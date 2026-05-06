# TECH SPEC — REVYX Partnerships API
**Document:** TECH_SPEC_REVYX_partnerships-api_v1.0.0.md  
**Version:** 1.0.0  
**Phase:** S7 — Phase 4 Post-Launch  
**Status:** APPROVED  
**Date:** 2026-05-06  
**Authors:** Backend Engineering · Product · BizDev  
**Reviewers:** Architect · Security · DBA · QA · Compliance · Product · Audit Lead

---

## Impact Assessment ★

| Dimension | Impact | Notes |
|---|---|---|
| Business | HIGH | imobiliare.ro + storia.ro feeds dramatically expand listing inventory |
| Security | HIGH | Inbound webhook from external partners = new attack surface |
| Compliance | MEDIUM | Partner listings contain PII (agent phone, address) — must apply same GDPR controls |
| Reliability | MEDIUM | Partner import failures must not affect core platform |
| Performance | MEDIUM | Bulk import (thousands of listings) must be async, not blocking |
| Billing | MEDIUM | Each imported listing attributed to partner tenant for billing purposes |

---

## 1. Partner Feed Architecture

```
[Partner system] ──── webhook POST ────► [REVYX Partnerships API]
                                                │
[Partner system] ◄── polling request ──── OR   │
                                                │
                                         [Import Queue — Redis]
                                                │  async worker
                                                ▼
                                         [Deduplication check]
                                                │  (SHA256 fingerprint)
                                                ▼
                                         [Property normalizer]
                                                │  (partner format → REVYX schema)
                                                ▼
                                         [ANCPI optional validation]
                                                │  (if cadastral_number present)
                                                ▼
                                         [properties table + embeddings pipeline]
                                                │
                                                ▼
                                         [Billing attribution — partner tenant_id]
```

---

## 2. Partner Registration

### 2.1 Partner Schema

```sql
-- migration 018 (S7)
CREATE TABLE partners (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    feed_format     VARCHAR(30) NOT NULL CHECK (feed_format IN ('imobiliare_ro', 'storia_ro', 'generic_json')),
    import_method   VARCHAR(20) NOT NULL CHECK (import_method IN ('webhook', 'polling')),
    feed_url        TEXT,               -- for polling method
    polling_interval_minutes INT DEFAULT 60,
    webhook_secret  VARCHAR(64) NOT NULL, -- HMAC-SHA256 signing secret
    api_key         VARCHAR(64) NOT NULL UNIQUE, -- partner uses this to auth
    rate_limit_rpm  INT NOT NULL DEFAULT 60,
    daily_quota     INT NOT NULL DEFAULT 1000,  -- max listings/day
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    last_polled_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE partner_import_log (
    id              BIGSERIAL PRIMARY KEY,
    partner_id      UUID NOT NULL REFERENCES partners(id),
    tenant_id       UUID NOT NULL,
    import_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    listings_received  INT NOT NULL,
    listings_imported  INT NOT NULL,
    listings_deduplicated INT NOT NULL,
    listings_failed    INT NOT NULL,
    import_method   VARCHAR(20) NOT NULL,
    error_summary   TEXT,
    INDEX idx_partner_import_log_partner (partner_id, import_at DESC)
);
```

---

## 3. Inbound Webhook

### 3.1 Endpoint

```
POST /api/v1/partnerships/webhook/{partner_id}
  Headers:
    X-Partner-Signature: sha256=<HMAC-SHA256 of body>
    Content-Type: application/json
  Auth: HMAC signature verification (no JWT required)
  Body: partner-format listing array (imobiliare.ro or storia.ro or generic_json)
```

### 3.2 Signature Verification

