# PROMPTS_MVP.md — Claude Code Session Prompts

> **Versiune:** 1.0.1  
> **Actualizat:** 2026-04-18 (PM review — prompts recalibrate la starea reală)  
> **Reguli:**  
> - Fiecare prompt = o sesiune Claude Code nouă  
> - Copiază blocul ``` integral → paste în sesiune  
> - Ordine strictă: F5a → F5b → F6-audit → F7  
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

## Index sesiuni rămase

| # | Ce face | Model | Timp | Branch |
|---|---------|-------|------|--------|
| F5a | API: Goals + Today + Dashboard handlers + server.go wiring | Sonnet | 45 min | `claude/api-handlers` |
| F5b | API: SRM + Achievements + Profile + Admin + AI client integrat | Sonnet | 45 min | `claude/api-handlers` |
| F6-audit | Frontend: build check + fix-uri după F5 | Sonnet | 30 min | `claude/frontend-mvp` |
| F7 | Smoke test + docs finale v1.1.0 | Sonnet | 30 min | `claude/smoke-test` |

**Model:** Sonnet pentru tot. Opus doar la decizie arhitecturală neprevăzută.  
**Total: 4 sesiuni, ~2.5–3h**

---

## Context actual (citește înainte de F5a)

**Ce există deja:**
- `backend/internal/api/handlers/handlers.go` — Auth complet (Register, Login, MFA, RefreshToken, Logout, ForgotPassword, ResetPassword)
- `backend/internal/api/server.go` — Config struct + routing (doar auth înregistrat)
- `backend/internal/engine/` — engine complet (F3 ✅)
- `backend/internal/scheduler/scheduler.go` — 12 jobs (F4 ✅)
- `backend/internal/ai/ai.go` — Client complet (GenerateTaskTexts, AnalyzeGO, SuggestGOCategory)
- `backend/internal/email/email.go` — Client complet (SendWelcome, SendSprintComplete, SendPasswordReset)
- Frontend: toate paginile și componentele există

**Ce lipsește și trebuie adăugat în F5:**
- `handlers.Handlers` struct: lipsește câmpul `ai *ai.Client`
- `handlers.New(...)`: lipsește parametrul `aiClient *ai.Client`
- `api.Config` struct: lipsește câmpul `AIClient *ai.Client`
- `server.go`: nu inițializează AI client, nu înregistrează rute business
- Toți handlerii de business (Goals, Today, Dashboard, SRM, Profile, Achievements, Admin)

---

## F5a — API Core (Goals + Today + Dashboard)

**Model: Sonnet** | **Max 45 min** | **Branch: `claude/api-handlers`**

```
Read CLAUDE.md v1.0.1. Branch: claude/api-handlers. Task: adaugă handlers Goals+Today+Dashboard în handlers.go + înregistrează rute în server.go.

## IMPORTANT: Citește CLAUDE.md secțiunea 8 (Integrări existente).
## ai.go și email.go DEJA EXISTĂ — NU le rescrie. Integrează-le.

## Citește max 3 fișiere
1. backend/internal/api/handlers/handlers.go — ce există (Auth complet)
2. backend/internal/ai/ai.go — funcțiile AI existente
3. backend/internal/engine/engine.go — funcții F3

## PASUL 1 — Extinde Handlers struct și New() în handlers.go

Adaugă câmpul ai la struct:
  type Handlers struct {
    db     *pgxpool.Pool
    redis  *redis.Client
    auth   *auth.Service
    engine *engine.Engine
    encKey []byte
    email  *email.Client
    ai     *ai.Client  // ← ADAUGĂ (nil dacă ANTHROPIC_API_KEY lipsește)
  }

Modifică New() să accepte aiClient:
  func New(pool *pgxpool.Pool, rdb *redis.Client, authSvc *auth.Service, eng *engine.Engine, encKey []byte, emailClient *email.Client, aiClient *ai.Client) *Handlers {
    return &Handlers{db: pool, redis: rdb, auth: authSvc, engine: eng, encKey: encKey, email: emailClient, ai: aiClient}
  }

## PASUL 2 — Adaugă handlerii în handlers.go

