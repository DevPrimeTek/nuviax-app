# PLAN.md — NuviaX: Demo Implementation Plan

> Versiune: 10.5.0 | Actualizat: 2026-04-02
> Obiectiv: flux complet register → goal → sprint → today → progress

---

## Cum folosești

1. Citești sesiunea curentă (prima cu ⏳)
2. Copiezi promptul exact într-o sesiune **NOUĂ** Claude Code
3. Claude Code implementează → commit → push → merge → deploy automat
4. Marchezi ✅ și treci la următoarea

**Un prompt = o sesiune nouă = un commit = un deploy pe VPS.**

---

## Progres

| # | Task | Status | Rezultat |
|---|------|--------|----------|
| 1 | SA-7 + CE-1 | ✅ DONE | Cron fix + trajectory null rezolvat |
| 2 | SA-1 | ✅ DONE | Grafic progres populat |
| 3 | SA-3 | ✅ DONE | SRM L1 auto-trigger activ |
| 4 | SA-4 + SA-5 | ✅ DONE | Buton L2 + reducere tasks |
| 5 | SA-2 + SA-6 | ✅ DONE | Achievements + SRM fallback |
| 6 | CI/CD tests | ⏳ PENDING | Tests la fiecare push (după demo) |

---

---

## SESIUNEA 1 — SA-7 + CE-1

> **Blockers rezolvate:** cron invalid care face L2 să nu se evalueze niciodată + trajectory null în ziua 1
> **Timp estimat:** 20 minute | **Model:** Sonnet

```
Read CLAUDE.md. Current version: 10.5.0. Branch: claude/sa7-ce1-XXXXX.
Task: two single-line fixes — SA-7 (invalid cron) + CE-1 (wrong table name).

## Files to read (max 2)
1. backend/internal/scheduler/scheduler.go — search "*/90" near jobRecalibrateRelevance
2. backend/internal/engine/level5_growth.go lines 80-95 — search "FROM goals"

## Fix 1 — SA-7 (scheduler.go)
Find:    "0 2 */90 * *"
Replace: "0 2 * * 0"
Only this line changes in this file.

## Fix 2 — CE-1 (level5_growth.go ~line 85)
Find:    FROM goals
Replace: FROM global_objectives
Only this line changes in this file.

## Verify (mandatory before commit)
grep "*/90" backend/internal/scheduler/scheduler.go         → EMPTY
grep "FROM goals" backend/internal/engine/level5_growth.go  → EMPTY

## Commit & push
git add backend/internal/scheduler/scheduler.go backend/internal/engine/level5_growth.go
git commit -m "fix: SA-7 cron weekly + CE-1 trajectory table name"
git push origin claude/sa7-ce1-XXXXX

## Update docs (same session, separate commit)
In CLAUDE.md section 4: SA-7 ✅  CE-1 ✅
In PLAN.md table: change session 1 from ⏳ PENDING to ✅ DONE
git add CLAUDE.md PLAN.md
git commit -m "docs: SA-7 CE-1 marked complete"
git push origin claude/sa7-ce1-XXXXX

## Then: open PR on GitHub → merge to main
Merge triggers GitHub Actions → auto build + deploy to VPS (~3 min).

## Session close output (print exactly)
═══════════════════════════════════════════════════
✅ SESSION 1 DONE — SA-7 + CE-1
cron: "0 2 * * 0" | table: FROM global_objectives
Deploy: GitHub Actions triggered on merge to main
Unblocks: TS-04 (indirect), TS-08
NEXT → open PLAN.md, copy Session 2 prompt
═══════════════════════════════════════════════════
```

**Ce se întâmplă după merge:**
- GitHub Actions rulează: test → build Docker → push DockerHub → SSH deploy VPS
- `jobRecalibrateRelevance` rulează în fiecare duminică la 02:00 UTC
- `GET /goals/:id/visualize` returnează 1 entry în ziua 1 (nu mai e null)

---

---

## SESIUNEA 2 — SA-1

> **Blocker rezolvat:** graficul de progres e gol — `growth_trajectories` nu se populează niciodată
> **Timp estimat:** 25 minute | **Model:** Sonnet

