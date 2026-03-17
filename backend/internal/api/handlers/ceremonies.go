package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/devprimetek/nuviax-app/internal/api/middleware"
	"github.com/devprimetek/nuviax-app/internal/db"
)

// ── GetLatestCeremony — GET /api/v1/ceremonies/:goalId ───────────
func (h *Handlers) GetLatestCeremony(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	goalID, err := uuid.Parse(c.Params("goalId"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID obiectiv invalid."})
	}

	// Verifică dreptul de acces
	if _, err := db.GetGoalByID(c.Context(), h.db, goalID, userID); err != nil {
		return notFound(c)
	}

	var ceremony struct {
		ID           uuid.UUID `json:"id"`
		SprintID     uuid.UUID `json:"sprint_id"`
		CeremonyTier string    `json:"ceremony_tier"`
		CeremonyData []byte    `json:"ceremony_data"`
		GeneratedAt  time.Time `json:"generated_at"`
	}

	err = h.db.QueryRow(c.Context(), `
		SELECT id, sprint_id, ceremony_tier, ceremony_data, generated_at
		FROM latest_ceremonies
		WHERE go_id = $1
	`, goalID).Scan(
		&ceremony.ID, &ceremony.SprintID, &ceremony.CeremonyTier,
		&ceremony.CeremonyData, &ceremony.GeneratedAt,
	)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Nicio ceremonie găsită."})
	}

	return c.JSON(ceremony)
}

// ── GetUnviewedCeremonies — GET /api/v1/ceremonies/unviewed ──────
func (h *Handlers) GetUnviewedCeremonies(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	type unviewedCeremony struct {
		ID           uuid.UUID `json:"id"`
		SprintID     uuid.UUID `json:"sprint_id"`
		GoalID       uuid.UUID `json:"goal_id"`
		CeremonyTier string    `json:"ceremony_tier"`
		CeremonyData []byte    `json:"ceremony_data"`
		GeneratedAt  time.Time `json:"generated_at"`
		GoalName     string    `json:"goal_name"`
	}

	rows, err := h.db.Query(c.Context(), `
		SELECT id, sprint_id, go_id, ceremony_tier, ceremony_data, generated_at, goal_name
		FROM unviewed_ceremonies
		WHERE user_id = $1
		ORDER BY generated_at DESC
	`, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Eroare la citirea ceremoniilor."})
	}
	defer rows.Close()

	var ceremonies []unviewedCeremony
	for rows.Next() {
		var item unviewedCeremony
		if err := rows.Scan(
			&item.ID, &item.SprintID, &item.GoalID,
			&item.CeremonyTier, &item.CeremonyData,
			&item.GeneratedAt, &item.GoalName,
		); err != nil {
			continue
		}
		ceremonies = append(ceremonies, item)
	}
	if err := rows.Err(); err != nil {
		return serverError(c, err)
	}

	if ceremonies == nil {
		ceremonies = []unviewedCeremony{}
	}
	return c.JSON(fiber.Map{"ceremonies": ceremonies})
}

// ── MarkCeremonyViewed — POST /api/v1/ceremonies/:id/view ────────
func (h *Handlers) MarkCeremonyViewed(c *fiber.Ctx) error {
	ceremonyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID ceremonie invalid."})
	}

	if _, err := h.db.Exec(c.Context(), `
		SELECT mark_ceremony_viewed($1)
	`, ceremonyID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Eroare la marcarea ceremoniei."})
	}

	return c.JSON(fiber.Map{"message": "Ceremonie marcată ca vizualizată."})
}
