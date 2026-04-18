# docs/testing/smoke-test-report.md — NuviaX MVP Smoke Test Report

> Versiune: 1.1.0  
> Data: 2026-04-18  
> Branch: claude/smoke-test-docs-jYYh7

---

## Rezultate verificări statice

| Check | Comandă | Rezultat |
|---|---|---|
| Secrete expuse | `grep -rn "sk-ant-\|re_[A-Za-z0-9]{20,}"` în docs + infra | ✅ ZERO |
| API opacity | `grep -rn "drift\|chaos_index\|weights\|threshold"` în handlers/ | ✅ ZERO |
| Build Go | `cd backend && go build ./...` | ✅ PASS |
| Unit tests engine | `go test ./internal/engine/... -v` | ✅ 12/12 PASS |

---

## Plan smoke test E2E (require: server live)

Rulează pe deployment real (`https://api.nuviax.app`) sau local cu Docker Compose (`infra/docker-compose.yml`).

```bash
BASE=https://api.nuviax.app/api/v1   # sau http://localhost:8080/api/v1

# 1. Register
curl -s -o /dev/null -w "%{http_code}" -X POST $BASE/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"smoke@test.com","password":"Smoke1234!","full_name":"Smoke Test"}'
# → 201 (user nou) sau 409 (deja există)

# 2. Login + token
TOKEN=$(curl -s -X POST $BASE/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"smoke@test.com","password":"Smoke1234!"}' | jq -r '.access_token')

# 3. AI validate GO
curl -s -H "Authorization: Bearer $TOKEN" -X POST $BASE/goals/analyze \
  -H "Content-Type: application/json" \
  -d '{"text":"Vreau să slăbesc"}'
# → {"needs_clarification": true, ...} (text vag → AI cere reformulare)

# 4. Create GO
curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" \
  -X POST $BASE/goals \
  -H "Content-Type: application/json" \
  -d '{"name":"Test MVP Goal","start_date":"2026-04-18","end_date":"2026-10-18","dominant_behavior_model":"INCREASE"}'
# → 201

# 5. Today
curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" $BASE/today
# → 200

# 6. Dashboard
curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" $BASE/dashboard
# → 200

# 7. API Opacity CRITIC — nu trebuie să apară câmpuri interne
curl -s -H "Authorization: Bearer $TOKEN" $BASE/goals | grep -i "drift\|chaos\|weight\|threshold"
# → ZERO rezultate

# 8. Admin guard (non-admin → 404, nu 403)
curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" $BASE/admin/stats
# → 404
```

---

## Criterii de acceptare

| Test | Criteriu |
|---|---|
| Register | 201 sau 409 |
| Login | access_token prezent în response |
| AI analyze | JSON cu câmpul `needs_clarification` |
| Create GO | 201 |
| Today | 200 |
| Dashboard | 200, câmpul `goals` prezent |
| API Opacity | ZERO câmpuri `drift/chaos/weight/threshold` în response |
| Admin guard | 404 pentru utilizator non-admin |

---

## Known constraints

- Smoke test E2E necesită PostgreSQL + Redis live (nu rulează în CI sandbox fără servicii).
- AI analyze: dacă `ANTHROPIC_API_KEY` lipsește, răspunsul vine din fallback rule-based (tot valid).
- Scheduler jobs (task generation, sprint close) sunt asincrone — nu se testează sincron în smoke.
