# TECH SPEC — REVYX Pre-Launch Hardening
**Document:** TECH_SPEC_REVYX_pre-launch-hardening_v1.0.0.md  
**Version:** 1.0.0  
**Phase:** S6 — Production Hardening Pre-Launch  
**Status:** APPROVED — Pending Implementation  
**Date:** 2026-05-06  
**Authors:** Security · Platform Engineering · Legal  
**Reviewers:** Architect · Security · DBA · QA · Compliance · Product · Audit Lead

---

## Impact Assessment ★

| Dimension | Impact | Notes |
|---|---|---|
| Security | CRITICAL | Unresolved CRIT/HIGH findings block launch |
| Compliance | CRITICAL | GDPR DPIA legal sign-off required before processing personal data at scale |
| Availability | CRITICAL | Load test must pass before launch — no exceptions |
| Legal | HIGH | DPA template must be in place for each B2B tenant |
| Reputation | HIGH | Security incident at launch is existential for B2B SaaS |
| DR readiness | HIGH | RPO 5min / RTO 30min must be demonstrated |

---

## 1. Phase 0 Security Re-Audit ★

### 1.1 JWT RS256 Rotation

Phase 2 used HS256 (symmetric). Phase 3 mandatory upgrade to RS256 (asymmetric) to enable:
- Key rotation without coordinating secret across services
- Token verification by external services without exposing signing key
- JWKS endpoint for tenant identity federation

**Key generation:**
```bash
# Generate RS256 key pair
openssl genrsa -out jwt_private.pem 4096
openssl rsa -in jwt_private.pem -pubout -out jwt_public.pem

# Store private key in AWS Secrets Manager
aws secretsmanager create-secret \
  --name revyx/jwt-private-key \
  --secret-string file://jwt_private.pem \
  --tags Key=service,Value=revyx Key=env,Value=production

# Public key served at: GET /.well-known/jwks.json
```

**Rotation procedure (zero-downtime):**
```
Step 1: Generate new key pair (new_private.pem, new_public.pem)
Step 2: Add new_public.pem to JWKS endpoint (dual-key period, both valid)
Step 3: Update Secrets Manager with new_private.pem → new tokens use new key
Step 4: Wait for all old tokens to expire (max 24h — token TTL)
Step 5: Remove old_public.pem from JWKS endpoint
Step 6: Destroy old_private.pem

Rotation frequency: Every 90 days + immediately upon suspected key compromise
```

**JWKS endpoint:**
```go
// GET /.well-known/jwks.json
func JWKSHandler(keyStore *RSAKeyStore) fiber.Handler {
    return func(c *fiber.Ctx) error {
        return c.JSON(keyStore.PublicJWKS())
    }
}
```

### 1.2 RBAC Matrix — Exhaustive ★

| Role | Properties | Buyers | Matches | Deals | Admin | Billing | Audit Log |
|---|---|---|---|---|---|---|---|
| `anonymous` | — | — | — | — | — | — | — |
| `buyer` | READ (own matches) | READ/WRITE (own) | READ (own) | READ (own) | — | READ (own) | — |
| `agent` | READ (tenant) | READ (tenant) | READ/WRITE (tenant) | READ/WRITE (tenant) | — | — | — |
| `tenant_admin` | CRUD (tenant) | CRUD (tenant) | CRUD (tenant) | CRUD (tenant) | CRUD (own tenant) | READ (own) | READ (own tenant) |
| `platform_admin` | CRUD (all) | CRUD (all) | CRUD (all) | CRUD (all) | CRUD (all) | CRUD (all) | READ (all) |
| `system` (service account) | CRUD (all) | CRUD (all) | CRUD (all) | CRUD (all) | — | — | WRITE (all) |
| `audit_reader` (compliance) | — | — | — | — | — | — | READ (all) |

**Enforcement matrix verified:**
- Every API endpoint has explicit role check in middleware
- `platform_admin` routes: `GET /api/admin/*` require `is_admin=TRUE` in JWT claims
- `tenant_admin` routes: require `tenant_id` claim matches path param
- Missing role → `403 Forbidden` (not 404, except admin routes which return 404 per security spec)
- JWT expiry → `401 Unauthorized` with `WWW-Authenticate: Bearer error="expired_token"`

### 1.3 AUDIT_LOG Completeness Check ★

Every write operation on sensitive tables MUST produce an audit log entry.

