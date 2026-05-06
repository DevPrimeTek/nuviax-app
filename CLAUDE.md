# CLAUDE.md — NuviaX Master Context (Source of Truth)

> Versiune: 1.6.0  
> Actualizat: 2026-05-06  
> **Regula #1:** orice sesiune Claude Code începe cu citirea acestui fișier.

---

## 0) Context proiect

NuviaX — platformă SaaS de goal management, construită pe **NuviaX Growth Framework Rev 5.6** (40 componente, C1–C40).

Stack: Go (Fiber v2) + PostgreSQL + Redis + Next.js 14 (TypeScript) + Claude Haiku 4.5 (AI) + Resend (email).

Proiectul a trecut printr-un **MVP Reset**. Engine-ul vechi (~30% conformitate) a fost eliminat. Reconstrucția se face pe faze F0–F7.

---

## 1) Stare curentă

| Fază | Status | Ce a produs |
|---|---|---|
| F0 — Reset engine + schema | ✅ | Cod vechi eliminat |
| F1 — Schema DB pentru MVP | ✅ | 32 tabele în schema public |
| F2 — Auth CSS standardizat | ✅ | Pagini auth consistente |
| F3 — Core Engine | ✅ | engine.go, srm.go, growth.go, helpers.go, 34 unit tests |
| F4 — Scheduler + SRM | ✅ | scheduler.go rewrite — 12 jobs, AI+Email integrate |
| F5 — API Handlers | ✅ | Auth 8 + Goals/Today/Dashboard 11 (F5a) + SRM/Achievements/Profile/Admin 11 endpoints (F5b) |
| F6 — Frontend MVP | ✅ | Build complet, zero erori TypeScript. Toate paginile și componentele funcționale. |
| F7 — Smoke Test + Docs | ✅ | Build PASS, 71 unit tests, smoke plan + docs v1.2.0 |
| F7.1 — Onboarding unblock | ✅ | AI `/goals/suggest-category` întoarce BM + directions; frontend trimite `dominant_behavior_model` la POST /goals |
| F7.2 — Onboarding SMART parsing | ✅ | Nou endpoint `POST /goals/parse`: AI generează 3 variante SMART clickabile; eliminat pasul `verify` cu input text; flux nou: input → parsing → suggestions → analyzing |
| F7.3 — Admin panel separat | ✅ | `/admin/login` standalone (propriu flux auth, nu trece prin login-ul app), `/api/admin/login` verifică `is_admin` înainte de setare cookies, middleware redirect `/admin/*` → `/admin/login`, bootstrap admin auto-promote/create prin `ADMIN_BOOTSTRAP_EMAIL`/`ADMIN_BOOTSTRAP_PASSWORD` |

**DB activă:** schema `public`, 32 tabele, migrări 001–013 din repo.

**Stare (2026-04-19):**
- F5a rezolvat: Goals/Today/Dashboard + AI client integrat în server.go ✅
- F5b rezolvat: SRM/Achievements/Profile/Admin handlers implementate ✅
- F5 complet ✅ — Frontend poate integra toate endpoint-urile
- F6 complet ✅ — Build success, zero erori TypeScript, audit endpoint-uri OK
- F7 complet ✅ — Build PASS, 71 unit tests, API opacity CLEAN, docs → v1.2.0
- F7.1 complet ✅ — Onboarding workflow E2E: AI → SMART check → categorie + BM + directions → user alege variantă → POST /goals cu BM → GO creat
- F7.2 complet ✅ — Onboarding SMART parsing: `POST /goals/parse` (AI `ParseAndSuggestGO`) → 3 variante SMART clickabile per GO → selecție → creare automată cu categorie + BM
- F7.3 complet ✅ — Admin panel complet separat: `/admin/login` (pagină dedicată, fără layout app), middleware propriu pentru `/admin/*`, `/api/admin/login` validează `is_admin` ÎNAINTE de setarea cookies, bootstrap admin idempotent via env vars

