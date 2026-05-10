# PROMPTS-FIX-SESSIONS — NuviaX MVP Fix Phase F8

> **Versiune:** 1.0.0
> **Data:** 2026-05-10
> **Owner:** Solution Architect
> **Convenție:** Prompturile sunt în engleză (pentru consistență cu Anthropic API). Rapoartele de output sunt în română (per cerință utilizator).

---

## Cum se folosește

1. PM verifică ce sub-fază urmează (`02-BACKLOG.md` → `00-PLAN-MASTER.md`)
2. PM verifică gate-ul precedent verde (`03-TESTING-STRATEGY.md`)
3. Specialistul desemnat (`01-TEAM-ROSTER.md`) deschide o sesiune Claude Code nouă pe **branch nou** (convenție `claude/fix-<phase>-<slug>`)
4. Copy-paste prompt-ul corespunzător sub-fazei
5. La final: PR draft → QA verifică gate → PM mută BACKLOG → Architect aprobă → merge

---

## SESSION FIX-01 — Schema Reconciliation (F8.1)

- **Owner:** DBA
- **Reviewer:** Architect
- **Model:** Sonnet 4.6
- **Branch:** `claude/fix-01-schema-reconciliation`
- **Backlog:** DEV-02, DEV-03, DEV-04, DEV-17
- **Estimare:** 60 min
- **Pre-condiții:** AUDIT-01 mergeable; access la DB de pe server pentru schema export (sau echivalent)

### Prompt (paste în sesiune Claude Code):

```
You are a Senior DBA working on NuviaX MVP fix phase F8.1 (Schema Reconciliation).

Read in this order:
1. CLAUDE.md
2. docs/audit/AUDIT-01-deviation-report.md (focus on DEV-02, DEV-03, DEV-04, DEV-17)
3. docs/fix-plan/02-BACKLOG.md (sections DEV-02 to DEV-04, DEV-17)
4. backend/migrations/001_schema.sql
5. backend/internal/db/db.go
6. backend/internal/db/queries.go
7. backend/internal/api/handlers/srm.go (for table refs)
8. backend/internal/scheduler/scheduler.go (for table refs)

Goal: produce a single new migration file backend/migrations/002_runtime_baseline.sql
that defines all tables currently referenced by Go code but missing from local migrations,
and aligns the users/sessions tables with the Go model.

Tables to define (idempotent, IF NOT EXISTS or DO $$ ... EXCEPTION WHEN duplicate_object):
- srm_events (id, go_id, user_id, level srm_level_enum, event_type text, trigger_reason text,
  confirmed_at timestamptz, confirmed_by uuid, created_at timestamptz default now())
  with CREATE TYPE srm_level_enum AS ENUM ('L1', 'L2', 'L3')
- context_adjustments (id, go_id, user_id, type context_adj_type_enum, valid_from date, valid_until date, created_at)
  with CREATE TYPE context_adj_type_enum AS ENUM ('ENERGY_LOW', 'ENERGY_HIGH', 'PAUSE')
- stagnation_events (id, go_id, user_id, days_inactive int, detected_at timestamptz)
- ceremonies (id, sprint_id UNIQUE, go_id, user_id, tier ceremony_tier_enum, viewed_at timestamptz, created_at)
  with CREATE TYPE ceremony_tier_enum AS ENUM ('BRONZE','SILVER','GOLD','PLATINUM')
- sprint_results (id, sprint_id UNIQUE, go_id, user_id, sprint_score numeric(4,3), grade char(2),
  ceremony_tier ceremony_tier_enum, created_at)
- go_metrics (id, go_id UNIQUE, user_id, gori_score numeric(5,4), ali_current numeric(5,4),
  ali_projected numeric(5,4), updated_at)
- achievement_badges (id, user_id, badge_type text, go_id, sprint_id, awarded_at timestamptz)
  UNIQUE (user_id, badge_type, sprint_id)
- growth_trajectories (id, go_id, snapshot_date date, actual_pct numeric(5,4),
  expected_pct numeric(5,4), delta numeric(6,4), trend trajectory_trend_enum, created_at)
  with CREATE TYPE trajectory_trend_enum AS ENUM ('ON_TRACK','AHEAD','BEHIND','CRITICAL')
  UNIQUE (go_id, snapshot_date)
- evolution_sprints (sprint_id PRIMARY KEY, evolution_score numeric(5,4),
  delta_performance numeric(6,4), consistency_weight numeric(5,4), detected_at timestamptz)

Also align users table — DEV-03:
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_encrypted text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS salt text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS full_name text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locale text DEFAULT 'ro';
ALTER TABLE users ADD COLUMN IF NOT EXISTS theme text DEFAULT 'dark';
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_secret text;
ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_enabled boolean DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT TRUE;

DEV-04 — sessions vs user_sessions: rename in migration with
ALTER TABLE IF EXISTS sessions RENAME TO user_sessions;
(idempotent — check current state first via DO $$ block)

Add to backend/internal/db/db.go::ensureGoalsTables the new statements OR replace
the function to load 002_runtime_baseline.sql via embed. Architect prefers embed for clarity.

Add a verification script: backend/scripts/schema-check.sh
- Spawns a temp Postgres (assume docker compose up testdb)
- Applies migrations in order
- Greps Go code for table refs (regex: FROM\s+(\w+)\b, INTO\s+(\w+)\b, UPDATE\s+(\w+)\b)
- Confirms each table exists in DB

Update CLAUDE.md (section 1) and ROADMAP.md (F8.1 → in_progress, then ✅ on completion).

Update docs/fix-plan/02-BACKLOG.md: change status of DEV-02, DEV-03, DEV-04, DEV-17 to IN_REVIEW.

Run: psql against a fresh test DB to verify migrations apply cleanly. Show output.

Write a Romanian summary at the end:
- Tabele adăugate
- Coloane adăugate la users
- Schema diff vs server (dacă disponibil)
- Probleme întâlnite
- Status backlog DEV-02..04, 17

Commit message:
fix(F8.1): reconcile DB schema with code references (DEV-02, DEV-03, DEV-04, DEV-17)

- Add migration 002_runtime_baseline.sql with 9 missing tables + 3 enums
- Align users table with Go model (email_encrypted, salt, mfa_*, etc.)
- Rename sessions → user_sessions (idempotent)
- Add schema-check.sh verification script
- Update CLAUDE.md, ROADMAP.md, BACKLOG.md

Do NOT modify any handlers, scheduler, or engine code in this session.
```