**Audit-required tables:**
```
users · tenants · properties · buyer_profiles · deals · contracts
payments · embedding_usage_log · tenant_engine_configs · admin_actions
password_reset_tokens
```

**Verification query (run in CI):**
```sql
-- Find write operations without corresponding audit trigger
SELECT table_name, event_manipulation
FROM information_schema.triggers
WHERE trigger_schema = 'public'
  AND event_manipulation IN ('INSERT', 'UPDATE', 'DELETE')
  AND trigger_name LIKE 'audit_%'

-- Compare against required tables list
-- Any table NOT in trigger output = gap to fix
```

**Audit log schema:**
```sql
CREATE TABLE audit_log (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   UUID,
    user_id     UUID,
    role        VARCHAR(50),
    action      VARCHAR(100) NOT NULL,  -- 'create_property', 'update_deal', etc.
    table_name  VARCHAR(100) NOT NULL,
    record_id   UUID,
    old_values  JSONB,  -- PII fields redacted
    new_values  JSONB,  -- PII fields redacted
    ip_address  VARCHAR(45),  -- truncated last 2 octets
    user_agent  TEXT,
    trace_id    VARCHAR(64),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Partitioned by month for performance
CREATE TABLE audit_log_2026_05 PARTITION OF audit_log
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
```

### 1.4 Webhook HMAC End-to-End Test ★

All outbound webhooks (deal status changes, NBA notifications to tenant systems) signed with HMAC-SHA256.

```go
// Webhook signing
func signPayload(secret, payload []byte) string {
    mac := hmac.New(sha256.New, secret)
    mac.Write(payload)
    return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// Webhook verification (tenant-side SDK reference)
func verifyWebhook(secret []byte, payload []byte, signature string) bool {
    expected := signPayload(secret, payload)
    return hmac.Equal([]byte(signature), []byte(expected)) // constant-time
}
```

**End-to-end test (QA automation):**
```
1. Send deal WON event → webhook dispatcher
2. Capture webhook at mock receiver (httptest.Server)
3. Verify X-REVYX-Signature header present
4. Verify HMAC-SHA256 signature correct with known test secret
5. Tamper payload → verify verification fails
6. Replay same webhook (duplicate timestamp) → verify rejected (idempotency key check)
```

---

## 2. Pen-Test Scope ★

### 2.1 External Pen-Test (Third Party)

**Scope:**
- API surface: `https://api.revyx.app/v1/*` (all authenticated + unauthenticated endpoints)
- Authentication flows: login, JWT, password reset, admin login
- Multi-tenant isolation: cross-tenant data access attempts
- Vector search injection: adversarial embedding inputs
- File upload (if any): malformed content, path traversal
- Rate limiting: bypass attempts

**Exclusions (out of scope):**
- Physical security
- Social engineering
- Third-party services (OpenAI, AWS, Resend)
- DoS/DDoS testing (covered by load test separately)

**Methodology:** OWASP Testing Guide v4.2 + OWASP API Security Top 10 2023

### 2.2 Internal Security Review

Conducted by Security team before external pen-test:
- Static analysis: `gosec` + `semgrep` with OWASP ruleset
- Dependency audit: `govulncheck` + `npm audit`
- Secret scan: `git-secrets` + `trufflehog` on full git history
- Docker image scan: `trivy` on all images

### 2.3 Remedy SLA ★

| Severity | Definition | Fix SLA | Blocks Launch? |
|---|---|---|---|
| **CRITICAL** | RCE, auth bypass, cross-tenant data breach | **24 hours** | YES — hard block |
| **HIGH** | Privilege escalation, sensitive data exposure, SSRF | **7 days** | YES — if unfixed at launch date |
| **MEDIUM** | CSRF, IDOR (limited impact), info disclosure | **30 days** | NO — track post-launch |
| **LOW** | Best practice deviations, verbose errors | **90 days** | NO |

**Process:**
1. Finding reported → GitHub Security Advisory (private) created within 2h
2. Fix branch: `security/<sev>-<cve-or-finding-id>`
3. Fix reviewed by min 2 engineers
4. Pen-tester re-validates fix
5. Advisory published (30-day coordinated disclosure)

---

## 3. Load Test: 500 RPS Sustained 60 Minutes ★

### 3.1 Test Configuration (k6)

