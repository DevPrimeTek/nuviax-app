-- Migration 008 — Avatar + User Settings
-- Adds avatar_url to users and a user_settings table for future preferences.

-- B-10: Photo avatar
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);

-- Ensure uploads directory note (actual dir created by Docker/entrypoint)
-- Files are stored at /app/uploads/avatars/{user_id}.{ext}
-- Served via GET /api/v1/settings/avatar (authenticated)
