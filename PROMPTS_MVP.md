# PROMPTS_MVP.md — Claude Code Session Prompts

> **Versiune:** 1.0.0  
> **Reguli:**  
> - Fiecare prompt = o sesiune Claude Code nouă  
> - Copiază blocul ``` integral → paste în sesiune  
> - Ordine strictă. Nu sări.  
> - Prompturi mari sunt split în sub-sesiuni (a/b) pentru a evita timeout

---

## Index prompturi

| Prompt | Ce face | Model | Estimare |
|--------|---------|-------|----------|
| **F0.1** | Cleanup fișiere moarte | Sonnet | 15 min |
| **F3a** | Engine core: scoring + validare + helpers | Sonnet | 45 min |
| **F3b** | Engine SRM + Growth + teste | Sonnet | 45 min |
| **F4** | Scheduler: 12 cron jobs | Sonnet | 45 min |
| **F5a** | API: goals + today + dashboard | Sonnet | 45 min |
| **F5b** | API: SRM + achievements + AI + email + admin | Sonnet | 45 min |
| **F6a** | Frontend: onboarding + today + goals + dashboard | Sonnet | 60 min |
| **F6b** | Frontend: profile + settings + achievements + componente | Sonnet | 60 min |
| **F7** | Smoke test + docs finale | Sonnet | 30 min |

**Total: 10 sesiuni, ~7–8h**

**Regulă model:**  
Sonnet pentru toată implementarea — arhitectura e deja definită în MVP_SCOPE.md.  
Opus doar dacă apare o decizie arhitecturală neprevăzută în timpul sesiunii.

---

## F0.1 — Cleanup

**Model: Sonnet** | **Timp: 15 min** | **Fișiere atinse: ~10**

```
Read CLAUDE.md v1.0.0. Branch: claude/cleanup. Task: eliminare fișiere moarte.

## Ce faci

Pas 1 — Verifică existența fiecărui fișier înainte de ștergere:
  ls -la infra/init-db.sql PLAN.md docs/DEMO_EXECUTION_PLAN.md docs/framework_100_percent_implementation_playbook.md docs/framework_workflow_deviations_stress_test.md

Pas 2 — Pentru fiecare fișier care există, verifică că nu e importat:
  grep -r "init-db" backend/ infra/docker-compose*.yml
  grep -r "PLAN.md" backend/ frontend/ CLAUDE.md ROADMAP.md
  (etc. pentru fiecare)

Pas 3 — Șterge fișierele confirmate ca moarte:
  git rm infra/init-db.sql
  git rm PLAN.md
  git rm docs/DEMO_EXECUTION_PLAN.md
  git rm docs/framework_100_percent_implementation_playbook.md
  git rm docs/framework_workflow_deviations_stress_test.md

Pas 4 — Arhivează (nu șterge):
  mkdir -p docs/archive
  git mv CHANGES.md docs/archive/CHANGES_v10.md

Pas 5 — Dacă există PROMPTS.md (cel vechi):
  git mv PROMPTS.md docs/archive/PROMPTS_v10.md

Pas 6 — Verifică build:
  cd backend && go build ./... && cd ..

Pas 7 — Commit:
  git add -A
  git commit -m "chore: cleanup — remove dead files from v10.x era, archive CHANGES"
```

---

## F3a — Engine Core (scoring + validare)

**Model: Sonnet** | **Timp: 45 min** | **Fișiere create: 2**

```
Read CLAUDE.md v1.0.0. Branch: claude/core-engine. Task: engine scoring + validare GO.

## Citește max 3 fișiere
1. FORMULAS_QUICK_REFERENCE.md — formulele exacte
2. backend/internal/db/queries.go — funcțiile DB existente
3. backend/migrations/001_base_schema.sql — structura tabelelor core

## Creează: backend/internal/engine/engine.go

Package engine. Import: context, math, time, uuid, pgxpool.

### Funcții publice:

func ValidateGO(name string, bm string, startDate, endDate time.Time, activeCount int) error
  // C2: bm must be in {CREATE, INCREASE, REDUCE, MAINTAIN, EVOLVE}
  // C3: activeCount must be < 3
  // C4: endDate - startDate ≤ 365 days
  // C14: name not empty, bm not empty
  // Return nil if valid, error with message if not

func ComputeExpected(dayInSprint int) float64
  // C5: return float64(dayInSprint) / 30.0

