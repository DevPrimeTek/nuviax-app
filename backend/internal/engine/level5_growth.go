package engine

// Level 5 — Growth Orchestration (C37-C40)
// Calculează progresul față de traiectoria așteptată și componentele de creștere.
// Valorile intermediare nu sunt niciodată expuse în afara engine-ului.

import (
	"context"
	"math"
	"time"

	"github.com/devprimetek/nuviax-app/internal/models"
)

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
