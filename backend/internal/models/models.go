package models

import (
	"time"

	"github.com/google/uuid"
)

// ── User ──────────────────────────────────────────────────────────────────────

type User struct {
	ID             uuid.UUID `db:"id"              json:"id"`
	EmailEncrypted string    `db:"email_encrypted" json:"-"`
	EmailHash      string    `db:"email_hash"      json:"-"`
	PasswordHash   string    `db:"password_hash"   json:"-"`
	Salt           string    `db:"salt"            json:"-"`
	FullName       *string   `db:"full_name"       json:"full_name,omitempty"`
	Locale         string    `db:"locale"          json:"locale"`
	Theme          string    `db:"theme"           json:"theme"`
	AvatarURL      *string   `db:"avatar_url"      json:"avatar_url,omitempty"`
	MFASecret      *string   `db:"mfa_secret"      json:"-"`
	MFAEnabled     bool      `db:"mfa_enabled"     json:"mfa_enabled"`
	IsActive       bool      `db:"is_active"       json:"is_active"`
	IsAdmin        bool      `db:"is_admin"        json:"is_admin"`
	CreatedAt      time.Time `db:"created_at"      json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"      json:"updated_at"`
}

// ── Session ───────────────────────────────────────────────────────────────────

type UserSession struct {
	ID            uuid.UUID  `db:"id"`
	UserID        uuid.UUID  `db:"user_id"`
	TokenHash     string     `db:"token_hash"`
	DeviceFP      *string    `db:"device_fp"`
	IPSubnet      *string    `db:"ip_subnet"`
	UserAgentHash *string    `db:"user_agent_hash"`
	ExpiresAt     time.Time  `db:"expires_at"`
	Revoked       bool       `db:"revoked"`
	CreatedAt     time.Time  `db:"created_at"`
}

// ── Auth Tokens ──────────────────────────────────────────────────────────────

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // secunde
}
