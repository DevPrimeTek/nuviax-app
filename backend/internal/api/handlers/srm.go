package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/api/middleware"
	"github.com/devprimetek/nuviax-app/internal/db"
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

	return c.JSON(fiber.Map{
		"goal_id":   goalID,
		"srm_level": srmLevel,
		"message":   srmMessage(srmLevel),
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
	if _, err := db.GetGoalByID(c.Context(), h.db, goalID, userID); err != nil {
		return notFound(c)
	}

	// TODO: engine.ConfirmSRML3(ctx, goalID) — triggers C34 Suspension + C35 Stabilization

	return c.JSON(fiber.Map{
		"goal_id": goalID,
		"message": "SRM Level 3 confirmat. Modul de stabilizare activat.",
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
