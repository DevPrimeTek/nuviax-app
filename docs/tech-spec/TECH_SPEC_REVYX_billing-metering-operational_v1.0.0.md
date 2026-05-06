# TECH SPEC — REVYX Billing & Metering Operational
**Document:** TECH_SPEC_REVYX_billing-metering-operational_v1.0.0.md  
**Version:** 1.0.0  
**Phase:** S7 — Phase 4 Post-Launch  
**Status:** APPROVED  
**Date:** 2026-05-06  
**Authors:** Backend Engineering · Finance · Product  
**Reviewers:** Architect · Security · DBA · QA · Compliance · Product · Audit Lead

---

## Impact Assessment ★

| Dimension | Impact | Notes |
|---|---|---|
| Revenue | CRITICAL | Billing correctness directly impacts revenue recognition |
| Security | HIGH | Stripe webhooks, payment data, tenant financial state — critical attack surface |
| Compliance | HIGH | Invoice data retention 10 years (Romanian fiscal law); GDPR for PII in invoices |
| Reliability | HIGH | Billing failures = revenue loss + tenant disruption |
| Infrastructure | MEDIUM | Stripe Meters API + webhook consumer + grace period state machine |
| Finance | HIGH | Internal cost allocation (AI + infra per tenant) for margin reporting |

---

## 1. Billing Plans

| Plan | Listings Limit | AI Model | SLA | Price (EUR/mo) |
|---|---|---|---|---|
| **Starter** | 500 | Local sentence-transformers only | 99.5% | 49 |
| **Growth** | 5,000 | Local + OpenAI opt-in | 99.9% | 199 |
| **Enterprise** | Unlimited | Dedicated infra + OpenAI | 99.9% custom | Custom |

**Overage policy:**
- Starter: listings 501+ blocked until next billing cycle or upgrade
- Growth: listings 5001+ billed at €0.02/listing/day overage
- Enterprise: no overage — negotiated capacity

```sql
-- migration 019 (S7)
ALTER TABLE tenants
  ADD COLUMN IF NOT EXISTS plan               VARCHAR(20) NOT NULL DEFAULT 'starter'
    CHECK (plan IN ('starter', 'growth', 'enterprise')),
  ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(50),
  ADD COLUMN IF NOT EXISTS stripe_subscription_id VARCHAR(50),
  ADD COLUMN IF NOT EXISTS billing_status     VARCHAR(20) NOT NULL DEFAULT 'active'
    CHECK (billing_status IN ('active', 'grace_period', 'suspended', 'cancelled')),
  ADD COLUMN IF NOT EXISTS grace_period_until TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS suspended_at       TIMESTAMPTZ;
```

---

## 2. Usage Metering

### 2.1 Metered Dimensions

| Meter Name (Stripe) | Source Table | Aggregation | Notes |
|---|---|---|---|
| `revyx_listings_count` | `properties` | max(daily count) per month | Peak count, not cumulative |
| `revyx_embedding_tokens` | `embedding_usage_log` | sum(tokens_used) per month | OpenAI only (local model billed flat) |
| `revyx_api_calls` | `api_usage_log` | count per month | Growth+ only; Starter has flat rate |
| `revyx_ai_tasks` | `ai_task_log` | count per month | GenerateTaskTexts, ParseAndSuggest, etc. |

### 2.2 Metering Collection

```go
// backend/internal/billing/metering.go
type MeteringService struct {
    db     *db.DB
    stripe *stripe.Client
    redis  *redis.Client
}

// Called after every successful property insert/delete
func (m *MeteringService) RecordListingChange(ctx context.Context, tenantID string, delta int) error {
    // Update Redis counter (fast path)
    key := fmt.Sprintf("meter:listings:%s:%s", tenantID, currentMonth())
    if delta > 0 {
        m.redis.IncrBy(ctx, key, int64(delta))
    } else {
        m.redis.DecrBy(ctx, key, int64(-delta))
    }
    m.redis.Expire(ctx, key, 40*24*time.Hour)
    return nil
}

// Called after every embedding API call
func (m *MeteringService) RecordEmbeddingUsage(ctx context.Context, tenantID string, tokens int, model string) error {
    _, err := m.db.InsertEmbeddingUsage(ctx, db.EmbeddingUsage{
        TenantID:  tenantID,
        Tokens:    tokens,
        Model:     model,
        CreatedAt: time.Now(),
    })
    return err
}

// Hourly job: flush Redis counters → Stripe Meter Events
func (m *MeteringService) FlushToStripe(ctx context.Context) error {
    tenants, _ := m.db.GetActiveTenants(ctx)
    for _, tenantID := range tenants {
        key := fmt.Sprintf("meter:listings:%s:%s", tenantID, currentMonth())
        count, _ := m.redis.Get(ctx, key).Int64()

        _, err := m.stripe.Billing.MeterEvents.New(&stripe.BillingMeterEventNewParams{
            EventName: stripe.String("revyx_listings_count"),
            Payload: map[string]string{
                "stripe_customer_id": m.db.GetStripeCustomerID(tenantID),
                "value":              strconv.FormatInt(count, 10),
            },
            Timestamp: stripe.Int64(time.Now().Unix()),
        })
        if err != nil {
            m.logger.Error("stripe meter flush failed", zap.String("tenant", tenantID), zap.Error(err))
        }
    }
    return nil
}
```

