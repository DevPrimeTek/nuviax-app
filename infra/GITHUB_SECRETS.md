# GitHub Secrets — NUViaX
# repo: devprimetek/nuviax-app
# Settings → Secrets and variables → Actions → New repository secret

## Secrete care SE REUTILIZEAZĂ (deja le ai la Profixer)

| Secret           | De unde îl iei                                                    |
|------------------|-------------------------------------------------------------------|
| SSH_DEPLOY_KEY   | `cat ~/.ssh/github_actions` pe server → copiezi tot, inclusiv header/footer |
| DOCKERHUB_TOKEN  | Copiezi valoarea din repo-ul Profixer (același cont devprimetek)  |

**SERVER_HOST** = `83.143.69.103`
**SERVER_USER** = `sbarbu`

Acestea nu sunt "secrete" sensibile dar le pui în Secrets ca să nu fie hardcodate în yml.

---

## Secrete NOI (specifice NUViaX)

| Secret            | Cum generezi                                      |
|-------------------|---------------------------------------------------|
| POSTGRES_PASSWORD | `openssl rand -base64 32`                         |
| REDIS_PASSWORD    | `openssl rand -base64 32`                         |
| JWT_PRIVATE_KEY   | Generat automat de setup-server.sh                |
| JWT_PUBLIC_KEY    | Generat automat de setup-server.sh                |
| ENCRYPTION_KEY    | Generat automat de setup-server.sh                |

---

## Pași exacti

### 1. SSH_DEPLOY_KEY — rulează pe server:
```bash
cat ~/.ssh/github_actions
```
Copiezi TOTUL (de la `-----BEGIN...` până la `-----END...`) și îl pui în secretul SSH_DEPLOY_KEY.

### 2. DOCKERHUB_TOKEN
Mergi la repo-ul Profixer pe GitHub → Settings → Secrets → copiezi valoarea DOCKERHUB_TOKEN.
Sau din DockerHub: hub.docker.com → contul tău → Security → Access Tokens.

### 3. Parolele noi (rulează pe server sau local):
```bash
echo "POSTGRES_PASSWORD=$(openssl rand -base64 32)"
echo "REDIS_PASSWORD=$(openssl rand -base64 32)"
```
Aceleași valori le pui și în /var/www/wxr-nuviax/infra/.env pe server.

### 4. JWT + ENCRYPTION — generate automat de setup-server.sh
Scriptul le generează și le afișează la final. Le copiezi în GitHub Secrets.

---

## Verificare finală după primul deploy:
```
https://api.nuviax.app/health  →  {"status":"ok","db":true,"redis":true}
https://nuviax.app             →  Frontend
```
