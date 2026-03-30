# docs/user-workflow.md — NuviaX User Workflow

> Version: 10.5.0 | Last updated: 2026-03-30

---

## 1. User Journey (End-to-End)

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
4. User completes tasks → `POST /tasks/:id/complete`
5. Each completion triggers server-side score recalculation (Level 2 engine, C19–C25)
6. End-of-day: scheduler runs daily check-in job; missed tasks recorded as regression events (`level2_execution.go`)

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
4. `GET /ceremonies/latest` returns unviewed ceremony → `CeremonyModal.tsx` displayed on next login
5. `POST /ceremonies/:id/viewed` marks ceremony as seen
6. Achievements stored in `achievements` table; `GET /achievements` returns full badge grid

### 1.8 Growth & Visualization

1. Trajectory data accumulated across sprints in `level5_growth.go`
2. `GET /goals/:id/visualization` → `ProgressCharts.tsx` renders LineChart + BarChart (Recharts)
3. Profile page: avatar, stats, activity heatmap, preferences (theme, language)
4. `PATCH /settings` persists theme + language to DB (`users.theme` — migration 012)
5. All score components (drift, chaos_index, weights, thresholds) remain server-only — never exposed in API responses

---

## 2. Goal Creation Flow

### 2.1 Onboarding Goal Wizard
### 2.2 AI Category Suggestion
### 2.3 Manual Category Selection
### 2.4 Behavior Model Selection (G-11)
### 2.5 Goal Submission & Validation
### 2.6 Initial Score Calculation

---

## 3. Daily Execution Flow

### 3.1 Today Page Overview
### 3.2 Task List Rendering
### 3.3 Task Completion
### 3.4 Daily Check-in
### 3.5 Score Recalculation Trigger
### 3.6 Progress Feedback (Grade, %)

---

## 4. SRM Flow (L1–L3)

### 4.1 SRM Entry Points
### 4.2 L1 — Daily Review
### 4.3 L2 — Weekly Review
### 4.4 L3 — Monthly Review
### 4.5 SRM Submission
### 4.6 Score Impact After SRM

---

## 5. Achievement Flow

### 5.1 Achievement Trigger Conditions
### 5.2 Ceremony Tiers (Tier 1–3)
### 5.3 Badge Award & Display
### 5.4 Achievement History (/profile)

---

## 6. Visualization Flow

### 6.1 Progress Bar & Grade Display
### 6.2 Activity Heatmap (52-week)
### 6.3 Goal Progress Cards
### 6.4 Profile Stats Overview
### 6.5 Dark/Light Theme Rendering

---

## 7. Test Scenarios

### 7.1 Happy Path — New User Full Journey
### 7.2 User With No Goals
### 7.3 Missed Daily Check-in
### 7.4 AI Suggestion Timeout (Fallback)
### 7.5 SRM Completion After Missed Period
### 7.6 Achievement Unlock Edge Case
### 7.7 Theme Persistence Across Sessions
### 7.8 Language Switch (EN / RU / RO)

---

## 8. Critical Checkpoints

### 8.1 Server-Side Calculation Enforcement
### 8.2 Opaque API Response Validation
### 8.3 JWT Auth on All Protected Routes
### 8.4 Admin 404 (Non-Admin Access)
### 8.5 Graceful Degradation (AI / Email Down)
### 8.6 Timing-Safe Forgot Password
