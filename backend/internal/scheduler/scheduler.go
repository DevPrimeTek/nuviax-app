package scheduler

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/devprimetek/nuviax-app/internal/cache"
	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/models"
	"github.com/devprimetek/nuviax-app/pkg/logger"
)

// Scheduler runs 5 background jobs that power the framework engine
type Scheduler struct {
	cron  *cron.Cron
	db    *pgxpool.Pool
	redis *redis.Client
}

func New(pool *pgxpool.Pool, rdb *redis.Client) *Scheduler {
	c := cron.New(
		cron.WithLocation(time.UTC),
		cron.WithLogger(cron.VerbosePrintfLogger(nil)),
	)
	return &Scheduler{cron: c, db: pool, redis: rdb}
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
// JOB 1 — Generare activități zilnice
// Rulează la 00:00 UTC — generează sarcinile de azi
// pentru toți userii cu obiective active
// ═══════════════════════════════════════════════════════════════
func (s *Scheduler) jobGenerateDailyTasks() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	today := time.Now().UTC().Truncate(24 * time.Hour)
	logger.Info("Job: GenerateDailyTasks", zap.Time("date", today))

	// Obține toți userii activi (au cel puțin un obiectiv ACTIVE)
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

		// Generează pentru fiecare obiectiv activ al userului
		goals, err := db.GetGoalsByUser(ctx, s.db, userID)
		if err != nil {
			failed++
			continue
		}

		for _, goal := range goals {
			if goal.Status != models.GoalActive {
				continue
			}

			// Verifică dacă e pauză activă
			adjs, _ := db.GetActiveAdjustments(ctx, s.db, goal.ID)
			if isOnPause(adjs, today) {
				logger.Info("Job: user on pause, skipping task gen",
					zap.String("user", userID.String()),
					zap.String("goal", goal.ID.String()))
				continue
			}

			sprint, err := db.GetCurrentSprint(ctx, s.db, goal.ID)
			if err != nil {
				continue
			}

			intensity := computeIntensityFromAdjs(adjs)
			taskCount := taskCountFromIntensity(intensity)

			checkpoints, _ := db.GetSprintCheckpoints(ctx, s.db, sprint.ID)
			activeCP := findActiveCheckpoint(checkpoints)
			if activeCP == nil {
				continue
			}

			texts := generateTaskTexts(goal, *activeCP, taskCount)
			for i, text := range texts {
				_, err := db.CreateTask(ctx, s.db,
					sprint.ID, goal.ID, userID, today,
					text, models.TaskMain, i)
				if err != nil {
					logger.Warn("Job: create task failed",
						zap.Error(err), zap.String("user", userID.String()))
				}
			}
			generated++
		}

		// Invalide cache
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
		score := computeGoalScore(ctx, s.db, g.id)
		grade := gradeFromScore(score)
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

	// Găsește userii care nu au completat nicio sarcină azi
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
		// Invalidează cache-ul pentru ziua următoare
		// Sarcinile de mâine vor fi regenerate la 00:00
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

	// Marchează ca COMPLETED checkpoints care au toate task-urile bifate
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

	// Găsește etapele care s-au terminat ieri
	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.go_id, s.sprint_number, g.end_date
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
		id, goalID    uuid.UUID
		number        int
		goalEndDate   time.Time
	}
	var expired []expiredSprint
	for rows.Next() {
		var sp expiredSprint
		rows.Scan(&sp.id, &sp.goalID, &sp.number, &sp.goalEndDate)
		expired = append(expired, sp)
	}

	closed, nextCreated := 0, 0

	for _, sp := range expired {
		// Calculează scorul final
		score := computeGoalScore(ctx, s.db, sp.goalID)
		grade := gradeFromScore(score)
		db.SaveSprintResult(ctx, s.db, sp.id, score, grade)
		db.CloseSprint(ctx, s.db, sp.id)
		closed++

		// Creează etapa următoare dacă obiectivul mai are timp
		nextStart := today
		nextEnd := today.AddDate(0, 0, 30)
		if nextEnd.After(sp.goalEndDate) {
			nextEnd = sp.goalEndDate
		}

		if nextStart.Before(sp.goalEndDate) {
			db.CreateSprint(ctx, s.db, sp.goalID, sp.number+1, nextStart, nextEnd)
			nextCreated++
		} else {
			// Obiectivul s-a terminat
			s.db.Exec(ctx,
				`UPDATE global_objectives SET status='COMPLETED', updated_at=NOW() WHERE id=$1`,
				sp.goalID)
			// Verifică dacă există obiective în waiting list de activat
			s.activateWaitingGoal(ctx, sp.goalID)
		}

		// Invalidează cache
		var userID uuid.UUID
		s.db.QueryRow(ctx, `SELECT user_id FROM global_objectives WHERE id=$1`, sp.goalID).Scan(&userID)
		if userID != uuid.Nil {
			cache.InvalidateDashboard(ctx, s.redis, userID.String())
		}
	}

	logger.Info("Job: CloseExpiredSprints done",
		zap.Int("closed", closed),
		zap.Int("next_sprints_created", nextCreated))
}

