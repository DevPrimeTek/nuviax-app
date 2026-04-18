package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/api/middleware"
)

// GET /srm/status/:goalId — returns latest SRM level for a goal (C33).
func (h *Handlers) GetSRMStatus(c *fiber.Ctx) error {
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

	var level string
	var triggeredAt time.Time
	err = h.db.QueryRow(c.Context(), `
		SELECT level, created_at FROM srm_events
		WHERE go_id=$1 ORDER BY created_at DESC LIMIT 1
	`, goalID).Scan(&level, &triggeredAt)
	if err != nil {
		return c.JSON(fiber.Map{"srm_level": "NONE", "triggered_at": nil})
	}

	return c.JSON(fiber.Map{"srm_level": level, "triggered_at": triggeredAt})
}

// POST /srm/confirm-l2/:goalId — user acknowledges L2 intervention (ENERGY_LOW).
func (h *Handlers) ConfirmSRML2(c *fiber.Ctx) error {
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

	h.db.Exec(c.Context(), `
		UPDATE srm_events SET confirmed_at=NOW()
		WHERE go_id=$1 AND level='L2' AND confirmed_at IS NULL
	`, goalID)

	h.db.Exec(c.Context(), `
		INSERT INTO context_adjustments (go_id, user_id, type)
		VALUES ($1, $2, 'ENERGY_LOW')
	`, goalID, userID)

	return c.JSON(fiber.Map{"ok": true})
}

// POST /srm/confirm-l3/:goalId — user confirms L3: goal is paused (C32 Pause).
func (h *Handlers) ConfirmSRML3(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("goalId"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	tag, err := h.db.Exec(c.Context(), `
		UPDATE global_objectives SET status='PAUSED', paused_at=NOW(), updated_at=NOW()
		WHERE id=$1 AND user_id=$2
	`, goalID, userID)
	if err != nil || tag.RowsAffected() == 0 {
		return notFound(c)
	}

	h.db.Exec(c.Context(), `
		INSERT INTO context_adjustments (go_id, user_id, type)
		VALUES ($1, $2, 'PAUSE')
	`, goalID, userID)

	return c.JSON(fiber.Map{"ok": true})
}
