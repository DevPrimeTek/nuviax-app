package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/devprimetek/nuviax-app/internal/ai"
	"github.com/devprimetek/nuviax-app/internal/api/middleware"
	"github.com/devprimetek/nuviax-app/internal/auth"
	"github.com/devprimetek/nuviax-app/internal/cache"
	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/internal/email"
	"github.com/devprimetek/nuviax-app/internal/engine"
	"github.com/devprimetek/nuviax-app/internal/models"
	"github.com/devprimetek/nuviax-app/pkg/crypto"
	"github.com/devprimetek/nuviax-app/pkg/logger"
)

type Handlers struct {
	db     *pgxpool.Pool
	redis  *redis.Client
	auth   *auth.Service
	engine *engine.Engine
	encKey []byte
	email  *email.Client // nil if RESEND_API_KEY not configured
	ai     *ai.Client    // nil if ANTHROPIC_API_KEY not configured
}

func New(pool *pgxpool.Pool, rdb *redis.Client, authSvc *auth.Service, eng *engine.Engine, encKey []byte, emailClient *email.Client, aiClient *ai.Client) *Handlers {
	return &Handlers{db: pool, redis: rdb, auth: authSvc, engine: eng, encKey: encKey, email: emailClient, ai: aiClient}
}

// ═══════════════════════════════════════════════════════════════
// AUTH HANDLERS
// ═══════════════════════════════════════════════════════════════

type registerReq struct {
	Email    string  `json:"email" validate:"required,email,max=255"`
	Password string  `json:"password" validate:"required,min=8,max=128"`
	FullName *string `json:"full_name" validate:"omitempty,max=100"`
	Locale   string  `json:"locale" validate:"omitempty,oneof=ro en ru"`
}

func (h *Handlers) Register(c *fiber.Ctx) error {
	var req registerReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Locale == "" {
		req.Locale = "ro"
	}

	// Verifică dacă există deja
	emailHash := crypto.SHA256Hex(req.Email)
	if _, err := db.GetUserByEmailHash(c.Context(), h.db, emailHash); err == nil {
		return c.Status(409).JSON(fiber.Map{"error": "Adresa de email este deja folosită."})
	}

	// Hash parolă
	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return serverError(c, err)
	}
	salt, _ := crypto.RandomHex(16)

	// Encrypt email
	encEmail, err := crypto.Encrypt(req.Email, h.encKey)
	if err != nil {
		return serverError(c, err)
	}

	user, err := db.CreateUser(c.Context(), h.db, encEmail, emailHash, hash, salt, req.Locale, req.FullName)
	if err != nil {
		return serverError(c, err)
	}

	tokens, err := h.createTokenPair(c, user, req.Email)
	if err != nil {
		return serverError(c, err)
	}

	db.WriteAudit(c.Context(), h.db, &user.ID, "REGISTER", crypto.SHA256Hex(c.IP()), crypto.SHA256Hex(c.Get("User-Agent")))

	// Send welcome email — fire-and-forget, non-blocking
	if h.email != nil {
		name := ""
		if req.FullName != nil {
			name = *req.FullName
		}
		go h.email.SendWelcome(context.Background(), req.Email, name) //nolint:errcheck
	}

	return c.Status(201).JSON(tokens)
}

type loginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *Handlers) Login(c *fiber.Ctx) error {
	var req loginReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := db.GetUserByEmailHash(c.Context(), h.db, crypto.SHA256Hex(req.Email))
	if err != nil || !crypto.CheckPassword(req.Password, user.PasswordHash) {
		// Timing-safe: același răspuns indiferent de motiv
		return c.Status(401).JSON(fiber.Map{"error": "Email sau parolă incorectă."})
	}

	// MFA dacă e activat
	if user.MFAEnabled {
		pending, _ := crypto.RandomHex(16)
		cache.SetMFAPending(c.Context(), h.redis, pending, user.ID.String())
		return c.Status(200).JSON(fiber.Map{
			"mfa_required": true,
			"mfa_token":    pending,
		})
	}

	tokens, err := h.createTokenPair(c, user, req.Email)
	if err != nil {
		return serverError(c, err)
	}

	db.WriteAudit(c.Context(), h.db, &user.ID, "LOGIN", crypto.SHA256Hex(c.IP()), crypto.SHA256Hex(c.Get("User-Agent")))
	return c.JSON(tokens)
}

