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
	"github.com/devprimetek/nuviax-app/internal/engine"
	"github.com/devprimetek/nuviax-app/pkg/logger"
)

// Scheduler runs background jobs powered by the NUViaX Framework Engine
type Scheduler struct {
	cron   *cron.Cron
	db     *pgxpool.Pool
	redis  *redis.Client
	engine *engine.Engine
}

func New(pool *pgxpool.Pool, rdb *redis.Client, eng *engine.Engine) *Scheduler {
	c := cron.New(
		cron.WithLocation(time.UTC),
		cron.WithLogger(cron.DefaultLogger),
	)
	return &Scheduler{cron: c, db: pool, redis: rdb, engine: eng}
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
	s.cron.AddFunc("0 2 */90 * *", s.jobRecalibrateRelevance)

	s.cron.Start()
	logger.Info("Scheduler started", zap.Int("jobs", 5))
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
		id, goalID uuid.UUID
		userID     uuid.UUID
		number     int
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
// JOB 5 — Recalibrare relevanță (la 90 zile, 02:00 UTC)
// Actualizează relevanța obiectivelor bazat pe comportament
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

	// Calculează scorul de sănătate pe 90 zile (opac)
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
			if total > 0 {
				healthScore := float64(done) / float64(total)
				s.db.Exec(ctx, `
					INSERT INTO go_metrics (go_id, metric_key, metric_value)
					VALUES ($1, 'health_90d', $2)
				`, goalID, healthScore)
				recalibrated++
			}
		}
		logger.Info("Job: RecalibrateRelevance done",
			zap.Int("recalibrated", recalibrated))
	}
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
