# Bug Report — Onboarding Workflow Blocked (2026-04-19)

> Versiune: 1.0  
> Branch fix: `claude/prepare-test-suite-jheWT`  
> Gravitate: P0 — utilizatorul nu putea trece de onboarding după register + email verification.

---

## 1. Simptome raportate

User raportează că după autentificare:
- Register + email verification funcționează.
- Onboardingul NU este integrat cu AI.
- GO-ul introdus nu e analizat, nu primește variante SMART, nu se salvează.
- Flow-ul definit (GO → AI check → SMART variants → choice) nu e respectat.

## 2. Root cause

Trei defecte independente care se încurcă:

### D1 — AI nu întorcea Behavior Model
`ai.SuggestGOCategory` întorcea doar `category` + `confidence` + `reasoning`. Pentru `POST /goals`, engine-ul (`engine.ValidateGO`) impune un `DominantBehaviorModel` din setul canonic C2 (`CREATE|INCREASE|REDUCE|MAINTAIN|EVOLVE`) — altfel returnează eroare de validare.

### D2 — Handler nu expunea BM + directions
`/goals/suggest-category` întorcea frontendul doar `category`. Frontendul nu avea cum să afle BM sau variantele de formulare SMART — deci nu putea trimite `dominant_behavior_model` la creare.

### D3 — Frontend nu trimitea `dominant_behavior_model`
`POST /goals` era trimis fără câmpul obligatoriu → backend răspundea **400**, frontend înghițea eroarea silent → user rămânea blocat pe ecranul de creare, fără feedback.

Rezultat combinat: onboardingul *arăta* că trimite date, dar nu se salva niciodată un GO. AI-ul nu era integrat în flux.

## 3. Fix aplicat

### 3.1 `backend/internal/ai/ai.go`
- `SuggestionResult` extins cu `BehaviorModel string` + `Directions []string` (JSON tags `behavior_model`, `directions`).
- Prompt-ul rescris cere categorie + BM canonic + 1-3 directions formulate SMART (Romanian).
- Validare post-parse: BM neconform e ignorat (cade pe fallback); directions sunt capate la 3.

### 3.2 `backend/internal/api/handlers/goals.go`
- `/goals/suggest-category` întoarce acum `category`, `confidence`, `behavior_model`, `directions[]`, `source` (`"ai"` sau `"fallback"`).
- Fallback rule-based (`fallbackCategory`, `fallbackBehaviorModel`, `fallbackDirections`) când `ANTHROPIC_API_KEY` lipsește sau AI nu răspunde. Garantează răspuns valid → onboardingul nu mai depinde strict de AI.

### 3.3 `frontend/app/app/onboarding/page.tsx`
- State nou: `suggestedBMs[]`, `suggestedDirections[][]`, `chosenDirections`.
- Ecran nou `direction`: dacă AI întoarce ≥ 2 directions, user alege prin chips înainte de creare.
- `POST /goals` trimite `dominant_behavior_model: <BM-ul propus sau "INCREASE" default>`.
- Erorile de creare apar explicit la user (`createError`).

## 4. Verificare

| Layer | Check | Rezultat |
|---|---|---|
| Engine | 34 unit tests (C1–C38 coverage) | ✅ PASS |
| AI | 11 unit tests: IsAvailable, New, parseLines, validBehaviorModels, SuggestionResult JSON contract | ✅ PASS |
| Handlers | 5 unit tests: fallbackCategory/BM/Directions keyword matching | ✅ PASS |
| Email | 9 unit tests: template rendering + IsAvailable | ✅ PASS |
| Crypto | 10 unit tests: bcrypt, AES-GCM, SHA256, random | ✅ PASS |
| **Total** | `go test ./internal/engine/... ./internal/ai/... ./internal/email/... ./internal/api/handlers/... ./pkg/crypto/...` | **✅ 71/71** |
| Frontend | `npx tsc --noEmit` | ✅ PASS |
| CI | `.github/workflows/test-unit.yml` rulează toată suita + API opacity guard + frontend type-check | ✅ configurat |
| Smoke | `backend/scripts/test_api.sh` verifică `needs_clarification`, `behavior_model`, `directions`, API opacity, admin 404 | ✅ extins |

## 5. Workflow final (respectă specificația user)

```
1. User se autentifică + verifică email
2. /onboarding → introduce text GO
3. POST /goals/analyze  →  AI verifică SMART
   ├─ needs_clarification: true  →  UI cere reformulare
   └─ needs_clarification: false →  pasul 4
4. POST /goals/suggest-category → AI întoarce
   { category, behavior_model, directions[] }
5. Dacă directions.length ≥ 2 → UI afișează chips → user alege varianta
6. POST /goals { name, description, start_date, end_date, dominant_behavior_model }
   → 201 Created → onboarding → dashboard
```

Dacă `ANTHROPIC_API_KEY` lipsește: fallback rule-based livrează un BM + directions valide → flow-ul rămâne operațional (graceful degradation, conform CLAUDE.md §4).

## 6. Non-regresii verificate

- API opacity: niciun câmp `drift|chaos_index|weights|threshold` nu apare în `/goals` response (guard grep în CI).
- Admin guard: non-admin primește 404 (nu 403) la `/admin/*` — verificat în `test_api.sh`.
- Engine-ul nu a fost modificat — doar AI, handler, frontend și teste.