```
Read CLAUDE.md. Current version: 10.5.0. Branch: claude/sa1-trajectories-XXXXX.
Task: call fn_compute_growth_trajectory inside jobComputeDailyScore.

## Context
PostgreSQL function fn_compute_growth_trajectory(goal_id, date) exists.
growth_trajectories table exists. Problem: the function is never called from Go.
jobComputeDailyScore (23:50 UTC) calls db.UpsertGoalScore() then stops.
Result: GET /goals/:id/visualize returns trajectory: null for every goal.

## Files to read (max 2)
1. backend/internal/scheduler/scheduler.go — find jobComputeDailyScore, read full body
2. backend/internal/db/queries.go — search "ComputeGrowthTrajectory" to see if it exists

## Add to scheduler.go
In jobComputeDailyScore, after db.UpsertGoalScore(ctx, goalID, score) succeeds:

  if err := db.ComputeGrowthTrajectory(ctx, goalID, time.Now()); err != nil {
      log.Printf("[scheduler] trajectory failed for %s: %v", goalID, err)
  }

## Add to db/queries.go only if function missing
  func (db *DB) ComputeGrowthTrajectory(ctx context.Context, goalID uuid.UUID, date time.Time) error {
      _, err := db.Pool.Exec(ctx,
          "SELECT fn_compute_growth_trajectory($1, $2)",
          goalID, date.UTC().Truncate(24*time.Hour),
      )
      return err
  }

## Verify
grep "ComputeGrowthTrajectory" backend/internal/scheduler/scheduler.go → 1 match
grep "fn_compute_growth_trajectory" backend/internal/db/queries.go      → 1 match

## Commit & push
git add backend/internal/scheduler/scheduler.go backend/internal/db/queries.go
git commit -m "feat: SA-1 wire fn_compute_growth_trajectory in jobComputeDailyScore"
git push origin claude/sa1-trajectories-XXXXX

## Update docs
In CLAUDE.md section 4: SA-1 ✅
In PLAN.md table: session 2 → ✅ DONE
git add CLAUDE.md PLAN.md
git commit -m "docs: SA-1 marked complete"
git push origin claude/sa1-trajectories-XXXXX

## Then: open PR → merge to main → GitHub Actions deploys automatically

## Session close output
═══════════════════════════════════════════════════
✅ SESSION 2 DONE — SA-1 trajectories
jobComputeDailyScore now populates growth_trajectories
Deploy: GitHub Actions triggered on merge
Unblocks: TS-03, TS-08
NEXT → open PLAN.md, copy Session 3 prompt
═══════════════════════════════════════════════════
```

**Ce se întâmplă după merge:**
- La fiecare rulare scheduler (23:50 UTC), `growth_trajectories` primește date
- `GET /goals/:id/visualize` returnează `trajectory` cu entries reale
- Graficul de progres din frontend afișează traiectoria real vs așteptată

---

---

## SESIUNEA 3 — SA-3

> **Blocker rezolvat:** SRM nu se activează — `GET /srm/status` returnează NONE chiar și după 5 zile ratate
> **Timp estimat:** 25 minute | **Model:** Sonnet

```
Read CLAUDE.md. Current version: 10.5.0. Branch: claude/sa3-srm-l1-XXXXX.
Task: insert srm_events L1 inside jobDetectStagnation after stagnation detected.

## Context
jobDetectStagnation (23:58 UTC) writes stagnation_events when inactive_days >= 5.
Problem: it never writes to srm_events. So GET /srm/status/:goalId always
returns srm_level: "NONE" no matter how many days the user misses.

## Files to read (max 2)
1. backend/internal/scheduler/scheduler.go — find jobDetectStagnation, read full body
2. backend/internal/db/queries.go — search "InsertSRMEvent" for exact signature

## Add to scheduler.go
Inside jobDetectStagnation, right after the stagnation_events insert succeeds:

  existingLevel, _ := db.GetActiveSRMLevel(ctx, goalID)
  if existingLevel == "" {
      if err := db.InsertSRMEvent(ctx, goalID, "L1", "stagnation_5days"); err != nil {
          log.Printf("[scheduler] SRM L1 failed for %s: %v", goalID, err)
      }
  }

Note: adapt InsertSRMEvent call to match the real signature found in step 2.

## Add to db/queries.go if GetActiveSRMLevel missing
  func (db *DB) GetActiveSRMLevel(ctx context.Context, goalID uuid.UUID) (string, error) {
      var level string
      err := db.Pool.QueryRow(ctx,
          `SELECT srm_level FROM srm_events
           WHERE go_id = $1 AND confirmed_at IS NULL
           ORDER BY triggered_at DESC LIMIT 1`,
          goalID,
      ).Scan(&level)
      if errors.Is(err, pgx.ErrNoRows) { return "", nil }
      return level, err
  }

## Verify
grep "stagnation_5days" backend/internal/scheduler/scheduler.go → 1 match
grep "GetActiveSRMLevel" backend/internal/scheduler/scheduler.go → 1 match

## Commit & push
git add backend/internal/scheduler/scheduler.go backend/internal/db/queries.go
git commit -m "feat: SA-3 SRM L1 auto-trigger after 5 consecutive inactive days"
git push origin claude/sa3-srm-l1-XXXXX

## Update docs
In CLAUDE.md section 4: SA-3 ✅
In PLAN.md table: session 3 → ✅ DONE
git add CLAUDE.md PLAN.md
git commit -m "docs: SA-3 marked complete"
git push origin claude/sa3-srm-l1-XXXXX

## Then: open PR → merge to main → GitHub Actions deploys automatically

## Session close output
═══════════════════════════════════════════════════
✅ SESSION 3 DONE — SA-3 SRM L1
jobDetectStagnation now writes srm_events L1 after 5 inactive days
Deploy: GitHub Actions triggered on merge
Unblocks: TS-04
NEXT → open PLAN.md, copy Session 4 prompt
═══════════════════════════════════════════════════
```