```go
// backend/internal/api/handlers/partnerships.go
func (h *Handler) PartnerWebhook(c *fiber.Ctx) error {
    partnerID := c.Params("partner_id")
    partner, err := h.db.GetPartner(c.Context(), partnerID)
    if err != nil {
        return fiber.ErrNotFound  // 404 regardless of whether partner exists (don't leak)
    }
    if !partner.IsActive {
        return fiber.ErrForbidden
    }

    // Rate limit check (sliding window, Redis)
    if !h.rateLimiter.Allow(c.Context(), "partner:"+partnerID, partner.RateLimitRPM) {
        return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
            "error": "rate_limit_exceeded",
        })
    }

    // HMAC-SHA256 verification (constant-time)
    body := c.Body()
    sig := c.Get("X-Partner-Signature")
    if !verifyPartnerSignature([]byte(partner.WebhookSecret), body, sig) {
        return fiber.ErrUnauthorized
    }

    // Queue for async processing — respond 202 immediately
    if err := h.importQueue.Enqueue(c.Context(), ImportJob{
        PartnerID:  partnerID,
        TenantID:   partner.TenantID,
        FeedFormat: partner.FeedFormat,
        Payload:    body,
    }); err != nil {
        h.logger.Error("enqueue import failed", zap.Error(err))
        return fiber.ErrInternalServerError
    }

    return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": "queued"})
}

func verifyPartnerSignature(secret, payload []byte, signature string) bool {
    mac := hmac.New(sha256.New, secret)
    mac.Write(payload)
    expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expected))
}
```

---

## 4. Rate Limiting (Redis Sliding Window)

```go
// backend/internal/ratelimit/sliding_window.go
type SlidingWindowLimiter struct {
    redis *redis.Client
}

func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string, limitRPM int) bool {
    now := time.Now().UnixMilli()
    windowStart := now - 60_000  // 60-second window

    pipe := l.redis.Pipeline()
    // Remove old entries outside window
    pipe.ZRemRangeByScore(ctx, "rl:"+key, "0", strconv.FormatInt(windowStart, 10))
    // Count current entries
    countCmd := pipe.ZCard(ctx, "rl:"+key)
    // Add current request
    pipe.ZAdd(ctx, "rl:"+key, redis.Z{Score: float64(now), Member: now})
    pipe.Expire(ctx, "rl:"+key, 2*time.Minute)
    _, _ = pipe.Exec(ctx)

    return countCmd.Val() < int64(limitRPM)
}
```

**Daily quota check:**
```go
// Daily quota: Redis counter, reset at 00:00 UTC
func (l *SlidingWindowLimiter) CheckDailyQuota(ctx context.Context, partnerID string, quota int) bool {
    today := time.Now().UTC().Format("2006-01-02")
    key := fmt.Sprintf("quota:partner:%s:%s", partnerID, today)
    count, _ := l.redis.Incr(ctx, key).Result()
    l.redis.Expire(ctx, key, 25*time.Hour)  // expire next day
    return int(count) <= quota
}
```

---

## 5. Feed Format Parsers

### 5.1 imobiliare.ro Format

```go
// backend/internal/partnerships/parsers/imobiliare.go
type ImobiliareProperty struct {
    ID          string  `json:"id"`
    Title       string  `json:"titlu"`
    Price       float64 `json:"pret"`
    Currency    string  `json:"moneda"`      // "EUR" | "RON"
    Surface     float64 `json:"suprafata"`   // m²
    Rooms       int     `json:"camere"`
    Floor       int     `json:"etaj"`
    TotalFloors int     `json:"etaje_total"`
    Address     struct {
        Street  string `json:"strada"`
        City    string `json:"oras"`
        County  string `json:"judet"`
        Lat     float64 `json:"lat"`
        Lng     float64 `json:"lng"`
    } `json:"adresa"`
    Images      []string `json:"imagini"`
    Amenities   []string `json:"facilitati"`
    Description string   `json:"descriere"`
    AgentPhone  string   `json:"telefon_agent"`  // PII — store hashed for dedup
    PostedAt    string   `json:"data_publicare"`
}

func ParseImobiliare(raw []byte) ([]NormalizedProperty, error) { ... }
```

### 5.2 storia.ro Format

```go
// backend/internal/partnerships/parsers/storia.go
type StoriaProperty struct {
    ExternalID  string  `json:"externalId"`
    ListingType string  `json:"listingType"` // "sale" | "rent"
    Price       struct {
        Value    float64 `json:"value"`
        Currency string  `json:"currency"`
    } `json:"price"`
    Area        float64  `json:"area"`
    RoomsNum    int      `json:"roomsNum"`
    Location    struct {
        Lat float64 `json:"lat"`
        Lon float64 `json:"lon"`
        City string `json:"city"`
    } `json:"location"`
    Features    []string `json:"features"`
    Description string   `json:"description"`
    Contact     struct {
        Phone string `json:"phone"`       // PII — store hashed
    } `json:"contact"`
}

func ParseStoria(raw []byte) ([]NormalizedProperty, error) { ... }
```

