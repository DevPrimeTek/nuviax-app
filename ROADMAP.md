# NuviaX — Roadmap de Dezvoltare

> Acest document reflectă starea curentă a proiectului și pașii următori în ordine de prioritate.
> Se actualizează la fiecare versiune majoră.
> **Ultima actualizare:** v10.3.1 — 2026-03-26

---

## Stare Curentă: v10.4.0

| Categorie | Status |
|-----------|--------|
| Backend Go — NUViaX Framework (40 componente) | ✅ 40/40 complet |
| Admin Panel (backend + frontend) | ✅ Funcțional — link nav condiționat per rol |
| Critical Gaps Stress Test (P0) | ✅ 5/5 rezolvate |
| Medium Gaps Stress Test (P1) | ✅ 10/12 implementate — **G-9, G-11 rămase** |
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

### v10.4.0 — Sprint 2: P1 Gaps (10/12)
- G-8: `computeSprintInternal` — formulă completă 40/25/25/10 (nu mai returnează doar completion rate)
- G-3: `computeChaosIndex` + `CheckChaosIndex` — CI ≥ 0.40 → SRM L2 automat din scheduler
- G-5: `ConsecutiveInactiveDays` + `IsStagnant` + job 11 `jobDetectStagnation`
- G-6: `IsVelocityControlActive` — ALI_projected > 1.15 → taskCount-- în `GenerateDailyTasks`
- G-2: Focus Rotation — stagnare ≥5 zile → taskCount++ (max 3)
- G-1: `ExtendSprintForPause` — sprint.end_date += pause_days în `SetPause` handler
- G-7: `CheckReactivationEligibility` + `ProposeReactivation` + job 12 `jobProposeReactivation`
- G-10: Future Vault — validateActivation returnează `VAULT:` signal → goal creat ca WAITING
- G-4: `ComputeGORI` — Global Objective Relevance Index per user
- G-12: `ConfirmSRML2` — endpoint nou `POST /srm/confirm-l2/:goalId`
- G-9: `jobRecalibrateRelevance` extins cu chaos_index storage + trigger automat SRM L2
- Migrație 010: `srm_events`, `reactivation_protocols`, `stagnation_events`

---

## ✅ Sprint 2 — P1 Gaps Stress Test (v10.4.0)
*12 gap-uri medii din simulare 120 zile — 10/12 implementate*

Valori de referință (din simulare):
- Retroactive window: **48h**
- Reactivation stability: **7 zile**
- Stagnation threshold: **5 zile consecutive**
- Chaos Index L2 threshold: **0.40**
- ALI Ambition Buffer: **1.0 – 1.15**
- Evolution delta: **≥5%**

| # | Gap | Fișier modificat | Status |
|---|-----|-----------------|-------:|
| G-1 | Deadline recalcul după pauză — extinde `end_date` sprint cu zilele de pauză | `level1_structural.go`, `handlers.go` | ✅ Implementat |
| G-2 | Focus Rotation — task extra pentru GO stagnant >5 zile | `engine.go` (GenerateDailyTasks) | ✅ Implementat |
| G-3 | Chaos Index formula — 0.40 threshold trigger SRM L2 automat | `level2_execution.go`, `scheduler.go` | ✅ Implementat |
| G-4 | GORI calculation — weighted average across active GOs | `engine.go` (ComputeGORI) | ✅ Implementat |
| G-5 | Stagnation detection — 5 zile consecutive → `stagnation_events` | `level2_execution.go`, `scheduler.go` | ✅ Implementat |
| G-6 | Velocity Control — ALI_projected > 1.15 → reduce task count | `level2_execution.go`, `engine.go` | ✅ Implementat |
| G-7 | Reactivation Protocol — 7 zile stabilitate → propunere automată | `level4_regulatory.go`, `scheduler.go` | ✅ Implementat |
| G-8 | Sprint score formula completă — 40/25/25/10 în `computeSprintInternal` | `level2_execution.go` | ✅ Implementat |
| G-9 | Annual Relevance Recalibration + Chaos Index storage în go_metrics | `scheduler.go` | ✅ Implementat |
| G-10 | Future Vault — max 3 active, goal nou → WAITING automat | `level4_regulatory.go`, `handlers.go` | ✅ Implementat |
| G-11 | Behavior Model dominance — EVOLVE override GO hibride | `engine.go` | ⏳ P2 — necesită câmp DB |
| G-12 | SRM flow complet — L2: confirmare user, L3: confirmare dublă | `srm.go`, `server.go` | ✅ Implementat |

**Migrație nouă:** `010_p1_gaps.sql` — tabele `srm_events`, `reactivation_protocols`, `stagnation_events`

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
