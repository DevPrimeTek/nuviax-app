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

Achievements and ceremonies are evaluated **after sprint close only** — not on individual task completion.

**Trigger chain (scheduler, daily at 01:00–01:05 UTC):**

1. `jobCloseExpiredSprints` (00:01 UTC) — sets `sprints.status = 'COMPLETED'` for sprints past their `end_date`
2. `jobDetectEvolutionSprints` (01:00 UTC) — queries sprints completed yesterday; calls `engine.MarkEvolutionSprint()` per sprint
   - Evolution condition: `current_sprint_score - prev_sprint_score >= 0.05` (delta threshold)
   - G-11 override: if `dominant_behavior_model` set, `ApplyEvolveOverride()` applies model-specific thresholds:
     - `ANALYTIC`: requires consistency >= 0.75 in addition to delta
     - `TACTICAL`: delta threshold lowered to 0.02 (more responsive to quick wins)
     - `STRATEGIC`: standard delta threshold (0.05)
     - `REACTIVE`: adaptive threshold based on recent performance volatility
   - Evolution detected → INSERT into `evolution_sprints` (idempotent via `ON CONFLICT sprint_id DO NOTHING`)
3. `jobGenerateCeremonies` (01:05 UTC) — queries sprints completed yesterday with no existing ceremony; calls `engine.GenerateCompletionCeremony()`

**⚠ SA-2 known gap:** `fn_award_achievement_if_earned()` exists in migration 006 but is never called from Go. `achievement_badges` is not populated by the scheduler. Badges only appear if inserted directly via DB. This is an open CRITICAL fix in Sprint 3.1.

### 5.2 Ceremony Tiers

Tier assignment in `engine.GenerateCompletionCeremony()` (`level5_growth.go:185`):

| Tier | Condition |
|---|---|
| `BRONZE` | `score < 0.75` (any sprint) |
| `SILVER` | `score >= 0.75` |
| `GOLD` | `score >= 0.90` AND not an evolution sprint |
| `PLATINUM` | `score >= 0.90` AND `isEvolution = true` |

- Score = `engine.ComputeSprintScore()` — opaque value; never exposed raw
- `isEvolution` flag passed from `jobGenerateCeremonies` query via `evolution_sprints` join
- Ceremony stored in `completion_ceremonies`: `sprint_id`, `go_id`, `ceremony_tier`, `viewed = false`
- On conflict (`ON CONFLICT sprint_id DO NOTHING`) — ceremony generated exactly once per sprint

### 5.3 Badge Storage and Award

**Tables involved:**
- `achievement_badges` — awarded badges: `id`, `user_id`, `badge_type`, `go_id`, `sprint_id`, `awarded_at`
- `completion_ceremonies` — sprint ceremonies: `id`, `sprint_id`, `go_id`, `ceremony_tier`, `viewed`, `generated_at`
- `evolution_sprints` — evolution markers: `sprint_id`, `evolution_score`, `delta_performance`, `consistency_weight`

**Read path (`GET /achievements`):**
1. Handler calls `engine.GetUserAchievements(ctx, userID)`
2. Query: `SELECT ... FROM achievement_badges WHERE user_id = $1 ORDER BY awarded_at DESC`
3. Returns `[]models.AchievementBadge` — never null, empty array `[]` on no badges (nil guard in handler)

**Progress path (`GET /achievements/progress`):**
1. Handler calls DB function directly: `SELECT * FROM get_achievement_progress($1)`
2. Returns progress toward each badge type (from migration 006)

**`fn_award_achievement_if_earned()` — when it should be called:**
- After each `jobCloseExpiredSprints` for the closed sprint
- After `jobDetectEvolutionSprints` when evolution is detected
- Currently not called (SA-2 gap) — must be wired to scheduler in Sprint 3.1

### 5.4 Frontend Display

**`/achievements` page (`achievements/page.tsx`):**
1. Fetches `GET /achievements` → renders badge grid from `achievements` array
2. Empty state: `achievements: []` → renders empty grid (no error shown)
3. Fetches `GET /achievements/progress` → renders progress bars per badge type

**Ceremony modal (`CeremonyModal.tsx`):**
1. Dashboard checks `GET /ceremonies/latest` on each login
2. If `viewed = false` → `CeremonyModal` rendered with tier (BRONZE/SILVER/GOLD/PLATINUM) and message
3. User dismisses → `POST /ceremonies/:id/viewed` → `viewed = true` in DB → modal not shown again
4. Colors/icons vary by tier — defined in `CeremonyModal.tsx` component

**`/profile` page:**
- Does not show achievement history directly; links to `/achievements`
- Shows activity heatmap (`ActivityHeatmap.tsx`) and stats — separate from badge system

