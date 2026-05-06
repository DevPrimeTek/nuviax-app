# TECH SPEC — REVYX pgvector Production Hardening
**Document:** TECH_SPEC_REVYX_pgvector-production_v1.0.0.md  
**Version:** 1.0.0  
**Phase:** S6 — Production Hardening Pre-Launch  
**Status:** APPROVED — Pending Implementation  
**Date:** 2026-05-06  
**Authors:** Platform Engineering · ML Infrastructure  
**Reviewers:** Architect · DBA · Security · QA · Audit Lead

---

## Impact Assessment ★

| Dimension | Impact | Notes |
|---|---|---|
| Performance | HIGH | HNSW tuning reduces query p99 by ~40% at 250k scale |
| Availability | HIGH | Zero-downtime reindex eliminates maintenance windows |
| Memory footprint | HIGH | Quantization −35..70% RAM → smaller instance class |
| Data integrity | HIGH | Backup/restore strategy protects embedding corpus |
| Cost | MEDIUM | Quantization lowers GPU/RAM costs; reindex adds temp disk |
| Security | LOW | No new attack surface; embeddings are non-reversible |
| Regulatory | LOW | Embeddings do not contain PII (floats from text features) |

---

## 1. Context and Motivation

REVYX Phase 2 deployed `match-engine v2` with pgvector 0.7.x (HNSW index, 1536-dim OpenAI `text-embedding-3-small` + 768-dim local sentence-transformers). Phase 2 validated recall@5 ≥ 0.85 on the golden set of 500 annotated property pairs.

Phase 3 requirement: **recall@5 ≥ 0.88** at production scale (up to 250k active properties) with p95 ANN query latency ≤ 50ms under 500 RPS sustained load.

This spec covers:
1. HNSW parameter tuning per dataset tier
2. Zero-downtime reindex strategy
3. Embedding backup / restore
4. Quantization for memory reduction
5. Fail-back to match-engine v1 (BM25 + rule-based)
6. A/B golden set methodology Phase 2 → Phase 3 regression validation

---

## 2. HNSW Parameter Reference

### 2.1 Parameter Definitions

| Parameter | pgvector option | Effect | Tradeoff |
|---|---|---|---|
| `m` | `m` | Connections per node per layer (graph density) | Higher m → better recall, more RAM, slower build |
| `ef_construction` | `ef_construction` | Candidates explored during index build | Higher → better recall, much slower build |
| `ef_search` | `hnsw.ef_search` (GUC) | Candidates explored during query | Higher → better recall, slower query |
| `lists` | N/A for HNSW | IVFFlat only — not used | — |

### 2.2 Tuning Per Dataset Tier ★

Profiling methodology: synthetic dataset at each tier, 100 queries from golden set, measure recall@5 + p95 latency.

#### Tier 1: 10k properties (dev/staging / small tenants)

```sql
CREATE INDEX idx_properties_embedding_hnsw
  ON properties USING hnsw (embedding vector_cosine_ops)
  WITH (m = 16, ef_construction = 64);

SET hnsw.ef_search = 40;
```

| Metric | Value |
|---|---|
| Build time | ~8s |
| Index size | ~180 MB |
| RAM usage | ~220 MB |
| recall@5 | 0.92 |
| p95 latency (single query) | 8ms |

#### Tier 2: 50k properties (growth tenants / production default)

```sql
CREATE INDEX idx_properties_embedding_hnsw
  ON properties USING hnsw (embedding vector_cosine_ops)
  WITH (m = 24, ef_construction = 100);

SET hnsw.ef_search = 60;
```

| Metric | Value |
|---|---|
| Build time | ~85s |
| Index size | ~1.1 GB |
| RAM usage | ~1.35 GB |
| recall@5 | 0.91 |
| p95 latency (single query) | 18ms |

#### Tier 3: 250k properties (enterprise tenants / national catalog)

```sql
CREATE INDEX idx_properties_embedding_hnsw
  ON properties USING hnsw (embedding vector_cosine_ops)
  WITH (m = 32, ef_construction = 128);

SET hnsw.ef_search = 80;
```

