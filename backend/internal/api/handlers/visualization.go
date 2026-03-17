package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/api/middleware"
	"github.com/devprimetek/nuviax-app/internal/db"
)

// ── GetProgressVisualization — GET /api/v1/goals/:id/visualize ───
func (h *Handlers) GetProgressVisualization(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID obiectiv invalid."})
	}

	// Verifică dreptul de acces
	if _, err := db.GetGoalByID(c.Context(), h.db, goalID, userID); err != nil {
		return notFound(c)
	}

	viz, err := h.engine.GenerateProgressVisualization(c.Context(), goalID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Eroare la generarea vizualizării."})
	}

	return c.JSON(viz)
}
