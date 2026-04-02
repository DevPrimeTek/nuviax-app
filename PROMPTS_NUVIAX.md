# PROMPTS.md — NuviaX: Session-Ready Implementation Prompts
# APPEND BELOW EXISTING CONTENT — starting after last prompt

---

## Sprint 3.1 — Sesiunea 1: SA-7 + CE-1 (cron fix + trajectory table bug)

```
Read CLAUDE.md. Current version: 10.5.0. Branch: claude/sa7-ce1-fix-XXXXX.
Task: fix jobRecalibrateRelevance invalid cron AND fix trajectory table name bug (CE-1).

## Why this first
SA-7 fix unblocks L2 chaos_index evaluation (needed by SA-3).
CE-1 fix ensures GET /goals/:id/visualize never returns trajectory: null on day 1.
Both are 1-line changes — lowest risk, highest unblocking value.

## Files to read (max 3)
1. backend/internal/scheduler/scheduler.go — find jobRecalibrateRelevance cron expression
2. backend/internal/engine/level5_growth.go lines 80-95 — find FROM goals bug

## Changes required

### Fix 1 — SA-7 (scheduler.go)
Find: "0 2 */90 * *"  (or similar invalid day-of-month expression)
Replace with: "0 2 * * 0"  (weekly, Sunday 02:00 UTC)
This is the ONLY change in scheduler.go.

### Fix 2 — CE-1 (level5_growth.go ~line 85)
Find: FROM goals  (inside the live snapshot fallback query)
Replace with: FROM global_objectives
This is the ONLY change in level5_growth.go.

## Verification (do NOT skip)
After changes, grep to confirm:
  grep "*/90" backend/internal/scheduler/scheduler.go  → must return EMPTY
  grep "FROM goals" backend/internal/engine/level5_growth.go  → must return EMPTY

## After
- Update CLAUDE.md section 4: mark SA-7 ✅, CE-1 ✅
- Update docs/testing/test-plan.md: SA-7 → ✅, CE-1 → ✅
- Commit: fix: SA-7 cron expression + CE-1 trajectory table name (level5_growth.go)
- Scenarios now unblocked: TS-04 (indirect), TS-08
```

---

## Sprint 3.1 — Sesiunea 2: SA-1 (growth_trajectories populated by scheduler)

```
Read CLAUDE.md. Current version: 10.5.0. Branch: claude/sa1-trajectories-XXXXX.
Task: wire fn_compute_growth_trajectory call inside jobComputeDailyScore.

## Context
growth_trajectories table exists (migration already applied).
fn_compute_growth_trajectory(goal_id, date) is a PostgreSQL function — already exists.
Problem: it is NEVER called from Go. jobComputeDailyScore runs at 23:50 UTC and
calls db.UpsertGoalScore() but does NOT call fn_compute_growth_trajectory after.

## Files to read (max 3)
1. backend/internal/scheduler/scheduler.go — find jobComputeDailyScore function
2. backend/internal/db/queries.go — check if ComputeGrowthTrajectory DB call exists
3. backend/internal/db/queries.go lines around UpsertGoalScore — understand signature

## Changes required

### In scheduler.go — inside jobComputeDailyScore
After the line that calls db.UpsertGoalScore(ctx, goalID, score), add:

  if err := db.ComputeGrowthTrajectory(ctx, goalID, time.Now()); err != nil {
      log.Printf("[scheduler] growth trajectory failed for %s: %v", goalID, err)
      // non-fatal: continue loop
  }

### In db/queries.go — add ComputeGrowthTrajectory function
func (db *DB) ComputeGrowthTrajectory(ctx context.Context, goalID uuid.UUID, date time.Time) error {
    _, err := db.Pool.Exec(ctx,
        "SELECT fn_compute_growth_trajectory($1, $2)",
        goalID, date.UTC().Truncate(24*time.Hour),
    )
    return err
}

## Verification
After implementation, confirm:
  grep "ComputeGrowthTrajectory" backend/internal/scheduler/scheduler.go  → 1 match
  grep "fn_compute_growth_trajectory" backend/internal/db/queries.go  → 1 match

## After
- Update CLAUDE.md section 4: SA-1 ✅
- Update docs/testing/test-plan.md: SA-1 ✅
- Commit: feat: SA-1 wire fn_compute_growth_trajectory in jobComputeDailyScore
- Scenarios now passing: TS-03 (after 2 scheduler runs), TS-08 (day 1 snapshot)
```

