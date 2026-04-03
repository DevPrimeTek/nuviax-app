# CLAUDE.md — NuviaX: Context Master

> **Read this file at the START of every new session. Confirm version and branch before any task.**
> Extended docs → `docs/` (not needed every session — use as reference only).

---

## 0. New Session Protocol

**Mandatory steps at start:**

```bash
# 1. Check repo state
git status
git log --oneline -5
git branch --show-current

# 2. Confirm before starting the task:
# "Read CLAUDE.md. Current version: X.X.X. Active branch: Y. Task: Z."
```

**Token rules:**
- Read **max 3 files** per task — do not explore globally
- Always specify exact path + function/line of interest
- One task per session — commit, then new session
- Use `/compact` after each major subtask (not at the end)
- **DO NOT read:** `node_modules/`, `vendor/`, `.next/`, `build/`, `dist/`

---

## 1. What is NuviaX

**NuviaX** is a SaaS platform for personal and professional goal management, based on the **NUViaX Growth Framework REV 5.6** — proprietary system with 5 levels (Layer 0 + Level 1–5) and 40 components (C1–C40).

**Core principle:** All calculations run exclusively server-side. The client receives only opaque results (%, grades A/B/C/D, task lists). No formula leaves the Go engine.

| Product | URL |
|---------|-----|
| App | `nuviax.app` (Next.js) |
| Landing | `nuviaxapp.com` (Next.js static) |
| API | `api.nuviax.app` (Go) |

**Owner:** DevPrimeTek — `github.com/DevPrimeTek/nuviax-app`
**Current version:** `10.4.2`
**Dev branch:** `claude/*` → PR → `main`

---

## 2. Tech Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Backend API | Go + Fiber v2 | Go 1.22, Fiber 2.52 |
| Database | PostgreSQL | 16 (Docker) |
| Cache/Sessions | Redis | 7 (Docker) |
| Frontend App | Next.js + TypeScript + Tailwind | Next.js 14, React 18 |
| Frontend Landing | Next.js (static) | Next.js 14 |
| Auth | JWT RS256 (RSA 4096-bit) | access: 15min, refresh: 7 days |
| Email | Resend.com (transactional) | — |
| AI | Anthropic Claude Haiku 4.5 | `claude-haiku-4-5-20251001` |
| CI/CD | GitHub Actions → DockerHub → VPS SSH | — |
| Proxy | nginx-proxy + acme-companion (jwilder) | shared |
| Deploy path | `/var/www/wxr-nuviax/` on VPS | — |

**Infrastructure keys status (as of 2026-03-29):**
- `ANTHROPIC_API_KEY` — ✅ Configured (GitHub Secrets + `.env`)
- `RESEND_API_KEY` — ✅ Configured (GitHub Secrets + `.env`)
- `EMAIL_FROM` — ✅ `noreply@nuviax.app`
- ⚠️ **Action required:** Rotate both API keys — real values were accidentally committed in `.env.example` (commits 76166c2, f678ae9). Revoke old keys on Anthropic Console and Resend Dashboard, generate new ones, update GitHub Secrets and VPS `.env`.

---

## 3. Model Selection

| Task | Recommended model |
|------|------------------|
| Code implementation, bugfix, routine refactoring | **Claude Sonnet** |
| Architectural decisions, critical review, new system design | **Claude Opus** |
| Rename, format, small edits under 20 lines | **Claude Haiku** |

**Rule:** Always start with **Sonnet**. Switch to Opus only if blocked on something architecturally complex.

---

## 4. Current State

**Version:** `10.5.0` | **Active sprint:** Sprint 4

| Category | Status |
|----------|--------|
| Framework REV 5.6 (40 components) | ✅ 40/40 complete |
| P0 Gaps (stress test) | ✅ 5/5 resolved (v10.1) |
| P1 Gaps (stress test) | ✅ 12/12 complete (v10.4.0 + v10.4.2) |
| Bugs B-2—B-11 | ✅ All resolved (v10.2) |
| AI Claude Haiku | ✅ Implemented (v10.2) — key configured |
| Email Resend | ✅ Implemented (v10.3) — key configured |
| Forgot/Reset password | ✅ Implemented (v10.3) |
| G-11 Behavior Model | ✅ Implemented (v10.4.2) — migration 011 |
| Admin panel | ✅ Implemented — manual admin account setup needed on VPS |
| Translations EN/RU | ✅ Implemented (v10.5.0) — today page PoC |
| AI Category Suggestion (Onboarding) | ✅ Implemented (v10.5.0) — 2s timeout, fallback |
| Activity Heatmap (/profile) | ✅ Implemented (v10.5.0) — 52-week GitHub-style grid |
| Dark/Light Theme (persistence) | ✅ Implemented (v10.5.0) — localStorage + backend (migration 012) |
| SA-7 cron fix | ✅ Fixed — `jobRecalibrateRelevance` now runs weekly (`0 2 * * 0`) |
| CE-1 table name fix | ✅ Fixed — trajectory query uses `global_objectives` (not `goals`) |
| SA-1 trajectory call | ✅ Fixed — `jobComputeDailyScore` now calls `fn_compute_growth_trajectory` |
| SA-3 SRM L1 auto-trigger | ✅ Fixed — `jobDetectStagnation` now writes `srm_events` L1 after 5 inactive days |
| Stripe monetization | 📅 Planned Sprint 4 |

