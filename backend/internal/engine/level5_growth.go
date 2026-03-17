package engine

// Level 5 — Growth Orchestration (C37-C40)
// Calculează progresul față de traiectoria așteptată și componentele de creștere.
// Valorile intermediare nu sunt niciodată expuse în afara engine-ului.

import (
	"context"
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

// C37 — computeProgressVsExpected: progresul real față de traiectoria liniară așteptată
// Ratio > 1 = înaintea planului, < 1 = în urmă
func (e *Engine) computeProgressVsExpected(goal *models.Goal, sprint *models.Sprint) float64 {
	now := time.Now().UTC()
	totalDuration := goal.EndDate.Sub(goal.StartDate).Hours()
	elapsed := now.Sub(goal.StartDate).Hours()
	if totalDuration <= 0 {
		return 0
	}
	expectedPct := elapsed / totalDuration

	var completedCP, totalCP int
	e.db.QueryRow(context.Background(), `
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
	ratio := actualPct / math.Max(expectedPct, 0.01)
	return clamp(ratio, 0, 1.2) / 1.2
}