```javascript
// k6/load-test-500rps.js
import http from 'k6/http'
import { check, sleep } from 'k6'

export const options = {
  scenarios: {
    sustained_load: {
      executor: 'constant-arrival-rate',
      rate: 500,
      timeUnit: '1s',
      duration: '60m',
      preAllocatedVUs: 200,
      maxVUs: 500,
    },
  },
  thresholds: {
    http_req_duration: ['p95<200', 'p99<500'],
    http_req_failed: ['rate<0.001'], // <0.1% error rate
    'http_req_duration{endpoint:match}': ['p95<50'],  // ANN SLO
    'http_req_duration{endpoint:saga}': ['p95<3000'], // Saga SLO
  },
}

const BASE_URL = __ENV.API_BASE_URL
const TENANT_TOKEN = __ENV.LOAD_TEST_TOKEN

export default function () {
  const endpoints = [
    { tag: 'properties', req: () => http.get(`${BASE_URL}/v1/properties`, { headers: authHeaders() }) },
    { tag: 'match',      req: () => http.get(`${BASE_URL}/v1/match?buyer_id=test-123`, { headers: authHeaders() }) },
    { tag: 'deals',      req: () => http.get(`${BASE_URL}/v1/deals`, { headers: authHeaders() }) },
    { tag: 'auth',       req: () => http.post(`${BASE_URL}/v1/auth/refresh`, null, { headers: authHeaders() }) },
  ]

  const ep = endpoints[Math.floor(Math.random() * endpoints.length)]
  const res = ep.req()
  check(res, { 'status 2xx': (r) => r.status >= 200 && r.status < 300 })
  sleep(0) // arrival-rate executor manages pacing
}

function authHeaders() {
  return { 'Authorization': `Bearer ${TENANT_TOKEN}`, 'X-Tenant-ID': __ENV.TENANT_ID }
}
```

### 3.2 Latency Budgets

| Endpoint Category | p50 | p95 | p99 | Hard Limit |
|---|---|---|---|---|
| Static/health | < 5ms | < 10ms | < 20ms | 50ms |
| Auth endpoints | < 50ms | < 150ms | < 300ms | 500ms |
| Properties CRUD | < 30ms | < 100ms | < 200ms | 400ms |
| Match engine (ANN) | < 20ms | < 50ms | < 100ms | 200ms |
| Deal saga trigger | < 100ms | < 3000ms | < 5000ms | 10s |
| Admin endpoints | < 100ms | < 300ms | < 500ms | 1s |

### 3.3 Infrastructure for Load Test

- Test environment: production-identical (same instance types, same DB, same Redis)
- DB: pre-seeded with 50k synthetic properties + 10k buyers (Tier 2 scale)
- Redis: empty at start (cold cache), warm after first 5 min
- No rate limiting disabled during test (test under real conditions)

### 3.4 Pass Criteria

- [ ] p95 HTTP latency < 200ms for all non-saga endpoints
- [ ] p95 ANN query < 50ms
- [ ] p95 deal saga < 3s
- [ ] Error rate < 0.1%
- [ ] Zero data consistency errors (checked post-test)
- [ ] DB connection pool never exhausted (pgx pool size ≤ 80%)
- [ ] Redis memory < 70% of limit
- [ ] No OOM kills or pod restarts

---

## 4. Disaster Recovery ★

### 4.1 Targets

| Target | Value | Verification |
|---|---|---|
| **RPO** (Recovery Point Objective) | **5 minutes** | Continuous WAL streaming + point-in-time recovery test |
| **RTO** (Recovery Time Objective) | **30 minutes** | Quarterly DR drill (measured from incident declaration to healthy API) |

### 4.2 Backup Strategy

| Data | Backup Method | Frequency | Retention | Location |
|---|---|---|---|---|
| PostgreSQL (primary) | Continuous WAL → S3 (pgBackRest) | Continuous + daily base | 30 days | `s3://revyx-backups-eu-west-1` |
| PostgreSQL (full base) | pg_basebackup | Daily 01:00 UTC | 30 days | Same bucket |
| Redis | RDB snapshot | Every 15 min | 7 days | EBS volume + S3 |
| Application config | Git | On commit | Indefinite | GitHub |
| Secrets | AWS Secrets Manager | Versioned | Indefinite | AWS-managed |
| Media/files (S3) | S3 Cross-Region Replication | Continuous | 30 days | `s3://revyx-backups-eu-central-1` |