---

## 6. Visualization Flow

### 6.1 Data Source: `growth_trajectories`

**Table schema (migration 006):**
- `go_id`, `snapshot_date`, `actual_pct`, `expected_pct`, `delta`, `trend`
- `trend` values: `ON_TRACK`, `AHEAD`, `BEHIND`, `CRITICAL`

**Population (SA-1 known gap):**
- `fn_compute_growth_trajectory()` SQL function exists in migration 006
- Currently **not called** from any Go scheduler job → table remains empty for all users
- Fix required in Sprint 3.1: call from `jobComputeDailyScore` (22:00 UTC) after `UpsertGoalScore()`

**Expected flow (post SA-1 fix):**
- `jobComputeDailyScore` runs daily at 22:00 UTC
- For each ACTIVE goal: computes score, upserts `go_scores`, then calls `fn_compute_growth_trajectory(goal_id, today)`
- One row inserted per day per goal into `growth_trajectories`
- After N days: N data points in trajectory → charts become meaningful

### 6.2 Fallback Logic (Single Snapshot)

When `growth_trajectories` is empty for a goal, `GenerateProgressVisualization()` (`level5_growth.go:82`) computes a live synthetic snapshot:

```
elapsed = now - goal.start_date
total   = goal.end_date - goal.start_date
expected_pct = elapsed / total  (time-linear)
actual_pct   = 0
delta        = -expected_pct
trend        = "ON_TRACK"
```

Returns array with exactly 1 entry. `trajectory` is never null or empty — this is guaranteed by the fallback.

### 6.3 Trajectory Freeze (SRM L3)

When SRM L3 is confirmed (`POST /srm/confirm-l3`):
- `FreezeExpectedTrajectory(sprint.ID)` sets `sprints.expected_pct_frozen = TRUE`, `frozen_expected_pct = <current_elapsed_ratio>`
- During visualization, `computeProgressVsExpected()` reads freeze flag: if frozen → uses stored `frozen_expected_pct` instead of real-time elapsed ratio
- Effect: `expected_pct` stops advancing while user is in stabilization mode
- Prevents drift loop paradox: score does not worsen while protocol is followed correctly
- Unfreeze: `UnfreezeExpectedTrajectory(sprint.ID)` called on reactivation

### 6.4 API Contract

**Endpoint:** `GET /api/v1/goals/:id/visualize`

**Auth:** JWT required; returns `404` if goal doesn't belong to caller.

**Response:**
```json
{
  "goal_id": "uuid",
  "trajectory": [
    {
      "date": "2026-03-28T00:00:00Z",
      "actual_pct": 0.42,
      "expected_pct": 0.45,
      "delta": -0.03,
      "trend": "ON_TRACK"
    }
  ]
}
```

**Never exposed:** raw score, drift, chaos_index, weights, thresholds — only `actual_pct`, `expected_pct`, `delta`, `trend`.

### 6.5 Progress Bar and Grade Display

**`GET /goals/:id`** returns:
- `progress_pct`: 0–100 integer (from `engine.ComputeProgressPct()`)
- `grade`: opaque string: `A+`, `A`, `B`, `C`, `D`
- `grade_label`: localized string (currently always Romanian: `auth.GradeLabel(grade, "ro")`)
- `days_left`: computed from current sprint `end_date` (not goal `end_date`) — B-3 fix

Frontend `GoalTabs.tsx` renders progress bar width from `progress_pct`, grade badge from `grade`.

### 6.6 Activity Heatmap

**Endpoint:** `GET /api/v1/profile/activity` — returns 365-day activity data.

**`ActivityHeatmap.tsx`:**
- Pure CSS grid, 52 columns (weeks) × 7 rows (days)
- Color scale based on completion rate per day (0 tasks → lightest, all tasks → darkest)
- Hover tooltip shows date + task count
- Rendered on `/profile` page below preferences section

### 6.7 Frontend Chart Component

**`ProgressCharts.tsx`** (Recharts library):
- `LineChart`: `actual_pct` vs `expected_pct` over time — shows trajectory divergence
- `BarChart`: per-sprint score comparison — shows evolution across sprints
- If trajectory has 1 point (fallback state): line chart renders as a single dot — not an error, expected until SA-1 is fixed
- Chart data fed from `GET /goals/:id/visualize` response; no client-side computation

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

**What must never break:** All score computation runs in Go engine only. No formula, weight, factor, or threshold may appear in any API response.