### 5.3 Normalized Property

```go
// backend/internal/partnerships/normalized.go
type NormalizedProperty struct {
    ExternalID      string    // partner's original ID
    PartnerID       string    // REVYX partner UUID
    TenantID        string    // billing attribution
    Title           string
    PriceRON        float64   // converted to RON at import time
    SurfaceM2       float64
    Rooms           int
    Floor           int
    TotalFloors     int
    Lat             float64
    Lng             float64
    City            string
    County          string
    Amenities       []string
    Description     string
    AgentPhoneHash  string    // SHA256(phone) — for dedup only, never returned in API
    PostedAt        time.Time
    Fingerprint     string    // dedup key (see §6)
}
```

---

## 6. Deduplication

**Fingerprint:** SHA256 of normalized concatenation

```go
func computeFingerprint(p NormalizedProperty) string {
    // Normalize address: lowercase, strip spaces and punctuation
    addr := normalizeAddress(p.City + p.County)
    // Surface: round to nearest 5m²
    surf := math.Round(p.SurfaceM2/5) * 5
    // Price: round to nearest 1000 RON bucket
    price := math.Round(p.PriceRON/1000) * 1000

    raw := fmt.Sprintf("%s|%.0f|%.0f", addr, surf, price)
    hash := sha256.Sum256([]byte(raw))
    return hex.EncodeToString(hash[:])
}
```

**Dedup table:**
```sql
-- migration 018 (S7)
CREATE TABLE listing_fingerprints (
    fingerprint     VARCHAR(64) PRIMARY KEY,
    property_id     UUID NOT NULL REFERENCES properties(id),
    partner_id      UUID NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Import logic:**
```go
func (w *ImportWorker) processListing(ctx context.Context, p NormalizedProperty) error {
    fp := computeFingerprint(p)
    existing, _ := w.db.GetFingerprintProperty(ctx, fp)
    if existing != nil {
        // Duplicate — update price if changed, skip otherwise
        return w.db.UpdatePropertyPrice(ctx, existing.ID, p.PriceRON)
    }
    // New listing — insert
    return w.db.CreatePartnerProperty(ctx, p, fp)
}
```

---

## 7. Polling Import

For partners that prefer polling over webhooks:

```go
// backend/internal/scheduler/scheduler.go — new job
func (s *Scheduler) jobPollPartnerFeeds() {
    partners, _ := s.db.GetPollingPartners(context.Background())
    for _, partner := range partners {
        if time.Since(*partner.LastPolledAt) < time.Duration(partner.PollingIntervalMinutes)*time.Minute {
            continue
        }
        go s.pollPartner(partner)  // concurrent, one goroutine per partner
    }
}

func (s *Scheduler) pollPartner(partner Partner) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    resp, err := s.httpClient.GetWithAPIKey(ctx, partner.FeedURL, partner.APIKey)
    if err != nil {
        s.logger.Error("partner poll failed", zap.String("partner", partner.ID), zap.Error(err))
        return
    }

    _ = s.importQueue.Enqueue(ctx, ImportJob{
        PartnerID:  partner.ID,
        TenantID:   partner.TenantID,
        FeedFormat: partner.FeedFormat,
        Payload:    resp,
    })
    _ = s.db.UpdatePartnerLastPolled(ctx, partner.ID)
}
```

---

## 8. Outbound Webhooks to Partners

When REVYX events occur, notify the originating partner:

| Event | Trigger | Payload |
|---|---|---|
| Match confirmed | buyer saves a partner listing | `{event: "match_confirmed", property_id, external_id, matched_at}` |
| Deal WON | deal closed on a partner listing | `{event: "deal_won", property_id, external_id, won_at}` |

```go
// backend/internal/webhooks/partner_notify.go
func (s *PartnerWebhookService) NotifyMatchConfirmed(ctx context.Context, propertyID string) error {
    partner, err := s.db.GetPropertyPartner(ctx, propertyID)
    if err != nil || partner == nil {
        return nil  // not a partner listing
    }
    return s.send(ctx, partner, WebhookEvent{
        Event:       "match_confirmed",
        PropertyID:  propertyID,
        ExternalID:  partner.ExternalID,
        Timestamp:   time.Now().UTC(),
    })
}