### Acceptance criteria (Gate F8.1):
- [ ] `psql` aplică migrările pe DB nou fără erori
- [ ] Toate cele 9 tabele lipsă există după migrare
- [ ] `schema-check.sh` exit 0
- [ ] BACKLOG: DEV-02..04, 17 → IN_REVIEW

---

## SESSION FIX-02 — Engine Restructure (F8.2)

- **Owner:** Backend Senior + Architect
- **Reviewer:** PM
- **Model:** Opus 4.7 (decizii API design) → Sonnet 4.6 (implementare)
- **Branch:** `claude/fix-02-engine-restructure`
- **Backlog:** DEV-11
- **Estimare:** 90 min
- **Pre-condiții:** F8.1 ✅

### Prompt:

```
You are a Senior Backend Engineer working on NuviaX MVP fix phase F8.2 (Engine Restructure).

Read in this order:
1. CLAUDE.md
2. docs/audit/AUDIT-01-deviation-report.md (focus on DEV-11)
3. docs/fix-plan/02-BACKLOG.md (DEV-11 + dependent items DEV-05, DEV-06, DEV-08, DEV-09, DEV-10, DEV-12, DEV-15)
4. FORMULAS_QUICK_REFERENCE.md
5. docs/user-workflow.md (sections 4, 5, 6 — SRM/Achievement/Visualization)
6. backend/internal/engine/ (all files)

Goal: Add the engine functions referenced throughout user-workflow.md but currently missing.
Do NOT remove existing functions — preserve and extend.

New files to create:
- backend/internal/engine/visualization.go
- backend/internal/engine/regulatory.go
- backend/internal/engine/evolution.go
- backend/internal/engine/clock.go (time injection helper for tests)

Functions to implement (signatures + behavior):

// visualization.go
func GenerateProgressVisualization(ctx context.Context, db *pgxpool.Pool, goalID uuid.UUID) (*ProgressData, error)
  - Query growth_trajectories for goal_id ORDER BY snapshot_date ASC
  - If empty, compute live snapshot:
    - SELECT start_date, end_date FROM global_objectives WHERE id=$1
    - elapsed = now - start_date; total = end_date - start_date
    - return single ProgressData entry { ActualPct: 0, ExpectedPct: elapsed/total clamped 0..1, Delta: -ExpectedPct, Trend: "ON_TRACK" }
  - Reads frozen_expected_pct if sprint frozen (see FreezeExpectedTrajectory)

type ProgressData struct {
  Date time.Time
  ActualPct, ExpectedPct, Delta float64
  Trend string // ON_TRACK, AHEAD, BEHIND, CRITICAL
}

func ComputeTrend(actual, expected float64) string
  delta := actual - expected
  switch {
  case delta >= 0.05: return "AHEAD"
  case delta >= -0.10: return "ON_TRACK"
  case delta >= -0.20: return "BEHIND"
  default: return "CRITICAL"
  }

// regulatory.go
func FreezeExpectedTrajectory(ctx context.Context, db *pgxpool.Pool, sprintID uuid.UUID) error
  - UPDATE sprints SET expected_pct_frozen=TRUE, frozen_expected_pct=<elapsed_ratio>
  - elapsed_ratio = days_since_sprint_start / 30.0 clamped

func UnfreezeExpectedTrajectory(ctx context.Context, db *pgxpool.Pool, sprintID uuid.UUID) error

func ApplySRMFallback(ctx context.Context, db *pgxpool.Pool, eventID, goalID, userID uuid.UUID, fallback string) error
  - switch fallback:
    case "PAUSE": UPDATE global_objectives SET status='PAUSED' WHERE id=$1 ...
    case "L1": INSERT srm_events (..., level='L1', event_type='AUTO_FALLBACK', ...)
    case "L2": INSERT srm_events (..., level='L2', event_type='AUTO_FALLBACK', ...) + INSERT context_adjustments ENERGY_LOW
    default: log + return nil

func CheckAndRecordRegressionEvent(ctx context.Context, db *pgxpool.Pool, goalID, userID uuid.UUID) (bool, error)
  - Check daily_tasks last 5 days for MAIN+DONE count = 0
  - If yes, INSERT srm_events (level='L1', event_type='STAGNATION_5D')
  - Return (triggered bool, err error)

func ComputeALIBreakdown(ctx context.Context, db *pgxpool.Pool, goalID uuid.UUID) (ALIBreakdown, error)
  - aliCurrent: tasks_done_so_far / expected_so_far
  - aliProjected: extrapolate to sprint end
  - inAmbitionBuffer: aliProjected in [1.0, 1.15]
  - velocityControlOn: aliProjected > 1.15

type ALIBreakdown struct {
  AliCurrent, AliProjected float64
  InAmbitionBuffer, VelocityControlOn bool
  GoalBreakdown []GoalALIPoint // empty for now in MVP simplified
  Note string
}

// evolution.go
func MarkEvolutionSprint(ctx context.Context, db *pgxpool.Pool, sprintID uuid.UUID) (bool, error)
  - Compute current sprint_score, prev sprint_score (sprint_number-1)
  - delta = current - prev
  - threshold = 0.05 (default), 0.02 if BM=TACTICAL, 0.05 if STRATEGIC, etc.
  - WAIT — current code uses BMs CREATE/INCREASE/REDUCE/MAINTAIN/EVOLVE. The "ANALYTIC/STRATEGIC/TACTICAL/REACTIVE" set in user-workflow.md is a documentation deviation.
  - Architect decision: keep MVP simplified — fixed threshold 0.05, ignore BM-specific override. POST-MVP will add overrides.
  - If delta >= 0.05: INSERT evolution_sprints ON CONFLICT DO NOTHING
  - Return (isEvolution bool, err)

func ApplyEvolveOverride(behaviorModel string, baseDelta float64) float64
  - For now: return baseDelta (passthrough) — kept as hook for POST-MVP

// growth.go (extend existing)
// MODIFY signature:
func CeremonyTier(score float64, isEvolution bool) string
  switch {
  case score >= 0.90 && isEvolution: return "PLATINUM"
  case score >= 0.90: return "GOLD"  // was PLATINUM in old code
  case score >= 0.80: return "GOLD"  // was GOLD
  case score >= 0.65: return "SILVER"
  default: return "BRONZE"
  }
  IMPORTANT: this changes behavior. Old GOLD threshold (>=0.80) now applies to <0.90. PLATINUM gated by isEvolution.
  Update CALLERS too — but in F8.4 callers session, NOT here.

func GenerateCompletionCeremony(ctx context.Context, db *pgxpool.Pool, sprintID, goID, userID uuid.UUID, score float64, isEvolution bool) error
  - tier = CeremonyTier(score, isEvolution)
  - INSERT ceremonies ... ON CONFLICT (sprint_id) DO NOTHING

// clock.go
type Clock interface {
  Now() time.Time
}
type RealClock struct{}
func (RealClock) Now() time.Time { return time.Now().UTC() }
type MockClock struct{ T time.Time }
func (m *MockClock) Now() time.Time { return m.T }
func (m *MockClock) Advance(d time.Duration) { m.T = m.T.Add(d) }

Update Engine struct to take Clock:
type Engine struct {
  db    *pgxpool.Pool
  redis *redis.Client
  clock Clock
}
Default New() uses RealClock.

Tests: write _test.go for each new function. Coverage target ≥ 80% on engine package.

Do NOT modify scheduler.go or handlers/*.go in this session.

Update CLAUDE.md (section 8 — Integrări existente / engine functions list).
Update ROADMAP.md F8.2 → in_progress, then ✅.
Update BACKLOG.md DEV-11 → IN_REVIEW.

Run: go test ./internal/engine/... -v -cover. Show output.

Romanian summary at end:
- Funcții adăugate (cu signature)
- Funcții modificate (CeremonyTier signature change!)
- Coverage real obținut
- Probleme/decizii notabile
- Backlog status

Commit message:
fix(F8.2): engine restructure — add visualization/regulatory/evolution functions (DEV-11)

- New files: visualization.go, regulatory.go, evolution.go, clock.go
- New funcs: GenerateProgressVisualization, FreezeExpectedTrajectory, UnfreezeExpectedTrajectory,
  ApplySRMFallback, CheckAndRecordRegressionEvent, ComputeALIBreakdown,
  MarkEvolutionSprint, ApplyEvolveOverride, GenerateCompletionCeremony
- BREAKING: CeremonyTier signature now includes isEvolution flag
- Tests added; coverage 8X% on engine package
```

