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
// Faza 1: template-uri cu context din goal; Faza 2: AI-assisted generation
func (e *Engine) generateTaskTexts(goal models.Goal, cp models.Checkpoint, count int) []string {
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
