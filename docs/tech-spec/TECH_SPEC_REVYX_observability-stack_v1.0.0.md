# TECH SPEC — REVYX Observability Stack
**Document:** TECH_SPEC_REVYX_observability-stack_v1.0.0.md  
**Version:** 1.0.0  
**Phase:** S6 — Production Hardening Pre-Launch  
**Status:** APPROVED — Pending Implementation  
**Date:** 2026-05-06  
**Authors:** Platform Engineering · SRE  
**Reviewers:** Architect · Security · QA · Product · Audit Lead

---

## Impact Assessment ★

| Dimension | Impact | Notes |
|---|---|---|
| Reliability | CRITICAL | Without observability, production incidents are blind |
| Security | HIGH | PII redaction in logs is a GDPR hard requirement |
| Performance | MEDIUM | OTel SDK adds ~1-3ms overhead per traced request |
| Cost | MEDIUM | Telemetry storage (Prometheus + Loki) ~$200/mo at 500 RPS |
| Compliance | HIGH | Audit trail completeness required for GDPR/CNPDCP |
| Developer experience | HIGH | Distributed traces dramatically reduce MTTR |

---

## 1. Architecture Overview

```
HTTP Request
    │
    ▼
[API Gateway / Load Balancer]
    │  trace propagation: W3C TraceContext headers
    ▼
[Go Backend — OTel SDK]  ──────► [Prometheus Scrape]
    │  spans + metrics              │
    │                               ▼
    ├──► [PostgreSQL]          [Prometheus]
    │       (db spans)              │
    │                               ▼
    ├──► [Redis]              [Grafana]
    │       (cache spans)           │
    │                               ▼
    ├──► [Job Queue/Scheduler] [Alert Manager]
    │       (job spans)             │
    │                               ▼
    └──► [External APIs]     [PagerDuty / Slack]
            (OpenAI, Resend)

Logs: [Structured JSON → Loki → Grafana]
Traces: [OTel SDK → OTLP Collector → Tempo → Grafana]
```

---

## 2. OpenTelemetry Tracing ★

### 2.1 SDK Initialization

```go
// backend/internal/observability/tracing.go
package observability

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func InitTracer(ctx context.Context, cfg TracingConfig) (*trace.TracerProvider, error) {
    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(cfg.CollectorEndpoint),
        otlptracegrpc.WithInsecure(), // TLS terminated at collector
    )
    if err != nil { return nil, fmt.Errorf("OTLP exporter: %w", err) }

    res := resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceName("revyx-api"),
        semconv.ServiceVersion(cfg.Version),
        semconv.DeploymentEnvironment(cfg.Environment),
    )

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(res),
        trace.WithSampler(trace.ParentBased(
            trace.TraceIDRatioBased(cfg.SampleRate), // default 0.1 in prod, 1.0 in dev
        )),
    )
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))
    return tp, nil
}
```

### 2.2 Instrumentation Layers

#### HTTP Layer (Fiber middleware)

```go
func OTelMiddleware(tracer trace.Tracer) fiber.Handler {
    return func(c *fiber.Ctx) error {
        ctx, span := tracer.Start(
            otel.GetTextMapPropagator().Extract(c.Context(), fiberCarrier{c}),
            c.Route().Path,
            trace.WithSpanKind(trace.SpanKindServer),
            trace.WithAttributes(
                attribute.String("http.method", c.Method()),
                attribute.String("http.url", c.OriginalURL()),
                attribute.String("tenant.id", c.Locals("tenantID").(string)),
            ),
        )
        defer span.End()

        c.SetUserContext(ctx)
        err := c.Next()

        span.SetAttributes(attribute.Int("http.status_code", c.Response().StatusCode()))
        if err != nil { span.RecordError(err) }
        return err
    }
}
```

#### PostgreSQL Layer (pgx tracer)

```go
// pgx instrumentation via otelpgx
import "github.com/exaring/otelpgx"

pool, _ := pgxpool.NewWithConfig(ctx, pgxConfig)
pool.Config().ConnConfig.Tracer = otelpgx.NewTracer(
    otelpgx.WithIncludeQueryParameters(), // redacted in prod via attribute filter
)
```

