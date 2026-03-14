#!/bin/bash
# ============================================================
# NUViaX — Deployment Verification Script
# Rulează acest script pe server după deployment
# ============================================================

set -e

echo "═══════════════════════════════════════════════════════"
echo "  NUViaX — Deployment Verification"
echo "═══════════════════════════════════════════════════════"
echo ""

FAIL=0

# ── Verificare containere ─────────────────────────────────
echo "▸ Verificare containere Docker..."
EXPECTED_CONTAINERS=("nuviax_db" "nuviax_redis" "nuviax_api" "nuviax_app" "nuviax_landing")

for container in "${EXPECTED_CONTAINERS[@]}"; do
    if docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
        echo "  ✓ ${container} rulează"
    else
        echo "  ✗ ${container} NU rulează!"
        FAIL=1
    fi
done
echo ""

# ── Verificare health endpoints ───────────────────────────
echo "▸ Verificare health endpoints..."

# API Health
echo -n "  API (http://localhost:8080/health): "
if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
    RESPONSE=$(curl -s http://localhost:8080/health)
    echo "✓ OK - $RESPONSE"
else
    echo "✗ FAIL"
    FAIL=1
fi

# App
echo -n "  App (http://localhost:3000): "
if curl -sf http://localhost:3000 > /dev/null 2>&1; then
    echo "✓ OK"
else
    echo "✗ FAIL"
    FAIL=1
fi

# Landing
echo -n "  Landing (http://localhost:3001): "
if curl -sf http://localhost:3001 > /dev/null 2>&1; then
    echo "✓ OK"
else
    echo "✗ FAIL"
    FAIL=1
fi
echo ""

# ── Verificare database ────────────────────────────────────
echo "▸ Verificare PostgreSQL..."
if docker exec nuviax_db pg_isready -U nuviax -d nuviax > /dev/null 2>&1; then
    echo "  ✓ PostgreSQL ready"
else
    echo "  ✗ PostgreSQL NOT ready"
    FAIL=1
fi
echo ""

# ── Verificare Redis ───────────────────────────────────────
echo "▸ Verificare Redis..."
REDIS_PASS=$(grep REDIS_PASSWORD /var/www/wxr-nuviax/infra/.env | cut -d= -f2)
if docker exec nuviax_redis redis-cli --no-auth-warning -a "$REDIS_PASS" ping > /dev/null 2>&1; then
    echo "  ✓ Redis ready"
else
    echo "  ✗ Redis NOT ready"
    FAIL=1
fi
echo ""

# ── Verificare variabile .env ──────────────────────────────
echo "▸ Verificare .env..."
ENV_FILE="/var/www/wxr-nuviax/infra/.env"

REQUIRED_VARS=("POSTGRES_PASSWORD" "REDIS_PASSWORD" "JWT_PRIVATE_KEY" "JWT_PUBLIC_KEY" "ENCRYPTION_KEY")

for var in "${REQUIRED_VARS[@]}"; do
    if grep -q "^${var}=" "$ENV_FILE" && ! grep -q "^${var}=CHANGE_ME" "$ENV_FILE"; then
        echo "  ✓ ${var} setat"
    else
        echo "  ✗ ${var} NU este setat corect!"
        FAIL=1
    fi
done
echo ""

# ── Verificare imagini Docker ──────────────────────────────
echo "▸ Verificare imagini Docker..."
EXPECTED_IMAGES=("devprimetek/nuviax-api:latest" "devprimetek/nuviax-app:latest" "devprimetek/nuviax-landing:latest")

for image in "${EXPECTED_IMAGES[@]}"; do
    if docker images --format '{{.Repository}}:{{.Tag}}' | grep -q "^${image}$"; then
        echo "  ✓ ${image}"
    else
        echo "  ✗ ${image} lipsește!"
        FAIL=1
    fi
done
echo ""

# ── Verificare rețele ──────────────────────────────────────
echo "▸ Verificare rețele Docker..."
EXPECTED_NETWORKS=("nginx_proxy" "nuviax_net")

for network in "${EXPECTED_NETWORKS[@]}"; do
    if docker network ls --format '{{.Name}}' | grep -q "^${network}$"; then
        echo "  ✓ ${network}"
    else
        echo "  ✗ ${network} lipsește!"
        FAIL=1
    fi
done
echo ""

# ── Verificare DNS (opțional) ─────────────────────────────
echo "▸ Verificare DNS (opțional)..."
DOMAINS=("nuviax.app" "api.nuviax.app" "nuviaxapp.com")

for domain in "${DOMAINS[@]}"; do
    if host "$domain" > /dev/null 2>&1; then
        IP=$(host "$domain" | grep "has address" | awk '{print $4}' | head -n1)
        if [ "$IP" == "83.143.69.103" ]; then
            echo "  ✓ ${domain} → ${IP}"
        else
            echo "  ⚠ ${domain} → ${IP} (așteptat: 83.143.69.103)"
        fi
    else
        echo "  ⚠ ${domain} - DNS nu s-a propagat încă"
    fi
done
echo ""

# ── Rezumat ────────────────────────────────────────────────
echo "═══════════════════════════════════════════════════════"
if [ $FAIL -eq 0 ]; then
    echo "  ✅ TOATE VERIFICĂRILE AU TRECUT!"
    echo "═══════════════════════════════════════════════════════"
    echo ""
    echo "  Aplicația este LIVE:"
    echo "  • API:     https://api.nuviax.app/health"
    echo "  • App:     https://nuviax.app"
    echo "  • Landing: https://nuviaxapp.com"
    echo ""
    echo "  Monitorizare:"
    echo "  • Logs API:     docker logs -f nuviax_api"
    echo "  • Logs App:     docker logs -f nuviax_app"
    echo "  • Logs Landing: docker logs -f nuviax_landing"
    echo ""
    exit 0
else
    echo "  ❌ UNELE VERIFICĂRI AU EȘUAT!"
    echo "═══════════════════════════════════════════════════════"
    echo ""
    echo "  Verifică:"
    echo "  • docker ps"
    echo "  • docker logs nuviax_api"
    echo "  • cat /var/www/wxr-nuviax/infra/.env"
    echo "  • docker compose -f infra/docker-compose.yml -f infra/docker-compose.frontend.yml ps"
    echo ""
    exit 1
fi
