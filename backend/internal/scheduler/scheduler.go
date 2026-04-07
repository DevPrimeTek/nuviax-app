// Package scheduler — NuviaX background cron jobs (F4)
//
// 12 jobs running on UTC schedule. Each job uses a 5-minute context timeout.
// AI (Claude Haiku) and Email (Resend) clients are optional — nil means graceful degradation.
package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	aiPkg "github.com/devprimetek/nuviax-app/internal/ai"
	"github.com/devprimetek/nuviax-app/internal/email"
	"github.com/devprimetek/nuviax-app/internal/engine"
	"github.com/devprimetek/nuviax-app/pkg/crypto"
	"github.com/devprimetek/nuviax-app/pkg/logger"
)

const jobTimeout = 5 * time.Minute

// Scheduler runs background jobs powered by the NuviaX Growth Framework Engine.
type Scheduler struct {
	cron   *cron.Cron
	db     *pgxpool.Pool
	ai     *aiPkg.Client  // optional — nil if ANTHROPIC_API_KEY not set
	email  *email.Client  // optional — nil if RESEND_API_KEY not set
	encKey []byte
}

// New creates a Scheduler. aiClient and emailClient may be nil (graceful degradation).
func New(pool *pgxpool.Pool, aiClient *aiPkg.Client, emailClient *email.Client, encKey []byte) *Scheduler {
	c := cron.New(
		cron.WithLocation(time.UTC),
		cron.WithLogger(cron.DefaultLogger),
	)
	return &Scheduler{cron: c, db: pool, ai: aiClient, email: emailClient, encKey: encKey}
}

// Start registers all 12 cron jobs and starts the scheduler.
func (s *Scheduler) Start() {
	// ── Night batch (00:00 UTC) ────────────────────────────────
	s.cron.AddFunc("0 0 * * *", s.jobCloseExpiredSprints)  // job 4
	s.cron.AddFunc("1 0 * * *", s.jobGenerateDailyTasks)   // job 1
	s.cron.AddFunc("5 0 * * *", s.jobStartNextSprints)     // job 5

	// ── Early morning (01:00 UTC) ──────────────────────────────
	s.cron.AddFunc("0 1 * * *", s.jobDetectEvolution)      // job 11
	s.cron.AddFunc("5 1 * * *", s.jobGenerateCeremonies)   // job 10
	s.cron.AddFunc("10 1 * * *", s.jobComputeGORI)         // job 12

	// ── Late night (23:xx UTC) ─────────────────────────────────
	s.cron.AddFunc("50 23 * * *", s.jobComputeDailyScore)   // job 2
	s.cron.AddFunc("55 23 * * *", s.jobCheckDailyProgress)  // job 3
	s.cron.AddFunc("58 23 * * *", s.jobCheckStagnation)     // job 8

	// ── Hourly ────────────────────────────────────────────────
	s.cron.AddFunc("0 * * * *", s.jobCheckSRMTimeouts)      // job 9

	// ── Weekly (Sunday UTC) ───────────────────────────────────
	s.cron.AddFunc("0 2 * * 0", s.jobRecalibrateRelevance)  // job 7
	s.cron.AddFunc("0 3 * * 0", s.jobComputeWeeklyALI)      // job 6

	s.cron.Start()
	logger.Info("Scheduler started", zap.Int("jobs", 12))
}

// Stop gracefully waits for running jobs to finish.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	logger.Info("Scheduler stopped")
}

// ── Job 1: Generate Daily Tasks (00:01 UTC) ────────────────────────────────────
// C23 Daily Stack Generator — creates task stack for each ACTIVE GO.

