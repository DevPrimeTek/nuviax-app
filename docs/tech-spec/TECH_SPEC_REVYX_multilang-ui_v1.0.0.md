# TECH SPEC — REVYX Multi-Language UI
**Document:** TECH_SPEC_REVYX_multilang-ui_v1.0.0.md  
**Version:** 1.0.0  
**Phase:** S7 — Phase 4 Post-Launch  
**Status:** APPROVED  
**Date:** 2026-05-06  
**Authors:** Frontend Engineering · Product  
**Reviewers:** Architect · Security · DBA · QA · Compliance · Product · Audit Lead

---

## Impact Assessment ★

| Dimension | Impact | Notes |
|---|---|---|
| User Experience | HIGH | Unlocks Romanian rural + Ukrainian diaspora markets (S7-4) |
| Security | MEDIUM | Cookie handling + language preference DB column require review |
| Compliance | HIGH | Ukrainian market: UI language ≠ data residency — strings only, no PII |
| Performance | LOW | next-intl adds ~2–4 KB per locale bundle (tree-shaken) |
| Maintainability | MEDIUM | Translation key discipline required; stale keys = build warning |
| SEO | MEDIUM | `hreflang` tags required for multi-locale routes |

---

## 1. Library Selection

**Chosen: `next-intl` v3.x** (compatible with Next.js 14 App Router)

Rationale vs `react-i18next`:
- Native App Router support (server components, layouts, route segments)
- No provider boilerplate in server components
- Built-in `getTranslations()` async server API — no hydration mismatch
- ICU message format (plurals, gender, number formatting)

```bash
npm install next-intl
```

---

## 2. Locale Configuration

### 2.1 Supported Locales

| Locale | Language | Default? | Direction |
|---|---|---|---|
| `ro` | Română | ✅ YES | LTR |
| `ru` | Русский | NO | LTR |

> **Note on directionality:** Both Romanian and Russian are LTR scripts. RTL readiness
> (CSS `[dir="rtl"]` overrides, logical CSS properties `margin-inline-start` instead of
> `margin-left`) is implemented proactively to avoid refactor cost when Arabic/Hebrew
> locales are added in Phase 5.

### 2.2 next-intl Configuration

```typescript
// frontend/i18n.config.ts
import { defineRouting } from 'next-intl/routing'

export const routing = defineRouting({
  locales: ['ro', 'ru'],
  defaultLocale: 'ro',
  localeDetection: true,          // uses Accept-Language header
  localePrefix: 'as-needed',      // /ro/* omitted; /ru/* explicit
})
```

```typescript
// frontend/middleware.ts — extend existing middleware
import createMiddleware from 'next-intl/middleware'
import { routing } from './i18n.config'

export default createMiddleware(routing)

export const config = {
  matcher: ['/((?!api|_next|.*\\..*).*)'],
}
```

---

## 3. Namespace Split

Translations are split into 5 namespaces to enable per-route code splitting and
to isolate concerns for human translators.

| Namespace | File | Used In |
|---|---|---|
| `common` | `messages/{locale}/common.json` | Layout, nav, buttons, errors, dates |
| `auth` | `messages/{locale}/auth.json` | Login, register, forgot-password, reset |
| `match` | `messages/{locale}/match.json` | Match list, filters, match card, NBA banner |
| `deal` | `messages/{locale}/deal.json` | Deal saga, timeline, contract, status labels |
| `admin` | `messages/{locale}/admin.json` | Admin panel — tenant admin + platform admin |

```
frontend/messages/
  ro/
    common.json
    auth.json
    match.json
    deal.json
    admin.json
  ru/
    common.json
    auth.json
    match.json
    deal.json
    admin.json
```

### 3.1 Key Naming Convention

```
<component>.<element>[.<modifier>]
```

Examples:
```json
// common.json
{
  "nav.properties": "Proprietăți",
  "nav.matches": "Potriviri",
  "nav.deals": "Tranzacții",
  "button.save": "Salvează",
  "button.cancel": "Anulează",
  "error.generic": "A apărut o eroare. Încearcă din nou.",
  "pagination.page_of": "{page} din {total}"
}
```

**Rules:**
- Keys are in English, snake_case with dot hierarchy
- No dynamic keys (no string concatenation to build key names)
- ICU plurals use `{count, plural, one {# element} other {# elemente}}`
- Never hard-code locale strings in component source — always via `t('key')`

### 3.2 Missing Key Policy

Build-time: `next-intl` strict mode enabled → missing key = TypeScript error.

```typescript
// frontend/global.d.ts
import ro_common from './messages/ro/common.json'

type Messages = typeof ro_common // Romanian is the authoritative key set
declare global {
  interface IntlMessages extends Messages {}
}
```

---

## 4. Translation Workflow

