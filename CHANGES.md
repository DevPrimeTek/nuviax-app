# NuviaX — Changelog

Format: `[vX.X.X] — Data — Descriere`
Convenție: **feat** (funcționalitate nouă), **fix** (bug), **refactor**, **docs**, **chore** (infra/config)

---

## [v10.4.1] — 2026-03-26

### feat: Admin Panel standalone + Setup script proprietar

**Admin page redesign — layout independent de aplicație:**
- `frontend/app/app/admin/page.tsx` — eliminat `AppShell`; înlocuit cu `AdminShell` propriu
- Fundal dark `#0a0a0f` complet separat de tema aplicației principale
- Top bar minimal: logo NuviaX + badge **ADMIN** + username + butoane Refresh / Deconectare
- Pagina de eroare cu butoane „← Dashboard" și „Login" (nu mai arată navigarea principală)
- Culorile nu mai folosesc variabilele CSS ale aplicației (funcționează independent)

**Setup script proprietar:**
- `scripts/setup_admin.sh` — creare cont admin via API + setare `is_admin=TRUE` în DB
  - Pasul 1: `POST /auth/register` (gestionează crypto corect: bcrypt cost 14 + AES-256-GCM)
  - Pasul 2: `UPDATE users SET is_admin=TRUE WHERE email_hash=SHA256(email)` via docker exec
  - Pasul 3: verificare finală + instrucțiuni de acces
  - Graceful: dacă utilizatorul există deja (HTTP 409) → continuă direct cu setarea admin

---

## [v10.4.0] — 2026-03-26

### feat: P1 Gaps Stress Test (10/12) + Migration 010

**Gap-uri implementate:**
- **G-8** — `computeSprintInternal` — formulă completă 40/25/25/10 (completion rate + consistency + context + quality), nu mai returnează doar completion rate simplu
- **G-3** — `computeChaosIndex` + `CheckChaosIndex` — Chaos Index ≥ 0.40 trigger automat SRM L2 din scheduler
- **G-5** — `ConsecutiveInactiveDays` + `IsStagnant` + job 11 `jobDetectStagnation` — ≥5 zile inactive → event în `stagnation_events`
- **G-6** — `IsVelocityControlActive` — ALI_projected > 1.15 → taskCount-- în `GenerateDailyTasks`
- **G-2** — Focus Rotation — GO stagnant ≥5 zile → taskCount++ (max 3) în `GenerateDailyTasks`
- **G-1** — `ExtendSprintForPause` — `sprint.end_date += pause_days` în `SetPause` handler (deadline recalcul după pauză)
- **G-7** — `CheckReactivationEligibility` + `ProposeReactivation` + job 12 `jobProposeReactivation` — PAUSED GO cu ≥7 zile stabilitate → propunere automată reactivare
- **G-10** — Future Vault — `validateActivation` returnează `VAULT:` signal → goal creat ca WAITING automat (max 3 obiective active)
- **G-4** — `ComputeGORI` — Global Objective Relevance Index ca medie ponderată per user
- **G-12** — `ConfirmSRML2` — endpoint `POST /srm/confirm-l2/:goalId` — confirmare user L2; SRM flow complet L1/L2/L3
- **G-9** — `jobRecalibrateRelevance` extins: chaos_index stocat în `go_metrics` + trigger automat SRM L2 la CI ≥ 0.40

**Rămas (P2):**
- **G-11** — Behavior Model dominance — EVOLVE override GO hibride — necesită câmp DB suplimentar; planificat în Sprint 3

### chore: Migration 010
- Tabele noi: `srm_events` (audit trail SRM L1/L2/L3), `reactivation_protocols` (7-day stability tracking), `stagnation_events` (consecutive inactive days log)

### docs: README.md actualizat complet
- Structura proiect actualizată cu 12 scheduler jobs
- 10 migrații documentate
- Toate endpoint-urile sincronizate
- Regula de menținere README adăugată în CLAUDE.md

---

## [v10.3.1] — 2026-03-26

