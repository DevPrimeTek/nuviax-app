package engine

// Level 4 — Regulatory Authority (C32-C36)
// Aplică regulile de business pentru activarea obiectivelor.
// Limitele și regulile sunt opace față de utilizator.

import (
	"context"

	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/models"
)

// C32 — validateActivation: verifică dacă un obiectiv poate fi activat
// Returnează: (poate fi activat, mesaj dacă nu poate / avertisment dacă da)
func (e *Engine) validateActivation(ctx context.Context, userID uuid.UUID, newGoal *models.Goal) (bool, string) {
	// Regula 1: Max 3 obiective active simultan
	activeCount, err := db.CountActiveGoals(ctx, e.db, userID)
	if err != nil || activeCount >= 3 {
		return false, "Poți lucra la maxim 3 obiective în același timp."
	}

	// Regula 2: Durată maximă 365 zile
	duration := newGoal.EndDate.Sub(newGoal.StartDate)
	if duration.Hours()/24 > 365 {
		return false, "Un obiectiv nu poate dura mai mult de 365 de zile."
	}

	// Regula 3: Verificare conflict temporal cu obiectivele existente
	existingGoals, _ := db.GetGoalsByUser(ctx, e.db, userID)
	for _, g := range existingGoals {
		if g.Status != models.GoalActive {
			continue
		}
		if e.hasTemporalOverlap(g, *newGoal) {
			// Nu blocăm — avertizăm (utilizatorul decide)
			return true, "Atenție: poate suprapune resurse cu un obiectiv existent."
		}
	}

	return true, ""
}

// C33 — hasTemporalOverlap: detectează suprapunerea temporală între două obiective
func (e *Engine) hasTemporalOverlap(existing, newGoal models.Goal) bool {
	return existing.StartDate.Before(newGoal.EndDate) &&
		newGoal.StartDate.Before(existing.EndDate)
}
