package engine

// Level 2 — Execution Engine (C19-C25)
// Calculează ratele de execuție și scorul de sprint.
// Toate metricile sunt opace față de exterior.

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/db"
)

// chaosIndexL2Threshold — Chaos Index >= 0.40 triggers SRM L2 (G-3)
const chaosIndexL2Threshold = 0.40

// stagnationThresholdDays — consecutive days without progress before Focus Rotation (G-5)
const stagnationThresholdDays = 5

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

// C20 — computeSprintInternal: scorul compozit al unui sprint (0-1)
// GAP G-8 FIX — formula completă: 40% completion + 25% consistency + 25% progress + 10% energy
func (e *Engine) computeSprintInternal(ctx context.Context, sprintID uuid.UUID) float64 {
	total, completed := e.querySprintTaskRatio(ctx, sprintID)
	if total == 0 {
		return 0
	}
	completionRate := clamp(float64(completed)/float64(total), 0, 1)

	// Load sprint for full formula — graceful fallback to completion-only on error
	sprint, err := db.GetSprintByID(ctx, e.db, sprintID)
	if err != nil {
		return completionRate
	}

	goal, err := db.GetGoalByID(ctx, e.db, sprint.GoalID, uuid.Nil)
	if err != nil {
		return completionRate
	}

	consistency := e.computeConsistency(ctx, sprint)
	progress := e.computeProgressVsExpected(ctx, goal, sprint)
	contextPenalty, energyBonus := e.computeContextFactors(ctx, sprint.GoalID)

	return clamp(
		completionRate*0.40+
			consistency*0.25+
			progress*0.25+
			energyBonus*0.10-
			contextPenalty*0.05,
		0, 1,
	)
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

// ── GAP G-3 — Chaos Index ───────────────────────────────────────────────────

// computeChaosIndex measures daily completion rate variability (G-3).
// CI = std_dev(daily_rates) / mean(daily_rates) — Coefficient of Variation.
// Requires at least 3 days of data; returns 0 otherwise.
func (e *Engine) computeChaosIndex(ctx context.Context, sprintID uuid.UUID) float64 {
	rows, err := e.db.Query(ctx, `
		SELECT
			CAST(COUNT(*) FILTER (WHERE completed = TRUE) AS FLOAT) /
			NULLIF(COUNT(*), 0)
		FROM daily_tasks
		WHERE sprint_id = $1 AND task_type = 'MAIN'
		GROUP BY task_date
		ORDER BY task_date
	`, sprintID)
	if err != nil {
		return 0
	}
	defer rows.Close()

	var rates []float64
	for rows.Next() {
		var rate float64
		if err := rows.Scan(&rate); err == nil {
			rates = append(rates, rate)
		}
	}

	if len(rates) < 3 {
		return 0 // not enough data
	}

	sum := 0.0
	for _, r := range rates {
		sum += r
	}
	mean := sum / float64(len(rates))
	if mean < 0.01 {
		return 1.0 // fully inactive = maximum chaos
	}

	sumSq := 0.0
	for _, r := range rates {
		d := r - mean
		sumSq += d * d
	}
	stdDev := math.Sqrt(sumSq / float64(len(rates)))
	return clamp(stdDev/mean, 0, 2)
}

// CheckChaosIndex returns the Chaos Index for a sprint and whether SRM L2 should trigger (G-3).
// CI >= 0.40 → L2 trigger. Public API — called by scheduler and SRM handlers.
func (e *Engine) CheckChaosIndex(ctx context.Context, sprintID uuid.UUID) (ci float64, triggerL2 bool) {
	ci = e.computeChaosIndex(ctx, sprintID)
	triggerL2 = ci >= chaosIndexL2Threshold
	return
}

// ── GAP G-5 — Stagnation Detection ─────────────────────────────────────────

// ConsecutiveInactiveDays returns the number of consecutive days without any
// completed main task for a given goal (G-5).
// Returns 0 if a task was completed today. Caps at 30.
func (e *Engine) ConsecutiveInactiveDays(ctx context.Context, goalID uuid.UUID) int {
	var lastCompleted time.Time
	err := e.db.QueryRow(ctx, `
		SELECT COALESCE(MAX(task_date), '0001-01-01'::date)
		FROM daily_tasks
		WHERE go_id = $1 AND task_type = 'MAIN' AND completed = TRUE
	`, goalID).Scan(&lastCompleted)

	if err != nil || lastCompleted.IsZero() || lastCompleted.Year() == 1 {
		// No completed tasks — count from goal creation
		var created time.Time
		e.db.QueryRow(ctx, `SELECT created_at FROM global_objectives WHERE id = $1`, goalID).Scan(&created)
		if created.IsZero() {
			return 0
		}
		days := int(time.Now().UTC().Truncate(24*time.Hour).Sub(created.Truncate(24*time.Hour)).Hours() / 24)
		if days > 30 {
			return 30
		}
		return days
	}

	days := int(time.Now().UTC().Truncate(24*time.Hour).Sub(lastCompleted.Truncate(24*time.Hour)).Hours() / 24)
	if days > 30 {
		return 30
	}
	return days
}

// IsStagnant returns true when a goal has been inactive for >= stagnationThresholdDays (G-5).
func (e *Engine) IsStagnant(ctx context.Context, goalID uuid.UUID) bool {
	return e.ConsecutiveInactiveDays(ctx, goalID) >= stagnationThresholdDays
}

// ── GAP G-6 — Velocity Control ──────────────────────────────────────────────

// IsVelocityControlActive checks if ALI_projected > 1.15 for a goal's current sprint (G-6).
// When true, task generation should reduce count by 1 to avoid overload.
func (e *Engine) IsVelocityControlActive(ctx context.Context, goalID uuid.UUID) bool {
	sprint, err := db.GetCurrentSprint(ctx, e.db, goalID)
	if err != nil || sprint == nil {
		return false
	}

	var completedTasks, totalTasks int
	e.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE completed = TRUE),
			COUNT(*)
		FROM daily_tasks
		WHERE sprint_id = $1 AND task_type = 'MAIN'
	`, sprint.ID).Scan(&completedTasks, &totalTasks)

	if totalTasks == 0 {
		return false
	}

	sprintDuration := sprint.EndDate.Sub(sprint.StartDate).Hours() / 24
	elapsed := time.Now().UTC().Sub(sprint.StartDate).Hours() / 24
	remaining := sprintDuration - elapsed

	if elapsed <= 0 || remaining <= 0 {
		return false
	}

	dailyRate := float64(completedTasks) / math.Max(elapsed, 1)
	projectedTotal := float64(completedTasks) + dailyRate*remaining
	expectedTotal := float64(totalTasks)

	if expectedTotal <= 0 {
		return false
	}

	aliProjected := projectedTotal / expectedTotal
	return aliProjected > 1.15 // Ambition Buffer upper limit
}
