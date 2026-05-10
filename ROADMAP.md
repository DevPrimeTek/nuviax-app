# ROADMAP.md — NuviaX Implementation Roadmap

> **Versiune:** 1.6.0  
> **Actualizat:** 2026-05-06  
> **Framework:** NuviaX Growth Framework Rev 5.6 (C1–C40)  
> **MVP Scope:** 29 componente (14 FULL + 15 SIMPLIFIED), 11 POST_MVP

> **🚨 STATUS BUILD:** F0–F7.3 complete. AUDIT-01 (2026-04-28) a identificat 28 devieri majore. **Faza F8 (8 sub-faze) este NECESARĂ pentru lansare MVP.** Plan: `docs/fix-plan/00-PLAN-MASTER.md`.

---

## Stare curentă

| Fază | Status | Componente acoperite | Data |
|---|---|---|---|
| F0 — Reset | ✅ | — | — |
| F1 — Schema DB | ✅ | Schema pentru toate C1–C40 | — |
| F2 — Auth CSS | ✅ | — | — |
| F0.1 — Cleanup | ✅ | Fișiere v10.x eliminate | 2026-04-06 |
| F3 — Core Engine | ✅ | C1–C8, C14, C19–C21, C24–C26, C28, C30, C37, C38 | 2026-04-07 |
| F4 — Scheduler + SRM | ✅ | C23, C27, C33, C32 (Pause only) | 2026-04-07 |
| F5 — API Handlers | ✅ | Auth + Goals + Today + Dashboard + SRM + Achievements + Profile + Admin — 30 endpoints | 2026-04-18 |
| F6 — Frontend MVP | ✅ | Build complet, zero erori TypeScript. Audit endpoint-uri OK. | 2026-04-18 |
| F7 — Smoke Test | ✅ | Build PASS, 12/12 unit tests, smoke plan documentat | 2026-04-18 |
| F7.1 — Onboarding unblock | ✅ | AI suggest-category + BM + directions integrate în onboarding | 2026-04-19 |
| F7.2 — Onboarding SMART parsing | ✅ | `POST /goals/parse` + `ParseAndSuggestGO` + flux suggestions clickabile | 2026-04-20 |
| F7.3 — Admin panel separat | ✅ | `/admin/login` standalone, middleware dedicat | 2026-04-21 |
| AUDIT-01 — System Alignment Audit | ✅ | 28 devieri identificate (8 CRIT + 9 HIGH + 7 MED + 4 LOW) | 2026-04-28 |
| **F8.1 — Schema Reconciliation** | ⏳ PENDING | DEV-02/03/04/17 — tabele DB lipsă local | — |
| **F8.2 — Engine Restructure** | ⏳ PENDING | DEV-11 — funcții engine lipsă | — |
| **F8.3 — API Security Hardening** | ⏳ PENDING | DEV-01 — sprint_score leak + opacity test | — |
| **F8.4 — Scheduler Wiring** | ⏳ PENDING | DEV-05/07/08/09/10/16/18/19 — apeluri reale | — |
| **F8.5 — Handler Hardening** | ⏳ PENDING | DEV-06/12/13/14/15/20/22/23/25/26/27/28 | — |
| **F8.6 — Frontend Polish** | ⏳ PENDING | DEV-21 + envelope updates frontend | — |
| **F8.7 — Integration & E2E Tests** | ⏳ PENDING | Automatizare TS-01..TS-12, CI pipeline | — |
| **F8.8 — Staging + Production Validation** | ⏳ PENDING | Deploy staging, smoke full, sign-off final | — |

### Stare detaliată F5

**Implementat (Auth):**
- `POST /auth/register` ✅
- `POST /auth/login` ✅
- `POST /auth/refresh` ✅
- `POST /auth/mfa/verify` ✅
- `POST /auth/mfa/enable` ✅
- `POST /auth/forgot-password` ✅
- `POST /auth/reset-password` ✅
- `POST /auth/logout` ✅

**Implementat F5a (Goals + Today + Dashboard):**
- Goals: `POST /goals/analyze` ✅, `POST /goals/suggest-category` ✅, `POST /goals` ✅, `GET /goals` ✅, `GET /goals/:id` ✅, `GET /goals/:id/visualize` ✅
- Today: `GET /today` ✅, `POST /today/complete/:id` ✅, `POST /today/personal` ✅, `POST /context/energy` ✅
- Dashboard: `GET /dashboard` ✅ (Redis cached 5 min)
- `handlers.Handlers` struct extins cu `ai *ai.Client` ✅
- `api.Config` extins cu `AIClient *ai.Client` ✅
- `server.go` rutele Goals/Today/Dashboard înregistrate ✅