### Acceptance criteria (Gate F8.2):
- [ ] Toate funcțiile listate există și au unit tests
- [ ] Coverage engine ≥ 80%
- [ ] Architect aprobă semantic alignment cu Framework
- [ ] BACKLOG: DEV-11 → IN_REVIEW

---

## SESSION FIX-03 — API Security Hardening (F8.3)

- **Owner:** Security Engineer
- **Reviewer:** Backend Senior + Architect
- **Model:** Opus 4.7
- **Branch:** `claude/fix-03-api-opacity`
- **Backlog:** DEV-01
- **Estimare:** 30 min
- **Pre-condiții:** F8.1 ✅ (poate rula în paralel cu F8.2)

### Prompt:

```
You are a Security Engineer working on NuviaX MVP fix phase F8.3 (API Opacity Hardening).

Read in this order:
1. CLAUDE.md (especially section 7 — Securitate engine)
2. docs/audit/AUDIT-01-deviation-report.md (DEV-01)
3. docs/user-workflow.md sections 8.1, 8.2, TS-12
4. backend/internal/api/handlers/achievements.go
5. backend/internal/api/handlers/*.go (all handlers — for full opacity audit)

Primary goal: fix DEV-01 — remove sprint_score from GET /ceremonies/:goalId response.

Secondary goal: write opacity test that prevents regression.

Step 1 — Fix DEV-01:
In backend/internal/api/handlers/achievements.go GetCeremony:
- Remove `sprint_score` from SELECT and from response
- Response should be: { id, tier, viewed_at }
- Adjust query: SELECT id, tier, viewed_at FROM ceremonies WHERE go_id=$1 ORDER BY created_at DESC LIMIT 1

Step 2 — Audit ALL handlers for opacity violations:
Run grep -rn "drift\|chaos_index\|continuity\|weights\|factors\|penalties\|score_components\|raw_score\|threshold" backend/internal/api/handlers/
Examine each match. If used in JSON output → CRITICAL fix. If used in computation only → OK.

Step 3 — Create opacity test:
File: backend/internal/api/handlers/opacity_test.go
- Use testify and httptest
- For each protected endpoint, mock JWT + minimal DB fixture
- Make request, parse JSON response (deep map[string]interface{})
- Walk recursively, collect all keys
- Assert no key matches forbidden list:
  forbidden := []string{"drift","chaos_index","continuity","weights","factors","penalties","score_components","raw_score","sprint_score","drift_comp","stagnation_comp","inconsistency_comp"}
- Helper: walkKeys(v interface{}, found *[]string)

List of endpoints to scan:
- GET /goals
- GET /goals/:id
- GET /goals/:id/visualize
- GET /today
- GET /dashboard
- GET /srm/status/:goalId
- GET /achievements
- GET /ceremonies/:goalId
- GET /profile/activity

Step 4 — Document allowed fields:
Update docs/integrations.md with section "API Opacity Contract":
List of allowed response fields per endpoint. This becomes contract for future endpoints.

Step 5 — Update CLAUDE.md section 7 with verification command:
go test -run TestOpacity ./internal/api/handlers/... -v

Update ROADMAP.md F8.3 → in_progress → ✅.
Update BACKLOG.md DEV-01 → IN_REVIEW.

Romanian summary:
- Câmpuri eliminate din GetCeremony
- Alte scurgeri găsite în audit (dacă există)
- Test opacity creat — număr endpoints acoperite
- Documentație actualizată

Commit message:
fix(F8.3): API opacity hardening — remove sprint_score leak (DEV-01)

- Remove sprint_score from GET /ceremonies/:goalId response
- Add opacity_test.go covering 9 protected endpoints
- Document API opacity contract in docs/integrations.md
- Update CLAUDE.md section 7 verification commands
```