### POST /goals/analyze — AI Validare GO (C9/C10)
  Parse {text string}
  Dacă h.ai != nil:
    ctx2, cancel := context.WithTimeout(c.Context(), 2*time.Second)
    defer cancel()
    result, err := h.ai.AnalyzeGO(ctx2, req.Text)
    Dacă err != nil → fallback
  Fallback rule-based: vagueTerms check + measurable check
  Return {needs_clarification bool, question string, hint string, source string}
  NICIODATĂ nu returna drift/chaos/weights/thresholds

### POST /goals/suggest-category
  Dacă h.ai != nil: ai.SuggestGOCategory(ctx 2s timeout)
  Return {category string, confidence float64}
  Fallback: {category: "", confidence: 0}

### POST /goals — CreateGoal (C3, C4, C12, C14)
  Parse {name, start_date, end_date, dominant_behavior_model, description?}
  engine.ValidateGO(name, bm, start, end, activeCount) → 400 dacă eroare
  Dacă activeCount >= 3 → INSERT GO cu status=WAITING (C12 Future Vault) → return 201
  Altfel → INSERT GO ACTIVE + INSERT sprint 30 zile + INSERT 3 checkpoints
  Return 201: {id, name, status, sprint_id} — FĂRĂ drift/chaos/weights

### GET /goals — ListGoals
  SELECT GO pentru user_id curent
  Return [{id, name, status, progress_pct, grade, behavior_model, end_date}]
  progress_pct și grade din ultima înregistrare go_scores

### GET /goals/:id — GetGoalDetail
  Return {id, name, status, progress_pct, grade, sprint_day, sprint_total, checkpoint_name}
  FĂRĂ drift, chaos_index, weights, thresholds

### GET /goals/:id/visualize
  SELECT growth_trajectories WHERE go_id = :id ORDER BY recorded_at
  Return [{date, progress_pct, expected_pct}]

### GET /today — GetToday
  SELECT daily_tasks WHERE user_id=? AND task_date=TODAY AND (status!='CANCELLED')
  JOIN cu sprint info pentru GO activ
  Return {goal_name, day_number, checkpoint{name, progress_pct}, streak_days, main_tasks[], personal_tasks[]}

### POST /today/complete/:id — CompleteTask
  UPDATE daily_tasks SET completed=TRUE, completed_at=NOW() WHERE id=:id AND user_id=?
  Return 200 {ok: true}

### POST /today/personal — AddPersonalTask (max 2/zi)
  Verifică COUNT(personal tasks today) < 2 → 400 dacă depășit
  INSERT daily_tasks cu type=PERSONAL
  Return 201 task creat

### POST /context/energy — SetEnergy
  Parse {level: "low"|"mid"|"hi"}
  INSERT context_adjustments cu type=ENERGY_LEVEL
  Return 200 {ok: true}

### GET /dashboard — GetDashboard
  SELECT active GOs cu ultima go_scores
  Return {goals: [{id, name, progress_pct, grade, streak_days}], active_count, tasks_today_done, tasks_today_total, srm_active bool}
  Cache în Redis 5 min (key: "dashboard:{userId}")

## PASUL 3 — Actualizează server.go

Adaugă AIClient în Config:
  type Config struct {
    ...
    AIClient    *ai.Client    // ← ADAUGĂ (nil dacă ANTHROPIC_API_KEY lipsește)
  }

Actualizează handlers.New() call:
  h := handlers.New(cfg.DB, cfg.Redis, authSvc, eng, encKey, cfg.EmailClient, cfg.AIClient)

Adaugă rute în grupul protected p:
  // Goals
  p.Post("/goals/analyze", h.AnalyzeGO)
  p.Post("/goals/suggest-category", h.SuggestGOCategory)
  p.Post("/goals", h.CreateGoal)
  p.Get("/goals", h.ListGoals)
  p.Get("/goals/:id", h.GetGoalDetail)
  p.Get("/goals/:id/visualize", h.GetGoalVisualize)
  // Today
  p.Get("/today", h.GetToday)
  p.Post("/today/complete/:id", h.CompleteTask)
  p.Post("/today/personal", h.AddPersonalTask)
  p.Post("/context/energy", h.SetEnergy)
  // Dashboard
  p.Get("/dashboard", h.GetDashboard)

