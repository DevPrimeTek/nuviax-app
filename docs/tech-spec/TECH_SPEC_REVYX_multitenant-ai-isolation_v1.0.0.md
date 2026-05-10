# TECH SPEC — REVYX Multi-Tenant AI Isolation
**Document:** TECH_SPEC_REVYX_multitenant-ai-isolation_v1.0.0.md  
**Version:** 1.0.0  
**Phase:** S6 — Production Hardening Pre-Launch  
**Status:** APPROVED — Pending Implementation  
**Date:** 2026-05-06  
**Authors:** Platform Engineering · Security · ML Infrastructure  
**Reviewers:** Architect · Security · DBA · Compliance · Audit Lead

---

## Impact Assessment ★

| Dimension | Impact | Notes |
|---|---|---|
| Security | CRITICAL | Cross-tenant data leakage prevention is a hard requirement |
| Architecture | HIGH | Schema-per-tenant vs shared schema decision has long-term implications |
| Cost | HIGH | Per-tenant KMS keys + billing attribution adds operational overhead |
| Scalability | HIGH | Index isolation strategy determines max tenant count |
| Compliance | CRITICAL | GDPR data sovereignty requires demonstrable isolation |
| Performance | MEDIUM | Tenant-scoped queries may miss shared optimizations |
| Operational complexity | HIGH | Per-tenant config + model variants require robust management plane |

---

## 1. Context

REVYX serves real estate platforms (tenants) each with their own property catalog, buyer profiles, and business rules. Phase 2 deployed a shared-schema model with `tenant_id` columns and RLS policies. Phase 3 requires:

1. **Data sovereignty opt-in** — tenant can require embeddings stay local (no OpenAI API call)
2. **Per-tenant model variant** — OpenAI vs local sentence-transformers, with billing attribution
3. **Isolated vector indexes** — no cross-tenant data in ANN results (even under adversarial queries)
4. **Per-tenant engine configuration** — APS/IS/NBA weights fully isolated
5. **Crypto isolation** — distinct KMS envelope keys per tenant
6. **Audit trail** — every cross-tenant query attempt logged and alerted

---

## 2. Tenant Data Sovereignty ★

### 2.1 Opt-In Modes

```go
type TenantEmbeddingMode string

const (
    ModeOpenAI     TenantEmbeddingMode = "openai"      // Default: uses OpenAI API
    ModeLocalOnly  TenantEmbeddingMode = "local_only"  // Embeddings generated on-premise
    ModeHybrid     TenantEmbeddingMode = "hybrid"      // Local for properties, OpenAI for search queries
)
```

**Stored in:** `tenants.embedding_mode` (migration 014 ★)

```sql
-- migration 014: tenant embedding config
ALTER TABLE tenants ADD COLUMN embedding_mode VARCHAR(20) NOT NULL DEFAULT 'openai';
ALTER TABLE tenants ADD COLUMN embedding_model VARCHAR(100);
ALTER TABLE tenants ADD COLUMN data_residency_region VARCHAR(10) DEFAULT 'eu-west-1';
ALTER TABLE tenants ADD COLUMN local_model_endpoint TEXT;

ALTER TABLE tenants ADD CONSTRAINT chk_embedding_mode
    CHECK (embedding_mode IN ('openai', 'local_only', 'hybrid'));
```

### 2.2 Cluster Routing for Local-Only Tenants

Local model inference runs on dedicated inference nodes (not the main API cluster). Routing is handled by the `EmbeddingRouter`:

```go
type EmbeddingRouter struct {
    openaiClient  *openai.Client
    localEndpoint string          // http://inference-node:8080/embed
    db            *pgxpool.Pool
    cache         *redis.Client
}

func (r *EmbeddingRouter) Embed(ctx context.Context, tenantID string, text string) ([]float32, error) {
    mode, err := r.getTenantMode(ctx, tenantID)
    if err != nil {
        return nil, fmt.Errorf("tenant mode lookup: %w", err)
    }

    switch mode {
    case ModeOpenAI:
        return r.embedOpenAI(ctx, text)
    case ModeLocalOnly:
        return r.embedLocal(ctx, tenantID, text)
    case ModeHybrid:
        return r.embedLocal(ctx, tenantID, text) // properties via local
    default:
        return nil, fmt.Errorf("unknown embedding mode: %s", mode)
    }
}

func (r *EmbeddingRouter) embedLocal(ctx context.Context, tenantID, text string) ([]float32, error) {
    endpoint := r.localEndpoint
    // Per-tenant dedicated endpoint if configured
    if ep := r.getTenantLocalEndpoint(ctx, tenantID); ep != "" {
        endpoint = ep
    }
    return callLocalModel(ctx, endpoint, text)
}
```

