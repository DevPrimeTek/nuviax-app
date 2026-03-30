# docs/testing/flows/goal-flow.md — Goal & Daily Execution Flows

> Part of: `docs/testing/` | Scenarios: `scenarios/critical.md` (TS-01, TS-02, TS-09)

---

## 1. User Journey (End-to-End)

### 1.1 Registration & Onboarding

1. User submits `POST /auth/register` (name, email, password)
2. Backend: bcrypt hash, insert `users`, generate JWT RS256 pair (access 15min / refresh 7d)
3. Welcome email sent via Resend (`email.go` → `WelcomeEmail`)
4. Frontend redirects → `/onboarding`
5. Onboarding page: user enters first goal title
6. `POST /goals/suggest-category` called with 2s hard timeout → Claude Haiku returns category suggestion or falls back silently
7. User selects category pill (HEALTH / CAREER / FINANCE / RELATIONSHIPS / LEARNING / CREATIVITY / OTHER) and optional `dominant_behavior_model` (G-11: ANALYTIC / STRATEGIC / TACTICAL / REACTIVE)
8. `POST /goals` creates entry in `global_objectives`

### 1.2 First Login

1. User submits `POST /auth/login` (email, password)
2. Backend: bcrypt verify, generate new JWT pair, store refresh token in Redis
3. Frontend stores access token in memory; refresh token in `httpOnly` cookie
4. `GET /settings` returns `theme` (dark/light) and language preference
5. Frontend applies `data-theme` + `nv_lang` from `localStorage` — anti-flash inline script runs before hydration
6. Redirect → `/dashboard`

### 1.3 Goal Setup

1. `GET /goals` returns list of user's `global_objectives`
2. User creates goal → `POST /goals` (title, category, deadline, optional `dominant_behavior_model`)
3. Engine Layer 0 (C1–C8) initializes base score via `engine.go`
4. `POST /goals/:id/sprint` creates first sprint in `sprints` table
5. Level 1 engine (`level1_structural.go`) generates initial tasks via Claude Haiku task generation
6. Scheduler cron (`scheduler.go`) generates daily tasks at midnight

### 1.4 Daily Loop

1. User opens `/today` → `GET /today` returns energy level + main tasks + personal tasks
2. User sets energy level (1–5) for the day
3. Task list rendered from active sprint's generated tasks
4. User completes tasks → `POST /today/complete/:id`
5. Each completion triggers server-side score recalculation (Level 2 engine, C19–C25)
6. End-of-day: scheduler runs daily check-in job; dashboard cache invalidated, checkpoint statuses updated

⚠️ Regression event recording NOT IMPLEMENTED (see SA-3, Sprint 3.1)

### 1.5 Progress Review

1. `GET /goals/:id` returns goal with progress % and grade (A+/A/B/C/D) — opaque output only
2. `GET /goals/:id/visualization` returns chart data points (Level 5, `visualization.go`)
3. `/dashboard` aggregates all goals — grades and progress bars rendered client-side from server values
4. `/recap` provides sprint summary: completion rate, streak, energy average
5. `GET /profile/activity` returns 365-day activity data → rendered as 52-week heatmap (`ActivityHeatmap.tsx`)

---

## 2. Goal Creation Flow

### 2.1 Input (User Action)

- User submits: `name`, `start_date`, `end_date` (YYYY-MM-DD), optional `description`, optional `dominant_behavior_model`, optional `waiting_list: true`
- Source: `/onboarding` wizard or `/goals` create form

### 2.2 Backend Processing (`handlers.go` → `CreateGoal`)

1. **Date validation:** `end_date > start_date`; max duration 365 days — returns `400` on failure
2. **G-10 capacity check:** `engine.ValidateGoalActivation()` — max 3 active goals
   - If at capacity and `waiting_list: false` → goal auto-routed to `WAITING` (`vaulted: true` in response)
   - If `waiting_list: true` → status set to `WAITING` directly
3. **DB insert:** `db.CreateGoal()` → row in `global_objectives`
4. **G-11 behavior model:** if `dominant_behavior_model` provided → `db.SetGoalBehaviorModel()` updates `global_objectives.dominant_behavior_model`
5. **Sprint creation (ACTIVE goals only):**
   - Sprint 1 end = `start_date + 30 days` (capped at `end_date`)
   - `db.CreateSprint()` → row in `sprints`
   - 3 checkpoints created: `"Fundament: <name>"`, `"Progres: <name>"`, `"Consolidare: <name>"`
   - `engine.GenerateDailyTasks()` called immediately (no waiting for midnight scheduler)