```
Developer adds key to messages/ro/*.json (source of truth)
        │
        ▼
CI check: script/check-translation-keys.ts
  → compares ro/ keys vs ru/ keys
  → reports missing keys in ru/ as WARNING (non-blocking for alpha)
  → reports EXTRA keys in ru/ as ERROR (blocker — stale key)
        │
        ▼
Human translator fills ru/ keys
  (Crowdin project OR direct PR to messages/ru/)
        │
        ▼
AI pre-fill (optional, for speed):
  scripts/ai-prefill-translations.ts
  → calls Claude claude-haiku-4-5-20251001 with prompt:
    "Translate these Romanian UI strings to Russian.
     Context: B2B real-estate SaaS platform.
     Return JSON with same keys."
  → Output saved as ru/*.json.draft (NOT committed as final)
  → Human translator reviews draft → approves → renames to .json
        │
        ▼
PR review → merge → Vercel deploys
```

### 4.1 AI Pre-fill Script

```typescript
// scripts/ai-prefill-translations.ts
import Anthropic from '@anthropic-ai/sdk'
import fs from 'fs'
import path from 'path'

const client = new Anthropic()

async function prefillNamespace(namespace: string) {
  const roPath = path.join('messages', 'ro', `${namespace}.json`)
  const ruDraftPath = path.join('messages', 'ru', `${namespace}.json.draft`)

  const roStrings = JSON.parse(fs.readFileSync(roPath, 'utf-8'))

  const response = await client.messages.create({
    model: 'claude-haiku-4-5-20251001',
    max_tokens: 4096,
    messages: [{
      role: 'user',
      content: `Translate the following Romanian UI strings to Russian for a B2B real-estate SaaS platform.
Preserve ICU message format placeholders ({count}, {page}, etc.) exactly.
Return only valid JSON with the same keys.

${JSON.stringify(roStrings, null, 2)}`,
    }],
  })

  const draft = response.content[0].type === 'text' ? response.content[0].text : ''
  fs.writeFileSync(ruDraftPath, draft)
  console.log(`Drafted: ${ruDraftPath}`)
}

for (const ns of ['common', 'auth', 'match', 'deal', 'admin']) {
  await prefillNamespace(ns)
}
```

---

## 5. Language Detection Chain

```
Request arrives
    │
    ▼
1. URL prefix check: /ru/* → locale = 'ru'
    │ (no prefix or /ro/*)
    ▼
2. Cookie check: revyx_locale=ru|ro
    │ (no cookie)
    ▼
3. Accept-Language header parse
   (e.g. "ru-UA,ru;q=0.9,uk;q=0.8" → 'ru')
    │ (no match or unsupported)
    ▼
4. Default: 'ro'
```

**next-intl handles steps 1 and 3 automatically** via `localeDetection: true`.  
Steps 2 (cookie) and user preference DB sync are handled by custom middleware extension:

```typescript
// frontend/middleware.ts — language cookie read
export default async function middleware(request: NextRequest) {
  // 1. Check user preference cookie (set on locale switch)
  const cookieLocale = request.cookies.get('revyx_locale')?.value
  if (cookieLocale && ['ro', 'ru'].includes(cookieLocale)) {
    // Redirect to appropriate locale prefix path
    return intlMiddleware(request) // next-intl respects locale prefix
  }
  return intlMiddleware(request)
}
```

---

## 6. Language Selector Component

```typescript
// frontend/components/LanguageSelector.tsx
'use client'

import { useLocale } from 'next-intl'
import { usePathname, useRouter } from 'next/navigation'
import { startTransition } from 'react'

const LOCALES = [
  { code: 'ro', label: 'RO', flag: '🇷🇴' },
  { code: 'ru', label: 'RU', flag: '🇷🇺' },
]

export function LanguageSelector() {
  const locale = useLocale()
  const router = useRouter()
  const pathname = usePathname()

  function switchLocale(next: string) {
    // Set cookie for persistence across sessions (1 year)
    document.cookie = `revyx_locale=${next};path=/;max-age=31536000;SameSite=Lax`

    startTransition(() => {
      // next-intl: replace locale segment in current path
      const newPath = pathname.replace(`/${locale}`, `/${next}`)
      router.replace(newPath)

      // Sync to backend (fire-and-forget)
      fetch('/api/proxy/users/preference', {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ language: next }),
      }).catch(() => {}) // non-critical
    })
  }

  return (
    <div className="flex gap-1" role="navigation" aria-label="Language selector">
      {LOCALES.map(({ code, label, flag }) => (
        <button
          key={code}
          onClick={() => switchLocale(code)}
          className={`px-2 py-1 text-sm rounded ${
            locale === code ? 'bg-primary text-white' : 'text-muted hover:bg-surface'
          }`}
          aria-current={locale === code ? 'true' : undefined}
        >
          {flag} {label}
        </button>
      ))}
    </div>
  )
}
```

### 6.1 User Preference — DB Schema