### fix: Admin Panel — is_admin expus în settings

- `GET /settings` returnează acum `is_admin` în response (din `UserSettings` struct)
- `AppShell.tsx` — link "Admin" vizibil **doar** pentru utilizatorii cu `is_admin=true`
- Link admin ascuns pentru utilizatori normali (nu apare în navigare)

### chore: Cleanup fișiere duplicate/outdated

- Șters: `NuviaX_UI_Mockup_v4.html` — mockup vechi (înlocuit de implementare reală)
- Șters: `ANALYSIS_REPORT.md` — raport inițial (integrat în CHANGES.md)
- Șters: `IMPLEMENTATION_CHECKLIST.md` — checklist vechi (înlocuit de ROADMAP.md)
- Șters: `TEST_REPORT.md` — raport generat automat (nu se ține în git)
- Șters: `frontend/infra/` — director duplicat (conținut mutat în `infra/`)
- Șters: `frontend/.github/workflows/` — workflows duplicate (există în `.github/workflows/`)

### docs: ROADMAP.md restructurat

- Sprint 1 și Sprint 2 marcate complet
- Sprint 3 (Traduceri + UX) și Sprint 4 (Monetizare) adăugate
- Decizii tehnice documentate

---

## [v10.3.0] — 2026-03-25

### feat: Integrare Email — Resend.com (E-1)

- `backend/internal/email/email.go` — client HTTP Resend API direct (stdlib, fără SDK extern)
  - `SendWelcomeEmail` — trimis la înregistrare (goroutine fire-and-forget)
  - `SendSprintCompleteEmail` — trimis la închidere sprint (din scheduler `jobCloseExpiredSprints`)
  - `SendPasswordResetEmail` — trimis la `POST /auth/forgot-password`
- Graceful degradation: dacă `RESEND_API_KEY` lipsește → log warning, fără erori aplicație

### feat: Forgot/Reset Password flow complet

- `POST /api/v1/auth/forgot-password` — timing-safe (mereu 200, previne user enumeration), token 1h TTL
- `POST /api/v1/auth/reset-password` — validare token single-use, actualizare parolă cu PBKDF2
- Frontend: `/auth/forgot-password` — formular email cu feedback vizual
- Frontend: `/auth/reset-password` — formular parolă nouă (citește token din URL query param)

### chore: Migration 009

- `password_reset_tokens` — `token_hash VARCHAR(64)`, `user_id UUID`, `expires_at TIMESTAMPTZ`, `used_at TIMESTAMPTZ`
- Index pe `token_hash` + constraint single-use (UPDATE `used_at`)

---

## [v10.2.0] — 2026-03-24

### fix: 10 Bug Fixes (B-2 through B-11) + AI Integration

**Bug Fixes:**
- **B-3** — Sprint zile: `daysLeft` folosea `goal.EndDate` (89 zile) → acum folosește `sprint.EndDate` (30 zile)
- **B-7** — Pagina Obiective: `GetGoals` returna array plat → acum returnează `{goals:[], waiting:[]}`
- **B-8** — Pagina Recap: `GET /recap/current` + `POST /goals/:id/recap` implementate complet
- **B-11** — CSS variabila `--ff-h` adăugată în `globals.css` (folosită în onboarding + profil)
- **B-5** — Energia zilnică: endpoint corectat `/today/energy` → `/context/energy`; mapare nivel `mid→normal`, `hi→high`; `goal_id` auto-detectat server-side
- **B-6** — Activități personale: adăugat input + buton "+" în `today/page.tsx` → `POST /today/personal`
- **B-9** — Setări conectate: schimbare parolă (modal + `POST /settings/password`) + export date (JSON download)
- **B-2** — Analiza GO: `AnalyzeGO` folosește Claude Haiku cu fallback pe analiza rule-based
- **B-4** — Activități zilnice: `generateTaskTexts` folosește Claude Haiku cu fallback pe template-uri statice
- **B-10** — Profil foto: avatar clickabil → upload `POST /settings/avatar`, stocare locală `/app/uploads/avatars/`

