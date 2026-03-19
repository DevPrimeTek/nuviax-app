package engine

// Level 2 — Execution Engine (C19-C25)
// Calculează ratele de execuție și scorul de sprint.
// Toate metricile sunt opace față de exterior.

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// querySprintTaskRatio este helper intern — interogare unică pentru total/completed sarcini MAIN
func (e *Engine) querySprintTaskRatio(ctx context.Context, sprintID uuid.UUID) (total, completed int) {
	e.db.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE completed = TRUE)
		FROM daily_tasks
		WHERE sprint_id = $1 AND task_type = 'MAIN'
	`, sprintID).Scan(&total, &completed)
	return
}

// C19 — computeCompletionRate: rata sarcinilor MAIN completate în sprint
func (e *Engine) computeCompletionRate(ctx context.Context, sprintID uuid.UUID) float64 {
	total, completed := e.querySprintTaskRatio(ctx, sprintID)
	if total == 0 {
		return 0
	}
	return float64(completed) / float64(total)
}

// C20 — computeSprintInternal: scorul brut al unui sprint (0-1)
func (e *Engine) computeSprintInternal(ctx context.Context, sprintID uuid.UUID) float64 {
	total, completed := e.querySprintTaskRatio(ctx, sprintID)
	if total == 0 {
		return 0
	}
	return clamp(float64(completed)/float64(total), 0, 1)
}

// GAP #15 — CheckAndRecordRegressionEvent detects when a goal's current progress
// metric value has dropped below the sprint start value (e.g. MRR loss, weight gain
// in a REDUCE goal going wrong). This cannot be handled by simple clamp(0,1) because
// the user genuinely moved backward — that event must be tracked explicitly.
//
// Returns true if a regression event was detected and recorded.
// The clamp logic in computeProgressVsExpected still applies afterward,
// but the regression flag enables SRM L1 auto-activation and UI alert.
func (e *Engine) CheckAndRecordRegressionEvent(
	ctx context.Context,
	goalID, sprintID, userID uuid.UUID,
	currentValue, sprintStartValue float64,
) (bool, error) {
	if currentValue >= sprintStartValue {
		return false, nil // Normal case — no regression
	}

	delta := currentValue - sprintStartValue // Always negative here

	// Avoid duplicate regression records for the same day
	var existingCount int
	e.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM regression_events
		WHERE go_id = $1 AND sprint_id = $2
		  AND detected_at::date = CURRENT_DATE
	`, goalID, sprintID).Scan(&existingCount)

	if existingCount > 0 {
		return true, nil // Already recorded today
	}

	_, err := e.db.Exec(ctx, `
		INSERT INTO regression_events
			(id, go_id, sprint_id, user_id, detected_at,
			 value_at_detection, value_at_sprint_start, regression_delta)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW(), $4, $5, $6)
	`, goalID, sprintID, userID, currentValue, sprintStartValue, delta)
	if err != nil {
		return false, fmt.Errorf("record regression event: %w", err)
	}

	return true, nil
}

// GetRegressionEvents returns open (unresolved) regression events for a sprint
func (e *Engine) GetRegressionEvents(ctx context.Context, sprintID uuid.UUID) ([]map[string]interface{}, error) {
	rows, err := e.db.Query(ctx, `
		SELECT id, go_id, detected_at, value_at_detection,
		       value_at_sprint_start, regression_delta
		FROM regression_events
		WHERE sprint_id = $1 AND resolved_at IS NULL
		ORDER BY detected_at DESC
	`, sprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var id, goID uuid.UUID
		var detectedAt time.Time
		var valDetection, valStart, delta float64
		if err := rows.Scan(&id, &goID, &detectedAt, &valDetection, &valStart, &delta); err != nil {
			continue
		}
		events = append(events, map[string]interface{}{
			"id":                    id,
			"goal_id":               goID,
			"detected_at":           detectedAt,
			"value_at_detection":    valDetection,
			"value_at_sprint_start": valStart,
			"regression_delta":      delta,
		})
	}
	return events, rows.Err()
}
