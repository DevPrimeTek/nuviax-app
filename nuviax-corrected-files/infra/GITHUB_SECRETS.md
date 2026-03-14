# GitHub Secrets — NUViaX

**Repository**: `DevPrimeTek/nuviax-app`  
**Path**: Settings → Secrets and variables → Actions → New repository secret

---

## 🔑 Secrete Necesare

### 1. Server SSH (pentru deployment)

| Secret | Valoare | Cum obții |
|--------|---------|-----------|
| `SSH_HOST` | `83.143.69.103` | IP-ul serverului |
| `SSH_PORT` | `22` | Portul SSH |
| `SSH_USER` | `sbarbu` | Username-ul pe server |
| `SSH_KEY` | [Cheie privată SSH] | `cat ~/.ssh/github_actions` pe server |

**Cum generezi SSH_KEY** (dacă nu există):
```bash
# Pe server ca sbarbu
ssh-keygen -t ed25519 -f ~/.ssh/github_actions -C "github-actions-nuviax"

# Adaugă în authorized_keys
cat ~/.ssh/github_actions.pub >> ~/.ssh/authorized_keys

# Copiază cheia privată pentru GitHub Secret
cat ~/.ssh/github_actions
# Copiezi TOTUL (de la -----BEGIN până la -----END-----)
```

---

### 2. DockerHub (pentru push imagini)

| Secret | Valoare | Cum obții |
|--------|---------|-----------|
| `DOCKERHUB_TOKEN` | [Token] | DockerHub → Account Settings → Security → New Access Token |

**Note**: 
- Folosește același cont `devprimetek` ca și pentru Profixer
- Poți reutiliza același token dacă îl ai salvat

---

### 3. Database & Redis (generate automat)

Aceste valori sunt generate de scriptul `infra/setup-server.sh`:

| Secret | Cum generezi manual | Unde găsești |
|--------|---------------------|--------------|
| `POSTGRES_PASSWORD` | `openssl rand -base64 32` | `/var/www/wxr-nuviax/infra/.env` |
| `REDIS_PASSWORD` | `openssl rand -base64 32` | `/var/www/wxr-nuviax/infra/.env` |

---

### 4. JWT & Encryption (generate automat)

Generate automat de `infra/setup-server.sh`:

| Secret | Unde găsești |
|--------|--------------|
| `JWT_PRIVATE_KEY` | `/var/www/wxr-nuviax/infra/.env` (linie foarte lungă, base64) |
| `JWT_PUBLIC_KEY` | `/var/www/wxr-nuviax/infra/.env` (base64) |
| `ENCRYPTION_KEY` | `/var/www/wxr-nuviax/infra/.env` (hex 64 chars) |

**Cum generezi manual** (dacă e necesar):
```bash
# JWT Keys
mkdir -p .keys
openssl genrsa -out .keys/jwt_private.pem 4096
openssl rsa -in .keys/jwt_private.pem -pubout -out .keys/jwt_public.pem

# Base64 encode pentru GitHub
cat .keys/jwt_private.pem | base64 -w 0  # → JWT_PRIVATE_KEY
cat .keys/jwt_public.pem | base64 -w 0   # → JWT_PUBLIC_KEY

# Encryption Key
openssl rand -hex 32  # → ENCRYPTION_KEY
```

---

## 📋 Checklist Setup

### Pasul 1: Setup Server
```bash
# Pe server ca sbarbu
cd /var/www
git clone git@github.com:DevPrimeTek/nuviax-app.git wxr-nuviax
cd wxr-nuviax
bash infra/setup-server.sh
```

Scriptul va afișa toate valorile necesare la final.

### Pasul 2: Configurare GitHub Secrets

1. Deschide: https://github.com/DevPrimeTek/nuviax-app/settings/secrets/actions
2. Click **"New repository secret"** pentru fiecare:

**Server SSH:**
- [ ] `SSH_HOST` = `83.143.69.103`
- [ ] `SSH_PORT` = `22`
- [ ] `SSH_USER` = `sbarbu`
- [ ] `SSH_KEY` = [output din `cat ~/.ssh/github_actions`]

**DockerHub:**
- [ ] `DOCKERHUB_TOKEN` = [token din DockerHub]

**Database & Cache:**
- [ ] `POSTGRES_PASSWORD` = [din output setup-server.sh]
- [ ] `REDIS_PASSWORD` = [din output setup-server.sh]

**Security:**
- [ ] `JWT_PRIVATE_KEY` = [din output setup-server.sh]
- [ ] `JWT_PUBLIC_KEY` = [din output setup-server.sh]
- [ ] `ENCRYPTION_KEY` = [din output setup-server.sh]

### Pasul 3: Configurare DNS

Adaugă A records pentru domeniul tău → `83.143.69.103`:

- [ ] `nuviax.app`
- [ ] `www.nuviax.app`
- [ ] `api.nuviax.app`
- [ ] `nuviaxapp.com`
- [ ] `www.nuviaxapp.com`

### Pasul 4: Test Deploy

```bash
# Local
git add .
git commit -m "Initial deployment setup"
git push origin main

# GitHub Actions va rula automat
# Verifică: https://github.com/DevPrimeTek/nuviax-app/actions
```

### Pasul 5: Verificare

După 5-10 minute:

```bash
# Health checks
curl https://api.nuviax.app/health
# Expected: {"status":"ok","db":true,"redis":true}

curl https://nuviax.app
# Expected: Frontend HTML

curl https://nuviaxapp.com
# Expected: Landing page HTML
```

---

## 🔒 Security Notes

1. **NICIODATĂ** nu comite `.env` în git
2. **NICIODATĂ** nu comiti fișiere din `.keys/` în git
3. Toate secretele sunt în GitHub Secrets (encrypted)
4. Pe server, fișierul `.env` este protejat (chmod 600)
5. Cheile JWT sunt separate de cod

---

## 🆘 Troubleshooting

### Secret values prea lungi
GitHub acceptă până la 65,536 caractere. JWT keys sunt lungi dar acceptate.

### Deploy failed - SSH connection refused
Verifică:
```bash
# Pe server
sudo systemctl status sshd

# Testează SSH key
ssh -i ~/.ssh/github_actions sbarbu@83.143.69.103
```

### Docker login failed
Verifică că `DOCKERHUB_TOKEN` este valid:
```bash
echo "TOKEN_VALUE" | docker login -u devprimetek --password-stdin
```

---

## 📞 Need Help?

Verifică logs:
- GitHub Actions: https://github.com/DevPrimeTek/nuviax-app/actions
- Server logs: `docker logs nuviax_api`
