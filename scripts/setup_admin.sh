#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# NuviaX — Setup cont admin proprietar
#
# Utilizare (SINGLE QUOTES pentru parolă — bash nu expandează caracterele speciale):
#   ./scripts/setup_admin.sh <email> '<parola>' [full_name]
#
# Exemplu:
#   ./scripts/setup_admin.sh sbarbu@nuviax.app 'NuviaX@sbarbu#2026Prm' 'sbarbu'
#
# Ce face scriptul:
#   1. Găsește IP-ul intern al containerului nuviax_api din rețeaua Docker
#   2. Înregistrează contul via API de pe host (curl → IP intern:8080)
#   3. Setează is_admin=TRUE în DB (docker exec nuviax_db psql)
#   4. Verifică și afișează statusul final
# ─────────────────────────────────────────────────────────────────────────────

set -euo pipefail

API_CONTAINER="${API_CONTAINER:-nuviax_api}"
DB_CONTAINER="${DB_CONTAINER:-nuviax_db}"
DB_USER="${DB_USER:-nuviax}"
DB_NAME="${DB_NAME:-nuviax}"
DOCKER_NETWORK="${DOCKER_NETWORK:-nuviax_net}"

EMAIL="${1:-}"
PASSWORD="${2:-}"
FULL_NAME="${3:-sbarbu}"
LOCALE="${4:-ro}"

# ── Validare input ────────────────────────────────────────────────────────────

if [[ -z "$EMAIL" || -z "$PASSWORD" ]]; then
  echo ""
  echo "  Utilizare: $0 <email> '<parola>' [full_name]"
  echo "  Exemplu:   $0 sbarbu@nuviax.app 'NuviaX@sbarbu#2026Prm' 'sbarbu'"
  echo "  IMPORTANT: folosește SINGLE QUOTES '' pentru parolă!"
  echo ""
  exit 1
fi

EMAIL_LOWER=$(echo "$EMAIL" | tr '[:upper:]' '[:lower:]')

echo ""
echo "  ┌─────────────────────────────────────────────────────┐"
echo "  │  NuviaX — Setup cont admin                          │"
echo "  └─────────────────────────────────────────────────────┘"
echo ""

# ── [0/3] Verificare containere ───────────────────────────────────────────────

echo "  [0/3] Verificare containere..."

for CNAME in "$API_CONTAINER" "$DB_CONTAINER"; do
  if ! docker ps --format '{{.Names}}' | grep -q "^${CNAME}$"; then
    echo ""
    echo "  ✗ Containerul '$CNAME' nu rulează."
    echo "    docker ps | grep nuviax"
    echo ""
    exit 1
  fi
done

echo "  ✓ nuviax_api și nuviax_db active"

# ── Determină IP-ul intern al containerului API ───────────────────────────────

# Încearcă rețeaua specifică, cu fallback la orice rețea disponibilă
API_IP=$(docker inspect "$API_CONTAINER" \
  --format "{{(index .NetworkSettings.Networks \"${DOCKER_NETWORK}\").IPAddress}}" \
  2>/dev/null || true)

if [[ -z "$API_IP" ]]; then
  # Fallback: primul IP din orice rețea
  API_IP=$(docker inspect "$API_CONTAINER" \
    --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' \
    2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' | head -1 || true)
fi

if [[ -z "$API_IP" ]]; then
  echo ""
  echo "  ✗ Nu pot determina IP-ul containerului $API_CONTAINER."
  echo "    docker inspect $API_CONTAINER --format '{{json .NetworkSettings.Networks}}'"
  echo ""
  exit 1
fi

echo "  ✓ API intern: http://${API_IP}:8080"
echo ""
echo "  Email  : $EMAIL_LOWER"
echo "  Nume   : $FULL_NAME"
echo ""

# ── [1/3] Înregistrare via API (curl de pe HOST → IP intern Docker) ───────────

echo "  [1/3] Înregistrare cont via API..."

# Construiește JSON — fără heredoc, fără caractere problematice
JSON="{\"email\":\"${EMAIL_LOWER}\",\"password\":\"${PASSWORD}\",\"full_name\":\"${FULL_NAME}\",\"locale\":\"${LOCALE}\"}"

HTTP_STATUS=$(curl -s \
  -o /tmp/nuviax_setup_response.json \
  -w "%{http_code}" \
  -X POST "http://${API_IP}:8080/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "${JSON}" \
  --connect-timeout 10 \
  --max-time 30 \
  2>/dev/null) || HTTP_STATUS="000"

case "$HTTP_STATUS" in
  201|200)
    echo "  ✓ Cont creat cu succes (HTTP $HTTP_STATUS)"
    ;;
  409)
    echo "  ℹ Contul există deja (HTTP 409) — continuăm cu setarea admin"
    ;;
  400)
    DETAIL=$(grep -o '"error":"[^"]*"' /tmp/nuviax_setup_response.json 2>/dev/null | head -1 || echo "")
    echo ""
    echo "  ✗ Date invalide (HTTP 400): $DETAIL"
    echo "    Parola trebuie să aibă minim 8 caractere."
    rm -f /tmp/nuviax_setup_response.json
    exit 1
    ;;
  000)
    echo ""
    echo "  ✗ Nu pot conecta la API (http://${API_IP}:8080)"
    echo "    Verifică că containerul rulează și ascultă:"
    echo "    docker logs $API_CONTAINER --tail 20"
    rm -f /tmp/nuviax_setup_response.json
    exit 1
    ;;
  *)
    DETAIL=$(cat /tmp/nuviax_setup_response.json 2>/dev/null || echo "")
    echo ""
    echo "  ✗ Eroare API (HTTP $HTTP_STATUS): $DETAIL"
    rm -f /tmp/nuviax_setup_response.json
    exit 1
    ;;
esac

# ── [2/3] Setare is_admin=TRUE în DB ─────────────────────────────────────────

echo "  [2/3] Setare drepturi admin..."

EMAIL_HASH=$(echo -n "$EMAIL_LOWER" | sha256sum | cut -d' ' -f1)

SQL_OUT=$(docker exec -i "$DB_CONTAINER" \
  psql -U "$DB_USER" -d "$DB_NAME" -t -A \
  -c "UPDATE users SET is_admin=TRUE WHERE email_hash='${EMAIL_HASH}' RETURNING id;" \
  2>&1)

if echo "$SQL_OUT" | grep -q "^UPDATE 0$\|^$"; then
  echo ""
  echo "  ✗ Utilizatorul nu a fost găsit. Utilizatori în DB:"
  docker exec -i "$DB_CONTAINER" \
    psql -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT id, full_name, is_admin, created_at FROM users ORDER BY created_at DESC LIMIT 5;"
  exit 1
fi

echo "  ✓ is_admin=TRUE setat"

# ── [3/3] Verificare finală ───────────────────────────────────────────────────

echo "  [3/3] Verificare..."

docker exec -i "$DB_CONTAINER" \
  psql -U "$DB_USER" -d "$DB_NAME" \
  -c "SELECT id, full_name, is_admin, is_active, created_at FROM users WHERE email_hash='${EMAIL_HASH}';"

echo ""
echo "  ┌─────────────────────────────────────────────────────┐"
echo "  │  ✓ Gata! Cont admin configurat.                     │"
echo "  └─────────────────────────────────────────────────────┘"
echo ""
echo "  1. Login : https://nuviax.app/auth/login"
echo "  2. Email : $EMAIL_LOWER"
echo "  3. Admin : https://nuviax.app/admin"
echo ""

rm -f /tmp/nuviax_setup_response.json