### 2.3 Billing Scheduler Jobs

```go
// Added to scheduler.go (S7)
// jobFlushMeteringToStripe — runs every hour
// jobGenerateMonthlyReport — runs on 1st of month at 08:00 UTC
// jobCheckGracePeriodExpiry — runs every 6 hours
```

---

## 3. Stripe Integration

### 3.1 Setup

```go
// backend/internal/billing/stripe_client.go
import "github.com/stripe/stripe-go/v76"

func NewStripeClient() *stripe.Client {
    stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
    return &stripe.Client{}
}
```

**Env vars (new, S7):**
```
STRIPE_SECRET_KEY=sk_live_...        # from Stripe dashboard
STRIPE_WEBHOOK_SECRET=whsec_...     # for webhook signature verification
STRIPE_PRICE_STARTER=price_xxx      # Stripe Price ID for Starter plan
STRIPE_PRICE_GROWTH=price_xxx       # Stripe Price ID for Growth plan
```

### 3.2 Webhook Consumer

```
POST /api/v1/billing/stripe-webhook
  Headers: Stripe-Signature (Stripe signs with STRIPE_WEBHOOK_SECRET)
  Body: Stripe event JSON
  Auth: Stripe signature verification (no JWT)
```

```go
// backend/internal/api/handlers/billing.go
var relevantEvents = map[string]bool{
    "invoice.payment_failed":     true,
    "invoice.payment_succeeded":  true,
    "customer.subscription.deleted": true,
    "customer.subscription.updated": true,
}

func (h *Handler) StripeWebhook(c *fiber.Ctx) error {
    payload := c.Body()
    sig := c.Get("Stripe-Signature")

    event, err := webhook.ConstructEvent(payload, sig, os.Getenv("STRIPE_WEBHOOK_SECRET"))
    if err != nil {
        return fiber.ErrUnauthorized
    }

    if !relevantEvents[string(event.Type)] {
        return c.SendStatus(fiber.StatusOK)  // ack unknown events
    }

    // Process async to stay within Stripe's 30s response window
    go h.billing.ProcessStripeEvent(context.Background(), event)
    return c.SendStatus(fiber.StatusOK)
}
```

### 3.3 Event Handlers

```go
// backend/internal/billing/events.go
func (b *BillingService) ProcessStripeEvent(ctx context.Context, event stripe.Event) {
    switch event.Type {
    case "invoice.payment_failed":
        b.handlePaymentFailed(ctx, event)
    case "invoice.payment_succeeded":
        b.handlePaymentSucceeded(ctx, event)
    case "customer.subscription.deleted":
        b.handleSubscriptionCancelled(ctx, event)
    case "customer.subscription.updated":
        b.handleSubscriptionUpdated(ctx, event)
    }
}

func (b *BillingService) handlePaymentFailed(ctx context.Context, event stripe.Event) {
    var invoice stripe.Invoice
    _ = json.Unmarshal(event.Data.Raw, &invoice)

    tenantID := b.db.GetTenantByStripeCustomer(ctx, invoice.Customer.ID)
    if tenantID == "" {
        return
    }

    // Start 7-day grace period
    graceUntil := time.Now().Add(7 * 24 * time.Hour)
    _ = b.db.UpdateTenantBillingStatus(ctx, tenantID, "grace_period", &graceUntil)

    // Notify tenant admin
    _ = b.email.SendPaymentFailedNotice(ctx, tenantID, graceUntil)

    b.logger.Warn("payment failed, grace period started",
        zap.String("tenant", tenantID),
        zap.Time("grace_until", graceUntil),
    )
}

func (b *BillingService) handlePaymentSucceeded(ctx context.Context, event stripe.Event) {
    var invoice stripe.Invoice
    _ = json.Unmarshal(event.Data.Raw, &invoice)

    tenantID := b.db.GetTenantByStripeCustomer(ctx, invoice.Customer.ID)
    _ = b.db.UpdateTenantBillingStatus(ctx, tenantID, "active", nil)
    _ = b.db.RecordInvoice(ctx, tenantID, invoice)
}
```

