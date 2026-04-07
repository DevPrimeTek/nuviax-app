# PROMPTS_MVP.md — Claude Code Session Prompts

> **Versiune:** 1.0.0  
> **Reguli:**  
> - Fiecare prompt = o sesiune Claude Code nouă  
> - Copiază blocul ``` integral → paste în sesiune  
> - Ordine strictă: F0.1 → F3a → F3b → F4 → F5a → F5b → F6a → F6b → F7  
> - Prompturi split în sub-sesiuni de max 45–60 min pentru a evita timeout

---

## Regulă GLOBALĂ — se aplică la FIECARE prompt

**ÎNAINTE de `git commit` în ORICE sesiune, execută obligatoriu PRE-COMMIT CHECKLIST din CLAUDE.md secțiunea 6. Aceasta include:**
1. Verificare secrete (grep chei API, parole în docs)
2. Verificare API opacity (grep drift/chaos în handlers)
3. Actualizare CLAUDE.md (marchează faza ✅ dacă e completă)
4. Actualizare ROADMAP.md (marchează faza ✅)
5. Actualizare README.md (endpoints, env vars, scheduler, versiune — verifică că NU conține secrete)
6. Actualizare docs/ afectate
7. Verificare versiuni identice în CLAUDE.md + README.md + ROADMAP.md
8. Cleanup imports/logs moarte

**Dacă sari peste PRE-COMMIT CHECKLIST = commit invalid.**

---

## Index

| # | Ce face | Model | Timp |
|---|---------|-------|------|
| F0.1 | Cleanup fișiere moarte + securizare README | Sonnet | 15 min |
| F3a | Engine core: scoring + validare + helpers | Sonnet | 45 min |
| F3b | Engine SRM + Growth + teste | Sonnet | 45 min |
| F4 | Scheduler: 12 cron jobs | Sonnet | 45 min |
| F5a | API: goals + today + dashboard | Sonnet | 45 min |
| F5b | API: SRM + achievements + AI + email + admin | Sonnet | 45 min |
| F6a | Frontend: onboarding + today + goals + dashboard | Sonnet | 60 min |
| F6b | Frontend: profile + settings + achievements + componente | Sonnet | 60 min |
| F7 | Smoke test + docs finale | Sonnet | 30 min |

**Model:** Sonnet pentru tot. Opus doar la decizie arhitecturală neprevăzută.  
**Total: 10 sesiuni, ~7–8h**

---

## F0.1 — Cleanup + Securizare README

**Model: Sonnet** | **Max 15 min**

```
Read CLAUDE.md v1.0.0. Branch: claude/cleanup. Task: cleanup fișiere moarte + securizare README.

## Pas 1 — Verifică și șterge fișiere moarte
Verifică existența, apoi grep că nu sunt referite:

  ls -la infra/init-db.sql PLAN.md docs/DEMO_EXECUTION_PLAN.md docs/framework_100_percent_implementation_playbook.md docs/framework_workflow_deviations_stress_test.md 2>/dev/null

Pentru fiecare care există:
  grep -r "numele_fisierului" backend/ frontend/ *.md
Dacă zero referințe → git rm <fișier>

## Pas 2 — Arhivează
  mkdir -p docs/archive
  git mv CHANGES.md docs/archive/CHANGES_v10.md 2>/dev/null || true
  git mv PROMPTS.md docs/archive/PROMPTS_v10.md 2>/dev/null || true

## Pas 3 — Securizare README.md
Verifică README.md pentru informații sensibile:
  grep -n "sk-ant-\|re_[A-Za-z]\|API_KEY.*=.*[A-Za-z0-9]\|password\|PRIVATE.*KEY\|162\.\|10\.\|192\.168" README.md
Dacă găsește ceva → ELIMINĂ. README nu trebuie să conțină:
  - Chei API reale sau parțiale
  - Parole
  - IP-uri server
  - Paths absolute de pe server
  - Orice credențiale

## Pas 4 — Build check
  cd backend && go build ./... && cd ..

## PRE-COMMIT CHECKLIST (OBLIGATORIU — din CLAUDE.md secțiunea 6)
  grep -rn "sk-ant-\|re_[A-Za-z0-9]\{20,\}" README.md CLAUDE.md ROADMAP.md infra/.env.example docs/ 2>/dev/null
  # ZERO rezultate

Actualizează CLAUDE.md:
  - Secțiunea 1: F0.1 → ✅
  - Secțiunea 5: elimină din lista de cleanup fișierele deja șterse

