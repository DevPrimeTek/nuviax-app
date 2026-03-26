#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# NuviaX — Setup cont admin proprietar
#
# Utilizare (IMPORTANT: folosește SINGLE QUOTES pentru parolă):
#   ./scripts/setup_admin.sh <email> '<parola>' [full_name]
#
# Exemplu:
#   ./scripts/setup_admin.sh sbarbu@nuviax.app 'NuviaX@sbarbu#2026Prm' 'sbarbu'
#
# ATENȚIE: Folosește ÎNTOTDEAUNA single quotes '' pentru parolă, nu double quotes "".
#          Bash interpretează ! și $ în double quotes ca comenzi speciale.
#
# Ce face scriptul:
#   1. Înregistrează contul via API (prin docker exec — funcționează pe VPS)
#   2. Setează is_admin=TRUE direct în baza de date
#   3. Verifică și afișează statusul final
#
# Cerințe:
#   - docker (containerele nuviax_api și nuviax_db trebuie să fie pornite)
#   - Rulat de pe VPS-ul unde rulează aplicația
# ─────────────────────────────────────────────────────────────────────────────

set -euo pipefail

# ── Configurare ───────────────────────────────────────────────────────────────

API_CONTAINER="${API_CONTAINER:-nuviax_api}"
DB_CONTAINER="${DB_CONTAINER:-nuviax_db}"
DB_USER="${DB_USER:-nuviax}"
DB_NAME="${DB_NAME:-nuviax}"
# API intern în container — nu prin nginx-proxy
API_INTERNAL="http://localhost:8080"

EMAIL="${1:-}"
PASSWORD="${2:-}"
FULL_NAME="${3:-sbarbu}"
LOCALE="${4:-ro}"

# ── Validare input ────────────────────────────────────────────────────────────

if [[ -z "$EMAIL" || -z "$PASSWORD" ]]; then
  echo ""
  echo "  ┌─────────────────────────────────────────────────────┐"
  echo "  │  NuviaX — Setup cont admin                          │"
  echo "  └─────────────────────────────────────────────────────┘"
  echo ""
  echo "  Utilizare:"
  echo "    $0 <email> '<parola>' [full_name]"
  echo ""
  echo "  IMPORTANT: folosește SINGLE QUOTES pentru parolă!"
  echo ""
  echo "  Exemplu:"
  echo "    $0 sbarbu@nuviax.app 'NuviaX@sbarbu#2026Prm' 'sbarbu'"
  echo ""
  exit 1
fi

EMAIL_LOWER=$(echo "$EMAIL" | tr '[:upper:]' '[:lower:]')

echo ""
echo "  ┌─────────────────────────────────────────────────────┐"
echo "  │  NuviaX — Setup cont admin                          │"
echo "  └─────────────────────────────────────────────────────┘"
echo ""
echo "  Email      : $EMAIL_LOWER"
echo "  Nume       : $FULL_NAME"
echo "  Locale     : $LOCALE"
echo "  API intern : $API_INTERNAL (via docker exec $API_CONTAINER)"
echo "  DB         : $DB_CONTAINER"
echo ""

# ── Verificare containere pornite ─────────────────────────────────────────────

echo "  [0/3] Verificare containere Docker..."

if ! docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${API_CONTAINER}$"; then
  echo ""
  echo "  ✗ Containerul '$API_CONTAINER' nu rulează."
  echo ""
  echo "  Pornește aplicația mai întâi:"
  echo "    cd /var/www/wxr-nuviax"
  echo "    docker compose -f infra/docker-compose.yml up -d"
  echo "    docker compose -f infra/docker-compose.frontend.yml up -d"
  echo ""
  echo "  Verifică containerele active:"
  echo "    docker ps --format 'table {{.Names}}\t{{.Status}}'"
  echo ""
  exit 1
fi

if ! docker ps --format '{{.Names}}' 2>/dev/null | grep -q "^${DB_CONTAINER}$"; then
  echo ""
  echo "  ✗ Containerul '$DB_CONTAINER' nu rulează."
  echo "    docker ps | grep nuviax"
  echo ""
  exit 1
fi

echo "  ✓ Containere active"

# ── Pasul 1: Înregistrare via API (din interiorul containerului API) ───────────

echo "  [1/3] Înregistrare cont via API intern..."

# Construiește JSON fără heredoc (evită probleme cu caractere speciale)
JSON_PAYLOAD="{\"email\":\"${EMAIL_LOWER}\",\"password\":\"${PASSWORD}\",\"full_name\":\"${FULL_NAME}\",\"locale\":\"${LOCALE}\"}"