func ComputeProgress(completedCheckpoints, totalCheckpoints int) float64
  // C24: if total == 0 return 0
  // return Clamp(float64(completed) / float64(total), 0, 1)

func ComputeDrift(realProgress, expected float64) float64
  // C25: return realProgress - expected (NOT clamped — drift can be negative)

func ComputeSprintTarget(annualTarget, currentProgress float64, sprintsRemaining int) float64
  // C20+C21: if remaining <= 0 return 0
  // brut = (annualTarget - currentProgress) / float64(remaining)
  // return brut * 0.80

func ComputeSprintScore(progressComp, consistencyComp, deviationComp float64) float64
  // C37: return Clamp(progressComp*0.50 + consistencyComp*0.30 + deviationComp*0.20, 0, 1)

func ComputeRelevance(impact, urgency, alignment, feasibility float64) float64
  // C11: return impact*0.35 + urgency*0.25 + alignment*0.25 + feasibility*0.15

func RelevanceToWeight(relevance float64) int
  // C7+C13: <0.40→1, <0.75→2, ≥0.75→3

func ScoreToGrade(score float64) string
  // ≥0.90 "A+", ≥0.80 "A", ≥0.65 "B", ≥0.45 "C", else "D"

## Creează: backend/internal/engine/helpers.go

func Clamp(x, min, max float64) float64
func ValidateBehaviorModel(bm string) bool
func CheckPriorityBalance(weights []int) bool
  // C8: sum(weights) ≤ 7

## NU implementa în această sesiune:
- SRM (F3b)
- GORI (F3b)
- Ceremonies (F3b)
- Queries DB — doar funcții pure (input → output)

## Verificare
  cd backend && go build ./internal/engine/... && go vet ./internal/engine/...

## Commit
  git add backend/internal/engine/engine.go backend/internal/engine/helpers.go
  git commit -m "feat(engine): core scoring — ValidateGO, Sprint Score, Drift, Progress (F3a)"
```

---

## F3b — Engine SRM + Growth + Teste

**Model: Sonnet** | **Timp: 45 min** | **Fișiere create: 3**

```
Read CLAUDE.md v1.0.0. Branch: claude/core-engine (continuare). Task: SRM logic + GORI + teste.

## Citește max 3 fișiere
1. backend/internal/engine/engine.go — funcțiile din F3a
2. backend/internal/engine/helpers.go — Clamp, ScoreToGrade
3. FORMULAS_QUICK_REFERENCE.md — formule SRM + GORI

## Creează: backend/internal/engine/srm.go

Package engine.

func IsDriftCritical(driftValues []float64) bool
  // C26: dacă ultimele 3 valori sunt toate < -0.15 → true

func ComputeChaosIndex(driftComp, stagnationComp, inconsistencyComp float64) float64
  // C28: contextDisruption = 0 (SIMPLIFIED)
  // return driftComp*0.30 + stagnationComp*0.25 + inconsistencyComp*0.25 + 0*0.20

func ChaosLevel(chaosIndex float64) string
  // <0.30 "GREEN", <0.40 "YELLOW", <0.60 "AMBER", ≥0.60 "RED"

func ComputeSRMFallback(hoursSince float64) string
  // ≥168 "PAUSE", ≥72 "L1", ≥24 "L2", else ""

## Creează: backend/internal/engine/growth.go

Package engine.

func ComputeGORI(sprintScores []float64, sprintsCompleted, sprintsTotal int) float64
  // C38 SIMPLIFIED: avg(sprintScores) * continuity
  // continuity = completed / max(total, 1) (excl SUSPENDED)
  // return Clamp(avg * continuity, 0, 1)

func GORIGrade(gori float64) string
  // ≥0.80 "Excellent", ≥0.60 "Good", ≥0.45 "Advisory", else "Early Recalibration"

func CeremonyTier(sprintScore float64) string
  // ≥0.90 "PLATINUM", ≥0.80 "GOLD", ≥0.65 "SILVER", else "BRONZE"

## Creează: backend/internal/engine/engine_test.go

Package engine_test. Minim 10 teste:

func TestValidateGO_ValidInput(t *testing.T)
func TestValidateGO_InvalidBM(t *testing.T)
func TestValidateGO_TooManyActive(t *testing.T)
func TestValidateGO_DurationOver365(t *testing.T)
func TestClamp_InRange(t *testing.T)
func TestClamp_Below(t *testing.T)
func TestClamp_Above(t *testing.T)
func TestScoreToGrade_AllBrackets(t *testing.T)
func TestSprintScore_Weights(t *testing.T)
  // Verifică: 1.0*0.50 + 1.0*0.30 + 1.0*0.20 == 1.0
  // Verifică: 0.0*0.50 + 0.0*0.30 + 0.0*0.20 == 0.0
