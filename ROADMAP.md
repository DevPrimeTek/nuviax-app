# NuviaX — Roadmap de Dezvoltare

> Acest document reflectă starea curentă a proiectului și pașii următori în ordine de prioritate.
> Se actualizează la fiecare versiune majoră.
> **Ultima actualizare:** v10.5.0 — 2026-03-30

---

## Stare Curentă: v10.5.0

| Categorie | Status |
|-----------|--------|
| Backend Go — NUViaX Framework (40 componente) | ✅ 40/40 complet |
| Admin Panel (backend + frontend) | ✅ Funcțional — link nav condiționat per rol |
| Critical Gaps Stress Test (P0) | ✅ 5/5 rezolvate |
| Medium Gaps Stress Test (P1) | ✅ 12/12 implementate — **toate complete** |
| Bug-uri UI/UX B-2—B-11 | ✅ Toate rezolvate (v10.2.0) |
| Integrare AI (Claude Haiku 4.5) | ✅ Implementat + key configurat |
| Integrare Email (Resend) | ✅ Implementat (v10.3) + key configurat |
| Forgot/Reset parolă | ✅ Implementat (v10.3): endpoint + pagini frontend |
| G-11 Behavior Model dominance | ✅ Implementat (v10.4.2): migration 011 |
| Tema Light CSS | ✅ Implementat (variabile + bloc light) |
| Structura proiect curată | ✅ Fișiere duplicate/outdated șterse (v10.3.1) |
| Traduceri EN/RU | ❌ Neimplementat — **următor în Sprint 3** |
| Monetizare (Stripe) | 📅 Planificat Sprint 4 |

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
| G-11 | Behavior Model dominance — EVOLVE override GO hibride | `level5_growth.go`, `handlers.go`, `migrations/011_behavior_model.sql` | ✅ Implementat (v10.4.2) |
| G-12 | SRM flow complet — L2: confirmare user, L3: confirmare dublă | `srm.go`, `server.go` | ✅ Implementat |

**Migrații:** `010_p1_gaps.sql` — tabele `srm_events`, `reactivation_protocols`, `stagnation_events` | `011_behavior_model.sql` — câmp `dominant_behavior_model` pe `global_objectives`

---

## Sprint 3 — Traduceri + UX Completare ← ACTIV

*Obiectiv: aplicație utilizabilă internațional, experiență utilizator completă*

### Prioritate înaltă

- [x] **G-11: Behavior Model dominance** ✅ COMPLET (v10.4.2) — `dominant_behavior_model VARCHAR(20)` pe `global_objectives`; `ApplyEvolveOverride()` în `level5_growth.go`; migration 011
- [ ] **Traduceri EN** — framework i18n: toate textele interfață; `frontend/app/lib/i18n.ts` cu `useTranslation()` hook; detectare limbă din `settings.language`; proof of concept pe `today/page.tsx`
- [ ] **Traduceri RU** — același framework, aceleași fișiere locale

### Prioritate medie

- [ ] **Onboarding îmbunătățit** — la pasul de clasificare GO, Claude Haiku sugerează categoria (SMART analysis) → utilizatorul confirmă sau corectează; fallback 2s
- [ ] **Statistici personale avansate** — calendar activitate tip GitHub heatmap în `/profile`; date din `daily_metrics`
- [ ] **Dark/Light theme toggle** — buton în navigare; salvat în `localStorage` + `settings.theme`

### Prioritate scăzută

- [ ] **Export PDF raport lunar** — `/recap` → PDF cu progres lunar (bibliotecă `pdf-lib` sau similar)
- [ ] **Notificări push PWA** — `manifest.json` + service worker; web push pentru remindere zilnice (opt-in din `settings`)

> **Prompts sesiune gata:** Vezi `PROMPTS.md` pentru context complet per task.

---

## 🔧 Sprint 3.1 — System Alignment Fixes (CRITICAL)

*Audit arhitectural 2026-03-30 — gaps descoperite între framework REV 5.6 și implementare reală*

---

#### SA-1 — Populare `growth_trajectories` din scheduler

- **Problemă:** `fn_compute_growth_trajectory()` există în migration 006 dar nu e niciodată apelată din Go. `growth_trajectories` rămâne goală → `ProgressCharts.tsx` afișează un singur punct sintetic cu `actual_pct=0`, graficele sunt inutilizabile.
- **Fișiere:** `backend/internal/scheduler/scheduler.go`
- **Tip:** Scheduler
- **Complexitate:** Low
- **Prioritate:** CRITICAL

---

#### SA-2 — Acordare badge-uri achievements (C39)

- **Problemă:** `fn_award_achievement_if_earned()` există în migration 006 dar nu e apelată niciodată din Go. `achievement_badges` rămâne goală pentru toți utilizatorii. Pagina `/achievements` afișează listă goală.
- **Fișiere:** `backend/internal/scheduler/scheduler.go`
- **Tip:** Scheduler
- **Complexitate:** Low
- **Prioritate:** HIGH