**Stare REVYX (2026-05-06) — S6 Phase 3 Production Hardening:**
- S6 Tech Specs complete ✅ — pgvector production, multi-tenant AI isolation, observability stack, pre-launch hardening
- S6 Workflow complete ✅ — Incident response cu SEV1-4 matrix, GDPR Art. 33-34 SLA, post-mortem template
- S6 Legal complete ✅ — TIA OpenAI Schrems II (draft, pending DPO sign-off)
- Phase 0 Security checklist actualizat ✅ — JWT RS256, RBAC exhaustiv, AUDIT_LOG completeness, BYPASSRLS CI check

**Stare REVYX (2026-05-06) — S7 Phase 4 Post-Launch:**
- S7-1 Multi-Language UI ✅ — next-intl, RO+RU, namespace split, AI pre-fill workflow, RTL-ready
- S7-2 ML Pricing Phase 3 ✅ — LightGBM, feature store, drift PSI, MLflow versioning, A/B 20/80, fallback rule-based
- S7-3 Churn Prediction ✅ — LightGBM classifier, AUC-ROC ≥0.78, NBA re-engagement, CS escalation, privacy-safe features
- S7-4 Market Expansion RO+UA ✅ — ANCPI cadastral, SIRUTA geocoding, ECB rates, UA translation pipeline, Schrems II confirmed
- S7-5 Partnerships API ✅ — imobiliare.ro/storia.ro feed import, HMAC webhook, Redis rate limit, SHA256 dedup, outbound notify
- S7-6 Billing Metering Operational ✅ — Stripe Meters API, usage metering, grace period FSM, invoice retention 10y, margin report

---

## 2) MVP Scope — Ce implementăm

**29 componente (14 FULL + 15 SIMPLIFIED). 11 componente POST_MVP.**

### FULL (implementare completă)
C1 Structural Supremacy, C2 Behavior Model, C3 Max 3 GO, C4 365-Day Max, C5 30-Day Sprint, C6 Normalization, C12 Future Vault, C14 GO Validation, C19 Sprint Structuring, C20 Sprint Target, C21 80% Rule, C24 Progress Computation, C25 Execution Variance, C37 Sprint Score

### SIMPLIFIED (logică de bază, fără edge cases avansate)
C7 Priority Weight, C8 Priority Balance, C9 Semantic Parsing, C10 BM Classification, C11 Relevance Scoring, C13 Relevance Thresholds, C22 Milestone Structuring, C23 Daily Stack, C26 Drift Engine, C27 Stagnation Detection, C28 Chaos Index, C30 Consistency Tracking, C32 Adaptive Context (doar Pause), C33 SRM (L1/L2/L3 fără timeout protocol), C38 GORI (fără variance penalty)

### POST_MVP (nu se implementează acum)
C15 Strategic Feasibility, C16 Capacity Calibration, C17 Deep Work Estimation, C18 Annual Recalibration, C29 Focus Rotation, C31 Behavioral Patterns, C34 Weighted Suspension, C35 Core Stabilization, C36 Reactivation Protocol, C39 Engagement Signal, C40 Sprint Reflection Gate

**Matrice completă cu justificări:** `MVP_SCOPE.md`

---

## 3) Start protocol sesiune

```text
Read CLAUDE.md v1.0.1. Branch: claude/<feature>. Task: <descriere>.
Files to read: <max 3>.
```

```bash
git status && git branch --show-current && git log --oneline -3
```

---

## 4) Reguli de lucru

**Model selection (regulă de bază — selectează modelul optim pentru sarcină):**

| Sarcină | Model | De ce |
|---|---|---|
| Implementare cod, bugfix, refactoring | **Sonnet** | Rapid, ieftin, suficient pentru cod structurat |
| Creare/modificare fișiere, migrări, config | **Sonnet** | Task-uri directe cu input/output clar |
| Teste unitare, integrare | **Sonnet** | Pattern repetitiv |
| Decizie arhitecturală nouă, design sistem | **Opus** | Necesită raționament profund |
| Debugging complex, root cause neclar | **Opus** | Necesită analiză multi-fișier |
| Review cod cu implicații de securitate | **Opus** | Necesită atenție la detalii subtile |

**Regulă implicită: Sonnet. Opus doar când task-ul necesită raționament, nu implementare.**

