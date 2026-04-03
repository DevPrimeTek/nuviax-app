package handlers

import (
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/api/middleware"
	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/models"
)

// ── GetSRMStatus — GET /api/v1/srm/status/:goalId ────────────────
func (h *Handlers) GetSRMStatus(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("goalId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID obiectiv invalid."})
	}

	// Verifică dreptul de acces
	if _, err := db.GetGoalByID(c.Context(), h.db, goalID, userID); err != nil {
		return notFound(c)
	}

	var srmLevel string
	err = h.db.QueryRow(c.Context(), `
		SELECT COALESCE(
			(SELECT srm_level::text
			 FROM srm_events
			 WHERE go_id = $1 AND revoked_at IS NULL
			 ORDER BY triggered_at DESC
			 LIMIT 1),
			'NONE'
		)
	`, goalID).Scan(&srmLevel)
	if err != nil {
		srmLevel = "NONE"
	}

	// GAP #8/#13 — Include ALI current vs projected breakdown to eliminate ambiguity
	aliInfo := h.computeALIBreakdown(c, userID, goalID)

	return c.JSON(fiber.Map{
		"goal_id":   goalID,
		"srm_level": srmLevel,
		"message":   srmMessage(srmLevel),
		"ali":       aliInfo,
	})
}

// ── ConfirmSRML2 — POST /api/v1/srm/confirm-l2/:goalId ───────────
// GAP G-12 — L2 requires single user confirmation to apply structural recalibration.
// L2 does NOT pause the goal — it reduces task intensity and schedules recalibration.
func (h *Handlers) ConfirmSRML2(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("goalId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID obiectiv invalid."})
	}

	// Verify access
	if _, err := db.GetGoalByID(c.Context(), h.db, goalID, userID); err != nil {
		return notFound(c)
	}

	// Mark the most recent pending L2 event as confirmed
	result, err := h.db.Exec(c.Context(), `
		UPDATE srm_events
		SET confirmed_at = NOW(), confirmed_by = $2
		WHERE id = (
			SELECT id FROM srm_events
			WHERE go_id = $1 AND srm_level = 'L2'
			  AND revoked_at IS NULL AND confirmed_at IS NULL
			ORDER BY triggered_at DESC
			LIMIT 1
		)
	`, goalID, userID)
	if err != nil {
		return serverError(c, err)
	}
	if result.RowsAffected() == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Niciun eveniment SRM L2 activ de confirmat."})
	}

	tomorrow := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
	endDate := tomorrow.AddDate(0, 0, 7)
	note := "srm_l2_confirm"
	if _, err := db.CreateContextAdjustment(c.Context(), h.db, goalID, userID,
		models.AdjEnergyLow, tomorrow, &endDate, &note); err != nil {
		return serverError(c, err)
	}

	return c.JSON(fiber.Map{
		"goal_id":   goalID,
		"message":   "SRM Level 2 confirmat. Recalibrare structurală aplicată. Intensitatea sarcinilor va fi ajustată.",
		"next_step": "Continuă activitățile la ritm redus. Dacă situația nu se îmbunătățește, SRM L3 poate fi necesar.",
	})
}

