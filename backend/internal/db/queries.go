package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devprimetek/nuviax-app/internal/models"
)

var ErrNotFound = errors.New("not found")

// ═══════════════════════════════════════════════════════════════
// USERS
// ═══════════════════════════════════════════════════════════════

func CreateUser(ctx context.Context, pool *pgxpool.Pool,
	emailEncrypted, emailHash, passwordHash, salt, locale string,
	fullName *string,
) (*models.User, error) {
	u := &models.User{}
	err := pool.QueryRow(ctx, `
		INSERT INTO users
			(email_encrypted, email_hash, password_hash, salt, full_name, locale)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, email_encrypted, email_hash, password_hash, salt,
		          full_name, locale, mfa_enabled, is_active, is_admin, created_at, updated_at
	`, emailEncrypted, emailHash, passwordHash, salt, fullName, locale).Scan(
		&u.ID, &u.EmailEncrypted, &u.EmailHash, &u.PasswordHash, &u.Salt,
		&u.FullName, &u.Locale, &u.MFAEnabled, &u.IsActive, &u.IsAdmin,
		&u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

func GetUserByEmailHash(ctx context.Context, pool *pgxpool.Pool, hash string) (*models.User, error) {
	u := &models.User{}
	err := pool.QueryRow(ctx, `
		SELECT id, email_encrypted, email_hash, password_hash, salt,
		       full_name, locale, mfa_secret, mfa_enabled, is_active, is_admin, created_at, updated_at
		FROM users WHERE email_hash = $1 AND is_active = TRUE
	`, hash).Scan(
		&u.ID, &u.EmailEncrypted, &u.EmailHash, &u.PasswordHash, &u.Salt,
		&u.FullName, &u.Locale, &u.MFASecret, &u.MFAEnabled, &u.IsActive, &u.IsAdmin,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func GetUserByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*models.User, error) {
	u := &models.User{}
	err := pool.QueryRow(ctx, `
		SELECT id, email_encrypted, email_hash, password_hash, salt,
		       full_name, locale, COALESCE(theme, 'dark'), avatar_url, mfa_secret, mfa_enabled, is_active, is_admin, created_at, updated_at
		FROM users WHERE id = $1 AND is_active = TRUE
	`, id).Scan(
		&u.ID, &u.EmailEncrypted, &u.EmailHash, &u.PasswordHash, &u.Salt,
		&u.FullName, &u.Locale, &u.Theme, &u.AvatarURL, &u.MFASecret, &u.MFAEnabled, &u.IsActive, &u.IsAdmin,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func UpdateUserTheme(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, theme string) error {
	_, err := pool.Exec(ctx,
		`UPDATE users SET theme=$1, updated_at=NOW() WHERE id=$2`, theme, userID)
	return err
}

func UpdateUserMFA(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, secret string, enabled bool) error {
	_, err := pool.Exec(ctx,
		`UPDATE users SET mfa_secret=$1, mfa_enabled=$2, updated_at=NOW() WHERE id=$3`,
		secret, enabled, userID)
	return err
}

func UpdateUserLocale(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, locale string) error {
	_, err := pool.Exec(ctx,
		`UPDATE users SET locale=$1, updated_at=NOW() WHERE id=$2`, locale, userID)
	return err
}

func UpdateUserPassword(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, newPasswordHash string) error {
	_, err := pool.Exec(ctx,
		`UPDATE users SET password_hash=$1, updated_at=NOW() WHERE id=$2`, newPasswordHash, userID)
	return err
}

func UpdateUserAvatar(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, avatarURL string) error {
	_, err := pool.Exec(ctx,
		`UPDATE users SET avatar_url=$1, updated_at=NOW() WHERE id=$2`, avatarURL, userID)
	return err
}

// ═══════════════════════════════════════════════════════════════
// SESSIONS
// ═══════════════════════════════════════════════════════════════

func CreateSession(ctx context.Context, pool *pgxpool.Pool,
	userID uuid.UUID, tokenHash string,
	deviceFP, ipSubnet, uaHash *string, expiresAt time.Time,
) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO user_sessions
			(user_id, token_hash, device_fp, ip_subnet, user_agent_hash, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, userID, tokenHash, deviceFP, ipSubnet, uaHash, expiresAt)
	return err
}

func GetSession(ctx context.Context, pool *pgxpool.Pool, tokenHash string) (*models.UserSession, error) {
	s := &models.UserSession{}
	err := pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked
		FROM user_sessions
		WHERE token_hash=$1 AND revoked=FALSE AND expires_at > NOW()
	`, tokenHash).Scan(&s.ID, &s.UserID, &s.TokenHash, &s.ExpiresAt, &s.Revoked)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return s, err
}

func RevokeSession(ctx context.Context, pool *pgxpool.Pool, tokenHash string) error {
	_, err := pool.Exec(ctx,
		`UPDATE user_sessions SET revoked=TRUE WHERE token_hash=$1`, tokenHash)
	return err
}

func RevokeSessionByID(ctx context.Context, pool *pgxpool.Pool, sessionID, userID uuid.UUID) error {
	_, err := pool.Exec(ctx,
		`UPDATE user_sessions SET revoked=TRUE WHERE id=$1 AND user_id=$2`, sessionID, userID)
	return err
}

func GetUserSessions(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]models.UserSession, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, user_id, device_fp, ip_subnet, user_agent_hash, expires_at, created_at
		FROM user_sessions
		WHERE user_id=$1 AND revoked=FALSE AND expires_at > NOW()
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []models.UserSession
	for rows.Next() {
		s := models.UserSession{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.DeviceFP, &s.IPSubnet, &s.UserAgentHash, &s.ExpiresAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// ═══════════════════════════════════════════════════════════════
// GOALS
// ═══════════════════════════════════════════════════════════════

func CountActiveGoals(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (int, error) {
	var count int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM global_objectives WHERE user_id=$1 AND status='ACTIVE'`, userID).Scan(&count)
	return count, err
}

func CreateGoal(ctx context.Context, pool *pgxpool.Pool,
	userID uuid.UUID, name string, desc *string,
	status models.GoalStatus, startDate, endDate time.Time,
) (*models.Goal, error) {
	g := &models.Goal{}
	err := pool.QueryRow(ctx, `
		INSERT INTO global_objectives
			(user_id, name, description, status, start_date, end_date)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, user_id, name, description, status, start_date, end_date, dominant_behavior_model, created_at, updated_at
	`, userID, name, desc, status, startDate, endDate).Scan(
		&g.ID, &g.UserID, &g.Name, &g.Description, &g.Status,
		&g.StartDate, &g.EndDate, &g.DominantBehaviorModel, &g.CreatedAt, &g.UpdatedAt,
	)
	return g, err
}

func GetGoalByID(ctx context.Context, pool *pgxpool.Pool, goalID, userID uuid.UUID) (*models.Goal, error) {
	g := &models.Goal{}
	err := pool.QueryRow(ctx, `
		SELECT id, user_id, name, description, status, start_date, end_date, dominant_behavior_model, created_at, updated_at
		FROM global_objectives WHERE id=$1 AND user_id=$2
	`, goalID, userID).Scan(
		&g.ID, &g.UserID, &g.Name, &g.Description, &g.Status,
		&g.StartDate, &g.EndDate, &g.DominantBehaviorModel, &g.CreatedAt, &g.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return g, err
}

func GetGoalsByUser(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]models.Goal, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, user_id, name, description, status, start_date, end_date, dominant_behavior_model, created_at, updated_at
		FROM global_objectives
		WHERE user_id=$1 AND status != 'ARCHIVED'
		ORDER BY
			CASE status WHEN 'ACTIVE' THEN 0 WHEN 'WAITING' THEN 1 ELSE 2 END,
			created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var goals []models.Goal
	for rows.Next() {
		g := models.Goal{}
		if err := rows.Scan(&g.ID, &g.UserID, &g.Name, &g.Description, &g.Status,
			&g.StartDate, &g.EndDate, &g.DominantBehaviorModel, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}
	return goals, nil
}

func UpdateGoalStatus(ctx context.Context, pool *pgxpool.Pool, goalID, userID uuid.UUID, status models.GoalStatus) error {
	_, err := pool.Exec(ctx,
		`UPDATE global_objectives SET status=$1, updated_at=NOW() WHERE id=$2 AND user_id=$3`,
		status, goalID, userID)
	return err
}

func SetGoalBehaviorModel(ctx context.Context, pool *pgxpool.Pool, goalID, userID uuid.UUID, behaviorModel *string) error {
	_, err := pool.Exec(ctx,
		`UPDATE global_objectives SET dominant_behavior_model=$1, updated_at=NOW() WHERE id=$2 AND user_id=$3`,
		behaviorModel, goalID, userID)
	return err
}

func ArchiveGoal(ctx context.Context, pool *pgxpool.Pool, goalID, userID uuid.UUID) error {
	return UpdateGoalStatus(ctx, pool, goalID, userID, models.GoalArchived)
}

// ═══════════════════════════════════════════════════════════════
// SPRINTS
// ═══════════════════════════════════════════════════════════════

func CreateSprint(ctx context.Context, pool *pgxpool.Pool,
	goalID uuid.UUID, number int, startDate, endDate time.Time,
) (*models.Sprint, error) {
	s := &models.Sprint{}
	err := pool.QueryRow(ctx, `
		INSERT INTO sprints (go_id, sprint_number, start_date, end_date, status)
		VALUES ($1,$2,$3,$4,'ACTIVE')
		RETURNING id, go_id, sprint_number, start_date, end_date, status, created_at
	`, goalID, number, startDate, endDate).Scan(
		&s.ID, &s.GoalID, &s.SprintNumber, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt,
	)
	return s, err
}

func GetSprintByID(ctx context.Context, pool *pgxpool.Pool, sprintID uuid.UUID) (*models.Sprint, error) {
	s := &models.Sprint{}
	err := pool.QueryRow(ctx, `
		SELECT id, go_id, sprint_number, start_date, end_date, status, created_at
		FROM sprints WHERE id = $1
	`, sprintID).Scan(
		&s.ID, &s.GoalID, &s.SprintNumber, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return s, err
}

func GetCurrentSprint(ctx context.Context, pool *pgxpool.Pool, goalID uuid.UUID) (*models.Sprint, error) {
	s := &models.Sprint{}
	err := pool.QueryRow(ctx, `
		SELECT id, go_id, sprint_number, start_date, end_date, status, created_at
		FROM sprints
		WHERE go_id=$1 AND status='ACTIVE'
		ORDER BY sprint_number DESC LIMIT 1
	`, goalID).Scan(
		&s.ID, &s.GoalID, &s.SprintNumber, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return s, err
}

func GetSprintHistory(ctx context.Context, pool *pgxpool.Pool, goalID uuid.UUID) ([]models.Sprint, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, go_id, sprint_number, start_date, end_date, status, created_at
		FROM sprints WHERE go_id=$1 ORDER BY sprint_number ASC
	`, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sprints []models.Sprint
	for rows.Next() {
		s := models.Sprint{}
		if err := rows.Scan(&s.ID, &s.GoalID, &s.SprintNumber, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		sprints = append(sprints, s)
	}
	return sprints, nil
}

func CloseSprint(ctx context.Context, pool *pgxpool.Pool, sprintID uuid.UUID) error {
	_, err := pool.Exec(ctx,
		`UPDATE sprints SET status='COMPLETED' WHERE id=$1`, sprintID)
	return err
}

// ═══════════════════════════════════════════════════════════════
// DAILY TASKS
// ═══════════════════════════════════════════════════════════════

func GetTodayTasks(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, date time.Time) ([]models.DailyTask, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, sprint_id, go_id, user_id, task_date,
		       task_text, task_type, sort_order, completed, completed_at, created_at
		FROM daily_tasks
		WHERE user_id=$1 AND task_date=$2
		ORDER BY task_type ASC, sort_order ASC
	`, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []models.DailyTask
	for rows.Next() {
		t := models.DailyTask{}
		if err := rows.Scan(
			&t.ID, &t.SprintID, &t.GoalID, &t.UserID, &t.TaskDate,
			&t.TaskText, &t.TaskType, &t.SortOrder, &t.Completed, &t.CompletedAt, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func CreateTask(ctx context.Context, pool *pgxpool.Pool,
	sprintID, goalID, userID uuid.UUID,
	taskDate time.Time, text string, taskType models.TaskType, order int,
) (*models.DailyTask, error) {
	t := &models.DailyTask{}
	err := pool.QueryRow(ctx, `
		INSERT INTO daily_tasks
			(sprint_id, go_id, user_id, task_date, task_text, task_type, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, sprint_id, go_id, user_id, task_date,
		          task_text, task_type, sort_order, completed, completed_at, created_at
	`, sprintID, goalID, userID, taskDate, text, taskType, order).Scan(
		&t.ID, &t.SprintID, &t.GoalID, &t.UserID, &t.TaskDate,
		&t.TaskText, &t.TaskType, &t.SortOrder, &t.Completed, &t.CompletedAt, &t.CreatedAt,
	)
	return t, err
}

func CountPersonalTasksToday(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, date time.Time) (int, error) {
	var count int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM daily_tasks
		 WHERE user_id=$1 AND task_date=$2 AND task_type='PERSONAL'`, userID, date).Scan(&count)
	return count, err
}

func CompleteTask(ctx context.Context, pool *pgxpool.Pool, taskID, userID uuid.UUID) error {
	tag, err := pool.Exec(ctx, `
		UPDATE daily_tasks
		SET completed=TRUE, completed_at=NOW()
		WHERE id=$1 AND user_id=$2 AND completed=FALSE
	`, taskID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func GetTaskByID(ctx context.Context, pool *pgxpool.Pool, taskID, userID uuid.UUID) (*models.DailyTask, error) {
	t := &models.DailyTask{}
	err := pool.QueryRow(ctx, `
		SELECT id, sprint_id, go_id, user_id, task_date,
		       task_text, task_type, sort_order, completed, completed_at, created_at
		FROM daily_tasks WHERE id=$1 AND user_id=$2
	`, taskID, userID).Scan(
		&t.ID, &t.SprintID, &t.GoalID, &t.UserID, &t.TaskDate,
		&t.TaskText, &t.TaskType, &t.SortOrder, &t.Completed, &t.CompletedAt, &t.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

// ═══════════════════════════════════════════════════════════════
// CHECKPOINTS
// ═══════════════════════════════════════════════════════════════

func GetSprintCheckpoints(ctx context.Context, pool *pgxpool.Pool, sprintID uuid.UUID) ([]models.Checkpoint, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, sprint_id, name, description, sort_order, status, progress_pct, completed_at
		FROM checkpoints WHERE sprint_id=$1 ORDER BY sort_order ASC
	`, sprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cps []models.Checkpoint
	for rows.Next() {
		c := models.Checkpoint{}
		if err := rows.Scan(&c.ID, &c.SprintID, &c.Name, &c.Description,
			&c.SortOrder, &c.Status, &c.ProgressPct, &c.CompletedAt); err != nil {
			return nil, err
		}
		cps = append(cps, c)
	}
	return cps, nil
}

func CreateCheckpoint(ctx context.Context, pool *pgxpool.Pool,
	sprintID uuid.UUID, name string, desc *string, order int,
) (*models.Checkpoint, error) {
	c := &models.Checkpoint{}
	err := pool.QueryRow(ctx, `
		INSERT INTO checkpoints (sprint_id, name, description, sort_order)
		VALUES ($1,$2,$3,$4)
		RETURNING id, sprint_id, name, description, sort_order, status, progress_pct, completed_at
	`, sprintID, name, desc, order).Scan(
		&c.ID, &c.SprintID, &c.Name, &c.Description, &c.SortOrder, &c.Status, &c.ProgressPct, &c.CompletedAt,
	)
	return c, err
}

// ═══════════════════════════════════════════════════════════════
// SCORES
// ═══════════════════════════════════════════════════════════════

func SaveSprintResult(ctx context.Context, pool *pgxpool.Pool,
	sprintID uuid.UUID, score float64, grade string,
) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO sprint_results (sprint_id, score_value, grade)
		VALUES ($1,$2,$3)
		ON CONFLICT (sprint_id) DO UPDATE
		  SET score_value=$2, grade=$3, computed_at=NOW()
	`, sprintID, score, grade)
	return err
}

func GetSprintResults(ctx context.Context, pool *pgxpool.Pool, goalID uuid.UUID) ([]models.SprintResult, error) {
	rows, err := pool.Query(ctx, `
		SELECT sr.id, sr.sprint_id, sr.score_value, sr.grade, sr.computed_at
		FROM sprint_results sr
		JOIN sprints s ON s.id = sr.sprint_id
		WHERE s.go_id = $1
		ORDER BY s.sprint_number ASC
	`, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []models.SprintResult
	for rows.Next() {
		r := models.SprintResult{}
		if err := rows.Scan(&r.ID, &r.SprintID, &r.ScoreValue, &r.Grade, &r.ComputedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

func UpsertGoalScore(ctx context.Context, pool *pgxpool.Pool,
	goalID uuid.UUID, score float64, grade string,
) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO go_scores (go_id, score_value, grade)
		VALUES ($1,$2,$3)
	`, goalID, score, grade)
	return err
}

func ComputeGrowthTrajectory(ctx context.Context, pool *pgxpool.Pool, goalID uuid.UUID, date time.Time) error {
	_, err := pool.Exec(ctx,
		"SELECT fn_compute_growth_trajectory($1, $2)",
		goalID, date.UTC().Truncate(24*time.Hour),
	)
	return err
}

func GetLatestGoalScore(ctx context.Context, pool *pgxpool.Pool, goalID uuid.UUID) (*models.GoalScore, error) {
	gs := &models.GoalScore{}
	err := pool.QueryRow(ctx, `
		SELECT id, go_id, score_value, grade, computed_at
		FROM go_scores WHERE go_id=$1
		ORDER BY computed_at DESC LIMIT 1
	`, goalID).Scan(&gs.ID, &gs.GoalID, &gs.ScoreValue, &gs.Grade, &gs.ComputedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return gs, err
}

// ═══════════════════════════════════════════════════════════════
// CONTEXT ADJUSTMENTS
// ═══════════════════════════════════════════════════════════════

func CreateContextAdjustment(ctx context.Context, pool *pgxpool.Pool,
	goalID, userID uuid.UUID, adjType models.AdjType,
	startDate time.Time, endDate *time.Time, note *string,
) (*models.ContextAdjustment, error) {
	return CreateRetroactivePause(ctx, pool, goalID, userID, adjType, startDate, endDate, note, false)
}

// CreateRetroactivePause — GAP #14 fix: supports pause registration with
// a retroactive start date (up to 48h in the past).
// When retroactive=true, the retroactive column is set so the engine
// knows NOT to penalize the missed days.
func CreateRetroactivePause(ctx context.Context, pool *pgxpool.Pool,
	goalID, userID uuid.UUID, adjType models.AdjType,
	startDate time.Time, endDate *time.Time, note *string,
	retroactive bool,
) (*models.ContextAdjustment, error) {
	ca := &models.ContextAdjustment{}
	err := pool.QueryRow(ctx, `
		INSERT INTO context_adjustments
			(go_id, user_id, adj_type, start_date, end_date, note, retroactive)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, go_id, user_id, adj_type, start_date, end_date, note, created_at
	`, goalID, userID, adjType, startDate, endDate, note, retroactive).Scan(
		&ca.ID, &ca.GoalID, &ca.UserID, &ca.AdjType,
		&ca.StartDate, &ca.EndDate, &ca.Note, &ca.CreatedAt,
	)
	return ca, err
}

func GetActiveAdjustments(ctx context.Context, pool *pgxpool.Pool, goalID uuid.UUID) ([]models.ContextAdjustment, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	rows, err := pool.Query(ctx, `
		SELECT id, go_id, user_id, adj_type, start_date, end_date, note, created_at
		FROM context_adjustments
		WHERE go_id=$1
		  AND start_date <= $2
		  AND (end_date IS NULL OR end_date >= $2)
		ORDER BY created_at DESC
	`, goalID, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var adjs []models.ContextAdjustment
	for rows.Next() {
		ca := models.ContextAdjustment{}
		if err := rows.Scan(&ca.ID, &ca.GoalID, &ca.UserID, &ca.AdjType,
			&ca.StartDate, &ca.EndDate, &ca.Note, &ca.CreatedAt); err != nil {
			return nil, err
		}
		adjs = append(adjs, ca)
	}
	return adjs, nil
}

// ═══════════════════════════════════════════════════════════════
// RECAP (B-8 fix)
// ═══════════════════════════════════════════════════════════════

// SprintRecapData holds the data needed for the recap page.
type SprintRecapData struct {
	SprintID       uuid.UUID
	GoalID         uuid.UUID
	SprintName     string
	GoalName       string
	SprintNumber   int
	Score          float64
	Grade          string
	DaysActive     int
	DaysTotal      int
	MRRDelta       int
}

// GetLastCompletedSprintRecap finds the most recently completed sprint for a user
// across all their goals, and returns the data needed for the recap page.
func GetLastCompletedSprintRecap(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (*SprintRecapData, error) {
	r := &SprintRecapData{}
	err := pool.QueryRow(ctx, `
		SELECT
			s.id, s.go_id, s.sprint_number,
			g.name,
			COALESCE(sr.score_value, 0),
			COALESCE(sr.grade, 'C'),
			CAST(EXTRACT(EPOCH FROM (s.end_date - s.start_date)) / 86400 AS INT),
			(
				SELECT COUNT(DISTINCT task_date)
				FROM daily_tasks
				WHERE sprint_id = s.id AND completed = TRUE
			)
		FROM sprints s
		JOIN global_objectives g ON g.id = s.go_id
		LEFT JOIN sprint_results sr ON sr.sprint_id = s.id
		WHERE g.user_id = $1 AND s.status = 'COMPLETED'
		ORDER BY s.end_date DESC
		LIMIT 1
	`, userID).Scan(
		&r.SprintID, &r.GoalID, &r.SprintNumber,
		&r.GoalName,
		&r.Score, &r.Grade,
		&r.DaysTotal, &r.DaysActive,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	r.SprintName = fmt.Sprintf("Etapa %d", r.SprintNumber)
	return r, nil
}

// GetLastCompletedSprintForGoal finds the most recently completed sprint for a specific goal.
func GetLastCompletedSprintForGoal(ctx context.Context, pool *pgxpool.Pool, goalID, userID uuid.UUID) (*models.Sprint, error) {
	s := &models.Sprint{}
	err := pool.QueryRow(ctx, `
		SELECT s.id, s.go_id, s.sprint_number, s.start_date, s.end_date, s.status, s.created_at
		FROM sprints s
		JOIN global_objectives g ON g.id = s.go_id
		WHERE s.go_id = $1 AND g.user_id = $2 AND s.status = 'COMPLETED'
		ORDER BY s.sprint_number DESC
		LIMIT 1
	`, goalID, userID).Scan(
		&s.ID, &s.GoalID, &s.SprintNumber, &s.StartDate, &s.EndDate, &s.Status, &s.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return s, err
}

// ═══════════════════════════════════════════════════════════════
// REFLECTIONS
// ═══════════════════════════════════════════════════════════════

func SaveReflection(ctx context.Context, pool *pgxpool.Pool,
	sprintID, userID uuid.UUID,
	q1, q2 *string, energyLevel *int,
) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO sprint_reflections (sprint_id, user_id, q1_answer, q2_answer, energy_level)
		VALUES ($1,$2,$3,$4,$5)
	`, sprintID, userID, q1, q2, energyLevel)
	return err
}

// ═══════════════════════════════════════════════════════════════
// AUDIT LOG
// ═══════════════════════════════════════════════════════════════

func WriteAudit(ctx context.Context, pool *pgxpool.Pool,
	userID *uuid.UUID, action, ipHash, uaHash string,
) {
	// Fire and forget — nu blocăm request-ul
	go func() {
		_, _ = pool.Exec(context.Background(), `
			INSERT INTO audit_log (user_id, action, ip_hash, ua_hash)
			VALUES ($1,$2,$3,$4)
		`, userID, action, ipHash, uaHash)
	}()
}

// ═══════════════════════════════════════════════════════════════
// ADMIN QUERIES
// ═══════════════════════════════════════════════════════════════

// GetPlatformStats returns aggregated platform statistics for the admin panel.
func GetPlatformStats(ctx context.Context, pool *pgxpool.Pool) (*models.PlatformStats, error) {
	s := &models.PlatformStats{}
	err := pool.QueryRow(ctx, `
		SELECT
			total_users, admin_users, new_users_7d, new_users_30d,
			active_goals, completed_goals, paused_goals, total_goals,
			active_sprints, completed_sprints,
			tasks_today, tasks_completed_today,
			srm_events_30d, srm_l3_events_30d,
			regression_events_30d,
			ceremonies_30d, badges_awarded_30d,
			computed_at
		FROM v_admin_platform_stats
	`).Scan(
		&s.TotalUsers, &s.AdminUsers, &s.NewUsers7d, &s.NewUsers30d,
		&s.ActiveGoals, &s.CompletedGoals, &s.PausedGoals, &s.TotalGoals,
		&s.ActiveSprints, &s.CompletedSprints,
		&s.TasksToday, &s.TasksCompletedToday,
		&s.SRMEvents30d, &s.SRML3Events30d,
		&s.RegressionEvents30d,
		&s.Ceremonies30d, &s.BadgesAwarded30d,
		&s.ComputedAt,
	)
	return s, err
}

// GetAdminUserList returns all users with computed stats for the admin panel.
func GetAdminUserList(ctx context.Context, pool *pgxpool.Pool) ([]models.AdminUserRecord, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			id, full_name, locale, is_active, is_admin, mfa_enabled,
			created_at, updated_at,
			active_goals, completed_goals, total_goals, completed_sprints,
			tasks_last_30d, last_active_at, active_sessions
		FROM v_admin_user_list
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.AdminUserRecord
	for rows.Next() {
		u := models.AdminUserRecord{}
		if err := rows.Scan(
			&u.ID, &u.FullName, &u.Locale, &u.IsActive, &u.IsAdmin, &u.MFAEnabled,
			&u.CreatedAt, &u.UpdatedAt,
			&u.ActiveGoals, &u.CompletedGoals, &u.TotalGoals, &u.CompletedSprints,
			&u.TasksLast30d, &u.LastActiveAt, &u.ActiveSessions,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// GetAuditLog returns the most recent audit entries for the admin panel.
func GetAuditLog(ctx context.Context, pool *pgxpool.Pool, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := pool.Query(ctx, `
		SELECT a.id, a.user_id, u.full_name, a.action, a.created_at
		FROM audit_log a
		LEFT JOIN users u ON u.id = a.user_id
		ORDER BY a.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var userID *uuid.UUID
		var fullName *string
		var action string
		var createdAt time.Time
		if err := rows.Scan(&id, &userID, &fullName, &action, &createdAt); err != nil {
			continue
		}
		entries = append(entries, map[string]interface{}{
			"id":         id,
			"user_id":    userID,
			"full_name":  fullName,
			"action":     action,
			"created_at": createdAt,
		})
	}
	return entries, rows.Err()
}

// SetUserActiveStatus activates or deactivates a user (admin only).
func SetUserActiveStatus(ctx context.Context, pool *pgxpool.Pool, targetUserID uuid.UUID, active bool) error {
	_, err := pool.Exec(ctx,
		`UPDATE users SET is_active=$1, updated_at=NOW() WHERE id=$2 AND is_admin=FALSE`,
		active, targetUserID)
	return err
}

// PromoteToAdmin grants admin privileges to a user (admin only).
func PromoteToAdmin(ctx context.Context, pool *pgxpool.Pool, targetUserID uuid.UUID) error {
	_, err := pool.Exec(ctx,
		`UPDATE users SET is_admin=TRUE, updated_at=NOW() WHERE id=$1`, targetUserID)
	return err
}

// ═══════════════════════════════════════════════════════════════
// STREAK CALCULATION
// ═══════════════════════════════════════════════════════════════

func GetStreakDays(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) (int, error) {
	// Numără zilele consecutive în care userul a completat cel puțin o sarcină
	var streak int
	err := pool.QueryRow(ctx, `
		WITH daily_completion AS (
			SELECT task_date,
				   BOOL_OR(completed) AS had_completion
			FROM daily_tasks
			WHERE user_id = $1
			  AND task_date <= CURRENT_DATE
			  AND task_type = 'MAIN'
			GROUP BY task_date
		),
		ordered AS (
			SELECT task_date,
				   had_completion,
				   ROW_NUMBER() OVER (ORDER BY task_date DESC) AS rn,
				   task_date - (CURRENT_DATE - (ROW_NUMBER() OVER (ORDER BY task_date DESC) - 1)::int) AS grp
			FROM daily_completion
			WHERE had_completion = TRUE
		)
		SELECT COUNT(*) FROM ordered WHERE grp = '0 days'::interval
	`, userID).Scan(&streak)
	if err != nil {
		return 0, fmt.Errorf("streak calc: %w", err)
	}
	return streak, nil
}

// ═══════════════════════════════════════════════════════════════
// PASSWORD RESET TOKENS
// ═══════════════════════════════════════════════════════════════

// CreatePasswordResetToken inserts a new reset token for the user.
// tokenHash is SHA-256 hex of the raw random token sent in the email link.
func CreatePasswordResetToken(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, tokenHash string) error {
	// Invalidate any previous unused tokens for this user first
	_, _ = pool.Exec(ctx,
		`DELETE FROM password_reset_tokens WHERE user_id=$1 AND used_at IS NULL`, userID)
	_, err := pool.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash)
		VALUES ($1, $2)
	`, userID, tokenHash)
	return err
}

// GetPasswordResetToken returns a valid (unused, not expired) token record.
func GetPasswordResetToken(ctx context.Context, pool *pgxpool.Pool, tokenHash string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT user_id FROM password_reset_tokens
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND expires_at > NOW()
	`, tokenHash).Scan(&userID)
	if err != nil {
		return uuid.Nil, ErrNotFound
	}
	return userID, nil
}

// MarkPasswordResetTokenUsed marks the token as consumed so it cannot be reused.
func MarkPasswordResetTokenUsed(ctx context.Context, pool *pgxpool.Pool, tokenHash string) error {
	_, err := pool.Exec(ctx,
		`UPDATE password_reset_tokens SET used_at=NOW() WHERE token_hash=$1`, tokenHash)
	return err
}
