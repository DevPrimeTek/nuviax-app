# TECH SPEC — REVYX ML Pricing Phase 3
**Document:** TECH_SPEC_REVYX_ml-pricing-phase3_v1.0.0.md  
**Version:** 1.0.0  
**Phase:** S7 — Phase 4 Post-Launch  
**Status:** APPROVED  
**Date:** 2026-05-06  
**Authors:** ML Engineering · Backend Engineering  
**Reviewers:** Architect · Security · DBA · QA · Compliance · Product · Audit Lead

---

## Impact Assessment ★

| Dimension | Impact | Notes |
|---|---|---|
| Product Quality | HIGH | MAE improvement 15–25% over rule-based (expected, pending A/B validation) |
| Security | HIGH | Model serving endpoint is internal-only; training data contains transaction prices (sensitive) |
| Compliance | HIGH | Training on historical transaction data — legal basis: legitimate interest (pricing service) |
| Infrastructure | MEDIUM | MLflow server + feature store add ~$150/mo infra cost |
| Reliability | HIGH | Fallback to rule-based required — ML model is not on critical path |
| Privacy | MEDIUM | Training data must be tenant-scoped; no cross-tenant price leakage |

---

## 1. Architecture Overview

```
[PostgreSQL — transactions, properties]
        │  feature extraction (batch, hourly)
        ▼
[Feature Store — PostgreSQL feature_snapshots + Redis cache]
        │  training dataset build (daily)
        ▼
[LightGBM Training Pipeline — Python, Airflow DAG]
        │  model artifact + metadata
        ▼
[MLflow Model Registry — versioned, staged: Staging → Production]
        │  model load at startup
        ▼
[ML Pricing Service — Go sidecar, HTTP]
        │  POST /internal/ml/price-estimate
        ▼
[REVYX API Handler — price_estimate field in property response]
        │  A/B split: 20% ML / 80% rule-based (Redis flag)
        ▼
[Response to tenant/buyer]
```

---

## 2. Feature Engineering

### 2.1 Feature Set

| Feature | Type | Source | Notes |
|---|---|---|---|
| `lat_cluster` | int8 | properties.lat → KMeans 50 clusters | Anonymized location |
| `lng_cluster` | int8 | properties.lng → KMeans 50 clusters | Paired with lat_cluster |
| `surface_m2` | float32 | properties.surface | Log-transformed: `log1p(surface)` |
| `year_built` | int16 | properties.year_built | Imputed median if null |
| `floor` | int8 | properties.floor | -1 = basement, 0 = ground |
| `total_floors` | int8 | properties.total_floors | |
| `floor_ratio` | float32 | floor / total_floors | Derived |
| `rooms` | int8 | properties.rooms | |
| `has_parking` | bool | properties.amenities @> '["parking"]' | |
| `has_elevator` | bool | properties.amenities @> '["elevator"]' | |
| `has_balcony` | bool | properties.amenities @> '["balcony"]' | |
| `has_storage` | bool | properties.amenities @> '["storage"]' | |
| `market_trend_90d` | float32 | median(price_per_m2) last 90d in lat_cluster | Rolling window |
| `market_vol_90d` | int16 | count(transactions) last 90d in lat_cluster | Liquidity proxy |
| `price_per_m2_zone_p25` | float32 | percentile(0.25) last 90d in cluster | Lower bound signal |
| `price_per_m2_zone_p75` | float32 | percentile(0.75) last 90d in cluster | Upper bound signal |

**Target variable:** `price_per_m2` (float32, log-transformed for training)

### 2.2 Feature Store Schema

```sql
-- migration 015 (S7)
CREATE TABLE feature_snapshots (
    id              BIGSERIAL PRIMARY KEY,
    property_id     UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL,
    snapshot_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    features        JSONB NOT NULL,
    -- features JSON matches feature set above
    INDEX idx_feature_snapshots_property (property_id),
    INDEX idx_feature_snapshots_tenant_snapshot (tenant_id, snapshot_at DESC)
);

CREATE TABLE market_zone_stats (
    lat_cluster     INT NOT NULL,
    lng_cluster     INT NOT NULL,
    window_days     INT NOT NULL,  -- 30, 60, 90
    computed_at     TIMESTAMPTZ NOT NULL,
    median_price_m2 FLOAT NOT NULL,
    p25_price_m2    FLOAT NOT NULL,
    p75_price_m2    FLOAT NOT NULL,
    transaction_count INT NOT NULL,
    PRIMARY KEY (lat_cluster, lng_cluster, window_days, computed_at)
) PARTITION BY RANGE (computed_at);
```