### Acceptance criteria (Gate F8.3):
- [ ] `go test -run TestOpacity ...` → all pass
- [ ] grep audit finds zero leaks în handlers
- [ ] Documentation updated
- [ ] BACKLOG: DEV-01 → IN_REVIEW

---

## SESSION FIX-04 — Scheduler Wiring (F8.4)

- **Owner:** Backend Senior
- **Reviewer:** Architect + PM
- **Model:** Sonnet 4.6
- **Branch:** `claude/fix-04-scheduler-wiring`
- **Backlog:** DEV-05, DEV-07, DEV-08, DEV-09, DEV-10, DEV-16, DEV-18, DEV-19
- **Estimare:** 90 min
- **Pre-condiții:** F8.1 ✅, F8.2 ✅

### Prompt:

```
You are a Senior Backend Engineer working on NuviaX MVP fix phase F8.4 (Scheduler Wiring).

Read in this order:
1. CLAUDE.md
2. docs/audit/AUDIT-01-deviation-report.md (DEV-05, DEV-07, DEV-08, DEV-09, DEV-10, DEV-16, DEV-18, DEV-19)
3. docs/fix-plan/02-BACKLOG.md (corresponding items)
4. docs/user-workflow.md sections 4, 5
5. backend/internal/scheduler/scheduler.go
6. backend/internal/engine/ (all engine functions added in F8.2)

Goal: wire scheduler jobs to engine functions. Make framework runtime alive.

DEV-05 (jobComputeDailyScore — populate growth_trajectories):
After existing UPSERT daily_scores, add:
  trend := engine.ComputeTrend(realProgress, expectedProgress)
  delta := realProgress - expectedProgress
  _, err := s.db.Exec(ctx, `
    INSERT INTO growth_trajectories (go_id, snapshot_date, actual_pct, expected_pct, delta, trend)
    VALUES ($1, $2, $3, $4, $5, $6)
    ON CONFLICT (go_id, snapshot_date) DO UPDATE SET
      actual_pct = EXCLUDED.actual_pct,
      expected_pct = EXCLUDED.expected_pct,
      delta = EXCLUDED.delta,
      trend = EXCLUDED.trend
  `, goID, today, realProgress, expectedProgress, delta, trend)

DEV-07 (jobGenerateCeremonies + jobCloseExpiredSprints — call achievement award):
Create SQL function in 002_runtime_baseline.sql (if not added in F8.1):
  CREATE OR REPLACE FUNCTION fn_award_achievement_if_earned(p_user_id uuid, p_sprint_id uuid)
    RETURNS void AS $$
  ... (implementation: check streak, sprint count, score thresholds; INSERT achievement_badges idempotent)
  $$ LANGUAGE plpgsql;

After ceremony INSERT in both jobs, add:
  _, err := s.db.Exec(ctx, `SELECT fn_award_achievement_if_earned($1, $2)`, userID, sprintID)
  // log error but don't fail the job

DEV-08 (jobCheckSRMTimeouts — apply fallback):
Replace the log-only block with:
  err := engine.ApplySRMFallback(ctx, s.db, eventID, goID, userID, fallbackAction)
  if err != nil { logger.Error("...", zap.Error(err)) }

DEV-09 (jobDetectEvolution — implement):
Replace placeholder with:
  rows, _ := s.db.Query(ctx, `
    SELECT id FROM sprints
    WHERE status = 'COMPLETED' AND completed_at >= NOW() - INTERVAL '24 hours'
      AND NOT EXISTS (SELECT 1 FROM evolution_sprints WHERE sprint_id = sprints.id)
  `)
  for each: engine.MarkEvolutionSprint(ctx, s.db, sprintID)

DEV-10 (jobComputeWeeklyALI — implement):
Replace placeholder with:
  rows, _ := s.db.Query(ctx, `SELECT id, user_id FROM global_objectives WHERE status='ACTIVE'`)
  for each:
    breakdown, _ := engine.ComputeALIBreakdown(ctx, s.db, goalID)
    UPSERT go_metrics SET ali_current = breakdown.AliCurrent, ali_projected = breakdown.AliProjected, updated_at=NOW()

DEV-16 (jobCloseExpiredSprints — tier with isEvolution):
Before tier := engine.CeremonyTier(sprintScore):
  var isEvolution bool
  _ = s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM evolution_sprints WHERE sprint_id=$1)`, sprintID).Scan(&isEvolution)
  tier := engine.CeremonyTier(sprintScore, isEvolution)