---

## 5. Current Sprint — Sprint 3

### Tasks in priority order

**G-11 — Behavior Model dominance** ✅ COMPLETE (v10.4.2)
- ✅ New field: `dominant_behavior_model VARCHAR(20)` on `global_objectives`
- ✅ Migration: `011_behavior_model.sql`
- ✅ `level5_growth.go`: `ApplyEvolveOverride()` for hybrid GOs (ANALYTIC/STRATEGIC/TACTICAL/REACTIVE)
- ✅ `handlers.go`: `CreateGoal` accepts optional `dominant_behavior_model`
- ✅ Models, DB queries: full `dominant_behavior_model` support

**i18n Translations EN + RU** ✅ COMPLETE (v10.5.0)
- ✅ `frontend/app/lib/i18n.ts` — `useTranslation()` hook, no external lib
- ✅ `public/locales/ro.json`, `en.json`, `ru.json` — all keys for today page
- ✅ `today/page.tsx` migrated as proof of concept
- Language detection: `localStorage('nv_lang')` → 'ro' default

**Improved AI Onboarding** ✅ COMPLETE (v10.5.0)
- ✅ `ai.go`: `SuggestGOCategory()` — 2s hard timeout, returns empty on failure
- ✅ `handlers.go`: `SuggestCategory` — POST /goals/suggest-category
- ✅ `onboarding/page.tsx`: debounced auto-suggest + category pill selector
- Categories: HEALTH, CAREER, FINANCE, RELATIONSHIPS, LEARNING, CREATIVITY, OTHER

**Personal activity statistics heatmap** ✅ COMPLETE (v10.5.0)
- ✅ `handlers.go`: `GetProfileActivity` — GET /profile/activity (last 365 days)
- ✅ `ActivityHeatmap.tsx` — pure CSS grid, 52 weeks, color scale, hover tooltip
- ✅ `profile/page.tsx` — heatmap section added below preferences

**Dark/Light theme toggle** ✅ COMPLETE (v10.5.0)
- ✅ `AppShell.tsx`: toggle button already present (sun/moon icon)
- ✅ `layout.tsx`: anti-flash inline script already present
- ✅ `handlers.go`: `UpdateSettings` now accepts + persists `theme`
- ✅ `GetSettings` returns `theme` from DB; `012_theme.sql` migration added

### Sprint 4 (next)
- Stripe: Pro subscription ($9.99/month) + Free tier limits + 14-day trial
- PWA + Push notifications
- Monthly PDF report export

---

## 6. Development Workflow

```bash
# Start session
git checkout -b claude/feature-name-XXXXX

# End session
git add [specific files — NOT git add .]
git commit -m "feat/fix/docs: clear description"
git push -u origin claude/feature-name-XXXXX
# → open PR to main on GitHub
```

**Commit conventions:**
```
feat:     new functionality
fix:      bug fix
docs:     documentation
refactor: restructure without new functionality
chore:    config, dependencies
```

**NEVER commit:** `.env`, `.env.*`, `.keys/`, `node_modules/`, `vendor/`
**NEVER put real keys in `.env.example`** — use placeholder values like `CHANGE_ME` or `sk-ant-...EXAMPLE`

---

## 7. README.md Rule

> **MANDATORY:** Update `README.md` at EVERY session that modifies:

| Event | What to update |
|-------|---------------|
| New version | `**Current version:**` line + Changelog table |
| New functionality | API Endpoints section |
| New migration | Database section (migration count, tables) |
| New scheduler job | Scheduler Jobs table |
| New/removed endpoint | API Endpoints section |
| Modified structure | Project Structure section |
| New env variable | Environment Variables section |

```bash
# Quick check at end of session
grep "Current version" README.md
grep "Current version" CLAUDE.md
# Must match
```

---

## 8. Quick Deployment

**Flow:** push to `main` → GitHub Actions → DockerHub → SSH VPS → health check

```bash
# Health check after deploy
curl https://api.nuviax.app/health
# Expected: {"status":"ok","db":true,"redis":true}
```

**Full details:** `docs/deployment.md`

---

## 9. Critical Rules

