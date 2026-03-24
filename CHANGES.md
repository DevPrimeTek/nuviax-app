# NuviaX — Changelog

Format: `[vX.X.X] — Data — Descriere`
Convenție: **feat** (funcționalitate nouă), **fix** (bug), **refactor**, **docs**, **chore** (infra/config)

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
