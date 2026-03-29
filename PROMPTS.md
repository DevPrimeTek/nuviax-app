# PROMPTS.md — NuviaX: Session-Ready Implementation Prompts

> Each prompt below is a complete, self-contained context block for a new Claude Code session.
> Copy the entire block (including the header) as your first message.
> Language: English (with Romanian comments where relevant).

---

## Sprint 3 — Task 1: i18n Translations (EN + RU)

```
Read CLAUDE.md first. Current version: 10.4.2. Branch: claude/i18n-translations-XXXXX. Task: implement i18n framework with EN and RU translations.

## Context
NuviaX frontend is Next.js 14 (App Router) with TypeScript + Tailwind.
Currently all UI text is hardcoded in Romanian directly in components.
No i18n library exists yet.

## Files to read (max 3)
1. frontend/app/app/today/page.tsx — proof-of-concept page to migrate
2. frontend/app/app/settings/page.tsx — contains settings.language field reference
3. frontend/app/lib/ — check if lib/ directory exists and what's in it

## What to implement
1. Create frontend/app/lib/i18n.ts
   - useTranslation() hook that reads language from user settings
   - Language detection order: settings.language → localStorage → 'ro' (default)
   - Type-safe: TranslationKey type derived from ro.json structure
   - No external library — keep it simple with a plain object lookup

2. Create translation files:
   - frontend/app/public/locales/ro.json — source of truth (Romanian)
   - frontend/app/public/locales/en.json — English
   - frontend/app/public/locales/ru.json — Russian
   Keys must cover all text in today/page.tsx only (proof of concept)

3. Migrate frontend/app/app/today/page.tsx to use useTranslation()
   - Replace hardcoded Romanian strings with t('key') calls
   - Do NOT migrate other pages — proof of concept only

4. Add 'language' to user settings type if not present
   - Check frontend/app/types/ or wherever UserSettings is typed
   - language field: 'ro' | 'en' | 'ru', default 'ro'

## Constraints
- No external i18n libraries (no next-intl, no i18next) — custom hook only
- Fallback: if key missing → return key itself (never crash)
- Do NOT add backend changes — language is frontend-only (stored in settings JSONB)
- Do NOT migrate all pages — today/page.tsx only as proof of concept

## After implementation
- Update README.md: add i18n to features
- Commit: feat: i18n framework + EN/RU translations (today page PoC)
- Update CLAUDE.md section 5: mark translation task with [x]
- Update ROADMAP.md Sprint 3: mark EN/RU translation tasks done
```

---

## Sprint 3 — Task 2: AI-Enhanced Onboarding

```
Read CLAUDE.md first. Current version: 10.4.2. Branch: claude/ai-onboarding-XXXXX. Task: improve GO creation onboarding with Haiku category suggestion.

## Context
NuviaX uses Claude Haiku 4.5 for task generation and GO analysis.
Existing AI client: backend/internal/ai/ai.go — read this file first.
Existing usage pattern: AnalyzeGOText() called from handlers.go.

## Files to read (max 3)
1. backend/internal/ai/ai.go — existing Haiku client + methods
2. frontend/app/app/onboarding/page.tsx — current onboarding flow
3. backend/internal/api/handlers/handlers.go — CreateGoal handler + AnalyzeGO

## What to implement

### Backend
1. In backend/internal/ai/ai.go, add method SuggestGOCategory(title, description string):
   - Prompt: given a goal title+description, suggest ONE category from:
     HEALTH, CAREER, FINANCE, RELATIONSHIPS, LEARNING, CREATIVITY, OTHER
   - Return: SuggestionResult{Category string, Confidence float64, Reasoning string}
   - Timeout: 2 seconds (strict — user is waiting in onboarding flow)
   - If timeout or error → return empty SuggestionResult (caller handles fallback)

2. Add new endpoint: POST /api/v1/goals/suggest-category
   - Body: {"title": "...", "description": "..."}
   - Response: {"category": "CAREER", "confidence": 0.85, "reasoning": "..."}
   - If AI unavailable → 200 with {"category": "", "confidence": 0, "reasoning": ""}
   - Protected by JWT middleware

### Frontend
3. In the onboarding GO classification step (onboarding/page.tsx):
   - After user types title+description → auto-call /goals/suggest-category
   - Show suggestion as pre-selected option with confidence badge
   - User can accept (click) or override by selecting different category
   - Loading state: spinner for max 2s, then show manual picker
   - If API returns empty category → skip suggestion, show manual picker directly

## Constraints
- Fallback is mandatory — onboarding must work without AI
- 2 second hard timeout on Haiku call
- Do NOT modify existing AnalyzeGO — this is a separate endpoint
- Graceful: if ANTHROPIC_API_KEY missing → skip suggestion silently

## After implementation
- Commit: feat: AI category suggestion in onboarding (Haiku, 2s timeout, fallback)
- Update CLAUDE.md section 5: mark AI onboarding task
- Update ROADMAP.md Sprint 3: mark onboarding task done
```