Note: this requires F8.2 CeremonyTier signature change. Imports unchanged.

DEV-18 (jobCheckDailyProgress — add 5-day-inactive trigger):
After existing IsDriftCritical loop, add second loop:
  triggered2, err := engine.CheckAndRecordRegressionEvent(ctx, s.db, goID, userID)
  // helper handles INSERT into srm_events idempotent

DEV-19 (jobCheckStagnation — also INSERT srm_events):
After INSERT stagnation_events, add:
  _, _ = s.db.Exec(ctx, `
    INSERT INTO srm_events (go_id, user_id, level, event_type)
    SELECT $1, $2, 'L1', 'STAGNATION_5D'
    WHERE NOT EXISTS (
      SELECT 1 FROM srm_events
      WHERE go_id=$1 AND event_type='STAGNATION_5D'
        AND created_at >= NOW() - INTERVAL '24 hours'
    )
  `, goID, userID)

Tests:
- backend/internal/scheduler/scheduler_test.go (new) — integration test using testcontainers Postgres
- TestSchedulerEndToEnd: simulate 30-day arc, verify all expected DB state
- Skip if DOCKER_HOST not set (CI will set it)

Update CLAUDE.md, ROADMAP.md, BACKLOG.md.

Run: go test ./internal/scheduler/... -v -tags=integration. Show output.

Romanian summary:
- Joburi modificate: count + descriere scurtă
- Fluxuri framework restabilite: trajectory, achievement, SRM timeout fallback, evolution, ALI
- Test integration coverage
- Backlog status

Commit message:
fix(F8.4): scheduler wiring — populate trajectory/achievements/SRM events (DEV-05/07/08/09/10/16/18/19)

- jobComputeDailyScore now writes growth_trajectories
- jobGenerateCeremonies + jobCloseExpiredSprints call fn_award_achievement_if_earned()
- jobCheckSRMTimeouts applies real fallback via engine.ApplySRMFallback
- jobDetectEvolution implements C31 simplified (delta 0.05 threshold)
- jobComputeWeeklyALI persists ALI breakdown to go_metrics
- jobCloseExpiredSprints uses CeremonyTier(score, isEvolution)
- jobCheckDailyProgress adds 5-day-inactive trigger
- jobCheckStagnation now writes srm_events L1
- Integration test scheduler_test.go added
```

### Acceptance criteria (Gate F8.4):
- [ ] Scheduler integration test pass
- [ ] Manual run: toate cele 12 joburi rulează clean
- [ ] DB state corect după simulare
- [ ] BACKLOG: DEV-05/07/08/09/10/16/18/19 → IN_REVIEW

---

## SESSION FIX-05 — Handler Hardening (F8.5)

- **Owner:** Backend Senior
- **Reviewer:** QA + Architect
- **Model:** Sonnet 4.6
- **Branch:** `claude/fix-05-handlers-hardening`
- **Backlog:** DEV-06, DEV-12, DEV-13, DEV-14, DEV-15, DEV-20, DEV-22, DEV-23, DEV-25, DEV-26, DEV-27, DEV-28
- **Estimare:** 90 min
- **Pre-condiții:** F8.2 ✅, F8.4 ✅

### Prompt:

```
You are a Senior Backend Engineer working on NuviaX MVP fix phase F8.5 (Handler Hardening).

Read in this order:
1. CLAUDE.md
2. docs/audit/AUDIT-01-deviation-report.md (the 12 DEV items in scope)
3. docs/fix-plan/02-BACKLOG.md (DEV-06, 12, 13, 14, 15, 20, 22, 23, 25, 26, 27, 28)
4. docs/user-workflow.md (full)
5. backend/internal/api/handlers/goals.go
6. backend/internal/api/handlers/today.go
7. backend/internal/api/handlers/srm.go
8. backend/internal/api/handlers/achievements.go
9. backend/internal/engine/ (functions from F8.2)

Apply these fixes:

DEV-06 (Day 1 visualization fallback):
In goals.go::GetGoalVisualize, after the rows loop, if len(points)==0:
  data, err := engine.GenerateProgressVisualization(c.Context(), h.db, goalID)
  if err == nil && data != nil {
    points = append(points, dataPoint{
      Date: data.Date.Format("2006-01-02"),
      ProgressPct: data.ActualPct,
      ExpectedPct: data.ExpectedPct,
    })
  }
  // returns at least 1 entry

DEV-12 (SRM L3 freeze):
In srm.go::ConfirmSRML3, after UPDATE global_objectives SET status='PAUSED':
  // get active sprint
  var sprintID uuid.UUID
  _ = h.db.QueryRow(...).Scan(&sprintID)
  if sprintID != uuid.Nil {
    _ = engine.FreezeExpectedTrajectory(c.Context(), h.db, sprintID)
  }
  // compute frozen_expected for response
  var frozenPct float64
  _ = h.db.QueryRow(`SELECT frozen_expected_pct FROM sprints WHERE id=$1`, sprintID).Scan(&frozenPct)
  return c.JSON(fiber.Map{"ok": true, "new_status": "PAUSED", "frozen_expected": frozenPct})