## PASUL 4 — Actualizează cmd/main.go sau cmd/api/main.go
Inițializează AI client la pornire:
  aiClient, _ := ai.New()   // nil dacă ANTHROPIC_API_KEY lipsește — graceful degradation
  cfg := api.Config{
    ...
    AIClient:   aiClient,
  }

## Verificare
  cd backend && go build ./... && go vet ./...
  grep -rn "drift\|chaos_index\|weights\|threshold" backend/internal/api/handlers/
  # ZERO rezultate

## PRE-COMMIT CHECKLIST (OBLIGATORIU)
  grep -rn "sk-ant-\|re_[A-Za-z0-9]\{20,\}" README.md CLAUDE.md ROADMAP.md 2>/dev/null  # ZERO
  grep -rn "drift\|chaos_index\|weights\|threshold" backend/internal/api/handlers/  # ZERO
  Actualizează README.md: endpoints Goals/Today/Dashboard marcate ✅
  Verifică versiuni: grep "Versiune\|1.0.1" CLAUDE.md README.md ROADMAP.md

## Commit
  git add backend/internal/api/
  git commit -m "feat(api): Goals, Today, Dashboard handlers + server.go wiring (F5a)"
```

---

## F5b — API Extended (SRM + Achievements + Profile + Admin)

**Model: Sonnet** | **Max 45 min** | **Branch: `claude/api-handlers` (continuare)**

```
Read CLAUDE.md v1.0.1. Branch: claude/api-handlers (continuare F5a). Task: SRM + achievements + profile + admin handlers.

## IMPORTANT: ai.go, email.go DEJA EXISTĂ. handlers.go din F5a are AI integrat.
## NU rescrie ce există. Adaugă doar ce lipsește.

## Citește max 3 fișiere
1. backend/internal/api/handlers/handlers.go — din F5a (verifică ce există)
2. backend/internal/engine/srm.go — ChaosLevel, ComputeSRMFallback
3. backend/internal/api/server.go — rutele din F5a

## Adaugă handlerii în handlers.go:

### GET /srm/status/:goalId
  SELECT srm_events WHERE go_id=:goalId ORDER BY created_at DESC LIMIT 1
  Return {srm_level: "L1"|"L2"|"L3"|"NONE", triggered_at}
  NICIODATĂ nu returna chaos_index, drift, weights

### POST /srm/confirm-l2/:goalId
  UPDATE srm_events SET confirmed_at=NOW() WHERE go_id=:goalId AND level='L2' AND confirmed_at IS NULL
  INSERT context_adjustments cu type=ENERGY_LOW
  Return 200 {ok: true}

### POST /srm/confirm-l3/:goalId
  UPDATE global_objectives SET status='PAUSED', paused_at=NOW() WHERE id=:goalId AND user_id=?
  INSERT context_adjustments cu type=PAUSE
  Return 200 {ok: true}

### GET /achievements
  SELECT achievements WHERE user_id=? ORDER BY earned_at DESC
  Return [{id, type, title, description, earned_at}]

### GET /ceremonies/:goalId
  SELECT ceremonies WHERE go_id=:goalId ORDER BY created_at DESC LIMIT 1
  Return {id, tier, sprint_score, viewed_at} sau null

### POST /ceremonies/:id/view
  UPDATE ceremonies SET viewed_at=NOW() WHERE id=:id AND go_id IN (GO-uri ale userului)
  Return 200 {ok: true}

### GET /profile/activity
  SELECT daily_metrics WHERE user_id=? AND recorded_at >= NOW()-365 days
  Return [{date, tasks_completed, active_minutes}] (365 entries)

### PATCH /settings
  Parse {theme?: "light"|"dark", locale?: "ro"|"en"|"ru"}
  UPDATE users SET preferences = preferences || jsonb_build_object(...)
  Return 200 {ok: true}

### GET /admin/stats — Admin guard (404 non-admin)
  middleware.RequireAdmin → dacă user.is_admin=false → 404 (nu 403!)
  SELECT COUNT(users), COUNT(active GOs), COUNT(tasks today)
  Return {users_total, active_goals, tasks_today}