### 2.3 Data Residency Enforcement

`local_only` tenants: all embedding calls validated at request time:

```go
// Middleware: block OpenAI calls for local_only tenants
func (m *DataSovereigntyMiddleware) ValidateEmbeddingCall(ctx context.Context, tenantID string, target string) error {
    mode, _ := m.getTenantMode(ctx, tenantID)
    if mode == ModeLocalOnly && target == "openai" {
        // Log as security event, return error
        m.audit.Log(ctx, AuditEvent{
            Type:     "data_sovereignty_violation_attempt",
            TenantID: tenantID,
            Severity: "HIGH",
            Detail:   "attempted OpenAI call for local_only tenant",
        })
        return ErrDataSovereigntyViolation
    }
    return nil
}
```

---

## 3. Per-Tenant Model Variant ★

### 3.1 Model Registry

```go
type EmbeddingModelConfig struct {
    Provider   string  // "openai" | "sentence-transformers" | "custom"
    ModelID    string  // "text-embedding-3-small" | "paraphrase-multilingual-mpnet-base-v2"
    Dimensions int     // 1536 | 768 | custom
    CostPer1M  float64 // USD per 1M tokens (for billing)
    MaxTokens  int     // context window
}

var ModelRegistry = map[string]EmbeddingModelConfig{
    "openai/text-embedding-3-small": {
        Provider: "openai", ModelID: "text-embedding-3-small",
        Dimensions: 1536, CostPer1M: 0.02, MaxTokens: 8191,
    },
    "local/paraphrase-multilingual": {
        Provider: "sentence-transformers",
        ModelID:  "paraphrase-multilingual-mpnet-base-v2",
        Dimensions: 768, CostPer1M: 0.0, MaxTokens: 512, // on-prem, no per-call cost
    },
    "local/romanian-bert": {
        Provider: "sentence-transformers", ModelID: "dumitrescuv/romanian-bert-base-uncased",
        Dimensions: 768, CostPer1M: 0.0, MaxTokens: 512,
    },
}
```

### 3.2 Billing Attribution ★

Each embedding call attributed to tenant's billing account:

```sql
-- table: embedding_usage_log (migration 015 ★)
CREATE TABLE embedding_usage_log (
    id           BIGSERIAL PRIMARY KEY,
    tenant_id    UUID NOT NULL REFERENCES tenants(id),
    model_key    VARCHAR(100) NOT NULL,
    token_count  INTEGER NOT NULL,
    cost_usd     NUMERIC(10,8) NOT NULL DEFAULT 0,
    purpose      VARCHAR(50) NOT NULL, -- 'property_index' | 'search_query' | 'bulk_reindex'
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_embedding_usage_tenant_date
    ON embedding_usage_log (tenant_id, created_at DESC);
```

```go
func (r *EmbeddingRouter) embedOpenAI(ctx context.Context, tenantID, text string) ([]float32, error) {
    resp, err := r.openaiClient.Embed(ctx, text)
    if err != nil { return nil, err }

    // Async billing attribution (non-blocking)
    go func() {
        model := ModelRegistry["openai/text-embedding-3-small"]
        cost := float64(resp.TokensUsed) / 1_000_000 * model.CostPer1M
        _ = r.logUsage(context.Background(), tenantID, "openai/text-embedding-3-small",
            resp.TokensUsed, cost, "property_index")
    }()

    return resp.Embedding, nil
}
```

### 3.3 Cost Dashboard API

