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
├── deploy/          # Scripturi deployment (local only)
└── .github/         # CI/CD workflows
\\\

## 🚀 Links

- **Live App:** https://nuviax.app
- **Landing:** https://nuviaxapp.com
- **API:** https://api.nuviax.app
- **Repository:** https://github.com/DevPrimeTek/nuviax-app

## 🛠️ Tech Stack

- **Frontend:** Next.js 14, React, TypeScript, Tailwind CSS
- **Backend:** Go 1.22, Fiber, PostgreSQL, Redis
- **Infrastructure:** Docker, GitHub Actions, VPS
- **Deployment:** PowerShell scripts, automated CI/CD

## 📋 Development

### Prerequisites

- Docker Desktop
- Git
- PowerShell 7.4.5 (Windows)

### Deployment

\\\powershell
# Navighează în folder deploy
cd deploy

# Rulează script deployment
.\Deploy-NuviaX-v8.3.ps1
\\\

## 📊 Changelog

<!-- Changelog entries managed by deploy script -->

### v8.3 - 16.03.2026
- ✅ Fix TypeScript errors (goals/page.tsx)
- ✅ Structură CSS corectată (packages → frontend/styles)
- ✅ README.md structurat
- ✅ Force sync GitHub cu local

---

**Versiune curentă:** 8.3  
**Ultima actualizare:** 16 March 2026  
**Status:** ✅ Production