**Optimizare tokenuri și prevenire timeout:**
- Citește max 3 fișiere per task, cu path exact
- NU explora global structura — e documentată mai jos
- NU citi: `node_modules/`, `vendor/`, `.next/`, `build/`, `dist/`, `go.sum`
- Un task per sesiune → commit → sesiune nouă
- Dacă nu știi path-ul, întreabă
- Sesiuni max 45–60 min — dacă task-ul e mai mare, split în sub-sesiuni
- Prompturile din `PROMPTS_MVP.md` sunt deja split pentru a evita timeout
- Dacă o sesiune se apropie de limită: commit WIP, notează unde ai rămas, continuă în sesiune nouă

**Branch și commit:**
- Branch: `claude/<feature-name>` — niciodată direct pe `main`
- `git add <fișiere specifice>` — niciodată `git add .`
- Convenții: `feat:` / `fix:` / `docs:` / `refactor:` / `chore:`

**NU comite niciodată:** `.env`, `.env.*`, `.keys/`, `*.pem`, `*.key`, `node_modules/`, `vendor/`

**REGULĂ STRICTĂ — Variabile de mediu și funcționalități necunoscute:**
- `ANTHROPIC_API_KEY`, `RESEND_API_KEY` și toate celelalte env vars **există pe server și în GitHub Secrets** chiar dacă nu sunt vizibile local în sesiunea curentă.
- **NICIODATĂ nu dezactiva, nu comenta și nu elimini cod** care depinde de o variabilă de mediu sau o resursă pe care nu o găsești local.
- **NICIODATĂ nu faci presupuneri** despre ce există sau nu pe server/în config.
- Dacă ceva lipsește, este neclar, sau nu înțelegi contextul → **ÎNTREABĂ mai întâi**. Nu improviza, nu simplifica, nu dezactiva.

---

## 5) Regula CLEAN PROJECT — OBLIGATORIE

Proiectul trebuie să fie curat în permanență. La fiecare sesiune:

**Înainte de commit, verifică:**
- Niciun fișier nefolosit adăugat
- Niciun import mort în fișierele modificate
- Niciun `console.log` / `fmt.Println` de debug lăsat
- Niciun comentariu TODO fără referință la fază (ex: `// TODO(F8): implement C39`)

**La orice fază completată (F3, F4, etc.):**
- Șterge fișiere care nu mai sunt relevante
- Șterge funcții/metode moarte din fișierele atinse
- Verifică `docs/` — șterge documente care descriu comportament vechi eliminat

**Fișiere deja identificate pentru cleanup (F0.1 ✅ complet — toate eliminate):**
- Toate fișierele v10.x au fost eliminate în F0.1 (2026-04-06)
- Nu există fișiere moarte cunoscute la momentul 2026-04-18

**Promptul de cleanup F0.1 este în `PROMPTS_MVP.md`.**

---

## 6) PRE-COMMIT CHECKLIST — BLOCKER (nu poți comite fără asta)

**FIECARE sesiune, ÎNAINTE de `git commit`, execută TOATE aceste comenzi. Nu sări niciuna.**

### Pas 1 — Securitate: verifică că nu expui secrete
```bash
grep -rn "sk-ant-\|re_[A-Za-z0-9]\{20,\}\|PRIVATE.*KEY.*=.*[A-Za-z0-9/+]" \
  README.md CLAUDE.md ROADMAP.md infra/.env.example docs/ 2>/dev/null
# ZERO rezultate. Dacă găsește ceva → ELIMINĂ IMEDIAT înainte de commit.
```

### Pas 2 — Securitate engine: verifică API opacity
```bash
grep -rn "drift\|chaos_index\|weights\|threshold" backend/internal/api/handlers/ 2>/dev/null
# ZERO rezultate în handlere. Aceste valori sunt INTERNE.
```

### Pas 2b — REVYX S6: verificări suplimentare (Phase 2/3 confirmate) ★
```bash
# Verifică că niciun query pe tabele tenant-scoped nu lipsește tenant_id filter
# (pattern: SELECT/INSERT/UPDATE/DELETE fără WHERE tenant_id — CI sqlfluff rule REVYX001)

# Verifică că niciun rol applicație nu are BYPASSRLS
# (rulat și ca CI step — scripts/check_bypassrls.sh)

# Verifică că PII fields (email, phone, cnp) nu apar în log output
grep -rn "zap.String(\"email\"\|zap.String(\"phone\"\|zap.String(\"cnp\"" backend/ 2>/dev/null
# ZERO rezultate — PII se loghează DOAR prin PIIRedactor (redactat)
```