```
GET /api/admin/tenants/:id/embedding-costs?from=2026-01-01&to=2026-05-01

Response:
{
  "tenant_id": "...",
  "period": { "from": "2026-01-01", "to": "2026-05-01" },
  "breakdown": [
    { "model": "openai/text-embedding-3-small", "calls": 12500, "tokens": 4200000, "cost_usd": 0.084 },
    { "model": "local/paraphrase-multilingual", "calls": 8300, "tokens": 0, "cost_usd": 0.0 }
  ],
  "total_cost_usd": 0.084
}
```

---

## 4. Tenant-Isolated Vector Indexes ★

### 4.1 Decision: Partial Indexes with tenant_id (Selected) vs Schema-Per-Tenant

**Option A: Schema-Per-Tenant**
- Complete DDL isolation: `tenant_<uuid>.properties`
- No RLS complexity
- Max tenant count limited by PostgreSQL schema overhead (~few hundred practical max)
- Migration complexity: schema creation per tenant onboarding
- Cross-tenant admin queries require `SET search_path`
- **Verdict: rejected** — operational complexity too high for MVP scale

**Option B: Partial Indexes with tenant_id WHERE clause (Selected ★)**
- Single shared table, RLS enforced at DB level
- Index per tenant: `CREATE INDEX ... WHERE tenant_id = '<uuid>'`
- Planner uses partial index for tenant-scoped queries automatically
- Scales to thousands of tenants with manageable index count
- Cross-tenant leakage prevented at both RLS + index level

```sql
-- Partial HNSW index per tenant (created at tenant onboarding)
CREATE INDEX CONCURRENTLY idx_properties_embedding_tenant_<uuid>
  ON properties USING hnsw (embedding vector_cosine_ops)
  WITH (m = 24, ef_construction = 100)
  WHERE tenant_id = '<uuid>';
```

**Index creation automation (tenant onboarding hook):**

```go
func (s *TenantService) OnboardTenant(ctx context.Context, t Tenant) error {
    // ... create tenant record, KMS key, etc.
    
    // Create isolated vector index
    params := selectHNSWParams(t.ExpectedPropertyCount)
    _, err := s.db.Exec(ctx, fmt.Sprintf(`
        CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_properties_embedding_%s
        ON properties USING hnsw (embedding vector_cosine_ops)
        WITH (m = %d, ef_construction = %d)
        WHERE tenant_id = $1
    `, sanitizeTenantID(t.ID), params.M, params.EFConstruction), t.ID)
    return err
}

// sanitizeTenantID: UUIDs are safe, but strip hyphens for index name
func sanitizeTenantID(id string) string {
    return strings.ReplaceAll(id, "-", "_")
}
```

### 4.2 RLS Policies (Reinforcement)

```sql
-- Enforce tenant isolation at DB level (defense-in-depth)
ALTER TABLE properties ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON properties
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- System role bypass (for admin/migration jobs only)
CREATE POLICY admin_bypass ON properties
    USING (current_setting('app.role', true) = 'system');
```

Application layer ALWAYS sets `app.current_tenant_id` before any query:

```go
func withTenantContext(ctx context.Context, conn *pgx.Conn, tenantID string) error {
    _, err := conn.Exec(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID)
    return err
}
```

### 4.3 Index Cleanup at Tenant Offboarding

```sql
-- Called by TenantService.OffboardTenant()
DROP INDEX CONCURRENTLY IF EXISTS idx_properties_embedding_<tenant_id_no_hyphens>;
```

---

## 5. Per-Tenant Engine Configuration (APS / IS / NBA) ★

### 5.1 Config Schema

Each tenant can override default weights and normalizers for the three main engines.

```sql
-- migration 016 ★
CREATE TABLE tenant_engine_configs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) UNIQUE,
    
    -- APS (Affinity Prediction Score) weights
    aps_price_weight     NUMERIC(4,3) DEFAULT 0.35,
    aps_location_weight  NUMERIC(4,3) DEFAULT 0.30,
    aps_features_weight  NUMERIC(4,3) DEFAULT 0.20,
    aps_semantic_weight  NUMERIC(4,3) DEFAULT 0.15,
    
    -- IS (Intent Score) config
    is_decay_days        INTEGER DEFAULT 30,
    is_min_interactions  INTEGER DEFAULT 3,
    
    -- NBA (Next Best Action) thresholds
    nba_hot_threshold    NUMERIC(4,3) DEFAULT 0.75,
    nba_warm_threshold   NUMERIC(4,3) DEFAULT 0.50,
    
    -- Normalizer overrides
    price_norm_percentile INTEGER DEFAULT 95, -- clamp at 95th percentile
    
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT aps_weights_sum CHECK (
        aps_price_weight + aps_location_weight + aps_features_weight + aps_semantic_weight = 1.0
    )
);
```