DEV-13 + DEV-14 (energy DB write + label normalization):
In today.go::SetEnergy:
  switch req.Level {
  case "low", "mid": req.Level = "low"
  case "normal", "mid": req.Level = "normal"
  case "high", "hi": req.Level = "high"
  default: return badRequest(c, "Nivel valid: low, normal, high")
  }
  // (handle alias normalization first; "mid" can mean low or normal — pick one. Doc says mid→normal, hi→high.)
  // Actually doc says: mid → normal, hi → high. So:
  // mid → normal; hi → high; rest as-is.

  if req.Level == "normal" { return c.JSON(fiber.Map{"ok": true}) }

  // Find user's primary active GO
  var goID uuid.UUID
  _ = h.db.QueryRow(...).Scan(&goID)
  if goID == uuid.Nil { return badRequest(c, "Nu există obiectiv activ.") }

  adjType := "ENERGY_LOW"
  if req.Level == "high" { adjType = "ENERGY_HIGH" }

  _, err := h.db.Exec(c.Context(), `
    INSERT INTO context_adjustments (go_id, user_id, type, valid_from, valid_until)
    VALUES ($1, $2, $3, CURRENT_DATE, CURRENT_DATE + 1)
  `, goID, userID, adjType)
  if err != nil { return serverError(c, err) }
  return c.JSON(fiber.Map{"ok": true})

DEV-15 (SRM status with ALI breakdown):
In srm.go::GetSRMStatus, replace simple JSON with:
  level, triggeredAt := /* existing query */
  ali, err := engine.ComputeALIBreakdown(c.Context(), h.db, goalID)
  if err != nil { ali = engine.ALIBreakdown{} }

  message := messageForLevel(level) // helper: returns localized RO string

  return c.JSON(fiber.Map{
    "goal_id": goalID,
    "srm_level": level,
    "message": message,
    "triggered_at": triggeredAt,
    "ali": fiber.Map{
      "ali_current": ali.AliCurrent,
      "ali_projected": ali.AliProjected,
      "in_ambition_buffer": ali.InAmbitionBuffer,
      "velocity_control_on": ali.VelocityControlOn,
      "note": ali.Note,
    },
  })

DEV-20 (sprint cap):
In goals.go::CreateGoal, after sprintEnd computation:
  if sprintEnd.After(endDate) {
    sprintEnd = endDate
  }

DEV-22 (suggest-category timeout):
In goals.go::SuggestGOCategory, wrap AI call:
  ctx2, cancel := context.WithTimeout(c.Context(), 2*time.Second)
  defer cancel()
  // pass ctx2 to ai.SuggestGOCategory if signature supports it; else add timeout via channel pattern

DEV-23 (analyze timeout):
Change context.WithTimeout(c.Context(), 8*time.Second) → 3*time.Second.

DEV-25 (streak):
In today.go::computeStreak, change checkDate := tomorrow → checkDate := today. Adjust loop logic to match: first row equals today OR yesterday counts; afterwards strict consecutive.
Add unit test: 7 days DONE up to yesterday, today no DONE → streak = 7.

DEV-26 (achievements envelope):
In achievements.go::ListAchievements:
  return c.JSON(fiber.Map{"achievements": items})
Coordinate frontend update in F8.6.

DEV-27 (ceremony envelope):
In achievements.go::GetCeremony, on no rows:
  return c.JSON(fiber.Map{"ceremony": nil})
On hit: return c.JSON(fiber.Map{"ceremony": fiber.Map{...}})

DEV-28 (behavior model field name):
Decision: API field name = "behavior_model" everywhere.
In goals.go::CreateGoal request struct: rename DominantBehaviorModel → BehaviorModel, json tag "behavior_model".
In response GetGoalDetail: already returns "behavior_model" — OK.
Add backwards-compat: accept BOTH "behavior_model" and "dominant_behavior_model" in CreateGoal request for 1 release. Log deprecation when old name used.
Coordinate frontend send field rename in F8.6.
Document in docs/integrations.md.

Tests:
- For each fix, add at least 1 unit test in *_test.go
- Use testfixtures pattern (introduce simple fixtures helper if missing)
- Coverage handlers ≥ 70%

Update CLAUDE.md, ROADMAP.md, BACKLOG.md.

Run: go test ./internal/api/handlers/... -v -cover. Show output.

Romanian summary:
- Endpoints modificate (count)
- Schimbări notabile contract API
- Coverage handlers
- Backlog status

Commit message:
fix(F8.5): handlers hardening — visualization fallback, SRM ALI, energy DB write (multiple DEVs)

- DEV-06: Day 1 fallback in GetGoalVisualize
- DEV-12: ConfirmSRML3 freezes trajectory + returns frozen_expected
- DEV-13/14: SetEnergy writes context_adjustments; normalizes labels
- DEV-15: GetSRMStatus computes & returns ALI breakdown
- DEV-20: Sprint creation caps at goal end_date
- DEV-22/23: AI timeouts tightened (2s/3s)
- DEV-25: Streak fix off-by-one
- DEV-26/27: Response envelopes for /achievements and /ceremonies
- DEV-28: behavior_model canonical field name
```

### Acceptance criteria (Gate F8.5):
- [ ] Toate testele unit pass; coverage ≥ 70% pe handlers
- [ ] Manual smoke: TS-04, TS-06, TS-08, TS-11 → green
- [ ] BACKLOG: 11 items → IN_REVIEW

---

## SESSION FIX-06 — Frontend Polish (F8.6)

- **Owner:** Frontend Senior
- **Reviewer:** UX + PM
- **Model:** Sonnet 4.6
- **Branch:** `claude/fix-06-frontend-polish`
- **Backlog:** DEV-21 (lead) + DEV-26/27/28 frontend portion
- **Estimare:** 60 min
- **Pre-condiții:** F8.5 ✅

### Prompt:

```
You are a Senior Frontend Engineer working on NuviaX MVP fix phase F8.6 (Frontend Polish).

Read in this order:
1. CLAUDE.md
2. docs/audit/AUDIT-01-deviation-report.md (DEV-21, DEV-26, DEV-27, DEV-28 frontend impact)
3. docs/fix-plan/02-BACKLOG.md
4. frontend/app/app/onboarding/page.tsx
5. frontend/app/app/achievements/page.tsx
6. frontend/app/components/SRMWarning.tsx
7. frontend/app/components/CeremonyModal.tsx (if exists)
8. frontend/app/app/today/page.tsx (energy selector)

