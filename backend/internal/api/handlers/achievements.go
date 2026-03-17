package handlers

import (
	"github.com/gofiber/fiber/v2"
	pgx "github.com/jackc/pgx/v5"

	"github.com/devprimetek/nuviax-app/internal/api/middleware"
	"github.com/devprimetek/nuviax-app/internal/models"
)

// ── GetUserAchievements — GET /api/v1/achievements ───────────────
func (h *Handlers) GetUserAchievements(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	achievements, err := h.engine.GetUserAchievements(c.Context(), userID)
	if err != nil {
		return serverError(c, err)
	}

	if achievements == nil {
		achievements = []models.AchievementBadge{}
	}
	return c.JSON(fiber.Map{"achievements": achievements})
}

// ── GetAchievementProgress — GET /api/v1/achievements/progress ───
func (h *Handlers) GetAchievementProgress(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	rows, err := h.db.Query(c.Context(), `
		SELECT * FROM get_achievement_progress($1)
	`, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Eroare la citirea progresului."})
	}

	progress, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		return serverError(c, err)
	}

	if progress == nil {
		progress = []map[string]any{}
	}
	return c.JSON(fiber.Map{"progress": progress})
}
