package engine

// Level 5 — Growth Orchestration (C37-C40)
// Calculează progresul față de traiectoria așteptată și componentele de creștere.
// Valorile intermediare nu sunt niciodată expuse în afara engine-ului.

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/models"
)

// ═══════════════════════════════════════════════════════════════
// PUBLIC API — Level 5 growth queries exposed to handlers
// ═══════════════════════════════════════════════════════════════

// GetUserAchievements returns all awarded achievement badges for a user (C39)
func (e *Engine) GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]models.AchievementBadge, error) {
	rows, err := e.db.Query(ctx, `
		SELECT id, user_id, badge_type, go_id, sprint_id, awarded_at
		FROM achievement_badges
		WHERE user_id = $1
		ORDER BY awarded_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var badges []models.AchievementBadge
	for rows.Next() {
		var b models.AchievementBadge
		if err := rows.Scan(&b.ID, &b.UserID, &b.BadgeType, &b.GoalID, &b.SprintID, &b.AwardedAt); err != nil {
			continue
		}
		badges = append(badges, b)
	}
	return badges, rows.Err()
}

// trajectoryPoint is the public shape returned by GenerateProgressVisualization
type trajectoryPoint struct {
	Date        time.Time `json:"date"`
	ActualPct   float64   `json:"actual_pct"`
	ExpectedPct float64   `json:"expected_pct"`
	Delta       float64   `json:"delta"`
	Trend       string    `json:"trend"`
}

// GenerateProgressVisualization returns trajectory data for the goal progress chart (C40)
// Only relative percentages and trend labels are exposed — no internal weights.
func (e *Engine) GenerateProgressVisualization(ctx context.Context, goalID uuid.UUID) (map[string]interface{}, error) {
	rows, err := e.db.Query(ctx, `
		SELECT snapshot_date, actual_pct, expected_pct, delta, trend
		FROM growth_trajectories
		WHERE go_id = $1
		ORDER BY snapshot_date ASC
	`, goalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trajectory []trajectoryPoint
	for rows.Next() {
		var p trajectoryPoint
		if err := rows.Scan(&p.Date, &p.ActualPct, &p.ExpectedPct, &p.Delta, &p.Trend); err != nil {
			continue
		}
		trajectory = append(trajectory, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// If no stored trajectory yet, compute a live snapshot
	if len(trajectory) == 0 {
		var start, end time.Time
		if err := e.db.QueryRow(ctx,
			`SELECT start_date, end_date FROM goals WHERE id = $1`, goalID,
		).Scan(&start, &end); err == nil {
			total := end.Sub(start).Hours()
			elapsed := time.Now().UTC().Sub(start).Hours()
			exp := 0.0
			if total > 0 {
				exp = math.Round((elapsed/total)*100) / 100
			}
			trajectory = append(trajectory, trajectoryPoint{
				Date:        time.Now().UTC(),
				ActualPct:   0,
				ExpectedPct: exp,
				Delta:       -exp,
				Trend:       "ON_TRACK",
			})
		}
	}

	return map[string]interface{}{
		"goal_id":    goalID,
		"trajectory": trajectory,
	}, nil
}

// MarkEvolutionSprint checks whether a completed sprint shows measurable evolution
// compared to the previous one and, if so, inserts an evolution_sprints record (C37).
// Returns (true, nil) when evolution was detected and recorded.
// Returns (false, nil) when delta is below threshold — NOT an error condition.
// Returns (false, err) only on actual DB or computation errors.
func (e *Engine) MarkEvolutionSprint(ctx context.Context, sprintID, goalID uuid.UUID) (bool, error) {
	score, _, err := e.ComputeSprintScore(ctx, sprintID)
	if err != nil {
		return false, fmt.Errorf("compute sprint score: %w", err)
	}

	// Compare against the most recent previously completed sprint
	var prevScore float64
	_ = e.db.QueryRow(ctx, `
		SELECT COALESCE(sr.score_value, 0)
		FROM sprints s
		JOIN sprint_results sr ON sr.sprint_id = s.id
		WHERE s.go_id = $1
		  AND s.status = 'COMPLETED'
		  AND s.id != $2
		ORDER BY s.sprint_number DESC
		LIMIT 1
	`, goalID, sprintID).Scan(&prevScore)

	delta := score - prevScore

	// G-11: Apply behavior model override if configured
	evolved, overrideErr := e.ApplyEvolveOverride(ctx, goalID, score, prevScore)
	if overrideErr != nil {
		// Log but don't fail — use standard threshold
		_ = overrideErr
	} else if evolved {
		// Behavior model indicated evolution — proceed with recording
		consistency := e.computeConsistencyForGoal(ctx, goalID)
		_, err = e.db.Exec(ctx, `
			INSERT INTO evolution_sprints
				(id, sprint_id, go_id, evolution_score, delta_performance,
				 consistency_weight, acceleration_factor, detected_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, 1.0, NOW())
			ON CONFLICT (sprint_id) DO NOTHING
		`, sprintID, goalID, score, delta, consistency)
		if err != nil {
			return false, fmt.Errorf("insert evolution_sprint (override): %w", err)
		}
		return true, nil
	}

	if delta < 0.05 {
		// Sub prag — nu e eroare, pur și simplu nu e evolution sprint
		return false, nil
	}

	consistency := e.computeConsistencyForGoal(ctx, goalID)

	_, err = e.db.Exec(ctx, `
		INSERT INTO evolution_sprints
			(id, sprint_id, go_id, evolution_score, delta_performance,
			 consistency_weight, acceleration_factor, detected_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, 1.0, NOW())
		ON CONFLICT (sprint_id) DO NOTHING
	`, sprintID, goalID, score, delta, consistency)
	if err != nil {
		return false, fmt.Errorf("insert evolution_sprint: %w", err)
	}
	return true, nil
}

// GenerateCompletionCeremony creates a completion_ceremonies record (C38)
// for a sprint that has just finished. Tier is determined from sprint score
// and whether the sprint also qualified as an evolution sprint.
func (e *Engine) GenerateCompletionCeremony(ctx context.Context, sprintID, goalID uuid.UUID, isEvolution bool) error {
	score, _, err := e.ComputeSprintScore(ctx, sprintID)
	if err != nil {
		return fmt.Errorf("compute sprint score: %w", err)
	}

	tier := "BRONZE"
	switch {
	case isEvolution && score >= 0.9:
		tier = "PLATINUM"
	case score >= 0.9:
		tier = "GOLD"
	case score >= 0.75:
		tier = "SILVER"
	}

	_, err = e.db.Exec(ctx, `
		INSERT INTO completion_ceremonies
			(id, sprint_id, go_id, ceremony_tier, ceremony_data, viewed, generated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, '{"auto_generated":true}'::jsonb, false, NOW())
		ON CONFLICT (sprint_id) DO NOTHING
	`, sprintID, goalID, tier)
	if err != nil {
		return fmt.Errorf("insert completion_ceremony: %w", err)
	}
	return nil
}

// computeConsistencyForGoal returns overall consistency across all sprints for a goal
func (e *Engine) computeConsistencyForGoal(ctx context.Context, goalID uuid.UUID) float64 {
	var activeDays, totalDays int
	e.db.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT task_date) FILTER (WHERE completed = TRUE),
			COUNT(DISTINCT task_date)
		FROM daily_tasks
		WHERE go_id = $1 AND task_type = 'MAIN'
	`, goalID).Scan(&activeDays, &totalDays)
	if totalDays == 0 {
		return 0
	}
	return float64(activeDays) / float64(totalDays)
}

