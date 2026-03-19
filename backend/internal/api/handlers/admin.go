package handlers

// Admin Panel Handlers — /api/v1/admin/*
// All endpoints require is_admin = TRUE (enforced by AdminOnly middleware).
// Dev-only endpoints (DB reset) additionally check APP_ENV environment variable.

import (
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/api/middleware"
	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/models"
	"github.com/devprimetek/nuviax-app/pkg/crypto"
)

// ── GET /api/v1/admin/stats ───────────────────────────────────────
// Returns aggregated platform statistics (users, goals, sprints, tasks, SRM events).
func (h *Handlers) AdminGetStats(c *fiber.Ctx) error {
	stats, err := db.GetPlatformStats(c.Context(), h.db)
	if err != nil {
		return serverError(c, err)
	}
	return c.JSON(stats)
}

// ── GET /api/v1/admin/users ───────────────────────────────────────
// Returns the full user list with per-user statistics.
func (h *Handlers) AdminGetUsers(c *fiber.Ctx) error {
	users, err := db.GetAdminUserList(c.Context(), h.db)
	if err != nil {
		return serverError(c, err)
	}
	if users == nil {
		users = []models.AdminUserRecord{}
	}
	return c.JSON(fiber.Map{
		"users": users,
		"total": len(users),
	})
}

// ── GET /api/v1/admin/audit?limit=100 ────────────────────────────
// Returns the most recent audit log entries.
func (h *Handlers) AdminGetAuditLog(c *fiber.Ctx) error {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	entries, err := db.GetAuditLog(c.Context(), h.db, limit)
	if err != nil {
		return serverError(c, err)
	}
	if entries == nil {
		entries = []map[string]interface{}{}
	}
	return c.JSON(fiber.Map{
		"entries": entries,
		"limit":   limit,
	})
}

// ── POST /api/v1/admin/users/:id/deactivate ──────────────────────
// Deactivates a user account (prevents login but preserves data).
func (h *Handlers) AdminDeactivateUser(c *fiber.Ctx) error {
	adminID := middleware.GetUserID(c)
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID utilizator invalid.")
	}

	if targetID == adminID {
		return badRequest(c, "Nu poți dezactiva propriul cont.")
	}

	if err := db.SetUserActiveStatus(c.Context(), h.db, targetID, false); err != nil {
		return serverError(c, err)
	}

	db.WriteAudit(c.Context(), h.db, &adminID, "ADMIN_DEACTIVATE_USER",
		crypto.SHA256Hex(c.IP()), crypto.SHA256Hex(c.Get("User-Agent")))

	return c.JSON(fiber.Map{
		"message":   "Utilizatorul a fost dezactivat.",
		"target_id": targetID,
	})
}

// ── POST /api/v1/admin/users/:id/activate ────────────────────────
// Re-activates a previously deactivated user account.
func (h *Handlers) AdminActivateUser(c *fiber.Ctx) error {
	adminID := middleware.GetUserID(c)
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID utilizator invalid.")
	}

	if err := db.SetUserActiveStatus(c.Context(), h.db, targetID, true); err != nil {
		return serverError(c, err)
	}

	db.WriteAudit(c.Context(), h.db, &adminID, "ADMIN_ACTIVATE_USER",
		crypto.SHA256Hex(c.IP()), crypto.SHA256Hex(c.Get("User-Agent")))

	return c.JSON(fiber.Map{
		"message":   "Utilizatorul a fost reactivat.",
		"target_id": targetID,
	})
}

