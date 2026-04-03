package scheduler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/devprimetek/nuviax-app/internal/cache"
	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/email"
	"github.com/devprimetek/nuviax-app/internal/engine"
	"github.com/devprimetek/nuviax-app/pkg/crypto"
	"github.com/devprimetek/nuviax-app/pkg/logger"
)

// Scheduler runs background jobs powered by the NUViaX Framework Engine
type Scheduler struct {
	cron   *cron.Cron
	db     *pgxpool.Pool
	redis  *redis.Client
	engine *engine.Engine
	email  *email.Client // optional: nil if RESEND_API_KEY not set
	encKey []byte
}

func New(pool *pgxpool.Pool, rdb *redis.Client, eng *engine.Engine, emailClient *email.Client, encKey []byte) *Scheduler {
	c := cron.New(
		cron.WithLocation(time.UTC),
		cron.WithLogger(cron.DefaultLogger),
	)
	return &Scheduler{cron: c, db: pool, redis: rdb, engine: eng, email: emailClient, encKey: encKey}
}

func (s *Scheduler) Start() {
	// Job 1 — Generare activități zilnice (00:00 UTC)
	s.cron.AddFunc("0 0 * * *", s.jobGenerateDailyTasks)

	// Job 2 — Calcul scor zilnic (23:50 UTC)
	s.cron.AddFunc("50 23 * * *", s.jobComputeDailyScore)

	// Job 3 — Verificare progres zilnic (23:55 UTC)
	s.cron.AddFunc("55 23 * * *", s.jobCheckDailyProgress)

	// Job 4 — Închidere etape expirate (00:01 UTC)
	s.cron.AddFunc("1 0 * * *", s.jobCloseExpiredSprints)

	// Job 5 — Recalibrare relevanță (la 90 de zile, 02:00 UTC)
	s.cron.AddFunc("0 2 * * 0", s.jobRecalibrateRelevance)

	// Job 11 — Detecție stagnare (23:58 UTC zilnic) — G-5
	s.cron.AddFunc("58 23 * * *", s.jobDetectStagnation)

	// Job 12 — Propunere reactivare obiective PAUSED (00:10 UTC zilnic) — G-7
	s.cron.AddFunc("10 0 * * *", s.jobProposeReactivation)

	// ── Level 4 & 5 jobs ──────────────────────────────────────

	// Job 6 — Detecție evoluție sprint (01:00 UTC zilnic)
	s.cron.AddFunc("0 1 * * *", s.jobDetectEvolutionSprints)

	// Job 7 — Generare ceremonies (01:05 UTC zilnic)
	s.cron.AddFunc("5 1 * * *", s.jobGenerateCeremonies)

	// Job 8 — Progres reactivare obiective (00:05 UTC zilnic)
	s.cron.AddFunc("5 0 * * *", s.jobProgressReactivation)

	// Job 9 — Verificare timeout SRM (orar)
	s.cron.AddFunc("0 * * * *", s.jobCheckSRMTimeouts)

	// Job 10 — Refresh progress overview (orar)
	s.cron.AddFunc("0 * * * *", s.jobRefreshProgressOverview)

	s.cron.Start()
	logger.Info("All jobs scheduled", zap.Int("total_jobs", len(s.cron.Entries())))
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	logger.Info("Scheduler stopped")
}