**Ce se întâmplă după merge:**
- După 5 zile fără tasks, `GET /srm/status/:goalId` returnează `srm_level: "L1"`
- Banner-ul SRM apare în dashboard cu mesaj de ajustare
- Guard activ: nu inserează L1 dacă există deja L1/L2/L3 activ

---

---

## SESIUNEA 4 — SA-4 + SA-5

> **Blockers rezolvate:** confirmarea L2 nu reduce tasks + butonul de confirmare lipsește din UI
> **Timp estimat:** 30 minute | **Model:** Sonnet

```
Read CLAUDE.md. Current version: 10.5.0. Branch: claude/sa4-sa5-XXXXX.
Task: SA-4 (ConfirmSRML2 creates context adjustment) + SA-5 (L2 confirm button in UI).

## Context
SA-4: ConfirmSRML2() in srm.go stamps confirmed_at correctly but never calls
      CreateContextAdjustment(). Next-day tasks are unchanged after L2 confirm.
SA-5: SRMWarning.tsx shows the L2 banner but has no confirm button.
      POST /srm/confirm-l2/:goalId route exists — just not wired in frontend.

## Files to read (max 3)
1. backend/internal/engine/srm.go — find ConfirmSRML2, read from confirmed_at stamp to end
2. backend/internal/db/queries.go — search "CreateContextAdjustment" + "AdjEnergyLow"
3. frontend/app/components/SRMWarning.tsx — find the L2 conditional section

## SA-4: add in srm.go inside ConfirmSRML2(), after confirmed_at is stamped
  tomorrow := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
  adj := db.ContextAdjustment{
      GoID:      goalID,
      AdjType:   db.AdjEnergyLow,
      StartDate: tomorrow,
      EndDate:   tomorrow.AddDate(0, 0, 7),
      Source:    "srm_l2_confirm",
  }
  if err := dbConn.CreateContextAdjustment(ctx, adj); err != nil {
      return fmt.Errorf("ConfirmSRML2 adj: %w", err)
  }

## SA-5: add in SRMWarning.tsx inside L2 conditional block, after message text
  {status.srm_level === 'L2' && !status.confirmed && (
    <button
      onClick={handleConfirmL2}
      className="mt-3 px-4 py-2 rounded-lg bg-amber-500 text-white text-sm font-medium hover:bg-amber-600 transition-colors"
    >
      Confirmare — Reduc intensitatea
    </button>
  )}

Add handleConfirmL2 before the return statement:
  const handleConfirmL2 = async () => {
    await fetch(`/api/v1/srm/confirm-l2/${goalId}`, {
      method: 'POST',
      headers: authHeaders(),
    });
    refreshSRMStatus();
  };

## Verify
grep "AdjEnergyLow" backend/internal/engine/srm.go       → 1 match inside ConfirmSRML2
grep "confirm-l2" frontend/app/components/SRMWarning.tsx  → 1 match

## Commit & push
git add backend/internal/engine/srm.go frontend/app/components/SRMWarning.tsx
git commit -m "feat: SA-4 ConfirmSRML2 reduces intensity + SA-5 SRMWarning L2 confirm button"
git push origin claude/sa4-sa5-XXXXX

## Update docs
In CLAUDE.md section 4: SA-4 ✅  SA-5 ✅
In PLAN.md table: session 4 → ✅ DONE
git add CLAUDE.md PLAN.md
git commit -m "docs: SA-4 SA-5 marked complete"
git push origin claude/sa4-sa5-XXXXX

## Then: open PR → merge to main → GitHub Actions deploys automatically

## Session close output
═══════════════════════════════════════════════════
✅ SESSION 4 DONE — SA-4 + SA-5
L2 confirm creates context adjustment + button visible in UI
Deploy: GitHub Actions triggered on merge
Unblocks: TS-05 (backend + frontend)
NEXT → open PLAN.md, copy Session 5 prompt
═══════════════════════════════════════════════════
```

