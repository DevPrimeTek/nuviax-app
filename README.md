# NuviaX - Growth Automation Platform

Platformă avansată de automatizare pentru creștere bazată pe framework-ul proprietar de engagement.

## 📁 Structura Proiectului

```
nuviax-app/
├── .github/
│   └── workflows/
│       ├── deploy.yml              # Backend deployment
│       └── deploy-frontend.yml     # Frontend deployment
├── backend/                         # Go API Server
│   ├── cmd/server/                 # Entry point
│   ├── internal/                   # Business logic
│   ├── pkg/                        # Shared packages
│   ├── Dockerfile
│   └── .dockerignore
├── frontend/
│   ├── app/                        # Main application (nuviax.app)
│   │   ├── src/
│   │   ├── package.json
│   │   └── Dockerfile
│   └── landing/                    # Landing page (nuviaxapp.com)
│       ├── src/
│       ├── package.json
│       └── Dockerfile
├── infra/                          # Infrastructure as Code
│   ├── docker-compose.yml          # Main services
│   ├── docker-compose.frontend.yml # Frontend overlay
│   ├── init-db.sql                 # Database schema
│   ├── .env.example                # Environment template
│   ├── deploy.sh                   # Deployment script
│   └── GITHUB_SECRETS.md           # Secrets documentation
└── README.md
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Git
- Node.js 20+ (pentru development local)
- Go 1.22+ (pentru development local)

### 1. Server Setup (One-time)

```bash
# Pe server ca user sbarbu
cd /var/www
git clone git@github.com:DevPrimeTek/nuviax-app.git wxr-nuviax
cd wxr-nuviax

# Rulează setup (generează secrets, pornește DB)
bash infra/setup-server.sh

# Configurează DNS A records → 83.143.69.103:
# - nuviax.app
# - www.nuviax.app
# - api.nuviax.app
# - nuviaxapp.com
# - www.nuviaxapp.com
```

### 2. GitHub Secrets Setup

Configurează în `Settings → Secrets and variables → Actions`:

| Secret | Valoare |
|--------|---------|
| `SSH_HOST` | `83.143.69.103` |
| `SSH_PORT` | `22` |
| `SSH_USER` | `sbarbu` |
| `SSH_KEY` | Conținutul din `~/.ssh/github_actions` |
| `DOCKERHUB_TOKEN` | Token DockerHub pentru `devprimetek` |
| `POSTGRES_PASSWORD` | Generat de `setup-server.sh` |
| `REDIS_PASSWORD` | Generat de `setup-server.sh` |
| `JWT_PRIVATE_KEY` | Generat de `setup-server.sh` |
| `JWT_PUBLIC_KEY` | Generat de `setup-server.sh` |
| `ENCRYPTION_KEY` | Generat de `setup-server.sh` |

### 3. Deploy

```bash
# Push pe main → deploy automat
git add .
git commit -m "Deploy NuviaX"
git push origin main

# Sau manual pe server:
cd /var/www/wxr-nuviax
bash infra/deploy.sh
```

## 🏗️ Arhitectură

### Stack Tehnologic

- **Backend**: Go 1.22 + Fiber
- **Frontend App**: Next.js 14 (React 18, TypeScript)
- **Landing**: Next.js 14 (Static)
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Proxy**: nginx-proxy + acme-companion (SSL automat)
- **Container**: Docker + Docker Compose

### Servicii

| Service | Port | Domain |
|---------|------|--------|
| API | 8080 | api.nuviax.app |
| App | 3000 | nuviax.app |
| Landing | 3001 | nuviaxapp.com |
| PostgreSQL | 5432 | Internal only |
| Redis | 6379 | Internal only |

### Rețele Docker

- `nginx_proxy` - Externă, shared cu Profixer
- `nuviax_net` - Internă, izolată

## 🔧 Development

### Backend Local

```bash
cd backend

# Install dependencies
go mod download

# Run locally (needs local PostgreSQL + Redis)
make run

# Run tests
make test

# Build Docker image
make docker-build
```

### Frontend Local

```bash
cd frontend/app

# Install dependencies
npm install

# Run dev server
npm run dev

# Build
npm run build
```

## 📊 Monitoring

### Health Checks

```bash
# API
curl https://api.nuviax.app/health
# Expected: {"status":"ok","db":true,"redis":true}

# App
curl https://nuviax.app

# Landing
curl https://nuviaxapp.com
```

### Logs

```bash
# Pe server
docker logs nuviax_api
docker logs nuviax_app
docker logs nuviax_landing
docker logs nuviax_db
docker logs nuviax_redis

# Follow logs
docker logs -f nuviax_api
```

## 🔐 Security

- JWT tokens cu RSA-4096
- AES-256 encryption pentru date sensibile
- PostgreSQL + Redis izolate în rețea internă
- SSL automat prin Let's Encrypt
- Secrets management prin GitHub Actions

## 📝 Deployment Workflow

1. **Push pe `main`** → GitHub Actions detectează schimbări
2. **Build**: 
   - Backend → `devprimetek/nuviax-api:latest`
   - App → `devprimetek/nuviax-app:latest`
   - Landing → `devprimetek/nuviax-landing:latest`
3. **Deploy**: SSH pe server, pull images, restart containers
4. **Health check**: Verifică că serviciile răspund
5. **Cleanup**: Șterge imagini vechi

## 🐛 Troubleshooting

### Container nu pornește

```bash
docker logs nuviax_api
docker compose -f infra/docker-compose.yml ps
```

### Database connection failed

```bash
# Verifică că PostgreSQL rulează
docker exec nuviax_db pg_isready -U nuviax

# Verifică parola în .env
grep POSTGRES_PASSWORD infra/.env
```

### SSL certificate issues

```bash
# Verifică nginx-proxy
docker logs nginx_proxy

# Verifică acme-companion
docker logs acme-companion
```

## 📞 Support

- **Repository**: https://github.com/DevPrimeTek/nuviax-app
- **Issues**: https://github.com/DevPrimeTek/nuviax-app/issues

## 📄 License

Proprietary - DevPrimeTek © 2025