**Items confirmate de Phase 2/3 (S5/S6):**
- [x] JWT RS256 cu rotație dual-key (zero-downtime) — spec: `TECH_SPEC_REVYX_pre-launch-hardening_v1.0.0.md §1.1`
- [x] RBAC matrix exhaustiv (7 roluri × 7 resurse) — spec: `§1.2`
- [x] AUDIT_LOG completeness: toate tabelele write-sensitive acoperite de triggers — spec: `§1.3`
- [x] Webhook HMAC-SHA256 constant-time verify — spec: `§1.4`
- [x] Cross-tenant query auditing + BYPASSRLS CI check — spec: `TECH_SPEC_REVYX_multitenant-ai-isolation_v1.0.0.md §7`
- [x] PII redaction în logs (PIIRedactor registry) — spec: `TECH_SPEC_REVYX_observability-stack_v1.0.0.md §3`
- [x] KMS envelope encryption per tenant — spec: `multitenant §6`
- [x] GDPR Art. 33-34 breach notification SLA (72h CNPDCP) — workflow: `WORKFLOW_REVYX_incident-response_v1.0.0.md §6`
- [x] TIA OpenAI Schrems II — spec: `docs/legal/TIA_OPENAI_v1.0.0.md` (pending DPO sign-off)

### Pas 3 — Actualizează CLAUDE.md
- Dacă o fază s-a completat → marchează ✅ în secțiunea 1
- Dacă fișiere au fost șterse/adăugate → actualizează secțiunea 5 (cleanup) și secțiunea 8 (index)

### Pas 4 — Actualizează ROADMAP.md
- Dacă o fază s-a completat → marchează ✅ + adaugă data
- Dacă scope s-a schimbat → actualizează tabelul componentelor

### Pas 5 — Actualizează README.md
- Dacă endpoints noi → adaugă în lista endpoints
- Dacă env vars noi → adaugă în tabel
- Dacă scheduler jobs noi → adaugă în secțiune
- Dacă versiune se schimbă → actualizează header
- **Verifică că README nu conține chei API, parole, IP-uri, paths sensibile**

### Pas 6 — Actualizează docs/ afectate
| Ce ai modificat | Ce fișier docs actualizezi |
|---|---|
| Engine (scoring/formule) | `FORMULAS_QUICK_REFERENCE.md` |
| Schema DB (migrare nouă) | `docs/database-reference.md` |
| Integrări (AI/Email) | `docs/integrations.md` |
| Infrastructură | `docs/deployment.md` |

### Pas 7 — Verificare finală versiuni
```bash
grep "Versiune\|versiune\|Version" CLAUDE.md README.md ROADMAP.md
# Toate trei trebuie să arate aceeași versiune
```

### Pas 8 — Cleanup fișiere modificate
```bash
# Verifică imports/requires moarte în fișierele modificate
# Verifică console.log / fmt.Println de debug
# Verifică comentarii TODO fără referință la fază
```

**Dacă oricare din pașii 1–7 eșuează → NU COMITE. Fix întâi, apoi commit.**

**Această regulă este NON-NEGOCIABILĂ. Un commit fără docs actualizate = commit invalid care va fi revertat.**

---

## 7) Securitate engine — INVARIANT

```
NICIODATĂ nu expune în API:
❌ drift, chaos_index, weights, factors, thresholds, formule

EXPUNE DOAR:
✅ Progress % | Grade (A+/A/B/C/D) | Ceremony tier | Achievement ID
```

Admin routes: JWT valid + `is_admin=TRUE` → non-admin primește 404 (nu 403).

---

## 8) Integrări existente — NU RESCRIE, INTEGREAZĂ

Codul pentru AI și Email **deja există** și funcționează. La rebuild (F3–F7), aceste fișiere trebuie **păstrate și integrate**, nu rescrise.