**How to verify:**
```bash
curl -H "Authorization: Bearer $TOKEN" https://api.nuviax.app/api/v1/goals/$GOAL_ID | \
  jq 'keys'
# Must NOT contain: drift, chaos_index, continuity, weights, factors,
#                   penalties, score_components, thresholds
```

**Allowed response keys:** `progress_pct`, `grade`, `grade_label`, `score` (opaque float 0–1), `actual_pct`, `expected_pct`, `delta`, `trend`, `tier`, `badge_type`.

**Failure looks like:** Any of the forbidden fields present in JSON response body. Immediate fix required — do not deploy.

---

### 8.2 Opaque API Response Validation

**What must never break:** Clients receive only grades (A+/A/B/C/D) and percentages. The numeric computation chain (C1–C40) is internal.

**How to verify:** Run TS-12 — inspect all goal, visualization, and SRM status responses. Use `jq 'to_entries[] | .key'` to enumerate all keys.

**Failure looks like:** `chaos_index: 0.42` or `weight_c7: 0.33` appearing in any API response. Even a logging endpoint must not expose these.

---

### 8.3 JWT Auth on All Protected Routes

**What must never break:** Every route except `/auth/login`, `/auth/register`, `/auth/forgot-password`, `/auth/reset-password`, and `/health` requires a valid JWT RS256 access token.

**How to verify:**
```bash
# Should return 401
curl https://api.nuviax.app/api/v1/goals
curl https://api.nuviax.app/api/v1/today
curl https://api.nuviax.app/api/v1/achievements
```

**Token behavior:** Access token expires in 15 minutes. Frontend proxy (`api/proxy/[...path]`) auto-refreshes using refresh token cookie. If refresh token expired → redirect to `/auth/login`.

**Failure looks like:** Any protected route returning `200` without `Authorization` header. Or `500` instead of `401` on missing token.

---

### 8.4 Admin 404 (Non-Admin Access)

**What must never break:** Admin panel returns `404` — not `403` — for non-admin users. The existence of the admin route must not be detectable.

**How to verify:**
```bash
# With a regular user token
curl -H "Authorization: Bearer $REGULAR_TOKEN" \
  https://api.nuviax.app/api/v1/admin/stats
# Must return 404, not 403 or 401
```

**Enforced by:** `middleware/admin.go` — checks `is_admin = TRUE` on `users` table; calls `notFound(c)` on failure (returns 404 body identical to other 404 responses).

**Failure looks like:** `403 Forbidden` (reveals route exists), or `401` (reveals route is protected). Any non-404 response is a disclosure failure.

---

### 8.5 Graceful Degradation (AI + Email Down)

**What must never break:** If Anthropic API is unreachable, onboarding continues — AI suggestion silently returns empty. If Resend is unreachable, registration succeeds — welcome email silently fails.

**How to verify (AI):**
```bash
# With ANTHROPIC_API_KEY unset or invalid
curl -X POST https://api.nuviax.app/api/v1/goals/suggest-category \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"goal_name":"Learn piano"}'
# Must return 200 with empty/null suggestion within 2 seconds
```

**How to verify (Email):**
- Set `RESEND_API_KEY=invalid` → `POST /auth/register` → must still return `201`

**Timeout enforcement:** `SuggestGOCategory()` has a 2-second hard context timeout. Response must arrive within that window regardless of Anthropic upstream latency.

**Failure looks like:** `500` error on registration when email fails. Or onboarding hanging >3 seconds when AI is down. Or `400/500` from suggest-category endpoint.

---

### 8.6 Timing-Safe Forgot Password

**What must never break:** `POST /auth/forgot-password` always returns `200` with a neutral message, regardless of whether the email exists in the DB. Response time must not differ between known and unknown emails.

**How to verify:**
```bash
# With known email
curl -X POST https://api.nuviax.app/api/v1/auth/forgot-password \
  -d '{"email":"real@user.com"}'
# → 200 {"message":"..."}

# With unknown email
curl -X POST https://api.nuviax.app/api/v1/auth/forgot-password \
  -d '{"email":"nobody@nowhere.com"}'
# → 200 {"message":"..."} — identical response
```

**Failure looks like:** `404` or `422` when email not found (user enumeration). Or measurably different response time between the two cases (timing side-channel).

---

## 9. Roadmap Test Mapping

Maps Sprint 3.1 System Alignment fixes (SA-1 through SA-7) to the test scenarios that verify them.

