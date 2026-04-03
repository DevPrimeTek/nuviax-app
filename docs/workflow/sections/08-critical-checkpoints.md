## 8. Critical Checkpoints

### CURRENT SYSTEM (REALITY)

### TARGET SYSTEM (FRAMEWORK)

### 8.1 Server-Side Calculation Enforcement

**What must never break:** All score computation runs in Go engine only. No formula, weight, factor, or threshold may appear in any API response.

**How to verify:**
```bash
curl -H "Authorization: Bearer $TOKEN" https://api.nuviax.app/api/v1/goals/$GOAL_ID | \
  jq 'keys'
# Must NOT contain: drift, chaos_index, continuity, weights, factors,
#                   penalties, score_components, thresholds
```

**Allowed response keys:** `progress_pct`, `grade`, `grade_label`, `score` (opaque float 0–1), `actual_pct`, `expected_pct`, `delta`, `trend`, `tier`, `badge_type`.

**Failure looks like:** Any of the forbidden fields present in JSON response body. Immediate fix required — do not deploy.

---

### 8.2 Opaque API Response Validation

**What must never break:** Clients receive only grades (A+/A/B/C/D) and percentages. The numeric computation chain (C1–C40) is internal.

**How to verify:** Run TS-12 — inspect all goal, visualization, and SRM status responses. Use `jq 'to_entries[] | .key'` to enumerate all keys.

**Failure looks like:** `chaos_index: 0.42` or `weight_c7: 0.33` appearing in any API response. Even a logging endpoint must not expose these.

---

### 8.3 JWT Auth on All Protected Routes

**What must never break:** Every route except `/auth/login`, `/auth/register`, `/auth/forgot-password`, `/auth/reset-password`, and `/health` requires a valid JWT RS256 access token.

**How to verify:**
```bash
# Should return 401
curl https://api.nuviax.app/api/v1/goals
curl https://api.nuviax.app/api/v1/today
curl https://api.nuviax.app/api/v1/achievements
```

**Token behavior:** Access token expires in 15 minutes. Frontend proxy (`api/proxy/[...path]`) auto-refreshes using refresh token cookie. If refresh token expired → redirect to `/auth/login`.

**Failure looks like:** Any protected route returning `200` without `Authorization` header. Or `500` instead of `401` on missing token.

---

### 8.4 Admin 404 (Non-Admin Access)

**What must never break:** Admin panel returns `404` — not `403` — for non-admin users. The existence of the admin route must not be detectable.

**How to verify:**
```bash
# With a regular user token
curl -H "Authorization: Bearer $REGULAR_TOKEN" \
  https://api.nuviax.app/api/v1/admin/stats
# Must return 404, not 403 or 401
```

**Enforced by:** `middleware/admin.go` — checks `is_admin = TRUE` on `users` table; calls `notFound(c)` on failure (returns 404 body identical to other 404 responses).

**Failure looks like:** `403 Forbidden` (reveals route exists), or `401` (reveals route is protected). Any non-404 response is a disclosure failure.

---

### 8.5 Graceful Degradation (AI + Email Down)

**What must never break:** If Anthropic API is unreachable, onboarding continues — AI suggestion silently returns empty. If Resend is unreachable, registration succeeds — welcome email silently fails.

**How to verify (AI):**
```bash
# With ANTHROPIC_API_KEY unset or invalid
curl -X POST https://api.nuviax.app/api/v1/goals/suggest-category \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"goal_name":"Learn piano"}'
# Must return 200 with empty/null suggestion within 2 seconds
```

**How to verify (Email):**
- Set `RESEND_API_KEY=invalid` → `POST /auth/register` → must still return `201`

**Timeout enforcement:** `SuggestGOCategory()` has a 2-second hard context timeout. Response must arrive within that window regardless of Anthropic upstream latency.

**Failure looks like:** `500` error on registration when email fails. Or onboarding hanging >3 seconds when AI is down. Or `400/500` from suggest-category endpoint.

---

### 8.6 Timing-Safe Forgot Password

**What must never break:** `POST /auth/forgot-password` always returns `200` with a neutral message, regardless of whether the email exists in the DB. Response time must not differ between known and unknown emails.

**How to verify:**
```bash
# With known email
curl -X POST https://api.nuviax.app/api/v1/auth/forgot-password \
  -d '{"email":"real@user.com"}'
# → 200 {"message":"..."}

# With unknown email
curl -X POST https://api.nuviax.app/api/v1/auth/forgot-password \
  -d '{"email":"nobody@nowhere.com"}'
# → 200 {"message":"..."} — identical response
```

**Failure looks like:** `404` or `422` when email not found (user enumeration). Or measurably different response time between the two cases (timing side-channel).

---

