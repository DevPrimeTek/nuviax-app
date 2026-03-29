# Deployment — NuviaX

> Infrastructură VPS, Docker, CI/CD, secrets.

---

## Infrastructură

| Parametru | Valoare |
|-----------|---------|
| VPS IP | `83.143.69.103` |
| SSH user | `sbarbu` |
| SSH port | `22` |
| Deploy path | `/var/www/wxr-nuviax/` |
| DNS registrar | name.com |
| Docker proxy | nginx-proxy + acme-companion (jwilder, shared cu alte proiecte) |

> **CRITIC SSH:** Serverul folosește `AuthorizedKeysFile /etc/ssh/authorized_keys/%u` (gestionat Puppet).
> Cheia deploy trebuie în `/etc/ssh/authorized_keys/sbarbu`, **NU** în `~/.ssh/authorized_keys`.

---

## Domenii → Containere

| Domeniu | Container | Port intern |
|---------|-----------|------------|
| `api.nuviax.app` | `nuviax_api` | 8080 |
| `nuviax.app` + `www` | `nuviax_app` | 3000 |
| `nuviaxapp.com` + `www` | `nuviax_landing` | 3001 |

---

## CI/CD Flow

```
push pe main
  → .github/workflows/deploy.yml        (backend)
  → .github/workflows/deploy-frontend.yml (frontend)
      ↓
  Test → Build Docker → Push DockerHub → SSH Deploy → Health Check
```

### Backend pipeline (deploy.yml)

| Job | Durata | Acțiune |
|-----|--------|---------|
| test | ~1 min | `go build ./...` + `go test ./...` |
| build | ~2 min | Docker build + push `devprimetek/nuviax-api:sha-xxxxxxx` |
| deploy | ~30 sec | SSH → git pull → docker pull → docker compose up nuviax_api |

### Frontend pipeline (deploy-frontend.yml)

Detectează schimbări cu `dorny/paths-filter`:
- `frontend/app/**` → build + push `devprimetek/nuviax-app:latest`
- `frontend/landing/**` → build + push `devprimetek/nuviax-landing:latest`
- Deploy: SSH → docker pull → docker compose up

---

## GitHub Secrets necesare

```
SSH_HOST          = 83.143.69.103
SSH_PORT          = 22
SSH_USER          = sbarbu
SSH_KEY           = [RSA private key]
DOCKERHUB_TOKEN   = [DockerHub access token - cont devprimetek]

POSTGRES_PASSWORD = [openssl rand -base64 32]
REDIS_PASSWORD    = [openssl rand -base64 32]

JWT_PRIVATE_KEY   = [RSA 4096-bit, base64 encoded]
JWT_PUBLIC_KEY    = [RSA 4096-bit public, base64 encoded]
ENCRYPTION_KEY    = [openssl rand -hex 32]

ANTHROPIC_API_KEY = sk-ant-...
RESEND_API_KEY    = re_...
```

**Ghid complet:** `infra/GITHUB_SECRETS.md`

---

## Deploy Manual

```bash
cd /var/www/wxr-nuviax
bash infra/deploy.sh
```

Sau direct:
```bash
docker compose -f infra/docker-compose.yml \
               -f infra/docker-compose.frontend.yml \
               up -d --no-build nuviax_api nuviax_app nuviax_landing
```

---

## Health Checks

```bash
# API
curl https://api.nuviax.app/health
# Expected: {"status":"ok","db":true,"redis":true}

# App
curl -sf https://nuviax.app

# Pe server — containere active
docker ps | grep nuviax

# Logs
docker logs nuviax_api --tail 50
docker logs nuviax_app --tail 20
```

---

## Primul Deploy (setup inițial)

```bash
# Pe server ca sbarbu
cd /var/www
git clone git@github.com:DevPrimeTek/nuviax-app.git wxr-nuviax
cd wxr-nuviax
bash infra/setup-server.sh
# → scriptul generează și afișează toate valorile pentru GitHub Secrets
```

---

## Troubleshooting

### SSH: unable to authenticate (publickey)
```bash
# Verifică AuthorizedKeysFile corect
sudo ls /etc/ssh/authorized_keys/sbarbu
# Nu ~/.ssh/authorized_keys — acesta NU e citit de SSH daemon
```

### Docker login failed
```bash
echo "TOKEN_VALUE" | docker login -u devprimetek --password-stdin
```

### Deploy failed după push
```
GitHub Actions → https://github.com/DevPrimeTek/nuviax-app/actions
→ click pe run eșuat → citește log-ul job-ului deploy
```

---

*Actualizat: v10.4.1 — 2026-03-26*