### GET /admin/users
  middleware.RequireAdmin → 404 dacă non-admin
  SELECT users cu email decriptat, created_at, last_login
  Return [{id, email, full_name, created_at, is_active, goals_count}]

### POST /admin/users/:id/deactivate
  middleware.RequireAdmin → 404 dacă non-admin
  UPDATE users SET is_active=FALSE WHERE id=:id
  Return 200 {ok: true}

## Adaugă rute în server.go (grupul p):
  // SRM
  p.Get("/srm/status/:goalId", h.GetSRMStatus)
  p.Post("/srm/confirm-l2/:goalId", h.ConfirmSRML2)
  p.Post("/srm/confirm-l3/:goalId", h.ConfirmSRML3)
  // Achievements + Ceremonies
  p.Get("/achievements", h.ListAchievements)
  p.Get("/ceremonies/:goalId", h.GetCeremony)
  p.Post("/ceremonies/:id/view", h.ViewCeremony)
  // Profile + Settings
  p.Get("/profile/activity", h.GetProfileActivity)
  p.Patch("/settings", h.UpdateSettings)
  // Admin (cu RequireAdmin middleware)
  admin := p.Group("/admin", middleware.RequireAdmin)
  admin.Get("/stats", h.AdminStats)
  admin.Get("/users", h.AdminUsers)
  admin.Post("/users/:id/deactivate", h.AdminDeactivateUser)

## Verificare
  cd backend && go build ./... && go vet ./...
  grep -rn "drift\|chaos_index\|weights\|threshold" backend/internal/api/handlers/  # ZERO

## PRE-COMMIT CHECKLIST (OBLIGATORIU)
  grep -rn "sk-ant-\|re_[A-Za-z0-9]\{20,\}" README.md CLAUDE.md ROADMAP.md 2>/dev/null  # ZERO
  grep -rn "drift\|chaos_index\|weights\|threshold" backend/internal/api/handlers/  # ZERO
  Actualizează CLAUDE.md: F5 → ✅
  Actualizează ROADMAP.md: F5 → ✅ cu data
  Actualizează README.md: toate endpoint-urile marcate ✅
  Verifică versiuni: grep "Versiune\|1.0.1" CLAUDE.md README.md ROADMAP.md  # toate identice

## Commit
  git add backend/internal/api/
  git commit -m "feat(api): SRM, achievements, profile, admin handlers (F5b)"
```

---

## F6-audit — Frontend Audit + Fix

**Model: Sonnet** | **Max 30–45 min** | **Branch: `claude/frontend-mvp`**

```
Read CLAUDE.md v1.0.1. Branch: claude/frontend-mvp. Task: audit frontend — build check + fix erori după F5.

## Context
Toate paginile și componentele frontend EXISTĂ deja:
  Pagini: onboarding, today, goals, goals/[id], dashboard, recap, profile, settings, achievements, admin
  Componente: AppShell, SRMWarning, CeremonyModal, ActivityHeatmap, ProgressCharts, GoalTabs

## Citește max 3 fișiere (cele cu erori de build dacă există)
1. frontend/app/app/onboarding/page.tsx — verifică fluxul AI
2. frontend/app/app/dashboard/page.tsx — verifică ce afișează
3. frontend/app/lib/api.ts — tipurile definite

## PASUL 1 — Build check
  cd frontend/app && npm run build 2>&1 | head -50
  Notează TOATE erorile TypeScript/build

## PASUL 2 — Audit endpoint calls
Verifică că fiecare pagină cheamă endpoint-uri care EXISTĂ după F5:
  Onboarding: POST /goals/analyze ✓, POST /goals/suggest-category ✓, POST /goals ✓
  Today: GET /today ✓, POST /today/complete/:id ✓, POST /today/personal ✓, POST /context/energy ✓
  Goals: GET /goals ✓, GET /goals/:id ✓
  Dashboard: GET /dashboard ✓
  SRM: GET /srm/status/:id ✓, POST /srm/confirm-l2/:id ✓, POST /srm/confirm-l3/:id ✓
  Achievements: GET /achievements ✓
  Profile: GET /profile/activity ✓
  Settings: PATCH /settings ✓
  Admin: GET /admin/stats ✓, GET /admin/users ✓

