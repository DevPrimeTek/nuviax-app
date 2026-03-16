# NuviaX App

Platform de management obiective bazată pe NuviaX Framework rev5.6.

## 📦 Structură

\\\
nuviax-app/
├── backend/          # Go API (Fiber)
├── frontend/
│   ├── app/         # Next.js 14 - Aplicația principală
│   └── landing/     # Next.js 14 - Landing page
├── infra/           # Docker Compose, deployment
├── deploy/          # Scripturi deployment (local)
└── .github/         # CI/CD workflows
\\\

## 🚀 Links

- **App:** https://nuviax.app
- **Landing:** https://nuviaxapp.com
- **API:** https://api.nuviax.app
- **Repo:** https://github.com/DevPrimeTek/nuviax-app

## 🛠️ Tech Stack

- Frontend: Next.js 14, React, TypeScript, Tailwind
- Backend: Go 1.22, Fiber, PostgreSQL, Redis
- Infrastructure: Docker, GitHub Actions, VPS

## 📋 Deployment

\\\powershell
cd deploy
.\Deploy-NuviaX-v9.0.ps1
\\\

## 📊 Changelog
### v10.0.0 - 16.03.2026
- 🚀 Deployment automat v10.0.0
- ✅ Fix: 404 error resolution
- ✅ Git sync și merge automat
- ✅ Backup pre-deployment
- ✅ Public assets verification


### v9.0 - 16.03.2026
- ✅ FIX CRITICAL: Dockerfile public/ folder handling
- ✅ Analiză completă structură proiect
- ✅ Creare automată foldere lipsă (public/, styles/)
- ✅ Verificare și creare favicon.ico
- ✅ README structurat permanent

**Versiune:** 10.0.0 | **Status:** ✅ Production Ready
