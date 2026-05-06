# TECH SPEC — REVYX Churn Prediction
**Document:** TECH_SPEC_REVYX_churn-prediction_v1.0.0.md  
**Version:** 1.0.0  
**Phase:** S7 — Phase 4 Post-Launch  
**Status:** APPROVED  
**Date:** 2026-05-06  
**Authors:** ML Engineering · Customer Success · Product  
**Reviewers:** Architect · Security · DBA · QA · Compliance · Product · Audit Lead

---

## Impact Assessment ★

| Dimension | Impact | Notes |
|---|---|---|
| Business | HIGH | B2B SaaS churn is existential — each tenant lost = significant MRR loss |
| Privacy | HIGH | Model must use only aggregate behavior metrics, never buyer personal data |
| Compliance | HIGH | Behavioral profiling of legal entities (tenants) requires documentation |
| Reliability | LOW | Churn model is async, not on request critical path |
| Infrastructure | LOW | Monthly retrain batch job, low resource cost |
| Customer Success | HIGH | Automated NBA re-engagement + CS escalation reduces manual triage |

---

## 1. Churn Definition

**Churn event:** A tenant (B2B) cancels subscription OR downgrades to Starter AND has
not renewed within the grace period.

**Churn risk window:** Model predicts probability of churn in the **next 30 days**,
evaluated daily.

**Label construction (training):**
```
is_churned = 1  IF tenant cancelled/downgraded in [T+1d, T+30d]
is_churned = 0  IF tenant remained active throughout [T+1d, T+30d]
Label assigned daily per tenant for the last 12 months of history.
```

---

## 2. Feature Engineering

### 2.1 Feature Set (Tenant-Level, Aggregate Only)

All features are computed over a rolling 30-day window unless specified.

| Feature | Type | Computation | Privacy |
|---|---|---|---|
| `login_frequency_7d` | float32 | count(distinct days with login) / 7 | Aggregate per tenant |
| `login_frequency_30d` | float32 | count(distinct days with login) / 30 | Aggregate per tenant |
| `active_agents_ratio` | float32 | active_agents / total_agents | No PII |
| `match_acceptance_rate` | float32 | accepted_matches / total_matches | Aggregate rate |
| `match_view_rate` | float32 | viewed_matches / total_matches | Aggregate rate |
| `deal_conversion_rate` | float32 | deals_closed / deals_initiated | Aggregate rate |
| `deal_won_rate_30d` | float32 | deal_WON / deal_initiated last 30d | Aggregate rate |
| `nba_response_rate` | float32 | nba_responded / nba_sent | Aggregate rate |
| `nba_response_latency_p50` | float32 | median hours to NBA response | Aggregate |
| `support_tickets_30d` | int8 | count(support_tickets) | No PII |
| `support_critical_30d` | int8 | count(tickets with priority=HIGH) | No PII |
| `billing_overdue_days` | int8 | days since last successful payment | No PII |
| `listing_count_trend` | float32 | (listings_30d - listings_60d) / max(listings_60d, 1) | Growth/decline |
| `plan_tier` | int8 | Starter=0, Growth=1, Enterprise=2 | No PII |
| `tenure_days` | int16 | days since tenant created | No PII |
| `days_since_last_admin_login` | int8 | capped at 60 | Aggregate |

**Categorical (label-encoded):**
- `contract_type`: monthly=0 / annual=1
- `region`: RO_urban=0 / RO_rural=1 / UA=2

**NOT included (privacy boundary):**
- Buyer names, emails, phone numbers, CNP
- Individual agent behavior (only aggregate tenant-level)
- Property addresses or listing content

### 2.2 Feature Store Schema

```sql
-- migration 016 (S7)
CREATE TABLE churn_feature_snapshots (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    snapshot_date   DATE NOT NULL,
    features        JSONB NOT NULL,
    churn_label     BOOLEAN,  -- NULL until T+30d outcome known
    labeled_at      TIMESTAMPTZ,
    PRIMARY KEY (tenant_id, snapshot_date)
);

CREATE TABLE churn_predictions (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           UUID NOT NULL,
    predicted_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    churn_probability   FLOAT NOT NULL CHECK (churn_probability BETWEEN 0 AND 1),
    risk_tier           VARCHAR(10) NOT NULL CHECK (risk_tier IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    model_version       VARCHAR(32) NOT NULL,
    top_risk_factors    JSONB,  -- top 3 SHAP features
    action_triggered    VARCHAR(50),  -- 'nba_reengagement' | 'cs_escalation' | null
    action_triggered_at TIMESTAMPTZ,
    INDEX idx_churn_predictions_tenant (tenant_id, predicted_at DESC)
);
```

---

## 3. Model Architecture

### 3.1 Algorithm: Gradient Boosting Classifier (LightGBM)

Same infrastructure as ML Pricing (S7-2), separate model in MLflow registry.