### Claude Haiku 4.5 — `backend/internal/ai/ai.go`
- **Model:** `claude-haiku-4-5-20251001`
- **Client:** HTTP direct (stdlib, fără SDK), timeout 12s
- **Env var:** `ANTHROPIC_API_KEY` (deja setat pe server)
- **Graceful degradation:** `ai.IsAvailable()` → dacă false, fallback pe rule-based

**Funcții existente (NU le rescrie):**

| Funcție | Ce face | Unde se apelează | Fallback |
|---|---|---|---|
| `GenerateTaskTexts(ctx, goalName, checkpoint, sprint, count)` | Generează 1-3 task-uri zilnice contextuale | Scheduler `jobGenerateDailyTasks` | Template-uri statice |
| `AnalyzeGO(ctx, goalText)` | Clasificare SMART + BM detection | Handler `POST /goals/analyze` | Rule-based (vagueTerms, measurableKeywords) |
| `SuggestGOCategory(ctx, title, desc)` | Sugerează categorie GO | Handler `POST /goals/suggest-category` | Empty response |
| `ParseAndSuggestGO(ctx, rawText)` | Parsează GO brut și generează 3 variante SMART cu categorie + BM (C9, C10) | Handler `POST /goals/parse` | Returnează textul original cu fallback category/BM |

**Flux AI în creare GO (F7.2 — curent):**
```
User introduce text GO (brut, poate fi vag) → POST /goals/parse
  → ai.ParseAndSuggestGO() cu 10s timeout
  → Returnează 3 variante SMART cu categorie + BM
  → Frontend arată suggestions clickabile (NU input text)
  → User alege o variantă → POST /goals cu behavior_model + domain detectate automat
  → La creare sprint: checkpoints generate
  → Scheduler 00:01 UTC: ai.GenerateTaskTexts() generează sarcini zilnice
```

**Flux AI legacy (endpoints păstrate, nu mai folosite în onboarding):**
```
POST /goals/analyze → ai.AnalyzeGO() → needsClarification + question + hint
POST /goals/suggest-category → ai.SuggestGOCategory() → category + BM + directions
```

### Resend.com — `backend/internal/email/email.go`
- **Client:** HTTP direct (stdlib, fără SDK), timeout 10s
- **Env vars:** `RESEND_API_KEY` + `EMAIL_FROM=noreply@nuviax.app` (ambele setate pe server)
- **Graceful degradation:** `email.IsAvailable()` → dacă false, log + return nil

**Funcții existente (NU le rescrie):**

| Funcție | Ce face | Unde se apelează |
|---|---|---|
| `SendWelcome(ctx, to, name)` | Welcome email la înregistrare | Handler `Register` → goroutine fire-and-forget |
| `SendPasswordReset(ctx, to, resetLink)` | Link reset parolă (1h TTL) | Handler `ForgotPassword` → goroutine |
| `SendSprintComplete(ctx, to, name, goal, grade, sprint)` | Notificare sprint completat | Scheduler `jobCloseExpiredSprints` |

**Flux email forgot-password (deja implementat):**
```
POST /auth/forgot-password
  → mereu returnează 200 (timing-safe, previne enumeration)
  → generează token crypto random 32 bytes → SHA256 hash
  → INSERT password_reset_tokens (migration 009, TTL 1h, single-use)
  → email.SendPasswordReset() cu link https://nuviax.app/auth/reset-password?token=X
  → goroutine (nu blochează response)

POST /auth/reset-password
  → validează token (hash match, nu expirat, nu folosit)
  → UPDATE users SET password_hash
  → marchează token used_at
```

### Regulă: PRESERVE, DON'T REWRITE
La orice sesiune F3–F7: citește `ai.go` și `email.go` ÎNAINTE de a modifica handlers/scheduler. **Integrează funcțiile existente, nu le rescrie.** Dacă o funcție lipsește sau e incompletă, adaugă — nu înlocui.

---

## 9) Index fișiere

**Governance:**
- `CLAUDE.md` — acest fișier
- `ROADMAP.md` — faze F0–F7 + post-MVP
- `MVP_SCOPE.md` — matrice C1–C40 cu justificări
- `PROMPTS_MVP.md` — prompturi pentru Claude Code (F0.1, F3–F7)