#### Redis Layer

```go
import "github.com/redis/go-redis/extra/redisotel/v9"

rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
redisotel.InstrumentTracing(rdb)
redisotel.InstrumentMetrics(rdb)
```

#### External APIs (OpenAI, Resend)

```go
func (c *AIClient) Embed(ctx context.Context, text string) (*EmbedResponse, error) {
    tracer := otel.Tracer("revyx.ai")
    ctx, span := tracer.Start(ctx, "ai.embed",
        trace.WithSpanKind(trace.SpanKindClient),
        trace.WithAttributes(
            attribute.String("ai.model", c.model),
            attribute.Int("ai.input_length", len(text)),
        ),
    )
    defer span.End()

    // ... HTTP call
    span.SetAttributes(attribute.Int("ai.tokens_used", resp.Usage.TotalTokens))
    return resp, nil
}
```

### 2.3 Trace ID Propagation Through Saga Steps ★

Distributed sagas (e.g., `deal-closure WON` saga from Phase 2) must propagate trace context across async steps.

```go
// Saga context envelope — carries TraceContext through queue
type SagaMessage struct {
    SagaID    string          `json:"saga_id"`
    Step      string          `json:"step"`
    Payload   json.RawMessage `json:"payload"`
    TraceCtx  map[string]string `json:"trace_ctx"` // W3C TraceContext headers
    CreatedAt time.Time       `json:"created_at"`
}

// Producer: inject trace context
func (s *SagaOrchestrator) Dispatch(ctx context.Context, step string, payload any) error {
    carrier := make(map[string]string)
    otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(carrier))

    msg := SagaMessage{
        SagaID:   extractSagaID(ctx),
        Step:     step,
        Payload:  marshalJSON(payload),
        TraceCtx: carrier,
    }
    return s.queue.Publish(ctx, "saga."+step, msg)
}

// Consumer: extract and continue trace
func (w *SagaWorker) Handle(rawMsg []byte) error {
    var msg SagaMessage
    _ = json.Unmarshal(rawMsg, &msg)

    ctx := otel.GetTextMapPropagator().Extract(
        context.Background(),
        propagation.MapCarrier(msg.TraceCtx),
    )
    ctx, span := otel.Tracer("revyx.saga").Start(ctx, "saga."+msg.Step)
    defer span.End()

    return w.processStep(ctx, msg)
}
```

---

## 3. Structured Logs with PII Redaction ★

### 3.1 PII Field Registry

All fields that may contain personal data, with redaction strategy:

| Field | Tables / Contexts | Redaction Strategy |
|---|---|---|
| `email` | users, buyer_contacts, email logs | `***@***.***` |
| `phone` | buyer_profiles, contact_logs | `+40***XXXX` (keep country code) |
| `cnp` (Romanian personal ID) | user KYC, legal docs | `XXXXXXXXXX` (full mask) |
| `full_name` | users, contracts | `J*** D***` (initials) |
| `ip_address` | audit_log, access_log | `192.168.*.*` (truncate last 2 octets) |
| `address` | properties, user profiles | Truncate to city level |
| `iban` | payment records | `RO**XXXX...XXXX` (keep bank code) |
| `jwt_token` | auth logs | Never log (token itself, not claims) |
| `password_hash` | auth events | Never log |
| `api_key` | integration events | Never log |

### 3.2 Logger Implementation