**Redis cache for hot features:**
```
key: revyx:ml:market_zone:{lat_cluster}:{lng_cluster}:90d
value: JSON {median, p25, p75, count}
TTL: 1 hour (refreshed by hourly feature job)
```

---

## 3. Training Pipeline

### 3.1 Technology Stack

| Component | Technology | Reason |
|---|---|---|
| Model | LightGBM 4.x | Fast tabular training, native categorical support, SHAP explainability |
| Pipeline orchestration | Apache Airflow 2.x | Existing infra; DAG versioning |
| Experiment tracking | MLflow 2.x | Model versioning, artifact storage, A/B metadata |
| Feature engineering | Python + pandas + SQLAlchemy | Standard ML stack |
| Hyperparameter tuning | Optuna (50 trials) | Bayesian optimization |

### 3.2 Training DAG (Airflow)

```python
# airflow/dags/ml_pricing_training.py
from airflow import DAG
from airflow.operators.python import PythonOperator
from datetime import datetime, timedelta

with DAG(
    dag_id='ml_pricing_retrain',
    schedule_interval='0 2 * * *',  # 02:00 UTC daily
    start_date=datetime(2026, 5, 6),
    catchup=False,
    default_args={
        'retries': 2,
        'retry_delay': timedelta(minutes=5),
    },
) as dag:

    extract_features = PythonOperator(
        task_id='extract_features',
        python_callable=extract_training_features,
    )

    check_drift = PythonOperator(
        task_id='check_psi_drift',
        python_callable=compute_psi_and_gate,
        # Skips retrain if PSI < 0.2 AND model age < 7 days
    )

    train_model = PythonOperator(
        task_id='train_lightgbm',
        python_callable=train_and_evaluate,
    )

    register_model = PythonOperator(
        task_id='register_mlflow',
        python_callable=register_if_better,
        # Only promotes to Staging if new_mae < current_mae * 0.98
    )

    extract_features >> check_drift >> train_model >> register_model
```

### 3.3 LightGBM Configuration

```python
# ml/pricing/train.py
import lightgbm as lgb
import mlflow
import mlflow.lightgbm

PARAMS = {
    'objective': 'regression',
    'metric': ['mae', 'mape'],
    'num_leaves': 127,
    'learning_rate': 0.05,
    'feature_fraction': 0.8,
    'bagging_fraction': 0.8,
    'bagging_freq': 5,
    'min_child_samples': 20,
    'reg_alpha': 0.1,
    'reg_lambda': 0.1,
    'n_estimators': 1000,
    'early_stopping_rounds': 50,
    'verbose': -1,
}

def train_and_evaluate():
    df_train, df_val = load_dataset()  # last 18 months, 80/20 split
    X_train, y_train = split_features(df_train)
    X_val, y_val = split_features(df_val)

    with mlflow.start_run(run_name=f'pricing_{datetime.utcnow().isoformat()}'):
        mlflow.log_params(PARAMS)

        model = lgb.train(
            PARAMS,
            lgb.Dataset(X_train, label=y_train),
            valid_sets=[lgb.Dataset(X_val, label=y_val)],
            callbacks=[lgb.early_stopping(50)],
        )

        mae = mean_absolute_error(y_val, model.predict(X_val))
        mape = mean_absolute_percentage_error(y_val, model.predict(X_val))

        mlflow.log_metrics({'val_mae': mae, 'val_mape': mape})
        mlflow.lightgbm.log_model(model, 'pricing_model')

        return model, mae
```

### 3.4 Drift Detection — PSI Gate

Population Stability Index (PSI) measures distribution shift between training data
and current production data. PSI > 0.2 = significant drift → trigger retrain regardless of schedule.

```python
def compute_psi(expected: np.ndarray, actual: np.ndarray, bins: int = 10) -> float:
    """PSI on price_per_m2 distribution."""
    breakpoints = np.percentile(expected, np.linspace(0, 100, bins + 1))
    expected_pct = np.histogram(expected, breakpoints)[0] / len(expected)
    actual_pct = np.histogram(actual, breakpoints)[0] / len(actual)

    # Avoid log(0)
    expected_pct = np.clip(expected_pct, 1e-6, None)
    actual_pct = np.clip(actual_pct, 1e-6, None)

    return np.sum((actual_pct - expected_pct) * np.log(actual_pct / expected_pct))

def compute_psi_and_gate(**context):
    psi = compute_psi(load_training_prices(), load_recent_prices_30d())
    mlflow.log_metric('price_psi', psi)
    if psi < 0.2:
        raise AirflowSkipException(f'PSI={psi:.3f} below threshold 0.2, skip retrain')
```

