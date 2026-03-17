package engine

// Level 2 — Execution Engine (C19-C25)
// Calculează ratele de execuție și scorul de sprint.
// Toate metricile sunt opace față de exterior.

import (
	"context"

	"github.com/google/uuid"
)

// C19 — computeCompletionRate: rata sarcinilor MAIN completate în sprint
func (e *Engine) computeCompletionRate(ctx context.Context, sprintID uuid.UUID) float64 {
	var total, completed int
	e.db.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE completed = TRUE)
		FROM daily_tasks
		WHERE sprint_id = $1 AND task_type = 'MAIN'
	`, sprintID).Scan(&total, &completed)
	if total == 0 {
		return 0
	}
	return float64(completed) / float64(total)
}

// C20 — computeSprintInternal: scorul brut al unui sprint (0-1)
func (e *Engine) computeSprintInternal(ctx context.Context, sprintID uuid.UUID) float64 {
	var total, completed int
	e.db.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE completed = TRUE)
		FROM daily_tasks
		WHERE sprint_id = $1 AND task_type = 'MAIN'
	`, sprintID).Scan(&total, &completed)
	if total == 0 {
		return 0
	}
	return clamp(float64(completed)/float64(total), 0, 1)
}