Apply these fixes:

DEV-21 (onboarding duration selection):
Add a new step between 'input' and 'parsing' OR within 'input' step:
- Three preset buttons per GO: 30 zile / 90 zile / 180 zile / 365 zile
- Default selected: 90
- Visual: pill buttons grouped horizontally below textarea
- State: goDurations: number[] (one per GO)
- On submit: calculate end_date based on selected duration per GO, send to /goals

DEV-26 (achievements envelope):
In frontend/app/app/achievements/page.tsx:
- Change fetch parser from `data` to `data.achievements` (handle both for 1-release transition)

DEV-27 (ceremony envelope):
In wherever CeremonyModal is fetched:
- Change parser from raw body to `data.ceremony` (handle null gracefully)

DEV-28 (behavior_model field name):
Search for `dominant_behavior_model` in frontend code. Replace with `behavior_model` in request bodies.
Onboarding sends behavior_model: g.behavior_model.

Energy selector (DEV-14 frontend portion):
In today page energy selector, ensure it sends `low`/`normal`/`high` (not `mid`/`hi`).

Build verification:
Run `npm run build` in frontend/app/. Must complete with 0 errors.
Run `npx tsc --noEmit`. Zero errors.

Update CLAUDE.md, ROADMAP.md, BACKLOG.md.

Romanian summary:
- Pagini modificate
- Componente noi (preset duration UI)
- Build status
- Backlog status

Commit message:
fix(F8.6): frontend polish — onboarding duration, response envelopes, behavior_model rename

- DEV-21: Onboarding now lets user choose 30/90/180/365 days per GO
- DEV-26/27: Updated parsers for /achievements and /ceremonies envelopes
- DEV-28: Frontend sends behavior_model (was dominant_behavior_model)
- Energy selector sends low/normal/high (DEV-14 frontend)
- Build clean, 0 TS errors
```

### Acceptance criteria (Gate F8.6):
- [ ] `npm run build` clean
- [ ] Manual onboarding: 4 durate distincte testate
- [ ] BACKLOG: DEV-21 + frontend portions → IN_REVIEW

---

## SESSION FIX-07 — Integration & E2E Tests (F8.7)

- **Owner:** Senior QA Tester
- **Reviewer:** DevOps + Architect + PM
- **Model:** Sonnet 4.6
- **Branch:** `claude/fix-07-test-automation`
- **Estimare:** 120 min
- **Pre-condiții:** F8.1–F8.6 toate ✅

### Prompt:

```
You are a Senior QA Tester working on NuviaX MVP fix phase F8.7 (Integration & E2E Tests).