---

## Sprint 3 — Task 3: Activity Heatmap Statistics

```
Read CLAUDE.md first. Current version: 10.4.2. Branch: claude/activity-heatmap-XXXXX. Task: add GitHub-style activity heatmap to /profile page.

## Context
User activity is tracked in daily_metrics table (one row per user per day).
Schema reference: docs/database-reference.md
Frontend profile: frontend/app/app/profile/page.tsx

## Files to read (max 3)
1. docs/database-reference.md — find daily_metrics table schema
2. frontend/app/app/profile/page.tsx — current profile page
3. backend/internal/api/handlers/handlers.go — find existing profile/settings handlers

## What to implement

### Backend
1. Add endpoint: GET /api/v1/profile/activity
   - Returns last 365 days of activity data
   - Query daily_metrics for current user, group by date
   - Response: {"activity": [{"date": "2026-01-15", "score": 0.75, "tasks_completed": 4}]}
   - score = 0 means no activity that day
   - Protected by JWT

### Frontend
2. Create component: frontend/app/app/components/ActivityHeatmap.tsx
   - GitHub-style 52-week grid (columns = weeks, rows = days Mon-Sun)
   - Color scale: 5 levels from --bg-card (no activity) to accent green (high activity)
   - Use CSS variables from globals.css — do NOT hardcode colors
   - Tooltip on hover: "Jan 15 — 4 tasks, score 75%"
   - No external chart library — pure CSS grid + div elements

3. Integrate into frontend/app/app/profile/page.tsx
   - Add "Activity" section below existing profile content
   - Fetch from /profile/activity on component mount
   - Show loading skeleton while fetching

## Constraints
- No recharts or chart.js for this — pure CSS grid only
- Must work with existing CSS variable system (globals.css)
- Empty state: show grey grid if no activity data yet

## After implementation
- Update README.md: add /profile/activity endpoint
- Commit: feat: activity heatmap in profile (52-week GitHub-style grid)
- Update ROADMAP.md Sprint 3: mark heatmap task done
```

---

## Sprint 3 — Task 4: Dark/Light Theme Toggle

```
Read CLAUDE.md first. Current version: 10.4.2. Branch: claude/theme-toggle-XXXXX. Task: add dark/light theme toggle that persists across sessions.

## Context
NuviaX already has CSS variables for both themes in globals.css (light theme block exists).
Current state: dark theme only, no toggle UI.
Settings page: frontend/app/app/settings/page.tsx
User settings: stored in users.settings JSONB field.

## Files to read (max 3)
1. frontend/app/app/styles/globals.css — find existing light theme variables
2. frontend/app/app/components/AppShell.tsx — navigation where toggle goes
3. backend/internal/api/handlers/handlers.go — find PATCH /settings handler

## What to implement

### Frontend
1. In AppShell.tsx, add theme toggle button in top navigation:
   - Icon: sun (light) / moon (dark)
   - Position: top right, near user avatar
   - On click: toggle 'data-theme' attribute on <html> element
   - Persist choice in localStorage key 'nuviax-theme'

2. In frontend/app/app/layout.tsx (or RootLayout):
   - On initial load, read localStorage and set data-theme before first render
   - Prevent flash: inline script in <head> to set theme immediately

3. Sync with backend settings:
   - When user toggles → also PATCH /settings with {"theme": "light"|"dark"}
   - On app load, if localStorage empty → read from user settings API
   - Fire-and-forget: don't block UI on settings save

### Backend
4. Verify PATCH /settings handler accepts 'theme' field in settings JSONB
   - If not supported yet → add theme field to settings update logic
   - No migration needed (already JSONB)

## Constraints
- localStorage is primary source of truth (instant, no flicker)
- Backend sync is secondary (best-effort, fire-and-forget)
- Do NOT change existing CSS variables — only add data-theme switching
- Verify globals.css light theme block is complete before referencing it

## After implementation
- Commit: feat: dark/light theme toggle (localStorage + settings sync)
- Update ROADMAP.md Sprint 3: mark theme toggle done
```

---

## Sprint 4 — Task 1: Stripe Subscription Integration

