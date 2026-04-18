package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/api/middleware"
)

type achievementItem struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EarnedAt    time.Time `json:"earned_at"`
}

// GET /achievements — list all achievements earned by the user.
func (h *Handlers) ListAchievements(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	rows, err := h.db.Query(c.Context(), `
		SELECT id, type, title, description, earned_at
		FROM achievements
		WHERE user_id=$1
		ORDER BY earned_at DESC
	`, userID)
	if err != nil {
		return serverError(c, err)
	}
	defer rows.Close()

	items := make([]achievementItem, 0)
	for rows.Next() {
		var a achievementItem
		if err := rows.Scan(&a.ID, &a.Type, &a.Title, &a.Description, &a.EarnedAt); err != nil {
			return serverError(c, err)
		}
		items = append(items, a)
	}

	return c.JSON(items)
}

// GET /ceremonies/:goalId — returns the latest ceremony for a goal, or null.
func (h *Handlers) GetCeremony(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("goalId"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	var ownerCheck uuid.UUID
	if err := h.db.QueryRow(c.Context(),
		`SELECT user_id FROM global_objectives WHERE id=$1 AND user_id=$2`,
		goalID, userID).Scan(&ownerCheck); err != nil {
		return notFound(c)
	}

	var id uuid.UUID
	var tier string
	var sprintScore float64
	var viewedAt *time.Time

	err = h.db.QueryRow(c.Context(), `
		SELECT id, tier, sprint_score, viewed_at
		FROM ceremonies
		WHERE go_id=$1
		ORDER BY created_at DESC LIMIT 1
	`, goalID).Scan(&id, &tier, &sprintScore, &viewedAt)
	if err != nil {
		return c.JSON(nil)
	}

	return c.JSON(fiber.Map{
		"id":           id,
		"tier":         tier,
		"sprint_score": sprintScore,
		"viewed_at":    viewedAt,
	})
}

// POST /ceremonies/:id/view — marks a ceremony as viewed.
func (h *Handlers) ViewCeremony(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	ceremonyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	tag, err := h.db.Exec(c.Context(), `
		UPDATE ceremonies SET viewed_at=COALESCE(viewed_at, NOW())
		WHERE id=$1
		  AND go_id IN (SELECT id FROM global_objectives WHERE user_id=$2)
	`, ceremonyID, userID)
	if err != nil || tag.RowsAffected() == 0 {
		return notFound(c)
	}

	return c.JSON(fiber.Map{"ok": true})
}