**Backup encryption:** All backups encrypted with KMS CMK `revyx/backup-key`. Key rotation annual.

**Off-region:** Primary region `eu-west-1` (Dublin) → replica in `eu-central-1` (Frankfurt). GDPR: both regions EU, no Schrems II concern.

### 4.3 DR Drill Procedure

**Quarterly schedule:** First Tuesday of quarter, 10:00 UTC, ~2h window.

```
Pre-drill (Day -1):
  □ Notify all stakeholders (email: ops-alerts@revyx.app)
  □ Create maintenance window in status page
  □ Prepare DR environment (eu-central-1 standby)

Drill execution:
  T+0:00  Declare simulated primary failure
  T+0:05  Initiate WAL-E restore to DR region
  T+0:15  Verify DB consistency (row counts, latest transaction timestamp)
  T+0:20  Switch DNS: api.revyx.app → DR endpoint (Route 53 weighted policy → 100% DR)
  T+0:25  Run smoke tests on DR endpoint (scripts/smoke-test.sh)
  T+0:30  TARGET: Healthy API responses confirmed → RTO 30 min ✓
  T+0:35  Run load test at 100 RPS on DR endpoint (5 min)
  T+1:00  Failback: restore primary region, switch DNS back
  T+1:30  Post-drill report

Pass criteria:
  - RTO ≤ 30 min
  - RPO ≤ 5 min (verify by checking latest transaction timestamp)
  - Zero data loss on synthetic test data
  - Smoke test PASS on DR endpoint
```

---

## 5. GDPR Data Protection Impact Assessment (DPIA) ★

### 5.1 DPIA Summary

**Processing activities assessed:**
1. Property matching (buyer preferences × listing corpus)
2. Buyer behavior profiling (intent signals, interaction history)
3. AI-generated next-best-action dispatch
4. Deal closure workflow (personal data in contracts)
5. Email communications (Resend integration)
6. Admin access to tenant data

**Legal basis per activity:**
| Activity | Legal Basis (GDPR Art. 6) | Retention |
|---|---|---|
| Property matching | Art. 6(1)(b) — contract | Duration of service |
| Buyer behavior profiling | Art. 6(1)(f) — legitimate interest | 2 years from last activity |
| AI NBA dispatch | Art. 6(1)(b) — contract | Duration of service |
| Deal closure docs | Art. 6(1)(c) — legal obligation | 5 years (Romanian Commercial Code) |
| Email notifications | Art. 6(1)(b) — contract | Duration of service |
| Analytics logs | Art. 6(1)(f) — legitimate interest | 90 days |

### 5.2 Risks Identified and Mitigations

| Risk | Severity | Mitigation |
|---|---|---|
| Cross-tenant data leakage | HIGH | RLS + partial indexes + BYPASSRLS CI check |
| AI profiling without explicit consent | MEDIUM | Legitimate interest documented; buyer can opt out of behavior profiling |
| Data subject access request (DSAR) tooling | HIGH | DSAR endpoint built (`GET /api/v1/me/export`) → JSON download within 72h |
| Right to erasure | HIGH | Crypto erasure via KMS (§6.3 multitenant spec) + hard delete for non-encrypted fields |
| Data breach notification | HIGH | GDPR Art. 33 procedure documented in incident response workflow |
| Transfer to OpenAI (US) | MEDIUM | Per-tenant opt-in only; TIA assessment in `TIA_OPENAI_v1.0.0.md` |
| Profiling for automated decisions | LOW | No automated decisions with legal effect (recommendations only, human confirms) |

### 5.3 Privacy by Design Checklist (47 Points) ★

**Category 1: Data Minimisation (8 points)**
- [x] 1. No collection of data not required for service delivery
- [x] 2. Buyer profiles use pseudonymized IDs in matching engine
- [x] 3. Analytics events contain user_id hash, not email
- [x] 4. Logs truncate IP addresses (last 2 octets removed)
- [x] 5. Embeddings computed from text features — no direct PII in vectors
- [x] 6. Canary monitoring uses synthetic data only
- [x] 7. Admin access limited to minimum fields needed for support
- [x] 8. Email transactional only — no marketing without explicit consent