Read in this order:
1. CLAUDE.md
2. docs/fix-plan/03-TESTING-STRATEGY.md (full)
3. docs/user-workflow.md (TS-01 to TS-12)
4. backend/internal/api/handlers/*_test.go (existing tests)
5. backend/internal/engine/*_test.go
6. backend/internal/scheduler/scheduler.go

Goal: Automate TS-01 through TS-12 end-to-end. Build CI pipeline. Coverage report.

Phase 1 — Integration tests (Go, build tag `integration`):
Create backend/internal/api/handlers/integration_test.go (or split per concern):
- TS-01_TestHappyPath: register → login → onboarding → first task complete → progress
- TS-02_TestVaultLimit: 4 GOs → 4th vaulted
- TS-03_TestTrajectory: 2 day simulate → trajectory has 2+ entries
- TS-04_TestSRML1Auto: 5-day-inactive simulate → srm_level=L1
- TS-05_TestSRML2Reduce: chaos threshold → confirm L2 → context_adjustment created
- TS-06_TestSRML3Freeze: L3 confirm → goal PAUSED + frozen_expected
- TS-07_TestAchievement: sprint close → ceremony + badge
- TS-08_TestVisualizationDay1: new goal → 1 entry trajectory
- TS-09_TestPersonalLimit: 3rd personal task rejected
- TS-10_TestThemePersist: PATCH theme → re-login → theme persists
- TS-11_TestAITimeout: AI mock with chaos → response < 2s
- TS-12_TestOpacity: full scan via opacity_test.go (already in F8.3)

Use testcontainers-go for Postgres. Real Redis or miniredis. Mock AI via interface.

Phase 2 — E2E tests (Playwright):
Create frontend/app/__tests__/e2e/:
- happy-path.spec.ts (TS-01)
- personal-task-limit.spec.ts (TS-09)
- theme-persist.spec.ts (TS-10)
- onboarding-duration.spec.ts (DEV-21 verify)

Phase 3 — CI pipeline:
Create .github/workflows/ci.yml:
- jobs: lint (golangci-lint), unit (go test), integration (go test -tags=integration), opacity (go test -run TestOpacity), schema-check (bash schema-check.sh), frontend-build (npm run build), frontend-typecheck (tsc --noEmit), e2e (playwright)
- runs on: pull_request, push to main
- required checks: all jobs

Phase 4 — Coverage report:
Run `go test ./... -cover -coverprofile=coverage.out`
Generate `docs/testing/F8.7-coverage-report.md` with:
- Per-package coverage
- Top 10 uncovered functions
- Coverage gates: engine ≥ 80%, handlers ≥ 70%

Phase 5 — Test gap analysis:
Create `docs/testing/F8.7-test-gaps.md`:
- Edge cases NOT covered (e.g., concurrent task completion, race conditions, network partition)
- Plan for POST-MVP coverage

Update CLAUDE.md, ROADMAP.md.

Run: full CI pipeline locally. All jobs green.

Romanian summary:
- Tests adăugate (count per nivel)
- CI jobs configurate
- Coverage real
- Test gaps identificate

Commit message:
test(F8.7): full test automation — integration, E2E, CI pipeline

- 12 integration tests covering TS-01 through TS-12
- 4 Playwright E2E specs
- CI pipeline with 7 required checks
- Coverage report: engine X%, handlers Y%
- Test gap analysis for POST-MVP
```

### Acceptance criteria (Gate F8.7):
- [ ] CI pipeline green on branch
- [ ] Coverage gates met
- [ ] Toate TS-01..TS-12 automate
- [ ] Test gap report livrat

---

## SESSION FIX-08 — Staging + Production Validation (F8.8)

- **Owner:** DevOps + QA + PM
- **Reviewer:** Architect + Security + Framework Owner (sign-off)
- **Model:** Opus 4.7 (decizii) + Sonnet 4.6 (execuție)
- **Branch:** `claude/fix-08-staging-validation`
- **Estimare:** 90 min
- **Pre-condiții:** F8.7 ✅

### Prompt:

```
You are leading NuviaX MVP fix phase F8.8 (Staging + Production Validation) — the FINAL gate before MVP launch.

Read in this order:
1. CLAUDE.md
2. docs/fix-plan/00-PLAN-MASTER.md section 6 (MVP launch criteria)
3. docs/fix-plan/03-TESTING-STRATEGY.md (gates F8.8)
4. docs/audit/AUDIT-01-deviation-report.md (final state)
5. docs/fix-plan/02-BACKLOG.md (verify all RESOLVED or ACCEPTED_POST_MVP)
6. infra/docker-compose.yml
7. infra/.env.example

Step 1 — Backlog audit:
Verify in docs/fix-plan/02-BACKLOG.md: zero items in OPEN/IN_PROGRESS. All in RESOLVED or ACCEPTED_POST_MVP with justification.

Step 2 — Pre-deploy checks:
- Run full CI locally: all green
- Run `gosec ./backend/...` → 0 HIGH/CRITICAL
- Run `npm audit --audit-level=high` in frontend → 0 issues
- Verify docker-compose builds clean: `docker compose -f infra/docker-compose.yml build`
- Verify .env.example has all required variables documented

Step 3 — Deploy staging:
Document the deploy procedure in docs/runbook/staging-deploy.md (new):
- Prerequisites (ENV vars)
- Steps (backup DB, deploy, run migrations, smoke check)
- Rollback procedure
Execute staging deploy (commands in runbook).
Verify backend logs clean for 5 minutes.
Verify frontend loads, /health returns 200.

Step 4 — Smoke manual full TS-01..TS-12 on staging:
Use docs/testing/F8.8-smoke-checklist.md (create new).
Each TS:
- Steps numerotate
- Expected
- Actual (record verbatim)
- PASS / FAIL
- Screenshot URL or note

Step 5 — Performance baseline:
Use k6 or vegeta, hit staging endpoints with realistic load:
- 50 concurrent users for 60s on each endpoint
- Capture P50/P95/P99 in `docs/testing/F8.8-performance-baseline.md`

Step 6 — Security scan:
- gosec full report
- npm audit JSON output
- Manual: bcrypt cost check, JWT alg check, admin 404 verify, forgot-password timing measurement
- Save to `docs/testing/F8.8-security-report.md`

Step 7 — Sign-off:
Each role commits sign-off file:
- docs/fix-plan/sign-off/architect.txt
- docs/fix-plan/sign-off/qa.txt
- docs/fix-plan/sign-off/security.txt
- docs/fix-plan/sign-off/devops.txt
- docs/fix-plan/sign-off/pm.txt
Each contains: name (or role label), date, "I approve MVP launch" + any conditions/concerns.

Step 8 — Final docs:
Update README.md to v1.5.0:
- Production deploy section
- Known limitations (post-MVP roadmap link)
Update CLAUDE.md to v1.5.0 — Phase F8 ✅, MVP READY.
Update ROADMAP.md F8.8 → ✅ + section "MVP launched 2026-XX-XX".

Final report: docs/testing/F8.8-mvp-launch-report.md
- Audit history (AUDIT-01 → 28 deviations)
- Fix history (8 fix sessions)
- Test coverage
- Performance baseline
- Security posture
- Outstanding POST-MVP items
- Launch authorization

Romanian summary at end:
- 28 devieri rezolvate / 28 (sau X / 28 + lista celor amânate POST-MVP)
- TS-01..TS-12: 12/12 PASS pe staging
- Performance: P95 = Xms (target Yms)
- Security: 0 HIGH
- Sign-off: 5/5 obținute
- MVP authorized for launch: DA / NU + condiții

Commit message:
release(F8.8): MVP launch validation — all gates green, sign-off obtained

- Staging deploy successful
- TS-01..TS-12: 12/12 pass
- Performance baseline documented
- Security: 0 HIGH/CRITICAL
- Sign-off: Architect, QA, Security, DevOps, PM
- AUDIT-01: 27/28 deviations RESOLVED, 1 POST-MVP
```

### Acceptance criteria (Gate F8.8 — FINAL):
- [ ] Toate gate-urile F8.1–F8.7 verzi
- [ ] Smoke staging: 12/12 PASS
- [ ] Performance documentat
- [ ] Security: 0 HIGH/CRITICAL
- [ ] 5 sign-off-uri prezente
- [ ] **MVP autorizat pentru lansare**

---

## Lecții pentru POST-MVP

După F8.8, deschide imediat:
- Repository post-mvp/ROADMAP-POST-MVP.md cu C15-C18, C29, C31 full, C34-C36, C39, C40
- Migrare la pre-commit hooks (pentru a nu repeta scurgerile de internal fields)
- Migrare la trunk-based development cu feature flags
- Quarterly schema review (auto-script de schema drift detection)
