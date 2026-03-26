package engine

// Level 4 — Regulatory Authority (C32-C36)
// Aplică regulile de business pentru activarea obiectivelor.
// Limitele și regulile sunt opace față de utilizator.

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/models"
)

// C32 — validateActivation: verifică dacă un obiectiv poate fi activat
// Returnează: (poate fi activat, mesaj dacă nu poate / avertisment dacă da)
func (e *Engine) validateActivation(ctx context.Context, userID uuid.UUID, newGoal *models.Goal) (bool, string) {
	// Regula 0: Validare date — StartDate trebuie să fie înainte de EndDate
	if !newGoal.StartDate.Before(newGoal.EndDate) {
		return false, "Data de start trebuie să fie înainte de data de final."
	}

	// Regula 1: Max 3 obiective active simultan — depășire → Vault (G-10)
	activeCount, err := db.CountActiveGoals(ctx, e.db, userID)
	if err != nil {
		return false, "Eroare internă la verificarea obiectivelor active. Încearcă din nou."
	}
	if activeCount >= 3 {
		// G-10: Instead of hard blocking, signal handler to vault the goal
		return true, vaultMessage
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

// ── GAP G-7 — Reactivation Protocol ────────────────────────────────────────

// reactivationStabilityDays — days without SRM events before reactivation is proposed (G-7)
const reactivationStabilityDays = 7

// CheckReactivationEligibility checks if a PAUSED goal can be proposed for reactivation (G-7).
// A goal is eligible when it has been in SRM L3 for >= 7 days without new SRM events.
// Returns: eligible, daysSinceLastSRM, error
func (e *Engine) CheckReactivationEligibility(ctx context.Context, goalID uuid.UUID) (bool, int, error) {
	var triggeredAt time.Time
	err := e.db.QueryRow(ctx, `
		SELECT triggered_at FROM srm_events
		WHERE go_id = $1 AND srm_level = 'L3' AND revoked_at IS NULL
		ORDER BY triggered_at DESC LIMIT 1
	`, goalID).Scan(&triggeredAt)
	if err != nil {
		return false, 0, nil // No L3 event — not in stabilization
	}

	daysSince := int(time.Now().UTC().Sub(triggeredAt).Hours() / 24)
	if daysSince < reactivationStabilityDays {
		return false, daysSince, nil
	}

	// Ensure no newer SRM events after the L3 trigger
	var recentCount int
	e.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM srm_events
		WHERE go_id = $1 AND triggered_at > $2 AND revoked_at IS NULL
	`, goalID, triggeredAt).Scan(&recentCount)

	if recentCount > 0 {
		return false, daysSince, nil
	}

	return true, daysSince, nil
}

// ProposeReactivation inserts a reactivation protocol record for an eligible PAUSED goal (G-7).
// The scheduler calls this; the protocol ramps intensity from 0.2 → 1.0 over ~8 days.
func (e *Engine) ProposeReactivation(ctx context.Context, goalID uuid.UUID) error {
	// Avoid duplicate proposals
	var existing int
	e.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM reactivation_protocols
		WHERE go_id = $1 AND completed_at IS NULL
	`, goalID).Scan(&existing)
	if existing > 0 {
		return nil // Already in progress
	}

	_, err := e.db.Exec(ctx, `
		INSERT INTO reactivation_protocols
			(id, go_id, current_day, current_intensity, srm1_disabled, srm2_threshold_adjusted, started_at)
		VALUES (gen_random_uuid(), $1, 1, 0.2, TRUE, TRUE, NOW())
		ON CONFLICT (go_id) DO UPDATE
		  SET current_day = 1, current_intensity = 0.2,
		      srm1_disabled = TRUE, srm2_threshold_adjusted = TRUE,
		      started_at = NOW(), completed_at = NULL, updated_at = NOW()
	`, goalID)
	return err
}

// ── GAP G-10 — Future Vault ────────────────────────────────────────────────

// vaultMessage is the special marker returned when a goal should go to Vault (G-10).
// The handler detects this prefix and creates the goal as WAITING instead of blocking.
const vaultMessage = "VAULT: Ai deja 3 obiective active. Noul obiectiv a fost adăugat în Vault-ul viitor și va fi activat automat când un slot se eliberează."

// ShouldVaultNewGoal returns true when the user is already at max active goals (G-10).
// Used by handlers to redirect new goals to WAITING (Vault) instead of blocking.
func (e *Engine) ShouldVaultNewGoal(ctx context.Context, userID uuid.UUID) (bool, error) {
	activeCount, err := db.CountActiveGoals(ctx, e.db, userID)
	if err != nil {
		return false, err
	}
	return activeCount >= 3, nil
}