```python
# ml/churn/train.py
PARAMS = {
    'objective': 'binary',
    'metric': ['auc', 'binary_logloss'],
    'num_leaves': 63,
    'learning_rate': 0.05,
    'feature_fraction': 0.8,
    'bagging_fraction': 0.8,
    'bagging_freq': 5,
    'min_child_samples': 30,  # higher — churn dataset smaller than pricing
    'scale_pos_weight': 4.0,  # class imbalance: ~20% churn rate expected
    'n_estimators': 500,
    'early_stopping_rounds': 30,
    'verbose': -1,
}
```

**Class imbalance handling:**
- `scale_pos_weight = (negative samples) / (positive samples)` — computed per training run
- Calibration: Platt scaling to align predicted probabilities to observed churn rate

### 3.2 Evaluation Metrics

| Metric | Target | Rationale |
|---|---|---|
| AUC-ROC | **≥ 0.78** | Primary eval — discrimination ability |
| Precision @ top 20% risk | ≥ 0.65 | CS team can only act on top-ranked tenants |
| Recall @ CRITICAL tier | ≥ 0.80 | Missing a CRITICAL churner is costly |
| Brier score | < 0.15 | Calibration quality |

### 3.3 Risk Tier Thresholds

Thresholds calibrated on validation set to achieve precision target:

| Tier | Probability Range | Action |
|---|---|---|
| `LOW` | < 0.25 | None — monitor |
| `MEDIUM` | 0.25 – 0.55 | Automated NBA re-engagement |
| `HIGH` | 0.55 – 0.80 | CS escalation (ticket created) |
| `CRITICAL` | ≥ 0.80 | CS escalation + VP Sales alert |

Thresholds stored in MLflow model metadata — adjustable without retraining.

---

## 4. Inference Pipeline

### 4.1 Daily Batch Scoring

```python
# airflow/dags/churn_scoring.py
with DAG(
    dag_id='churn_daily_scoring',
    schedule_interval='0 6 * * *',  # 06:00 UTC daily, after feature refresh
    start_date=datetime(2026, 5, 6),
    catchup=False,
) as dag:

    build_features = PythonOperator(
        task_id='build_daily_features',
        python_callable=build_churn_features,
        # Computes all features for all active tenants
    )

    score_tenants = PythonOperator(
        task_id='score_tenants',
        python_callable=run_churn_inference,
    )

    trigger_actions = PythonOperator(
        task_id='trigger_actions',
        python_callable=dispatch_churn_actions,
        # MEDIUM+ → NBA, HIGH+ → CS ticket
    )

    build_features >> score_tenants >> trigger_actions
```

```python
def run_churn_inference():
    model = mlflow.lightgbm.load_model('models:/revyx_churn/Production')
    thresholds = load_tier_thresholds()  # from MLflow model metadata

    tenants = query_active_tenants()
    features_df = build_feature_matrix(tenants)
    probabilities = model.predict(features_df)

    # SHAP top-3 features per tenant (for CS context)
    explainer = shap.TreeExplainer(model)
    shap_values = explainer.shap_values(features_df)

    results = []
    for i, tenant_id in enumerate(tenants):
        prob = float(probabilities[i])
        tier = classify_tier(prob, thresholds)
        top_factors = get_top_shap_factors(shap_values[i], features_df.columns, n=3)

        results.append({
            'tenant_id': tenant_id,
            'churn_probability': prob,
            'risk_tier': tier,
            'top_risk_factors': top_factors,
            'model_version': model.metadata.run_id[:8],
        })

    bulk_insert_churn_predictions(results)
    return results
```

---

## 5. Automated Actions

### 5.1 NBA Re-Engagement (MEDIUM+)

When a tenant reaches `MEDIUM` or higher risk, the NBA (Next Best Action) engine
inserts a re-engagement action:

```go
// backend/internal/nba/rules.go — new rule type: churn_reengagement
type ChurnReengagementAction struct {
    TenantID        string
    RiskTier        string
    TopRiskFactors  []string  // from SHAP
}

func (r *NBAEngine) ProcessChurnAlert(ctx context.Context, alert ChurnReengagementAction) error {
    msg := r.buildReengagementMessage(alert)
    // Delivered via: in-app notification to tenant_admin + optional email
    return r.dispatch(ctx, alert.TenantID, NBA_REENGAGEMENT, msg)
}

func (r *NBAEngine) buildReengagementMessage(alert ChurnReengagementAction) string {
    // Human-readable, NOT revealing the model or probability score
    // Example: "Your team's match acceptance rate has dropped this week.
    //           Review the latest recommendations to close pending deals faster."
    // Top risk factors translated to human-readable suggestions (hardcoded map)
    return r.reengagementTemplate(alert.TopRiskFactors)
}
```

**Privacy rule:** NBA message must NOT reveal churn probability or risk tier to tenant.
The tenant sees a helpful nudge, not a "we think you're about to churn" alert.

### 5.2 CS Escalation (HIGH+)

```go
// backend/internal/churn/actions.go
func (s *ChurnService) EscalateToCS(ctx context.Context, pred ChurnPrediction) error {
    ticket := SupportTicket{
        TenantID:    pred.TenantID,
        Priority:    ticketPriority(pred.RiskTier),  // HIGH→P2, CRITICAL→P1
        Subject:     fmt.Sprintf("[CHURN RISK] Tenant %s — %s", pred.TenantID, pred.RiskTier),
        Body:        buildCSTicketBody(pred),  // includes top risk factors + historical trend
        Tags:        []string{"churn-prediction", strings.ToLower(pred.RiskTier)},
        AssignedTo:  "cs-team",
    }
    if pred.RiskTier == "CRITICAL" {
        ticket.CcEmails = []string{csVPEmail()}
    }
    return s.ticketing.Create(ctx, ticket)
}
```