### feat: Integrare Claude Haiku 4.5
- `internal/ai/ai.go` — client HTTP direct (fără SDK), model `claude-haiku-4-5-20251001`
- Metode: `GenerateTaskTexts` (activități zilnice contextualizate) + `AnalyzeGO` (clasificare SMART)
- Graceful degradation: dacă `ANTHROPIC_API_KEY` lipsește → fallback automat pe reguli
- Engine expune `AnalyzeGOText()` public pentru handlers

### chore: Migration 008
- `avatar_url VARCHAR(500)` adăugat pe `users`
- Director upload: `/app/uploads/avatars/`

---

## [v10.1.0] — 2026-03-24

### feat: Admin Panel complet
- Pagină admin frontend (`/admin`) cu 4 tab-uri: Statistici, Utilizatori, Jurnal Audit, Sistem
- Backend: 8 rute `/api/v1/admin/*` securizate cu `AdminOnly` middleware (`is_admin=TRUE`)
- Statistici platformă: 20+ metrici (utilizatori, obiective, sprinturi, SRM, badge-uri)
- Management utilizatori: dezactivare / activare / promovare admin
- Resetare DB dev: `POST /admin/db/reset` (disponibil doar `APP_ENV=development`)
- Sistem: DB pool stats, scheduler jobs, versiune PostgreSQL

### fix: 5 Critical Gaps din Stress Test (P0)
- **GAP #14** — Pauze retroactive (max 48h); câmp `retroactive_start_date` în SetPause
- **GAP #15** — Detecție regresie (valoare sub sprint start); tabela `regression_events`
- **GAP #20** — Freeze traiectorie în Stabilization Mode; `FreezeExpectedTrajectory()`
- **GAP #8** — ALI per-goal vs total; breakdown complet în răspuns SRM
- **GAP #13** — ALI_current vs ALI_projected; câmpuri separate în response

### feat: ConfirmSRML3 — implementat complet
- Suspendă obiectivul (status PAUSED) + eveniment SRM L3 + freeze traiectorie

### chore: Migration 007
- `is_admin` pe `users`; `retroactive` pe `context_adjustments`
- Tabele noi: `regression_events`, `ali_snapshots`
- Coloane sprint: `expected_pct_frozen`, `frozen_expected_pct`
- Views admin: `v_admin_platform_stats`, `v_admin_user_list`
- Funcție: `fn_dev_reset_data(admin_id)` — dev only, cu protecție is_admin

### docs
- `CLAUDE.md` — fișier context master pentru continuarea sesiunilor (citit la start)
- `CHANGES.md` — rescris ca changelog structurat și complet

---

## [v10.0.0] — 2026-03-19

### feat: NUViaX Framework REV 5.6 — 40/40 componente implementate
- Engine Go cu 6 fișiere: `engine.go`, `level1_structural.go` → `level5_growth.go`
- **Layer 0 (C1-C8):** Drift, Chaos, Continuity, GORI, Visibility, Priority
- **Level 1 (C9-C18):** Sprint Architecture, Checkpoints, Smart Reactivation
- **Level 2 (C19-C25):** Execution Matrix, Daily Stack, Progress, Velocity
- **Level 3 (C26-C31):** Context Events, Energy, Pause Analytics, Rhythm, Consistency, Behavioral
- **Level 4 (C32-C36):** Adaptive Context, SRM, Stabilization, Reactivation, Regulatory
- **Level 5 (C37-C40):** Evolution Detection, Ceremonies, Achievements, Visualization

### feat: Level 5 Growth Orchestration
- `MarkEvolutionSprint` — detecție delta ≥5% → inserare `evolution_sprints`
- `GenerateCompletionCeremony` — tier BRONZE/SILVER/GOLD/PLATINUM din sprint score
- `GetUserAchievements` — query `achievement_badges` cu 10 tipuri de badge
- `GenerateProgressVisualization` — trajectory data + fallback snapshot live

