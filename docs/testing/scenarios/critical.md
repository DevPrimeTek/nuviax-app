# docs/testing/scenarios/critical.md — Critical Test Scenarios & Checkpoints

> Part of: `docs/testing/` | See flows in `flows/` for implementation details

---

## Test Scenarios (TS-01 through TS-12)

### TS-01 — Happy Path: New User Full Journey

**Steps:**
1. `POST /auth/register` with valid name, email, password
2. `POST /auth/login` → receive access token + refresh token cookie
3. `GET /settings` → assert `theme` field present
4. Navigate to `/onboarding` → type goal title → wait for AI suggestion (`POST /goals/suggest-category`)
5. Select category → `POST /goals` with `start_date`, `end_date` (30-day range)
6. `GET /today` → assert `main_tasks` array is non-empty, `day_number = 1`
7. `POST /today/complete/:id` on first task → assert `200`
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
2. Miss all main tasks for 5 consecutive days (do not call `POST /today/complete/:id` on any `MAIN` task)
3. Simulate `jobDetectStagnation` run (scheduler job at 23:58 UTC)
4. `GET /srm/status/:goalId`

**Expected result (post SA-3 fix):** `srm_level = "L1"`, `message = "Ajustare automată activă. Ritmul a fost redus ușor."`. `stagnation_events` has row with `inactive_days >= 5`. `srm_events` has row with `srm_level = 'L1'`.

**Current behavior:** `jobDetectStagnation` populates `stagnation_events` correctly but does NOT write to `srm_events`. `GET /srm/status` returns `"NONE"`. ⚠️ SA-3 NOT IMPLEMENTED.

**Failure indicator:**
- `srm_level = "NONE"` after 5 inactive days → expected until SA-3 is applied
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

**Expected result (post SA-4 fix):** L2 confirmed; `confirmed_at` stamped on `srm_events`; `CreateContextAdjustment(AdjEnergyLow)` called; next day's task count is reduced. Goal status remains `ACTIVE`.

**Current behavior (pre-fix):** `confirmed_at` is stamped but no context adjustment is created. Task count is unchanged the next day. ⚠️ SA-4 NOT IMPLEMENTED.

**Failure indicator:**
- `404` on `POST /srm/confirm-l2` → no active unconfirmed L2 event found
- Task count unchanged after L2 confirmation → SA-4 not yet applied
- Goal status becomes `PAUSED` → L2 incorrectly escalating to L3 behavior

---

### TS-06 — SRM L3 Pauses Goal and Freezes Trajectory

**Steps:**
1. Trigger active L3 SRM condition (via manual `INSERT INTO srm_events` with `srm_level = 'L3'` for testing, or via L2 escalation)
2. `GET /srm/status/:goalId` → assert `srm_level = "L3"`
3. `POST /srm/confirm-l3/:goalId`
4. `GET /goals/:id` → check `status`
5. `GET /goals/:id/visualization` → check `frozen_expected` is a fixed value

**Expected result:** `200` → `new_status: "PAUSED"`, `frozen_expected` is a float between 0–1. `global_objectives.status = 'PAUSED'` in DB. `sprint_trajectories` frozen (drift loop prevented, GAP #20).

**Failure indicator:**
- Goal status still `ACTIVE` → L3 confirm not updating `global_objectives`
- `frozen_expected = 0` when goal is partway through → trajectory freeze not computing elapsed time correctly

---

### TS-07 — Achievement Unlocks and Ceremony Appears

**Steps:**
1. Create goal; complete sprint (30 days of tasks OR manually close sprint via `POST /goals/:id/sprint/close`)
2. Simulate scheduler run: `jobDetectEvolutionSprints` (01:00 UTC) then `jobGenerateCeremonies` (01:05 UTC)
3. `GET /ceremonies/:goalId` → assert ceremony present with `tier` field (BRONZE/SILVER/GOLD/PLATINUM)
4. `GET /achievements` → assert non-empty badge array ⚠️ will return `[]` until SA-2 is implemented
5. `POST /ceremonies/:id/view` → assert `200`
6. `GET /ceremonies/:goalId` → assert returned ceremony has `viewed = true`

**Expected result (post SA-2 fix):** Sprint closure generates ceremony. `CeremonyModal.tsx` shows on next login. Achievements recorded. `viewed` flag correctly prevents re-display.

**Current behavior (pre-fix):** Ceremony is generated correctly. `GET /achievements` returns `[]` — `fn_award_achievement_if_earned()` is never called (SA-2 NOT IMPLEMENTED).

**Failure indicator:**
- `GET /ceremonies/:goalId` returns `404` after sprint close → `jobGenerateCeremonies` not running or sprint close not setting `status = 'COMPLETED'`
- `tier` missing from ceremony → `engine.GenerateCompletionCeremony()` failing silently

---

### TS-08 — Visualization Is Not Empty on Day 1

**Steps:**
1. Create active goal
2. Immediately call `GET /goals/:id/visualization` (before any scheduler run)

**Expected result (post table-name bugfix):** `trajectory` array has exactly 1 entry (live snapshot). `actual_pct: 0`, `expected_pct > 0` (time-based fraction), `trend: "ON_TRACK"`.

**Current behavior:** `trajectory: null` — fallback query uses `FROM goals` (wrong table); fix is `FROM global_objectives` in `level5_growth.go:85`. ⚠️ This test will fail until that bug is fixed.

**Failure indicator:**
- `trajectory: null` → table name bug not yet fixed
- `expected_pct = 0` on a goal started 1+ days ago → start/end date computation broken

---

### TS-09 — Personal Task Limit Enforced

**Steps:**
1. `POST /today/personal` → task 1 created → assert `201`
2. `POST /today/personal` → task 2 created → assert `201`
3. `POST /today/personal` → task 3 attempt → assert `422` with error message

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

## Critical Checkpoints

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
