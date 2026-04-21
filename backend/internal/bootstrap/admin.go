// Package bootstrap ensures critical infrastructure users exist on startup.
//
// EnsureAdmin creates (or promotes) a single admin account using the
// ADMIN_BOOTSTRAP_EMAIL and ADMIN_BOOTSTRAP_PASSWORD environment variables.
// This is idempotent:
//   - If the user doesn't exist, it is created with is_admin = TRUE.
//   - If the user exists but isn't admin, is_admin is flipped to TRUE.
//   - If the user exists and is admin, nothing happens.
//
// The password is only used at user creation time; existing accounts keep
// their current password (use the standard password reset flow to change it).
// This avoids silent password rotations when the env var is updated.
package bootstrap

import (
	"context"
	"encoding/hex"
	"errors"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devprimetek/nuviax-app/internal/db"
	"github.com/devprimetek/nuviax-app/pkg/crypto"
	"github.com/devprimetek/nuviax-app/pkg/logger"
	"go.uber.org/zap"
)

func EnsureAdmin(ctx context.Context, pool *pgxpool.Pool, rawEncKey []byte) {
	email := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_BOOTSTRAP_EMAIL")))
	password := os.Getenv("ADMIN_BOOTSTRAP_PASSWORD")

	if email == "" || password == "" {
		return
	}

	encKey, err := parseEncKey(rawEncKey)
	if err != nil {
		logger.Error("admin bootstrap: invalid ENCRYPTION_KEY", zap.Error(err))
		return
	}

	emailHash := crypto.SHA256Hex(email)

	// Caz 1: user-ul există deja — promovăm la admin dacă nu e deja
	existing, err := db.GetUserByEmailHash(ctx, pool, emailHash)
	if err == nil && existing != nil {
		if existing.IsAdmin {
			return
		}
		if _, err := pool.Exec(ctx, `UPDATE users SET is_admin = TRUE, updated_at = NOW() WHERE id = $1`, existing.ID); err != nil {
			logger.Error("admin bootstrap: promote failed", zap.Error(err))
			return
		}
		logger.Info("admin bootstrap: user promoted to admin", zap.String("email_hash", emailHash[:12]))
		return
	}

	// Caz 2: user-ul nu există — îl creăm
	hash, err := crypto.HashPassword(password)
	if err != nil {
		logger.Error("admin bootstrap: hash failed", zap.Error(err))
		return
	}
	salt, _ := crypto.RandomHex(16)
	encEmail, err := crypto.Encrypt(email, encKey)
	if err != nil {
		logger.Error("admin bootstrap: encrypt failed", zap.Error(err))
		return
	}

	fullName := "Administrator"
	user, err := db.CreateUser(ctx, pool, encEmail, emailHash, hash, salt, "ro", &fullName)
	if err != nil {
		logger.Error("admin bootstrap: create failed", zap.Error(err))
		return
	}
	if _, err := pool.Exec(ctx, `UPDATE users SET is_admin = TRUE, updated_at = NOW() WHERE id = $1`, user.ID); err != nil {
		logger.Error("admin bootstrap: promote after create failed", zap.Error(err))
		return
	}
	logger.Info("admin bootstrap: admin user created", zap.String("email_hash", emailHash[:12]))
}

// parseEncKey accepts 32 raw bytes or 64-hex characters, matching api.parseEncKey.
func parseEncKey(raw []byte) ([]byte, error) {
	s := string(raw)
	if len(s) == 64 {
		if key, err := hex.DecodeString(s); err == nil && len(key) == 32 {
			return key, nil
		}
	}
	if len(raw) == 32 {
		return raw, nil
	}
	return nil, errors.New("ENCRYPTION_KEY must be 32 bytes or 64-char hex")
}

