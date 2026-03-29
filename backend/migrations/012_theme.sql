-- ═══════════════════════════════════════════════════════════════
-- Migration 012: User Theme Preference
-- Adds theme column to users for server-side persistence.
-- Complements localStorage (nv_theme) as primary source of truth.
-- ═══════════════════════════════════════════════════════════════

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS theme VARCHAR(10) NOT NULL DEFAULT 'dark'
        CHECK (theme IN ('dark', 'light'));

COMMENT ON COLUMN users.theme IS 'UI theme preference: dark (default) or light';