```sql
-- migration 014 (new, S7)
ALTER TABLE users ADD COLUMN IF NOT EXISTS
  language VARCHAR(5) NOT NULL DEFAULT 'ro'
  CHECK (language IN ('ro', 'ru'));

COMMENT ON COLUMN users.language IS 'UI locale preference; ISO 639-1';
```

### 6.2 API Endpoint — Preference Sync

```go
// PATCH /api/v1/users/preference
// Auth: JWT required (any authenticated user)
// Body: { "language": "ro" | "ru" }
func (h *Handler) UpdateUserPreference(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)
    var body struct {
        Language string `json:"language" validate:"required,oneof=ro ru"`
    }
    if err := c.BodyParser(&body); err != nil {
        return fiber.ErrBadRequest
    }
    if err := h.validate.Struct(body); err != nil {
        return fiber.ErrBadRequest
    }
    return h.db.UpdateUserLanguage(c.Context(), userID, body.Language)
}
```

---

## 7. RTL Readiness (Proactive)

Although `ro` and `ru` are both LTR, all CSS uses **logical properties** to enable
future RTL locale addition without component refactors:

| Do NOT use | Use instead |
|---|---|
| `margin-left` | `margin-inline-start` |
| `padding-right` | `padding-inline-end` |
| `text-align: left` | `text-align: start` |
| `float: right` | `float: inline-end` |
| `border-left` | `border-inline-start` |

```css
/* globals.css — RTL root setup */
[dir="rtl"] {
  font-family: var(--font-rtl, var(--font-sans));
}
```

```typescript
// frontend/app/[locale]/layout.tsx
export default async function LocaleLayout({ children, params }: Props) {
  const { locale } = params
  const dir = ['ar', 'he', 'fa'].includes(locale) ? 'rtl' : 'ltr'

  return (
    <html lang={locale} dir={dir}>
      <body>{children}</body>
    </html>
  )
}
```

---

## 8. SEO — hreflang Tags

```typescript
// frontend/app/[locale]/layout.tsx — head metadata
export async function generateMetadata({ params }: Props) {
  return {
    alternates: {
      languages: {
        'ro': 'https://revyx.app',
        'ru': 'https://revyx.app/ru',
        'x-default': 'https://revyx.app',
      },
    },
  }
}
```

---

## 9. CI Checks

```yaml
# .github/workflows/i18n.yml
name: i18n key audit
on: [push, pull_request]
jobs:
  check-keys:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - run: npm ci
      - run: npx ts-node scripts/check-translation-keys.ts
        # Exits non-zero if: extra key in ru/ vs ro/, OR key count mismatch > 20%
```

---

## 10. Audit Checkpoint — S7-1 Multi-Language UI ★

**Architect:** `next-intl` v3 App Router integration is the correct choice for Next.js 14. Server-component-safe `getTranslations()` avoids hydration mismatches. Namespace split aligns with route-level code splitting. Logical CSS properties for RTL proofing is the right investment now rather than later. Language detection chain (URL → cookie → Accept-Language → default) is complete. ✅

**Security:** Cookie `revyx_locale` uses `SameSite=Lax` — correct. Language preference PATCH endpoint requires JWT auth — correct. AI pre-fill script uses ANTHROPIC_API_KEY from env (existing) — no new secret surface. Ukrainian users accessing from EU: language strings contain zero PII, no data sovereignty implication. Language column on `users` table is not PII. ✅

**DBA:** `ALTER TABLE users ADD COLUMN language VARCHAR(5) DEFAULT 'ro'` is backward-compatible (DEFAULT value, not NOT NULL without DEFAULT). Migration is idempotent (`IF NOT EXISTS`). Language preference PATCH must be covered by existing `users` audit trigger (no schema gap introduced). ✅

**QA:** Coverage required: (1) locale switch persists across page refresh via cookie, (2) user preference sync PATCH called on switch, (3) Accept-Language `ru-UA` header routes to `ru` locale, (4) missing translation key triggers TS error at build, (5) RTL CSS logical properties verified via Playwright snapshot with synthetic RTL locale. ✅

**Compliance:** Russian locale adds Russian-language users. GDPR applies if any Russian users are in EU (likely, given diaspora scope). No new legal basis required — existing privacy policy covers language preference as a functional data element. Ukrainian users: UI strings in Ukrainian/Russian are in scope for S7-4 (that spec governs data residency). ✅

**Product:** Language selector placement: top-right nav bar, visible on all authenticated and public pages. Default `ro` is correct for primary market. AI pre-fill draft workflow (not auto-committed) gives translators a starting point without shipping untested machine translations. ✅

**Audit Lead:** **No hard blockers.** Items to track:
- [ ] Migration 014 run in production before deploy
- [ ] All `ru/` namespaces reviewed by native Russian speaker before GA (AI draft ≠ final)
- [ ] `hreflang` tags verified in production via Google Search Console

---

*End of TECH_SPEC_REVYX_multilang-ui_v1.0.0.md*