---

## 4. Grace Period & Suspension State Machine

```
[active]
    │   invoice.payment_failed
    ▼
[grace_period] ─── invoice.payment_succeeded ──► [active]
    │
    │   grace_period_until elapsed (checked every 6h)
    ▼
[suspended]
    │   tenant pays manually + admin reactivates
    ▼
[active]

[grace_period or suspended]
    │   customer.subscription.deleted
    ▼
[cancelled]
```

```go
// jobCheckGracePeriodExpiry — runs every 6 hours
func (s *Scheduler) jobCheckGracePeriodExpiry() {
    tenants, _ := s.db.GetGracePeriodExpired(context.Background())
    for _, tenantID := range tenants {
        _ = s.db.UpdateTenantBillingStatus(context.Background(), tenantID, "suspended", nil)
        s.logger.Warn("tenant suspended — grace period elapsed", zap.String("tenant", tenantID))
        // Notify tenant: account suspended, contact support
        _ = s.email.SendSuspensionNotice(context.Background(), tenantID)
    }
}
```

**Suspended tenant behavior:**
- API requests return `402 Payment Required` with `{"error": "account_suspended", "support_url": "..."}`
- Read-only access preserved for 30 days (data download)
- After 30 days: data scheduled for deletion (GDPR Art. 17 basis: contract terminated)

---

## 5. Invoice Storage

```sql
-- migration 019 (S7)
CREATE TABLE invoices (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL REFERENCES tenants(id),
    stripe_invoice_id   VARCHAR(50) UNIQUE NOT NULL,
    amount_eur          FLOAT NOT NULL,
    status              VARCHAR(20) NOT NULL,  -- 'paid' | 'open' | 'void' | 'uncollectible'
    billing_period_start DATE NOT NULL,
    billing_period_end   DATE NOT NULL,
    line_items          JSONB NOT NULL,  -- breakdown: listings, embeddings, API calls
    pdf_url             TEXT,            -- Stripe-hosted invoice PDF
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at             TIMESTAMPTZ,
    INDEX idx_invoices_tenant (tenant_id, billing_period_start DESC)
);

-- Retention: 10 years (Romanian fiscal law art. 25 din Legea contabilității)
-- Implemented via Postgres table archiving policy (move to archive schema after 2 years)
```

---

## 6. Internal Cost Allocation & Margin Report

