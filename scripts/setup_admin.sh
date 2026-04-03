#!/usr/bin/env bash
# NuviaX — Setup cont admin (robust)
# Acceptă email SAU username simplu (fără @).
# Dacă primește username, îl convertește automat în <username>@nuviax.app
#
# Exemple:
#   ./scripts/setup_admin.sh sbarbu_admin 'NuviaXAdmin#2026'
#   ./scripts/setup_admin.sh sbarbu_admin@nuviax.app 'NuviaXAdmin#2026' 'Sbarbu Admin'

set -euo pipefail

RAW_ID="${1:-}"
PASSWORD="${2:-}"
FULL_NAME="${3:-}"

if [[ -z "$RAW_ID" || -z "$PASSWORD" ]]; then
  echo "Utilizare: $0 <username_sau_email> '<parola>' [full_name]"
  echo "Exemplu : $0 sbarbu_admin 'NuviaXAdmin#2026' 'Sbarbu Admin'"
  exit 1
fi

# Normalizează identitatea de login
RAW_ID_LOWER=$(echo "$RAW_ID" | tr '[:upper:]' '[:lower:]')
if [[ "$RAW_ID_LOWER" == *"@"* ]]; then
  EMAIL="$RAW_ID_LOWER"
  USERNAME="${RAW_ID_LOWER%%@*}"
else
  USERNAME="$RAW_ID_LOWER"
  EMAIL="${RAW_ID_LOWER}@nuviax.app"
fi

if [[ -z "$FULL_NAME" ]]; then
  FULL_NAME="$USERNAME"
fi

echo ""
echo "NuviaX — Admin access bootstrap"
echo "Username: $USERNAME"
echo "Email   : $EMAIL"
echo ""

# ── Verificare containere ─────────────────────────────────────────
echo "[0/5] Verificare containere..."
for C in nuviax_app nuviax_api nuviax_db; do
  docker ps --format '{{.Names}}' | grep -q "^${C}$" || {
    echo "✗ Containerul '$C' nu rulează."
    exit 1
  }
done
echo "✓ Containere active"

# ── Register via API (idempotent) ─────────────────────────────────
cat > /tmp/nv_register.js << 'NODEOF'
const http = require('http');
const data = JSON.stringify({
  email: process.env.NV_EMAIL,
  password: process.env.NV_PASS,
  full_name: process.env.NV_NAME,
  locale: 'ro'
});
const req = http.request({
  host: 'nuviax_api', port: 8080,
  path: '/api/v1/auth/register', method: 'POST',
  headers: { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(data) }
}, res => {
  let body = '';
  res.on('data', d => body += d);
  res.on('end', () => {
    if (res.statusCode === 201) { console.log('CREATED'); process.exit(0); }
    if (res.statusCode === 409) { console.log('EXISTS'); process.exit(0); }
    console.error('ERR ' + res.statusCode + ': ' + body);
    process.exit(1);
  });
});
req.on('error', e => { console.error('CONN_ERR: ' + e.message); process.exit(1); });
req.write(data);
req.end();
NODEOF

echo "[1/5] Înregistrare cont (sau detectare existent)..."
docker cp /tmp/nv_register.js nuviax_app:/tmp/nv_register.js

REGISTER_RESULT=$(docker exec \
  -e NV_EMAIL="$EMAIL" \
  -e NV_PASS="$PASSWORD" \
  -e NV_NAME="$FULL_NAME" \
  nuviax_app node /tmp/nv_register.js 2>&1) || {
    echo "✗ Eroare register: $REGISTER_RESULT"
    docker exec nuviax_app rm -f /tmp/nv_register.js || true
    rm -f /tmp/nv_register.js
    exit 1
  }

docker exec nuviax_app rm -f /tmp/nv_register.js || true
rm -f /tmp/nv_register.js

[[ "$REGISTER_RESULT" == "CREATED" ]] && echo "✓ Cont creat"
[[ "$REGISTER_RESULT" == "EXISTS"  ]] && echo "ℹ Contul există deja — continuăm"

# ── Promote user to admin ─────────────────────────────────────────
echo "[2/5] Setare is_admin=TRUE..."
EMAIL_HASH=$(echo -n "$EMAIL" | sha256sum | cut -d' ' -f1)

UPDATED_ID=$(docker exec -i nuviax_db \
  psql -U nuviax -d nuviax -t -A \
  -c "UPDATE users SET is_admin=TRUE WHERE email_hash='${EMAIL_HASH}' RETURNING id;" 2>&1)

if [[ -z "$UPDATED_ID" || "$UPDATED_ID" == "UPDATE 0" ]]; then
  echo "✗ Utilizatorul nu a fost găsit în DB după register."
  docker exec -i nuviax_db psql -U nuviax -d nuviax \
    -c "SELECT id, full_name, is_admin, created_at FROM users ORDER BY created_at DESC LIMIT 10;"
  exit 1
fi

echo "✓ Utilizator promovat admin"

# ── Verify login works ────────────────────────────────────────────
echo "[3/5] Verificare login API..."
LOGIN_CODE=$(docker exec -i nuviax_app sh -lc "curl -s -o /tmp/nv_login_resp.json -w '%{http_code}' \
  -X POST http://nuviax_api:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}'")

if [[ "$LOGIN_CODE" != "200" ]]; then
  echo "✗ Login a eșuat (HTTP $LOGIN_CODE)."
  echo "Răspuns:"
  docker exec -i nuviax_app cat /tmp/nv_login_resp.json || true
  exit 1
fi
echo "✓ Login API valid"

# ── Verify admin flag in DB ───────────────────────────────────────
echo "[4/5] Verificare flag admin..."
docker exec -i nuviax_db psql -U nuviax -d nuviax \
  -c "SELECT id, full_name, is_admin, is_active FROM users WHERE email_hash='${EMAIL_HASH}';"

# ── Final instructions ────────────────────────────────────────────
echo "[5/5] Instrucțiuni acces panel"
echo ""
echo "✅ Gata. Folosește aceste date la login:"
echo "   Email   : $EMAIL"
echo "   Password: $PASSWORD"
echo ""
echo "Apoi intră la: https://nuviax.app/admin"
echo "Dacă vezi 404/Acces restricționat, fă logout/login și refresh hard (Ctrl+Shift+R)."
echo ""