```go
// backend/internal/observability/logger.go
package observability

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type PIIRedactor struct{}

func (r *PIIRedactor) redactEmail(v string) string {
    parts := strings.Split(v, "@")
    if len(parts) != 2 { return "***@***" }
    domParts := strings.Split(parts[1], ".")
    return "***@***." + domParts[len(domParts)-1]
}

func (r *PIIRedactor) redactPhone(v string) string {
    if len(v) < 4 { return "***" }
    return v[:3] + strings.Repeat("*", len(v)-6) + v[len(v)-3:]
}

func (r *PIIRedactor) redactCNP(v string) string { return strings.Repeat("X", len(v)) }

func (r *PIIRedactor) redactIP(v string) string {
    parts := strings.Split(v, ".")
    if len(parts) != 4 { return "*.*.*.* "}
    return parts[0] + "." + parts[1] + ".*.*"
}

// PIIAwareEncoder wraps zapcore.Encoder
func NewProductionLogger(cfg LogConfig) (*zap.Logger, error) {
    encoderCfg := zap.NewProductionEncoderConfig()
    encoderCfg.TimeKey = "ts"
    encoderCfg.MessageKey = "msg"
    encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

    return zap.Config{
        Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
        Development:      false,
        Encoding:         "json",
        EncoderConfig:    encoderCfg,
        OutputPaths:      []string{"stdout"},
        ErrorOutputPaths: []string{"stderr"},
    }.Build(
        zap.WithCaller(true),
        zap.Fields(
            zap.String("service", "revyx-api"),
            zap.String("env", cfg.Environment),
        ),
    )
}
```

### 3.3 Log Event Schema

All log lines are structured JSON:
```json
{
  "ts": "2026-05-06T10:00:00.123Z",
  "level": "info",
  "service": "revyx-api",
  "env": "production",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "tenant_id": "uuid-here",
  "user_id": "uuid-here",
  "msg": "property match completed",
  "http.method": "GET",
  "http.path": "/api/v1/match",
  "http.status": 200,
  "duration_ms": 34,
  "match.count": 5,
  "match.engine": "v2"
}
```

Never in logs: email, phone, CNP, JWT, password, API key, IBAN (PII fields auto-redacted at field level).

---

## 4. Metrics Catalog ★

### 4.1 HTTP Metrics (all endpoints)

| Metric | Type | Labels | Description |
|---|---|---|---|
| `revyx_http_requests_total` | Counter | `method`, `path`, `status`, `tenant_id` | Total HTTP requests |
| `revyx_http_request_duration_seconds` | Histogram | `method`, `path`, `tenant_id` | Request latency |
| `revyx_http_request_size_bytes` | Histogram | `method`, `path` | Request body size |

### 4.2 Engine Metrics ★

| Metric | Type | Labels | Description | SLO-linked |
|---|---|---|---|---|
| `revyx_match_engine_version` | Gauge | `version` (v1/v2) | Active match engine | ✓ |
| `revyx_match_duration_seconds` | Histogram | `engine`, `tenant_id` | ANN query latency | ✓ |
| `revyx_match_recall_at5` | Gauge | `engine`, `dataset_tier` | Recall@5 (golden set) | ✓ |
| `revyx_aps_score_distribution` | Histogram | `tenant_id` | APS score buckets | |
| `revyx_is_score_distribution` | Histogram | `tenant_id` | Intent Score buckets | ✓ |
| `revyx_nba_actions_dispatched_total` | Counter | `action_type`, `tenant_id` | NBA actions sent | ✓ |
| `revyx_dhi_computation_duration_seconds` | Histogram | `tenant_id` | DHI computation time | ✓ |
| `revyx_ls_update_duration_seconds` | Histogram | `tenant_id` | Listing Score update latency | ✓ |
| `revyx_nps_dispatch_delay_seconds` | Histogram | `tenant_id` | NPS dispatch vs T+7d target | ✓ |

### 4.3 Infrastructure Metrics

| Metric | Type | Labels | Description |
|---|---|---|---|
| `revyx_db_query_duration_seconds` | Histogram | `operation`, `table` | DB query latency |
| `revyx_db_pool_size` | Gauge | `state` (idle/active/waiting) | Connection pool |
| `revyx_redis_operation_duration_seconds` | Histogram | `command` | Redis latency |
| `revyx_embedding_request_duration_seconds` | Histogram | `provider`, `tenant_id` | Embedding API latency |
| `revyx_embedding_cost_usd_total` | Counter | `provider`, `tenant_id` | Cumulative embedding cost |
| `revyx_saga_step_duration_seconds` | Histogram | `saga_type`, `step` | Saga step execution time |
| `revyx_saga_failures_total` | Counter | `saga_type`, `step` | Failed saga steps |