| Fix | Description | Verified By | Priority |
|---|---|---|---|
| SA-1 | `growth_trajectories` populated by scheduler | TS-03, TS-08 | CRITICAL |
| SA-2 | `fn_award_achievement_if_earned()` called from Go | TS-07 | HIGH |
| SA-3 | SRM L1 auto-trigger wired in `jobCheckDailyProgress` | TS-04 | CRITICAL |
| SA-4 | SRM L2 confirm creates `ENERGY_LOW` context adjustment | TS-05 | CRITICAL |
| SA-5 | `SRMWarning.tsx` L2 confirm button present in UI | TS-05 (frontend) | CRITICAL |
| SA-6 | `jobCheckSRMTimeouts` applies fallback state change | TS-06 | HIGH |
| SA-7 | `jobRecalibrateRelevance` cron `*/90` → `*/7` fix | TS-04 (indirect) | HIGH |

---

### SA-1 → TS-03, TS-08

**Fix:** Add call to `fn_compute_growth_trajectory(goal_id, today)` inside `jobComputeDailyScore` after `db.UpsertGoalScore()`.

**TS-03 verifies:** After 2 scheduler runs, `GET /goals/:id/visualize` returns ≥2 trajectory entries.

**TS-08 verifies:** On Day 1 (before any scheduler run), live fallback returns exactly 1 entry — not empty.

---

### SA-2 → TS-07

**Fix:** Call `fn_award_achievement_if_earned(user_id, sprint_id)` inside `jobGenerateCeremonies` after each successful `GenerateCompletionCeremony()`.

**TS-07 verifies:** `GET /achievements` returns non-empty array after sprint close. `achievement_badges` has row for the user.

---

### SA-3 → TS-04

**Fix:** In `jobCheckDailyProgress` — after regression detection loop — call `engine.CheckAndRecordRegressionEvent()` and insert into `srm_events` with `srm_level = 'L1'` when regression detected.

**TS-04 verifies:** After 5 consecutive missed days, `GET /srm/status/:goalId` returns `srm_level = "L1"`.

---

### SA-4 → TS-05 (backend)

**Fix:** In `ConfirmSRML2()` (`srm.go`) — after stamping `confirmed_at` — call `db.CreateContextAdjustment()` with `adjType = AdjEnergyLow` starting tomorrow, to actually reduce next-day task intensity.

**TS-05 verifies:** Task count the day after L2 confirmation is lower than baseline day.

---

### SA-5 → TS-05 (frontend)

**Fix:** In `SRMWarning.tsx` — add conditional confirm button when `srm_level === 'L2'`; on click call `POST /srm/confirm-l2/:goalId`; on success refresh SRM status.

**TS-05 verifies:** L2 banner has actionable button; confirmation dismisses banner without page reload.

---

### SA-6 → TS-06

**Fix:** Replace `// TODO: engine.ApplySRMFallback(ctx, goalID, fallback)` with actual state application — insert `srm_events` row with computed fallback level; if fallback = `L1`, reduce intensity without pausing.

**TS-06 verifies:** Goal with unconfirmed L3 after timeout does not remain blocked indefinitely.

---

### SA-7 → TS-04 (indirect)

**Fix:** Change cron expression `"0 2 */90 * *"` → `"0 2 * * 0"` (weekly Sunday at 02:00 UTC) or `"0 2 */7 * *"` — `*/90` is invalid in day-of-month field.

**TS-04 indirect:** `jobRecalibrateRelevance` must run for `CheckChaosIndex()` to evaluate L2 threshold. Without this fix, L2 auto-trigger (which TS-05 depends on) never fires.

---

## 10. Post-Fix Validation Checklist

Run after all SA-1 through SA-7 fixes are deployed. All items must pass before Sprint 3.1 is closed.

- [ ] **TS-03** — `GET /goals/:id/visualize` returns ≥2 trajectory entries after 2 scheduler runs
- [ ] **TS-04** — `GET /srm/status/:goalId` returns `srm_level: "L1"` after 5 consecutive missed days
- [ ] **TS-05** — SRM L2 banner has confirm button; after confirm, next-day task count is reduced
- [ ] **TS-06** — L3 unconfirmed >N hours → fallback applied; goal not stuck indefinitely
- [ ] **TS-07** — Sprint close → `GET /achievements` returns ≥1 badge; `GET /ceremonies/latest` returns ceremony with correct tier
- [ ] **TS-08** — Day 1 visualization returns exactly 1 entry; `trajectory` never null or empty
- [ ] **TS-12** — Zero internal fields (`drift`, `chaos_index`, `weights`, thresholds) in any API response
- [ ] **8.3** — All protected routes return `401` without Authorization header
- [ ] **8.4** — Admin routes return `404` for non-admin users
- [ ] **8.6** — `POST /auth/forgot-password` returns `200` for both known and unknown emails
- [ ] **Cron fix (SA-7)** — `jobRecalibrateRelevance` runs without error; verify via scheduler logs