```
Read CLAUDE.md first. Current version: 10.4.x. Branch: claude/stripe-integration-XXXXX. Task: implement Stripe Pro subscription ($9.99/month).

## Context
This is a major feature — read ROADMAP.md Sprint 4 section first.
NuviaX will have two tiers: Free (limited) and Pro ($9.99/month).
Payment processor: Stripe.

## Prerequisites before starting (check these exist)
- GitHub Secrets: STRIPE_SECRET_KEY, STRIPE_WEBHOOK_SECRET
- VPS .env: same keys present
- Stripe account created at dashboard.stripe.com
- Product + Price created in Stripe dashboard (get the price_id)

## Files to read (max 3)
1. backend/cmd/server/main.go — understand server initialization
2. backend/internal/api/server.go — where to add webhook route
3. docs/database-reference.md — users table structure

## What to implement (in order — this is multi-session work)

### Session A — Database + Backend skeleton
1. Create migration 012_stripe.sql:
   - ALTER TABLE users ADD COLUMN subscription_status VARCHAR(20) DEFAULT 'free'
   - ALTER TABLE users ADD COLUMN stripe_customer_id VARCHAR(50)
   - ALTER TABLE users ADD COLUMN trial_ends_at TIMESTAMPTZ
   - ALTER TABLE users ADD COLUMN subscription_ends_at TIMESTAMPTZ
   CHECK constraint: subscription_status IN ('free', 'trial', 'pro', 'cancelled')

2. Create backend/internal/stripe/stripe.go:
   - CreateCustomer(email string) (customerID string, err error)
   - CreateCheckoutSession(customerID, priceID, successURL, cancelURL string) (sessionURL string, err error)
   - GetSubscription(subscriptionID string) (status string, err error)
   - Use Stripe API directly (stdlib net/http, no SDK)

3. Add endpoints:
   - POST /api/v1/billing/checkout — create Stripe Checkout session
   - GET /api/v1/billing/status — return user's subscription_status + trial_ends_at
   - POST /api/v1/webhooks/stripe — handle checkout.session.completed, customer.subscription.deleted

### Session B — Free tier enforcement
4. Create middleware: backend/internal/api/middleware/subscription.go
   - ProOnly() middleware: check subscription_status IN ('trial', 'pro')
   - If free → return 403 {"error": "pro_required", "upgrade_url": "/billing"}

5. Apply ProOnly to:
   - POST /api/v1/goals (limit: free users can have max 1 active goal)
   - GET /api/v1/goals/analyze (AI analysis: pro only)

### Session C — Frontend billing page
6. Create frontend/app/app/billing/page.tsx:
   - Show current plan (free/trial/pro)
   - "Upgrade to Pro" button → calls /billing/checkout → redirect to Stripe
   - Show trial end date if on trial
   - Show subscription end date if cancelled

## Constraints
- NEVER store raw Stripe secret key in frontend
- Webhook must verify Stripe-Signature header
- Free tier degradation must be graceful (show upgrade prompt, not error)
- Test with Stripe test mode (test keys) before production

## After implementation
- Update README.md: add billing endpoints, env vars
- Update CHANGES.md with version bump
- Update CLAUDE.md section 4: mark Stripe as implemented
```

---

## Maintenance — Security: Rotate Exposed API Keys

```
URGENT — do this before any other task.

Read CLAUDE.md first. Current version: 10.4.2. Branch: claude/security-key-rotation-XXXXX.
Task: clean up accidentally committed API keys from .env.example.

## Problem
Real API keys were committed to .env.example in commits:
- f678ae9 (ANTHROPIC_API_KEY)
- 76166c2, ddd5e9e, 958a4ce (RESEND_API_KEY)

The git history is public. Both keys must be considered compromised.

## What to do (in this session)

1. Read infra/.env.example — verify current state of keys

2. Replace real key values with safe placeholders:
   ANTHROPIC_API_KEY=sk-ant-...YOUR_KEY_HERE
   RESEND_API_KEY=re_...YOUR_KEY_HERE

3. Commit: chore: remove real API keys from .env.example (use placeholders)

4. Instruct owner to:
   a. Go to console.anthropic.com → revoke old key → generate new one
   b. Go to resend.com → revoke old key → generate new one
   c. Update GitHub Secrets with new keys
   d. SSH to VPS → update infra/.env → restart nuviax_api container:
      docker compose -f infra/docker-compose.yml up -d --no-build nuviax_api

## Note
The old keys will remain in git history. To fully remove them would require
git history rewriting (filter-branch/BFG) which is destructive and breaks CI.
The priority is: revoke old keys immediately, commit clean .env.example now.
Inform the owner to decide on history rewriting separately.
```

---

## Docs — Update project-structure.md for v10.4.2

```
Read CLAUDE.md first. Current version: 10.4.2. Branch: claude/docs-update-XXXXX.
Task: update docs/project-structure.md to reflect v10.4.2 changes.

## Files to read (max 3)
1. docs/project-structure.md — current state (shows v10.4.1, missing migration 011)
2. backend/migrations/ — list all 12 migration files
3. backend/internal/engine/level5_growth.go lines 1-50 — check ApplyEvolveOverride signature

## What to update in docs/project-structure.md
1. Version header: v10.4.1 → v10.4.2
2. Migration count: 010 → 011 in the migrations listing
3. Add migration 011: 011_behavior_model.sql — dominant_behavior_model on global_objectives
4. In engine section, note ApplyEvolveOverride() added to level5_growth.go

## After
- Commit: docs: update project-structure.md for v10.4.2 (migration 011, ApplyEvolveOverride)
```

---

*Last updated: 2026-03-29 — v10.4.2 — initial prompt library created*
*Add new prompts here as features are planned. Mark with ✅ when session is complete.*