**Ce se întâmplă după merge:**
- Butonul "Confirmare — Reduc intensitatea" apare în banner-ul L2
- A doua zi după confirmare, `GenerateDailyTasks` produce mai puține tasks
- Fluxul SRM L2 complet funcțional end-to-end

---

---

## SESIUNEA 5 — SA-2 + SA-6

> **Blockers rezolvate:** achievements returnează [] + goaluri cu L3 blocate indefinit
> **Timp estimat:** 30 minute | **Model:** Sonnet

```
Read CLAUDE.md. Current version: 10.5.0. Branch: claude/sa2-sa6-XXXXX.
Task: SA-2 (award achievements on sprint close) + SA-6 (SRM L3 timeout fallback).

## Context
SA-2: jobGenerateCeremonies calls GenerateCompletionCeremony() correctly but never
      calls fn_award_achievement_if_earned(). GET /achievements always returns [].
SA-6: jobCheckSRMTimeouts has "// TODO: engine.ApplySRMFallback(...)" never replaced.
      Goals with unconfirmed L3 stay blocked indefinitely.

## Files to read (max 3)
1. backend/internal/scheduler/scheduler.go — find jobGenerateCeremonies AND jobCheckSRMTimeouts
2. backend/internal/db/queries.go — search "AwardAchievement" (likely missing)
3. backend/internal/engine/srm.go — search "ComputeSRMFallback" (likely missing)

## SA-2: add in scheduler.go inside jobGenerateCeremonies, after ceremony generated successfully
  if err := db.AwardAchievementIfEarned(ctx, userID, sprintID); err != nil {
      log.Printf("[scheduler] achievement failed for sprint %s: %v", sprintID, err)
  }

Add to db/queries.go if missing:
  func (db *DB) AwardAchievementIfEarned(ctx context.Context, userID, sprintID uuid.UUID) error {
      _, err := db.Pool.Exec(ctx,
          "SELECT fn_award_achievement_if_earned($1, $2)", userID, sprintID)
      return err
  }

## SA-6: in scheduler.go inside jobCheckSRMTimeouts, replace the TODO line with
  fallback := engine.ComputeSRMFallback(currentLevel, hoursUnconfirmed)
  if err := db.InsertSRMEvent(ctx, goalID, fallback, "srm_timeout_fallback"); err != nil {
      log.Printf("[scheduler] SRM fallback failed: %v", err)
  }

Add to srm.go if ComputeSRMFallback missing:
  func ComputeSRMFallback(current string, hours float64) string {
      if current == "L3" && hours > 72 { return "L1" }
      return current
  }

## Verify
grep "AwardAchievementIfEarned" backend/internal/scheduler/scheduler.go → 1 match
grep "TODO.*ApplySRMFallback" backend/internal/scheduler/scheduler.go    → EMPTY

## Commit & push
git add backend/internal/scheduler/scheduler.go backend/internal/db/queries.go backend/internal/engine/srm.go
git commit -m "feat: SA-2 award achievements on sprint close + SA-6 SRM L3 timeout fallback"
git push origin claude/sa2-sa6-XXXXX

## Update docs
In CLAUDE.md section 4: SA-2 ✅  SA-6 ✅
In PLAN.md table: session 5 → ✅ DONE
git add CLAUDE.md PLAN.md
git commit -m "docs: SA-2 SA-6 complete — Sprint 3.1 all SA fixes done"
git push origin claude/sa2-sa6-XXXXX

## Run regression checklist
Open docs/testing/scenarios/regression.md → run every item in Post-Fix Validation Checklist.

## Then: open PR → merge to main → GitHub Actions deploys automatically

## Session close output
═══════════════════════════════════════════════════
✅ SESSION 5 DONE — SA-2 + SA-6
Achievements awarded on sprint close + SRM L3 fallback active
Deploy: GitHub Actions triggered on merge
Unblocks: TS-06, TS-07
🎉 DEMO READY — all 5 sessions complete
Run final checklist from PLAN.md before demo
NEXT → Session 6 (CI/CD) after demo confirmed
═══════════════════════════════════════════════════
```