For each tenant, track:
- **AI cost:** OpenAI API calls (embedding + completions) — from `embedding_usage_log` × current OpenAI price
- **Infra cost:** Compute allocation (vCPU-hours × tenant's share of total requests)
- **Revenue:** Monthly invoice amount (from `invoices`)
- **Gross margin:** `(revenue - ai_cost - infra_cost) / revenue`

```sql
-- Monthly cost report query (run on 1st of each month)
SELECT
    t.id AS tenant_id,
    t.name,
    t.plan,
    SUM(eu.tokens_used) FILTER (WHERE eu.model = 'text-embedding-3-small') AS embedding_tokens,
    SUM(eu.tokens_used) * 0.00000002 AS estimated_openai_cost_eur,  -- $0.02/1M tokens
    COUNT(DISTINCT p.id) AS active_listings,
    i.amount_eur AS monthly_revenue,
    ROUND(
        (i.amount_eur - SUM(eu.tokens_used) * 0.00000002) / NULLIF(i.amount_eur, 0) * 100, 2
    ) AS gross_margin_pct
FROM tenants t
LEFT JOIN embedding_usage_log eu ON eu.tenant_id = t.id
    AND eu.created_at >= date_trunc('month', NOW() - INTERVAL '1 month')
    AND eu.created_at < date_trunc('month', NOW())
LEFT JOIN properties p ON p.tenant_id = t.id AND p.deleted_at IS NULL
LEFT JOIN invoices i ON i.tenant_id = t.id
    AND i.billing_period_start = date_trunc('month', NOW() - INTERVAL '1 month')::date
    AND i.status = 'paid'
GROUP BY t.id, t.name, t.plan, i.amount_eur
ORDER BY monthly_revenue DESC NULLS LAST;
```

**Delivery:** Monthly report generated as JSON, stored in `billing_reports` table, accessible via `/api/admin/billing/margin-report?month=2026-05` (platform_admin only).

---

## 7. Plan Limit Enforcement

```go
// backend/internal/billing/enforcement.go
func (b *BillingService) CheckListingQuota(ctx context.Context, tenantID string) error {
    tenant, _ := b.db.GetTenant(ctx, tenantID)

    if tenant.BillingStatus == "suspended" || tenant.BillingStatus == "cancelled" {
        return ErrAccountSuspended
    }

    count, _ := b.db.CountTenantListings(ctx, tenantID)
    limit := planLimit(tenant.Plan)

    if limit > 0 && count >= limit {
        return ErrListingQuotaExceeded
    }
    return nil
}

func planLimit(plan string) int {
    switch plan {
    case "starter":  return 500
    case "growth":   return 5000
    case "enterprise": return 0  // unlimited
    default:         return 500
    }
}
```

This check is called in `CreateProperty` handler before inserting a new listing.

---

## 8. Audit Checkpoint — S7-6 Billing & Metering Operational ★

**Architect:** Stripe Meters API (usage-based billing) is the correct approach — avoids building a custom metering system. Hourly flush to Stripe (not real-time) is sufficient for monthly billing cycles and reduces Stripe API call volume. Grace period state machine (active → grace_period → suspended → cancelled) covers all billing lifecycle states. Redis counters for listing metering (fast path) + hourly DB flush is the correct two-tier approach. ✅

**Security:** `STRIPE_SECRET_KEY` and `STRIPE_WEBHOOK_SECRET` added to Secrets Manager (not env files). Stripe webhook consumer MUST verify `Stripe-Signature` header before processing — implemented via `webhook.ConstructEvent`. Invoice PDF URLs are Stripe-hosted (HTTPS, time-limited) — never stored raw, accessed via Stripe API. `402 Payment Required` for suspended tenants must be returned BEFORE processing any read/write request — enforce in billing middleware, not per-handler. ✅

**DBA:** `invoices` table: 10-year retention policy required (Romanian fiscal law). Implement via: (1) archive schema after 2 years (move to `archive.invoices`), (2) no DELETE from `invoices` for fiscal period. `tenants.billing_status` must be covered by audit trigger (status transitions are sensitive). `embedding_usage_log` is already in S6 schema — confirm `tokens_used` column exists (add if not). ✅

**QA:** Required tests: (1) `invoice.payment_failed` → grace period set in DB + email sent, (2) `invoice.payment_succeeded` → billing_status back to 'active', (3) grace period expiry job → tenant suspended, (4) suspended tenant → 402 on property create, (5) listing quota exceeded (Starter >500) → 429 on create, (6) Stripe webhook with invalid signature → 401, (7) duplicate webhook event (same invoice ID) → idempotent (no double-grace-period). ✅

**Compliance:** Invoice storage: 10 years per art. 25 Legea contabilității 82/1991. Invoices contain tenant company name/address (not personal data of natural persons in B2B context). Suspension after grace period with 30-day read-only window satisfies GDPR Art. 17 (erasure on contract termination) by giving tenants time to export data. **Action required: Terms of Service must specify grace period duration and data deletion timeline.** ✅

**Product:** Grace period 7 days is standard for B2B SaaS — confirm with Finance. Enterprise plan requires custom Stripe contract (not Meters API) — Enterprise billing is out of scope for initial S7 automated billing (manual invoice process until S8). Margin report must be reviewed monthly by Finance Lead before distribution. ✅

**Audit Lead:** **Items to track before billing go-live:**
- [ ] Stripe production account configured (meters + webhooks + pricing)
- [ ] `STRIPE_SECRET_KEY` + `STRIPE_WEBHOOK_SECRET` in production Secrets Manager
- [ ] Invoice retention policy implemented (archive schema + no-DELETE policy for fiscal records)
- [ ] Terms of Service updated with grace period + data deletion timeline (Legal sign-off)
- [ ] Billing middleware (402 enforcement) verified in staging before production

---

*End of TECH_SPEC_REVYX_billing-metering-operational_v1.0.0.md*
