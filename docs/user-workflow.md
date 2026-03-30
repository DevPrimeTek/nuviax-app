# docs/user-workflow.md — NuviaX User Workflow

> Version: 10.5.0 | Last updated: 2026-03-30

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
4. User completes tasks → `POST /tasks/:id/complete`
5. Each completion triggers server-side score recalculation (Level 2 engine, C19–C25)
6. End-of-day: scheduler runs daily check-in job; missed tasks recorded as regression events (`level2_execution.go`)

### 1.5 Progress Review

1. `GET /goals/:id` returns goal with progress % and grade (A+/A/B/C/D) — opaque output only
2. `GET /goals/:id/visualization` returns chart data points (Level 5, `visualization.go`)
3. `/dashboard` aggregates all goals — grades and progress bars rendered client-side from server values
4. `/recap` provides sprint summary: completion rate, streak, energy average
5. `GET /profile/activity` returns 365-day activity data → rendered as 52-week heatmap (`ActivityHeatmap.tsx`)

### 1.6 SRM Triggers

1. Level 4 engine (`level4_regulatory.go`) evaluates SRM conditions after each score update
2. `SRMWarning.tsx` banner appears on dashboard when SRM is active
3. SRM levels: L1 (daily review) → L2 (weekly review) → L3 (monthly review), escalating
4. User completes SRM → `POST /srm` — score recalculated with regulatory adjustments
5. Successful SRM exits warning state; failed SRM may escalate level

### 1.7 Achievement & Ceremony Flow

1. Level 5 engine (`level5_growth.go`) evaluates achievement conditions post-score-update
2. Sprint close (scheduler cron) → `ApplyEvolveOverride()` runs for hybrid GO behavior models
3. Ceremony assigned: BRONZE / SILVER / GOLD / PLATINUM based on sprint performance
4. `GET /ceremonies/latest` returns unviewed ceremony → `CeremonyModal.tsx` displayed on next login
5. `POST /ceremonies/:id/viewed` marks ceremony as seen
6. Achievements stored in `achievements` table; `GET /achievements` returns full badge grid

### 1.8 Growth & Visualization

1. Trajectory data accumulated across sprints in `level5_growth.go`
2. `GET /goals/:id/visualization` → `ProgressCharts.tsx` renders LineChart + BarChart (Recharts)
3. Profile page: avatar, stats, activity heatmap, preferences (theme, language)
4. `PATCH /settings` persists theme + language to DB (`users.theme` — migration 012)
5. All score components (drift, chaos_index, weights, thresholds) remain server-only — never exposed in API responses

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

**Completing a task (`POST /tasks/:id/complete`):**
1. `db.CompleteTask()` — sets `completed = TRUE`, records timestamp
2. Both today-tasks and dashboard Redis caches invalidated
3. No immediate score recalculation — score computed on-demand via `engine.ComputeGoalScore()`

**Adding a personal task (`POST /tasks/personal`):**
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

**`POST /tasks/:id/complete`:** `200` + `{ "message": "Activitate bifată." }`

**`POST /context/energy` (low/high):** `200` + `{ "message": "...Activitățile de mâine vor fi adaptate." }`

### 3.5 Frontend Behavior

1. `/today` renders main tasks + personal tasks in separate lists
2. Checkbox tap → optimistic UI update → `POST /tasks/:id/complete`
3. Energy selector: 3 options (low / normal / high); selection calls `POST /context/energy`; no page reload
4. "Add personal task" button disabled after 2 tasks/day (client enforced + server enforced)
5. Streak counter and checkpoint banner update on each page load (no real-time push)

---

## 4. SRM Flow (L1–L3)

### 4.1 Input (User Action)

- SRM is **engine-triggered**, not user-initiated
- User confirms L2 or L3 when banner is shown
- Entry points: dashboard SRM warning banner (`SRMWarning.tsx`), goal detail page
- Endpoints: `GET /srm/status/:goalId`, `POST /srm/confirm-l2/:goalId`, `POST /srm/confirm-l3/:goalId`

### 4.2 Backend Processing (`srm.go` + `level4_regulatory.go`)

**SRM Status (`GET /srm/status/:goalId`):**
1. Queries `srm_events` for most recent non-revoked event
2. Returns `srm_level: NONE / L1 / L2 / L3`
3. Computes ALI breakdown (current + projected): `computeALIBreakdown()`
   - `ALI_current` = tasks completed / expected by now
   - `ALI_projected` = if current pace continues to sprint end
   - Ambition buffer zone: `ALI_projected` between 1.0–1.15 → Velocity Control warning
   - `velocity_control_on: true` if `ALI_projected > 1.15`