**Ce se întâmplă după merge:**
- La închiderea unui sprint → `GET /achievements` returnează badges câștigate
- Goaluri cu L3 neconfirmat >72h → trec automat la L1 (nu mai sunt blocate)
- **DEMO READY** — toate SA fixes complete

---

---

## SESIUNEA 6 — CI/CD Tests *(după demo)*

> **Notă:** Această sesiune nu blochează demo-ul. O faci după ce demo-ul e confirmat.
> **Timp estimat:** 40 minute | **Model:** Sonnet

```
Read CLAUDE.md. Current version: 10.5.0. Branch: claude/cicd-tests-XXXXX.
Task: GitHub Actions test suite — unit tests + integration tests.
E2E Playwright is Sprint 4, not this session.

## Files to read (max 3)
1. .github/workflows/ — list existing files
2. backend/go.mod — Go version for actions/setup-go
3. backend/internal/engine/ — list files to know which test files exist

## Create .github/workflows/test-unit.yml
name: Unit Tests
on: [push, pull_request]
jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: backend/go.mod
      - name: Vet
        working-directory: backend
        run: go vet ./...
      - name: Unit tests
        working-directory: backend
        run: go test ./internal/engine/... -v -count=1

## Create .github/workflows/test-integration.yml
name: Integration Tests
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
jobs:
  integration:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: nuviax_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-retries 5
      redis:
        image: redis:7-alpine
        options: --health-cmd "redis-cli ping"
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: backend/go.mod
      - name: Integration tests
        working-directory: backend
        run: go test ./... -tags=integration -v -count=1
        env:
          TEST_DB_URL: postgres://postgres:postgres@localhost:5432/nuviax_test
          TEST_REDIS_URL: redis://localhost:6379

## Create backend/internal/engine/srm_test.go if file missing
//go:build !integration

package engine

import "testing"

func TestComputeSRMFallback(t *testing.T) {
    cases := []struct{ c string; h float64; want string }{
        {"L3", 73, "L1"},
        {"L3", 24, "L3"},
        {"L2", 50, "L2"},
    }
    for _, tc := range cases {
        if got := ComputeSRMFallback(tc.c, tc.h); got != tc.want {
            t.Errorf("got %q want %q", got, tc.want)
        }
    }
}

## Commit & push
git add .github/workflows/test-unit.yml .github/workflows/test-integration.yml
git add backend/internal/engine/srm_test.go
git commit -m "feat: CI/CD unit + integration tests (E2E Playwright is Sprint 4)"
git push origin claude/cicd-tests-XXXXX

## Update docs
In README.md: add CI badge for test-unit and test-integration workflows
In PLAN.md table: session 6 → ✅ DONE
git add README.md PLAN.md
git commit -m "docs: CI/CD active — Sprint 3.1 fully complete"
git push origin claude/cicd-tests-XXXXX

## Then: open PR → merge to main → tests run on every future push

## Session close output
═══════════════════════════════════════════════════
✅ SESSION 6 DONE — CI/CD Tests
Unit + integration tests active in GitHub Actions
Every push to main: tests → build → deploy (if green)
Sprint 3.1 COMPLETE
═══════════════════════════════════════════════════
```

**Ce se întâmplă după merge:**
- Fiecare PR sau push declanșează testele automat
- Deploy-ul pe VPS se face doar dacă testele trec
- Regresiile sunt prinse înainte să ajungă în producție

---

## Checklist final demo

```bash
# Rulează după sesiunea 5 e pe VPS

# 1. API healthy
curl https://api.nuviax.app/health
# → {"status":"ok","db":true,"redis":true}

# 2. Trajectory nu mai e null
curl -H "Authorization: Bearer TOKEN" \
  https://api.nuviax.app/api/v1/goals/GOAL_ID/visualize
# → trajectory: [...] nu null

# 3. SRM status
curl -H "Authorization: Bearer TOKEN" \
  https://api.nuviax.app/api/v1/srm/status/GOAL_ID
# → srm_level: "NONE" / "L1" / "L2" / "L3"

# 4. Zero câmpuri interne în răspunsuri API
# Verifică manual că drift, chaos_index, weights NU apar
```

Checklist complet: `docs/testing/scenarios/regression.md`

---

*Actualizat de Claude Code la finalul fiecărei sesiuni*