func TestComputeGORI_Basic(t *testing.T)
func TestCeremonyTier_AllBrackets(t *testing.T)
func TestIsDriftCritical_ThreeDays(t *testing.T)

## Verificare
  cd backend && go test ./internal/engine/... -v -count=1

## După
  Actualizează CLAUDE.md: F3 → ✅
  Actualizează ROADMAP.md: F3 → ✅

## Commit
  git add backend/internal/engine/
  git commit -m "feat(engine): SRM logic, GORI, ceremonies, unit tests (F3b)"
```

---

## F4 — Scheduler

**Model: Sonnet** | **Timp: 45 min** | **Fișiere modificate: 1**

```
Read CLAUDE.md v1.0.0. Branch: claude/scheduler. Task: rewrite scheduler cu 12 jobs.

## Citește max 3 fișiere
1. backend/internal/engine/engine.go — funcțiile publice din F3
2. backend/internal/scheduler/scheduler.go — structura existentă (rewrite)
3. backend/migrations/001_base_schema.sql — structura daily_tasks, sprints

## Ce faci: rewrite backend/internal/scheduler/scheduler.go

Struct Scheduler { db *pgxpool.Pool, engine *engine.Engine }
func NewScheduler(db, engine) *Scheduler
func (s *Scheduler) Start() — înregistrează toate jobs cu robfig/cron

### 12 Jobs (fiecare e o metodă pe Scheduler):

1. jobGenerateDailyTasks — 00:01 UTC
   Query GO ACTIVE → pentru fiecare: INSERT 2 daily_tasks MAIN
   Text: template generic "Activitate pentru <goal_name>"

2. jobComputeDailyScore — 23:50 UTC
   Query GO ACTIVE → engine.ComputeProgress() + engine.ComputeDrift()
   INSERT go_scores (score_value, grade)

3. jobCheckDailyProgress — 23:55 UTC
   Query go_scores ultimele 3 zile per GO → engine.IsDriftCritical()
   Dacă true → INSERT srm_events level=L1

4. jobCloseExpiredSprints — 00:00 UTC
   WHERE status=ACTIVE AND end_date < TODAY
   engine.ComputeSprintScore() → INSERT sprint_results
   tier = engine.CeremonyTier() → INSERT ceremonies
   UPDATE sprints status=COMPLETED

5. jobStartNextSprints — 00:05 UTC
   GO ACTIVE fără sprint ACTIVE → INSERT sprints + 3 checkpoints

6. jobComputeWeeklyALI — duminică 03:00
   Placeholder: UPDATE go_metrics SET ali_value=...

7. jobRecalibrateRelevance — duminică 02:00
   engine.ComputeChaosIndex() → dacă ≥0.40 INSERT srm_events L2

8. jobCheckStagnation — 23:58 UTC
   COUNT completed=FALSE per GO, 5+ zile → INSERT stagnation_events

9. jobCheckSRMTimeouts — orar
   Query srm_events L3 neconfirmate → engine.ComputeSRMFallback()

10. jobGenerateCeremonies — 01:05 UTC
    Sprints COMPLETED fără ceremony → engine.CeremonyTier() → INSERT

11. jobDetectEvolution — 01:00 UTC
    Placeholder simplu: log "evolution check"

12. jobComputeGORI — 01:10 UTC
    engine.ComputeGORI() → UPDATE go_metrics

### Fiecare job:
  - context.WithTimeout 5 minute
  - logger.Info la start și finish cu count
  - Erori logate, nu panic

## NU implementa:
  - AI task generation (rămâne template generic, AI e în F5)
  - Physical Delta Safety Signal
  - SRM Timeout Protocol complet (24h/72h/7d re-propunere)
  - Reactivation ramp

## Verificare
  cd backend && go build ./internal/scheduler/...

## După
  Actualizează CLAUDE.md: F4 → ✅
  Actualizează README.md: secțiunea scheduler

## Commit
  git add backend/internal/scheduler/scheduler.go
  git commit -m "feat(scheduler): 12 cron jobs + SRM runtime (F4)"
