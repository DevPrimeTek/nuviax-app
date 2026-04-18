package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/pkg/crypto"
)

// GET /admin/stats — platform-wide counters. Requires admin (404 for non-admin via middleware).
func (h *Handlers) AdminStats(c *fiber.Ctx) error {
	var usersTotal, activeGoals, tasksToday int

	h.db.QueryRow(c.Context(), `SELECT COUNT(*) FROM users WHERE is_active=TRUE`).Scan(&usersTotal)
	h.db.QueryRow(c.Context(), `SELECT COUNT(*) FROM global_objectives WHERE status='ACTIVE'`).Scan(&activeGoals)
	h.db.QueryRow(c.Context(), `SELECT COUNT(*) FROM daily_tasks WHERE task_date=CURRENT_DATE`).Scan(&tasksToday)

	return c.JSON(fiber.Map{
		"users_total":  usersTotal,
		"active_goals": activeGoals,
		"tasks_today":  tasksToday,
	})
}

type adminUserRow struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	FullName   *string   `json:"full_name"`
	CreatedAt  time.Time `json:"created_at"`
	IsActive   bool      `json:"is_active"`
	GoalsCount int       `json:"goals_count"`
}

// GET /admin/users — list all users with decrypted email. Requires admin.
func (h *Handlers) AdminUsers(c *fiber.Ctx) error {
	rows, err := h.db.Query(c.Context(), `
		SELECT u.id, u.email_encrypted, u.full_name, u.created_at, u.is_active,
		       COUNT(g.id) AS goals_count
		FROM users u
		LEFT JOIN global_objectives g ON g.user_id = u.id
		GROUP BY u.id, u.email_encrypted, u.full_name, u.created_at, u.is_active
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		return serverError(c, err)
	}
	defer rows.Close()

	users := make([]adminUserRow, 0)
	for rows.Next() {
		var u adminUserRow
		var encEmail string
		if err := rows.Scan(&u.ID, &encEmail, &u.FullName, &u.CreatedAt, &u.IsActive, &u.GoalsCount); err != nil {
			return serverError(c, err)
		}
		if dec, err := crypto.Decrypt(encEmail, h.encKey); err == nil {
			u.Email = dec
		}
		users = append(users, u)
	}

	return c.JSON(users)
}

// POST /admin/users/:id/deactivate — soft-deletes a user account. Requires admin.
func (h *Handlers) AdminDeactivateUser(c *fiber.Ctx) error {
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID invalid.")
	}

	h.db.Exec(c.Context(),
		`UPDATE users SET is_active=FALSE, updated_at=NOW() WHERE id=$1`,
		targetID)

	return c.JSON(fiber.Map{"ok": true})
}
