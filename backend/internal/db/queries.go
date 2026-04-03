package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devprimetek/nuviax-app/internal/models"
)

var ErrNotFound = errors.New("not found")

// ═══════════════════════════════════════════════════════════════
// USERS
// ═══════════════════════════════════════════════════════════════

func CreateUser(ctx context.Context, pool *pgxpool.Pool,
	emailEncrypted, emailHash, passwordHash, salt, locale string,
	fullName *string,
) (*models.User, error) {
	u := &models.User{}
	err := pool.QueryRow(ctx, `
		INSERT INTO users
			(email_encrypted, email_hash, password_hash, salt, full_name, locale)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, email_encrypted, email_hash, password_hash, salt,
		          full_name, locale, mfa_enabled, is_active, is_admin, created_at, updated_at
	`, emailEncrypted, emailHash, passwordHash, salt, fullName, locale).Scan(
		&u.ID, &u.EmailEncrypted, &u.EmailHash, &u.PasswordHash, &u.Salt,
		&u.FullName, &u.Locale, &u.MFAEnabled, &u.IsActive, &u.IsAdmin,
		&u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

func GetUserByEmailHash(ctx context.Context, pool *pgxpool.Pool, hash string) (*models.User, error) {
	u := &models.User{}
	err := pool.QueryRow(ctx, `
		SELECT id, email_encrypted, email_hash, password_hash, salt,
		       full_name, locale, mfa_secret, mfa_enabled, is_active, is_admin, created_at, updated_at
		FROM users WHERE email_hash = $1 AND is_active = TRUE
	`, hash).Scan(
		&u.ID, &u.EmailEncrypted, &u.EmailHash, &u.PasswordHash, &u.Salt,
		&u.FullName, &u.Locale, &u.MFASecret, &u.MFAEnabled, &u.IsActive, &u.IsAdmin,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func GetUserByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*models.User, error) {
	u := &models.User{}
	err := pool.QueryRow(ctx, `
		SELECT id, email_encrypted, email_hash, password_hash, salt,
		       full_name, locale, COALESCE(theme, 'dark'), avatar_url, mfa_secret, mfa_enabled, is_active, is_admin, created_at, updated_at
		FROM users WHERE id = $1 AND is_active = TRUE
	`, id).Scan(
		&u.ID, &u.EmailEncrypted, &u.EmailHash, &u.PasswordHash, &u.Salt,
		&u.FullName, &u.Locale, &u.Theme, &u.AvatarURL, &u.MFASecret, &u.MFAEnabled, &u.IsActive, &u.IsAdmin,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func UpdateUserMFA(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, secret string, enabled bool) error {
	_, err := pool.Exec(ctx,
		`UPDATE users SET mfa_secret=$1, mfa_enabled=$2, updated_at=NOW() WHERE id=$3`,
		secret, enabled, userID)
	return err
}

func UpdateUserPassword(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, newPasswordHash string) error {
	_, err := pool.Exec(ctx,
		`UPDATE users SET password_hash=$1, updated_at=NOW() WHERE id=$2`, newPasswordHash, userID)
	return err
}

// ═══════════════════════════════════════════════════════════════
// SESSIONS
// ═══════════════════════════════════════════════════════════════

func CreateSession(ctx context.Context, pool *pgxpool.Pool,
	userID uuid.UUID, tokenHash string,
	deviceFP, ipSubnet, uaHash *string, expiresAt time.Time,
) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO user_sessions
			(user_id, token_hash, device_fp, ip_subnet, user_agent_hash, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, userID, tokenHash, deviceFP, ipSubnet, uaHash, expiresAt)
	return err
}

func GetSession(ctx context.Context, pool *pgxpool.Pool, tokenHash string) (*models.UserSession, error) {
	s := &models.UserSession{}
	err := pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked
		FROM user_sessions
		WHERE token_hash=$1 AND revoked=FALSE AND expires_at > NOW()
	`, tokenHash).Scan(&s.ID, &s.UserID, &s.TokenHash, &s.ExpiresAt, &s.Revoked)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return s, err
}

func RevokeSession(ctx context.Context, pool *pgxpool.Pool, tokenHash string) error {
	_, err := pool.Exec(ctx,
		`UPDATE user_sessions SET revoked=TRUE WHERE token_hash=$1`, tokenHash)
	return err
}

// ═══════════════════════════════════════════════════════════════
// AUDIT LOG
// ═══════════════════════════════════════════════════════════════

func WriteAudit(ctx context.Context, pool *pgxpool.Pool,
	userID *uuid.UUID, action, ipHash, uaHash string,
) {
	// Fire and forget — nu blocăm request-ul
	go func() {
		_, _ = pool.Exec(context.Background(), `
			INSERT INTO audit_log (user_id, action, ip_hash, ua_hash)
			VALUES ($1,$2,$3,$4)
		`, userID, action, ipHash, uaHash)
	}()
}

// ═══════════════════════════════════════════════════════════════
// PASSWORD RESET
// ═══════════════════════════════════════════════════════════════

func CreatePasswordResetToken(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, tokenHash string) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '1 hour')
	`, userID, tokenHash)
	return err
}

func GetPasswordResetToken(ctx context.Context, pool *pgxpool.Pool, tokenHash string) (uuid.UUID, error) {
	var userID uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT user_id FROM password_reset_tokens
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND expires_at > NOW()
	`, tokenHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return userID, err
}

func MarkPasswordResetTokenUsed(ctx context.Context, pool *pgxpool.Pool, tokenHash string) error {
	_, err := pool.Exec(ctx,
		`UPDATE password_reset_tokens SET used_at = NOW() WHERE token_hash = $1`, tokenHash)
	return err
}