// ── POST /api/v1/admin/users/:id/promote ─────────────────────────
// Promotes a regular user to admin status.
func (h *Handlers) AdminPromoteUser(c *fiber.Ctx) error {
	adminID := middleware.GetUserID(c)
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return badRequest(c, "ID utilizator invalid.")
	}

	if err := db.PromoteToAdmin(c.Context(), h.db, targetID); err != nil {
		return serverError(c, err)
	}

	db.WriteAudit(c.Context(), h.db, &adminID, "ADMIN_PROMOTE_ADMIN",
		crypto.SHA256Hex(c.IP()), crypto.SHA256Hex(c.Get("User-Agent")))

	return c.JSON(fiber.Map{
		"message":   "Utilizatorul a primit drepturi de admin.",
		"target_id": targetID,
	})
}

// ── POST /api/v1/admin/db/reset ───────────────────────────────────
// DEV ONLY — Deletes all non-admin user data and resets the database
// to a clean state. Protected by:
//  1. AdminOnly middleware (is_admin = TRUE)
//  2. APP_ENV environment variable must be "development"
//  3. Confirmation token must be sent in the request body
func (h *Handlers) AdminDevReset(c *fiber.Ctx) error {
	// Guard: only available in development environment
	if os.Getenv("APP_ENV") != "development" {
		return c.Status(403).JSON(fiber.Map{
			"error": "Resetarea bazei de date este disponibilă doar în mediul de development.",
		})
	}

	type resetReq struct {
		ConfirmText string `json:"confirm_text"` // must be "RESET_ALL_DATA"
	}
	var req resetReq
	if err := c.BodyParser(&req); err != nil || req.ConfirmText != "RESET_ALL_DATA" {
		return c.Status(400).JSON(fiber.Map{
			"error": `Confirmare invalidă. Trimite {"confirm_text": "RESET_ALL_DATA"}.`,
		})
	}

	adminID := middleware.GetUserID(c)

	var deletedUsers, deletedGoals, deletedTasks int
	err := h.db.QueryRow(c.Context(),
		`SELECT deleted_users, deleted_goals, deleted_tasks
		 FROM fn_dev_reset_data($1)`,
		adminID,
	).Scan(&deletedUsers, &deletedGoals, &deletedTasks)
	if err != nil {
		return serverError(c, err)
	}

	db.WriteAudit(c.Context(), h.db, &adminID, "ADMIN_DEV_RESET",
		crypto.SHA256Hex(c.IP()), crypto.SHA256Hex(c.Get("User-Agent")))

	return c.JSON(fiber.Map{
		"message":       "Baza de date a fost resetată. Datele admin sunt păstrate.",
		"deleted_users": deletedUsers,
		"deleted_goals": deletedGoals,
		"deleted_tasks": deletedTasks,
	})
}

// ── GET /api/v1/admin/health ──────────────────────────────────────
// Returns detailed system health including scheduler job status,
// DB connection pool stats, and Redis status.
func (h *Handlers) AdminGetHealth(c *fiber.Ctx) error {
	poolStats := h.db.Stat()

	var dbVersion string
	h.db.QueryRow(c.Context(), `SELECT version()`).Scan(&dbVersion)

	var tableCount int
	h.db.QueryRow(c.Context(),
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'`,
	).Scan(&tableCount)

	// Recent scheduler job activity (last 24h audit entries with SCHEDULER_ prefix)
	var schedulerJobCount int
	h.db.QueryRow(c.Context(), `
		SELECT COUNT(*) FROM audit_log
		WHERE action LIKE 'SCHEDULER_%'
		  AND created_at >= NOW() - INTERVAL '24 hours'
	`).Scan(&schedulerJobCount)

	return c.JSON(fiber.Map{
		"status": "ok",
		"database": fiber.Map{
			"version":              dbVersion,
			"table_count":          tableCount,
			"pool_total_conns":     poolStats.TotalConns(),
			"pool_idle_conns":      poolStats.IdleConns(),
			"pool_acquired_conns":  poolStats.AcquiredConns(),
			"pool_max_conns":       poolStats.MaxConns(),
		},
		"scheduler": fiber.Map{
			"jobs_last_24h": schedulerJobCount,
		},
		"environment": os.Getenv("APP_ENV"),
	})
}
