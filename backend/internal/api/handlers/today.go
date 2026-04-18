package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/api/middleware"
)

// GET /today — Daily stack for the user's primary active GO.
func (h *Handlers) GetToday(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	today := time.Now().UTC().Format("2006-01-02")

	// Primary active GO + sprint
	var goalID, sprintID uuid.UUID
	var goalName string
	var sprintNumber int
	var sprintStart time.Time
	goErr := h.db.QueryRow(c.Context(), `
		SELECT go.id, go.name, s.id, s.sprint_number, s.start_date
		FROM global_objectives go
		JOIN sprints s ON s.go_id = go.id AND s.status = 'ACTIVE'
		WHERE go.user_id = $1 AND go.status = 'ACTIVE'
		ORDER BY go.created_at LIMIT 1
	`, userID).Scan(&goalID, &goalName, &sprintID, &sprintNumber, &sprintStart)

	dayNumber := 0
	if goErr == nil && !sprintStart.IsZero() {
		dayNumber = int(time.Since(sprintStart).Hours()/24) + 1
		if dayNumber < 1 {
			dayNumber = 1
		}
		if dayNumber > 30 {
			dayNumber = 30
		}
	}

	var progressPct float64
	if goErr == nil {
		_ = h.db.QueryRow(c.Context(), `
			SELECT COALESCE(real_progress, 0) FROM daily_scores
			WHERE go_id=$1 ORDER BY score_date DESC LIMIT 1
		`, goalID).Scan(&progressPct)
	}

	streak := h.computeStreak(c.Context(), userID)

	// Tasks for today (not cancelled/skipped)
	rows, err := h.db.Query(c.Context(), `
		SELECT id, title, task_type::text, status::text
		FROM daily_tasks
		WHERE user_id=$1 AND task_date=$2 AND status != 'SKIPPED'
		ORDER BY task_type, created_at
	`, userID, today)

	type taskItem struct {
		ID    uuid.UUID `json:"id"`
		Title string    `json:"title"`
		Type  string    `json:"type"`
		Done  bool      `json:"done"`
	}

	mainTasks := make([]taskItem, 0)
	personalTasks := make([]taskItem, 0)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var t taskItem
			var taskType, status string
			if err := rows.Scan(&t.ID, &t.Title, &taskType, &status); err != nil {
				continue
			}
			t.Type = taskType
			t.Done = status == "DONE"
			if taskType == "OPTIONAL" {
				personalTasks = append(personalTasks, t)
			} else {
				mainTasks = append(mainTasks, t)
			}
		}
	}

	checkpointName := ""
	if goErr == nil && sprintNumber > 0 {
		checkpointName = fmt.Sprintf("Sprint %d — Day %d/30", sprintNumber, dayNumber)
	}

	return c.JSON(fiber.Map{
		"goal_name":  goalName,
		"day_number": dayNumber,
		"checkpoint": fiber.Map{
			"name":         checkpointName,
			"progress_pct": progressPct,
		},
		"streak_days":    streak,
		"main_tasks":     mainTasks,
		"personal_tasks": personalTasks,
	})
}

// POST /today/complete/:id — Mark a task as DONE.
func (h *Handlers) CompleteTask(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	tag, err := h.db.Exec(c.Context(), `
		UPDATE daily_tasks
		SET status='DONE', completed_at=NOW()
		WHERE id=$1 AND user_id=$2 AND status='PENDING'
	`, taskID, userID)
	if err != nil {
		return serverError(c, err)
	}
	if tag.RowsAffected() == 0 {
		return notFound(c)
	}

	return c.JSON(fiber.Map{"ok": true})
}

// POST /today/personal — Add a personal task (max 2/day, type OPTIONAL).
func (h *Handlers) AddPersonalTask(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	today := time.Now().UTC().Format("2006-01-02")

	var req struct {
		Title string `json:"title"`
	}
	if err := c.BodyParser(&req); err != nil || req.Title == "" {
		return badRequest(c, "Titlul sarcinii este obligatoriu.")
	}

	var personalCount int
	_ = h.db.QueryRow(c.Context(), `
		SELECT COUNT(*) FROM daily_tasks
		WHERE user_id=$1 AND task_date=$2 AND task_type='OPTIONAL'
	`, userID, today).Scan(&personalCount)

	if personalCount >= 2 {
		return badRequest(c, "Maxim 2 sarcini personale pe zi.")
	}

	// Require an active GO to attach the task to
	var goID, sprintID uuid.UUID
	if err := h.db.QueryRow(c.Context(), `
		SELECT go.id, s.id
		FROM global_objectives go
		JOIN sprints s ON s.go_id = go.id AND s.status = 'ACTIVE'
		WHERE go.user_id = $1 AND go.status = 'ACTIVE'
		LIMIT 1
	`, userID).Scan(&goID, &sprintID); err != nil {
		return c.Status(422).JSON(fiber.Map{"error": "Nu există un obiectiv activ."})
	}

	var taskID uuid.UUID
	var title string
	if err := h.db.QueryRow(c.Context(), `
		INSERT INTO daily_tasks (go_id, sprint_id, user_id, task_date, title, task_type)
		VALUES ($1,$2,$3,$4,$5,'OPTIONAL')
		RETURNING id, title
	`, goID, sprintID, userID, today, req.Title).Scan(&taskID, &title); err != nil {
		return serverError(c, err)
	}

	return c.Status(201).JSON(fiber.Map{
		"id":    taskID,
		"title": title,
		"type":  "OPTIONAL",
		"done":  false,
	})
}

// POST /context/energy — Record energy level for the day.
// context_adjustments table is pending migration — accepted and acknowledged without DB write.
func (h *Handlers) SetEnergy(c *fiber.Ctx) error {
	var req struct {
		Level string `json:"level"`
	}
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}
	switch req.Level {
	case "low", "mid", "hi":
	default:
		return badRequest(c, "Nivel valid: low, mid, hi")
	}
	return c.JSON(fiber.Map{"ok": true})
}

// computeStreak returns consecutive days (backwards from today) with at least one DONE task.
func (h *Handlers) computeStreak(ctx context.Context, userID uuid.UUID) int {
	rows, err := h.db.Query(ctx, `
		SELECT task_date FROM daily_tasks
		WHERE user_id=$1 AND status='DONE' AND task_date <= CURRENT_DATE
		GROUP BY task_date ORDER BY task_date DESC LIMIT 30
	`, userID)
	if err != nil {
		return 0
	}
	defer rows.Close()

	streak := 0
	// Start checkDate at tomorrow so "today" counts as day 1
	checkDate := time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, 1)

	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			break
		}
		d = d.UTC().Truncate(24 * time.Hour)
		expected := checkDate.AddDate(0, 0, -1)
		if d.Equal(expected) {
			streak++
			checkDate = expected
		} else {
			break
		}
	}
	return streak
}
