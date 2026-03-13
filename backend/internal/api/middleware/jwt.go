package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/devprimetek/nuviax-app/internal/auth"
	"github.com/devprimetek/nuviax-app/internal/cache"
)

const UserIDKey = "userID"
const UserEmailKey = "userEmail"

func JWTAuth(authSvc *auth.Service, rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extrage token din header
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{
				"error": "Autentificare necesară.",
			})
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		// Parse și validare
		claims, err := authSvc.ParseAccessToken(tokenStr)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Token invalid sau expirat.",
			})
		}

		// Verifică blacklist (logout)
		if cache.IsTokenBlacklisted(c.Context(), rdb, claims.ID) {
			return c.Status(401).JSON(fiber.Map{
				"error": "Sesiune încheiată. Te rog autentifică-te din nou.",
			})
		}

		// Injectează în context
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Token corupt."})
		}
		c.Locals(UserIDKey, userID)
		c.Locals(UserEmailKey, claims.Email)

		return c.Next()
	}
}

// GetUserID extrage userID din context Fiber
func GetUserID(c *fiber.Ctx) uuid.UUID {
	id, _ := c.Locals(UserIDKey).(uuid.UUID)
	return id
}

func GetUserEmail(c *fiber.Ctx) string {
	email, _ := c.Locals(UserEmailKey).(string)
	return email
}