// ═══════════════════════════════════════════════════════════════
// JOB 5 — Recalibrare relevanță (la 90 zile)
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

	// Actualizează metadata obiective (opac)
	// Calculează un scor de "sănătate" bazat pe ultimele 90 zile
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
				// Stochează ca metrică opacă
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

func isOnPause(adjs []models.ContextAdjustment, today time.Time) bool {
	for _, a := range adjs {
		if a.AdjType == models.AdjPause {
			if !today.Before(a.StartDate) {
				if a.EndDate == nil || !today.After(*a.EndDate) {
					return true
				}
			}
		}
	}
	return false
}

func computeIntensityFromAdjs(adjs []models.ContextAdjustment) float64 {
	base := 1.0
	for _, a := range adjs {
		switch a.AdjType {
		case models.AdjEnergyLow:
			base = 0.6
		case models.AdjEnergyHigh:
			base = 1.2
		}
	}
	return base
}

func taskCountFromIntensity(intensity float64) int {
	if intensity >= 1.2 {
		return 3
	} else if intensity >= 1.0 {
		return 2
	}
	return 1
}

func findActiveCheckpoint(cps []models.Checkpoint) *models.Checkpoint {
	for i := range cps {
		if cps[i].Status == models.CheckpointInProgress ||
			cps[i].Status == models.CheckpointUpcoming {
			return &cps[i]
		}
	}
	return nil
}

func generateTaskTexts(goal models.Goal, cp models.Checkpoint, count int) []string {
	base := cp.Name
	all := []string{
		"Lucrează la: " + base,
		"Avansează cu: " + base,
		"Finalizează o parte din: " + base,
	}
	if count > len(all) {
		count = len(all)
	}
	return all[:count]
}

func computeGoalScore(ctx context.Context, pool *pgxpool.Pool, goalID uuid.UUID) float64 {
	var total, completed int
	pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE dt.completed = TRUE)
		FROM daily_tasks dt
		JOIN sprints s ON s.id = dt.sprint_id
		WHERE s.go_id = $1 AND dt.task_type = 'MAIN'
		  AND dt.task_date <= CURRENT_DATE
	`, goalID).Scan(&total, &completed)

	if total == 0 {
		return 0
	}

	completionRate := float64(completed) / float64(total)

	// Consistență — distribuție zilnică
	var activeDays, totalDays int
	pool.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT task_date) FILTER (WHERE completed = TRUE),
			COUNT(DISTINCT task_date)
		FROM daily_tasks
		WHERE go_id = $1 AND task_type = 'MAIN' AND task_date <= CURRENT_DATE
	`, goalID).Scan(&activeDays, &totalDays)

	consistency := 0.0
	if totalDays > 0 {
		consistency = float64(activeDays) / float64(totalDays)
	}

	// Scor compozit (opac)
	score := completionRate*0.65 + consistency*0.35
	return math.Min(math.Max(score, 0), 1)
}

func gradeFromScore(score float64) string {
	switch {
	case score >= 0.90:
		return "A+"
	case score >= 0.80:
		return "A"
	case score >= 0.70:
		return "B"
	case score >= 0.60:
		return "C"
	default:
		return "D"
	}
}

func (s *Scheduler) autoCompleteCheckpoints(ctx context.Context) {
	// Dacă 80%+ din sarcinile unui checkpoint sunt bifate → marchează COMPLETED
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
	// Obține user-ul obiectivului completat
	var userID uuid.UUID
	if err := s.db.QueryRow(ctx,
		`SELECT user_id FROM global_objectives WHERE id=$1`, completedGoalID,
	).Scan(&userID); err != nil {
		return
	}

	// Verifică dacă are slot liber (< 3 active)
	var activeCount int
	s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM global_objectives WHERE user_id=$1 AND status='ACTIVE'`,
		userID).Scan(&activeCount)

	if activeCount >= 3 {
		return
	}

	// Activează primul obiectiv din waiting list (cel mai vechi)
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

	// Creează Sprint 1 pentru noul obiectiv
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