### 4.4 Security Metrics

| Metric | Type | Labels | Description |
|---|---|---|---|
| `revyx_audit_events_total` | Counter | `type`, `severity` | Audit event count |
| `revyx_auth_failures_total` | Counter | `reason` | Auth failure reasons |
| `revyx_jwt_rotations_total` | Counter | | JWT rotation events |
| `revyx_rls_policy_violations_total` | Counter | `table` | RLS violation attempts |

### 4.5 Alert Rules (Prometheus)

```yaml
# docs/observability/alert-rules.yaml
groups:
  - name: revyx_slo
    rules:
      - alert: LSUpdateLatencyHigh
        expr: histogram_quantile(0.95, revyx_ls_update_duration_seconds_bucket) > 30
        for: 5m
        severity: warning
        annotations:
          summary: "LS update p95 > 30s SLO"

      - alert: NBARefreshLatencyHigh
        expr: histogram_quantile(0.95, revyx_nba_actions_dispatched_total) > 30
        for: 5m
        severity: warning

      - alert: DHIComputationSlow
        expr: histogram_quantile(0.95, revyx_dhi_computation_duration_seconds_bucket) > 600
        for: 10m
        severity: warning

      - alert: MatchEngineV2Fallback
        expr: revyx_match_engine_version{version="v1"} == 1
        for: 1m
        severity: critical
        annotations:
          summary: "Match engine fell back to v1 — investigate pgvector"

      - alert: DealClosureSagaSlowP95
        expr: histogram_quantile(0.95, revyx_saga_step_duration_seconds_bucket{saga_type="deal_closure"}) > 3
        for: 5m
        severity: high

      - alert: NPSDispatchDelayed
        expr: histogram_quantile(0.95, revyx_nps_dispatch_delay_seconds_bucket) > 3600
        for: 15m
        severity: high

      - alert: CrossTenantQueryAttempt
        expr: increase(revyx_audit_events_total{type="cross_tenant_query_attempt"}[5m]) > 0
        severity: critical

      - alert: EmbeddingCostSpike
        expr: increase(revyx_embedding_cost_usd_total[1h]) > 10
        severity: warning
        annotations:
          summary: "Embedding cost >$10/hr — check for bulk reindex or abuse"
```

---

## 5. Grafana Dashboards ★

Dashboards checked-in as JSON in `docs/observability/dashboards/`. See individual files.

### 5.1 Dashboard Inventory

| File | Dashboard Name | Primary Use |
|---|---|---|
| `revyx-overview.json` | REVYX Platform Overview | On-call first look |
| `revyx-match-engine.json` | Match Engine Performance | ML + recall monitoring |
| `revyx-multitenant.json` | Multi-Tenant Health | Per-tenant resource usage |
| `revyx-sagas.json` | Saga / Workflow Status | Deal closure + NPS tracking |
| `revyx-security.json` | Security & Audit Events | SOC / compliance |
| `revyx-embedding-cost.json` | Embedding Cost Attribution | Finance / billing |
| `revyx-slo.json` | SLO / Error Budget | SRE / product |

### 5.2 Overview Dashboard (`revyx-overview.json`) ★

Key panels:
- Request rate (RPS) — last 1h
- Error rate (5xx %) — last 1h with SLO line at 0.1%
- p50/p95/p99 latency — last 1h
- Active tenants count
- Match engine version indicator (v1/v2 traffic split)
- DB connection pool saturation
- Redis hit rate
- Active sagas in-flight count