Actualizează ROADMAP.md:
  - F0.1 → ✅ cu data de azi

Verifică versiuni:
  grep "Versiune\|versiune" CLAUDE.md README.md ROADMAP.md

## Commit
  git add -A
  git commit -m "chore: cleanup v10.x dead files, secure README, archive CHANGES (F0.1)"
```

---

## F3a — Engine Core

**Model: Sonnet** | **Max 45 min**

```
Read CLAUDE.md v1.0.0. Branch: claude/core-engine. Task: engine scoring + validare GO.

## Citește max 3 fișiere
1. FORMULAS_QUICK_REFERENCE.md — formulele exacte
2. backend/internal/db/queries.go — funcțiile DB existente
3. backend/migrations/001_base_schema.sql — structura tabelelor core

## Creează: backend/internal/engine/engine.go

Package engine. Import: context, math, time, uuid, pgxpool.

func ValidateGO(name string, bm string, startDate, endDate time.Time, activeCount int) error
  // C2: bm ∈ {CREATE, INCREASE, REDUCE, MAINTAIN, EVOLVE}
  // C3: activeCount < 3
  // C4: endDate - startDate ≤ 365 days
  // C14: name not empty, bm not empty

func ComputeExpected(dayInSprint int) float64
  // C5: float64(dayInSprint) / 30.0

func ComputeProgress(completedCheckpoints, totalCheckpoints int) float64
  // C24: Clamp(float64(completed) / float64(total), 0, 1)

func ComputeDrift(realProgress, expected float64) float64
  // C25: realProgress - expected (NOT clamped)

func ComputeSprintTarget(annualTarget, currentProgress float64, sprintsRemaining int) float64
  // C20+C21: (annualTarget - currentProgress) / remaining × 0.80

func ComputeSprintScore(progressComp, consistencyComp, deviationComp float64) float64
  // C37: Clamp(progress×0.50 + consistency×0.30 + deviation×0.20, 0, 1)

func ComputeRelevance(impact, urgency, alignment, feasibility float64) float64
  // C11: impact×0.35 + urgency×0.25 + alignment×0.25 + feasibility×0.15

func RelevanceToWeight(relevance float64) int
  // C7+C13: <0.40→1, <0.75→2, ≥0.75→3

func ScoreToGrade(score float64) string
  // ≥0.90 "A+", ≥0.80 "A", ≥0.65 "B", ≥0.45 "C", else "D"

## Creează: backend/internal/engine/helpers.go

func Clamp(x, min, max float64) float64
func ValidateBehaviorModel(bm string) bool
func CheckPriorityBalance(weights []int) bool  // C8: sum ≤ 7

## NU implementa aici: SRM, GORI, Ceremonies, queries DB — sunt în F3b

## Verificare
  cd backend && go build ./internal/engine/... && go vet ./internal/engine/...

## PRE-COMMIT CHECKLIST (OBLIGATORIU)
  grep -rn "sk-ant-\|re_[A-Za-z0-9]\{20,\}" README.md CLAUDE.md 2>/dev/null  # ZERO
  Actualizează FORMULAS_QUICK_REFERENCE.md dacă formulele implementate diferă
  Verifică versiuni: grep "Versiune" CLAUDE.md README.md ROADMAP.md

## Commit
  git add backend/internal/engine/engine.go backend/internal/engine/helpers.go
  git commit -m "feat(engine): core scoring — ValidateGO, Sprint Score, Drift, Progress (F3a)"
```

---

## F3b — Engine SRM + Growth + Teste

**Model: Sonnet** | **Max 45 min**

```
Read CLAUDE.md v1.0.0. Branch: claude/core-engine (continuare). Task: SRM + GORI + teste.

## Citește max 3 fișiere
1. backend/internal/engine/engine.go — funcțiile din F3a
2. backend/internal/engine/helpers.go — Clamp, ScoreToGrade
3. FORMULAS_QUICK_REFERENCE.md — formule SRM + GORI

## Creează: backend/internal/engine/srm.go

func IsDriftCritical(driftValues []float64) bool
  // C26: ultimele 3 valori toate < -0.15 → true

func ComputeChaosIndex(driftComp, stagnationComp, inconsistencyComp float64) float64
  // C28: drift×0.30 + stagnation×0.25 + inconsistency×0.25 + 0×0.20

func ChaosLevel(chaosIndex float64) string
  // <0.30 GREEN, <0.40 YELLOW, <0.60 AMBER, ≥0.60 RED