**Implementat F5b (SRM + Achievements + Profile + Admin):**
- SRM: `GET /srm/status/:goalId` ✅, `POST /srm/confirm-l2/:goalId` ✅, `POST /srm/confirm-l3/:goalId` ✅
- Achievements: `GET /achievements` ✅, `GET /ceremonies/:goalId` ✅, `POST /ceremonies/:id/view` ✅
- Profile: `GET /profile/activity` ✅, `PATCH /settings` ✅
- Admin: `GET /admin/stats` ✅, `GET /admin/users` ✅, `POST /admin/users/:id/deactivate` ✅

### Stare detaliată F6

Toate paginile și componentele **există în cod**:
- Pagini: `onboarding`, `today`, `goals`, `goals/[id]`, `dashboard`, `recap`, `profile`, `settings`, `achievements`, `admin`
- Componente: `AppShell`, `SRMWarning`, `CeremonyModal`, `ActivityHeatmap`, `ProgressCharts`, `GoalTabs`, `DashboardClientLayer`

**Stare:** Paginile cheamă endpoint-urile backend care lipsesc. Vor funcționa automat după F5.  
**Acțiune necesară:** Verificare build + audit după F5 complet.

---

## F0 — Reset ✅
Engine vechi eliminat, repo curățat.

## F1 — Schema DB ✅
32 tabele în schema `public`, migrări 001–013 active pe VPS.

## F2 — Auth CSS ✅
Pagini auth cu clase CSS standardizate.

---

## F0.1 — Cleanup ✅ (2026-04-06)

Fișiere moarte din era v10.x eliminate: `infra/init-db.sql`, `PLAN.md`, `PROMPTS.md`, `CHANGES.md`, `docs/DEMO_EXECUTION_PLAN.md`, `docs/framework_100_percent_implementation_playbook.md`, `docs/framework_workflow_deviations_stress_test.md`. README verificat — zero secrete expuse.

---

## F3 — Core Engine ✅ (2026-04-07)

**Componentele implementate:**

| C# | Componentă | Scope | Ce se codează |
|---|---|---|---|
| C1 | Structural Supremacy | FULL | Principiu — aplicat prin design |
| C2 | Behavior Model | FULL | Validare 5 BM la input |
| C5 | 30-Day Sprint | FULL | Expected(t) = t/30 |
| C6 | Normalization | FULL | Funcție clamp(x, 0, 1) |
| C7 | Priority Weight | SIMPLIFIED | Mapping Relevance → weight 1/2/3 |
| C8 | Priority Balance | SIMPLIFIED | Check Σ(weights) ≤ 7 |
| C14 | GO Validation | FULL | Deadline, BM, metric check |
| C19 | Sprint Structuring | FULL | Creare sprint 30 zile |
| C20 | Sprint Target | FULL | (Target-Progress)/Remaining × 0.80 |
| C21 | 80% Rule | FULL | Factor 0.80 |
| C24 | Progress Computation | FULL | Σ(completed×weight)/Σ(total) |
| C25 | Execution Variance | FULL | Drift_raw = Real - Expected |
| C26 | Drift Engine | SIMPLIFIED | Drift + trigger SRM L1 la -0.15/3d |
| C28 | Chaos Index | SIMPLIFIED | Formula 4 componente, 4 praguri |
| C30 | Consistency | SIMPLIFIED | active_days/eligible_days |
| C37 | Sprint Score | FULL | P×0.50 + C×0.30 + D×0.20 |
| C38 | GORI | SIMPLIFIED | Media sprint scores × continuity |

**Fișiere create:**
- `backend/internal/engine/engine.go` — scoring, drift, progress
- `backend/internal/engine/srm.go` — SRM L1/L2/L3 logic
- `backend/internal/engine/growth.go` — GORI, ceremonies, trajectories
- `backend/internal/engine/helpers.go` — clamp, grade, validare

---

## F4 — Scheduler + SRM Runtime ✅ (2026-04-07)

**Componentele implementate:**

| C# | Componentă | Scope | Ce se codează |
|---|---|---|---|
| C23 | Daily Stack | SIMPLIFIED | 1-3 tasks/zi per GO |
| C27 | Stagnation Detection | SIMPLIFIED | ≥5 zile fără progres |
| C32 | Adaptive Context | SIMPLIFIED | Doar Planned Pause |
| C33 | SRM (3 Levels) | SIMPLIFIED | L1 auto, L2 notify, L3 confirm |

**12 scheduler jobs** în `backend/internal/scheduler/scheduler.go`

---

## F5 — API Handlers ✅ (2026-04-18)

