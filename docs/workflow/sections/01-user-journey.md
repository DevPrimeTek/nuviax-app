## 1. User Journey (End-to-End)

### CURRENT SYSTEM (REALITY)

### TARGET SYSTEM (FRAMEWORK)

### 1.1 Registration & Onboarding

1. User submits `POST /auth/register` (name, email, password)
2. Backend: bcrypt hash, insert `users`, generate JWT RS256 pair (access 15min / refresh 7d)
3. Welcome email sent via Resend (`email.go` → `WelcomeEmail`)
4. Frontend redirects → `/onboarding`
5. Onboarding page: user enters first goal title
6. `POST /goals/suggest-category` called with 2s hard timeout → Claude Haiku returns category suggestion or falls back silently
7. User selects category pill (HEALTH / CAREER / FINANCE / RELATIONSHIPS / LEARNING / CREATIVITY / OTHER) and optional `dominant_behavior_model` (G-11: ANALYTIC / STRATEGIC / TACTICAL / REACTIVE)
8. `POST /goals` creates entry in `global_objectives`

### 1.2 First Login

1. User submits `POST /auth/login` (email, password)
2. Backend: bcrypt verify, generate new JWT pair, store refresh token in Redis
3. Frontend stores access token in memory; refresh token in `httpOnly` cookie
4. `GET /settings` returns `theme` (dark/light) and language preference
5. Frontend applies `data-theme` + `nv_lang` from `localStorage` — anti-flash inline script runs before hydration
6. Redirect → `/dashboard`

### 1.3 Goal Setup

1. `GET /goals` returns list of user's `global_objectives`
2. User creates goal → `POST /goals` (title, category, deadline, optional `dominant_behavior_model`)
3. Engine Layer 0 (C1–C8) initializes base score via `engine.go`
4. `POST /goals/:id/sprint` creates first sprint in `sprints` table
5. Level 1 engine (`level1_structural.go`) generates initial tasks via Claude Haiku task generation
6. Scheduler cron (`scheduler.go`) generates daily tasks at midnight

### 1.4 Daily Loop

1. User opens `/today` → `GET /today` returns energy level + main tasks + personal tasks
2. User sets energy level (1–5) for the day
3. Task list rendered from active sprint's generated tasks
4. User completes tasks → `POST /today/complete/:id`
5. Each completion triggers server-side score recalculation (Level 2 engine, C19–C25)
6. End-of-day: scheduler runs daily check-in job; dashboard cache invalidated, checkpoint statuses updated ⚠️ regression event recording NOT IMPLEMENTED (see SA-3, Sprint 3.1)

### 1.5 Progress Review

1. `GET /goals/:id` returns goal with progress % and grade (A+/A/B/C/D) — opaque output only
2. `GET /goals/:id/visualization` returns chart data points (Level 5, `visualization.go`)
3. `/dashboard` aggregates all goals — grades and progress bars rendered client-side from server values
4. `/recap` provides sprint summary: completion rate, streak, energy average
5. `GET /profile/activity` returns 365-day activity data → rendered as 52-week heatmap (`ActivityHeatmap.tsx`)

### 1.6 SRM Triggers

1. Level 4 engine (`level4_regulatory.go`) evaluates SRM conditions after each score update
2. `SRMWarning.tsx` banner appears on dashboard when SRM is active
3. SRM levels: L1 (daily review) → L2 (weekly review) → L3 (monthly review), escalating
4. User completes SRM → `POST /srm` — score recalculated with regulatory adjustments
5. Successful SRM exits warning state; failed SRM may escalate level

### 1.7 Achievement & Ceremony Flow

1. Level 5 engine (`level5_growth.go`) evaluates achievement conditions post-score-update
2. Sprint close (scheduler cron) → `ApplyEvolveOverride()` runs for hybrid GO behavior models
3. Ceremony assigned: BRONZE / SILVER / GOLD / PLATINUM based on sprint performance
4. `GET /ceremonies/:goalId` returns latest ceremony for goal → `CeremonyModal.tsx` displayed on next login
5. `POST /ceremonies/:id/view` marks ceremony as seen
6. Achievements stored in `achievements` table; `GET /achievements` returns full badge grid

### 1.8 Growth & Visualization

1. Trajectory data accumulated across sprints in `level5_growth.go`
2. `GET /goals/:id/visualization` → `ProgressCharts.tsx` renders LineChart + BarChart (Recharts)
3. Profile page: avatar, stats, activity heatmap, preferences (theme, language)
4. `PATCH /settings` persists theme + language to DB (`users.theme` — migration 012)
5. All score components (drift, chaos_index, weights, thresholds) remain server-only — never exposed in API responses

---