// ═══════════════════════════════════════════════════════════════
// JOB 1 — Generare activități zilnice (00:00 UTC)
// Rulează la 00:00 UTC — generează sarcinile de azi
// pentru toți userii cu obiective active
// ═══════════════════════════════════════════════════════════════
func (s *Scheduler) jobGenerateDailyTasks() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	logger.Info("Job: GenerateDailyTasks", zap.Time("date", today))

	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT user_id
		FROM global_objectives
		WHERE status = 'ACTIVE'
	`)
	if err != nil {
		logger.Error("GenerateDailyTasks: query users", zap.Error(err))
		return
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			userIDs = append(userIDs, id)
		}
	}

	generated, skipped, failed := 0, 0, 0

	for _, userID := range userIDs {
		// Verifică dacă există deja sarcini pentru azi
		var count int
		s.db.QueryRow(ctx,
			`SELECT COUNT(*) FROM daily_tasks WHERE user_id=$1 AND task_date=$2`,
			userID, today).Scan(&count)
		if count > 0 {
			skipped++
			continue
		}

		// Delegă generarea sarcinilor engine-ului
		tasks, err := s.engine.GenerateDailyTasks(ctx, userID, today)
		if err != nil {
			logger.Warn("Job: GenerateDailyTasks engine error",
				zap.Error(err), zap.String("user", userID.String()))
			failed++
			continue
		}

		generated += len(tasks)
		cache.InvalidateTodayTasks(ctx, s.redis, userID.String(), today.Format("2006-01-02"))
		cache.InvalidateDashboard(ctx, s.redis, userID.String())
	}

	logger.Info("Job: GenerateDailyTasks done",
		zap.Int("generated", generated),
		zap.Int("skipped", skipped),
		zap.Int("failed", failed))
}

// ═══════════════════════════════════════════════════════════════
// JOB 2 — Calcul scor zilnic (23:50 UTC)
// Calculează și stochează scorul zilnic pentru toate obiectivele
// ═══════════════════════════════════════════════════════════════
func (s *Scheduler) jobComputeDailyScore() {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	logger.Info("Job: ComputeDailyScore", zap.Time("date", today))

	rows, err := s.db.Query(ctx, `
		SELECT id, user_id FROM global_objectives WHERE status = 'ACTIVE'
	`)
	if err != nil {
		logger.Error("ComputeDailyScore: query goals", zap.Error(err))
		return
	}
	defer rows.Close()

	type goalRef struct{ id, userID uuid.UUID }
	var goals []goalRef
	for rows.Next() {
		var g goalRef
		rows.Scan(&g.id, &g.userID)
		goals = append(goals, g)
	}

	for _, g := range goals {
		score, grade, err := s.engine.ComputeGoalScore(ctx, g.id, g.userID)
		if err != nil {
			logger.Warn("Job: ComputeGoalScore error",
				zap.Error(err), zap.String("goal", g.id.String()))
			continue
		}
		db.UpsertGoalScore(ctx, s.db, g.id, score, grade)
		if err := db.ComputeGrowthTrajectory(ctx, s.db, g.id, time.Now()); err != nil {
			logger.Warn("[scheduler] trajectory failed",
				zap.Error(err), zap.String("goal", g.id.String()))
		}
	}

	logger.Info("Job: ComputeDailyScore done", zap.Int("goals", len(goals)))
}

// ═══════════════════════════════════════════════════════════════
// JOB 3 — Verificare progres zilnic (23:55 UTC)
// Ajustează intensitatea pentru ziua următoare
// ═══════════════════════════════════════════════════════════════
func (s *Scheduler) jobCheckDailyProgress() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	logger.Info("Job: CheckDailyProgress", zap.Time("date", today))

	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT dt.user_id
		FROM daily_tasks dt
		WHERE dt.task_date = $1
		  AND dt.task_type = 'MAIN'
		  AND NOT EXISTS (
			SELECT 1 FROM daily_tasks dt2
			WHERE dt2.user_id = dt.user_id
			  AND dt2.task_date = $1
			  AND dt2.completed = TRUE
		  )
	`, today)
	if err != nil {
		logger.Error("CheckDailyProgress: query", zap.Error(err))
		return
	}
	defer rows.Close()

	missed := 0
	for rows.Next() {
		var userID uuid.UUID
		rows.Scan(&userID)
		cache.InvalidateDashboard(ctx, s.redis, userID.String())
		missed++
	}

	// Actualizează checkpoints in-progress
	_, err = s.db.Exec(ctx, `
		UPDATE checkpoints cp
		SET status = 'IN_PROGRESS'
		WHERE cp.status = 'UPCOMING'
		  AND cp.sort_order = (
			SELECT MIN(sort_order) FROM checkpoints
			WHERE sprint_id = cp.sprint_id AND status = 'UPCOMING'
		  )
		  AND EXISTS (
			SELECT 1 FROM sprints s
			WHERE s.id = cp.sprint_id AND s.status = 'ACTIVE'
		  )
	`)
	if err != nil {
		logger.Warn("CheckDailyProgress: update checkpoints", zap.Error(err))
	}

	s.autoCompleteCheckpoints(ctx)

	logger.Info("Job: CheckDailyProgress done", zap.Int("missed_users", missed))
}

