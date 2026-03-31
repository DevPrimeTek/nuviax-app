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
2. Simulate scheduler run: trigger `jobComputeDailyScore` (cron: `50 23 * * *` — runs at **23:50 UTC**; or wait 24h in prod)
3. Complete at least one task on Day 1
4. Simulate Day 2: trigger daily score job again (second 23:50 UTC run)
5. `GET /api/v1/goals/:id/visualize`

**Expected result (post SA-1 fix):** `trajectory` array contains ≥2 entries, each with `date`, `actual_pct`, `expected_pct`, `delta`, `trend`.

**Fallback behavior to verify:** If only 1 day has passed (trajectory = 0 DB rows), `GenerateProgressVisualization` returns a single live-computed snapshot — `actual_pct: 0`, `expected_pct > 0`, `trend: "ON_TRACK"`. Not a failure; verify array is never null/empty. ⚠️ Fallback currently broken — see TS-08.

**Failure indicator:**
- `trajectory: null` or `trajectory: []` → SA-1 not applied, or live snapshot fallback failing (CE-1 bug)
- Missing `expected_pct` field → engine returning internal weight data (critical rules violation)
- `404` on endpoint → wrong URL used; route is `/visualize` not `/visualization`

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
1. Trigger active L3 SRM condition (via manual `INSERT INTO srm_events (id, go_id, srm_level, triggered_at) VALUES (gen_random_uuid(), $goalId, 'L3', NOW())` for testing, or via L2 escalation)
2. `GET /api/v1/srm/status/:goalId` → assert `srm_level = "L3"`
3. `POST /api/v1/srm/confirm-l3/:goalId`
4. Check step 3 response body for `new_status` and `frozen_expected` fields
5. `GET /api/v1/goals/:id` → assert `status = "PAUSED"` in goal object
6. `GET /api/v1/goals/:id/visualize` → assert `expected_pct` is a fixed value (does not advance on subsequent calls)