### feat: Frontend componente Level 4/5
- `CeremonyModal.tsx` — tier colors, mark-viewed POST, achievement list
- `DashboardClientLayer.tsx` — polls `/ceremonies/unviewed` la mount
- `SRMWarning.tsx` — bannere L1 (galben) / L2 (portocaliu) / L3 (roșu) cu confirmare
- `ProgressCharts.tsx` — LineChart real vs așteptat + BarChart delta (Recharts 3.8)
- `GoalTabs.tsx` — tab switcher client Prezentare / Progres
- `achievements/page.tsx` — badge grid grupat pe categorii

### chore: Migration 006 (Level 5 Growth)
- Tabele: `growth_milestones`, `achievement_badges`, `completion_ceremonies`, `growth_trajectories`
- 7 views de analiză creștere
- Materialized view: `mv_user_stats` (refresh orar)
- Funcții: `fn_compute_growth_trajectory`, `fn_award_achievement_if_earned`
- Triggers: `trg_milestone_check`, `trg_achievement_award`, `trg_trajectory_snapshot`

### chore: Scheduler — 10 background jobs (cron, UTC)
- `00:00` — Generare activități zilnice
- `23:50` — Calcul scor zilnic
- `23:55` — Verificare progres
- `00:01` — Închidere etape expirate
- `*/90d 02:00` — Recalibrare relevanță
- `01:00` — Detecție evolution sprint
- `01:05` — Generare ceremonies
- `00:05` — Progres reactivare obiective
- `orar` — Verificare timeout SRM + Refresh matview

---

## [v9.x] — 2026-03-17 — 2026-03-18

### docs: Analiză completă 11 puncte (ANALYSIS_REPORT.md)
- Bug-uri critice identificate: B-3 (sprint days calcul greșit), B-7 (goals API mismatch), B-8 (recap 404)
- Bug-uri majore: B-5 (energy nivel), B-6 (personal tasks add), B-9 (settings incomplete), B-11 (CSS variabile lipsă)
- Bug-uri medii: B-2 (AI analysis), B-4 (task generation static), B-10 (foto profil)
- Nota: identificate în această versiune, nerezolvate

### feat: Framework REV 5.6 — baza Level 1-4
- Engine Level 1-4 implementat
- Migrations 001-005 aplicate (23 tabele, views, triggers)
- Handlers: auth, goals, tasks, sprints, context, settings
- Scheduler parțial (jobs 1-5)

---

## [v8.x – v1.x] — 2026-03-13 — 2026-03-17

### Early development — 10-15 iterații (nesistematizate)

**Arhitectură și infrastructură:**
- Structură Go (Fiber v2) + Next.js 14 inițializată
- Docker Compose: PostgreSQL 16 + Redis 7 + Go API + Next.js App + Landing
- GitHub Actions CI/CD: push `main` → build → DockerHub → SSH deploy VPS `83.143.69.103`
- Proxy: nginx-proxy + acme-companion (shared cu alte proiecte pe VPS)
- Domenii: `nuviax.app`, `api.nuviax.app`, `nuviaxapp.com`

**Securitate:**
- JWT RS256 (RSA 4096-bit), access token 15min, refresh 7 zile
- Criptare email AES-256-GCM
- Hashing parole PBKDF2 cu salt
- Rate limiting: 100 req/min global, 10 req/min auth endpoints
- Session management cu device fingerprinting + IP subnet

**Frontend baza:**
- Onboarding wizard (3 obiective cu activare secvențială)
- Dashboard cu sprint card, stats, SRM warnings
- Auth pages (login, register)
- Today/tasks, goals list/detail, settings, profile

**Corecții structurale:**
- Corectare nomenclatură: `web` → `app` în toate containerele și workflow-urile
- Fix Dockerfile handling `public/` folder
- Fix 404 pe deploy inițial
- README structurat cu documentație completă

---

*Bug-urile identificate în v9.x rămân deschise pentru rezolvare în v10.2+*
*Referință completă bug-uri: `ANALYSIS_REPORT.md`*