# Rulează curl din interiorul containerului nuviax_api — acces direct la localhost:8080
HTTP_RESULT=$(docker exec "$API_CONTAINER" \
  curl -s -o /tmp/nv_reg.json -w "%{http_code}" \
  -X POST "${API_INTERNAL}/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  --data-raw "${JSON_PAYLOAD}" \
  --connect-timeout 10 \
  --max-time 30 \
  2>/dev/null || echo "ERR")

HTTP_STATUS="${HTTP_RESULT: -3}"

if [[ "$HTTP_RESULT" == "ERR" || "$HTTP_STATUS" == "000" ]]; then
  echo ""
  echo "  ✗ Nu pot conecta la API intern ($API_INTERNAL)."
  echo "    Verifică că procesul din container ascultă pe portul 8080:"
  echo "    docker exec $API_CONTAINER ps aux"
  echo "    docker logs $API_CONTAINER --tail 20"
  echo ""
  exit 1
elif [[ "$HTTP_STATUS" == "201" || "$HTTP_STATUS" == "200" ]]; then
  echo "  ✓ Cont înregistrat cu succes (HTTP $HTTP_STATUS)"
elif [[ "$HTTP_STATUS" == "409" ]]; then
  echo "  ℹ Contul există deja (HTTP 409) — continuăm cu setarea admin"
elif [[ "$HTTP_STATUS" == "400" ]]; then
  DETAIL=$(docker exec "$API_CONTAINER" cat /tmp/nv_reg.json 2>/dev/null | \
    grep -o '"error":"[^"]*"' | head -1 || echo "date invalide")
  echo ""
  echo "  ✗ Eroare la înregistrare (HTTP 400): $DETAIL"
  echo "    Verifică că parola are minim 8 caractere."
  echo ""
  exit 1
else
  DETAIL=$(docker exec "$API_CONTAINER" cat /tmp/nv_reg.json 2>/dev/null || echo "")
  echo ""
  echo "  ✗ Eroare API (HTTP $HTTP_STATUS): $DETAIL"
  echo ""
  exit 1
fi

# ── Pasul 2: Setare is_admin=TRUE în DB ──────────────────────────────────────

echo "  [2/3] Setare drepturi admin în baza de date..."

# SHA-256 al email-ului (identic cu crypto.SHA256Hex din Go)
EMAIL_HASH=$(echo -n "$EMAIL_LOWER" | sha256sum | cut -d' ' -f1)

SQL_RESULT=$(docker exec -i "$DB_CONTAINER" \
  psql -U "$DB_USER" -d "$DB_NAME" -t -A \
  -c "UPDATE users SET is_admin = TRUE WHERE email_hash = '${EMAIL_HASH}' RETURNING id, full_name, is_admin;" \
  2>&1)

if echo "$SQL_RESULT" | grep -q "^UPDATE 0$"; then
  echo ""
  echo "  ✗ Utilizatorul nu a fost găsit în baza de date."
  echo "    email_hash calculat = $EMAIL_HASH"
  echo ""
  echo "  Utilizatori existenți în DB:"
  docker exec -i "$DB_CONTAINER" \
    psql -U "$DB_USER" -d "$DB_NAME" \
    -c "SELECT id, full_name, is_admin, created_at FROM users ORDER BY created_at DESC LIMIT 5;"
  echo ""
  exit 1
elif echo "$SQL_RESULT" | grep -qiE "error|fatal"; then
  echo ""
  echo "  ✗ Eroare SQL: $SQL_RESULT"
  exit 1
else
  echo "  ✓ is_admin=TRUE setat cu succes"
fi

# ── Pasul 3: Verificare finală ────────────────────────────────────────────────

echo "  [3/3] Verificare status cont..."

docker exec -i "$DB_CONTAINER" \
  psql -U "$DB_USER" -d "$DB_NAME" \
  -c "SELECT id, full_name, is_admin, is_active, created_at FROM users WHERE email_hash = '${EMAIL_HASH}';" \
  2>&1

echo ""
echo "  ┌─────────────────────────────────────────────────────┐"
echo "  │  ✓ Cont admin configurat cu succes!                 │"
echo "  └─────────────────────────────────────────────────────┘"
echo ""
echo "  Cum te conectezi:"
echo "  ─────────────────────────────────────────────────────"
echo "  1. Login  : https://nuviax.app/auth/login"
echo "  2. Email  : $EMAIL_LOWER"
echo "  3. Admin  : https://nuviax.app/admin"
echo "  ─────────────────────────────────────────────────────"
echo ""

# Curăță temporarele din container
docker exec "$API_CONTAINER" rm -f /tmp/nv_reg.json 2>/dev/null || true