---

#### SA-3 — Trigger automat SRM L1

- **Problemă:** SRM L2 și L3 au triggere implementate. SRM L1 nu e declanșat niciodată din niciun job sau handler. `CheckAndRecordRegressionEvent()` există în `level2_execution.go` dar nu e apelată. Cascada L1→L2→L3 este ruptă la primul nivel.
- **Fișiere:** `backend/internal/scheduler/scheduler.go` (`jobCheckDailyProgress`)
- **Tip:** Scheduler
- **Complexitate:** Medium
- **Prioritate:** CRITICAL

---

#### SA-4 — SRM L2: aplicare efectivă intensitate redusă (backend)

- **Problemă:** `ConfirmSRML2` marchează evenimentul ca confirmat dar nu creează nicio ajustare de context (`ENERGY_LOW`). Intensitatea zilei următoare rămâne nemodificată. Mesajul "Intensitatea sarcinilor va fi ajustată" este fals.
- **Fișiere:** `backend/internal/api/handlers/srm.go` (`ConfirmSRML2`)
- **Tip:** Backend
- **Complexitate:** Low
- **Prioritate:** CRITICAL

---

#### SA-5 — SRM L2: buton confirmare în frontend

- **Problemă:** `SRMWarning.tsx` afișează warning L2 dar nu are buton de confirmare. `POST /srm/confirm-l2/:goalId` există ca endpoint dar nu poate fi apelat din UI. Confirmarea L2 este imposibilă pentru utilizator.
- **Fișiere:** `frontend/app/components/SRMWarning.tsx`
- **Tip:** Frontend
- **Complexitate:** Low
- **Prioritate:** CRITICAL

---

#### SA-6 — SRM timeout fallback: implementare efectivă (C33)

- **Problemă:** `jobCheckSRMTimeouts` rulează orar, detectează timeout-uri L3 și calculează fallback-ul (L2/L1/PAUSE), dar nu aplică nicio schimbare de stare: `// TODO: engine.ApplySRMFallback(ctx, goalID, fallback)`. Goal-urile cu L3 neconfirmat rămân blocate indefinit.
- **Fișiere:** `backend/internal/scheduler/scheduler.go` (`jobCheckSRMTimeouts`)
- **Tip:** Scheduler + Backend
- **Complexitate:** Medium
- **Prioritate:** HIGH

---

#### SA-7 — Fix cron `jobRecalibrateRelevance` (G-9)

- **Problemă:** Expresia cron `"0 2 */90 * *"` — `*/90` în câmpul zi-a-lunii este invalid (lunile au max 31 zile). Job-ul nu rulează niciodată → Chaos Index și SRM L2 auto-trigger din G-9 nu funcționează.
- **Fișiere:** `backend/internal/scheduler/scheduler.go`
- **Tip:** Scheduler
- **Complexitate:** Low
- **Prioritate:** HIGH

---

| # | Task | Tip | Complexitate | Prioritate |
|---|------|-----|-------------|-----------|
| SA-1 | Populare `growth_trajectories` din scheduler | Scheduler | Low | CRITICAL |
| SA-2 | Acordare achievements (C39) | Scheduler | Low | HIGH |
| SA-3 | Trigger automat SRM L1 | Scheduler | Medium | CRITICAL |
| SA-4 | SRM L2 aplicare intensitate (backend) | Backend | Low | CRITICAL |
| SA-5 | SRM L2 buton confirmare (frontend) | Frontend | Low | CRITICAL |
| SA-6 | SRM timeout fallback efectiv (C33) | Scheduler | Medium | HIGH |
| SA-7 | Fix cron `jobRecalibrateRelevance` | Scheduler | Low | HIGH |

---

## Sprint 4 — Monetizare (Planificat)
*Doar după ce aplicația funcționează complet și are utilizatori reali*

### Stripe Integration

- [ ] **Stripe Checkout** — subscripție lunară ($9.99) / anuală ($89.99)
- [ ] **Free tier limits** — max 1 GO activ, fără vizualizare avansată, fără achievements
- [ ] **Pro tier** — max 3 GO active, toate funcționalitățile, AI analysis nelimitat
- [ ] **Trial 14 zile** — acces Pro complet fără card la înregistrare
- [ ] **Webhook Stripe** — `POST /api/v1/webhooks/stripe` — actualizare `users.subscription_status` în timp real
- [ ] **Billing portal** — redirect Stripe Customer Portal din `/settings`
- [ ] **Enforcement middleware** — verificare `subscription_status` la endpoint-uri Pro

### Backend

- [ ] **Migration 012** — `users.subscription_status ENUM`, `users.stripe_customer_id`, `users.trial_ends_at`
- [ ] **GitHub Secret nou** — `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`

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
