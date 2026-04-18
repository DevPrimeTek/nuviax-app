package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/api/middleware"
	"github.com/devprimetek/nuviax-app/internal/cache"
	"github.com/devprimetek/nuviax-app/internal/engine"
)

type dashGoal struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	ProgressPct float64   `json:"progress_pct"`
	Grade       string    `json:"grade"`
	StreakDays  int       `json:"streak_days"`
}

type dashData struct {
	Goals           []dashGoal `json:"goals"`
	ActiveCount     int        `json:"active_count"`
	TasksTodayDone  int        `json:"tasks_today_done"`
	TasksTodayTotal int        `json:"tasks_today_total"`
	SRMActive       bool       `json:"srm_active"`
}

// GET /dashboard — Summary of active GOs and today's progress (Redis-cached 5 min).
func (h *Handlers) GetDashboard(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	userIDStr := userID.String()

	var cached dashData
	if err := cache.GetDashboard(c.Context(), h.redis, userIDStr, &cached); err == nil {
		return c.JSON(cached)
	}

	rows, err := h.db.Query(c.Context(), `
		SELECT go.id, go.name, COALESCE(ds.real_progress, 0) AS progress_pct
		FROM global_objectives go
		LEFT JOIN LATERAL (
			SELECT real_progress FROM daily_scores
			WHERE go_id = go.id
			ORDER BY score_date DESC LIMIT 1
		) ds ON TRUE
		WHERE go.user_id = $1 AND go.status = 'ACTIVE'
		ORDER BY go.created_at
	`, userID)
	if err != nil {
		return serverError(c, err)
	}
	defer rows.Close()

	goals := make([]dashGoal, 0)
	for rows.Next() {
		var g dashGoal
		if err := rows.Scan(&g.ID, &g.Name, &g.ProgressPct); err != nil {
			continue
		}
		g.Grade = engine.ScoreToGrade(g.ProgressPct)
		goals = append(goals, g)
	}

	streak := h.computeStreak(c.Context(), userID)
	for i := range goals {
		goals[i].StreakDays = streak
	}

	today := time.Now().UTC().Format("2006-01-02")
	var tasksDone, tasksTotal int
	_ = h.db.QueryRow(c.Context(), `
		SELECT
			COUNT(*) FILTER (WHERE status = 'DONE'),
			COUNT(*) FILTER (WHERE status IN ('DONE', 'PENDING'))
		FROM daily_tasks
		WHERE user_id=$1 AND task_date=$2
	`, userID, today).Scan(&tasksDone, &tasksTotal)

	data := dashData{
		Goals:           goals,
		ActiveCount:     len(goals),
		TasksTodayDone:  tasksDone,
		TasksTodayTotal: tasksTotal,
		SRMActive:       false,
	}

	_ = cache.SetDashboard(c.Context(), h.redis, userIDStr, data)

	return c.JSON(data)
}