**Expected result:** Step 3 returns `200` with body:
```json
{
  "goal_id": "...",
  "new_status": "PAUSED",
  "frozen_expected": <float 0–1>,
  "message": "SRM Level 3 confirmat...",
  "next_step": "Reactivarea automată va fi propusă după 7 zile de stabilitate."
}
```
`global_objectives.status = 'PAUSED'` in DB. Sprint `expected_pct_frozen = TRUE` in `sprints` table (drift loop prevented, GAP #20).

⚠️ **CE-7 — frozen_expected divergence:** `frozen_expected` in the API response (step 3) is computed using `goal.StartDate/goal.EndDate`. The value stored in `sprints.frozen_expected_pct` (used by trajectory engine) is computed using `sprint.StartDate/goal.EndDate`. These values will differ for any goal past Sprint 1. Do not assert both values are equal — test them separately if needed.

**Failure indicator:**
- Step 3 returns `201` → wrong; actual code returns `200` (no explicit status set)
- `new_status` missing from step 3 response → handler changed
- Goal status still `ACTIVE` after step 5 → L3 confirm not updating `global_objectives`
- `frozen_expected = 0` in step 3 when goal started >0 days ago → elapsed time computation broken (check `goal.StartDate`)
- `expected_pct` still advancing in visualize response → `FreezeExpectedTrajectory()` not called or sprint freeze flag not set

---

### TS-07 — Achievement Unlocks and Ceremony Appears

**Steps:**
1. Create goal; complete sprint (30 days of tasks OR manually close sprint via `POST /api/v1/sprints/:sprintId/close` — requires **sprint ID**, not goal ID)
2. Simulate scheduler run: `jobDetectEvolutionSprints` (01:00 UTC) then `jobGenerateCeremonies` (01:05 UTC)
3. `GET /api/v1/ceremonies/:goalId` → assert ceremony present with `ceremony_tier` field = `BRONZE`, `SILVER`, `GOLD`, or `PLATINUM`
4. `GET /api/v1/achievements` → assert response contains `achievements` array ⚠️ will return `{"achievements":[]}` until SA-2 is implemented
5. `POST /api/v1/ceremonies/:ceremonyId/view` → assert `200` with `{"message":"Ceremonie marcată ca vizualizată."}`
6. `GET /api/v1/ceremonies/:goalId` again → assert `ceremony_tier` is still present (viewed flag lives in DB; endpoint returns latest ceremony regardless)

**Expected result (post SA-2 fix):** Sprint closure generates ceremony. `CeremonyModal.tsx` shows on next login via `GET /ceremonies/unviewed`. Achievements recorded. `viewed` flag prevents re-display in modal.

**Current behavior (pre-fix):** Ceremony generated correctly, `ceremony_tier` populated. `GET /achievements` returns `{"achievements":[]}` — `fn_award_achievement_if_earned()` never called. ⚠️ SA-2 NOT IMPLEMENTED.

**Failure indicator:**
- `GET /ceremonies/:goalId` returns `404` after sprint close → `jobGenerateCeremonies` not running, or sprint not closed with `status = 'COMPLETED'`
- `ceremony_tier` missing from response → `engine.GenerateCompletionCeremony()` failing silently
- `POST /sprints/:id/close` returns `404` → wrong ID used (must be sprint UUID, not goal UUID)
- `achievements` key missing from `GET /achievements` response → handler changed (currently wraps in `{"achievements":[...]}`)
- Step 5 returns `500` → `mark_ceremony_viewed()` DB function missing or failing

---

### TS-08 — Visualization Is Not Empty on Day 1

**Steps:**
1. Create active goal
2. Immediately call `GET /api/v1/goals/:id/visualize` (before any scheduler run)

**Expected result (post CE-1 bugfix):** Response contains a `trajectory` array with exactly 1 entry (live snapshot):
```json
{
  "trajectory": [
    { "actual_pct": 0, "expected_pct": <float > 0>, "delta": <negative float>, "trend": "ON_TRACK" }
  ]
}
```

⚠️ **CURRENTLY FAILS — CE-1 (production bug):** Fallback query in `level5_growth.go:85` reads `FROM goals` but table is `global_objectives`. Query silently returns no rows → fallback snapshot not inserted → API returns `trajectory: null`. This is a code bug, not a missing feature.

**Fix required:** Change `FROM goals` → `FROM global_objectives` at `level5_growth.go:85`. This is independent of SA-1.

**Failure indicator:**
- `trajectory: null` → CE-1 table name bug not yet fixed (`level5_growth.go:85`)
- `trajectory: []` → fallback returns empty array instead of nil; acceptable after fix if at least 1 entry expected
- `expected_pct = 0` on a goal started 1+ days ago → start/end date computation broken in fallback
- `404` on request → wrong URL used; route is `/visualize` not `/visualization`

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
1. `GET /api/v1/goals/:id` on any active goal — pipe through `jq 'keys'`
2. `GET /api/v1/goals/:id/visualize` — pipe through `jq '.trajectory[0] | keys'`
3. `GET /api/v1/srm/status/:goalId` — pipe through `jq 'keys'`
4. `GET /api/v1/goals/:id/progress` — pipe through `jq 'keys'`
5. Inspect all response keys

**Expected result:** None of the following fields appear in any response: `drift`, `chaos_index`, `continuity`, `weights`, `factors`, `penalties`, `score_components`, numeric thresholds (0.25, 0.40, 0.60).

```bash
# Quick check
curl -s -H "Authorization: Bearer $TOKEN" https://api.nuviax.app/api/v1/goals/$GOAL_ID \
  | jq '[keys[] | select(. == "drift" or . == "chaos_index" or . == "weights")]'
# Must return: []
```

⚠️ **CE-8 — `score` field conflict:** Checkpoint 8.1 lists `score (opaque float 0–1)` as an allowed key. CLAUDE.md "EXPOSE ONLY" list does not include a raw float score — only `Progress %`, `Grade`, `Ceremony`, `Achievements`. `GET /goals/:id/progress` currently returns `{"score": <float>, "grade": ..., "progress_pct": ...}`. Whether the `score` float violates CLAUDE.md is **unresolved** — treat any raw score float in a goal response as a flag for review.

**Failure indicator:**
- Any of the forbidden fields present → critical rules violation; do not deploy; fix immediately
- `score` float appears in `GET /goals/:id` main response → review against CLAUDE.md critical rules
- `404` on step 2 → wrong URL used; route is `/visualize` not `/visualization`

---

### TS-13 — WAITING Goal Promoted to ACTIVE Has Tasks Available Same Day

**Context:** This tests scheduler-driven promotion only. Manual activation via `POST /api/v1/goals/:id/activate` is unaffected — that handler already calls `GenerateDailyTasks()` immediately.

**Steps:**
1. Create a 4th goal when 3 are already active → assert response has `"vaulted": true`, `"status": "WAITING"`
2. Archive one active goal (`DELETE /api/v1/goals/:id`) to free a slot
3. Simulate `activateWaitingGoal` scheduler run (nightly job) — or wait for midnight UTC
4. `GET /api/v1/goals` → assert the previously-WAITING goal now appears in `goals` array with `status = "ACTIVE"`
5. `GET /api/v1/today` → assert `main_tasks` is non-empty on the same calendar day as promotion

**Expected result (post M-2 fix):** Promoted goal's Sprint 1 is created and `main_tasks` are generated within the same scheduler run. `GET /today` returns tasks immediately — no 24-hour gap.

⚠️ **CURRENTLY FAILS — M-2 (known gap):** `activateWaitingGoal` creates the sprint but does NOT call `engine.GenerateDailyTasks()`. Tasks will not appear until the next 00:00 UTC cycle (~24 hour gap). Manual activation via `POST /goals/:id/activate` works correctly.

**Fix required:** Add `engine.GenerateDailyTasks(ctx, userID, today)` inside `activateWaitingGoal` in `scheduler.go` after sprint creation.

**Failure indicator:**
- Step 5: `main_tasks: []` after scheduler promotion → M-2 gap not yet fixed
- Step 5: `main_tasks` non-empty after manual `POST /goals/:id/activate` → confirms manual path works, only scheduler path broken
- Step 4: goal still shows `status = "WAITING"` → slot not freed or scheduler not run

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

**Allowed response keys:** `progress_pct`, `grade`, `grade_label`, `actual_pct`, `expected_pct`, `delta`, `trend`, `tier`, `badge_type`.

⚠️ **CE-8 — Unresolved conflict on `score` field:** `GET /goals/:id/progress` currently returns `score` (opaque float 0–1). CLAUDE.md "EXPOSE ONLY" list specifies: Progress %, Grade, Ceremony, Achievements — no raw float. Until this is resolved:
- Do NOT add new endpoints that return raw `score` floats
- Existing `score` in `/goals/:id/progress` and dashboard responses must be reviewed before Sprint 4
- Treat any `score` float in a goal detail response (`GET /goals/:id`) as a violation

**Failure looks like:** Any of the explicitly forbidden fields (`drift`, `chaos_index`, `continuity`, `weights`, `factors`, `penalties`) present in any API response body. Immediate fix required — do not deploy.

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
