# docs/testing/smoke-test-report.md — NuviaX MVP Smoke Test Report

> Versiune: 1.2.0  
> Data: 2026-04-19  
> Branch: claude/prepare-test-suite-jheWT

---

## Rezultate verificări statice

| Check | Comandă | Rezultat |
|---|---|---|
| Secrete expuse | `grep -rn "sk-ant-\|re_[A-Za-z0-9]{20,}"` în docs + infra | ✅ ZERO |
| API opacity | `grep -rn "drift\|chaos_index\|weights\|threshold"` în handlers/ | ✅ ZERO |
| Build Go | `cd backend && go build ./...` | ✅ PASS |
| Unit tests backend | `go test ./internal/engine/... ./internal/ai/... ./internal/email/... ./internal/api/handlers/... ./pkg/crypto/...` | ✅ 71/71 PASS |
| Frontend type-check | `npx tsc --noEmit` în `frontend/app` | ✅ PASS |

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

# 3. AI validate GO (SMART check)
curl -s -H "Authorization: Bearer $TOKEN" -X POST $BASE/goals/analyze \
  -H "Content-Type: application/json" \
  -d '{"text":"Vreau să slăbesc"}'
# → {"needs_clarification": true, ...} (text vag → AI cere reformulare)

# 3b. AI suggest category + behavior model + directions (used by onboarding UI)
curl -s -H "Authorization: Bearer $TOKEN" -X POST $BASE/goals/suggest-category \
  -H "Content-Type: application/json" \
  -d '{"title":"Vreau să alerg 5km în 30 minute","description":""}'
# → {"category":"HEALTH","behavior_model":"INCREASE","directions":["..."], "source":"ai|fallback"}
# Onboarding UI propune user-ului să aleagă una dintre directions când sunt ≥ 2.

# 4. Create GO (dominant_behavior_model e OBLIGATORIU per engine.ValidateGO)
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

---

## Changelog 1.2.0 (2026-04-19)

Onboarding blocker rezolvat — flow-ul real al user-ului e acum acoperit end-to-end:

1. **AI suggest-category extins (C2+C9)**: `SuggestGOCategory` întoarce acum și `behavior_model` (din set canonic C2: CREATE/INCREASE/REDUCE/MAINTAIN/EVOLVE) și `directions[]` (variante formulate SMART).
2. **Handler `/goals/suggest-category`**: expune `behavior_model` + `directions` + `source` ("ai" sau "fallback"). Când `ANTHROPIC_API_KEY` lipsește, fallback-ul rule-based garantează un răspuns valid (niciodată vid).
3. **Frontend onboarding**: trimite `dominant_behavior_model` în `POST /goals` (câmpul era obligatoriu în `engine.ValidateGO` dar frontend-ul nu îl trimitea → onboarding bloca cu 400 silent). Dacă AI întoarce ≥ 2 directions, UI-ul afișează chips pentru alegere înainte de creare.
4. **Testare**: 12 → 71 teste unitare backend (engine, AI, email, handlers fallback, crypto). `test_api.sh` verifică acum suggest-category contract + API opacity + admin guard. `test-unit.yml` rulează întreaga suită + frontend type-check în CI.