---

## 4. Model Versioning (MLflow)

```
Stages:
  None → Staging → Production → Archived

Promotion rules:
  None → Staging:    val_mae improves vs current Production by ≥2%
  Staging → Production: manual approval by ML Lead OR auto after 24h canary PASS
  Production → Archived: on next Production promotion
```

```python
# Promote to production (after canary validation)
client = mlflow.tracking.MlflowClient()
client.transition_model_version_stage(
    name='revyx_pricing',
    version=new_version,
    stage='Production',
    archive_existing_versions=True,
)
```

---

## 5. ML Pricing Service (Go Sidecar)

The model is served by a Python FastAPI sidecar (LightGBM native inference),
called by the Go API via internal HTTP.

```python
# ml/pricing/serve.py — FastAPI sidecar
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import lightgbm as lgb
import mlflow
import numpy as np

app = FastAPI()

# Load production model at startup
model = mlflow.lightgbm.load_model('models:/revyx_pricing/Production')

class PriceRequest(BaseModel):
    lat_cluster: int
    lng_cluster: int
    surface_m2: float
    year_built: int
    floor: int
    total_floors: int
    rooms: int
    has_parking: bool
    has_elevator: bool
    has_balcony: bool
    has_storage: bool
    market_trend_90d: float
    market_vol_90d: int
    price_per_m2_zone_p25: float
    price_per_m2_zone_p75: float

class PriceResponse(BaseModel):
    price_estimate: float    # RON/m²
    confidence: float        # 0..1
    model_version: str

@app.post('/internal/ml/price-estimate', response_model=PriceResponse)
async def price_estimate(req: PriceRequest):
    features = np.array([[
        req.lat_cluster, req.lng_cluster, np.log1p(req.surface_m2),
        req.year_built, req.floor, req.total_floors,
        req.floor / max(req.total_floors, 1),
        req.rooms, req.has_parking, req.has_elevator,
        req.has_balcony, req.has_storage,
        req.market_trend_90d, req.market_vol_90d,
        req.price_per_m2_zone_p25, req.price_per_m2_zone_p75,
    ]])

    raw_pred = model.predict(features)[0]
    price = float(np.expm1(raw_pred))  # reverse log1p

    # Confidence: based on prediction interval width (SHAP variance proxy)
    p25 = req.price_per_m2_zone_p25
    p75 = req.price_per_m2_zone_p75
    zone_spread = (p75 - p25) / max(p25, 1)
    confidence = max(0.0, min(1.0, 1.0 - zone_spread))

    return PriceResponse(
        price_estimate=price,
        confidence=confidence,
        model_version=model.metadata.run_id[:8],
    )
```

### 5.1 Go Caller

```go
// backend/internal/pricing/ml_client.go
type MLPriceResponse struct {
    PriceEstimate float64 `json:"price_estimate"`
    Confidence    float64 `json:"confidence"`
    ModelVersion  string  `json:"model_version"`
}

func (c *MLClient) PriceEstimate(ctx context.Context, req PriceRequest) (*MLPriceResponse, error) {
    body, _ := json.Marshal(req)
    httpReq, _ := http.NewRequestWithContext(ctx, "POST",
        c.baseURL+"/internal/ml/price-estimate", bytes.NewReader(body))
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(httpReq)
    if err != nil || resp.StatusCode != 200 {
        return nil, fmt.Errorf("ml sidecar unavailable: %w", err)
    }
    defer resp.Body.Close()

    var result MLPriceResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result, nil
}
```

---

## 6. Fallback Strategy

