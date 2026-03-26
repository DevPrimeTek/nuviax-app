#!/usr/bin/env bash
# NuviaX — Setup cont admin
# Utilizare: ./scripts/setup_admin.sh sbarbu@nuviax.app 'ParolaTA' 'Nume'
# IMPORTANT: single quotes pentru parolă

set -euo pipefail

EMAIL_LOWER=$(echo "${1:-}" | tr '[:upper:]' '[:lower:]')
PASSWORD="${2:-}"
FULL_NAME="${3:-sbarbu}"

[[ -z "$EMAIL_LOWER" || -z "$PASSWORD" ]] && {
  echo "Utilizare: $0 <email> '<parola>' [full_name]"
  echo "Exemplu:   $0 sbarbu@nuviax.app 'NuviaX@sbarbu#2026Prm' 'sbarbu'"
  exit 1
}

echo ""
echo "NuviaX — Setup admin: $EMAIL_LOWER"
echo ""

# ── Verificare containere ─────────────────────────────────────────
echo "[0/3] Verificare containere..."
for C in nuviax_app nuviax_db; do
  docker ps --format '{{.Names}}' | grep -q "^${C}$" || {
    echo "✗ Containerul '$C' nu rulează. Pornește aplicația mai întâi."
    exit 1
  }
done
echo "✓ Containere active"

# ── Scrie script Node.js în /tmp pe HOST ──────────────────────────
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

# ── Copiază scriptul în container și rulează-l ────────────────────
echo "[1/3] Înregistrare cont..."
docker cp /tmp/nv_register.js nuviax_app:/tmp/nv_register.js

RESULT=$(docker exec \
  -e NV_EMAIL="$EMAIL_LOWER" \
  -e NV_PASS="$PASSWORD" \
  -e NV_NAME="$FULL_NAME" \
  nuviax_app node /tmp/nv_register.js 2>&1) || {
    echo "✗ Eroare: $RESULT"
    docker exec nuviax_app rm -f /tmp/nv_register.js
    rm -f /tmp/nv_register.js
    exit 1
  }

docker exec nuviax_app rm -f /tmp/nv_register.js
rm -f /tmp/nv_register.js

[[ "$RESULT" == "CREATED" ]] && echo "✓ Cont creat"
[[ "$RESULT" == "EXISTS"  ]] && echo "ℹ Contul există deja — continuăm"

# ── Setare is_admin în DB ─────────────────────────────────────────
echo "[2/3] Setare admin în DB..."
EMAIL_HASH=$(echo -n "$EMAIL_LOWER" | sha256sum | cut -d' ' -f1)

UPDATED=$(docker exec -i nuviax_db \
  psql -U nuviax -d nuviax -t -A \
  -c "UPDATE users SET is_admin=TRUE WHERE email_hash='${EMAIL_HASH}' RETURNING id;" 2>&1)

[[ -z "$UPDATED" || "$UPDATED" == "UPDATE 0" ]] && {
  echo "✗ Utilizatorul nu a fost găsit în DB."
  echo "  Utilizatori existenți:"
  docker exec -i nuviax_db psql -U nuviax -d nuviax \
    -c "SELECT full_name, is_admin, created_at FROM users ORDER BY created_at DESC LIMIT 5;"
  exit 1
}
echo "✓ is_admin=TRUE setat"

# ── Verificare finală ─────────────────────────────────────────────
echo "[3/3] Verificare..."
docker exec -i nuviax_db psql -U nuviax -d nuviax \
  -c "SELECT id, full_name, is_admin, is_active FROM users WHERE email_hash='${EMAIL_HASH}';"

echo ""
echo "✓ Gata!"
echo "  Login : https://nuviax.app/auth/login"
echo "  Email : $EMAIL_LOWER"
echo "  Admin : https://nuviax.app/admin"
echo ""