---

## Sprint 3.1 — Sesiunea 3: SA-3 (SRM L1 auto-trigger)

```
Read CLAUDE.md. Current version: 10.5.0. Branch: claude/sa3-srm-l1-XXXXX.
Task: wire SRM L1 auto-trigger inside jobCheckDailyProgress after stagnation detection.

## Context
jobDetectStagnation (23:58 UTC) populates stagnation_events correctly when
inactive_days >= 5. Problem: it does NOT write to srm_events with srm_level = 'L1'.
So GET /srm/status always returns "NONE" even after 5 missed days.

## Files to read (max 3)
1. backend/internal/scheduler/scheduler.go — find jobDetectStagnation function
2. backend/internal/engine/srm.go — find CheckAndRecordRegressionEvent or L1 trigger logic
3. backend/internal/db/queries.go — find or check InsertSRMEvent signature

## Changes required

### In scheduler.go — inside jobDetectStagnation, after stagnation_events insert
After detecting inactive_days >= 5 for a goal, add:

  srmEvent := engine.SRMEvent{
      GoID:          goalID,
      SRMLevel:      "L1",
      TriggerReason: "stagnation_5days",
      TriggeredAt:   time.Now(),
  }
  if err := db.InsertSRMEvent(ctx, srmEvent); err != nil {
      log.Printf("[scheduler] SRM L1 insert failed for %s: %v", goalID, err)
  }

Only insert L1 if no existing active srm_event with level L1 or higher exists for this goal.
Add a guard: check db.GetActiveSRMLevel(ctx, goalID) — if already L1/L2/L3, skip.

## Verification
  grep "SRMLevel.*L1" backend/internal/scheduler/scheduler.go  → 1 match
  grep "stagnation_5days" backend/internal/scheduler/scheduler.go  → 1 match

## After
- Update CLAUDE.md section 4: SA-3 ✅
- Update docs/testing/test-plan.md: SA-3 ✅
- Commit: feat: SA-3 SRM L1 auto-trigger in jobDetectStagnation
- Scenarios now passing: TS-04
```

---

## Sprint 3.1 — Sesiunea 4: SA-4 + SA-5 (SRM L2 reduces intensity + frontend button)

```
Read CLAUDE.md. Current version: 10.5.0. Branch: claude/sa4-sa5-srm-l2-XXXXX.
Task: SA-4 (backend — ConfirmSRML2 creates context adjustment) + SA-5 (frontend — SRMWarning.tsx confirm button).

## Context
SA-4: ConfirmSRML2() in srm.go stamps confirmed_at correctly but does NOT call
CreateContextAdjustment(). So next-day task count is unchanged after L2 confirm.

SA-5: SRMWarning.tsx shows the L2 banner but has no confirm button.
POST /srm/confirm-l2/:goalId route exists — just not wired in the frontend.

## Files to read (max 3)
1. backend/internal/engine/srm.go — find ConfirmSRML2 function
2. backend/internal/db/queries.go — find CreateContextAdjustment signature
3. frontend/app/components/SRMWarning.tsx — find L2 conditional section

## Changes required

### SA-4: srm.go — ConfirmSRML2()
After stamping confirmed_at on srm_events, add:

  tomorrow := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
  adj := db.ContextAdjustment{
      GoID:      goalID,
      AdjType:   db.AdjEnergyLow,
      StartDate: tomorrow,
      EndDate:   tomorrow.AddDate(0, 0, 7), // 7-day window
      Source:    "srm_l2_confirm",
  }
  if err := dbConn.CreateContextAdjustment(ctx, adj); err != nil {
      return fmt.Errorf("ConfirmSRML2 context adjustment: %w", err)
  }

### SA-5: SRMWarning.tsx — inside L2 conditional block
Add below the L2 message text:

  {status.srm_level === 'L2' && !status.confirmed && (
    <button
      onClick={handleConfirmL2}
      className="mt-3 px-4 py-2 rounded-lg bg-amber-500 text-white text-sm font-medium hover:bg-amber-600 transition-colors"
    >
      Confirmare — Reduc intensitatea
    </button>
  )}

Add handleConfirmL2 function before return:
  const handleConfirmL2 = async () => {
    await fetch(`/api/v1/srm/confirm-l2/${goalId}`, { method: 'POST', headers: authHeaders() });
    refreshSRMStatus(); // call existing refresh function
  };

## Verification
  grep "AdjEnergyLow" backend/internal/engine/srm.go  → 1 match in ConfirmSRML2
  grep "confirm-l2" frontend/app/components/SRMWarning.tsx  → 1 match

## After
- Update CLAUDE.md section 4: SA-4 ✅, SA-5 ✅
- Update docs/testing/test-plan.md: SA-4 ✅, SA-5 ✅
- Commit: feat: SA-4 ConfirmSRML2 creates context adjustment + SA-5 SRMWarning L2 button
- Scenarios now passing: TS-05 (backend + frontend)
```