func ComputeSRMFallback(hoursSince float64) string
  // ≥168 PAUSE, ≥72 L1, ≥24 L2, else ""

## Creează: backend/internal/engine/growth.go

func ComputeGORI(sprintScores []float64, completed, total int) float64
  // C38: Clamp(avg(scores) × (completed/max(total,1)), 0, 1)

func GORIGrade(gori float64) string
func CeremonyTier(sprintScore float64) string
  // ≥0.90 PLATINUM, ≥0.80 GOLD, ≥0.65 SILVER, else BRONZE

## Creează: backend/internal/engine/engine_test.go — minim 12 teste
  TestValidateGO_ValidInput, TestValidateGO_InvalidBM, TestValidateGO_TooManyActive,
  TestValidateGO_DurationOver365, TestClamp_InRange, TestClamp_Below, TestClamp_Above,
  TestScoreToGrade_AllBrackets, TestSprintScore_Weights, TestComputeGORI_Basic,
  TestCeremonyTier_AllBrackets, TestIsDriftCritical_ThreeDays

## Verificare
  cd backend && go test ./internal/engine/... -v -count=1

## PRE-COMMIT CHECKLIST (OBLIGATORIU)
  Actualizează CLAUDE.md: secțiunea 1 → F3 → ✅
  Actualizează ROADMAP.md: F3 → ✅ cu data
  Verifică versiuni: grep "Versiune" CLAUDE.md README.md ROADMAP.md

## Commit
  git add backend/internal/engine/
  git commit -m "feat(engine): SRM logic, GORI, ceremonies, 12 unit tests (F3b)"
```

---

## F4 — Scheduler

**Model: Sonnet** | **Max 45 min**

```
Read CLAUDE.md v1.0.0. Branch: claude/scheduler. Task: rewrite scheduler 12 jobs.

## IMPORTANT: Citește CLAUDE.md secțiunea 8 (Integrări existente) ÎNAINTE de a scrie cod.
## ai.go și email.go DEJA EXISTĂ — nu le rescrie, integrează-le.

## Citește max 3 fișiere
1. backend/internal/ai/ai.go — funcțiile AI existente (GenerateTaskTexts, AnalyzeGO)
2. backend/internal/email/email.go — funcțiile email existente (SendWelcome, SendSprintComplete)
3. backend/internal/scheduler/scheduler.go — structura existentă (rewrite)

## Rewrite: backend/internal/scheduler/scheduler.go

Struct Scheduler { db *pgxpool.Pool, ai *ai.Client, email *email.Client }
NewScheduler: acceptă ai.Client și email.Client (pot fi nil — graceful degradation)
12 jobs — fiecare cu context.WithTimeout(5min) + logger.Info start/finish:

1. jobGenerateDailyTasks (00:01) — GO ACTIVE → pentru fiecare:
   DACĂ s.ai != nil:
     texts, err := s.ai.GenerateTaskTexts(ctx, goalName, checkpointName, sprintNum, 2)
     Dacă err != nil → fallback pe template generic
   ALTFEL:
     texts = []string{"Activitate pentru <goalName>", "Continuare <checkpointName>"}
   INSERT daily_tasks cu textele generate

2. jobComputeDailyScore (23:50) — engine.ComputeProgress+Drift → INSERT go_scores
3. jobCheckDailyProgress (23:55) — engine.IsDriftCritical → INSERT srm_events L1

4. jobCloseExpiredSprints (00:00) — ACTIVE+expired →
   SprintScore → INSERT sprint_results → CeremonyTier → INSERT ceremonies → COMPLETED
   DACĂ s.email != nil:
     go s.email.SendSprintComplete(ctx, userEmail, userName, goalName, grade, sprintNum)
   (fire-and-forget, nu blochează job-ul)

5. jobStartNextSprints (00:05) — GO ACTIVE fără sprint → INSERT sprints+checkpoints
6. jobComputeWeeklyALI (duminică 03:00) — placeholder UPDATE go_metrics
7. jobRecalibrateRelevance (duminică 02:00) — ChaosIndex → dacă ≥0.40 INSERT srm_events L2
8. jobCheckStagnation (23:58) — 5+ zile fără completed → INSERT stagnation_events
9. jobCheckSRMTimeouts (orar) — L3 neconfirmate → ComputeSRMFallback
10. jobGenerateCeremonies (01:05) — sprints COMPLETED fără ceremony → INSERT
11. jobDetectEvolution (01:00) — placeholder log
12. jobComputeGORI (01:10) — engine.ComputeGORI → UPDATE go_metrics