```

---

## F5a — API Handlers Core

**Model: Sonnet** | **Timp: 45 min** | **Fișiere modificate: 2–3**

```
Read CLAUDE.md v1.0.0. Branch: claude/api-handlers. Task: endpoints goals + today + dashboard.

## Citește max 3 fișiere
1. backend/internal/api/server.go — routing existent
2. backend/internal/api/handlers/handlers.go — structura existentă
3. backend/internal/engine/engine.go — funcțiile F3

## Ce adaugi (la handlere existente, nu rewrite):

### POST /goals — CreateGoal
  Parse body: name, start_date, end_date, description, dominant_behavior_model
  engine.ValidateGO(name, bm, start, end, activeCount)
  relevance = engine.ComputeRelevance(0.7, 0.5, 0.6, 0.8) // defaults MVP
  weight = engine.RelevanceToWeight(relevance)
  INSERT global_objectives
  Dacă activeCount >= 3 → status=WAITING, return {vaulted: true}
  Altfel → status=ACTIVE, creare sprint + 3 checkpoints
  Return 201 + goal object

### GET /goals — ListGoals
  Query global_objectives WHERE user_id=X ORDER BY created_at DESC
  Return [{id, name, status, progress_pct, grade}]

### GET /goals/:id — GetGoalDetail
  Query goal + sprint activ + checkpoints
  SECURITATE: return DOAR {name, status, progress_pct, grade, sprint_day, sprint_total, checkpoints}
  NICIODATĂ: drift, chaos, weights, thresholds

### GET /goals/:id/visualize — GetVisualization
  Query growth_trajectories WHERE go_id=X ORDER BY snapshot_date
  Return [{date, actual_pct, expected_pct, delta}]

### GET /today — GetTodayTasks
  Query daily_tasks WHERE user_id=X AND task_date=TODAY
  Return [{id, text, completed, task_type}]

### POST /tasks/:id/complete — CompleteTask
  UPDATE daily_tasks SET completed=TRUE, completed_at=NOW() WHERE id=X AND user_id=Y
  Return 200

### GET /dashboard — GetDashboard
  Query goals + today task count per user
  Return [{goal_name, progress_pct, grade, tasks_today}]

## SECURITATE — verificare obligatorie la final:
  grep -rn "drift\|chaos_index\|weights\|threshold" backend/internal/api/handlers/
  # ZERO rezultate permise

## Verificare
  cd backend && go build ./...

## Commit
  git add backend/internal/api/
  git commit -m "feat(api): goals, today, dashboard endpoints (F5a)"
```

---

## F5b — API SRM + Achievements + AI + Email + Admin

**Model: Sonnet** | **Timp: 45 min** | **Fișiere modificate: 3–4**

```
Read CLAUDE.md v1.0.0. Branch: claude/api-handlers (continuare). Task: SRM + achievements + integrări.

## Citește max 3 fișiere
1. backend/internal/api/handlers/handlers.go — handlere din F5a
2. backend/internal/ai/ai.go — verifică structura existentă
3. backend/internal/email/email.go — verifică structura existentă

## Endpoints de adăugat:

### SRM
  GET /srm/status/:goalId → {srm_level, triggered_at} (NICIODATĂ chaos/drift)
  POST /srm/confirm-l2/:goalId → INSERT context_adjustments ENERGY_LOW
  POST /srm/confirm-l3/:goalId → UPDATE GO status=PAUSED

### Achievements
  GET /achievements → query achievement_badges WHERE user_id=X
  GET /ceremonies/:goalId → query ceremonies WHERE go_id=X ORDER BY created_at DESC LIMIT 1
  POST /ceremonies/:id/view → UPDATE viewed_at=NOW()

### Profile
  GET /profile/activity → query daily_metrics WHERE user_id=X last 365 days
  PATCH /settings → UPDATE users SET theme=X, locale=Y

### Admin (404 pentru non-admin, NU 403)
  GET /admin/stats → count users, goals, sprints, active SRM
  GET /admin/users → list users cu is_active, is_admin
  POST /admin/users/:id/deactivate → UPDATE is_active=FALSE
  POST /admin/db/reset → fn_dev_reset_data (doar APP_ENV=development)