### 5.2 Config Loader with Strict Isolation

```go
type TenantEngineConfig struct {
    APS struct {
        PriceWeight    float64
        LocationWeight float64
        FeaturesWeight float64
        SemanticWeight float64
    }
    IS struct {
        DecayDays       int
        MinInteractions int
    }
    NBA struct {
        HotThreshold  float64
        WarmThreshold float64
    }
    PriceNormPercentile int
}

func (s *ConfigService) GetTenantConfig(ctx context.Context, tenantID string) (*TenantEngineConfig, error) {
    // Cache: Redis key `engine_config:<tenant_id>`, TTL 5min
    cached, err := s.cache.Get(ctx, "engine_config:"+tenantID).Result()
    if err == nil {
        var cfg TenantEngineConfig
        _ = json.Unmarshal([]byte(cached), &cfg)
        return &cfg, nil
    }

    // DB fallback — uses RLS, tenant_id enforced
    row := s.db.QueryRow(ctx, `
        SELECT aps_price_weight, aps_location_weight, aps_features_weight, aps_semantic_weight,
               is_decay_days, is_min_interactions,
               nba_hot_threshold, nba_warm_threshold, price_norm_percentile
        FROM tenant_engine_configs WHERE tenant_id = $1
    `, tenantID)
    
    cfg := &TenantEngineConfig{}
    err = row.Scan(
        &cfg.APS.PriceWeight, &cfg.APS.LocationWeight,
        &cfg.APS.FeaturesWeight, &cfg.APS.SemanticWeight,
        &cfg.IS.DecayDays, &cfg.IS.MinInteractions,
        &cfg.NBA.HotThreshold, &cfg.NBA.WarmThreshold,
        &cfg.PriceNormPercentile,
    )
    if errors.Is(err, pgx.ErrNoRows) {
        return s.defaultConfig(), nil // tenant uses platform defaults
    }
    return cfg, err
}
```

### 5.3 Cross-Tenant Config Leakage Test ★

```go
// TestCrossTenantConfigIsolation — in chaos drill suite
func TestCrossTenantConfigIsolation(t *testing.T) {
    tenantA := createTestTenant(t, TenantEngineConfig{APS: {PriceWeight: 0.70}})
    tenantB := createTestTenant(t, TenantEngineConfig{APS: {PriceWeight: 0.20}})

    cfgA, _ := configService.GetTenantConfig(ctx, tenantA.ID)
    cfgB, _ := configService.GetTenantConfig(ctx, tenantB.ID)

    assert.Equal(t, 0.70, cfgA.APS.PriceWeight)
    assert.Equal(t, 0.20, cfgB.APS.PriceWeight)
    assert.NotEqual(t, cfgA.APS.PriceWeight, cfgB.APS.PriceWeight, "config leakage detected")
}
```

---

## 6. Crypto Isolation: KMS Data Keys Per Tenant ★

### 6.1 Envelope Encryption Architecture

Extends `deal-closure v1 §6.5` envelope encryption to all tenant data at rest.

```
Tenant Data (plaintext)
    │ encrypt with DEK (AES-256-GCM)
    ▼
Encrypted Data (stored in DB/S3)

DEK (Data Encryption Key, 256-bit random)
    │ encrypt with KMS CMK (Customer Master Key)
    ▼
Encrypted DEK (stored in tenants.kms_encrypted_dek)
```

Each tenant has:
- Dedicated CMK ARN: `arn:aws:kms:eu-west-1:ACCOUNT:key/<tenant-kms-id>`
- Encrypted DEK stored in `tenants.kms_encrypted_dek` (binary column)
- DEK rotated annually or on-demand (GDPR Art. 17 erasure compliance)