```
NEVER expose in API:
❌ drift, chaos_index, continuity, weights, factors, penalties
❌ thresholds (0.25, 0.40, 0.60), formulas, score components

EXPOSE ONLY:
✅ Progress % (0-100)
✅ Grade (A+, A, B, C, D)
✅ Ceremony (tier, message, badge)
✅ Achievements (ID, name, icon)
```

**Admin 404:** admin panel returns 404 (not 403) for non-admins.
**Timing-safe:** `forgot-password` always returns 200.
**Graceful degradation:** AI and Email work without their respective keys.
**No real secrets in git:** `.env.example` must contain only placeholder values.

---

## 10. References

| Resource | Location |
|----------|---------|
| **User workflow + test scenarios** | **`docs/user-workflow.md`** |
| Repo structure + design system | `docs/project-structure.md` |
| Full DB schema + migrations | `docs/database-reference.md` |
| AI (Haiku) + Email (Resend) details | `docs/integrations.md` |
| VPS, Docker, CI/CD, secrets | `docs/deployment.md` |
| Bug history + stress test gaps | `docs/history-bugs-gaps.md` |
| Framework formulas (C1-C40) | `FORMULAS_QUICK_REFERENCE.md` |
| Development roadmap | `ROADMAP.md` |
| Detailed changelog | `CHANGES.md` |
| GitHub Secrets guide | `infra/GITHUB_SECRETS.md` |
| **Implementation prompts (session-ready)** | **`PROMPTS.md`** |
| Client action items | `CLIENT_TODO.md` |

---

## 11. User Workflow & Testing (MANDATORY)

> `docs/user-workflow.md` is the **source of truth** for all feature validation.
> A feature that is not covered by a passing test scenario does not exist.

---

### Before implementing any feature

1. Open `docs/user-workflow.md`
2. Identify the relevant flow section (Sections 2–6)
3. Identify the related test scenarios (Section 7, TS-xx)
4. Confirm the implementation will satisfy the inputs, DB changes, API response, and frontend behavior defined in that section

```
# Example — implementing SRM L2 intensity reduction:
# → Read Section 4 (SRM Flow)
# → Read TS-05
# → Verify: ConfirmSRML2 creates context adjustment + task count drops next day
```

---

### After implementing any feature

State explicitly which test scenarios should now pass:

```
# Example commit message or session close note:
# "SA-4 fix complete — TS-05 should pass: L2 confirm creates ENERGY_LOW
#  context adjustment; next-day task count reduced."
```

Do not close a session without naming the test scenarios the implementation covers.

---

### Hard rules

```
NEVER implement a feature without first reading its flow in user-workflow.md
NEVER mark a task complete if its TS-xx scenario would still fail
NEVER assume a feature works — identify the exact DB change and API response
NEVER ship a fix for SA-1 through SA-7 without running its mapped test scenario
```

---

### Quick lookup

| Task type | Where to look first |
|---|---|
| New feature | Section matching the flow (2–6) → nearest TS-xx |
| Bug fix | TS-xx that reproduces the bug → Critical Checkpoints (8) |
| Scheduler job | Section 6.1 (trajectories) or Section 5.1 (achievements) |
| SRM change | Section 4 → TS-04, TS-05, TS-06 |
| API response change | Section 8.1–8.2 (opaque API rules) → TS-12 |
| Auth/security change | Section 8.3–8.6 |

**Reference:** `docs/user-workflow.md` — Sections 9–10 map all Sprint 3.1 fixes (SA-1 through SA-7) to their test scenarios.

---

## 12. Prompt Optimization Rule (MANDATORY)

> Before starting any task from PROMPTS.md, Claude MUST:

### Step 1 — Read only what's needed
Based on the prompt's "Files to read (max 3)" list, read EXACTLY those files.
Do NOT explore additional files unless a file is missing (then ask the owner).

### Step 2 — Validate before coding
Answer these questions from the files read:
- Does the function/line referenced in the prompt actually exist?
- If a function is missing, create it as specified — do not skip.
- If a line number is wrong, find the correct location and proceed.

### Step 3 — One change at a time
Make changes in the exact order listed in the prompt.
After each change, grep to verify (as specified in the prompt's Verification block).
Do NOT proceed to the next change if grep shows unexpected results — report to owner.

### Step 4 — Close the session correctly
- State which TS-xx scenarios are now satisfied
- Update CLAUDE.md section 4 status
- Update docs/testing/test-plan.md gap status
- Commit with exact message format from prompt
- Do NOT add extra files to the commit unless the prompt specifies them

### Why this rule exists
Context windows are expensive. A session that reads 10 files instead of 3
uses 3× the tokens for the same output. Every file read beyond the prompt's
list must be justified by a missing reference — not curiosity or "just in case".

*Last updated: 2026-03-30 — v10.5.0 — user-workflow.md complete, Section 11 added*

*Last updated: 2026-04-02 — v10.5.1 — Prompt Optimization Rule, Section 12 added*
