# TECH SPEC — REVYX Market Expansion: Romania Rural + Ukraine Diaspora
**Document:** TECH_SPEC_REVYX_market-expansion-ro-ua_v1.0.0.md  
**Version:** 1.0.0  
**Phase:** S7 — Phase 4 Post-Launch  
**Status:** APPROVED  
**Date:** 2026-05-06  
**Authors:** Product · Backend Engineering · Legal  
**Reviewers:** Architect · Security · DBA · QA · Compliance · Product · Audit Lead

---

## Impact Assessment ★

| Dimension | Impact | Notes |
|---|---|---|
| Business | HIGH | Romania rural + Ukrainian diaspora = estimated 40% addressable market expansion |
| Compliance | CRITICAL | Ukraine: no EU adequacy decision → data residency strictly EU, zero transfer |
| Security | MEDIUM | ANCPI API integration = new external dependency surface |
| Infrastructure | MEDIUM | Multi-currency rates job + UA content translation pipeline |
| Legal | HIGH | Schrems II compliance for UA users confirmed: data stays in EU |
| Localization | HIGH | Builds on S7-1 multi-language (RU locale) |

---

## 1. Romania Rural / Mic-Urban Expansion

### 1.1 Problem Statement

Current geocoding assumes major urban centers (București, Cluj, Timișoara, etc.) with
established lat/lng precision. Romania rural:
- Addresses reference UAT (Unitate Administrativ-Teritorială) codes, not street grids
- Cadastral data requires ANCPI (Agenția Națională de Cadastru și Publicitate Imobiliară) integration
- Postal codes in rural areas map to multiple localities

### 1.2 ANCPI Integration

**Purpose:** Validate property cadastral number (număr cadastral) and retrieve
official surface/ownership data for rural listings.

**API:** ANCPI eTerra REST API (requires institutional account + SSL client certificate)

```go
// backend/internal/ancpi/client.go
type ANCPIClient struct {
    baseURL    string
    httpClient *http.Client  // mTLS configured
    apiKey     string        // from ANCPI_API_KEY env var
}

type CadastralInfo struct {
    CadastralNumber   string  `json:"nr_cadastral"`
    UAT               string  `json:"uat_code"`       // SIRUTA code
    LocalityName      string  `json:"localitate"`
    OfficialSurface   float64 `json:"suprafata_m2"`   // official m² from ANCPI
    OwnershipType     string  `json:"regim_proprietate"`
    LastUpdated       string  `json:"data_actualizare"`
    IsEncumbered      bool    `json:"are_sarcini"`    // ipoteci, servituți
}

func (c *ANCPIClient) GetCadastralInfo(ctx context.Context, cadastralNum string) (*CadastralInfo, error) {
    req, err := http.NewRequestWithContext(ctx, "GET",
        fmt.Sprintf("%s/imobile/%s", c.baseURL, url.PathEscape(cadastralNum)), nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("X-API-Key", c.apiKey)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("ancpi: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode == 404 {
        return nil, ErrCadastralNotFound
    }
    var info CadastralInfo
    return &info, json.NewDecoder(resp.Body).Decode(&info)
}
```

**Graceful degradation:** If ANCPI API unavailable or cadastral number not found,
listing is accepted without official data — flagged as `cadastral_verified: false`
in the property record. Buyers see a badge indicating unverified cadastral status.

```sql
-- migration 017 (S7)
ALTER TABLE properties
  ADD COLUMN IF NOT EXISTS cadastral_number    VARCHAR(30),
  ADD COLUMN IF NOT EXISTS cadastral_verified  BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS uat_code            VARCHAR(10),  -- SIRUTA code
  ADD COLUMN IF NOT EXISTS official_surface_m2 FLOAT;

CREATE INDEX idx_properties_uat ON properties (uat_code) WHERE uat_code IS NOT NULL;
```

### 1.3 UAT Code Geocoding

UAT (SIRUTA) codes enable geocoding of rural localities without street-level precision.

**Reference dataset:** SIRUTA database (public, updated annually by INS Romania)

```sql
-- migration 017 (S7) — static reference table
CREATE TABLE uat_geocodes (
    siruta_code  VARCHAR(10) PRIMARY KEY,
    uat_name     VARCHAR(100) NOT NULL,
    county_code  VARCHAR(5) NOT NULL,  -- județ
    uat_type     VARCHAR(20) NOT NULL, -- 'municipiu' | 'oras' | 'comuna'
    lat          FLOAT NOT NULL,
    lng          FLOAT NOT NULL,
    population   INT,
    -- Bounding box for zone clustering (S7-2 feature: lat_cluster, lng_cluster)
    bbox_sw_lat  FLOAT,
    bbox_sw_lng  FLOAT,
    bbox_ne_lat  FLOAT,
    bbox_ne_lng  FLOAT
);

-- Loaded from INS SIRUTA CSV at migration time
-- ~3200 UAT entries (all Romanian administrative units)
```