## PASUL 3 — Fix erori de build
  Repară DOAR erorile de TypeScript/linting care blochează build-ul
  NU adăuga funcționalitate nouă — scopul e build success

## PASUL 4 — Verificare build final
  cd frontend/app && npm run build
  # ZERO erori de build

## PRE-COMMIT CHECKLIST (OBLIGATORIU)
  grep -rn "sk-ant-\|API_KEY.*=.*[A-Za-z0-9]" frontend/app/  # ZERO
  Actualizează CLAUDE.md: F6 → ✅
  Actualizează ROADMAP.md: F6 → ✅ cu data
  Verifică versiuni: grep "Versiune" CLAUDE.md README.md ROADMAP.md

## Commit
  git add frontend/app/
  git commit -m "fix(frontend): build errors fixed, F6 complete"
```

---

## F7 — Smoke Test + Docs Finale

**Model: Sonnet** | **Max 30 min** | **Branch: `claude/smoke-test`**

```
Read CLAUDE.md v1.0.1. Branch: claude/smoke-test. Task: E2E smoke test + docs finale v1.1.0.

## PASUL 1 — Smoke tests curl

  # 1. Register
  curl -s -o /dev/null -w "%{http_code}" -X POST localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d '{"email":"smoke@test.com","password":"Smoke1234!","full_name":"Smoke Test"}'
  # → 201 sau 409

  # 2. Login + token
  TOKEN=$(curl -s -X POST localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"smoke@test.com","password":"Smoke1234!"}' | jq -r '.access_token')

  # 3. AI validate GO
  curl -s -H "Authorization: Bearer $TOKEN" -X POST localhost:8080/api/v1/goals/analyze \
    -H "Content-Type: application/json" \
    -d '{"text":"Vreau să slăbesc"}'
  # → {needs_clarification: true, ...} sau {needs_clarification: false}

  # 4. Create GO
  curl -s -H "Authorization: Bearer $TOKEN" -X POST localhost:8080/api/v1/goals \
    -H "Content-Type: application/json" \
    -d '{"name":"Test MVP Goal","start_date":"2026-04-18","end_date":"2026-10-18","dominant_behavior_model":"INCREASE"}'
  # → 201

  # 5. Today
  curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/today
  # → 200

  # 6. Dashboard
  curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/dashboard
  # → 200 cu goals

  # 7. API Opacity check CRITIC
  curl -s -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/goals | grep -i "drift\|chaos\|weight\|threshold"
  # → ZERO rezultate

  # 8. Admin guard
  curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" localhost:8080/api/v1/admin/stats
  # → 404

## PASUL 2 — Dacă buguri → fix sau docs/known-issues.md

## PASUL 3 — PRE-COMMIT CHECKLIST FINAL (OBLIGATORIU)

  # Securitate completă
  grep -rn "sk-ant-\|re_[A-Za-z0-9]\{20,\}\|password.*=.*[A-Za-z]" README.md CLAUDE.md ROADMAP.md infra/.env.example 2>/dev/null
  # ZERO

  # API opacity
  grep -rn "drift\|chaos_index\|weights\|threshold" backend/internal/api/handlers/
  # ZERO

  # Actualizare docs finale v1.1.0
  Actualizează README.md: versiune → 1.1.0
  Actualizează CLAUDE.md: F5–F7 → ✅, versiune → 1.1.0
  Actualizează ROADMAP.md: F0–F7 → ✅, versiune → 1.1.0

  # Verificare versiuni
  grep "1.1.0" CLAUDE.md README.md ROADMAP.md
  # Toate trei: 1.1.0

## Commit
  git add README.md CLAUDE.md ROADMAP.md
  git commit -m "docs: v1.1.0 — MVP complete, F0–F7 verified"
```

---

## Sesiunile F0.1–F4 (COMPLETATE ✅)

Sesiunile F0.1, F3a, F3b, F4 au fost executate cu succes (commits pe branch-urile respective, mergere în main).  
Nu mai sunt necesare — documentate doar ca referință istorică.

---

*v1.0.1 | 2026-04-18*  
*4 sesiuni rămase: F5a + F5b + F6-audit + F7 (~2.5–3h total)*  
*PRE-COMMIT CHECKLIST obligatoriu în fiecare sesiune*