// C37 — computeProgressVsExpected: progresul real față de traiectoria liniară așteptată
// Ratio > 1 = înaintea planului, < 1 = în urmă
//
// GAP #20 FIX — Stabilization Mode Freeze:
// When SRM Level 3 is active (Stabilization Mode), the expected trajectory is frozen
// at the value it had when L3 was triggered. Without this fix, expectedPct keeps
// advancing while the user is in stabilization, creating a "drift loop paradox" where
// the drift score worsens even when the user correctly follows the protocol.
func (e *Engine) computeProgressVsExpected(ctx context.Context, goal *models.Goal, sprint *models.Sprint) float64 {
	now := time.Now().UTC()
	totalDuration := goal.EndDate.Sub(goal.StartDate).Hours()
	elapsed := now.Sub(goal.StartDate).Hours()
	if totalDuration <= 0 {
		return 0
	}

	// GAP #20 — Check if sprint has a frozen expected_pct (Stabilization Mode active)
	var isFrozen bool
	var frozenPct *float64
	e.db.QueryRow(ctx, `
		SELECT expected_pct_frozen, frozen_expected_pct
		FROM sprints WHERE id = $1
	`, sprint.ID).Scan(&isFrozen, &frozenPct)

	var expectedPct float64
	if isFrozen && frozenPct != nil {
		// Use the frozen value — do not advance during Stabilization Mode
		expectedPct = *frozenPct
	} else {
		expectedPct = elapsed / totalDuration
	}

	var completedCP, totalCP int
	e.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'COMPLETED'),
			COUNT(*)
		FROM checkpoints
		WHERE sprint_id = $1
	`, sprint.ID).Scan(&completedCP, &totalCP)

	if totalCP == 0 {
		return clamp(expectedPct, 0, 1) // Fără checkpointuri: progres temporal pur
	}

	actualPct := float64(completedCP) / float64(totalCP)
	// ratio > 1 = înaintea planului, < 1 = în urmă; clampăm la [0, 1] pentru scor opac
	ratio := actualPct / math.Max(expectedPct, 0.01)
	return clamp(ratio, 0, 1)
}

// FreezeExpectedTrajectory freezes the expected_pct on the sprint when SRM L3
// (Stabilization Mode) is activated, preventing the drift loop paradox (GAP #20).
// Call this when SRM L3 is confirmed by the user.
func (e *Engine) FreezeExpectedTrajectory(ctx context.Context, sprintID uuid.UUID) error {
	// Compute current expected_pct before freezing
	var startDate, endDate time.Time
	err := e.db.QueryRow(ctx, `
		SELECT s.start_date, g.end_date
		FROM sprints s
		JOIN global_objectives g ON g.id = s.go_id
		WHERE s.id = $1
	`, sprintID).Scan(&startDate, &endDate)
	if err != nil {
		return fmt.Errorf("freeze trajectory: load dates: %w", err)
	}

	now := time.Now().UTC()
	total := endDate.Sub(startDate).Hours()
	elapsed := now.Sub(startDate).Hours()
	currentExpected := 0.0
	if total > 0 {
		currentExpected = math.Min(elapsed/total, 1.0)
	}

	_, err = e.db.Exec(ctx, `
		UPDATE sprints
		SET expected_pct_frozen = TRUE,
		    frozen_expected_pct = $2
		WHERE id = $1 AND expected_pct_frozen = FALSE
	`, sprintID, currentExpected)
	return err
}

// UnfreezeExpectedTrajectory resumes normal trajectory tracking when Stabilization
// Mode ends (reactivation protocol complete). This allows expectedPct to catch up
// from the frozen value rather than jumping to the current real time position.
func (e *Engine) UnfreezeExpectedTrajectory(ctx context.Context, sprintID uuid.UUID) error {
	_, err := e.db.Exec(ctx, `
		UPDATE sprints
		SET expected_pct_frozen = FALSE,
		    frozen_expected_pct = NULL
		WHERE id = $1
	`, sprintID)
	return err
}

// ApplyEvolveOverride applies behavior model-specific evolution logic to hybrid GO (Global Objectives).
// For goals with a dominant_behavior_model set, this override adjusts the evolution sprint calculation
// to account for behavioral characteristics:
// - ANALYTIC: requires higher consistency (0.75+) for evolution detection
// - STRATEGIC: balanced approach, standard evolution thresholds
// - TACTICAL: lower threshold (0.02 delta), more responsive to quick wins
// - REACTIVE: adaptive threshold based on recent performance volatility
// Returns true if evolution override was applied, false if standard logic applies.
func (e *Engine) ApplyEvolveOverride(ctx context.Context, goalID uuid.UUID, baseScore, prevScore float64) (bool, error) {
	var behaviorModel *string
	err := e.db.QueryRow(ctx, `
		SELECT dominant_behavior_model FROM global_objectives WHERE id = $1
	`, goalID).Scan(&behaviorModel)
	if err != nil {
		return false, nil // No behavior model set — use standard logic
	}

	if behaviorModel == nil {
		return false, nil // Auto-detected behavior model — standard logic
	}

	delta := baseScore - prevScore

	switch *behaviorModel {
	case "ANALYTIC":
		// ANALYTIC goals require higher consistency before marking evolution
		consistency := e.computeConsistencyForGoal(ctx, goalID)
		if delta >= 0.05 && consistency >= 0.75 {
			return true, nil // Evolution qualifies with high consistency
		}

	case "TACTICAL":
		// TACTICAL goals are more responsive to incremental progress
		if delta >= 0.02 {
			return true, nil // Faster evolution recognition
		}

	case "REACTIVE":
		// REACTIVE goals adapt threshold based on recent volatility
		var volatility float64
		_ = e.db.QueryRow(ctx, `
			SELECT COALESCE(STDDEV(sr.score_value), 0)
			FROM sprints s
			JOIN sprint_results sr ON sr.sprint_id = s.id
			WHERE s.go_id = $1 AND s.status = 'COMPLETED'
			ORDER BY s.sprint_number DESC LIMIT 5
		`, goalID).Scan(&volatility)

		// Adaptive threshold: lower when consistent, higher when volatile
		threshold := 0.05
		if volatility > 0.15 {
			threshold = 0.08
		} else if volatility < 0.05 {
			threshold = 0.03
		}

		if delta >= threshold {
			return true, nil
		}

	case "STRATEGIC":
		fallthrough
	default:
		// STRATEGIC uses standard threshold
		if delta >= 0.05 {
			return true, nil
		}
	}

	return false, nil
}