type mfaVerifyReq struct {
	MFAToken string `json:"mfa_token" validate:"required"`
	Code     string `json:"code" validate:"required,len=6"`
}

func (h *Handlers) MFAVerify(c *fiber.Ctx) error {
	var req mfaVerifyReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	userIDStr, err := cache.GetMFAPending(c.Context(), h.redis, req.MFAToken)
	if err != nil || userIDStr == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Token MFA expirat sau invalid."})
	}

	userID, _ := uuid.Parse(userIDStr)
	user, err := db.GetUserByID(c.Context(), h.db, userID)
	if err != nil || user.MFASecret == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Eroare autentificare."})
	}

	// Decrypt MFA secret
	secret, err := crypto.Decrypt(*user.MFASecret, h.encKey)
	if err != nil || !auth.ValidateTOTP(secret, req.Code) {
		return c.Status(401).JSON(fiber.Map{"error": "Cod incorect."})
	}

	cache.DelMFAPending(c.Context(), h.redis, req.MFAToken)

	email, _ := crypto.Decrypt(user.EmailEncrypted, h.encKey)
	tokens, err := h.createTokenPair(c, user, email)
	if err != nil {
		return serverError(c, err)
	}

	db.WriteAudit(c.Context(), h.db, &user.ID, "LOGIN_MFA", crypto.SHA256Hex(c.IP()), "")
	return c.JSON(tokens)
}

func (h *Handlers) MFAEnable(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	user, err := db.GetUserByID(c.Context(), h.db, userID)
	if err != nil {
		return serverError(c, err)
	}

	email, _ := crypto.Decrypt(user.EmailEncrypted, h.encKey)
	key, err := auth.GenerateTOTPSecret(email)
	if err != nil {
		return serverError(c, err)
	}

	// Encrypt și salvează secretul
	encSecret, err := crypto.Encrypt(key.Secret(), h.encKey)
	if err != nil {
		return serverError(c, err)
	}
	if err := db.UpdateUserMFA(c.Context(), h.db, userID, encSecret, true); err != nil {
		return serverError(c, err)
	}

	return c.JSON(fiber.Map{
		"secret":  key.Secret(),
		"qr_url":  key.URL(),
		"message": "Scanează codul QR în aplicația de autentificare.",
	})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *Handlers) RefreshToken(c *fiber.Ctx) error {
	var req refreshReq
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}

	tokenHash := crypto.SHA256Hex(req.RefreshToken)
	session, err := db.GetSession(c.Context(), h.db, tokenHash)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Sesiune invalidă sau expirată."})
	}

	user, err := db.GetUserByID(c.Context(), h.db, session.UserID)
	if err != nil {
		return serverError(c, err)
	}

	// Revocă vechiul refresh token (rotație)
	db.RevokeSession(c.Context(), h.db, tokenHash)

	email, _ := crypto.Decrypt(user.EmailEncrypted, h.encKey)
	tokens, err := h.createTokenPair(c, user, email)
	if err != nil {
		return serverError(c, err)
	}

	return c.JSON(tokens)
}

func (h *Handlers) Logout(c *fiber.Ctx) error {
	header := c.Get("Authorization")
	if header != "" && strings.HasPrefix(header, "Bearer ") {
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		jti, err := h.auth.GetJTI(tokenStr)
		if err == nil {
			cache.BlacklistToken(c.Context(), h.redis, jti, 15*time.Minute)
		}
	}

	var req refreshReq
	if err := c.BodyParser(&req); err == nil && req.RefreshToken != "" {
		db.RevokeSession(c.Context(), h.db, crypto.SHA256Hex(req.RefreshToken))
	}

	userID := middleware.GetUserID(c)
	db.WriteAudit(c.Context(), h.db, &userID, "LOGOUT", "", "")
	cache.InvalidateDashboard(c.Context(), h.redis, userID.String())

	return c.JSON(fiber.Map{"message": "Deconectat cu succes."})
}

// ═══════════════════════════════════════════════════════════════
// FORGOT PASSWORD / RESET PASSWORD
// ═══════════════════════════════════════════════════════════════