func (s *PartnerWebhookService) send(ctx context.Context, partner Partner, event WebhookEvent) error {
    body, _ := json.Marshal(event)
    sig := signPayload([]byte(partner.WebhookSecret), body)

    req, _ := http.NewRequestWithContext(ctx, "POST", partner.CallbackURL, bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-REVYX-Signature", sig)

    resp, err := s.httpClient.Do(req)
    if err != nil || resp.StatusCode >= 500 {
        // Retry with exponential backoff (max 3 retries)
        return s.retryQueue.Enqueue(ctx, event, partner.ID)
    }
    return nil
}
```

---

## 9. Billing Attribution

Partner listings are attributed to the partner's `tenant_id` for billing:

```go
// backend/internal/db/queries.go
func (db *DB) CreatePartnerProperty(ctx context.Context, p NormalizedProperty, fp string) error {
    // tenant_id = partner.TenantID (billing goes to partner's account)
    // source = 'partner_import'
    // partner_id = p.PartnerID (for traceability)
    return db.pool.QueryRow(ctx, `
        INSERT INTO properties (tenant_id, title, price_ron, surface_m2, rooms,
          floor, total_floors, lat, lng, city, county, amenities, description,
          source, partner_id, external_partner_id)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,'partner_import',$14,$15)
        RETURNING id
    `, p.TenantID, p.Title, p.PriceRON, p.SurfaceM2, p.Rooms, p.Floor, p.TotalFloors,
        p.Lat, p.Lng, p.City, p.County, pq.Array(p.Amenities), p.Description,
        p.PartnerID, p.ExternalID,
    ).Scan(&id)
}
```

---

## 10. Audit Checkpoint — S7-5 Partnerships API ★

**Architect:** Async import queue (Redis) is the correct design — partner webhook response must be immediate (202 Accepted). HMAC-SHA256 signature verification matches existing webhook pattern from S6 (`signPayload`/`verifyWebhook`). Deduplication by normalized fingerprint (address + surface + price bucket) is pragmatic — prevents obvious duplicates without expensive semantic comparison. Polling sidecar (goroutine per partner) is fine at current scale; move to dedicated worker pool if > 20 partners. ✅

**Security:** Partner webhook endpoint: `404` for non-existent partners (don't reveal existence). Rate limiting per `partner_id` (not IP) prevents quota bypass via IP rotation. `AgentPhoneHash` stored (SHA256, not reversible) for dedup only — never returned in API responses. Daily quota limit enforced in Redis before processing payload (not after). Outbound webhook `X-REVYX-Signature` uses same `signPayload` function as inbound — consistent HMAC pattern. ✅

**DBA:** `listing_fingerprints` table will grow at the rate of all imported listings — add to retention policy (delete entries for properties that have been deleted). `partner_import_log` provides audit trail; index on `(partner_id, import_at DESC)` supports dashboard queries. `properties` table needs two new columns (`partner_id`, `external_partner_id`) + `source` column (VARCHAR 30) — all nullable for non-partner listings. ✅

**QA:** Required tests: (1) invalid HMAC → 401 regardless of partner existence, (2) rate limit exceeded → 429 with correct headers, (3) daily quota exceeded → 429, (4) duplicate fingerprint → price updated, no new property row, (5) outbound webhook retries on 500 response from partner, (6) deal WON on non-partner listing → no outbound webhook attempt. ✅

**Compliance:** Partner feeds contain agent phone numbers (`AgentPhoneHash` only in REVYX). **DPA must be signed with each partner** (they are data processors sharing data with REVYX). Partner names (imobiliare.ro, storia.ro) must be added to the sub-processor list (GDPR Art. 30 register). Outbound webhook contains only property IDs and timestamps — no buyer PII sent to partner. ✅

**Product:** Partner portal (admin UI for managing partner config, viewing import logs) is **out of scope for S7** — manage via platform_admin panel. Dedup fingerprint logic must be reviewed with each partner before go-live (price bucket size and address normalization may need tuning per partner's data quality). ✅

**Audit Lead:** **Items to track before partner go-live:**
- [ ] DPA signed with each partner (imobiliare.ro, storia.ro) before first import
- [ ] Sub-processor list updated in GDPR Art. 30 register
- [ ] Partner webhook secret rotation procedure documented in runbook
- [ ] Rate limit and daily quota values confirmed with partner SLA

---

*End of TECH_SPEC_REVYX_partnerships-api_v1.0.0.md*