---

## Sprint 3.1 — Sesiunea 5: SA-2 + SA-6 (achievements + SRM fallback)

```
Read CLAUDE.md. Current version: 10.5.0. Branch: claude/sa2-sa6-XXXXX.
Task: SA-2 (fn_award_achievement_if_earned call) + SA-6 (ApplySRMFallback implementation).

## Context
SA-2: jobGenerateCeremonies calls GenerateCompletionCeremony() but never calls
fn_award_achievement_if_earned(user_id, sprint_id). So GET /achievements always returns [].

SA-6: jobCheckSRMTimeouts has a TODO comment:
  // TODO: engine.ApplySRMFallback(ctx, goalID, fallback)
This means goals with unconfirmed L3 are never resolved.

## Files to read (max 3)
1. backend/internal/scheduler/scheduler.go — find jobGenerateCeremonies and jobCheckSRMTimeouts
2. backend/internal/db/queries.go — find AwardAchievement or fn_award_achievement_if_earned call
3. backend/internal/engine/srm.go — find ApplySRMFallback signature (may not exist yet)

## Changes required

### SA-2: scheduler.go — jobGenerateCeremonies
After successful GenerateCompletionCeremony(), add:

  if err := db.AwardAchievementIfEarned(ctx, userID, sprintID); err != nil {
      log.Printf("[scheduler] achievement award failed for sprint %s: %v", sprintID, err)
      // non-fatal
  }

### SA-2: db/queries.go — add AwardAchievementIfEarned
func (db *DB) AwardAchievementIfEarned(ctx context.Context, userID, sprintID uuid.UUID) error {
    _, err := db.Pool.Exec(ctx,
        "SELECT fn_award_achievement_if_earned($1, $2)",
        userID, sprintID,
    )
    return err
}

### SA-6: scheduler.go — jobCheckSRMTimeouts
Replace the TODO line with:

  fallbackLevel := engine.ComputeSRMFallback(currentLevel, hoursUnconfirmed)
  srmEvent := engine.SRMEvent{
      GoID:          goalID,
      SRMLevel:      fallbackLevel,
      TriggerReason: "srm_timeout_fallback",
      TriggeredAt:   time.Now(),
  }
  if err := db.InsertSRMEvent(ctx, srmEvent); err != nil {
      log.Printf("[scheduler] SRM fallback insert failed: %v", err)
  }

If engine.ComputeSRMFallback does not exist, create it in srm.go:
  func ComputeSRMFallback(current string, hoursUnconfirmed float64) string {
      if current == "L3" && hoursUnconfirmed > 72 { return "L1" }
      return current
  }

## Verification
  grep "AwardAchievementIfEarned" backend/internal/scheduler/scheduler.go  → 1 match
  grep "TODO.*ApplySRMFallback" backend/internal/scheduler/scheduler.go  → must return EMPTY

## After
- Update CLAUDE.md section 4: SA-2 ✅, SA-6 ✅
- Update docs/testing/test-plan.md: SA-2 ✅, SA-6 ✅
- Commit: feat: SA-2 fn_award_achievement_if_earned + SA-6 ApplySRMFallback implementation
- Scenarios now passing: TS-06, TS-07
- Sprint 3.1 post-fix checklist: all SA items ✅ → run full validation checklist in regression.md
```

---

## Sprint 3.1 — Sesiunea 6: CI/CD Tests (GitHub Actions)

