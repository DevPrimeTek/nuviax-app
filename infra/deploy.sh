#!/bin/bash
# ============================================================
# NUViaX — Deploy & Health Check Script
# Rulează acest script după ce faci push pe main
# sau manual pentru debugging
# ============================================================

set -e

COMPOSE_FILES="-f infra/docker-compose.yml -f infra/docker-compose.frontend.yml"
SERVICES="nuviax_api nuviax_app nuviax_landing"

echo "═══════════════════════════════════════"
echo "  NUViaX Deployment"
echo "═══════════════════════════════════════"

# ── Pull latest images ─────────────────────────────────────
echo ""
echo "▸ Pulling latest images..."
docker pull devprimetek/nuviax-api:latest || true
docker pull devprimetek/nuviax-app:latest || true
docker pull devprimetek/nuviax-landing:latest || true

# ── Stop old containers ────────────────────────────────────
echo ""
echo "▸ Stopping old containers..."
docker compose $COMPOSE_FILES stop $SERVICES

# ── Start services ─────────────────────────────────────────
echo ""
echo "▸ Starting services..."
docker compose $COMPOSE_FILES up -d --no-build $SERVICES

# ── Wait for services ──────────────────────────────────────
echo ""
echo "▸ Waiting for services to be ready..."
sleep 10

# ── Health checks ──────────────────────────────────────────
echo ""
echo "▸ Running health checks..."
echo ""

# API Health Check
echo -n "  API (http://localhost:8080/health): "
if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
    echo "✓ OK"
else
    echo "✗ FAIL"
    docker logs nuviax_api --tail 50
fi

# App Health Check
echo -n "  App (http://localhost:3000): "
if curl -sf http://localhost:3000 > /dev/null 2>&1; then
    echo "✓ OK"
else
    echo "✗ FAIL"
    docker logs nuviax_app --tail 50
fi

# Landing Health Check
echo -n "  Landing (http://localhost:3001): "
if curl -sf http://localhost:3001 > /dev/null 2>&1; then
    echo "✓ OK"
else
    echo "✗ FAIL"
    docker logs nuviax_landing --tail 50
fi

# ── Public URLs ────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════"
echo "  Public URLs (după configurare DNS)"
echo "═══════════════════════════════════════"
echo ""
echo "  API:     https://api.nuviax.app/health"
echo "  App:     https://nuviax.app"
echo "  Landing: https://nuviaxapp.com"
echo ""

# ── Cleanup ────────────────────────────────────────────────
echo "▸ Cleanup old images..."
docker image prune -f

echo ""
echo "✓ Deployment complete!"
echo ""
