package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/devprimetek/nuviax-app/internal/api/middleware"
	"github.com/devprimetek/nuviax-app/internal/db"
)

type activityEntry struct {
	Date           string `json:"date"`
	TasksCompleted int    `json:"tasks_completed"`
	ActiveMinutes  int    `json:"active_minutes"`
}

// GET /settings — returns user display settings (name, locale, theme).
func (h *Handlers) GetSettings(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	user, err := db.GetUserByID(c.Context(), h.db, userID)
	if err != nil {
		return serverError(c, err)
	}
	name := ""
	if user.FullName != nil {
		name = *user.FullName
	}
	return c.JSON(fiber.Map{
		"full_name":  name,
		"user_name":  name,
		"locale":     user.Locale,
		"theme":      user.Theme,
		"mfa_enabled": user.MFAEnabled,
	})
}

// GET /profile/activity — returns daily activity metrics for the past 365 days.
func (h *Handlers) GetProfileActivity(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	rows, err := h.db.Query(c.Context(), `
		SELECT recorded_at::date::text, tasks_completed, active_minutes
		FROM daily_metrics
		WHERE user_id=$1 AND recorded_at >= NOW() - INTERVAL '365 days'
		ORDER BY recorded_at ASC
	`, userID)
	if err != nil {
		return serverError(c, err)
	}
	defer rows.Close()

	entries := make([]activityEntry, 0)
	for rows.Next() {
		var e activityEntry
		if err := rows.Scan(&e.Date, &e.TasksCompleted, &e.ActiveMinutes); err != nil {
			return serverError(c, err)
		}
		entries = append(entries, e)
	}

	return c.JSON(entries)
}

// PATCH /settings — updates user theme and/or locale preference.
func (h *Handlers) UpdateSettings(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req struct {
		Theme  *string `json:"theme"`
		Locale *string `json:"locale"`
	}
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	if req.Theme == nil && req.Locale == nil {
		return c.JSON(fiber.Map{"ok": true})
	}

	if req.Theme != nil && *req.Theme != "light" && *req.Theme != "dark" {
		return badRequest(c, "Temă invalidă: 'light' sau 'dark'.")
	}
	if req.Locale != nil {
		switch *req.Locale {
		case "ro", "en", "ru":
		default:
			return badRequest(c, "Localizare invalidă: 'ro', 'en' sau 'ru'.")
		}
	}

	if req.Theme != nil {
		h.db.Exec(c.Context(), `
			UPDATE users
			SET preferences = COALESCE(preferences, '{}') || jsonb_build_object('theme', $1::text),
			    updated_at = NOW()
			WHERE id=$2
		`, *req.Theme, userID)
	}
	if req.Locale != nil {
		h.db.Exec(c.Context(), `
			UPDATE users
			SET preferences = COALESCE(preferences, '{}') || jsonb_build_object('locale', $1::text),
			    updated_at = NOW()
			WHERE id=$2
		`, *req.Locale, userID)
	}

	return c.JSON(fiber.Map{"ok": true})
}