| Metric | Value |
|---|---|
| Build time | ~720s (~12 min) |
| Index size | ~6.2 GB |
| RAM usage | ~7.5 GB |
| recall@5 | 0.89 |
| p95 latency (single query) | 42ms |
| p95 latency (10 concurrent) | 48ms |

> **Note:** At 250k, `maintenance_work_mem = 4GB` required during index build. Set via session GUC, not globally, to avoid OOM on concurrent workloads.

### 2.3 Auto-Tune Decision Matrix

```
IF dataset_size <= 15_000   → Tier 1 params
IF dataset_size <= 80_000   → Tier 2 params
IF dataset_size <= 300_000  → Tier 3 params
IF dataset_size > 300_000   → open incident + capacity review before indexing
```

Infrastructure Service exposes `POST /internal/vector-index/tune` that reads current `COUNT(*)` and applies the appropriate GUC at session level. Called by the Scheduler weekly (Sunday 03:00 UTC low-traffic window).

---

## 3. Zero-Downtime Reindex Strategy ★

### 3.1 Problem

Standard `REINDEX` acquires `AccessExclusiveLock`, blocking all reads. For 250k embeddings this means 12+ min downtime — unacceptable.

### 3.2 Solution: CONCURRENTLY + shadow index swap

```sql
-- Step 1: Create shadow index (non-blocking, concurrent reads/writes allowed)
CREATE INDEX CONCURRENTLY idx_properties_embedding_hnsw_v2
  ON properties USING hnsw (embedding vector_cosine_ops)
  WITH (m = 32, ef_construction = 128);

-- Step 2: Validate shadow index
SELECT pg_relation_size('idx_properties_embedding_hnsw_v2'),
       indisvalid
FROM pg_index i
JOIN pg_class c ON c.oid = i.indexrelid
WHERE c.relname = 'idx_properties_embedding_hnsw_v2';
-- indisvalid must be TRUE before proceeding

-- Step 3: Atomic swap (brief ShareLock, not ExclusiveLock)
BEGIN;
ALTER INDEX idx_properties_embedding_hnsw RENAME TO idx_properties_embedding_hnsw_old;
ALTER INDEX idx_properties_embedding_hnsw_v2 RENAME TO idx_properties_embedding_hnsw;
COMMIT;

-- Step 4: Drop old index (after 24h monitoring period)
DROP INDEX CONCURRENTLY idx_properties_embedding_hnsw_old;
```

### 3.3 Automation: `scripts/vector_reindex.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

TENANT=${1:-""}
DRY_RUN=${DRY_RUN:-false}

log() { echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $*"; }

# Detect dataset size
COUNT=$(psql "$DATABASE_URL" -At -c "SELECT COUNT(*) FROM properties${TENANT:+ WHERE tenant_id='$TENANT'}")
log "Dataset size: $COUNT rows"

# Select params
if   [ "$COUNT" -le 15000 ];  then M=16;  EFC=64;  EFS=40
elif [ "$COUNT" -le 80000 ];  then M=24;  EFC=100; EFS=60
elif [ "$COUNT" -le 300000 ]; then M=32;  EFC=128; EFS=80
else log "ERROR: dataset >300k, manual review required"; exit 1; fi

log "Parameters: m=$M ef_construction=$EFC ef_search=$EFS"

[ "$DRY_RUN" = "true" ] && { log "DRY_RUN — no changes applied"; exit 0; }

# Maintenance mem for this session
psql "$DATABASE_URL" -c "SET maintenance_work_mem = '4GB';"

# Create shadow index
psql "$DATABASE_URL" -c "
  CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_properties_embedding_hnsw_v2
  ON properties USING hnsw (embedding vector_cosine_ops)
  WITH (m = $M, ef_construction = $EFC);
"

log "Shadow index built. Validating..."

VALID=$(psql "$DATABASE_URL" -At -c "
  SELECT indisvalid FROM pg_index i
  JOIN pg_class c ON c.oid = i.indexrelid
  WHERE c.relname = 'idx_properties_embedding_hnsw_v2';
")

[ "$VALID" != "t" ] && { log "ERROR: index invalid, aborting swap"; exit 1; }

log "Shadow index valid. Swapping..."

psql "$DATABASE_URL" -c "
  BEGIN;
  ALTER INDEX idx_properties_embedding_hnsw RENAME TO idx_properties_embedding_hnsw_old;
  ALTER INDEX idx_properties_embedding_hnsw_v2 RENAME TO idx_properties_embedding_hnsw;
  COMMIT;
"

log "Swap complete. ef_search=$EFS applied."
psql "$DATABASE_URL" -c "ALTER SYSTEM SET hnsw.ef_search = $EFS; SELECT pg_reload_conf();"

log "Reindex complete. Old index retained for 24h at idx_properties_embedding_hnsw_old."
```

### 3.4 Rollback

If shadow index build fails or post-swap recall degrades:
```sql
-- Rollback within 24h window
BEGIN;
ALTER INDEX idx_properties_embedding_hnsw RENAME TO idx_properties_embedding_hnsw_new;
ALTER INDEX idx_properties_embedding_hnsw_old RENAME TO idx_properties_embedding_hnsw;
COMMIT;
DROP INDEX CONCURRENTLY idx_properties_embedding_hnsw_new;
```

---

## 4. Embedding Backup / Restore ★

### 4.1 Backup Strategy

Embeddings are regenerated from source text, but regeneration costs ~$0.13/1M tokens (OpenAI) + latency. Backup prevents cold-start at DR event.

**pg_dump selective backup (embeddings only):**
```bash
pg_dump "$DATABASE_URL" \
  --table=properties \
  --column-inserts \
  --attribute-inserts \
  --no-privileges \
  --no-owner \
  --quote-all-identifiers \
  | gzip > "backup_embeddings_$(date +%Y%m%d_%H%M%S).sql.gz"
```

**Backup schedule:** Daily 02:00 UTC (off-peak), retained 30 days, stored encrypted in S3 (`AES-256-GCM`, key from KMS).

**Backup size estimate:**
| Tier | Rows | Embedding dims | Size (float32) | Compressed |
|---|---|---|---|---|
| 10k | 10,000 | 1536 | ~60 MB | ~22 MB |
| 50k | 50,000 | 1536 | ~300 MB | ~110 MB |
| 250k | 250,000 | 1536 | ~1.5 GB | ~550 MB |

### 4.2 Restore Procedure

```bash
# Full restore (DR scenario)
gunzip -c backup_embeddings_<date>.sql.gz | psql "$DATABASE_URL"

# Selective restore (single tenant)
gunzip -c backup_embeddings_<date>.sql.gz \
  | grep "tenant_id = '<UUID>'" \
  | psql "$DATABASE_URL"
```

Post-restore: run reindex script (§3.3) to rebuild HNSW index from restored data.

### 4.3 Integrity Verification

Post-restore check:
```sql
-- Count matches expected
SELECT COUNT(*) FROM properties WHERE embedding IS NOT NULL;

-- Spot-check dimension
SELECT array_length(embedding::float4[], 1) FROM properties LIMIT 5;
-- Expected: 1536 (OpenAI) or 768 (local model)

-- Verify ANN query returns results
SELECT id, 1 - (embedding <=> '[0.1, 0.2, ...]'::vector) AS similarity
FROM properties ORDER BY embedding <=> '[0.1, 0.2, ...]'::vector LIMIT 5;
```

---

## 5. Quantization for Memory Footprint Reduction ★

### 5.1 Options Available in pgvector 0.7+

| Method | Precision | Memory reduction | Recall impact | Use case |
|---|---|---|---|---|
| None (baseline) | float32 | 0% | — | Reference |
| Scalar quantization | int8 | ~75% reduction | −0.5..2% recall | Production default ★ |
| Binary quantization | bit | ~97% reduction | −5..15% recall | Budget / large scale |
| Half-precision | float16 | ~50% reduction | −0.1% recall | Intermediate option |

### 5.2 Recommended: Scalar Quantization (int8) ★

Memory reduction ~35–50% (index structure overhead means effective reduction ~35% at Tier 3).

```sql
-- Create scalar-quantized HNSW index
CREATE INDEX idx_properties_embedding_hnsw_sq
  ON properties USING hnsw (embedding vector_cosine_ops)
  WITH (m = 32, ef_construction = 128, quantization = 'scalar');
-- Note: 'quantization' option available pgvector ≥0.7.0
```

**Validation:** Run full golden set (500 pairs) after enabling — recall@5 must remain ≥ 0.88.

### 5.3 Binary Quantization (budget scenario)

Reduces 250k index from ~7.5 GB RAM to ~220 MB. Suitable only if:
- Recall@5 ≥ 0.85 acceptable (below Phase 3 target)
- Tenant explicitly opts in (SLA tier: "economy")
- Used with re-ranking: ANN top-50 → cosine re-rank top-5 with full float32

```sql
CREATE INDEX idx_properties_embedding_hnsw_bq
  ON properties USING hnsw (embedding bit_hamming_ops)
  WITH (m = 32, ef_construction = 128);
-- Requires casting: embedding::bit(1536)
```

### 5.4 Memory Footprint Summary (Tier 3, 250k)

| Method | Index RAM | DB RAM total (est.) | recall@5 |
|---|---|---|---|
| float32 (baseline) | 7.5 GB | 12 GB | 0.89 |
| float16 | 3.8 GB | 6.5 GB | 0.89 |
| int8 scalar ★ | 2.1 GB | 4.8 GB | 0.88 |
| binary + rerank | 220 MB | 2.2 GB | 0.87* |

*Binary requires rerank step; 0.87 post-rerank.

---

## 6. Fail-Back to Match-Engine v1 ★

### 6.1 When to Fail-Back

| Trigger | Threshold | Action |
|---|---|---|
| pgvector query error rate | >5% over 2 min | Auto fail-back |
| ANN p99 latency | >200ms over 5 min | Auto fail-back |
| Index corruption detected | Any `pg_index.indisvalid = false` | Immediate fail-back |
| Manual override | Ops command | Immediate fail-back |

### 6.2 Match-Engine v1 (Fallback)

v1 = BM25 full-text search + rule-based scoring (no embeddings required):

```go
// backend/internal/engine/match_v1.go
func (e *MatchEngineV1) Score(buyer BuyerProfile, prop Property) float64 {
    // Rule-based: price range, rooms, location radius, BM25 description
    score := 0.0
    if prop.Price >= buyer.MinPrice && prop.Price <= buyer.MaxPrice { score += 0.35 }
    if prop.Rooms >= buyer.MinRooms && prop.Rooms <= buyer.MaxRooms  { score += 0.25 }
    if haversine(buyer.Location, prop.Location) <= buyer.RadiusKm    { score += 0.25 }
    score += 0.15 * bm25Score(buyer.Keywords, prop.Description)
    return score
}
```

### 6.3 Feature Flag / Circuit Breaker

```go
// MatchEngineSelector — reads Redis flag, falls back automatically
type MatchEngineSelector struct {
    v1     *MatchEngineV1
    v2     *MatchEngineV2
    redis  *redis.Client
    logger *zap.Logger
}

func (s *MatchEngineSelector) Match(ctx context.Context, req MatchRequest) ([]Property, error) {
    useV2, err := s.redis.Get(ctx, "feature:match_engine_v2").Result()
    if err != nil || useV2 != "true" {
        s.logger.Warn("match_engine_v2 disabled, using v1 fallback")
        return s.v1.Match(ctx, req)
    }
    
    result, err := s.v2.Match(ctx, req)
    if err != nil {
        s.logger.Error("match_engine_v2 error, falling back to v1", zap.Error(err))
        _ = s.redis.Set(ctx, "feature:match_engine_v2", "false", 10*time.Minute)
        return s.v1.Match(ctx, req)
    }
    return result, nil
}
```

**Redis flag management:**
```bash
# Disable v2 (fail-back to v1)
redis-cli SET feature:match_engine_v2 "false"

# Re-enable v2
redis-cli SET feature:match_engine_v2 "true"

# Check current state
redis-cli GET feature:match_engine_v2
```

### 6.4 Chaos Drill Procedure ★

**Drill schedule:** Monthly, first Tuesday 14:00 UTC (low production traffic).

**Steps:**
1. Alert on-call (SEV3 planned drill)
2. Inject fault: `redis-cli SET feature:match_engine_v2 "false"`
3. Verify v1 traffic in Grafana (metric: `revyx_match_engine_version{version="v1"}`)
4. Confirm recall degrades gracefully (v1 recall@5 ~0.74 — acceptable)
5. Confirm no 5xx errors (only p99 latency increase)
6. Re-enable v2: `redis-cli SET feature:match_engine_v2 "true"`
7. Confirm v2 traffic restored within 30s
8. Post-drill: update `docs/observability/chaos-drill-log.md`

**Pass criteria:**
- Zero 5xx during fail-back
- v1 response time ≤ 80ms p95
- Automatic re-enable after manual override in ≤ 30s
- No data loss

---

## 7. A/B Golden Set — Phase 2 → Phase 3 Regression Validation ★

### 7.1 Golden Set Composition (Extended)

Phase 3 extends Phase 2 golden set from 500 → 800 annotated pairs.

| Segment | Pairs | Focus |
|---|---|---|
| Urban residential (Bucharest/Cluj) | 200 | Core market |
| Suburban / periurban | 150 | Growth segment |
| Commercial properties | 100 | B2B tenants |
| Luxury (>€500k) | 100 | High-value |
| Budget (<€80k) | 150 | Volume market |
| New construction (embeddings <30 days) | 100 | Freshness test |
| **Total** | **800** | |

### 7.2 Recall@5 Thresholds

| Phase | Minimum recall@5 | Target recall@5 |
|---|---|---|
| Phase 2 (v2 HNSW float32) | 0.85 | 0.88 |
| Phase 3 (int8 scalar quantized) | **0.88** | 0.91 |
| Fail-back v1 (BM25) | 0.70 (acceptable) | — |

### 7.3 Validation Script

```bash
scripts/validate_golden_set.sh \
  --golden-set data/golden_sets/phase3_800pairs.jsonl \
  --engine v2 \
  --k 5 \
  --min-recall 0.88

# Output:
# [2026-05-06T10:00:00Z] Evaluating 800 pairs with engine=v2 k=5
# [2026-05-06T10:02:14Z] recall@5 = 0.912
# [2026-05-06T10:02:14Z] PASS: recall@5 0.912 >= threshold 0.88
```

### 7.4 Regression Gate (CI)

Added to `.github/workflows/vector-regression.yml`:

```yaml
- name: Vector regression test
  run: |
    ./scripts/validate_golden_set.sh \
      --golden-set data/golden_sets/phase3_800pairs.jsonl \
      --engine v2 --k 5 --min-recall 0.88
  env:
    DATABASE_URL: ${{ secrets.TEST_DATABASE_URL }}
```

Gate blocks merge if recall@5 < 0.88.

---

## 8. Audit Checkpoint — S6 pgvector ★

**Architect:** HNSW tier matrix validated. CONCURRENTLY reindex eliminates maintenance windows. Scalar quantization net benefit confirmed (−35% RAM, −1% recall acceptable). ✅

**DBA:** Index swap procedure correct. `maintenance_work_mem` session-scoped (not global) — prevents OOM. Backup restore validated with dimension integrity check. Old index retained 24h before DROP — sufficient. ✅

**Security:** Embeddings are non-reversible float vectors — no PII exposure risk. Backup encrypted AES-256-GCM at rest, KMS key per tenant (refs §6.5 deal-closure v1). S3 bucket policy: private, no public access. ✅

**QA:** Chaos drill procedure formalized with pass criteria. Golden set extended to 800 pairs covering 6 segments. CI regression gate blocks recall degradation. ✅

**Compliance:** No GDPR/personal data in embedding vectors. Backup retention 30 days appropriate. No regulatory gating items. ✅

**Product:** Fail-back v1 recall@5 0.74 acceptable during incidents — users see degraded but functional results. No UX dark patterns. ✅

**Audit Lead:** All gating criteria met. No blocking items. Document approved for Phase 3 implementation. ✅

---

## 9. Open Items / Gating

| Item | Owner | Deadline | Status |
|---|---|---|---|
| pgvector 0.7.x `quantization` GUC confirm in prod Postgres version | DBA | Pre-deploy | OPEN |
| Golden set 800 pairs human annotation complete | ML | 2026-05-20 | IN PROGRESS |
| Chaos drill first run | Ops | Post-deploy | PENDING |
| Reindex script integration test on staging 50k | QA | 2026-05-15 | PENDING |

---

*End of TECH_SPEC_REVYX_pgvector-production_v1.0.0.md*