// ═══════════════════════════════════════════════════════════════
// JOB 4 — Închidere etape expirate (00:01 UTC)
// ═══════════════════════════════════════════════════════════════
func (s *Scheduler) jobCloseExpiredSprints() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	logger.Info("Job: CloseExpiredSprints", zap.Time("date", today))

	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.go_id, s.sprint_number, g.end_date, g.user_id
		FROM sprints s
		JOIN global_objectives g ON g.id = s.go_id
		WHERE s.status = 'ACTIVE'
		  AND s.end_date < $1
	`, today)
	if err != nil {
		logger.Error("CloseExpiredSprints: query", zap.Error(err))
		return
	}
	defer rows.Close()

	type expiredSprint struct {
		id, goalID  uuid.UUID
		userID      uuid.UUID
		number      int
		goalEndDate time.Time
	}
	var expired []expiredSprint
	for rows.Next() {
		var sp expiredSprint
		rows.Scan(&sp.id, &sp.goalID, &sp.number, &sp.goalEndDate, &sp.userID)
		expired = append(expired, sp)
	}

	closed, nextCreated := 0, 0

	for _, sp := range expired {
		// Scorul sprint-ului via engine (opac)
		score, grade, err := s.engine.ComputeSprintScore(ctx, sp.id)
		if err != nil {
			score, grade = 0, "D"
		}
		db.SaveSprintResult(ctx, s.db, sp.id, score, grade)
		db.CloseSprint(ctx, s.db, sp.id)
		closed++

		// Send sprint-complete email — fire-and-forget
		if s.email != nil && sp.userID != uuid.Nil {
			sprintNum := sp.number
			sprintGrade := grade
			goalID := sp.goalID
			userID := sp.userID
			go func() {
				user, err := db.GetUserByID(context.Background(), s.db, userID)
				if err != nil {
					return
				}
				userEmail, _ := crypto.Decrypt(user.EmailEncrypted, s.encKey)
				if userEmail == "" {
					return
				}
				name := ""
				if user.FullName != nil {
					name = *user.FullName
				}
				goal, _ := db.GetGoalByID(context.Background(), s.db, goalID, userID)
				goalName := ""
				if goal != nil {
					goalName = goal.Name
				}
				_ = s.email.SendSprintComplete(context.Background(), userEmail, name, goalName, sprintGrade, sprintNum)
			}()
		}

		nextStart := today
		nextEnd := today.AddDate(0, 0, 30)
		if nextEnd.After(sp.goalEndDate) {
			nextEnd = sp.goalEndDate
		}

		if nextStart.Before(sp.goalEndDate) {
			db.CreateSprint(ctx, s.db, sp.goalID, sp.number+1, nextStart, nextEnd)
			nextCreated++
		} else {
			s.db.Exec(ctx,
				`UPDATE global_objectives SET status='COMPLETED', updated_at=NOW() WHERE id=$1`,
				sp.goalID)
			s.activateWaitingGoal(ctx, sp.goalID)
		}

		if sp.userID != uuid.Nil {
			cache.InvalidateDashboard(ctx, s.redis, sp.userID.String())
		}
	}

	logger.Info("Job: CloseExpiredSprints done",
		zap.Int("closed", closed),
		zap.Int("next_sprints_created", nextCreated))
}

// ═══════════════════════════════════════════════════════════════
// JOB 5 — Recalibrare relevanță (la 90 zile, 02:00 UTC) — G-9
// Actualizează relevanța obiectivelor bazat pe comportamentul ultimelor 90 zile.
// Stochează health_90d și chaos_index în go_metrics.
// ═══════════════════════════════════════════════════════════════
func (s *Scheduler) jobRecalibrateRelevance() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	logger.Info("Job: RecalibrateRelevance")

	// Curăță sesiunile expirate
	res, err := s.db.Exec(ctx,
		`DELETE FROM user_sessions WHERE expires_at < NOW() OR revoked = TRUE`)
	if err == nil {
		logger.Info("Job: cleaned expired sessions",
			zap.Int64("deleted", res.RowsAffected()))
	}

	// Curăță audit log mai vechi de 180 zile
	s.db.Exec(ctx, `DELETE FROM audit_log WHERE created_at < NOW() - INTERVAL '180 days'`)

	// G-9: Calculează scorul de sănătate pe 90 zile + Chaos Index per sprint activ
	rows, err := s.db.Query(ctx, `
		SELECT go_id,
			   COUNT(*) FILTER (WHERE completed) AS done,
			   COUNT(*) AS total
		FROM daily_tasks
		WHERE task_type = 'MAIN'
		  AND task_date > NOW() - INTERVAL '90 days'
		GROUP BY go_id
	`)
	if err == nil {
		defer rows.Close()
		recalibrated := 0
		for rows.Next() {
			var goalID uuid.UUID
			var done, total int
			rows.Scan(&goalID, &done, &total)
			if total == 0 {
				continue
			}

			healthScore := float64(done) / float64(total)
			s.db.Exec(ctx, `
				INSERT INTO go_metrics (go_id, metric_key, metric_value)
				VALUES ($1, 'health_90d', $2)
			`, goalID, healthScore)

			// G-9: also store Chaos Index for the active sprint
			sprint, sprintErr := db.GetCurrentSprint(ctx, s.db, goalID)
			if sprintErr == nil && sprint != nil {
				ci, triggerL2 := s.engine.CheckChaosIndex(ctx, sprint.ID)
				s.db.Exec(ctx, `
					INSERT INTO go_metrics (go_id, metric_key, metric_value)
					VALUES ($1, 'chaos_index', $2)
				`, goalID, ci)

				// G-3: auto-trigger SRM L2 if chaos index >= 0.40 (only if no L2 in last 7 days)
				if triggerL2 {
					s.db.Exec(ctx, `
						INSERT INTO srm_events (id, go_id, srm_level, trigger_reason)
						SELECT gen_random_uuid(), $1, 'L2', 'chaos_index_threshold'
						WHERE NOT EXISTS (
							SELECT 1 FROM srm_events
							WHERE go_id = $1 AND srm_level = 'L2'
							  AND triggered_at > NOW() - INTERVAL '7 days'
							  AND revoked_at IS NULL
						)
					`, goalID)
					logger.Info("Job: SRM L2 triggered by Chaos Index",
						zap.String("goal_id", goalID.String()),
						zap.Float64("chaos_index", ci))
				}
			}

			recalibrated++
		}
		logger.Info("Job: RecalibrateRelevance done",
			zap.Int("recalibrated", recalibrated))
	}
}

// ═══════════════════════════════════════════════════════════════
// JOB 6 — Detect Evolution Sprints (01:00 UTC zilnic)
// Verifică sprints completate ieri și detectează evoluție (C37)
// ═══════════════════════════════════════════════════════════════
func (s *Scheduler) jobDetectEvolutionSprints() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	yesterday := time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	logger.Info("Job: DetectEvolutionSprints", zap.Time("date", yesterday))

	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.go_id
		FROM sprints s
		WHERE s.status = 'COMPLETED'
		  AND DATE(s.updated_at) = $1
	`, yesterday)
	if err != nil {
		logger.Error("DetectEvolutionSprints: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	evolutionCount := 0
	for rows.Next() {
		var sprintID, goalID uuid.UUID
		if err := rows.Scan(&sprintID, &goalID); err != nil {
			continue
		}
		if isEvolution, err := s.engine.MarkEvolutionSprint(ctx, sprintID, goalID); err == nil && isEvolution {
			evolutionCount++
		}
	}

	logger.Info("Job: DetectEvolutionSprints done",
		zap.Int("evolution_sprints", evolutionCount))
}

// ═══════════════════════════════════════════════════════════════
// JOB 7 — Generate Ceremonies (01:05 UTC zilnic)
// Generează ceremonies pentru sprints completate ieri (C38)
// ═══════════════════════════════════════════════════════════════
func (s *Scheduler) jobGenerateCeremonies() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	yesterday := time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	logger.Info("Job: GenerateCeremonies", zap.Time("date", yesterday))

	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.go_id,
		       EXISTS(SELECT 1 FROM evolution_sprints WHERE sprint_id = s.id) AS is_evolution
		FROM sprints s
		WHERE s.status = 'COMPLETED'
		  AND DATE(s.updated_at) = $1
		  AND NOT EXISTS (
			  SELECT 1 FROM completion_ceremonies
			  WHERE sprint_id = s.id
		  )
	`, yesterday)
	if err != nil {
		logger.Error("GenerateCeremonies: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	ceremonyCount := 0
	for rows.Next() {
		var sprintID, goalID uuid.UUID
		var isEvolution bool
		if err := rows.Scan(&sprintID, &goalID, &isEvolution); err != nil {
			continue
		}
		if err := s.engine.GenerateCompletionCeremony(ctx, sprintID, goalID, isEvolution); err == nil {
			ceremonyCount++
		}
	}

	logger.Info("Job: GenerateCeremonies done",
		zap.Int("ceremonies_generated", ceremonyCount))
}

// ═══════════════════════════════════════════════════════════════
// JOB 8 — Progress Reactivation (00:05 UTC zilnic)
// Avansează intensitatea pentru goals în reactivation (C36)
// ═══════════════════════════════════════════════════════════════
func (s *Scheduler) jobProgressReactivation() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	logger.Info("Job: ProgressReactivation")

	rows, err := s.db.Query(ctx, `
		SELECT go_id, current_day, current_intensity
		FROM reactivation_protocols
		WHERE completed_at IS NULL
	`)
	if err != nil {
		logger.Error("ProgressReactivation: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	type reactivationRow struct {
		goalID           uuid.UUID
		currentDay       int
		currentIntensity float64
	}
	var active []reactivationRow
	for rows.Next() {
		var r reactivationRow
		if err := rows.Scan(&r.goalID, &r.currentDay, &r.currentIntensity); err == nil {
			active = append(active, r)
		}
	}

	progressed, completed := 0, 0
	for _, r := range active {
		newDay := r.currentDay + 1
		newIntensity := 0.2 + float64(newDay)*0.1

		if newIntensity >= 1.0 {
			_, err := s.db.Exec(ctx, `
				UPDATE reactivation_protocols
				SET completed_at = NOW(), current_intensity = 1.0, updated_at = NOW()
				WHERE go_id = $1 AND completed_at IS NULL
			`, r.goalID)
			if err == nil {
				completed++
			}
		} else {
			_, err := s.db.Exec(ctx, `
				UPDATE reactivation_protocols
				SET current_day = $1, current_intensity = $2, updated_at = NOW()
				WHERE go_id = $3 AND completed_at IS NULL
			`, newDay, newIntensity, r.goalID)
			if err == nil {
				progressed++
			}
		}
	}

	logger.Info("Job: ProgressReactivation done",
		zap.Int("progressed", progressed),
		zap.Int("completed", completed))
}

// ═══════════════════════════════════════════════════════════════
// JOB 9 — Check SRM Timeouts (orar)
// Verifică SRM L3 neconfirmat mai mult de N ore (C33)
// ═══════════════════════════════════════════════════════════════
func (s *Scheduler) jobCheckSRMTimeouts() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	logger.Info("Job: CheckSRMTimeouts")

	rows, err := s.db.Query(ctx, `
		SELECT go_id,
		       EXTRACT(EPOCH FROM (NOW() - triggered_at)) / 3600 AS hours_since
		FROM srm_events
		WHERE srm_level = 'L3'
		  AND revoked_at IS NULL
	`)
	if err != nil {
		logger.Error("CheckSRMTimeouts: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	timeoutCount := 0
	for rows.Next() {
		var goalID uuid.UUID
		var hoursSince float64
		if err := rows.Scan(&goalID, &hoursSince); err != nil {
			continue
		}

		var fallback string
		switch {
		case hoursSince >= 168: // 7 days
			fallback = "PAUSE"
		case hoursSince >= 72:
			fallback = "L1"
		case hoursSince >= 24:
			fallback = "L2"
		default:
			continue
		}

		logger.Info("SRM L3 timeout — applying fallback",
			zap.String("goal_id", goalID.String()),
			zap.String("fallback", fallback),
			zap.Float64("hours_since", hoursSince))
		// TODO: engine.ApplySRMFallback(ctx, goalID, fallback)
		timeoutCount++
	}

	logger.Info("Job: CheckSRMTimeouts done",
		zap.Int("timeouts_processed", timeoutCount))
}

// ═══════════════════════════════════════════════════════════════
// JOB 10 — Refresh Progress Overview (orar)
// Refresh materialized view pentru analytics (C40)
// ═══════════════════════════════════════════════════════════════
func (s *Scheduler) jobRefreshProgressOverview() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	logger.Info("Job: RefreshProgressOverview")

	if _, err := s.db.Exec(ctx, "SELECT refresh_progress_overview()"); err != nil {
		logger.Error("RefreshProgressOverview: failed", zap.Error(err))
		return
	}

	logger.Info("Job: RefreshProgressOverview done")
}

// ═══════════════════════════════════════════════════════════════
// HELPERS (package-private)
// ═══════════════════════════════════════════════════════════════

func (s *Scheduler) autoCompleteCheckpoints(ctx context.Context) {
	s.db.Exec(ctx, `
		UPDATE checkpoints cp
		SET status = 'COMPLETED', completed_at = NOW()
		WHERE cp.status = 'IN_PROGRESS'
		  AND (
			SELECT CAST(COUNT(*) FILTER (WHERE dt.completed) AS FLOAT)
				/ NULLIF(COUNT(*), 0)
			FROM daily_tasks dt
			JOIN sprints s ON s.id = dt.sprint_id
			WHERE s.id = cp.sprint_id AND dt.task_type = 'MAIN'
		  ) >= 0.80
	`)
}

// ═══════════════════════════════════════════════════════════════
// JOB 11 — Detect Stagnation (23:58 UTC zilnic) — G-5
// Înregistrează stagnare când un obiectiv are >= 5 zile consecutive fără progres.
// ═══════════════════════════════════════════════════════════════
func (s *Scheduler) jobDetectStagnation() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	logger.Info("Job: DetectStagnation")

	rows, err := s.db.Query(ctx, `
		SELECT id, user_id FROM global_objectives WHERE status = 'ACTIVE'
	`)
	if err != nil {
		logger.Error("DetectStagnation: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	type goalRef struct{ id, userID uuid.UUID }
	var goals []goalRef
	for rows.Next() {
		var g goalRef
		if err := rows.Scan(&g.id, &g.userID); err == nil {
			goals = append(goals, g)
		}
	}

	detected := 0
	for _, g := range goals {
		inactiveDays := s.engine.ConsecutiveInactiveDays(ctx, g.id)
		if inactiveDays < 5 {
			continue
		}

		// Record stagnation event (idempotent — unique per day)
		_, err := s.db.Exec(ctx, `
			INSERT INTO stagnation_events (go_id, user_id, inactive_days)
			VALUES ($1, $2, $3)
			ON CONFLICT (go_id, detected_at::date) DO NOTHING
		`, g.id, g.userID, inactiveDays)
		if err == nil {
			detected++
			cache.InvalidateDashboard(ctx, s.redis, g.userID.String())
			// SA-3: auto-trigger SRM L1 if no active SRM event exists for this goal
			existingLevel, _ := db.GetActiveSRMLevel(ctx, s.db, g.id)
			if existingLevel == "" {
				if err := db.InsertSRMEvent(ctx, s.db, g.id, "L1", "stagnation_5days"); err != nil {
					logger.Error("[scheduler] SRM L1 failed", zap.String("goal", g.id.String()), zap.Error(err))
				}
			}
		}
	}

	logger.Info("Job: DetectStagnation done", zap.Int("stagnant_goals", detected))
}

// ═══════════════════════════════════════════════════════════════
// JOB 12 — Propose Reactivation for eligible PAUSED goals (00:10 UTC) — G-7
// Detectează obiective PAUSED cu >= 7 zile stabilitate → propune reactivare.
// ═══════════════════════════════════════════════════════════════
func (s *Scheduler) jobProposeReactivation() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	logger.Info("Job: ProposeReactivation")

	rows, err := s.db.Query(ctx, `
		SELECT id, user_id FROM global_objectives WHERE status = 'PAUSED'
	`)
	if err != nil {
		logger.Error("ProposeReactivation: query failed", zap.Error(err))
		return
	}
	defer rows.Close()

	type goalRef struct{ id, userID uuid.UUID }
	var goals []goalRef
	for rows.Next() {
		var g goalRef
		if err := rows.Scan(&g.id, &g.userID); err == nil {
			goals = append(goals, g)
		}
	}

	proposed := 0
	for _, g := range goals {
		eligible, daysSince, err := s.engine.CheckReactivationEligibility(ctx, g.id)
		if err != nil || !eligible {
			continue
		}

		if err := s.engine.ProposeReactivation(ctx, g.id); err == nil {
			proposed++
			logger.Info("Job: reactivation proposed",
				zap.String("goal_id", g.id.String()),
				zap.Int("days_stable", daysSince))
			cache.InvalidateDashboard(ctx, s.redis, g.userID.String())
		}
	}

	logger.Info("Job: ProposeReactivation done", zap.Int("proposed", proposed))
}

func (s *Scheduler) activateWaitingGoal(ctx context.Context, completedGoalID uuid.UUID) {
	var userID uuid.UUID
	if err := s.db.QueryRow(ctx,
		`SELECT user_id FROM global_objectives WHERE id=$1`, completedGoalID,
	).Scan(&userID); err != nil {
		return
	}

	var activeCount int
	s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM global_objectives WHERE user_id=$1 AND status='ACTIVE'`,
		userID).Scan(&activeCount)

	if activeCount >= 3 {
		return
	}

	var waitingID uuid.UUID
	err := s.db.QueryRow(ctx, `
		SELECT id FROM global_objectives
		WHERE user_id=$1 AND status='WAITING'
		ORDER BY created_at ASC LIMIT 1
	`, userID).Scan(&waitingID)
	if err != nil {
		return
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	s.db.Exec(ctx, `
		UPDATE global_objectives
		SET status='ACTIVE', start_date=$1, updated_at=NOW()
		WHERE id=$2
	`, today, waitingID)

	var goalEndDate time.Time
	s.db.QueryRow(ctx, `SELECT end_date FROM global_objectives WHERE id=$1`, waitingID).Scan(&goalEndDate)
	sprintEnd := today.AddDate(0, 0, 30)
	if sprintEnd.After(goalEndDate) {
		sprintEnd = goalEndDate
	}
	db.CreateSprint(ctx, s.db, waitingID, 1, today, sprintEnd)

	cache.InvalidateDashboard(ctx, s.redis, userID.String())
	logger.Info("Job: activated waiting goal",
		zap.String("goal_id", waitingID.String()),
		zap.String("user_id", userID.String()))
}