CS ticket body includes:
- Tenant name + plan + tenure
- Risk tier (for internal CS use only)
- Top 3 risk factors in plain language
- 30-day trend chart URL (Grafana panel link)
- Suggested action checklist (from playbook)

---

## 6. Monthly Retraining

```python
# airflow/dags/churn_retrain.py
with DAG(
    dag_id='churn_monthly_retrain',
    schedule_interval='0 3 1 * *',  # 1st of each month, 03:00 UTC
    start_date=datetime(2026, 6, 1),
    catchup=False,
) as dag:

    label_outcomes = PythonOperator(
        task_id='label_30d_outcomes',
        python_callable=backfill_churn_labels,
        # Sets churn_label on snapshots where T+30d has elapsed
    )

    train_new_model = PythonOperator(
        task_id='train_churn_model',
        python_callable=train_churn_classifier,
    )

    evaluate = PythonOperator(
        task_id='evaluate_auc_gate',
        python_callable=evaluate_and_gate,
        # Fails DAG if AUC-ROC < 0.78 — keeps old model in production
    )

    label_outcomes >> train_new_model >> evaluate
```

---

## 7. Privacy Architecture

```
Training data:
  ✅ Aggregate behavior rates (login_frequency, match_acceptance_rate, etc.)
  ✅ Billing metadata (overdue_days, plan_tier)
  ✅ Tenant-level counts (support_tickets, active_agents)
  ❌ Buyer names, emails, CNP, phone numbers
  ❌ Individual agent login timestamps
  ❌ Property addresses
  ❌ Chat/message content

Inference output stored in churn_predictions:
  ✅ tenant_id (UUID, no PII)
  ✅ churn_probability (float)
  ✅ top_risk_factors (feature names — no values, no PII)
  ❌ Never stored: individual buyer identifiers, deal party names

GDPR basis for behavioral analysis:
  Legal entity (tenant) profiling for contract management
  Basis: Article 6(1)(b) — performance of contract (SaaS service agreement)
  Documented in: DPIA §3.4 (update required for churn model — see Compliance note below)
```

---

## 8. Audit Checkpoint — S7-3 Churn Prediction ★

**Architect:** Daily batch scoring (not real-time) is correct — churn is a lagging signal, sub-minute latency adds no value. Separate MLflow model registry entry for churn vs pricing is correct (different feature sets, different retraining cadence). SHAP explainability for top-3 risk factors is required for CS utility and GDPR Art. 22 automated decision safeguards. Risk tier thresholds stored in model metadata (not hardcoded) enables tuning without retraining — good design. ✅

**Security:** CS ticket body contains risk tier and model factors — this is internal CS-only data, must not be exposed to the tenant directly. NBA re-engagement message must be validated to contain zero probability/tier information. `churn_predictions` table must have row-level security: readable only by `system` role and `audit_reader` — NOT by `tenant_admin`. ✅

**DBA:** `churn_feature_snapshots` will accumulate ~365 rows/tenant/year — add retention policy: delete snapshots older than 18 months (keep enough for retraining). `churn_predictions` index on `(tenant_id, predicted_at DESC)` is correct for CS dashboard queries. Label backfill job must use `UPDATE ... WHERE labeled_at IS NULL AND snapshot_date <= NOW() - INTERVAL '30 days'` — include tenant_id in WHERE to prevent cross-tenant update. ✅

**QA:** Required tests: (1) AUC-ROC gate blocks model promotion if < 0.78, (2) CRITICAL tier triggers VP Sales CC, (3) NBA message contains zero probability/tier text (string assert), (4) tenant_admin API returns no churn data (403/404), (5) backfill labeling is idempotent (run twice → same result). ✅

**Compliance:** Churn model profiles legal entities (tenants), not natural persons directly. However, behavior features are derived from agent activity (natural persons employed by tenant). GDPR Art. 22 (automated decision-making) applies if churn score directly triggers suspension — it does not (CS human reviews before suspension). **Action required: update DPIA §3.4 to document churn model data flows.** Legal basis: Art. 6(1)(b) for contract management is defensible. Retention policy for churn_predictions: 24 months. ✅

**Product:** Re-engagement NBA message copy must be A/B tested before GA — message tone is critical (helpful, not alarming). CRITICAL tier escalation to VP Sales: ensure CS lead reviews the alert before VP is contacted (avoid noise). Churn dashboard for CS team: Grafana panel showing risk distribution across tenant base. ✅

**Audit Lead:** **1 hard requirement before GA:**
- [ ] DPIA §3.4 updated to document churn model, signed by DPO before production scoring begins
- [ ] `churn_predictions` RLS policy deployed and verified before first batch run

---

*End of TECH_SPEC_REVYX_churn-prediction_v1.0.0.md*
