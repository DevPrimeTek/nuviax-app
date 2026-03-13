#!/bin/bash
# ============================================================
# NUViaX — Server Setup Script
# Rulează O SINGURĂ DATĂ pe server ca user: sbarbu
# Directorul aplicației: /var/www/wxr-nuviax
# ============================================================

set -e
echo "═══════════════════════════════════════"
echo "  NUViaX Server Setup"
echo "═══════════════════════════════════════"

# ── 1. Creare director (pattern ca wxr-profixer) ────────────
echo "▸ Creare /var/www/wxr-nuviax..."
sudo mkdir -p /var/www/wxr-nuviax
sudo chown sbarbu:sbarbu /var/www/wxr-nuviax
cd /var/www/wxr-nuviax

# ── 2. Clone repo ────────────────────────────────────────────
echo "▸ Clone monorepo (SSH)...
# Dacă primești eroare SSH, rulează mai întâi:
# cat ~/.ssh/github_actions.pub
# și adaugă cheia în GitHub → Settings → Deploy keys"
git clone git@github.com:DevPrimeTek/nuviax-app.git .

# ── 3. Creare .env ────────────────────────────────────────────
echo "▸ Creare .env din template..."
cp infra/.env.example infra/.env

# ── 4. Generare parole PostgreSQL + Redis ────────────────────
echo "▸ Generare parole DB și Redis..."
PG_PASS=$(openssl rand -base64 32)
REDIS_PASS=$(openssl rand -base64 32)
sed -i "s|POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=${PG_PASS}|" infra/.env
sed -i "s|REDIS_PASSWORD=.*|REDIS_PASSWORD=${REDIS_PASS}|" infra/.env

# ── 5. Generare chei JWT RSA 4096 ────────────────────────────
echo "▸ Generare chei JWT RSA 4096..."
mkdir -p /var/www/wxr-nuviax/.keys
openssl genrsa -out /var/www/wxr-nuviax/.keys/jwt_private.pem 4096 2>/dev/null
openssl rsa -in /var/www/wxr-nuviax/.keys/jwt_private.pem \
    -pubout -out /var/www/wxr-nuviax/.keys/jwt_public.pem 2>/dev/null
chmod 600 /var/www/wxr-nuviax/.keys/jwt_private.pem

JWT_PRIV=$(cat /var/www/wxr-nuviax/.keys/jwt_private.pem | base64 -w 0)
JWT_PUB=$(cat /var/www/wxr-nuviax/.keys/jwt_public.pem | base64 -w 0)
sed -i "s|JWT_PRIVATE_KEY=.*|JWT_PRIVATE_KEY=${JWT_PRIV}|" infra/.env
sed -i "s|JWT_PUBLIC_KEY=.*|JWT_PUBLIC_KEY=${JWT_PUB}|" infra/.env

# ── 6. Generare ENCRYPTION_KEY AES-256 ───────────────────────
echo "▸ Generare ENCRYPTION_KEY..."
ENC_KEY=$(openssl rand -hex 32)
sed -i "s|ENCRYPTION_KEY=.*|ENCRYPTION_KEY=${ENC_KEY}|" infra/.env

# ── 7. Verificare nginx_proxy rulează ────────────────────────
echo "▸ Verificare nginx_proxy..."
if docker ps --format '{{.Names}}' | grep -q "^nginx_proxy$"; then
    echo "   ✅ nginx_proxy rulează"
else
    echo "   ❌ nginx_proxy nu rulează! Pornește mai întâi Profixer."
    exit 1
fi

# ── 8. Creare rețea nuviax_net ───────────────────────────────
echo "▸ Creare rețea nuviax_net..."
docker network create nuviax_net 2>/dev/null && echo "   ✅ nuviax_net creată" \
    || echo "   ℹ nuviax_net există deja"

# ── 9. Start DB + Redis ──────────────────────────────────────
echo "▸ Start PostgreSQL + Redis..."
docker compose -f infra/docker-compose.yml up -d nuviax_db nuviax_redis

echo "▸ Aștept PostgreSQL să fie ready..."
until docker exec nuviax_db pg_isready -U nuviax -d nuviax 2>/dev/null; do
    printf "."; sleep 2
done
echo ""
echo "   ✅ PostgreSQL gata"

# ── 10. Gitignore ────────────────────────────────────────────
grep -q "^\.env$" .gitignore 2>/dev/null || echo ".env" >> .gitignore
grep -q "^\.keys" .gitignore 2>/dev/null || echo ".keys/" >> .gitignore

# ── 11. Afișează valorile pentru GitHub Secrets ──────────────
echo ""
echo "═══════════════════════════════════════════════════════"
echo "  Copiază aceste valori în GitHub Secrets"
echo "  repo: devprimetek/nuviax-app"
echo "═══════════════════════════════════════════════════════"
echo ""
echo "SERVER_HOST=83.143.69.103"
echo "SERVER_USER=sbarbu"
echo ""
echo "--- SSH_DEPLOY_KEY (reutilizezi cheia existentă) ---"
echo "cat ~/.ssh/github_actions    ← copiezi tot ce afișează"
echo ""
echo "--- DOCKERHUB_TOKEN ---"
echo "Copiezi din repo Profixer (același token, același cont)"
echo ""
echo "--- POSTGRES_PASSWORD ---"
echo "${PG_PASS}"
echo ""
echo "--- REDIS_PASSWORD ---"
echo "${REDIS_PASS}"
echo ""
echo "--- ENCRYPTION_KEY ---"
echo "${ENC_KEY}"
echo ""
echo "--- JWT_PRIVATE_KEY (base64, lung) ---"
echo "${JWT_PRIV}" | head -c 80
echo "... [valoarea completă e în infra/.env]"
echo ""
echo "--- JWT_PUBLIC_KEY (base64) ---"
echo "${JWT_PUB}" | head -c 80
echo "... [valoarea completă e în infra/.env]"
echo ""
echo "Valorile complete JWT sunt în: /var/www/wxr-nuviax/infra/.env"
echo ""
echo "═══════════════════════════════════════"
echo "  ✅ Setup complet!"
echo ""
echo "  Pași următori:"
echo "  1. Copiază valorile de mai sus în GitHub Secrets"
echo "  2. Configurează DNS (A record → 83.143.69.103):"
echo "     nuviax.app, www.nuviax.app, api.nuviax.app"
echo "     nuviaxapp.com, www.nuviaxapp.com"
echo "  3. Push pe main → deploy automat"
echo "  4. Verifică: https://api.nuviax.app/health"
echo "═══════════════════════════════════════"