6. **Cache:** `cache.InvalidateDashboard()` clears Redis dashboard cache for user

### 2.3 DB Changes

| Table | Operation |
|---|---|
| `global_objectives` | INSERT — new goal row with status ACTIVE / WAITING |
| `global_objectives` | UPDATE `dominant_behavior_model` (if G-11 provided) |
| `sprints` | INSERT Sprint 1 (ACTIVE goals only) |
| `checkpoints` | INSERT × 3 (ACTIVE goals only) |
| `daily_tasks` | INSERT tasks for today (ACTIVE goals only, via engine) |

### 2.4 API Response

**ACTIVE goal (standard):** `201` + goal object

**Auto-vaulted to WAITING (G-10):**
```json
{
  "goal": { ... },
  "message": "Ai deja 3 obiective active. Obiectivul a fost adăugat în Vault-ul viitor...",
  "status": "WAITING",
  "vaulted": true
}
```

**Validation error:** `400` / `422` + `{ "error": "..." }`

### 2.5 Frontend Behavior

1. On success (ACTIVE): redirect → `/today`; dashboard cache busted → fresh load
2. On `vaulted: true`: show vault notice banner; stay on `/goals`
3. On `422`: display inline error; no redirect
4. AI category suggestion (`POST /goals/suggest-category`) runs debounced on title input before form submit — result pre-fills category pill; no block if timeout (2s fallback)

---

## 3. Daily Execution Flow

### 3.1 Input (User Action)

- User opens `/today`
- User sets energy level (low / normal / high)
- User completes tasks
- User optionally adds personal tasks (max 2/day)

### 3.2 Backend Processing (`handlers.go` → `GetTodayTasks`, `CompleteTask`, `SetEnergy`)

**Loading today's tasks (`GET /today`):**
1. Redis cache checked first (`cache.GetTodayTasks`) — returns immediately on hit
2. On miss: `db.GetTodayTasks()` for today's date
3. If DB returns empty: `engine.GenerateDailyTasks()` called on-demand (fallback to scheduler)
4. Tasks split into `MAIN` (sprint-generated) and `PERSONAL` (user-added)
5. Streak days, current checkpoint (status `IN_PROGRESS`), day-in-sprint number all fetched
6. Response cached in Redis

**Setting energy (`POST /context/energy`):**
1. Frontend label normalized: `mid → normal`, `hi → high`
2. `normal` energy → no DB action, returns `200` immediately
3. `low` / `high` → `db.CreateContextAdjustment()` with `adjType = AdjEnergyLow / AdjEnergyHigh`, active from today to tomorrow
4. Today's Redis task cache invalidated → next load regenerates with adjusted intensity

**Completing a task (`POST /today/complete/:id`):**
1. `db.CompleteTask()` — sets `completed = TRUE`, records timestamp
2. Both today-tasks and dashboard Redis caches invalidated
3. No immediate score recalculation — score computed on-demand via `engine.ComputeGoalScore()`

**Adding a personal task (`POST /today/personal`):**
1. Max 2 personal tasks/day enforced via `db.CountPersonalTasksToday()`
2. Active sprint resolved from first ACTIVE goal
3. `db.CreateTask()` inserts with `task_type = PERSONAL`
4. Today-tasks cache invalidated

### 3.3 DB Changes

| Table | Operation |
|---|---|
| `daily_tasks` | UPDATE `completed = TRUE` on task completion |
| `context_adjustments` | INSERT energy adjustment (low/high only) |
| `daily_tasks` | INSERT personal task |

### 3.4 API Response

**`GET /today`:**
```json
{
  "date": "2026-03-30T00:00:00Z",
  "goal_name": "...",
  "day_number": 14,
  "main_tasks": [ { "id": "...", "text": "...", "completed": false } ],
  "personal_tasks": [ ... ],
  "done_count": 2,
  "total_count": 5,
  "streak_days": 7,
  "checkpoint": { "id": "...", "name": "Progres: ...", "status": "IN_PROGRESS" }
}
```

**`POST /today/complete/:id`:** `200` + `{ "message": "Activitate bifată." }`

**`POST /context/energy` (low/high):** `200` + `{ "message": "...Activitățile de mâine vor fi adaptate." }`

### 3.5 Frontend Behavior

1. `/today` renders main tasks + personal tasks in separate lists
2. Checkbox tap → optimistic UI update → `POST /today/complete/:id`
3. Energy selector: 3 options (low / normal / high); selection calls `POST /context/energy`; no page reload
4. "Add personal task" button disabled after 2 tasks/day (client enforced + server enforced)
5. Streak counter and checkpoint banner update on each page load (no real-time push)