## Verificare
  cd backend && go build ./internal/scheduler/...

## PRE-COMMIT CHECKLIST (OBLIGATORIU)
  Actualizează CLAUDE.md: F4 → ✅
  Actualizează ROADMAP.md: F4 → ✅
  Actualizează README.md: adaugă secțiune scheduler cu 12 jobs + cron times
  Verifică versiuni: grep "Versiune" CLAUDE.md README.md ROADMAP.md

## Commit
  git add backend/internal/scheduler/
  git commit -m "feat(scheduler): 12 cron jobs + SRM runtime (F4)"
```

---

## F5a — API Core

**Model: Sonnet** | **Max 45 min**

```
Read CLAUDE.md v1.0.0. Branch: claude/api-handlers. Task: endpoints goals + today + dashboard + AI validare GO.

## IMPORTANT: Citește CLAUDE.md secțiunea 8 (Integrări existente).
## ai.go și email.go DEJA EXISTĂ — NU le rescrie. Integrează-le în handlers.

## Citește max 3 fișiere
1. backend/internal/ai/ai.go — funcțiile AI existente (AnalyzeGO, SuggestGOCategory)
2. backend/internal/api/handlers/handlers.go — handlere existente (verifică ce există deja)
3. backend/internal/engine/engine.go — funcții F3

## Handlers struct trebuie să accepte ai + email (pot fi nil):
  type Handlers struct { db, redis, engine, ai *ai.Client, email *email.Client, encKey }
  Dacă ai==nil → fallback rule-based. Dacă email==nil → skip email.

## Adaugă la routing (păstrează ce funcționează deja):

### POST /goals/analyze — AI Validare GO (C9/C10)
  Verifică dacă handlerul AnalyzeGO DEJA EXISTĂ — dacă da, păstrează-l!
  Flux: parse text → dacă h.ai != nil: ai.AnalyzeGO(ctx, text) cu 2s timeout
  → return {needs_clarification, question, hint, source:"ai"}
  Fallback: rule-based (vagueTerms, measurable, time patterns)

### POST /goals/suggest-category — AI Sugestie Categorie
  Dacă h.ai != nil: ai.SuggestGOCategory(ctx, title, desc) cu 2s timeout
  Return {category, confidence} sau {category:"", confidence:0} la fallback

### POST /goals — CreateGoal (cu AI pre-validare)
  ValidateGO, ComputeRelevance, RelevanceToWeight
  Dacă activeCount >= 3 → WAITING + vaulted:true
  Altfel → ACTIVE + sprint + 3 checkpoints
  Return 201 (FĂRĂ drift/chaos/weights)

### GET /goals — ListGoals: [{id, name, status, progress_pct, grade}]
### GET /goals/:id — GetGoalDetail: DOAR progress_pct, grade, status, sprint info
### GET /goals/:id/visualize — growth_trajectories data
### GET /today — daily_tasks WHERE today AND user_id
### POST /tasks/:id/complete — UPDATE completed=TRUE
### GET /dashboard — goals overview per user

## SECURITATE — check OBLIGATORIU:
  grep -rn "drift\|chaos_index\|weights\|threshold" backend/internal/api/handlers/
  # ZERO rezultate

## PRE-COMMIT CHECKLIST (OBLIGATORIU)
  Actualizează README.md: adaugă endpoints goals, today, dashboard
  Verifică secrete: grep "sk-ant-\|password" README.md  # ZERO
  Verifică versiuni: grep "Versiune" CLAUDE.md README.md ROADMAP.md

## Commit
  git add backend/internal/api/
  git commit -m "feat(api): goals, today, dashboard endpoints (F5a)"
```

---

## F5b — API Extended

**Model: Sonnet** | **Max 45 min**

```
Read CLAUDE.md v1.0.0. Branch: claude/api-handlers (continuare). Task: SRM + achievements + email wiring + admin.

## IMPORTANT: ai.go și email.go DEJA EXISTĂ cu funcții complete.
## NU le rescrie. Verifică că sunt WIRED corect în handlers și server.go.

## Citește max 3 fișiere
1. backend/internal/api/handlers/handlers.go — din F5a
2. backend/internal/ai/ai.go — VERIFICĂ funcțiile existente, nu rescrie
3. backend/internal/email/email.go — VERIFICĂ funcțiile existente, nu rescrie

