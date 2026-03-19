package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devprimetek/nuviax-app/internal/db"
)

// AdminOnly verifies that the authenticated user has is_admin = TRUE.
// Must be used after JWTAuth middleware so that the userID is available in locals.
func AdminOnly(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID.String() == "00000000-0000-0000-0000-000000000000" {
			return c.Status(401).JSON(fiber.Map{"error": "Autentificare necesară."})
		}

		user, err := db.GetUserByID(c.Context(), pool, userID)
		if err != nil || !user.IsAdmin {
			// Return generic 404 to avoid leaking that an admin panel exists
			return c.Status(404).JSON(fiber.Map{"error": "Pagina nu există."})
		}

		return c.Next()
	}
}