## AI — backend/internal/ai/ai.go
  Verifică dacă fișierul există și ce conține.
  Dacă există cu funcții vechi → păstrează structura, adaugă/modifică:
    SuggestGOCategory(title, desc) → POST Haiku, 2s timeout, fallback empty
    AnalyzeGOText(text) → POST Haiku, 2s timeout, fallback empty
    GenerateTaskTexts(goal, milestone, count) → POST Haiku, 2s timeout, fallback template

  Endpoint API:
    POST /goals/suggest-category → ai.SuggestGOCategory → 200 always
    POST /goals/analyze → ai.AnalyzeGOText → 200 always

## Email — backend/internal/email/email.go
  Verifică dacă fișierul există.
  SendWelcome(email, name) — POST Resend API, fallback: log + return nil
  SendPasswordReset(email, token) — POST Resend API, fallback: log + return nil
  Wire: register handler → SendWelcome, forgot-password → SendPasswordReset

## SECURITATE check final:
  grep -rn "drift\|chaos_index\|weights\|threshold" backend/internal/api/handlers/

## După
  Actualizează CLAUDE.md: F5 → ✅
  Actualizează README.md: lista endpoints completă
  Actualizează infra/.env.example: ANTHROPIC_API_KEY, RESEND_API_KEY

## Commit
  git add backend/internal/
  git commit -m "feat(api): SRM, achievements, AI, email, admin (F5b)"
```

---

## F6a — Frontend Core Pages

**Model: Sonnet** | **Timp: 60 min** | **Fișiere create: 4–5 pagini**

```
Read CLAUDE.md v1.0.0. Branch: claude/frontend-mvp. Task: pagini core frontend.

## Citește max 3 fișiere
1. frontend/app/app/auth/login/page.tsx — referință stil
2. frontend/app/styles/globals.css — CSS variables
3. frontend/app/app/api/proxy/[...path]/route.ts — cum funcționează proxy

## Implementează 4 pagini:

### /onboarding (frontend/app/app/onboarding/page.tsx)
  Wizard 3 pași:
  1. Input: goal name + description
  2. Date picker: start_date, end_date + select BM (5 opțiuni)
  3. Confirmare → POST /api/proxy/goals → redirect /today
  State management cu useState. Fără AI suggestion (simplificare — se adaugă separat).

### /today (frontend/app/app/today/page.tsx)
  useEffect → fetch GET /api/proxy/today
  Task cards cu checkbox
  onClick checkbox → POST /api/proxy/tasks/:id/complete
  Actualizare state DUPĂ confirmare API (nu optimistic)
  Empty state: "Nu ai activități pentru azi"

### /goals (frontend/app/app/goals/page.tsx)
  useEffect → fetch GET /api/proxy/goals
  Card per goal: name, progress bar (width: progress_pct%), grade badge, status pill
  onClick → router.push(/goals/[id])

### /dashboard (frontend/app/app/dashboard/page.tsx)
  useEffect → fetch GET /api/proxy/dashboard
  Cards overview: goal name + progress bar + grade
  Secțiune: "X activități azi"
  SRM check: fetch GET /api/proxy/srm/status/:goalId per GO
  Dacă srm_level != "NONE" → afișează SRMWarning banner

## Stil: folosește clasele din app.css (page, greet, greet-title, etc.)
## NU implementa: i18n, charts Recharts, heatmap (sunt în F6b sau post-MVP)

## Verificare
  cd frontend/app && npm run build

## Commit
  git add frontend/app/app/onboarding/ frontend/app/app/today/ frontend/app/app/goals/ frontend/app/app/dashboard/
  git commit -m "feat(frontend): onboarding, today, goals, dashboard pages (F6a)"
```

---

## F6b — Frontend Extended + Componente

**Model: Sonnet** | **Timp: 60 min** | **Fișiere create: 5–6 pagini + 3 componente**

```
Read CLAUDE.md v1.0.0. Branch: claude/frontend-mvp (continuare). Task: pagini secundare + componente.

## Citește max 3 fișiere
1. frontend/app/app/today/page.tsx — referință stil din F6a
2. frontend/app/components/ — verifică ce componente există
3. frontend/app/styles/globals.css — CSS variables

## Pagini:

### /goals/[id] (frontend/app/app/goals/[id]/page.tsx)
  Fetch GET /api/proxy/goals/:id
  Secțiuni: progress %, grade mare, sprint day X/30, milestones list
  NU implementa chart Recharts — doar text + progress bar

### /profile (frontend/app/app/profile/page.tsx)
  Fetch GET /api/proxy/settings
  Avatar placeholder, full_name, email
  Statistici simple: "X obiective active", "Y zile consecutive"