func (s *Scheduler) jobGenerateDailyTasks() {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	logger.Info("jobGenerateDailyTasks: start")

	rows, err := s.db.Query(ctx, `
		SELECT g.id, g.name, g.user_id, s.id AS sprint_id, s.sprint_number
		FROM global_objectives g
		JOIN sprints s ON s.go_id = g.id
		WHERE g.status = 'ACTIVE' AND s.status = 'ACTIVE'
	`)
	if err != nil {
		logger.Error("jobGenerateDailyTasks: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	inserted := 0

	for rows.Next() {
		var goID, userID, sprintID uuid.UUID
		var goalName string
		var sprintNum int

		if err := rows.Scan(&goID, &goalName, &userID, &sprintID, &sprintNum); err != nil {
			logger.Error("jobGenerateDailyTasks: scan error", zap.Error(err))
			continue
		}

		// Skip if tasks already exist for today
		var existing int
		_ = s.db.QueryRow(ctx,
			`SELECT COUNT(*) FROM daily_tasks WHERE sprint_id = $1 AND task_date = $2`,
			sprintID, today,
		).Scan(&existing)
		if existing > 0 {
			continue
		}

		// Generate task texts: AI first, fallback on error or nil client
		checkpointName := fmt.Sprintf("Sprint %d", sprintNum)
		var texts []string

		if s.ai != nil {
			aiCtx, aiCancel := context.WithTimeout(ctx, 12*time.Second)
			generated, aiErr := s.ai.GenerateTaskTexts(aiCtx, goalName, checkpointName, sprintNum, 2)
			aiCancel()
			if aiErr == nil && len(generated) > 0 {
				texts = generated
			} else if aiErr != nil {
				logger.Error("jobGenerateDailyTasks: AI error, using fallback",
					zap.String("go_id", goID.String()), zap.Error(aiErr))
			}
		}

		if len(texts) == 0 {
			texts = []string{
				fmt.Sprintf("Activitate pentru %s", goalName),
				fmt.Sprintf("Continuare %s", checkpointName),
			}
		}

		taskTypes := []string{"MAIN", "SUPPORT"}
		for i, text := range texts {
			tt := "MAIN"
			if i < len(taskTypes) {
				tt = taskTypes[i]
			}
			if _, err := s.db.Exec(ctx, `
				INSERT INTO daily_tasks (go_id, sprint_id, user_id, task_date, title, task_type)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, goID, sprintID, userID, today, text, tt); err != nil {
				logger.Error("jobGenerateDailyTasks: insert error",
					zap.String("go_id", goID.String()), zap.Error(err))
			} else {
				inserted++
			}
		}
	}

	logger.Info("jobGenerateDailyTasks: done", zap.Int("tasks_inserted", inserted))
}

// ── Job 2: Compute Daily Score (23:50 UTC) ────────────────────────────────────
// C24 Progress Computation + C25 Execution Variance.

func (s *Scheduler) jobComputeDailyScore() {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	logger.Info("jobComputeDailyScore: start")

	rows, err := s.db.Query(ctx, `
		SELECT g.id, g.user_id, s.id, s.start_date,
		       COUNT(dt.id) FILTER (WHERE dt.status = 'DONE')  AS done,
		       COUNT(dt.id)                                     AS total
		FROM global_objectives g
		JOIN sprints s ON s.go_id = g.id
		LEFT JOIN daily_tasks dt ON dt.sprint_id = s.id AND dt.task_date <= CURRENT_DATE
		WHERE g.status = 'ACTIVE' AND s.status = 'ACTIVE'
		GROUP BY g.id, g.user_id, s.id, s.start_date
	`)
	if err != nil {
		logger.Error("jobComputeDailyScore: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	upserted := 0

	for rows.Next() {
		var goID, userID, sprintID uuid.UUID
		var sprintStart time.Time
		var done, total int64

		if err := rows.Scan(&goID, &userID, &sprintID, &sprintStart, &done, &total); err != nil {
			logger.Error("jobComputeDailyScore: scan error", zap.Error(err))
			continue
		}

		dayInSprint := int(today.Sub(sprintStart.Truncate(24*time.Hour)).Hours()/24) + 1
		if dayInSprint < 1 {
			dayInSprint = 1
		}

		realProgress := engine.ComputeProgress(int(done), int(total))
		expectedProgress := engine.ComputeExpected(dayInSprint)
		drift := engine.ComputeDrift(realProgress, expectedProgress)

		if _, err := s.db.Exec(ctx, `
			INSERT INTO daily_scores
				(go_id, user_id, sprint_id, score_date, real_progress, expected_progress, drift, tasks_done, tasks_total)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (go_id, score_date) DO UPDATE SET
				real_progress     = EXCLUDED.real_progress,
				expected_progress = EXCLUDED.expected_progress,
				drift             = EXCLUDED.drift,
				tasks_done        = EXCLUDED.tasks_done,
				tasks_total       = EXCLUDED.tasks_total,
				computed_at       = NOW()
		`, goID, userID, sprintID, today, realProgress, expectedProgress, drift, done, total); err != nil {
			logger.Error("jobComputeDailyScore: upsert error",
				zap.String("go_id", goID.String()), zap.Error(err))
		} else {
			upserted++
		}
	}

	logger.Info("jobComputeDailyScore: done", zap.Int("scores_upserted", upserted))
}

// ── Job 3: Check Daily Progress (23:55 UTC) ───────────────────────────────────
// C26 Drift Engine — inserts SRM L1 event when drift is critical.

func (s *Scheduler) jobCheckDailyProgress() {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	logger.Info("jobCheckDailyProgress: start")

	rows, err := s.db.Query(ctx, `
		SELECT ds.go_id, g.user_id,
		       array_agg(ds.drift ORDER BY ds.score_date DESC) AS drifts
		FROM daily_scores ds
		JOIN global_objectives g ON g.id = ds.go_id
		WHERE ds.score_date >= CURRENT_DATE - INTERVAL '7 days'
		  AND g.status = 'ACTIVE'
		GROUP BY ds.go_id, g.user_id
	`)
	if err != nil {
		logger.Error("jobCheckDailyProgress: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	triggered := 0
	for rows.Next() {
		var goID, userID uuid.UUID
		var drifts []float64

		if err := rows.Scan(&goID, &userID, &drifts); err != nil {
			logger.Error("jobCheckDailyProgress: scan error", zap.Error(err))
			continue
		}

		if !engine.IsDriftCritical(drifts) {
			continue
		}

		if _, err := s.db.Exec(ctx, `
			INSERT INTO srm_events (go_id, user_id, level, event_type, created_at)
			VALUES ($1, $2, 'L1', 'DRIFT_CRITICAL', NOW())
		`, goID, userID); err != nil {
			logger.Error("jobCheckDailyProgress: insert srm_event error",
				zap.String("go_id", goID.String()), zap.Error(err))
		} else {
			triggered++
		}
	}

	logger.Info("jobCheckDailyProgress: done", zap.Int("l1_events", triggered))
}

// ── Job 4: Close Expired Sprints (00:00 UTC) ──────────────────────────────────
// C37 Sprint Score + C20 Sprint Target — closes sprints, sends email notification.

func (s *Scheduler) jobCloseExpiredSprints() {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	logger.Info("jobCloseExpiredSprints: start")

	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.go_id, s.user_id, s.sprint_number,
		       g.name AS goal_name,
		       u.email_encrypted, u.full_name
		FROM sprints s
		JOIN global_objectives g ON g.id = s.go_id
		JOIN users u ON u.id = s.user_id
		WHERE s.status = 'ACTIVE' AND s.end_date < CURRENT_DATE
	`)
	if err != nil {
		logger.Error("jobCloseExpiredSprints: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	closed := 0
	for rows.Next() {
		var sprintID, goID, userID uuid.UUID
		var sprintNum int
		var goalName, emailEnc string
		var fullName *string

		if err := rows.Scan(&sprintID, &goID, &userID, &sprintNum,
			&goalName, &emailEnc, &fullName); err != nil {
			logger.Error("jobCloseExpiredSprints: scan error", zap.Error(err))
			continue
		}

		// Fetch sprint metrics from daily_scores
		var progressComp, consistencyComp, deviationComp float64
		_ = s.db.QueryRow(ctx, `
			SELECT
				COALESCE(MAX(real_progress), 0),
				COALESCE(
					COUNT(DISTINCT score_date) FILTER (WHERE tasks_done > 0)::float / 30.0,
					0
				),
				COALESCE(1.0 - LEAST(AVG(ABS(drift)), 1.0), 0.5)
			FROM daily_scores
			WHERE sprint_id = $1
		`, sprintID).Scan(&progressComp, &consistencyComp, &deviationComp)

		sprintScore := engine.ComputeSprintScore(progressComp, consistencyComp, deviationComp)
		grade := engine.ScoreToGrade(sprintScore)
		tier := engine.CeremonyTier(sprintScore)

		// Mark sprint COMPLETED
		if _, err := s.db.Exec(ctx, `
			UPDATE sprints
			SET status = 'COMPLETED', sprint_score = $1, grade = $2, completed_at = NOW()
			WHERE id = $3
		`, sprintScore, grade, sprintID); err != nil {
			logger.Error("jobCloseExpiredSprints: update sprint error",
				zap.String("sprint_id", sprintID.String()), zap.Error(err))
			continue
		}

		// Insert sprint result record
		if _, err := s.db.Exec(ctx, `
			INSERT INTO sprint_results (sprint_id, go_id, user_id, sprint_score, grade, ceremony_tier, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW())
			ON CONFLICT (sprint_id) DO NOTHING
		`, sprintID, goID, userID, sprintScore, grade, tier); err != nil {
			logger.Error("jobCloseExpiredSprints: insert sprint_results error",
				zap.String("sprint_id", sprintID.String()), zap.Error(err))
		}

		// Insert ceremony
		if _, err := s.db.Exec(ctx, `
			INSERT INTO ceremonies (sprint_id, go_id, user_id, tier, created_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (sprint_id) DO NOTHING
		`, sprintID, goID, userID, tier); err != nil {
			logger.Error("jobCloseExpiredSprints: insert ceremony error",
				zap.String("sprint_id", sprintID.String()), zap.Error(err))
		}

		closed++

		// Send sprint complete email — fire-and-forget
		if s.email != nil {
			name := "utilizator"
			if fullName != nil && *fullName != "" {
				name = *fullName
			}
			emailEnc := emailEnc
			sprintNum := sprintNum
			goalName := goalName
			grade := grade
			go func() {
				emailCtx, emailCancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer emailCancel()
				plainEmail, decErr := crypto.Decrypt(emailEnc, s.encKey)
				if decErr != nil {
					logger.Error("jobCloseExpiredSprints: decrypt email error", zap.Error(decErr))
					return
				}
				if err := s.email.SendSprintComplete(emailCtx, plainEmail, name, goalName, grade, sprintNum); err != nil {
					logger.Error("jobCloseExpiredSprints: send email error", zap.Error(err))
				}
			}()
		}
	}

	logger.Info("jobCloseExpiredSprints: done", zap.Int("sprints_closed", closed))
}

// ── Job 5: Start Next Sprints (00:05 UTC) ─────────────────────────────────────
// C19 Sprint Structuring — creates a new sprint for ACTIVE GOs without one.

func (s *Scheduler) jobStartNextSprints() {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	logger.Info("jobStartNextSprints: start")

	// ACTIVE GOs with no ACTIVE sprint
	rows, err := s.db.Query(ctx, `
		SELECT g.id, g.user_id,
		       COALESCE(MAX(sp.sprint_number), 0) AS last_sprint
		FROM global_objectives g
		LEFT JOIN sprints sp ON sp.go_id = g.id
		WHERE g.status = 'ACTIVE'
		  AND NOT EXISTS (
		      SELECT 1 FROM sprints s2
		      WHERE s2.go_id = g.id AND s2.status = 'ACTIVE'
		  )
		GROUP BY g.id, g.user_id
	`)
	if err != nil {
		logger.Error("jobStartNextSprints: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	started := 0

	for rows.Next() {
		var goID, userID uuid.UUID
		var lastSprint int

		if err := rows.Scan(&goID, &userID, &lastSprint); err != nil {
			logger.Error("jobStartNextSprints: scan error", zap.Error(err))
			continue
		}

		nextNum := lastSprint + 1
		startDate := today
		endDate := today.AddDate(0, 0, 30)

		if _, err := s.db.Exec(ctx, `
			INSERT INTO sprints (go_id, user_id, sprint_number, start_date, end_date, status)
			VALUES ($1, $2, $3, $4, $5, 'ACTIVE')
			ON CONFLICT (go_id, sprint_number) DO NOTHING
		`, goID, userID, nextNum, startDate, endDate); err != nil {
			logger.Error("jobStartNextSprints: insert sprint error",
				zap.String("go_id", goID.String()), zap.Error(err))
		} else {
			started++
		}
	}

	logger.Info("jobStartNextSprints: done", zap.Int("sprints_started", started))
}

// ── Job 6: Compute Weekly ALI (Sunday 03:00 UTC) ──────────────────────────────
// C38 GORI placeholder — ALI computation deferred to post-MVP.

func (s *Scheduler) jobComputeWeeklyALI() {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()
	_ = ctx

	// TODO(F8): implement C38 ALI computation (post-MVP)
	logger.Info("jobComputeWeeklyALI: placeholder — ALI computation deferred to post-MVP")
}

// ── Job 7: Recalibrate Relevance (Sunday 02:00 UTC) ───────────────────────────
// C28 Chaos Index — triggers SRM L2 when chaos ≥ 0.40.

func (s *Scheduler) jobRecalibrateRelevance() {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	logger.Info("jobRecalibrateRelevance: start")

	rows, err := s.db.Query(ctx, `
		SELECT ds.go_id, g.user_id,
		       COALESCE(AVG(ABS(ds.drift)), 0)             AS drift_comp,
		       COALESCE(
		           1.0 - (COUNT(DISTINCT ds.score_date) FILTER (WHERE ds.tasks_done > 0)::float
		               / NULLIF(COUNT(DISTINCT ds.score_date), 0)
		           ), 0
		       )                                            AS stagnation_comp,
		       COALESCE(STDDEV(ds.real_progress), 0)       AS inconsistency_comp
		FROM daily_scores ds
		JOIN global_objectives g ON g.id = ds.go_id
		WHERE ds.score_date >= CURRENT_DATE - INTERVAL '30 days'
		  AND g.status = 'ACTIVE'
		GROUP BY ds.go_id, g.user_id
	`)
	if err != nil {
		logger.Error("jobRecalibrateRelevance: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	triggered := 0
	for rows.Next() {
		var goID, userID uuid.UUID
		var driftComp, stagnationComp, inconsistencyComp float64

		if err := rows.Scan(&goID, &userID, &driftComp, &stagnationComp, &inconsistencyComp); err != nil {
			logger.Error("jobRecalibrateRelevance: scan error", zap.Error(err))
			continue
		}

		chaosIndex := engine.ComputeChaosIndex(driftComp, stagnationComp, inconsistencyComp)
		if chaosIndex < 0.40 {
			continue
		}

		if _, err := s.db.Exec(ctx, `
			INSERT INTO srm_events (go_id, user_id, level, event_type, created_at)
			VALUES ($1, $2, 'L2', 'CHAOS_INDEX_HIGH', NOW())
		`, goID, userID); err != nil {
			logger.Error("jobRecalibrateRelevance: insert srm_event error",
				zap.String("go_id", goID.String()), zap.Error(err))
		} else {
			triggered++
		}
	}

	logger.Info("jobRecalibrateRelevance: done", zap.Int("l2_events", triggered))
}

// ── Job 8: Check Stagnation (23:58 UTC) ───────────────────────────────────────
// C27 Stagnation Detection — 5+ consecutive days with no completed tasks.

func (s *Scheduler) jobCheckStagnation() {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	logger.Info("jobCheckStagnation: start")

	rows, err := s.db.Query(ctx, `
		SELECT g.id, g.user_id
		FROM global_objectives g
		JOIN sprints s ON s.go_id = g.id AND s.status = 'ACTIVE'
		WHERE g.status = 'ACTIVE'
		  AND NOT EXISTS (
		      SELECT 1 FROM daily_tasks dt
		      WHERE dt.go_id = g.id
		        AND dt.status = 'DONE'
		        AND dt.task_date >= CURRENT_DATE - INTERVAL '5 days'
		  )
	`)
	if err != nil {
		logger.Error("jobCheckStagnation: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	inserted := 0
	for rows.Next() {
		var goID, userID uuid.UUID
		if err := rows.Scan(&goID, &userID); err != nil {
			logger.Error("jobCheckStagnation: scan error", zap.Error(err))
			continue
		}

		if _, err := s.db.Exec(ctx, `
			INSERT INTO stagnation_events (go_id, user_id, days_inactive, detected_at)
			VALUES ($1, $2, 5, NOW())
		`, goID, userID); err != nil {
			logger.Error("jobCheckStagnation: insert error",
				zap.String("go_id", goID.String()), zap.Error(err))
		} else {
			inserted++
		}
	}

	logger.Info("jobCheckStagnation: done", zap.Int("stagnation_events", inserted))
}

// ── Job 9: Check SRM Timeouts (hourly) ────────────────────────────────────────
// C33 SRM — unconfirmed L3 events trigger ComputeSRMFallback.

func (s *Scheduler) jobCheckSRMTimeouts() {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	logger.Info("jobCheckSRMTimeouts: start")

	rows, err := s.db.Query(ctx, `
		SELECT id, go_id, created_at
		FROM srm_events
		WHERE level = 'L3'
		  AND confirmed_at IS NULL
		  AND created_at <= NOW() - INTERVAL '24 hours'
	`)
	if err != nil {
		if isTableMissing(err) {
			logger.Info("jobCheckSRMTimeouts: srm_events table not ready, skipping")
			return
		}
		logger.Error("jobCheckSRMTimeouts: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	processed := 0
	for rows.Next() {
		var eventID, goID uuid.UUID
		var createdAt time.Time

		if err := rows.Scan(&eventID, &goID, &createdAt); err != nil {
			logger.Error("jobCheckSRMTimeouts: scan error", zap.Error(err))
			continue
		}

		hoursSince := time.Since(createdAt).Hours()
		fallbackAction := engine.ComputeSRMFallback(hoursSince)

		logger.Info("jobCheckSRMTimeouts: L3 fallback computed",
			zap.String("event_id", eventID.String()),
			zap.String("go_id", goID.String()),
			zap.String("action", fallbackAction),
			zap.Float64("hours_since", hoursSince),
		)
		processed++
	}

	logger.Info("jobCheckSRMTimeouts: done", zap.Int("l3_processed", processed))
}

// ── Job 10: Generate Ceremonies (01:05 UTC) ───────────────────────────────────
// C37 Sprint Score — inserts ceremony records for completed sprints without one.

func (s *Scheduler) jobGenerateCeremonies() {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	logger.Info("jobGenerateCeremonies: start")

	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.go_id, s.user_id, s.sprint_score
		FROM sprints s
		WHERE s.status = 'COMPLETED'
		  AND s.sprint_score IS NOT NULL
		  AND NOT EXISTS (
		      SELECT 1 FROM ceremonies c WHERE c.sprint_id = s.id
		  )
	`)
	if err != nil {
		if isTableMissing(err) {
			logger.Info("jobGenerateCeremonies: ceremonies table not ready, skipping")
			return
		}
		logger.Error("jobGenerateCeremonies: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	inserted := 0
	for rows.Next() {
		var sprintID, goID, userID uuid.UUID
		var sprintScore float64

		if err := rows.Scan(&sprintID, &goID, &userID, &sprintScore); err != nil {
			logger.Error("jobGenerateCeremonies: scan error", zap.Error(err))
			continue
		}

		tier := engine.CeremonyTier(sprintScore)

		if _, err := s.db.Exec(ctx, `
			INSERT INTO ceremonies (sprint_id, go_id, user_id, tier, created_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (sprint_id) DO NOTHING
		`, sprintID, goID, userID, tier); err != nil {
			logger.Error("jobGenerateCeremonies: insert error",
				zap.String("sprint_id", sprintID.String()), zap.Error(err))
		} else {
			inserted++
		}
	}

	logger.Info("jobGenerateCeremonies: done", zap.Int("ceremonies_inserted", inserted))
}

// ── Job 11: Detect Evolution (01:00 UTC) ──────────────────────────────────────
// C31 Behavioral Patterns — placeholder (post-MVP).

func (s *Scheduler) jobDetectEvolution() {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()
	_ = ctx

	// TODO(F8): implement C31 behavioral pattern detection (post-MVP)
	logger.Info("jobDetectEvolution: placeholder — behavioral pattern detection deferred to post-MVP")
}

// ── Job 12: Compute GORI (01:10 UTC) ──────────────────────────────────────────
// C38 GORI — computes Growth Objective Rate Index and persists to go_metrics.

func (s *Scheduler) jobComputeGORI() {
	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	logger.Info("jobComputeGORI: start")

	rows, err := s.db.Query(ctx, `
		SELECT g.id, g.user_id,
		       array_agg(s.sprint_score ORDER BY s.sprint_number) FILTER (WHERE s.sprint_score IS NOT NULL) AS scores,
		       COUNT(s.id) FILTER (WHERE s.status = 'COMPLETED') AS completed,
		       COUNT(s.id)                                        AS total
		FROM global_objectives g
		JOIN sprints s ON s.go_id = g.id
		WHERE g.status IN ('ACTIVE', 'COMPLETED')
		GROUP BY g.id, g.user_id
		HAVING COUNT(s.id) > 0
	`)
	if err != nil {
		logger.Error("jobComputeGORI: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	updated := 0
	for rows.Next() {
		var goID, userID uuid.UUID
		var scores []float64
		var completed, total int64

		if err := rows.Scan(&goID, &userID, &scores, &completed, &total); err != nil {
			logger.Error("jobComputeGORI: scan error", zap.Error(err))
			continue
		}

		if len(scores) == 0 {
			continue
		}

		gori := engine.ComputeGORI(scores, int(completed), int(total))

		if _, err := s.db.Exec(ctx, `
			INSERT INTO go_metrics (go_id, user_id, gori_score, updated_at)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (go_id) DO UPDATE SET
				gori_score = EXCLUDED.gori_score,
				updated_at = NOW()
		`, goID, userID, gori); err != nil {
			logger.Error("jobComputeGORI: upsert error",
				zap.String("go_id", goID.String()), zap.Error(err))
		} else {
			updated++
		}
	}

	logger.Info("jobComputeGORI: done", zap.Int("metrics_updated", updated))
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// isTableMissing returns true when the error is a PostgreSQL "undefined table" error.
// Used for jobs that reference tables which may not exist yet during development.
func isTableMissing(err error) bool {
	if err == nil {
		return false
	}
	// pgx wraps pgconn.PgError; check by message for "undefined_table" (42P01)
	return pgx.ErrNoRows != err && len(err.Error()) > 0 &&
		containsAny(err.Error(), "42P01", "undefined_table", "does not exist")
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
	}
	return false
}