## Endpoints:
GET /srm/status/:goalId → {srm_level, triggered_at} NICIODATĂ chaos/drift
POST /srm/confirm-l2/:goalId → INSERT context_adjustments ENERGY_LOW
POST /srm/confirm-l3/:goalId → UPDATE GO status=PAUSED
GET /achievements → badge list
GET /ceremonies/:goalId → ultima ceremonie
POST /ceremonies/:id/view → UPDATE viewed_at
GET /profile/activity → daily_metrics 365 zile
PATCH /settings → theme, locale
GET /admin/stats → 404 non-admin
GET /admin/users → 404 non-admin
POST /admin/users/:id/deactivate → 404 non-admin

## WIRING EMAIL (verifică, nu rescrie):
Handler Register (DEJA EXISTĂ) trebuie să apeleze:
  if h.email != nil {
    go h.email.SendWelcome(context.Background(), email, name)
  }
  Verifică în handlers.go că acest apel EXISTĂ. Dacă nu → adaugă-l.

Handler ForgotPassword (DEJA EXISTĂ) trebuie să apeleze:
  if h.email != nil {
    go h.email.SendPasswordReset(context.Background(), email, resetLink)
  }
  Verifică în handlers.go că acest apel EXISTĂ. Dacă nu → adaugă-l.

## WIRING AI (verifică, nu rescrie):
Verifică că server.go inițializează:
  aiClient, _ := ai.New()  // nil dacă ANTHROPIC_API_KEY lipsește
  emailClient, _ := email.New()  // nil dacă RESEND_API_KEY lipsește
  handlers := NewHandlers(db, redis, engine, aiClient, emailClient, encKey)

## ENV VARS necesare (DEJA pe server, doar verifică .env.example):
  ANTHROPIC_API_KEY=sk-ant-...YOUR_KEY_HERE
  RESEND_API_KEY=re_...YOUR_KEY_HERE
  EMAIL_FROM=NuviaX <noreply@nuviax.app>

## SECURITATE check:
  grep -rn "drift\|chaos_index\|weights\|threshold" backend/internal/api/handlers/  # ZERO

## PRE-COMMIT CHECKLIST (OBLIGATORIU)
  Actualizează CLAUDE.md: F5 → ✅
  Actualizează ROADMAP.md: F5 → ✅
  Actualizează README.md: endpoints complete + env vars (ANTHROPIC_API_KEY, RESEND_API_KEY)
  Actualizează infra/.env.example dacă env vars noi
  Verifică secrete: grep "sk-ant-" README.md infra/.env.example  # ZERO chei reale

## Commit
  git add backend/internal/
  git commit -m "feat(api): SRM, achievements, AI, email, admin (F5b)"
```

---

## F6a — Frontend Core

**Model: Sonnet** | **Max 60 min**

```
Read CLAUDE.md v1.0.0. Branch: claude/frontend-mvp. Task: 4 pagini core.

## Citește max 3 fișiere
1. frontend/app/app/auth/login/page.tsx — referință stil
2. frontend/app/styles/globals.css — CSS vars
3. frontend/app/app/api/proxy/[...path]/route.ts — API proxy

## 4 pagini:

/onboarding — wizard 4 pași cu AI:
  Pas 1: input text goal name + description
  Pas 2: POST /api/proxy/goals/analyze → AI validează GO
    Dacă needs_clarification=true → afișează question + hint → user reformulează → re-submit
    Dacă needs_clarification=false → trecem la pas 3
    Dacă AI nu răspunde în 2s → skip validare, trecem direct la pas 3
  Pas 3: date (start, end) + select BM (5 opțiuni) + AI suggest-category (pre-fill)
    POST /api/proxy/goals/suggest-category cu title → pre-selectează categoria
    User poate override categoria sugerată
  Pas 4: confirmare → POST /api/proxy/goals → redirect /today

/today — GET /today → task cards cu checkbox → POST /tasks/:id/complete (update DUPĂ confirmare API)
/goals — GET /goals → card list cu progress bar + grade → click → /goals/[id]
/dashboard — GET /dashboard → cards overview + tasks count + SRM banner dacă activ

## Stil: clase din app.css. NU implementa: i18n, charts, heatmap

## Verificare
  cd frontend/app && npm run build

## PRE-COMMIT CHECKLIST (OBLIGATORIU)
  Verifică secrete: grep -rn "sk-ant-\|API_KEY\|password" frontend/app/app/  # ZERO
  Verifică versiuni: grep "Versiune" CLAUDE.md README.md ROADMAP.md