**Geocoding fallback chain:**
```
1. Street address → Google Maps Geocoding API (urban, existing)
2. Cadastral number → ANCPI lookup → UAT code → uat_geocodes centroid
3. Postal code → custom RO postal code table → county centroid
4. UAT name text match → uat_geocodes → centroid
5. County centroid (last resort)
```

---

## 2. Ukraine — Diaspora Market

### 2.1 Market Definition

Target: Ukrainian nationals living in EU countries (Romania, Poland, Germany, etc.)
who use REVYX to find real-estate in Romania for purchase/rent.

**Not in scope:** Properties located in Ukraine. No Ukrainian-domiciled tenants.

### 2.2 Data Residency — Schrems II Compliance

Ukraine has **no EU adequacy decision** under GDPR Art. 45.

**Binding rule: zero personal data transfer to Ukraine or to any system hosted outside EU.**

```
User UA accesses revyx.app (hosted EU) → request handled by EU infrastructure
Data stored: EU only (existing AWS eu-central-1 or equivalent)
UA users receive same GDPR protections as EU users
No sub-processor in Ukraine
No CDN PoP in Ukraine for personal data (static assets OK via CDN)
```

This is confirmed in TIA (Transfer Impact Assessment) doc:
`docs/legal/TIA_OPENAI_v1.0.0.md` — covers OpenAI as sub-processor.
**New TIA not required for UA market** (no data transferred to UA).

### 2.3 Multi-Currency Support

Ukrainian diaspora users may price properties in EUR, RON, or UAH for reference.

**Currencies supported:**
| Currency | Code | Use case |
|---|---|---|
| Romanian Leu | RON | Default for RO properties |
| Euro | EUR | International reference |
| Ukrainian Hryvnia | UAH | Diaspora reference price |

**Exchange rate source:** European Central Bank (ECB) Reference Rates API
(free, official EU source, updated every business day at ~16:00 CET)

```go
// backend/internal/currency/ecb.go
type ECBClient struct {
    httpClient *http.Client
    redis      *redis.Client
}

const ecbURL = "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml"

func (c *ECBClient) FetchRates(ctx context.Context) (map[string]float64, error) {
    // Check Redis cache first (TTL 6h — rates update once daily)
    cached, err := c.redis.Get(ctx, "revyx:ecb:rates").Result()
    if err == nil {
        var rates map[string]float64
        _ = json.Unmarshal([]byte(cached), &rates)
        return rates, nil
    }

    resp, err := c.httpClient.Get(ecbURL)
    // Parse XML, extract RON and UAH rates (both vs EUR base)
    // Cache result in Redis, TTL 6h
    ...
}

// Scheduler job: jobRefreshExchangeRates — runs daily at 17:00 UTC (after ECB publish)
```

```sql
-- migration 017 (S7)
CREATE TABLE exchange_rates (
    id          SERIAL PRIMARY KEY,
    base        VARCHAR(3) NOT NULL DEFAULT 'EUR',
    currency    VARCHAR(3) NOT NULL,
    rate        FLOAT NOT NULL,  -- units of currency per 1 EUR
    fetched_at  TIMESTAMPTZ NOT NULL,
    source      VARCHAR(50) NOT NULL DEFAULT 'ecb',
    UNIQUE (base, currency, fetched_at::date)
);
```

**Property display:** Prices stored in RON (canonical). UI converts on the fly using
latest cached rate. Conversion is display-only — all storage, billing, and API
responses use RON.

### 2.4 Ukrainian Language Listings — Content Moderation + Translation

Tenants may submit property listings with descriptions in Ukrainian (Cyrillic script).

**Indexing pipeline:**
```
Tenant submits listing (description_ua: Ukrainian text)
        │
        ▼
Moderation check:
  - Length check (min 20 chars, max 5000 chars)
  - Language detection (lingua-go or similar)
  - Basic content filter (slurs, prohibited content)
        │
        ▼
AI translation → Romanian for platform indexing:
  POST ai.TranslateUA(ctx, ukrainianText) → romanianText
  Model: claude-haiku-4-5-20251001 (existing AI client)
  Stored in: properties.description (RO) + properties.description_original (original)
        │
        ▼
Embedding generation on Romanian text (existing pipeline)
```

```go
// backend/internal/ai/ai.go — NEW function (add to existing file)
func (c *Client) TranslateUA(ctx context.Context, text string) (string, error) {
    if !c.IsAvailable() {
        return text, nil  // fallback: use original text for embedding
    }
    resp, err := c.complete(ctx, fmt.Sprintf(
        "Translate the following Ukrainian real-estate listing text to Romanian. "+
            "Output only the translation, no commentary.\n\n%s", text,
    ), 1000)
    if err != nil {
        return text, err
    }
    return resp, nil
}
```

