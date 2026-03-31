# docs/testing/scenarios/regression.md — Regression & Sprint 3.1 Fix Mapping

> Part of: `docs/testing/` | See `scenarios/critical.md` for full TS-xx scenario steps

---

## SA Fix → Test Scenario Mapping

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

## SA-1 → TS-03, TS-08

**Fix:** Add call to `fn_compute_growth_trajectory(goal_id, today)` inside `jobComputeDailyScore` (cron: `50 23 * * *` — **23:50 UTC**) after `db.UpsertGoalScore()`.

**TS-03 verifies:** After 2 scheduler runs, `GET /goals/:id/visualize` returns ≥2 trajectory entries.

**TS-08 verifies:** On Day 1 (before any scheduler run), live fallback returns exactly 1 entry — not empty.

---

## SA-2 → TS-07

**Fix:** Call `fn_award_achievement_if_earned(user_id, sprint_id)` inside `jobGenerateCeremonies` after each successful `GenerateCompletionCeremony()`.

**TS-07 verifies:** `GET /achievements` returns non-empty array after sprint close. `achievement_badges` has row for the user.

---

## SA-3 → TS-04

**Fix:** In `jobCheckDailyProgress` — after regression detection loop — call `engine.CheckAndRecordRegressionEvent()` and insert into `srm_events` with `srm_level = 'L1'` when regression detected.

**TS-04 verifies:** After 5 consecutive missed days, `GET /srm/status/:goalId` returns `srm_level = "L1"`.

---

## SA-4 → TS-05 (backend)

**Fix:** In `ConfirmSRML2()` (`srm.go`) — after stamping `confirmed_at` — call `db.CreateContextAdjustment()` with `adjType = AdjEnergyLow` starting tomorrow, to actually reduce next-day task intensity.

**TS-05 verifies:** Task count the day after L2 confirmation is lower than baseline day.

---

## SA-5 → TS-05 (frontend)

**Fix:** In `SRMWarning.tsx` — add conditional confirm button when `srm_level === 'L2'`; on click call `POST /srm/confirm-l2/:goalId`; on success refresh SRM status.

**TS-05 verifies:** L2 banner has actionable button; confirmation dismisses banner without page reload.

---

## SA-6 → TS-06

**Fix:** Replace `// TODO: engine.ApplySRMFallback(ctx, goalID, fallback)` with actual state application — insert `srm_events` row with computed fallback level; if fallback = `L1`, reduce intensity without pausing.

**TS-06 verifies:** Goal with unconfirmed L3 after timeout does not remain blocked indefinitely.

---

## SA-7 → TS-04 (indirect)

**Fix:** Change cron expression `"0 2 */90 * *"` → `"0 2 * * 0"` (weekly Sunday at 02:00 UTC) or `"0 2 */7 * *"` — `*/90` is invalid in day-of-month field.

**TS-04 indirect:** `jobRecalibrateRelevance` must run for `CheckChaosIndex()` to evaluate L2 threshold. Without this fix, L2 auto-trigger (which TS-05 depends on) never fires.

---

## Post-Fix Validation Checklist

Run after all SA-1 through SA-7 fixes are deployed. All items must pass before Sprint 3.1 is closed.

- [ ] **TS-03** — `GET /goals/:id/visualize` returns ≥2 trajectory entries after 2 scheduler runs
- [ ] **TS-04** — `GET /srm/status/:goalId` returns `srm_level: "L1"` after 5 consecutive missed days
- [ ] **TS-05** — SRM L2 banner has confirm button; after confirm, next-day task count is reduced
- [ ] **TS-06** — L3 unconfirmed >N hours → fallback applied; goal not stuck indefinitely
- [ ] **TS-07** — Sprint close → `GET /achievements` returns `achievements` array with ≥1 badge; `GET /ceremonies/:goalId` returns ceremony with `ceremony_tier` field set
- [ ] **TS-08** — Day 1 visualization returns exactly 1 entry; `trajectory` never null or empty
- [ ] **TS-12** — Zero internal fields (`drift`, `chaos_index`, `weights`, thresholds) in any API response
- [ ] **8.3** — All protected routes return `401` without Authorization header
- [ ] **8.4** — Admin routes return `404` for non-admin users
- [ ] **8.6** — `POST /auth/forgot-password` returns `200` for both known and unknown emails
- [ ] **Cron fix (SA-7)** — `jobRecalibrateRelevance` runs without error; verify via scheduler logs
- [ ] **CE-1 bugfix** — `GET /goals/:id/visualize` on Day 1 returns `trajectory` with exactly 1 entry (not null); fix: `FROM goals` → `FROM global_objectives` at `level5_growth.go:85`
- [ ] **CE-7 review** — `frozen_expected` in `POST /srm/confirm-l3` response matches expectation; note DB stores sprint-based value, API returns goal-based value — verify acceptable divergence
- [ ] **CE-8 review** — Audit all handlers for `score` float field in responses; resolve conflict with CLAUDE.md "EXPOSE ONLY" rules before Sprint 4
- [ ] **TS-13** — WAITING goal promoted to ACTIVE via `POST /goals/:id/activate`; `GET /today` returns non-empty `main_tasks` on the same day

---

## Additional Known Gaps (Post-Validation)

### CE-7 — frozen_expected DB/API Divergence → TS-06

**Issue:** Two different calculations produce two different `frozen_expected` values:

- **DB value** (`FreezeExpectedTrajectory` in `level5_growth.go:285`):
  ```
  total   = goal.EndDate - sprint.StartDate
  elapsed = now - sprint.StartDate
  ```
- **API response value** (`ConfirmSRML3` in `srm.go:132`):
  ```
  total   = goal.EndDate - goal.StartDate
  elapsed = now - goal.StartDate
  ```

For Sprint 1 (sprint.StartDate == goal.StartDate) these are equal. For Sprint 2+, they diverge. The trajectory engine uses the DB value; the response only reflects the API value.

**TS-06 verifies:** `frozen_expected` is non-zero and plausible. Does not assert DB == API value. Divergence must be resolved before multi-sprint freeze behavior is relied upon.

---

### M-2 — WAITING → ACTIVE Promotion Does Not Generate Daily Tasks → TS-13

**Issue:** `activateWaitingGoal` (scheduler, `scheduler.go`) creates Sprint 1 and checkpoints but does **not** call `engine.GenerateDailyTasks()`. A WAITING goal promoted to ACTIVE at midnight will have a sprint but no tasks until the next 00:00 UTC scheduler cycle — a gap of up to 24 hours.

Manual activation via `POST /api/v1/goals/:id/activate` (handler `ActivateGoal`) **does** call `GenerateDailyTasks()` immediately.

**Affected path:** Only the scheduler-driven promotion (nightly auto-activation of vaulted goals). Manual user activation is unaffected.

**TS-13 verifies:** After scheduler promotes a WAITING goal to ACTIVE, `GET /today` returns non-empty `main_tasks` on the same calendar day — currently fails for scheduler-promoted goals.

**Fix required:** Add `engine.GenerateDailyTasks(ctx, userID, today)` call inside `activateWaitingGoal` after sprint creation.