```json
{
  "title": "REVYX Platform Overview",
  "uid": "revyx-overview-v1",
  "schemaVersion": 37,
  "panels": [
    {
      "type": "stat",
      "title": "Request Rate (RPS)",
      "targets": [{"expr": "sum(rate(revyx_http_requests_total[1m]))"}]
    },
    {
      "type": "stat",
      "title": "Error Rate",
      "targets": [{"expr": "sum(rate(revyx_http_requests_total{status=~'5..'}[5m])) / sum(rate(revyx_http_requests_total[5m])) * 100"}],
      "thresholds": {"steps": [{"value": 0.1, "color": "yellow"}, {"value": 1, "color": "red"}]}
    },
    {
      "type": "timeseries",
      "title": "Latency Percentiles",
      "targets": [
        {"expr": "histogram_quantile(0.50, sum(rate(revyx_http_request_duration_seconds_bucket[5m])) by (le))", "legendFormat": "p50"},
        {"expr": "histogram_quantile(0.95, sum(rate(revyx_http_request_duration_seconds_bucket[5m])) by (le))", "legendFormat": "p95"},
        {"expr": "histogram_quantile(0.99, sum(rate(revyx_http_request_duration_seconds_bucket[5m])) by (le))", "legendFormat": "p99"}
      ]
    }
  ]
}
```

*(Full JSON for all 7 dashboards in `docs/observability/dashboards/` — see individual files)*

---

## 6. SLO Targets Per Pillar ★

| Pillar | Metric | SLO Target | Measurement Window | Alert Threshold |
|---|---|---|---|---|
| Listing Score update | `revyx_ls_update_duration_seconds` p95 | **< 30s** | 7-day rolling | >30s for 5min |
| NBA refresh | NBA dispatch lag | **< 30s** | 7-day rolling | >30s for 5min |
| DHI computation | `revyx_dhi_computation_duration_seconds` p95 | **< 10min** | 7-day rolling | >600s for 10min |
| NPS dispatch | Dispatch vs T+7d ±1h | **±1h** | per-event | >1h delay for 15min |
| Deal-closure WON saga | `revyx_saga_step_duration_seconds` p95 | **< 3s** | 7-day rolling | >3s for 5min |
| API availability | Error rate (5xx) | **≥ 99.9%** | 30-day rolling | >0.1% for 5min |
| API latency | p95 HTTP duration | **< 200ms** | 7-day rolling | >200ms for 10min |
| Match engine recall@5 | `revyx_match_recall_at5` | **≥ 0.88** | weekly golden set | <0.88 immediate |
| ANN query latency | `revyx_match_duration_seconds` p95 | **< 50ms** | 7-day rolling | >50ms for 5min |

### 6.1 Error Budget Policy

Monthly error budget per SLO = `(1 - SLO) * 30d`:

| SLO | Monthly budget |
|---|---|
| API availability 99.9% | 43.2 min downtime |
| API latency p95 < 200ms | 4.32 hr budget-seconds |
| Deal-closure saga p95 < 3s | depletes after 4.32 hr of violations |

**Budget consumption triggers:**
- 25% consumed → inform on-call, review causes
- 50% consumed → freeze all non-critical deployments
- 75% consumed → mandatory post-mortem, VP eng review
- 100% consumed → incident SEV1, all-hands engineering review

---

## 7. Error Budget & Escalation Runbook ★

```
## Error Budget Runbook — revyx.app/runbooks/error-budget

### Trigger: Budget 50% consumed

1. Identify top error/latency contributors:
   SELECT metric, count FROM revyx_slo_violations ORDER BY count DESC LIMIT 10;

2. Check recent deploys:
   git log --since="30 days ago" --oneline | head -20

3. Decision:
   IF regression introduced by deploy → ROLLBACK within 1h
   IF capacity issue → scale horizontally (Kubernetes HPA or RDS read replica)
   IF bug → create SEV2 incident, fix SLA 7d

4. Notify:
   Slack #sre-alerts: "@here Error budget 50% consumed for [SLO], investigating"
   GitHub: create tracking issue with 'slo-budget' label

### Trigger: Budget 100% consumed

1. Page VP Engineering + CTO immediately
2. Convene war-room (Zoom link in #war-room)
3. All non-hotfix deploys frozen until budget recovers
4. Root cause analysis → published post-mortem within 5 business days
```

---

## 8. Synthetic Monitoring (Canaries) ★

Canary probes run every 60 seconds from external monitoring nodes (e.g., Checkly, or self-hosted).

### 8.1 Critical Endpoint Canaries