```sql
-- migration 017 ★
ALTER TABLE tenants
    ADD COLUMN kms_key_arn     VARCHAR(256),
    ADD COLUMN kms_encrypted_dek BYTEA,
    ADD COLUMN kms_key_created_at TIMESTAMPTZ DEFAULT NOW(),
    ADD COLUMN kms_key_rotated_at TIMESTAMPTZ;
```

### 6.2 Key Provisioning at Tenant Onboarding

```go
func (s *TenantService) ProvisionKMSKey(ctx context.Context, tenantID string) error {
    // 1. Create CMK in KMS
    cmk, err := s.kms.CreateKey(ctx, &kms.CreateKeyInput{
        Description: fmt.Sprintf("REVYX tenant %s data key", tenantID),
        KeyUsage:    "ENCRYPT_DECRYPT",
        Tags: []kms.Tag{
            {Key: "tenant_id", Value: tenantID},
            {Key: "service", Value: "revyx"},
        },
    })
    if err != nil { return fmt.Errorf("create CMK: %w", err) }

    // 2. Generate DEK, encrypt with CMK
    dek := make([]byte, 32)
    if _, err := rand.Read(dek); err != nil { return err }
    
    encDEK, err := s.kms.Encrypt(ctx, &kms.EncryptInput{
        KeyId:     *cmk.KeyMetadata.KeyId,
        Plaintext: dek,
    })
    if err != nil { return fmt.Errorf("encrypt DEK: %w", err) }

    // 3. Store in DB
    _, err = s.db.Exec(ctx, `
        UPDATE tenants SET
            kms_key_arn = $1,
            kms_encrypted_dek = $2,
            kms_key_created_at = NOW()
        WHERE id = $3
    `, *cmk.KeyMetadata.Arn, encDEK.CiphertextBlob, tenantID)
    
    // Wipe DEK from memory immediately
    for i := range dek { dek[i] = 0 }
    return err
}
```

### 6.3 GDPR Art. 17 — Cryptographic Erasure

When tenant requests data deletion, DEK is destroyed at KMS level:

```go
func (s *TenantService) CryptoErase(ctx context.Context, tenantID string) error {
    kmsArn, err := s.getKMSArn(ctx, tenantID)
    if err != nil { return err }

    // Schedule CMK deletion (7-day minimum AWS waiting period)
    _, err = s.kms.ScheduleKeyDeletion(ctx, &kms.ScheduleKeyDeletionInput{
        KeyId:               kmsArn,
        PendingWindowInDays: 7,
    })
    if err != nil { return fmt.Errorf("schedule KMS deletion: %w", err) }

    // Mark tenant as crypto-erased
    _, err = s.db.Exec(ctx, `
        UPDATE tenants SET
            status = 'crypto_erased',
            kms_key_arn = NULL,
            kms_encrypted_dek = NULL
        WHERE id = $1
    `, tenantID)
    
    s.audit.Log(ctx, AuditEvent{
        Type: "crypto_erasure_scheduled", TenantID: tenantID,
        Detail: fmt.Sprintf("CMK scheduled for deletion: %s", kmsArn),
    })
    return err
}
```

---

## 7. Cross-Tenant Query Auditing ★

### 7.1 Rule: No Query Without tenant_id

All production queries involving tenant-scoped tables MUST include `tenant_id` in the WHERE clause. Enforced at two levels:

**Level 1: CI SQL Lint (sqlfluff)**

```yaml
# .sqlfluff
[sqlfluff]
templater = jinja
dialect = postgres

[rules]
rule_REVYX001 = enabled  # custom rule: require tenant_id filter

# scripts/lint_sql.py (custom rule)
# Scans backend/internal/db/queries.go for queries missing tenant_id
```

CI step blocks PRs where SQL queries on tenant tables lack `tenant_id` predicate.

**Level 2: Runtime Audit Middleware**

