## 9. Roadmap Test Mapping

### CURRENT SYSTEM (REALITY)

### TARGET SYSTEM (FRAMEWORK)

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

**Fix:** Add call to `fn_compute_growth_trajectory(goal_id, today)` inside `jobComputeDailyScore` (23:50 UTC) after `db.UpsertGoalScore()`.

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