```sql
-- migration 017 (S7)
ALTER TABLE properties
  ADD COLUMN IF NOT EXISTS description_original TEXT,  -- original language
  ADD COLUMN IF NOT EXISTS description_lang     VARCHAR(5) DEFAULT 'ro';
  -- 'ro' | 'ru' | 'uk'
```

### 2.5 Compliance — Ukrainian Users

```
Data processing basis for UA users:
  - Article 6(1)(b): contract performance (they use the service)
  - No transfer outside EU: confirmed
  - Privacy policy: must be available in Russian (S7-1 multilang covers this)
  - Right to erasure: same GDPR Art. 17 flow as EU users
  
Sub-processor note:
  - OpenAI: TranslateUA calls go to OpenAI (via existing ai.go client)
  - Covered by TIA_OPENAI_v1.0.0.md (EU-US SCCs)
  - Translation input = listing description text (no PII)
```

---

## 3. Scheduler Integration

New job added to `backend/internal/scheduler/scheduler.go`:

```go
// jobRefreshExchangeRates — daily at 17:00 UTC
func (s *Scheduler) jobRefreshExchangeRates() {
    rates, err := s.ecbClient.FetchRates(context.Background())
    if err != nil {
        s.logger.Error("ecb rate refresh failed", zap.Error(err))
        return
    }
    if err := s.db.InsertExchangeRates(context.Background(), rates); err != nil {
        s.logger.Error("exchange rate DB insert failed", zap.Error(err))
    }
}
```

---

## 4. API Changes

### 4.1 Property Create/Update — New Fields

```
POST /api/v1/properties
  New optional fields:
    cadastral_number: string
    uat_code: string
    description_original: string
    description_lang: "ro"|"ru"|"uk"
    currency_display: "RON"|"EUR"|"UAH"  // display preference only
```

### 4.2 Currency Conversion Endpoint

```
GET /api/v1/currency/rates
  Response: { base: "RON", rates: { EUR: 0.2012, UAH: 8.34 }, updated_at: "..." }
  Auth: public (no auth required — rates are public data)
  Cache: Redis 6h
```

---

## 5. Audit Checkpoint — S7-4 Market Expansion RO+UA ★

**Architect:** ANCPI integration as optional enrichment (not blocking listing creation) is the correct degraded-mode design. SIRUTA lookup table (3200 rows) is small enough for in-memory cache. ECB rates via Redis with 6h TTL is correct — rates are stable within a day. Property prices stored canonically in RON with display conversion is the right single-source-of-truth approach. Ukrainian listing translation using existing `ai.go` `TranslateUA` function follows the "integrate, don't rewrite" rule. ✅

**Security:** ANCPI mTLS client certificate must be stored in AWS Secrets Manager (not in env vars). `ANCPI_API_KEY` added to env var list (infra/.env.example). ECB API is public HTTP — no auth needed, no secret surface. Translation of listing descriptions via OpenAI: listing text may contain addresses (not personal data per se, but sensitive). TranslateUA must not pass buyer PII — enforcement: call only on `description` field, never on buyer/agent fields. ✅

**DBA:** Migration 017 adds 5 columns to `properties` (large table) — use `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` (non-locking in PostgreSQL 11+). SIRUTA table loaded at migration time: ensure CSV is included in the migration `data/` directory. `exchange_rates` UNIQUE constraint on `(base, currency, fetched_at::date)` prevents duplicate daily inserts from retried jobs. ✅

**QA:** Required tests: (1) ANCPI lookup unavailable → listing accepted with `cadastral_verified: false`, (2) ECB rate fetch → rates stored in DB + Redis cache, (3) currency conversion endpoint returns valid JSON, (4) Ukrainian description text triggers TranslateUA → RO translation stored in `description`, original in `description_original`, (5) UA user request served from EU infra (no cross-region redirect to non-EU endpoint). ✅

**Compliance:** Schrems II — UA user data stays in EU: ✅ confirmed by architecture (EU-only infra, no UA PoP for personal data). OpenAI for TranslateUA: listing descriptions contain no PII (addresses are property addresses, not personal). TIA_OPENAI_v1.0.0.md covers this use case. **Action required: privacy policy update to mention Ukrainian user support and language availability.** DPIA: no new category of personal data — UA users are subject to same profile as EU users. ✅

**Product:** ANCPI verification badge in UI ("Verificat ANCPI" ✓) is a trust differentiator for rural listings. UAH display rate must show "only indicative" disclaimer (ECB rate is reference, not market rate). Ukrainian listing moderation: content filter must be reviewed for Ukrainian-language patterns (existing Romanian filter insufficient). ✅

**Audit Lead:** **Items to track before GA:**
- [ ] ANCPI institutional account + SSL certificate provisioned
- [ ] Privacy policy updated for Ukrainian user support (Legal sign-off)
- [ ] Content moderation extended to Ukrainian language input
- [ ] SIRUTA CSV included in migration 017 data/ directory

---

*End of TECH_SPEC_REVYX_market-expansion-ro-ua_v1.0.0.md*
