# NUViaX — Fișiere Corectate pentru GitHub

## 📦 Conținut

Acest folder conține toate fișierele corectate pentru proiectul NuviaX, cu nomenclatură consistentă (`app` în loc de `web`).

## 🚀 Cum să Folosești

### Opțiunea 1: Înlocuire Manuală (Recomandat)

1. **Backup**:
   ```bash
   cd /path/to/nuviax-app
   git checkout -b backup-$(date +%Y%m%d)
   git push origin backup-$(date +%Y%m%d)
   ```

2. **Copiază fișierele**:
   ```bash
   # Workflows
   cp -f nuviax-corrected-files/.github/workflows/* .github/workflows/
   
   # Backend
   cp -f nuviax-corrected-files/backend/Dockerfile backend/
   cp -f nuviax-corrected-files/backend/.dockerignore backend/
   
   # Frontend
   cp -f nuviax-corrected-files/frontend/app/Dockerfile frontend/app/
   cp -f nuviax-corrected-files/frontend/landing/Dockerfile frontend/landing/
   
   # Infrastructure
   cp -f nuviax-corrected-files/infra/* infra/
   
   # Root
   cp -f nuviax-corrected-files/README.md .
   cp -f nuviax-corrected-files/.gitignore .
   ```

3. **Review & Commit**:
   ```bash
   git status
   git diff
   git add .
   git commit -m "fix: correct structure and nomenclature (web → app)"
   git push origin main
   ```

### Opțiunea 2: Clone Fresh (Pentru Testing)

```bash
# Creează un director nou
mkdir nuviax-app-new
cd nuviax-app-new

# Copiază fișierele corectate
cp -r /path/to/nuviax-corrected-files/* .

# Adaugă restul codului din repository-ul existent
# (backend/internal, frontend/app/src, etc.)

# Init git și push
git init
git add .
git commit -m "Initial commit with corrected structure"
git remote add origin git@github.com:DevPrimeTek/nuviax-app.git
git push -f origin main
```

## ⚠️ IMPORTANT Înainte de Push

1. **Configurează GitHub Secrets** (vezi `infra/GITHUB_SECRETS.md`)
2. **Verifică DNS** (A records către 83.143.69.103)
3. **Testează local** dacă este posibil

## 📋 Fișiere Incluse

```
.
├── .github/workflows/
│   ├── deploy.yml
│   └── deploy-frontend.yml
├── backend/
│   ├── Dockerfile
│   └── .dockerignore
├── frontend/
│   ├── app/Dockerfile
│   └── landing/Dockerfile
├── infra/
│   ├── docker-compose.yml
│   ├── docker-compose.frontend.yml
│   ├── .env.example
│   ├── init-db.sql
│   ├── deploy.sh
│   └── GITHUB_SECRETS.md
├── README.md
├── .gitignore
└── CHANGES.md
```

## ✅ Checklist Post-Upload

- [ ] Fișiere încărcate pe GitHub
- [ ] GitHub Secrets configurate
- [ ] DNS configurat
- [ ] Push pe main executat
- [ ] GitHub Actions rulează cu succes
- [ ] Health checks OK

## 📞 Support

Consultă `CHANGES.md` pentru detalii despre modificări.
