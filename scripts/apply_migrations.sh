#!/usr/bin/env bash
# NuviaX — Safe migration runner for Dockerized DB
# Fixes relative-path issues from apply_all.sql (which uses \i include directives).

set -euo pipefail

CONTAINER="${1:-nuviax_db}"
DB_USER="${DB_USER:-nuviax}"
DB_NAME="${DB_NAME:-nuviax}"
SRC_DIR="backend/migrations"
DST_DIR="/tmp/nuviax_migrations_$$"

echo "[1/4] Verific container DB: $CONTAINER"
docker ps --format '{{.Names}}' | grep -q "^${CONTAINER}$" || {
  echo "✗ Containerul '$CONTAINER' nu rulează."
  exit 1
}

echo "[2/4] Copiere migrații în container..."
docker exec "$CONTAINER" sh -lc "rm -rf '$DST_DIR' && mkdir -p '$DST_DIR'"
docker cp "$SRC_DIR/." "$CONTAINER:$DST_DIR/"

echo "[3/4] Aplicare apply_all.sql în container..."
docker exec -i "$CONTAINER" sh -lc "cd '$DST_DIR' && psql -v ON_ERROR_STOP=1 -U '$DB_USER' -d '$DB_NAME' -f apply_all.sql"

echo "[4/4] Cleanup..."
docker exec "$CONTAINER" sh -lc "rm -rf '$DST_DIR'"

echo "✅ Migrations aplicate cu succes."