```
Read CLAUDE.md. Current version: 10.5.0. Branch: claude/cicd-tests-XXXXX.
Task: implement 3-layer test suite in GitHub Actions — Go unit, integration, E2E (Playwright).
Priority order: unit → integration → E2E. Implement unit + integration in this session.
E2E (Playwright) = separate session when unit + integration are green.

## Files to read (max 3)
1. .github/workflows/ — list all existing workflow files
2. backend/go.mod — check Go version and test dependencies
3. Makefile or any existing test scripts at root

## What to implement

### Layer 1: Go Unit Tests (.github/workflows/test-unit.yml)
Triggers: push to any branch, PR to main.
Steps:
  - actions/checkout@v4
  - actions/setup-go@v5 with go-version from go.mod
  - go test ./internal/engine/... -v -count=1
  - go test ./internal/db/... -v -count=1 (skip DB tests with build tag: //go:build !integration)
  - go vet ./...
  - golangci-lint run (if .golangci.yml exists; else skip)

### Layer 2: Integration Tests (.github/workflows/test-integration.yml)
Triggers: push to main, PR to main.
Services: postgres:16-alpine, redis:7-alpine (GitHub Actions services block).
Env: TEST_DB_URL from service, TEST_REDIS_URL from service.
Steps:
  - actions/checkout@v4
  - actions/setup-go@v5
  - Wait for postgres: pg_isready loop (max 30s)
  - Run migrations: go run ./cmd/migrate/main.go (or equivalent)
  - go test ./... -tags=integration -v -count=1
  - Report: upload test results as artifact

### Create test files if missing
backend/internal/engine/engine_test.go — test opaque API rule:
  func TestEngineNeverExposesDrift(t *testing.T) {
      // Call engine with mock goal data
      // Assert response struct has no drift, chaos_index, weights fields
  }

backend/internal/engine/srm_test.go — test ComputeSRMFallback:
  func TestComputeSRMFallbackL3After72h(t *testing.T) {
      result := ComputeSRMFallback("L3", 73)
      if result != "L1" { t.Errorf("expected L1, got %s", result) }
  }

## Constraints
- Integration tests must use build tag: //go:build integration
- Unit tests must NOT require DB connection
- Workflows must NOT use real API keys — use mock env vars
- E2E (Playwright) = NOT in this session; add placeholder job with skip

## After
- Update CLAUDE.md: add CI/CD test status section
- Update README.md: add CI badges for both workflows
- Commit: feat: CI/CD unit + integration test workflows (GitHub Actions)
- Note in commit: E2E Playwright planned Sprint 4
```

---

## Regula de analiză prompturi (pentru CLAUDE.md)

Adaugă secțiunea următoare în CLAUDE.md, după secțiunea 11 (User Workflow & Testing):

```markdown
---

## 12. Prompt Optimization Rule (MANDATORY)

> Before starting any task from PROMPTS.md, Claude MUST:

### Step 1 — Read only what's needed
Based on the prompt's "Files to read (max 3)" list, read EXACTLY those files.
Do NOT explore additional files unless a file is missing (then ask the owner).

### Step 2 — Validate before coding
Answer these questions from the files read:
- Does the function/line referenced in the prompt actually exist?
- If a function is missing, create it as specified — do not skip.
- If a line number is wrong, find the correct location and proceed.

### Step 3 — One change at a time
Make changes in the exact order listed in the prompt.
After each change, grep to verify (as specified in the prompt's Verification block).
Do NOT proceed to the next change if grep shows unexpected results — report to owner.

### Step 4 — Close the session correctly
- State which TS-xx scenarios are now satisfied
- Update CLAUDE.md section 4 status
- Update docs/testing/test-plan.md gap status
- Commit with exact message format from prompt
- Do NOT add extra files to the commit unless the prompt specifies them

### Why this rule exists
Context windows are expensive. A session that reads 10 files instead of 3
uses 3× the tokens for the same output. Every file read beyond the prompt's
list must be justified by a missing reference — not curiosity or "just in case".
```

---

*Last updated: 2026-04-02 — v10.5.0 — Sprint 3.1 SA fix prompts + CI/CD + Prompt Optimization Rule*
*Sessions: SA-7+CE-1 → SA-1 → SA-3 → SA-4+SA-5 → SA-2+SA-6 → CI/CD*
*Mark each with ✅ after session is committed and pushed.*