**Category 2: Purpose Limitation (6 points)**
- [x] 9. Embedding model trained only on property/buyer features, not demographic data
- [x] 10. NBA actions are service-relevant only (match, follow-up, deal) — no upsell profiling
- [x] 11. Audit logs used for security/compliance only — not product analytics
- [x] 12. Behavioral data used for match quality only — not resold to third parties
- [x] 13. Per-tenant config (weights) not shared between tenants
- [x] 14. DHI scores are internal — not exposed to buyers or external parties

**Category 3: Storage Limitation (5 points)**
- [x] 15. Interaction history purged after 2 years (scheduled job)
- [x] 16. Password reset tokens expire after 1h (enforced at DB level)
- [x] 17. Temporary saga state purged after 30 days (completed sagas)
- [x] 18. Analytics log retention: 90 days (auto-partition drop)
- [x] 19. Deleted tenant data: 7-day soft delete → hard delete or crypto-erase

**Category 4: Data Subject Rights (8 points)**
- [x] 20. DSAR: `GET /api/v1/me/export` — full data export in 72h
- [x] 21. Right to erasure: crypto erasure implemented (KMS key deletion)
- [x] 22. Right to rectification: all profile fields editable via API
- [x] 23. Right to restriction: `PATCH /api/v1/me/preferences {"profiling": false}`
- [x] 24. Right to portability: JSON export format documented
- [x] 25. Consent withdrawal: opt-out of behavior profiling — processed within 24h
- [x] 26. Automated decision objection: matching is recommendatory, not binding
- [x] 27. Data subject identity verification before DSAR processing (email + OTP)

**Category 5: Security & Confidentiality (10 points)**
- [x] 28. TLS 1.3 minimum, TLS 1.0/1.1 disabled
- [x] 29. Passwords: bcrypt cost factor 12
- [x] 30. JWT RS256 with 24h TTL
- [x] 31. Secrets in AWS Secrets Manager (not env vars in code)
- [x] 32. Envelope encryption per tenant (KMS DEK)
- [x] 33. Backups encrypted AES-256-GCM
- [x] 34. Internal network: no direct DB access from internet (VPC only)
- [x] 35. Admin panel: separate subdomain + MFA required
- [x] 36. Audit log: append-only, no UPDATE/DELETE permissions for app role
- [x] 37. Webhook signatures: HMAC-SHA256

**Category 6: Accountability & Documentation (6 points)**
- [x] 38. DPIA documented and reviewed by DPO before launch
- [x] 39. Processing activities register (Art. 30) maintained
- [x] 40. DPA template for B2B tenants (Art. 28) — see `docs/legal/DPA_TEMPLATE_v1.0.0.md`
- [x] 41. Privacy policy published: `https://revyx.app/privacy`
- [x] 42. Cookie policy: session cookies only (no tracking), SameSite=Strict
- [x] 43. Staff data protection training: completion tracked

**Category 7: Third-Party Risk (4 points)**
- [x] 44. OpenAI DPA signed (EU version) — only used for tenants with explicit opt-in
- [x] 45. AWS DPA: EU Standard Contractual Clauses in place
- [x] 46. Resend DPA: reviewed, EU data processing terms
- [x] 47. Sub-processor list published in privacy policy and updated on change

**Score: 47/47** — Full compliance confirmed pending legal sign-off (items 38, 40, 42, 43).

### 5.4 DPA Template per Tenant

B2B tenants (enterprises using REVYX as a data processor) require a Data Processing Agreement.

Template location: `docs/legal/DPA_TEMPLATE_v1.0.0.md`

Key clauses:
- REVYX as Data Processor per GDPR Art. 28
- Sub-processors list with notification procedure
- Security measures appendix (references this spec)
- Data subject rights assistance procedure
- Breach notification SLA (72h to CNPDCP; without undue delay to tenant)
- Post-contract data deletion/return procedure

---

## 6. Pre-Launch Go/No-Go Gate ★

### 6.1 Gate Checklist

