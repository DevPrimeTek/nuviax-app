package engine

// Level 1 — Structural Authority (C9-C18)
// Gestionează generarea sarcinilor, intensitatea și structura execuției zilnice.
// Niciun detaliu nu iese din acest fișier în afara package-ului.

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/models"
)

// C9 — computeIntensity: determină intensitatea zilei bazată pe context
// Cu ajustări multiple, AdjEnergyLow are prioritate maximă (siguranță) față de AdjEnergyHigh.
func (e *Engine) computeIntensity(adjs []models.ContextAdjustment) float64 {
	base := 1.0
	hasLow := false
	hasHigh := false
	for _, a := range adjs {
		switch a.AdjType {
		case models.AdjEnergyLow:
			hasLow = true
		case models.AdjEnergyHigh:
			hasHigh = true
		}
	}
	// AdjEnergyLow are prioritate — dacă utilizatorul e obosit, reducem intensitatea
	if hasLow {
		base = 0.6
	} else if hasHigh {
		base = 1.2
	}
	return base
}

// C10 — taskCountFromIntensity: mapează intensitatea la numărul de sarcini (1-3)
func (e *Engine) taskCountFromIntensity(intensity float64) int {
	switch {
	case intensity >= 1.2:
		return 3
	case intensity >= 1.0:
		return 2
	default:
		return 1
	}
}

// C11 — generateTasksFromCheckpoints: generează sarcinile zilnice din checkpoint-ul activ
func (e *Engine) generateTasksFromCheckpoints(
	ctx context.Context,
	checkpoints []models.Checkpoint,
	goal models.Goal,
	sprint *models.Sprint,
	userID uuid.UUID,
	date time.Time,
	count int,
) []models.DailyTask {
	var tasks []models.DailyTask
	activeCP := findActiveCheckpoint(checkpoints)
	if activeCP == nil {
		return tasks
	}

	taskTexts := e.generateTaskTexts(goal, *activeCP, count)
	for i, text := range taskTexts {
		t, err := db.CreateTask(ctx, e.db,
			sprint.ID, goal.ID, userID, date,
			text, models.TaskMain, i)
		if err == nil {
			tasks = append(tasks, *t)
		}
	}
	return tasks
}

// C12 — generateTaskTexts: construiește textele sarcinilor din template-uri contextuale
// Dacă AI (Claude Haiku) este disponibil, generează sarcini contextualizate.
// Fallback: template-uri statice (B-4 fix).
func (e *Engine) generateTaskTexts(goal models.Goal, cp models.Checkpoint, count int) []string {
	if e.ai != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		tasks, err := e.ai.GenerateTaskTexts(ctx, goal.Name, cp.Name, 1, count)
		if err == nil && len(tasks) > 0 {
			return tasks
		}
		// AI failed — fall through to static templates
	}
	// Static fallback templates
	base := cp.Name
	templates := []string{
		"Lucrează 30 min la: " + base,
		"Avansează concret cu: " + base,
		"Finalizează un pas din: " + base,
	}
	if count > len(templates) {
		count = len(templates)
	}
	return templates[:count]
}

// C13 — findActiveCheckpoint: găsește checkpoint-ul activ sau primul disponibil
func findActiveCheckpoint(cps []models.Checkpoint) *models.Checkpoint {
	for i := range cps {
		if cps[i].Status == models.CheckpointInProgress ||
			cps[i].Status == models.CheckpointUpcoming {
			return &cps[i]
		}
	}
	return nil
}

// ── GAP G-1 — Deadline Recalcul After Pause ─────────────────────────────────

// ExtendSprintForPause extends the active sprint's end_date by pauseDays (G-1).
// This ensures the sprint deadline is fair after a planned absence — the user
// doesn't lose sprint time to a pause they reported honestly.
// Only affects ACTIVE sprints; completed/skipped sprints are not extended.
func (e *Engine) ExtendSprintForPause(ctx context.Context, goalID uuid.UUID, pauseDays int) error {
	if pauseDays <= 0 {
		return nil
	}

	sprint, err := db.GetCurrentSprint(ctx, e.db, goalID)
	if err != nil {
		return nil // no active sprint — nothing to extend
	}

	_, err = e.db.Exec(ctx, `
		UPDATE sprints
		SET end_date = end_date + ($2 * INTERVAL '1 day')
		WHERE id = $1 AND status = 'ACTIVE'
	`, sprint.ID, pauseDays)
	return err
}