**Framework (citește doar la nevoie):**
- `docs/framework/rev5_6/00-intro.md` — principii + changelog
- `docs/framework/rev5_6/10-layer0.md` — C1–C8
- `docs/framework/rev5_6/20-level1.md` — C9–C18
- `docs/framework/rev5_6/30-level2.md` — C19–C25
- `docs/framework/rev5_6/40-level3.md` — C26–C31
- `docs/framework/rev5_6/50-level4.md` — C32–C36
- `docs/framework/rev5_6/60-level5.md` — C37–C40
- `FORMULAS_QUICK_REFERENCE.md` — formule rapide

**Backend (Go):**
- `backend/internal/api/server.go` — routing
- `backend/internal/api/middleware/*.go` — JWT, admin
- `backend/internal/api/handlers/*.go` — handlere
- `backend/internal/engine/*.go` — scoring (F3)
- `backend/internal/scheduler/scheduler.go` — cron (F4)
- `backend/internal/db/db.go` + `queries.go` — DB
- `backend/migrations/*.sql` — schema

**Frontend:**
- `frontend/app/middleware.ts` — route protection
- `frontend/app/app/auth/*/page.tsx` — auth (F2 ✅)
- `frontend/app/app/api/proxy/[...path]/route.ts` — API proxy
- `frontend/app/styles/globals.css` — design system

**Infra:**
- `infra/docker-compose.yml` — containere
- `infra/.env.example` — variabile env
- `infra/verify-deployment.sh` — verificare post-deploy
- `scripts/setup_admin.sh` — bootstrap admin

**REVYX — Tech Specs (S6):**
- `docs/tech-spec/TECH_SPEC_REVYX_pgvector-production_v1.0.0.md` — HNSW tuning, reindex, quantization, fail-back
- `docs/tech-spec/TECH_SPEC_REVYX_multitenant-ai-isolation_v1.0.0.md` — per-tenant data sovereignty, KMS, cross-tenant audit
- `docs/tech-spec/TECH_SPEC_REVYX_observability-stack_v1.0.0.md` — OTel, logs PII redact, metrics catalog, SLOs, Grafana
- `docs/tech-spec/TECH_SPEC_REVYX_pre-launch-hardening_v1.0.0.md` — JWT RS256, RBAC, pen-test, load test, DPIA, go/no-go gate
- `docs/workflow/WORKFLOW_REVYX_incident-response_v1.0.0.md` — SEV matrix, on-call, war-room, GDPR breach notification
- `docs/legal/TIA_OPENAI_v1.0.0.md` — Transfer Impact Assessment Schrems II (draft, pending DPO sign-off)
- `docs/observability/dashboards/` — Grafana dashboard JSON (pending SRE delivery)

**REVYX — Tech Specs (S7):**
- `docs/tech-spec/TECH_SPEC_REVYX_multilang-ui_v1.0.0.md` — next-intl i18n, RO+RU, namespace split, AI pre-fill, RTL-ready
- `docs/tech-spec/TECH_SPEC_REVYX_ml-pricing-phase3_v1.0.0.md` — LightGBM pricing, feature store, PSI drift, MLflow, A/B test
- `docs/tech-spec/TECH_SPEC_REVYX_churn-prediction_v1.0.0.md` — B2B churn model, risk tiers, NBA re-engagement, CS escalation
- `docs/tech-spec/TECH_SPEC_REVYX_market-expansion-ro-ua_v1.0.0.md` — ANCPI, SIRUTA, ECB rates, UA translation, Schrems II
- `docs/tech-spec/TECH_SPEC_REVYX_partnerships-api_v1.0.0.md` — partner feed import, rate limiting, dedup, outbound webhooks
- `docs/tech-spec/TECH_SPEC_REVYX_billing-metering-operational_v1.0.0.md` — Stripe Meters, grace FSM, invoice retention, margin report

---

## 10) Definition of Done per task

1. Implementare conform MVP_SCOPE (FULL sau SIMPLIFIED)
2. Zero câmpuri interne expuse în API
3. Test minim rulat
4. Docs actualizate (regula SYNC-ON-COMMIT)
5. Cleanup: zero fișiere/imports/logs moarte
6. Commit + push pe branch `claude/*`
