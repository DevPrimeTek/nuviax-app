# NuviaX — Roadmap de Dezvoltare

> Acest document reflectă starea curentă a proiectului și pașii următori în ordine de prioritate.
> Se actualizează la fiecare versiune majoră.
> **Ultima actualizare:** v10.3.1 — 2026-03-26

---

## Stare Curentă: v10.3.1

| Categorie | Status |
|-----------|--------|
| Backend Go — NUViaX Framework (40 componente) | ✅ 40/40 complet |
| Admin Panel (backend + frontend) | ✅ Funcțional — link nav condiționat per rol |
| Critical Gaps Stress Test (P0) | ✅ 5/5 rezolvate |
| Medium Gaps Stress Test (P1) | ⏳ 0/12 — **prioritate curentă** |
| Bug-uri UI/UX B-2—B-11 | ✅ Toate rezolvate (v10.2.0) |
| Integrare AI (Claude Haiku 4.5) | ✅ Implementat + graceful fallback |
| Integrare Email (Resend) | ✅ Implementat (v10.3): welcome + reset + sprint |
| Forgot/Reset parolă | ✅ Implementat (v10.3): endpoint + pagini frontend |
| Tema Light CSS | ✅ Implementat (variabile + bloc light) |
| Structura proiect curată | ✅ Fișiere duplicate/outdated șterse (v10.3.1) |
| Traduceri EN/RU | ❌ Neimplementat |
| Monetizare (Stripe) | 📅 Planificat târziu |

---

## ✅ Sprint 1 — COMPLET (v10.2.0 + v10.3.x)

### v10.1.0 — Admin Panel + P0 Gaps
- Admin panel complet (stats, users, audit, health, dev-reset)
- 5 gap-uri critice stress test rezolvate (retroactive pause, regression events, ALI, freeze trajectory)

### v10.2.0 — 10 Bug Fixes + Claude Haiku AI
- Toate bug-urile B-2—B-11 rezolvate
- `internal/ai/ai.go` — Claude Haiku 4.5 pentru task generation + GO analysis

### v10.3.0 — Resend Email Integration
- `internal/email/email.go` — client Resend (Welcome, PasswordReset, SprintComplete)
- `POST /auth/forgot-password` + `POST /auth/reset-password` (timing-safe)
- Pagini frontend `/auth/forgot-password` + `/auth/reset-password`
- Migration 009: `password_reset_tokens`

### v10.3.1 — Admin Fix + Cleanup
- `is_admin` expus în `/settings` response
- Link "Admin" în navigare — vizibil **doar** pentru admini
- Fișiere duplicate/outdated șterse: mockup, checklist, raport vechi, duplicate infra

---

## 🎯 Sprint 2 — P1 Gaps Stress Test (CURENT)
*12 gap-uri medii din simulare 120 zile — toate cu impact real pe comportamentul framework-ului*

Valori de referință (din simulare):
- Retroactive window: **48h**
- Reactivation stability: **7 zile**
- Stagnation threshold: **5 zile consecutive**
- Chaos Index L2 threshold: **0.40**
- ALI Ambition Buffer: **1.0 – 1.15**
- Evolution delta: **≥5%**

| # | Gap | Fișier de modificat | Prioritate |
|---|-----|--------------------|-----------:|
| G-1 | Deadline recalcul după pauză — extinde `end_date` sprint cu zilele de pauză | `level1_structural.go` | 🔴 P1 |
| G-2 | Focus Rotation — redirecționează atenția spre GO cu stagnare >5 zile | `engine.go`, `level2_execution.go` | 🔴 P1 |
| G-3 | Chaos Index formula exactă — 0.40 threshold trigger SRM L2 | `level2_execution.go` | 🔴 P1 |
| G-4 | GORI calculation complet — ponderi per sprint (completion + consistency + progress + energy) | `engine.go` | 🔴 P1 |
| G-5 | Stagnation detection granular — 5 zile consecutive fără progres | `level2_execution.go`, `scheduler.go` | 🟠 P1 |
| G-6 | Velocity Control activare — când ALI_projected > 1.15 | `level1_structural.go` | 🟠 P1 |
| G-7 | Reactivation Protocol — 7 zile stabilitate → propunere reactivare | `level4_regulatory.go` | 🟠 P1 |
| G-8 | Sprint score formula completă — 40% completion, 25% consistency, 25% progress, 10% energy | `engine.go`, `level1_structural.go` | 🔴 P1 |
| G-9 | Annual Relevance Recalibration — la 90 zile per GO activ | `scheduler.go`, `level3_adaptive.go` | 🟡 P2 |
| G-10 | Future Vault cu recalibration — max 3 active, restul în Vault automat | `level4_regulatory.go` | 🟡 P2 |
| G-11 | Behavior Model dominance — EVOLVE override la GO-uri hibride | `engine.go` | 🟡 P2 |
| G-12 | SRM flow complet — L1: automat, L2: confirmare user, L3: confirmare dublă | `srm.go`, `level4_regulatory.go` | 🟠 P1 |

---

## Sprint 3 — Traduceri + UX Completare

- [ ] **Traduceri EN** — textele interfețe RO → EN (framework `T{}` există în AppShell)
- [ ] **Traduceri RU** — RO → RU
- [ ] **Onboarding îmbunătățit** — AI suggestions la clasificare GO
- [ ] **Notificări push** — PWA web push pentru reminders zilnice
- [ ] **Statistici personale avansate** — calendar activitate în pagina profil

---

## Sprint 4 — Monetizare (Planificat)
*Doar după ce aplicația funcționează complet și are utilizatori reali*

- [ ] **Stripe integration** — subscripție lunară/anuală
- [ ] **Free tier limits** — max 1 GO, fără vizualizare, fără achievements
- [ ] **Pro tier** — 3 GO, toate funcționalitățile
- [ ] **Trial 14 zile** — acces complet fără card
- [ ] **Webhook Stripe** — actualizare status subscripție în timp real

---

## Decizii Tehnice

| Subiect | Decizie | Status |
|---------|---------|--------|
| Stocare foto profil | Local VPS `/app/uploads/avatars/` | ✅ Implementat — upgrade R2 mai târziu |
| Email tranzacțional | Resend.com (3K/lună gratis) | ✅ Implementat |
| AI provider | Anthropic Claude Haiku 4.5 ($4-5/lună) | ✅ Implementat |
| Mobile | PWA pentru v1, React Native mai târziu | 📅 Planificat |
| Analytics | Plausible (GDPR, ieftin) | 📅 Planificat |
| Full-text search | PostgreSQL tsvector | 📅 Planificat |

---

*Actualizare roadmap: la fiecare versiune majoră sau decizie arhitecturală importantă*
