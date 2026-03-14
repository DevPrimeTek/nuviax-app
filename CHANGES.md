# NUViaX — Fișiere Corectate

## ✅ Modificări Implementate

### 1. Nomenclatură Consistentă
- **Schimbat**: `web` → `app` în toate fișierele
- **Imagini Docker**: `nuviax-web` → `nuviax-app`
- **Servicii**: `nuviax_web` → `nuviax_app`
- **Căi**: `apps/web` → `frontend/app`

### 2. Fișiere Corectate

#### `.github/workflows/`
- ✓ `deploy.yml` - Backend deployment workflow
- ✓ `deploy-frontend.yml` - Frontend deployment workflow (app + landing)

#### `backend/`
- ✓ `Dockerfile` - Multi-stage build optimizat
- ✓ `.dockerignore` - Exclude fișiere inutile din build

#### `frontend/app/`
- ✓ `Dockerfile` - Next.js build pentru aplicația principală
  - Path corect: `frontend/app` (nu `apps/web`)

#### `frontend/landing/`
- ✓ `Dockerfile` - Next.js build pentru landing page
  - Path corect: `frontend/landing` (nu `apps/landing`)

#### `infra/`
- ✓ `docker-compose.yml` - Configurație completă (DB, Redis, API, App, Landing)
- ✓ `docker-compose.frontend.yml` - Overlay pentru frontend cu Traefik
- ✓ `.env.example` - Template pentru variabile de mediu
- ✓ `init-db.sql` - Schema PostgreSQL inițială
- ✓ `deploy.sh` - Script de deployment și health check
- ✓ `GITHUB_SECRETS.md` - Documentație pentru secrets

#### Root
- ✓ `README.md` - Documentație completă
- ✓ `.gitignore` - Exclude .env, .keys, etc.

## 📋 Pași pentru Aplicare

### 1. Backup Repository Existent
```bash
cd /path/to/local/nuviax-app
git checkout -b backup-before-corrections
git push origin backup-before-corrections
```

### 2. Aplicare Fișiere Corectate
```bash
# Copiază fișierele din acest folder în repository-ul tău local
cp -r nuviax-corrected-files/.github .
cp -r nuviax-corrected-files/backend/Dockerfile backend/
cp -r nuviax-corrected-files/backend/.dockerignore backend/
cp -r nuviax-corrected-files/frontend/app/Dockerfile frontend/app/
cp -r nuviax-corrected-files/frontend/landing/Dockerfile frontend/landing/
cp -r nuviax-corrected-files/infra/* infra/
cp nuviax-corrected-files/README.md .
cp nuviax-corrected-files/.gitignore .
```

### 3. Verificare Modificări
```bash
git status
git diff
```

### 4. Commit & Push
```bash
git add .
git commit -m "fix: correct nomenclature and structure (web → app)"
git push origin main
```

### 5. Configurare GitHub Secrets
Urmează instrucțiunile din `infra/GITHUB_SECRETS.md`

### 6. Deploy
GitHub Actions va detecta push-ul și va rula automat deployment.

## ⚠️ Important

1. **NU ȘTERGE** fișierele existente fără backup
2. **VERIFICĂ** că toate path-urile sunt corecte
3. **CONFIGUREAZĂ** GitHub Secrets înainte de push
4. **TESTEAZĂ** local cu `docker compose` înainte de deploy

## 🔍 Verificare Post-Deploy

```bash
# Health checks
curl https://api.nuviax.app/health
curl https://nuviax.app
curl https://nuviaxapp.com

# Docker logs pe server
docker logs nuviax_api
docker logs nuviax_app
docker logs nuviax_landing
```

## 📞 Support

În caz de probleme, verifică:
- GitHub Actions logs: https://github.com/DevPrimeTek/nuviax-app/actions
- Docker logs pe server
- `infra/.env` pentru configurație