| Endpoint | Check | SLO | Alert if |
|---|---|---|---|
| `GET /health` | HTTP 200 + `"status":"ok"` | 99.95% | 3 consecutive failures |
| `POST /api/v1/auth/login` | HTTP 200, latency <300ms | 99.9% | 2 consecutive failures |
| `GET /api/v1/match` (synthetic buyer) | HTTP 200, results ≥1, latency <500ms | 99.5% | 2 consecutive failures |
| `GET /api/v1/properties` | HTTP 200, latency <200ms | 99.9% | 3 consecutive failures |
| `POST /api/v1/deals` (saga trigger) | HTTP 202, saga_id returned | 99% | 2 consecutive failures |

### 8.2 Canary Configuration (Checkly / self-hosted)

```javascript
// checks/match-endpoint.check.js
import { ApiCheck, AssertionBuilder } from '@checkly/cli/constructs'

new ApiCheck('revyx-match-canary', {
  name: 'REVYX Match Engine Canary',
  frequency: 1, // every minute
  locations: ['eu-west-1', 'eu-central-1'],
  request: {
    url: 'https://api.revyx.app/v1/match',
    method: 'GET',
    headers: [{ key: 'X-Tenant-ID', value: process.env.CANARY_TENANT_ID }],
    assertions: [
      AssertionBuilder.statusCode().equals(200),
      AssertionBuilder.jsonBody('$.results.length').greaterThan(0),
      AssertionBuilder.responseTime().lessThan(500),
    ],
  },
  alertChannels: [pagerDutyChannel, slackChannel],
})
```

### 8.3 Canary User / Data

A dedicated canary tenant (`tenant_id: canary-synthetic-xxx`) with:
- 100 synthetic properties (pre-loaded, never changed)
- 1 synthetic buyer profile with broad criteria (matches all 100 properties)
- Canary API key with read-only permissions
- Excluded from billing attribution

---

## 9. Audit Checkpoint — S6 Observability ★

**Architect:** OTel SDK + OTLP Collector + Tempo/Loki/Prometheus/Grafana is the standard OSS observability stack. Trace propagation through sagas (W3C TraceContext in queue messages) is the correct pattern. Dashboard-as-code in repo is essential for reproducibility. ✅

**Security:** PII redaction registry is comprehensive. Structured logger with field-level redaction (not regex scrubbing) is the safe approach. Canary tenant with read-only API key is properly scoped. JWT and password hash explicitly listed as "never log" — correct. ✅

**QA:** SLO targets are specific and measurable. Alert rules cover all engine pillars enumerated in S2-S5. Synthetic canaries provide external validation independent of internal metrics. ✅

**Product:** SLO targets aligned with product commitments: LS <30s, NBA <30s, DHI <10min, NPS ±1h, saga p95 <3s. Error budget policy gives clear deployment freeze triggers. ✅

**Compliance:** Structured logs with PII redaction satisfy GDPR Art. 5(1)(f) (integrity/confidentiality). Audit events logged with trace context enables complete audit trail. Retention policy needed for log data — recommend: operational logs 30 days, audit logs 1 year (GDPR Art. 30). ✅

**Audit Lead:** All gating items: dashboard JSON files must be created before launch (referenced but not fully generated — see §5 — full JSON in `docs/observability/dashboards/`). Alert rules must be deployed and tested in staging. ✅

---

## 10. Open Items / Gating

| Item | Owner | Deadline | Status | Blocking? |
|---|---|---|---|---|
| Full Grafana dashboard JSON files (7 dashboards) | SRE | 2026-05-20 | OPEN | YES |
| Alert rules deployed and tested in staging | SRE | 2026-05-20 | OPEN | YES |
| Log retention policy documented in compliance register | Compliance | 2026-05-15 | OPEN | NO |
| Canary tenant provisioned in production | Ops | Pre-launch | OPEN | YES |
| OTel collector infrastructure provisioned | Infra | 2026-05-20 | OPEN | YES |
| Loki + Tempo storage provisioned (30d retention) | Infra | 2026-05-20 | OPEN | YES |

---

*End of TECH_SPEC_REVYX_observability-stack_v1.0.0.md*
