#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────────────
# NuviaX — Setup cont admin proprietar
#
# Utilizare:
#   ./scripts/setup_admin.sh [email] [password] [full_name]
#
# Exemplu:
#   ./scripts/setup_admin.sh sbarbu@nuviax.app "NuviaX@sbarbu!2026#Prm" "sbarbu"
#
# Ce face scriptul:
#   1. Înregistrează contul via API (dacă nu există deja)
#   2. Setează is_admin=TRUE direct în baza de date
#   3. Verifică și afișează statusul final
#
# Cerințe:
#   - curl (disponibil pe orice sistem Linux)
#   - docker (pentru acces la containerul DB)
#   - API-ul NuviaX trebuie să fie pornit
# ─────────────────────────────────────────────────────────────────────────────

set -euo pipefail

# ── Configurare ───────────────────────────────────────────────────────────────

API_URL="${API_URL:-https://api.nuviax.app}"
DB_CONTAINER="${DB_CONTAINER:-nuviax_db}"
DB_USER="${DB_USER:-nuviax}"
DB_NAME="${DB_NAME:-nuviax}"

EMAIL="${1:-sbarbu@nuviax.app}"
PASSWORD="${2:-}"
FULL_NAME="${3:-sbarbu}"
LOCALE="${4:-ro}"

# ── Validare input ────────────────────────────────────────────────────────────

if [[ -z "$PASSWORD" ]]; then
  echo ""
  echo "  ┌─────────────────────────────────────────────────────┐"
  echo "  │  NuviaX — Setup cont admin                          │"
  echo "  └─────────────────────────────────────────────────────┘"
  echo ""
  echo "  Utilizare: $0 <email> <parola> [full_name] [locale]"
  echo ""
  echo "  Exemplu:"
  echo "    $0 sbarbu@nuviax.app \"NuviaX@sbarbu!2026#Prm\" \"sbarbu\" ro"
  echo ""
  exit 1
fi

EMAIL_LOWER=$(echo "$EMAIL" | tr '[:upper:]' '[:lower:]')

echo ""
echo "  ┌─────────────────────────────────────────────────────┐"
echo "  │  NuviaX — Setup cont admin                          │"
echo "  └─────────────────────────────────────────────────────┘"
echo ""
echo "  Email    : $EMAIL_LOWER"
echo "  Nume     : $FULL_NAME"
echo "  Locale   : $LOCALE"
echo "  API URL  : $API_URL"
echo "  DB       : $DB_CONTAINER"
echo ""

# ── Pasul 1: Înregistrare via API ─────────────────────────────────────────────

echo "  [1/3] Înregistrare cont via API..."

REGISTER_PAYLOAD=$(cat <<EOF
{
  "email": "$EMAIL_LOWER",
  "password": "$PASSWORD",
  "full_name": "$FULL_NAME",
  "locale": "$LOCALE"
}
EOF
)

HTTP_STATUS=$(curl -s -o /tmp/nuviax_register_response.json -w "%{http_code}" \
  -X POST "$API_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "$REGISTER_PAYLOAD" \
  --connect-timeout 10 \
  --max-time 30 \
  2>/dev/null || echo "000")

if [[ "$HTTP_STATUS" == "000" ]]; then
  echo ""
  echo "  ✗ Nu pot conecta la API: $API_URL"
  echo "    Verifică că serviciul nuviax_api este pornit."
  echo "    docker ps | grep nuviax_api"
  echo ""
  exit 1
elif [[ "$HTTP_STATUS" == "201" || "$HTTP_STATUS" == "200" ]]; then
  echo "  ✓ Cont înregistrat cu succes (HTTP $HTTP_STATUS)"
elif [[ "$HTTP_STATUS" == "409" ]]; then
  echo "  ℹ Contul există deja (HTTP 409) — continuăm cu setarea admin"
elif [[ "$HTTP_STATUS" == "400" ]]; then
  DETAIL=$(cat /tmp/nuviax_register_response.json 2>/dev/null | grep -o '"error":"[^"]*"' | head -1 || echo "date invalide")
  echo ""
  echo "  ✗ Eroare la înregistrare (HTTP 400): $DETAIL"
  echo "    Verifică că parola are minim 8 caractere."
  echo ""
  exit 1
else
  DETAIL=$(cat /tmp/nuviax_register_response.json 2>/dev/null || echo "")
  echo ""
  echo "  ✗ Eroare neașteptată (HTTP $HTTP_STATUS): $DETAIL"
  echo ""
  exit 1
fi

# ── Pasul 2: Setare is_admin=TRUE în DB ──────────────────────────────────────

echo "  [2/3] Setare drepturi admin în baza de date..."

# Calculează SHA-256 al email-ului (identic cu crypto.SHA256Hex din Go)
EMAIL_HASH=$(echo -n "$EMAIL_LOWER" | sha256sum | cut -d' ' -f1)

SQL_RESULT=$(docker exec -i "$DB_CONTAINER" \
  psql -U "$DB_USER" -d "$DB_NAME" -t -A \
  -c "UPDATE users SET is_admin = TRUE WHERE email_hash = '$EMAIL_HASH' RETURNING id, full_name, is_admin;" \
  2>&1)

if echo "$SQL_RESULT" | grep -q "UPDATE 0"; then
  echo ""
  echo "  ✗ Utilizatorul nu a fost găsit în baza de date."
  echo "    email_hash = $EMAIL_HASH"
  echo ""
  echo "  Debug — utilizatori existenți:"
  docker exec -i "$DB_CONTAINER" \
    psql -U "$DB_USER" -d "$DB_NAME" -t \
    -c "SELECT id, full_name, is_admin, created_at FROM users ORDER BY created_at DESC LIMIT 5;"
  echo ""
  exit 1
elif echo "$SQL_RESULT" | grep -q "|"; then
  echo "  ✓ is_admin=TRUE setat cu succes"
elif echo "$SQL_RESULT" | grep -q "error\|ERROR"; then
  echo ""
  echo "  ✗ Eroare SQL: $SQL_RESULT"
  exit 1
fi

# ── Pasul 3: Verificare finală ─────────────────────────────────────────────────

echo "  [3/3] Verificare status cont..."

VERIFY=$(docker exec -i "$DB_CONTAINER" \
  psql -U "$DB_USER" -d "$DB_NAME" -t -A \
  -c "SELECT id, full_name, is_admin, is_active, created_at FROM users WHERE email_hash = '$EMAIL_HASH';" \
  2>&1)

echo ""
echo "  ┌─────────────────────────────────────────────────────┐"
echo "  │  ✓ Cont admin configurat cu succes!                 │"
echo "  └─────────────────────────────────────────────────────┘"
echo ""
echo "  Date cont:"
echo "  $VERIFY" | tr '|' '\t'
echo ""
echo "  ─────────────────────────────────────────────────────"
echo "  Cum te conectezi:"
echo ""
echo "  1. Deschide: https://nuviax.app/auth/login"
echo "  2. Email   : $EMAIL_LOWER"
echo "  3. Accesează panoul admin: https://nuviax.app/admin"
echo "  ─────────────────────────────────────────────────────"
echo ""

# Curăță fișierele temporare
rm -f /tmp/nuviax_register_response.json