// ── ConfirmSRML3 — POST /api/v1/srm/confirm-l3/:goalId ───────────
func (h *Handlers) ConfirmSRML3(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("goalId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID obiectiv invalid."})
	}

	// Verifică dreptul de acces
	goal, err := db.GetGoalByID(c.Context(), h.db, goalID, userID)
	if err != nil {
		return notFound(c)
	}

	// Get current sprint to freeze its expected trajectory (GAP #20)
	sprint, sprintErr := db.GetCurrentSprint(c.Context(), h.db, goalID)

	// Suspend the goal (SRM L3 → status PAUSED)
	if _, err := h.db.Exec(c.Context(), `
		UPDATE global_objectives SET status = 'PAUSED', updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, goalID, userID); err != nil {
		return serverError(c, err)
	}

	// Record the SRM L3 event
	h.db.Exec(c.Context(), `
		INSERT INTO srm_events (id, go_id, srm_level, trigger_reason, triggered_at)
		VALUES (gen_random_uuid(), $1, 'L3', 'user_confirmed_stabilization', NOW())
	`, goalID)

	// GAP #20 — Freeze expected trajectory to prevent drift loop paradox
	frozenPct := 0.0
	if sprintErr == nil && sprint != nil {
		h.engine.FreezeExpectedTrajectory(c.Context(), sprint.ID)

		// Compute the frozen value for the response
		total := goal.EndDate.Sub(goal.StartDate).Hours()
		elapsed := time.Now().UTC().Sub(goal.StartDate).Hours()
		if total > 0 {
			frozenPct = math.Min(elapsed/total, 1.0)
		}
	}

	return c.JSON(fiber.Map{
		"goal_id":         goalID,
		"new_status":      models.GoalPaused,
		"frozen_expected": frozenPct,
		"message":         "SRM Level 3 confirmat. Modul de stabilizare activat. Obiectivul este suspendat temporar.",
		"next_step":       "Reactivarea automată va fi propusă după 7 zile de stabilitate.",
	})
}

// srmMessage returnează un mesaj user-friendly pentru nivelul SRM
func srmMessage(level string) string {
	switch level {
	case "L1":
		return "Ajustare automată activă. Ritmul a fost redus ușor."
	case "L2":
		return "Ajustare structurală în curs. Am recalibrat obiectivele."
	case "L3":
		return "Resetare strategică necesară. Confirmă pentru a activa modul de stabilizare."
	default:
		return "Totul merge bine!"
	}
}

// GAP #8/#13 — computeALIBreakdown returns both current and projected ALI values.
// ALI_current = actual ambition load based on tasks completed so far.
// ALI_projected = projected load if the current pace continues to sprint end.
// This eliminates the ambiguity between "what has been done" vs "what will happen".
// Ambition Buffer zone: ALI_projected between 1.0 and 1.15 triggers Velocity Control.
func (h *Handlers) computeALIBreakdown(c *fiber.Ctx, userID, targetGoalID uuid.UUID) fiber.Map {
	// Get all active goals for the user to compute total load
	goals, err := db.GetGoalsByUser(c.Context(), h.db, userID)
	if err != nil {
		return fiber.Map{"error": "could not compute ALI"}
	}

	type goalALI struct {
		GoalID       uuid.UUID `json:"goal_id"`
		ALICurrent   float64   `json:"ali_current"`
		ALIProjected float64   `json:"ali_projected"`
	}

	var breakdown []goalALI
	totalCurrent := 0.0
	totalProjected := 0.0

	for _, g := range goals {
		if g.Status != models.GoalActive {
			continue
		}

		sprint, err := db.GetCurrentSprint(c.Context(), h.db, g.ID)
		if err != nil || sprint == nil {
			continue
		}

		// ALI_current: tasks completed / tasks expected by now
		var completedTasks, totalTasks int
		h.db.QueryRow(c.Context(), `
			SELECT
				COUNT(*) FILTER (WHERE completed = TRUE),
				COUNT(*)
			FROM daily_tasks
			WHERE sprint_id = $1 AND task_type = 'MAIN'
		`, sprint.ID).Scan(&completedTasks, &totalTasks)

		sprintDuration := sprint.EndDate.Sub(sprint.StartDate).Hours() / 24
		elapsed := time.Now().UTC().Sub(sprint.StartDate).Hours() / 24
		remaining := sprintDuration - elapsed

		aliCurrent := 0.0
		if totalTasks > 0 && elapsed > 0 {
			dailyRate := float64(completedTasks) / math.Max(elapsed, 1)
			expectedByNow := math.Round(elapsed) * (float64(totalTasks) / sprintDuration)
			if expectedByNow > 0 {
				aliCurrent = dailyRate / (float64(totalTasks) / sprintDuration)
			}
		}

		// ALI_projected: if current pace continues to end of sprint
		aliProjected := aliCurrent
		if remaining > 0 && elapsed > 0 {
			dailyRate := float64(completedTasks) / math.Max(elapsed, 1)
			projectedTotal := float64(completedTasks) + dailyRate*remaining
			expectedTotal := float64(totalTasks)
			if expectedTotal > 0 {
				aliProjected = projectedTotal / expectedTotal
			}
		}

		breakdown = append(breakdown, goalALI{
			GoalID:       g.ID,
			ALICurrent:   math.Round(aliCurrent*1000) / 1000,
			ALIProjected: math.Round(aliProjected*1000) / 1000,
		})

		totalCurrent += aliCurrent
		totalProjected += aliProjected
	}

	activeGoalCount := float64(len(breakdown))
	if activeGoalCount > 0 {
		totalCurrent /= activeGoalCount
		totalProjected /= activeGoalCount
	}

	inAmbitionBuffer := totalProjected > 1.0 && totalProjected <= 1.15
	velocityControlOn := totalProjected > 1.15

	return fiber.Map{
		"ali_current":         math.Round(totalCurrent*1000) / 1000,
		"ali_projected":       math.Round(totalProjected*1000) / 1000,
		"in_ambition_buffer":  inAmbitionBuffer,
		"velocity_control_on": velocityControlOn,
		"goal_breakdown":      breakdown,
		// GAP #13 disambiguation note in response
		"note": "ali_current = progres realizat până acum. ali_projected = proiecție la finalul sprintului.",
	}
}