```go
// QueryAuditMiddleware wraps pgx.Pool
type QueryAuditMiddleware struct {
    inner  pgxpool.Pool
    audit  AuditLogger
    tables []string // tables requiring tenant_id: ["properties", "buyers", "matches", ...]
}

func (m *QueryAuditMiddleware) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
    if m.isTenantTable(sql) && !m.hasTenantFilter(sql, args) {
        tenantID, _ := ctx.Value(ctxKeyTenantID).(string)
        m.audit.Log(ctx, AuditEvent{
            Type:     "cross_tenant_query_attempt",
            TenantID: tenantID,
            Severity: "CRITICAL",
            SQL:      redactSensitiveSQL(sql),
        })
        return pgconn.CommandTag{}, ErrMissingTenantFilter
    }
    return m.inner.Exec(ctx, sql, args...)
}
```

### 7.2 BYPASSRLS Enforcement

No application role may have `BYPASSRLS`. Checked in CI:

```bash
# scripts/check_bypassrls.sh
BYPASS_ROLES=$(psql "$DATABASE_URL" -At -c "
    SELECT rolname FROM pg_roles
    WHERE rolbypassrls = true AND rolname NOT IN ('postgres', 'rds_superuser')
")
if [ -n "$BYPASS_ROLES" ]; then
    echo "ERROR: Non-system roles with BYPASSRLS: $BYPASS_ROLES"
    exit 1
fi
echo "PASS: No unauthorized BYPASSRLS roles"
```

Run in CI and as a weekly scheduled audit job.

### 7.3 Alert Rules

```yaml
# Prometheus alert rules
- alert: CrossTenantQueryAttempt
  expr: increase(revyx_audit_events_total{type="cross_tenant_query_attempt"}[5m]) > 0
  severity: critical
  annotations:
    summary: "Cross-tenant query attempt detected"
    runbook: "docs/runbooks/cross-tenant-isolation-breach.md"

- alert: BYPASSRLSRoleDetected
  expr: revyx_security_bypassrls_roles_count > 0
  severity: critical
```

---

## 8. Audit Checkpoint — S6 Multi-Tenant Isolation ★

**Architect:** Partial index per tenant is the correct trade-off at current scale. Schema-per-tenant deferred to post-MVP if tenant count exceeds 500. EmbeddingRouter cleanly separates routing from business logic. Config isolation via RLS + cache key scoping is sound. ✅

**Security:** KMS envelope encryption per tenant is production-grade. Crypto erasure via KMS key deletion satisfies GDPR Art. 17 without bulk data delete. BYPASSRLS CI check is essential — approved. Cross-tenant audit events at CRITICAL severity with immediate alert is correct. ✅

**DBA:** Partial HNSW indexes per tenant: postgres planner uses them correctly when WHERE clause matches. RLS + `SET LOCAL app.current_tenant_id` pattern is correct (LOCAL scope prevents session leakage). `aps_weights_sum = 1.0` constraint prevents misconfiguration. ✅

**Compliance:** Per-tenant KMS keys satisfy data sovereignty requirements. Crypto erasure documented for GDPR Art. 17. Embedding logs retention policy needed — recommend 90 days billing data, 30 days operational. ✅

**Product:** `local_only` mode required for enterprise tenants (banking/insurance sector buyers). Cost dashboard API sufficient for billing integration. ✅

**Audit Lead:** CRITICAL gating item: BYPASSRLS CI check must pass in staging before production deploy. All other items OPEN (implementation pending). ✅

---

## 9. Open Items / Gating

| Item | Owner | Deadline | Status | Blocking? |
|---|---|---|---|---|
| KMS provisioning IAM roles in production AWS | Infra | 2026-05-15 | OPEN | YES |
| BYPASSRLS CI check passing on staging | DBA | 2026-05-15 | OPEN | YES |
| Per-tenant index creation load test (100 tenants) | QA | 2026-05-20 | OPEN | NO |
| Billing attribution integration with payment system | Product | 2026-06-01 | OPEN | NO |
| `local_only` inference node provisioned in staging | ML Infra | 2026-05-20 | OPEN | NO |
| Embedding usage log retention policy documented | Compliance | 2026-05-15 | OPEN | NO |

---

*End of TECH_SPEC_REVYX_multitenant-ai-isolation_v1.0.0.md*