**L1 — Automatic adjustment (no user action required):**
- Triggered by `level4_regulatory.go` during scheduler run
- Task intensity reduced automatically
- No user confirmation needed
- `srm_events` row inserted with `srm_level = 'L1'`

**L2 — Structural recalibration (`POST /srm/confirm-l2/:goalId`):**
1. Verifies access: `db.GetGoalByID()` — returns `404` if not owner
2. `UPDATE srm_events SET confirmed_at = NOW(), confirmed_by = $2` on most recent unconfirmed L2
3. If no active unconfirmed L2 event → `404`
4. Task intensity adjusted; sprint structure recalibrated by engine
5. Goal status remains `ACTIVE`

**L3 — Strategic reset (`POST /srm/confirm-l3/:goalId`):**
1. Verifies goal ownership
2. `UPDATE global_objectives SET status = 'PAUSED'`
3. `INSERT INTO srm_events` with `srm_level = 'L3'`, trigger_reason = `'user_confirmed_stabilization'`
4. `engine.FreezeExpectedTrajectory(sprint.ID)` — freezes projected trajectory to prevent drift loop paradox (GAP #20)
5. `frozen_expected` percentage computed from elapsed time / total goal duration

### 4.3 DB Changes

| Table | Operation | When |
|---|---|---|
| `srm_events` | INSERT new event | L1/L2/L3 trigger |
| `srm_events` | UPDATE `confirmed_at` | L2 user confirmation |
| `srm_events` | INSERT with `trigger_reason` | L3 user confirmation |
| `global_objectives` | UPDATE `status = 'PAUSED'` | L3 only |
| `sprint_trajectories` | UPDATE frozen flag | L3 only (GAP #20) |

### 4.4 API Response

**`GET /srm/status/:goalId`:**
```json
{
  "goal_id": "...",
  "srm_level": "L2",
  "message": "Ajustare structurală în curs. Am recalibrat obiectivele.",
  "ali": {
    "ali_current": 0.72,
    "ali_projected": 0.68,
    "in_ambition_buffer": false,
    "velocity_control_on": false,
    "goal_breakdown": [ ... ],
    "note": "ali_current = progres realizat până acum. ali_projected = proiecție la finalul sprintului."
  }
}
```

**`POST /srm/confirm-l2`:** `200` + message + `next_step`

**`POST /srm/confirm-l3`:** `200` + `new_status: PAUSED` + `frozen_expected` percentage + `next_step`

### 4.5 Frontend Behavior

1. `SRMWarning.tsx` banner displayed on dashboard when `srm_level ≠ NONE`
2. L1: informational banner only — no action button
3. L2: banner shows "Confirm recalibration" button → calls `POST /srm/confirm-l2`; on success banner dismissed
4. L3: banner shows "Activate stabilization mode" button → calls `POST /srm/confirm-l3`; on success goal card shows `PAUSED` badge; reactivation proposed after 7 days (scheduler)

---

## 5. Achievement Flow

### 5.1 Achievement Trigger Conditions
### 5.2 Ceremony Tiers (Tier 1–3)
### 5.3 Badge Award & Display
### 5.4 Achievement History (/profile)

---

## 6. Visualization Flow

### 6.1 Progress Bar & Grade Display
### 6.2 Activity Heatmap (52-week)
### 6.3 Goal Progress Cards
### 6.4 Profile Stats Overview
### 6.5 Dark/Light Theme Rendering

---

## 7. Test Scenarios

---

### TS-01 — Happy Path: New User Full Journey

**Steps:**
1. `POST /auth/register` with valid name, email, password
2. `POST /auth/login` → receive access token + refresh token cookie
3. `GET /settings` → assert `theme` field present
4. Navigate to `/onboarding` → type goal title → wait for AI suggestion (`POST /goals/suggest-category`)
5. Select category → `POST /goals` with `start_date`, `end_date` (30-day range)
6. `GET /today` → assert `main_tasks` array is non-empty, `day_number = 1`
7. `POST /tasks/:id/complete` on first task → assert `200`
8. `GET /today` again → assert `done_count = 1`
9. `GET /goals/:id` → assert `progress_pct > 0`, `grade` is non-empty string

**Expected result:** User registered, goal created, Sprint 1 active, first task completable on same day as registration.

**Failure indicator:**
- `main_tasks` empty on first `/today` load → `engine.GenerateDailyTasks()` not called at goal creation
- `progress_pct = 0` after task completion → score not computing
- Missing `grade` in goal response → engine not returning opaque output

---

### TS-02 — Goal Capacity Limit (G-10 Vault)

**Steps:**
1. Create 3 active goals (all with `waiting_list: false`)
2. Attempt `POST /goals` for a 4th goal with `waiting_list: false`
3. Check response body

**Expected result:** `201` with `"vaulted": true`, `"status": "WAITING"`. Goal created but not active.

**Failure indicator:**
- `status = ACTIVE` on 4th goal → G-10 capacity check bypassed
- `422` error → vaulted goal not created at all

---

### TS-03 — Trajectory Has >1 Data Points

**Steps:**
1. Create an active goal (Day 1)
2. Simulate scheduler run: trigger `jobComputeDailyScore` (or wait 24h in prod)
3. Complete at least one task on Day 1
4. Simulate Day 2: trigger daily score job again
5. `GET /goals/:id/visualization`

**Expected result:** `trajectory` array contains ≥2 entries, each with `date`, `actual_pct`, `expected_pct`, `delta`, `trend`.

**Fallback behavior to verify:** If only 1 day has passed (trajectory = 0 DB rows), `GenerateProgressVisualization` returns a single live-computed snapshot — `actual_pct: 0`, `expected_pct > 0`, `trend: "ON_TRACK"`. Not a failure; verify array is never null/empty.

**Failure indicator:**
- `trajectory: null` or `trajectory: []` → live snapshot fallback also failing
- Missing `expected_pct` field → engine returning internal weight data (critical rules violation)

---

### TS-04 — SRM L1 Triggers Automatically

**Steps:**
1. Create active goal
2. Miss all main tasks for 5 consecutive days (do not call `POST /tasks/:id/complete` on any `MAIN` task)
3. Simulate `jobDetectStagnation` run (scheduler job at 23:58 UTC)
4. `GET /srm/status/:goalId`

**Expected result:** `srm_level = "L1"`, `message = "Ajustare automată activă. Ritmul a fost redus ușor."`. `stagnation_events` has row with `inactive_days >= 5`.

**Failure indicator:**
- `srm_level = "NONE"` after 5 inactive days → `jobDetectStagnation` not running or threshold check failing
- `srm_level = "L2"` immediately → L2 chaos index threshold reached before L1; verify `chaos_index < 0.40`

---

### TS-05 — SRM L2 Reduces Task Intensity

**Steps:**
1. Create active goal; complete 0 tasks for several days until chaos index reaches threshold (≥ 0.40)
2. Simulate `jobRecalibrateRelevance` scheduler run
3. Verify `srm_events` has row with `srm_level = 'L2'`, `trigger_reason = 'chaos_index_threshold'`
4. `GET /srm/status/:goalId` → assert `srm_level = "L2"`
5. `POST /srm/confirm-l2/:goalId`
6. `GET /today` next day → compare task count vs pre-L2 baseline

**Expected result:** L2 confirmed; `confirmed_at` stamped on `srm_events`. Next day's task count is reduced (engine applies lower intensity). Goal status remains `ACTIVE`.

**Failure indicator:**
- `404` on `POST /srm/confirm-l2` → no active unconfirmed L2 event found
- Task count unchanged after L2 confirmation → engine not applying intensity reduction
- Goal status becomes `PAUSED` → L2 incorrectly escalating to L3 behavior

---

### TS-06 — SRM L3 Pauses Goal and Freezes Trajectory

**Steps:**
1. Trigger active L3 SRM condition (via manual `INSERT INTO srm_events` with `srm_level = 'L3'` for testing, or via L2 escalation)
2. `GET /srm/status/:goalId` → assert `srm_level = "L3"`
3. `POST /srm/confirm-l3/:goalId`
4. `GET /goals/:id` → check `status`
5. `GET /goals/:id/visualization` → check `frozen_expected` is a fixed value

**Expected result:** `201` → `new_status: "PAUSED"`, `frozen_expected` is a float between 0–1. `global_objectives.status = 'PAUSED'` in DB. `sprint_trajectories` frozen (drift loop prevented, GAP #20).

**Failure indicator:**
- Goal status still `ACTIVE` → L3 confirm not updating `global_objectives`
- `frozen_expected = 0` when goal is partway through → trajectory freeze not computing elapsed time correctly

---

### TS-07 — Achievement Unlocks and Ceremony Appears

**Steps:**
1. Create goal; complete sprint (30 days of tasks OR manually close sprint via `POST /goals/:id/sprint/close`)
2. Simulate scheduler run: `jobDetectEvolutionSprints` (01:00 UTC) then `jobGenerateCeremonies` (01:05 UTC)
3. `GET /ceremonies/latest` → assert ceremony present with `tier` field (BRONZE/SILVER/GOLD/PLATINUM)
4. `GET /achievements` → assert non-empty badge array
5. `POST /ceremonies/:id/viewed` → assert `200`
6. `GET /ceremonies/latest` → assert no unviewed ceremony returned

**Expected result:** Sprint closure generates ceremony. `CeremonyModal.tsx` shows on next login. Achievements recorded. `viewed` flag correctly prevents re-display.

**Failure indicator:**
- `GET /ceremonies/latest` returns `404` after sprint close → `jobGenerateCeremonies` not running or sprint close not setting `status = 'COMPLETED'`
- `tier` missing from ceremony → `engine.GenerateCompletionCeremony()` failing silently

---

### TS-08 — Visualization Is Not Empty on Day 1

**Steps:**
1. Create active goal
2. Immediately call `GET /goals/:id/visualization` (before any scheduler run)

**Expected result:** `trajectory` array has exactly 1 entry (live snapshot). `actual_pct: 0`, `expected_pct > 0` (time-based fraction), `trend: "ON_TRACK"`.

**Failure indicator:**
- `trajectory: []` or `trajectory: null` → live snapshot fallback in `GenerateProgressVisualization` not executing
- `expected_pct = 0` on a goal started 1+ days ago → start/end date computation broken

---

### TS-09 — Personal Task Limit Enforced

**Steps:**
1. `POST /tasks/personal` → task 1 created → assert `201`
2. `POST /tasks/personal` → task 2 created → assert `201`
3. `POST /tasks/personal` → task 3 attempt → assert `422` with error message

**Expected result:** Third personal task rejected with `"Poți adăuga maxim 2 activități personale pe zi."`.

**Failure indicator:**
- Third task accepted → `CountPersonalTasksToday` not checking correctly
- `422` on first task → goal not active, no sprint found

---

### TS-10 — Theme and Language Persist Across Sessions

**Steps:**
1. `PATCH /settings` with `{ "theme": "light" }`
2. Logout (`POST /auth/logout`)
3. Login again (`POST /auth/login`)
4. `GET /settings` → assert `theme = "light"`
5. On frontend: reload `/today` → assert `data-theme="light"` on `<html>` before hydration (anti-flash script)

**Expected result:** Theme persisted in `users.theme` (migration 012). Anti-flash inline script reads `nv_theme` from `localStorage` and applies `data-theme` before React hydration.

**Failure indicator:**
- `GET /settings` returns `theme = "dark"` after setting light → `UpdateSettings` not persisting `theme` field
- Flash of dark theme on page load → anti-flash script not running or `localStorage` not synced with DB value on login

---

### TS-11 — AI Category Suggestion Timeout Fallback

**Steps:**
1. Simulate Anthropic API unavailable (block outbound or set invalid key)
2. `POST /goals/suggest-category` with `{ "goal_name": "Run a marathon" }`
3. Assert response arrives within ~2 seconds

**Expected result:** `200` with empty suggestion (`""` or `null`) — no error, no hang. Onboarding continues normally. User must select category manually.

**Failure indicator:**
- Response takes >3 seconds → 2s hard timeout in `SuggestGOCategory()` not enforced
- `500` error returned → graceful degradation not working; AI failure should never block user flow

---

### TS-12 — Opaque API — No Internal Data Exposed

**Steps:**
1. Call `GET /goals/:id` on any active goal
2. Call `GET /goals/:id/visualization`
3. Call `GET /srm/status/:goalId`
4. Inspect all response bodies

**Expected result:** None of the following fields appear anywhere in any response: `drift`, `chaos_index`, `continuity`, `weights`, `factors`, `penalties`, `score_components`, numeric thresholds (0.25, 0.40, 0.60).

Allowed fields: `progress_pct` (0–100), `grade` (A+/A/B/C/D), `actual_pct`, `expected_pct`, `trend`, `tier`, `badge`.

**Failure indicator:**
- Any internal computation field present in response → critical rules violation; security fix required immediately

---

## 8. Critical Checkpoints

### 8.1 Server-Side Calculation Enforcement
### 8.2 Opaque API Response Validation
### 8.3 JWT Auth on All Protected Routes
### 8.4 Admin 404 (Non-Admin Access)
### 8.5 Graceful Degradation (AI / Email Down)
### 8.6 Timing-Safe Forgot Password