**Componentele de implementat:**

| C# | Componentă | Scope | Ce se codează |
|---|---|---|---|
| C3 | Max 3 Active GO | FULL | Validare la POST /goals |
| C4 | 365-Day Max | FULL | Validare la POST /goals |
| C9 | Semantic Parsing | SIMPLIFIED | AI extrage domain/direction/metric |
| C10 | BM Classification | SIMPLIFIED | AI atribuie BM cu confidence |
| C11 | Relevance Scoring | SIMPLIFIED | Scor fix la creare |
| C12 | Future Vault | FULL | Status WAITING |
| C13 | Relevance Thresholds | SIMPLIFIED | Floor 0.30, mapping weight |

**Fișiere de modificat:**
- `backend/internal/api/handlers/handlers.go` — adaugă handlers Goals/Today/Dashboard/SRM/Profile/Admin
- `backend/internal/api/server.go` — adaugă ai.Client în Config + înregistrează toate rutele

**Estimare:** 2×45 min (F5a + F5b)

---

## F6 — Frontend MVP ✅ (2026-04-18)

**Build:** `npm run build` — zero erori TypeScript, 23 pagini compilate.

**Pagini verificate:**
- `onboarding`, `today`, `goals`, `goals/[id]`, `dashboard`
- `recap`, `profile`, `settings`, `achievements`, `admin`

**Componente verificate:**
- `AppShell`, `SRMWarning`, `CeremonyModal`, `ActivityHeatmap`, `ProgressCharts`, `GoalTabs`, `DashboardClientLayer`

**Audit endpoint-uri:** Toate paginile cheamă endpoint-uri F5 existente (Goals, Today, Dashboard, SRM, Achievements, Profile, Settings, Admin).

---

## F7 — Smoke Test + Docs ✅

**Verificare:** build PASS, 12/12 unit tests PASS, API opacity CLEAN, smoke test plan documentat.  
**Docs:** README.md, CLAUDE.md, ROADMAP.md → v1.1.0. Raport: `docs/testing/smoke-test-report.md`.

---

## AUDIT-01 — System Alignment Audit ✅ (2026-04-28)

Audit complet codebase vs Framework Rev 5.6 + scenarii TS-01..TS-12.

**Output:** `docs/audit/AUDIT-01-deviation-report.md` (475 linii) — 28 devieri.

| Severitate | Count | Status |
|------------|-------|--------|
| CRITICAL | 8 | 0 RESOLVED |
| HIGH | 9 | 0 RESOLVED |
| MEDIUM | 7 | 0 RESOLVED |
| LOW | 4 | 1 RESOLVED (DEV-24) |

**SA-1..SA-7 status:** SA-4, SA-5, SA-7 ✅ rezolvate; SA-3 ⚠️ parțial; SA-1, SA-2, SA-6 ❌ deschise.

---

## F8 — MVP Fix Phase ⏳ PENDING (estimat 10–12h)

> **Obiectiv:** rezolvare completă a celor 27 devieri rămase din AUDIT-01; aducere MVP la calitate de lansare. Sursa de adevăr: `docs/fix-plan/00-PLAN-MASTER.md`.

### F8.1 — Schema Reconciliation
- **Owner:** DBA  |  **Estimare:** 60 min
- **Backlog:** DEV-02 (9 tabele lipsă), DEV-03 (users column drift), DEV-04 (sessions vs user_sessions), DEV-17 (achievements canonical name)
- **Output:** `backend/migrations/002_runtime_baseline.sql`, `backend/scripts/schema-check.sh`
- **Gate:** schema check exit 0; `db.RunMigrations()` curat

### F8.2 — Engine Restructure
- **Owner:** Backend Senior + Architect  |  **Estimare:** 90 min
- **Backlog:** DEV-11 (funcții engine lipsă)
- **Output:** `backend/internal/engine/{visualization,regulatory,evolution,clock}.go`
- **Funcții:** GenerateProgressVisualization, FreezeExpectedTrajectory, ApplySRMFallback, MarkEvolutionSprint, ComputeALIBreakdown, GenerateCompletionCeremony, CheckAndRecordRegressionEvent
- **Gate:** coverage engine ≥ 80%

### F8.3 — API Security Hardening
- **Owner:** Security Engineer  |  **Estimare:** 30 min
- **Backlog:** DEV-01 (`sprint_score` leak)
- **Output:** opacity_test.go cu 9 endpoints, `docs/integrations.md` API contract
- **Gate:** zero matches în opacity scan