## Commit
  git add frontend/app/app/onboarding/ frontend/app/app/today/ frontend/app/app/goals/ frontend/app/app/dashboard/
  git commit -m "feat(frontend): onboarding, today, goals, dashboard (F6a)"
```

---

## F6b — Frontend Extended

**Model: Sonnet** | **Max 60 min**

```
Read CLAUDE.md v1.0.0. Branch: claude/frontend-mvp (continuare). Task: pagini secundare + componente.

## Citește max 3 fișiere
1. frontend/app/app/today/page.tsx — referință din F6a
2. frontend/app/components/ — ce există
3. frontend/app/styles/globals.css — CSS vars

## Pagini:
/goals/[id] — detaliu: progress %, grade, sprint day X/30, milestones
/profile — avatar, name, stats simple
/settings — theme toggle + language selector
/achievements — badge grid 10 types
/admin — verifică dacă există, dacă nu: 2 tab-uri Stats+Users, guard is_admin

## Componente:
SRMWarning.tsx — banner L1/L2/L3 cu buton confirm
AppShell.tsx — verifică, adaugă links noi la sidebar
CeremonyModal.tsx — modal per tier, buton "Am văzut"

## Verificare
  cd frontend/app && npm run build

## PRE-COMMIT CHECKLIST (OBLIGATORIU)
  Actualizează CLAUDE.md: F6 → ✅
  Actualizează ROADMAP.md: F6 → ✅
  Verifică secrete în frontend: grep -rn "sk-ant-\|API_KEY" frontend/  # ZERO
  Verifică versiuni: grep "Versiune" CLAUDE.md README.md ROADMAP.md

## Commit
  git add frontend/app/
  git commit -m "feat(frontend): profile, settings, achievements, admin, components (F6b)"
```

---

## F7 — Smoke Test + Docs Finale

**Model: Sonnet** | **Max 30 min**

```
Read CLAUDE.md v1.0.0. Branch: claude/smoke-test. Task: E2E test + docs finale.

## Teste curl — execută și notează rezultatul

  # Auth
  curl -s -o /dev/null -w "%{http_code}" -X POST localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d '{"email":"smoke@test.com","password":"Smoke1234!","full_name":"Smoke Test"}'
  # → 201 sau 409

  TOKEN=$(curl -s -X POST localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"smoke@test.com","password":"Smoke1234!"}' | jq -r '.access_token')

  # Create GO
  curl -s -H "Authorization: Bearer $TOKEN" -X POST localhost:8080/api/v1/goals \
    -H "Content-Type: application/json" \
    -d '{"name":"Test MVP","start_date":"2026-04-07","end_date":"2026-10-07","dominant_behavior_model":"INCREASE"}'
  # → 201

  # Dashboard
  curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/dashboard
  # → 200 cu goals

  # Opacity check CRITIC
  curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/goals | grep -i "drift\|chaos\|weight\|threshold"
  # → ZERO rezultate

  # Admin guard
  curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/admin/stats
  # → 404

## Dacă buguri minore → fix acum. Majore → docs/known-issues.md

## PRE-COMMIT CHECKLIST FINAL (OBLIGATORIU)

  # Securitate completă
  grep -rn "sk-ant-\|re_[A-Za-z0-9]\{20,\}\|password.*=.*[A-Za-z]" README.md CLAUDE.md ROADMAP.md infra/.env.example 2>/dev/null
  # ZERO

  # API opacity
  grep -rn "drift\|chaos_index\|weights\|threshold" backend/internal/api/handlers/
  # ZERO

  # Actualizare docs finale
  Actualizează README.md: versiune 1.1.0, endpoints complete, scheduler 12 jobs
  Actualizează CLAUDE.md: F3–F7 → ✅, versiune 1.1.0
  Actualizează ROADMAP.md: F0–F7 → ✅, versiune 1.1.0

  # Verificare versiuni
  grep "Versiune\|versiune\|1.1.0" CLAUDE.md README.md ROADMAP.md
  # Toate trei: 1.1.0

## Commit
  git add README.md CLAUDE.md ROADMAP.md docs/
  git commit -m "docs: v1.1.0 — MVP complete, all F0–F7 verified"
```

---

*v1.0.0 | 2026-04-06*  
*10 sesiuni Sonnet, ~7–8h total*  
*PRE-COMMIT CHECKLIST obligatoriu în fiecare sesiune*