### /settings (frontend/app/app/settings/page.tsx)
  Theme toggle dark/light → localStorage + PATCH /api/proxy/settings
  Language selector (ro/en/ru) → PATCH /api/proxy/settings
  Funcțional dar texte rămân în română (i18n e F9)

### /achievements (frontend/app/app/achievements/page.tsx)
  Fetch GET /api/proxy/achievements
  Grid cu 10 badge types, locked (gri) / unlocked (color)
  Folosește BADGE_META din fișierul existent dacă există

### /admin (frontend/app/app/admin/page.tsx)
  Verifică dacă pagina există deja — dacă da, păstreaz-o
  Dacă nu: creare simplificată cu 2 tab-uri: Stats + Users
  Guard: dacă !is_admin → redirect /dashboard

## Componente shared:

### frontend/app/components/SRMWarning.tsx
  Props: level (L1/L2/L3), goalId
  L1: banner galben "Ritmul tău a încetinit"
  L2: banner portocaliu + buton "Confirmă ajustare" → POST /srm/confirm-l2
  L3: banner roșu + buton "Confirmă reset" → POST /srm/confirm-l3

### frontend/app/components/layout/AppShell.tsx
  Verifică dacă există — dacă da, adaugă link-urile noi la sidebar
  Dacă nu: creare cu sidebar (links: Dashboard, Today, Goals, Profile, Settings, Achievements)
  Link Admin vizibil doar dacă is_admin=true

### frontend/app/components/CeremonyModal.tsx
  Props: tier (BRONZE/SILVER/GOLD/PLATINUM), goalName, score
  Modal cu color per tier, buton "Am văzut" → POST /ceremonies/:id/view

## Verificare
  cd frontend/app && npm run build

## După
  Actualizează CLAUDE.md: F6 → ✅
  Actualizează ROADMAP.md: F6 → ✅

## Commit
  git add frontend/app/
  git commit -m "feat(frontend): profile, settings, achievements, admin, components (F6b)"
```

---

## F7 — Smoke Test + Docs

**Model: Sonnet** | **Timp: 30 min** | **Fișiere modificate: 3 docs**

```
Read CLAUDE.md v1.0.0. Branch: claude/smoke-test. Task: verificare E2E + docs finale.

## Test — parcurge în ordine, notează rezultatul fiecărui pas

1. Auth:
  curl -s -o /dev/null -w "%{http_code}" -X POST localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d '{"email":"smoke@test.com","password":"Smoke1234!","full_name":"Smoke Test"}'
  # Așteptat: 201 sau 409

2. Login:
  TOKEN=$(curl -s -X POST localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"smoke@test.com","password":"Smoke1234!"}' | jq -r '.access_token')
  echo $TOKEN
  # Așteptat: JWT valid

3. Create GO:
  curl -s -H "Authorization: Bearer $TOKEN" -X POST localhost:8080/api/v1/goals \
    -H "Content-Type: application/json" \
    -d '{"name":"Test MVP Goal","start_date":"2026-04-07","end_date":"2026-10-07","dominant_behavior_model":"INCREASE"}'
  # Așteptat: 201 + goal object

4. Today:
  curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/today
  # Așteptat: 200 + tasks array

5. Dashboard:
  curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/dashboard
  # Așteptat: 200 + goals array cu progress_pct, grade

6. Opacity check:
  curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/goals | grep -i "drift\|chaos\|weight\|threshold"
  # Așteptat: ZERO rezultate

7. Admin guard:
  curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/admin/stats
  # Așteptat: 404 (nu 403, nu 401)

## Dacă buguri
  Minore → fix acum
  Majore → notează în docs/known-issues.md

## Actualizare docs finale

README.md:
  - Versiune: 1.1.0 (MVP complete)
  - Lista endpoints actualizată
  - Scheduler: 12 jobs

CLAUDE.md:
  - F3–F7 → ✅
  - Versiune: 1.1.0

ROADMAP.md:
  - F0–F7 → ✅

## Commit
  git add README.md CLAUDE.md ROADMAP.md
  git commit -m "docs: v1.1.0 — MVP complete, all phases F0–F7 verified"
```

---

*v1.0.0 | 2026-04-06*  
*10 sesiuni, ~7–8h total*  
*Toate pe Sonnet. Opus doar la decizie arhitecturală neprevăzută.*
