# CLAUDE.md — NuviaX Master Context (Source of Truth)

> Versiune: 1.0.0 (post-reset)  
> Actualizat: 2026-04-06  
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
| F0.1 — Cleanup fișiere moarte | ✅ | Fișiere v10.x eliminate |
| F1 — Schema DB pentru MVP | ✅ | 32 tabele în schema public |
| F2 — Auth CSS standardizat | ✅ | Pagini auth consistente |
| F3 — Core Engine | ⏳ | — |
| F4 — Scheduler + SRM | ⏳ | — |
| F5 — API Handlers | ⏳ | — |
| F6 — Frontend MVP | ⏳ | — |
| F7 — Smoke Test + Docs | ⏳ | — |

**DB activă:** schema `public`, 32 tabele, migrări 001–013 din repo.

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
Read CLAUDE.md v1.0.0. Branch: claude/<feature>. Task: <descriere>.
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

**Fișiere v10.x eliminate (F0.1 ✅):**
- `infra/init-db.sql`, `PLAN.md`, `PROMPTS.md`, `CHANGES.md` — eliminate
- `docs/DEMO_EXECUTION_PLAN.md`, `docs/framework_100_percent_implementation_playbook.md`, `docs/framework_workflow_deviations_stress_test.md` — eliminate

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

## 8) Index fișiere

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

---

## 9) Definition of Done per task

1. Implementare conform MVP_SCOPE (FULL sau SIMPLIFIED)
2. Zero câmpuri interne expuse în API
3. Test minim rulat
4. Docs actualizate (regula SYNC-ON-COMMIT)
5. Cleanup: zero fișiere/imports/logs moarte
6. Commit + push pe branch `claude/*`
