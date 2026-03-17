package engine

// Level 3 — Adaptive Intelligence (C26-C31)
// Analizează consistența, contextul și adaptează scorurile în timp.
// Logica de penalizare și bonus rămâne strict internă.

import (
	"context"

	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/models"
)

// C26 — computeConsistency: distribuția zilnică a completărilor în sprint
// Un scor bun = sarcini completate uniform, nu totul la final
func (e *Engine) computeConsistency(ctx context.Context, sprint *models.Sprint) float64 {
	var activeDays, totalDays int
	e.db.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT task_date) FILTER (WHERE completed = TRUE),
			COUNT(DISTINCT task_date)
		FROM daily_tasks
		WHERE sprint_id = $1
		  AND task_type = 'MAIN'
		  AND task_date <= CURRENT_DATE
	`, sprint.ID).Scan(&activeDays, &totalDays)

	if totalDays == 0 {
		return 0
	}
	return float64(activeDays) / float64(totalDays)
}

// C27 — computeContextFactors: penalizare/bonus bazat pe ajustările de context
// Pauza planificată NU penalizează — onestitatea e recompensată
func (e *Engine) computeContextFactors(ctx context.Context, goalID uuid.UUID) (penalty, bonus float64) {
	adjs, _ := db.GetActiveAdjustments(ctx, e.db, goalID)
	for _, a := range adjs {
		switch a.AdjType {
		case models.AdjEnergyHigh:
			bonus = 0.1
		case models.AdjEnergyLow:
			penalty = 0.03 // Penalizare mică — userul a fost onest
		case models.AdjPause:
			// Pauza planificată NU penalizează
		}
	}
	return
}
