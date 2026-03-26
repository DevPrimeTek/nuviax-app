-- Migration 009 — Password Reset Tokens
-- Supports the forgot-password / reset-password flow.
-- Tokens expire after 1 hour and can only be used once.

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(64) NOT NULL UNIQUE,   -- SHA-256 hex of the raw token
    expires_at  TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '1 hour',
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_prt_user  ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_prt_token ON password_reset_tokens(token_hash);

-- Auto-cleanup: remove tokens older than 24h (housekeeping, run via cron or manually)
-- DELETE FROM password_reset_tokens WHERE created_at < NOW() - INTERVAL '24 hours';