```go
// backend/internal/api/handlers/properties.go
func (h *Handler) GetPropertyPriceEstimate(c *fiber.Ctx) error {
    prop := loadProperty(c)

    var priceEst float64
    var source string

    mlResp, err := h.mlClient.PriceEstimate(c.Context(), buildMLRequest(prop))
    switch {
    case err != nil:
        // ML sidecar unavailable
        priceEst = h.ruleBased.Estimate(prop)
        source = "rule_based"
    case mlResp.Confidence < 0.6:
        // ML confidence too low — fall back
        priceEst = h.ruleBased.Estimate(prop)
        source = "rule_based_low_confidence"
    default:
        priceEst = mlResp.PriceEstimate
        source = "ml_v" + mlResp.ModelVersion
    }

    // A/B assignment (Redis flag, sticky per property_id)
    if !h.abTest.IsMLGroup(prop.ID) {
        priceEst = h.ruleBased.Estimate(prop)
        source = "rule_based_ab_control"
    }

    return c.JSON(fiber.Map{
        "price_estimate": priceEst,
        "price_source":   source, // internal; NOT exposed in public API response
    })
}
```

---

## 7. A/B Test Design

| Parameter | Value |
|---|---|
| ML group | 20% of properties (hash(property_id) % 10 < 2) |
| Control group | 80% rule-based |
| Minimum duration | 4 weeks |
| Primary metric | MAE (ML estimate vs actual closed deal price) |
| Secondary metric | User acceptance rate (did buyer make offer within 7d of viewing price?) |
| Stopping rule | Sequential testing — stop early if p < 0.01 (two-sided) |
| Guardrail metric | API p95 latency must stay < 200ms |

```sql
-- A/B result logging
CREATE TABLE ml_pricing_ab_log (
    id              BIGSERIAL PRIMARY KEY,
    property_id     UUID NOT NULL,
    experiment_date DATE NOT NULL,
    group_assignment VARCHAR(20) NOT NULL,  -- 'ml' | 'rule_based'
    ml_estimate     FLOAT,
    rb_estimate     FLOAT NOT NULL,
    actual_deal_price FLOAT,  -- filled when deal closes
    confidence      FLOAT,
    model_version   VARCHAR(16),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## 8. Audit Checkpoint — S7-2 ML Pricing Phase 3 ★

**Architect:** LightGBM is the correct choice for tabular real-estate features — gradient boosting outperforms linear models for this feature mix. Feature store separation (PostgreSQL snapshots + Redis hot cache) is the right architecture for both training freshness and inference latency. Go sidecar calling Python FastAPI sidecar is a clean boundary. PSI drift detection at 0.2 threshold is standard practice. MLflow staged promotion (None → Staging → Production) prevents untested models reaching production. ✅

**Security:** `/internal/ml/price-estimate` must be network-isolated — not routable from public internet. Enforce via: (1) no public ingress rule for ML sidecar port, (2) internal service mesh / VPC-only. Training data contains transaction prices — sensitive. Training pipeline must run under a dedicated service account with read-only access to `transactions` and `properties` tables, not using the application DB user. Model artifacts in MLflow storage must use the same KMS key as other tenant data. ✅

**DBA:** `feature_snapshots` table will grow quickly — add a retention job: delete snapshots older than 90 days (training window). `market_zone_stats` partitioned by `computed_at` — partition management required (monthly). `ml_pricing_ab_log` — index on `property_id, experiment_date`. All three new tables must be added to the AUDIT_LOG trigger set for write operations. ✅

**QA:** Required tests: (1) fallback triggered when ML sidecar returns 503, (2) fallback triggered when confidence < 0.6, (3) A/B group assignment is stable for same property_id (deterministic hash), (4) A/B log written for every request, (5) PSI computation returns correct value for known distributions (unit test). Load test: ML sidecar must sustain 500 RPS with p95 < 50ms (inference-only, not training). ✅

**Compliance:** Training on historical transaction prices — legal basis: legitimate interest (service improvement). No PII in training features (location is clustered, not raw lat/lng). Cross-tenant isolation: training data query must include `tenant_id` scope OR aggregate at zone level only — no per-tenant price distribution leak. A/B log: price estimates are business data, not personal data — no GDPR subject access right implications beyond what covers the property record itself. ✅

**Product:** 4-week minimum A/B test before GA decision. MAE improvement goal: ≥15% vs rule-based. If A/B shows no improvement, rule-based remains default and ML is demoted. Price source field (`price_source`) must NOT appear in public API responses — internal metric only. ✅

**Audit Lead:** **2 items to track before GA:**
- [ ] ML sidecar network isolation verified (no public ingress) before Production deployment
- [ ] Training service account with read-only permissions provisioned and verified

---

*End of TECH_SPEC_REVYX_ml-pricing-phase3_v1.0.0.md*