// POST /auth/forgot-password — initiates password reset flow.
// Always returns 200 to prevent email enumeration.
func (h *Handlers) ForgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.JSON(fiber.Map{"message": "Dacă adresa există, vei primi un email."})
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		return c.JSON(fiber.Map{"message": "Dacă adresa există, vei primi un email."})
	}

	user, err := db.GetUserByEmailHash(c.Context(), h.db, crypto.SHA256Hex(req.Email))
	if err != nil {
		// User not found — return success anyway (timing-safe)
		return c.JSON(fiber.Map{"message": "Dacă adresa există, vei primi un email."})
	}

	// Generate a cryptographically random token (32 bytes → 64 hex chars)
	rawToken, err := crypto.RandomHex(32)
	if err != nil {
		return c.JSON(fiber.Map{"message": "Dacă adresa există, vei primi un email."})
	}
	tokenHash := crypto.SHA256Hex(rawToken)

	if err := db.CreatePasswordResetToken(c.Context(), h.db, user.ID, tokenHash); err != nil {
		return c.JSON(fiber.Map{"message": "Dacă adresa există, vei primi un email."})
	}

	// Send reset email — fire-and-forget
	if h.email != nil {
		resetLink := fmt.Sprintf("https://nuviax.app/auth/reset-password?token=%s", rawToken)
		go h.email.SendPasswordReset(context.Background(), req.Email, resetLink) //nolint:errcheck
	}

	db.WriteAudit(c.Context(), h.db, &user.ID, "PASSWORD_RESET_REQUEST",
		crypto.SHA256Hex(c.IP()), crypto.SHA256Hex(c.Get("User-Agent")))

	return c.JSON(fiber.Map{"message": "Dacă adresa există, vei primi un email."})
}

// POST /auth/reset-password — validates token and sets new password.
func (h *Handlers) ResetPassword(c *fiber.Ctx) error {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "Date invalide.")
	}
	if len(req.NewPassword) < 8 {
		return badRequest(c, "Parola nouă trebuie să aibă cel puțin 8 caractere.")
	}
	if req.Token == "" {
		return badRequest(c, "Token invalid sau expirat.")
	}

	tokenHash := crypto.SHA256Hex(strings.TrimSpace(req.Token))
	userID, err := db.GetPasswordResetToken(c.Context(), h.db, tokenHash)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Token invalid sau expirat."})
	}

	newHash, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		return serverError(c, err)
	}

	if err := db.UpdateUserPassword(c.Context(), h.db, userID, newHash); err != nil {
		return serverError(c, err)
	}
	if err := db.MarkPasswordResetTokenUsed(c.Context(), h.db, tokenHash); err != nil {
		return serverError(c, err)
	}

	db.WriteAudit(c.Context(), h.db, &userID, "PASSWORD_RESET_COMPLETE",
		crypto.SHA256Hex(c.IP()), crypto.SHA256Hex(c.Get("User-Agent")))

	return c.JSON(fiber.Map{"message": "Parola a fost resetată cu succes. Te poți autentifica."})
}

// ═══════════════════════════════════════════════════════════════
// HELPERS
// ═══════════════════════════════════════════════════════════════

func (h *Handlers) createTokenPair(c *fiber.Ctx, user *models.User, email string) (*models.AuthTokens, error) {
	accessToken, err := h.auth.GenerateAccessToken(user.ID, email)
	if err != nil {
		return nil, err
	}
	refreshToken, err := h.auth.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Salvează refresh token
	tokenHash := crypto.SHA256Hex(refreshToken)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	fp := fingerprint(c)
	ipSubnet := subnet(c.IP())
	uaHash := crypto.SHA256Hex(c.Get("User-Agent"))
	db.CreateSession(c.Context(), h.db, user.ID, tokenHash, &fp, &ipSubnet, &uaHash, expiresAt)

	return &models.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 min
	}, nil
}

func fingerprint(c *fiber.Ctx) string {
	raw := c.Get("User-Agent") + c.IP()
	return crypto.SHA256Hex(raw)[:16]
}

func subnet(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return strings.Join(parts[:3], ".") + ".0"
	}
	return ip
}

func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": msg})
}

func notFound(c *fiber.Ctx) error {
	return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Resursa nu a fost găsită."})
}

func serverError(c *fiber.Ctx, err error) error {
	logger.Error("handler error", zap.Error(err), zap.String("path", c.Path()), zap.String("method", c.Method()))
	return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Eroare internă. Încearcă din nou."})
}