### F8.4 — Scheduler Wiring
- **Owner:** Backend Senior  |  **Estimare:** 90 min
- **Backlog:** DEV-05, DEV-07, DEV-08, DEV-09, DEV-10, DEV-16, DEV-18, DEV-19
- **Output:** scheduler apelează engine functions; growth_trajectories/achievement_badges/srm_events populate corect
- **Gate:** integration test scheduler verde

### F8.5 — Handler Hardening
- **Owner:** Backend Senior  |  **Estimare:** 90 min
- **Backlog:** DEV-06, DEV-12, DEV-13, DEV-14, DEV-15, DEV-20, DEV-22, DEV-23, DEV-25, DEV-26, DEV-27, DEV-28
- **Output:** Day-1 viz fallback, SRM L3 freeze, energy DB write, ALI breakdown în SRM status
- **Gate:** TS-04, TS-06, TS-08, TS-11 manual green

### F8.6 — Frontend Polish
- **Owner:** Frontend Senior + UX  |  **Estimare:** 60 min
- **Backlog:** DEV-21 + frontend portion DEV-26/27/28
- **Output:** onboarding cu selecție durată; envelope-uri sincronizate
- **Gate:** `npm run build` clean, manual onboarding 4 durate

### F8.7 — Integration & E2E Tests
- **Owner:** Senior QA  |  **Estimare:** 120 min
- **Output:** TS-01..TS-12 automate, Playwright E2E, CI pipeline `.github/workflows/ci.yml`
- **Gate:** CI green, coverage handlers ≥ 70%, engine ≥ 80%

### F8.8 — Staging + Production Validation
- **Owner:** DevOps + QA + PM  |  **Estimare:** 90 min
- **Output:** deploy staging, smoke 12/12, perf baseline, security scan, 5 sign-off
- **Gate FINAL:** MVP autorizat pentru lansare

### Drum critic
```
F8.1 → F8.2 → F8.4 → F8.5 → F8.7 → F8.8
        ↓                ↑
       F8.3 (paralel) F8.6 (paralel)
```

---

# POST-MVP

| Fază | Conținut | Componente |
|---|---|---|
| F8 | Stripe monetizare | — |
| F9 | i18n complet (RO/EN/RU) | — |
| F10 | CI/CD + Testing (Unit/Integration/E2E) | — |
| F11 | Scale (caching, monitoring, security) | — |
| F12 | Componente POST_MVP | C15, C16, C17, C18, C29, C31, C34, C35, C36, C39, C40 |
| F13 | Growth (PWA, export, analytics) | — |
| F14 | Final Release (legal, launch) | — |

---

## Referințe

| Document | Conținut |
|---|---|
| `CLAUDE.md` | Context master + reguli sesiune |
| `MVP_SCOPE.md` | Matrice C1–C40 cu justificări |
| `PROMPTS_MVP.md` | Prompturi Claude Code F5a–F7 (actualizate) |
| `FORMULAS_QUICK_REFERENCE.md` | Formule engine |
| `docs/framework/rev5_6/` | Framework Rev 5.6 complet |

---

---

## REVYX S7 — Phase 4 Post-Launch ✅ (2026-05-06)

Platform REVYX live în producție (go/no-go gate trecut S6). S7 = expansiune funcțională.

| Livrabil | Status | Tech Spec |
|---|---|---|
| S7-1 Multi-Language UI (RO+RU) | ✅ | `TECH_SPEC_REVYX_multilang-ui_v1.0.0.md` |
| S7-2 ML Pricing Phase 3 (LightGBM) | ✅ | `TECH_SPEC_REVYX_ml-pricing-phase3_v1.0.0.md` |
| S7-3 Churn Prediction B2B | ✅ | `TECH_SPEC_REVYX_churn-prediction_v1.0.0.md` |
| S7-4 Market Expansion RO Rural + UA Diaspora | ✅ | `TECH_SPEC_REVYX_market-expansion-ro-ua_v1.0.0.md` |
| S7-5 Partnerships API (imobiliare.ro, storia.ro) | ✅ | `TECH_SPEC_REVYX_partnerships-api_v1.0.0.md` |
| S7-6 Billing & Metering Operational (Stripe) | ✅ | `TECH_SPEC_REVYX_billing-metering-operational_v1.0.0.md` |

**Migrări noi (S7):** 014 (users.language), 015 (feature_snapshots, market_zone_stats), 016 (churn_feature_snapshots, churn_predictions), 017 (properties ANCPI fields, uat_geocodes, exchange_rates), 018 (partners, partner_import_log, listing_fingerprints), 019 (tenants billing columns, invoices)

---

*v1.6.0 | 2026-05-06 — REVYX S7 Phase 4 Post-Launch: 6 livrabile tech spec complete*
