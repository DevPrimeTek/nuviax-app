# docs/testing/flows/visualization.md — Visualization Flow

> Part of: `docs/testing/` | Scenarios: `scenarios/critical.md` (TS-03, TS-08, TS-10, TS-12)

---

## Current System (Reality)

**`growth_trajectories` population:**
- ❌ `fn_compute_growth_trajectory()` exists in migration 006 but is NEVER called from Go — table empty for all users (SA-1)
- ❌ `jobComputeDailyScore` does NOT call the trajectory function after `UpsertGoalScore()`

**Fallback (Day 1 snapshot):**
- ❌ `GenerateProgressVisualization()` (`level5_growth.go:85`) queries `FROM goals` — wrong table; actual table is `global_objectives` — query returns no rows silently
- ❌ `trajectory: null` in API response on Day 1 — TS-08 fails in production

**Trajectory freeze (SRM L3):**
- ✅ `FreezeExpectedTrajectory(sprint.ID)` sets `expected_pct_frozen = TRUE` + stores `frozen_expected_pct`
- ✅ `computeProgressVsExpected()` respects freeze flag — `expected_pct` stops advancing during stabilization

**Progress bar & grade (`GET /goals/:id`):**
- ✅ `progress_pct` (0–100) returned from `engine.ComputeProgressPct()`
- ✅ `grade` (A+/A/B/C/D) returned — opaque; no internal components exposed
- ✅ `days_left` computed from sprint `end_date` (B-3 fix)

**Activity heatmap:**
- ✅ `GET /profile/activity` returns 365-day data
- ✅ `ActivityHeatmap.tsx` — 52×7 CSS grid, color scale, hover tooltip

**Frontend charts:**
- ✅ `ProgressCharts.tsx` renders LineChart + BarChart from `GET /goals/:id/visualize`
- ❌ LineChart renders single dot when trajectory empty (expected degraded state until SA-1 fixed)

---

## Target System (Framework)

**`growth_trajectories` population:**
- ✅ `jobComputeDailyScore` (22:00 UTC) calls `fn_compute_growth_trajectory(goal_id, today)` after each `UpsertGoalScore()`
- ✅ One row per day per ACTIVE goal → N days = N data points → charts meaningful

**Fallback (Day 1 snapshot):**
- ✅ `GenerateProgressVisualization()` queries `FROM global_objectives` correctly
- ✅ Returns exactly 1 synthetic entry: `actual_pct: 0`, `expected_pct > 0` (time-linear), `trend: "ON_TRACK"`
- ✅ `trajectory` never null or empty — TS-08 passes

**Trajectory freeze (SRM L3):**
- ✅ `expected_pct` frozen at confirmed-L3 moment; drift loop paradox prevented (GAP #20)
- ✅ `UnfreezeExpectedTrajectory()` called on goal reactivation — expected_pct resumes advancing

**Progress bar & grade:**
- ✅ All score components (drift, chaos_index, weights, thresholds) server-only — never in response
- ✅ `grade_label` localized per user language (not hardcoded `"ro"`)

**Activity heatmap:**
- ✅ Completion rate per day drives color scale; empty days distinct from missing data

**Frontend charts:**
- ✅ ≥2 trajectory points → LineChart shows divergence between `actual_pct` and `expected_pct`
- ✅ BarChart shows per-sprint score evolution across sprint history

---

## 6.1 Data Source: `growth_trajectories`

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

## 6.2 Fallback Logic (Single Snapshot)

When `growth_trajectories` is empty for a goal, `GenerateProgressVisualization()` (`level5_growth.go:82`) attempts to compute a live synthetic snapshot:

```
elapsed = now - goal.start_date
total   = goal.end_date - goal.start_date
expected_pct = elapsed / total  (time-linear)
actual_pct   = 0
delta        = -expected_pct
trend        = "ON_TRACK"
```

**⚠️ Known bug:** The fallback query at `level5_growth.go:85` reads `FROM goals` but the actual table is `global_objectives`. The query silently returns no rows → fallback snapshot is never inserted → `trajectory` is `null` in the API response. **TS-08 currently fails in production.** Fix: change `FROM goals` → `FROM global_objectives` in `GenerateProgressVisualization()`.

## 6.3 Trajectory Freeze (SRM L3)

When SRM L3 is confirmed (`POST /srm/confirm-l3`):
- `FreezeExpectedTrajectory(sprint.ID)` sets `sprints.expected_pct_frozen = TRUE`, `frozen_expected_pct = <current_elapsed_ratio>`
- During visualization, `computeProgressVsExpected()` reads freeze flag: if frozen → uses stored `frozen_expected_pct` instead of real-time elapsed ratio
- Effect: `expected_pct` stops advancing while user is in stabilization mode
- Prevents drift loop paradox: score does not worsen while protocol is followed correctly
- Unfreeze: `UnfreezeExpectedTrajectory(sprint.ID)` called on reactivation

## 6.4 API Contract

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

## 6.5 Progress Bar and Grade Display

**`GET /goals/:id`** returns:
- `progress_pct`: 0–100 integer (from `engine.ComputeProgressPct()`)
- `grade`: opaque string: `A+`, `A`, `B`, `C`, `D`
- `grade_label`: localized string (currently always Romanian: `auth.GradeLabel(grade, "ro")`)
- `days_left`: computed from current sprint `end_date` (not goal `end_date`) — B-3 fix

Frontend `GoalTabs.tsx` renders progress bar width from `progress_pct`, grade badge from `grade`.

## 6.6 Activity Heatmap

**Endpoint:** `GET /api/v1/profile/activity` — returns 365-day activity data.

**`ActivityHeatmap.tsx`:**
- Pure CSS grid, 52 columns (weeks) × 7 rows (days)
- Color scale based on completion rate per day (0 tasks → lightest, all tasks → darkest)
- Hover tooltip shows date + task count
- Rendered on `/profile` page below preferences section

## 6.7 Frontend Chart Component

**`ProgressCharts.tsx`** (Recharts library):
- `LineChart`: `actual_pct` vs `expected_pct` over time — shows trajectory divergence
- `BarChart`: per-sprint score comparison — shows evolution across sprints
- If trajectory has 1 point (fallback state): line chart renders as a single dot — not an error, expected until SA-1 is fixed
- Chart data fed from `GET /goals/:id/visualize` response; no client-side computation

## 6.8 Theme and Language Persistence

**`PATCH /settings`** persists `theme` (dark/light) and language to DB (`users.theme` — migration 012).

**`GET /settings`** returns:
- `theme`: `"dark"` or `"light"` (default: `"dark"`)
- Language preference

**Frontend flow:**
- On login: `GET /settings` → sync `nv_theme` to `localStorage` → `data-theme` applied
- Anti-flash inline script in `layout.tsx` reads `localStorage` and applies `data-theme` before React hydration
- `AppShell.tsx` toggle button (sun/moon icon) updates both DB and `localStorage`

**Test scenario:** TS-10 — Theme and Language Persist Across Sessions
