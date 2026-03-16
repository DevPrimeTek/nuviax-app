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
		          full_name, locale, mfa_enabled, is_active, created_at, updated_at
	`, emailEncrypted, emailHash, passwordHash, salt, fullName, locale).Scan(
		&u.ID, &u.EmailEncrypted, &u.EmailHash, &u.PasswordHash, &u.Salt,
		&u.FullName, &u.Locale, &u.MFAEnabled, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

func GetUserByEmailHash(ctx context.Context, pool *pgxpool.Pool, hash string) (*models.User, error) {
	u := &models.User{}
	err := pool.QueryRow(ctx, `
		SELECT id, email_encrypted, email_hash, password_hash, salt,
		       full_name, locale, mfa_secret, mfa_enabled, is_active, created_at, updated_at
		FROM users WHERE email_hash = $1 AND is_active = TRUE
	`, hash).Scan(
		&u.ID, &u.EmailEncrypted, &u.EmailHash, &u.PasswordHash, &u.Salt,
		&u.FullName, &u.Locale, &u.MFASecret, &u.MFAEnabled, &u.IsActive,
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
		       full_name, locale, mfa_secret, mfa_enabled, is_active, created_at, updated_at
		FROM users WHERE id = $1 AND is_active = TRUE
	`, id).Scan(
		&u.ID, &u.EmailEncrypted, &u.EmailHash, &u.PasswordHash, &u.Salt,
		&u.FullName, &u.Locale, &u.MFASecret, &u.MFAEnabled, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
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
		RETURNING id, user_id, name, description, status, start_date, end_date, created_at, updated_at
	`, userID, name, desc, status, startDate, endDate).Scan(
		&g.ID, &g.UserID, &g.Name, &g.Description, &g.Status,
		&g.StartDate, &g.EndDate, &g.CreatedAt, &g.UpdatedAt,
	)
	return g, err
}

func GetGoalByID(ctx context.Context, pool *pgxpool.Pool, goalID, userID uuid.UUID) (*models.Goal, error) {
	g := &models.Goal{}
	err := pool.QueryRow(ctx, `
		SELECT id, user_id, name, description, status, start_date, end_date, created_at, updated_at
		FROM global_objectives WHERE id=$1 AND user_id=$2
	`, goalID, userID).Scan(
		&g.ID, &g.UserID, &g.Name, &g.Description, &g.Status,
		&g.StartDate, &g.EndDate, &g.CreatedAt, &g.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return g, err
}

func GetGoalsByUser(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]models.Goal, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, user_id, name, description, status, start_date, end_date, created_at, updated_at
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
			&g.StartDate, &g.EndDate, &g.CreatedAt, &g.UpdatedAt); err != nil {
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

// CreateDefaultCheckpoints creează 3 checkpoints standard pentru un sprint nou.
// Faza 1 (UPCOMING) devine IN_PROGRESS prin scheduler la 23:55 UTC.
func CreateDefaultCheckpoints(ctx context.Context, pool *pgxpool.Pool, sprintID uuid.UUID, goalName string) {
	label := goalName
	if len(label) > 50 {
		label = label[:50]
	}
	phases := []struct {
		name  string
		order int
	}{
		{"Faza 1 — Start: " + label, 0},
		{"Faza 2 — Execuție: " + label, 1},
		{"Faza 3 — Finalizare: " + label, 2},
	}
	for _, p := range phases {
		n := p.name
		CreateCheckpoint(ctx, pool, sprintID, n, nil, p.order) //nolint:errcheck
	}
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
	ca := &models.ContextAdjustment{}
	err := pool.QueryRow(ctx, `
		INSERT INTO context_adjustments
			(go_id, user_id, adj_type, start_date, end_date, note)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, go_id, user_id, adj_type, start_date, end_date, note, created_at
	`, goalID, userID, adjType, startDate, endDate, note).Scan(
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