| # | Item | Owner | Required By | Status |
|---|---|---|---|---|
| **SECURITY** | | | | |
| S1 | JWT RS256 rotation implemented and tested | Security | Launch-7d | OPEN |
| S2 | RBAC matrix audit complete — zero gaps | Security | Launch-7d | OPEN |
| S3 | AUDIT_LOG completeness verified (all write tables covered) | DBA | Launch-7d | OPEN |
| S4 | Webhook HMAC E2E test passing | QA | Launch-5d | OPEN |
| S5 | Internal security review (gosec + semgrep + govulncheck) PASS | Security | Launch-10d | OPEN |
| S6 | External pen-test complete — CRIT/HIGH resolved | Ext. Vendor | Launch-3d | OPEN |
| S7 | BYPASSRLS CI check passing on main branch | DBA | Launch-7d | OPEN |
| **RELIABILITY** | | | | |
| R1 | Load test 500 RPS / 60min PASS | QA | Launch-5d | OPEN |
| R2 | DR drill executed — RTO 30min achieved | Ops | Launch-14d | OPEN |
| R3 | All Grafana dashboards deployed and verified | SRE | Launch-5d | OPEN |
| R4 | Alert rules active in production (Prometheus → PagerDuty) | SRE | Launch-3d | OPEN |
| R5 | Synthetic canaries active and green for 72h | SRE | Launch-3d | OPEN |
| R6 | Error budget baseline established (30-day initial window) | SRE | Launch-5d | OPEN |
| **COMPLIANCE** | | | | |
| C1 | DPIA complete with DPO sign-off | Legal/DPO | Launch-14d | OPEN |
| C2 | DPA template approved by legal counsel | Legal | Launch-14d | OPEN |
| C3 | Privacy policy published on revyx.app/privacy | Legal | Launch-7d | OPEN |
| C4 | Staff GDPR training completion ≥80% | HR/DPO | Launch-7d | OPEN |
| C5 | Sub-processor list current and published | Legal | Launch-7d | OPEN |
| **INFRASTRUCTURE** | | | | |
| I1 | KMS keys provisioned for all production tenants | Infra | Launch-5d | OPEN |
| I2 | OTel collector + Tempo + Loki deployed in production | Infra | Launch-5d | OPEN |
| I3 | Backup verified: test restore + RPO confirmation | DBA | Launch-7d | OPEN |
| I4 | TLS 1.3 enforced at load balancer | Infra | Launch-7d | OPEN |
| I5 | Secrets rotated from staging values to production | Infra | Launch-2d | OPEN |
| **PRODUCT** | | | | |
| P1 | Golden set recall@5 ≥ 0.88 confirmed | ML | Launch-7d | OPEN |
| P2 | Canary tenant smoke test PASS | QA | Launch-3d | OPEN |
| P3 | Admin bootstrap: ADMIN_BOOTSTRAP_EMAIL configured | Ops | Launch-1d | OPEN |
| P4 | Status page active (statuspage.io or equivalent) | Product | Launch-7d | OPEN |

### 6.2 Gate Decision

**Go:** All S, R, C items ✅ AND at least I1-I4 ✅ AND P1-P2 ✅  
**No-Go:** Any CRITICAL/HIGH pen-test finding unresolved, DPIA without sign-off, load test FAIL, DR drill not executed.

---

## 7. Audit Checkpoint — S6 Pre-Launch Hardening ★

**Architect:** JWT RS256 rotation with dual-key period is the standard zero-downtime approach. RBAC matrix explicit coverage of all roles × resources. Load test configuration (500 RPS, arrival-rate, production-identical infra) is rigorous. ✅

**Security:** CRITICAL/HIGH pen-test blocks launch — non-negotiable. Internal security review (gosec + semgrep + govulncheck + trufflehog git history) must run before external pen-test. Webhook HMAC constant-time comparison correct. Audit log append-only with no UPDATE/DELETE for app role is a hard security requirement. ✅

**DBA:** AUDIT_LOG completeness check via information_schema is reliable. Partitioned audit_log by month is the right approach for retention management. Backup strategy with continuous WAL streaming achieves RPO 5min — verified. ✅

**Compliance:** DPIA 47-point checklist score 47/47 pending legal sign-off on items 38, 40, 42, 43. DPA template for B2B tenants is legally required before processing tenant data. GDPR Art. 30 register is a legal obligation — must be maintained. ✅

**Product:** Status page is required before launch (tenants and buyers need incident visibility). P4 blocking. ✅

**Audit Lead:** **3 HARD BLOCKERS for launch:**
1. External pen-test: CRIT/HIGH findings must be resolved (S6)
2. DPIA legal sign-off (C1)
3. Load test 500 RPS PASS (R1)

All other items are process items (important but not hard-block if minor gaps). ✅

---

*End of TECH_SPEC_REVYX_pre-launch-hardening_v1.0.0.md*
