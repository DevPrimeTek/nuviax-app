# ROADMAP.md — NuviaX Implementation Roadmap

> **Versiune:** 1.0.0  
> **Actualizat:** 2026-04-06  
> **Framework:** NuviaX Growth Framework Rev 5.6 (C1–C40)  
> **MVP Scope:** 29 componente (14 FULL + 15 SIMPLIFIED), 11 POST_MVP

---

## Stare curentă

| Fază | Status | Componente acoperite |
|---|---|---|
| F0 — Reset | ✅ | — |
| F1 — Schema DB | ✅ | Schema pentru toate C1–C40 |
| F2 — Auth CSS | ✅ | — |
| F0.1 — Cleanup | ✅ | — |
| F3 — Core Engine | ✅ | C1–C8, C14, C19–C21, C24–C26, C28, C30, C37, C38 |
| F4 — Scheduler + SRM | ⏳ | C23, C27, C33, C32 (Pause only) |
| F5 — API Handlers | ⏳ | C9, C10, C11, C12, C13, C3 (Vault) |
| F6 — Frontend MVP | ⏳ | UI pentru toate componentele active |
| F7 — Smoke Test | ⏳ | Verificare E2E |

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

**Fișiere de creat:**
- `backend/internal/engine/engine.go` — scoring, drift, progress
- `backend/internal/engine/srm.go` — SRM L1/L2/L3 logic
- `backend/internal/engine/growth.go` — GORI, ceremonies, trajectories
- `backend/internal/engine/helpers.go` — clamp, grade, validare

**Estimare:** 2–3h

---

## F4 — Scheduler + SRM Runtime ⏳

**Componentele implementate:**

| C# | Componentă | Scope | Ce se codează |
|---|---|---|---|
| C23 | Daily Stack | SIMPLIFIED | 1-3 tasks/zi per GO |
| C27 | Stagnation Detection | SIMPLIFIED | ≥5 zile fără progres |
| C32 | Adaptive Context | SIMPLIFIED | Doar Planned Pause |
| C33 | SRM (3 Levels) | SIMPLIFIED | L1 auto, L2 notify, L3 confirm |

**12 scheduler jobs** — detalii în `PROMPTS_MVP.md`

**Fișiere de creat/modificat:**
- `backend/internal/scheduler/scheduler.go` — rewrite complet

**Estimare:** 2h

---

## F5 — API Handlers ⏳

**Componentele implementate:**

| C# | Componentă | Scope | Ce se codează |
|---|---|---|---|
| C3 | Max 3 Active GO | FULL | Validare la POST /goals |
| C4 | 365-Day Max | FULL | Validare la POST /goals |
| C9 | Semantic Parsing | SIMPLIFIED | AI extrage domain/direction/metric |
| C10 | BM Classification | SIMPLIFIED | AI atribuie BM cu confidence |
| C11 | Relevance Scoring | SIMPLIFIED | Scor fix la creare |
| C12 | Future Vault | FULL | Status WAITING |
| C13 | Relevance Thresholds | SIMPLIFIED | Floor 0.30, mapping weight |

**Endpoints:** auth, goals, today, SRM, dashboard, profile, achievements, admin  
**Integrări:** Claude Haiku (2s timeout + fallback), Resend (fallback)

**Fișiere de creat/modificat:**
- `backend/internal/api/handlers/*.go`
- `backend/internal/ai/ai.go`
- `backend/internal/email/email.go`

**Estimare:** 2–3h

---

## F6 — Frontend MVP ⏳

**Pagini:** onboarding, today, goals, goals/[id], dashboard, recap, profile, settings, achievements, admin

**Estimare:** 3–4h

---

## F7 — Smoke Test + Docs ⏳

**Verificare:** auth → create GO → sprint → tasks → complete → score → dashboard  
**Docs:** README.md, CLAUDE.md final

**Estimare:** 1h

---

## Ordinea de execuție

```
F0.1 (cleanup) → F3 (engine) → F4 (scheduler) → F5 (API) → F6 (frontend) → F7 (test)
```

**Timp total estimat: 11–14h**

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
| `PROMPTS_MVP.md` | Prompturi Claude Code F0.1–F7 |
| `FORMULAS_QUICK_REFERENCE.md` | Formule engine |
| `docs/framework/rev5_6/` | Framework Rev 5.6 complet |

---

*v1.0.0 | 2026-04-06*
