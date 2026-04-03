# CLAUDE.md — NuviaX Master Operating Rules (Source of Truth)

> Versiune: 11.1.0  
> Actualizat: 2026-04-03  
> **Regula #1:** orice prompt/sesiune începe cu analiza acestui fișier.

---

## 0) Start protocol (obligatoriu în fiecare sesiune)

1. Citește `CLAUDE.md` cap-coadă.
2. Confirmă versiunea + branch + task.
3. Citește **doar fișierele minim necesare** din indexul de mai jos.

Comenzi standard:

```bash
git status
git log --oneline -5
git branch --show-current
```

Template de start:

```text
Read CLAUDE.md v11.1.0. Branch: <branch>. Task: <task>. Files to inspect: <max 3 initially>.
```

---

## 1) Token/request optimization rules (Codex + Claude Code)

- Nu face scan global inutil.
- Pleacă din `CLAUDE.md` -> deschide doar fișierele relevante task-ului.
- Extinde contextul progresiv (max 3 fișiere inițial, apoi incremental).
- Pentru debugging: întâi verifici config+middleware+ruta, apoi handlers/DB, apoi UI.
- Nu marca "root cause" fără dovadă în cod.

---

## 2) File index (referințe oficiale)

## 2.1 Core governance docs
- `CLAUDE.md` — reguli sesiune + index
- `PLAN.md` — workstreams tehnice + release gating
- `ROADMAP.md` — milestones M1..M4
- `README.md` — quickstart + ops checks

## 2.2 Framework alignment docs
- `docs/framework/rev5_6/README.md` (index)
- `docs/framework/rev5_6/00-intro.md`
- `docs/framework/rev5_6/10-layer0.md` ... `60-level5.md`
- `docs/framework_100_percent_implementation_playbook.md`
- `docs/framework_workflow_deviations_stress_test.md`

## 2.3 Runtime backend (Go)
- `backend/internal/api/server.go` — routing map
- `backend/internal/api/middleware/jwt.go` — auth guard
- `backend/internal/api/middleware/admin.go` — admin guard
- `backend/internal/api/handlers/*.go` — API behavior
- `backend/internal/engine/*.go` — core scoring logic
- `backend/internal/scheduler/scheduler.go` — cron orchestration
- `backend/internal/db/queries.go` — DB access
- `backend/migrations/*.sql` — schema truth

## 2.4 Frontend app
- `frontend/app/middleware.ts` — route protection
- `frontend/app/app/admin/page.tsx` — admin UI
- `frontend/app/app/api/proxy/[...path]/route.ts` — API proxy with cookie auth
- `frontend/app/app/auth/login/page.tsx` — login flow

## 2.5 Test docs
- `docs/testing/test-plan.md` — master test plan
- `docs/testing/scenarios/critical.md`
- `docs/testing/scenarios/regression.md`
- `docs/workflow/README.md` + `docs/workflow/sections/*.md` (workflow structurat)

---

## 3) Admin panel runbook (DevOps quick fix)

## Simptom
User se loghează în aplicație, dar `/admin` nu se deschide.

## Cauze tipice
1. user nu are `is_admin=TRUE`;
2. login făcut cu username în loc de email;
3. sesiune/token vechi după promovare admin.

## Soluție standard (idempotent)

```bash
bash scripts/setup_admin.sh sbarbu_admin 'NuviaXAdmin#2026' 'Sbarbu Admin'
```

Acest script:
- convertește automat `sbarbu_admin` -> `sbarbu_admin@nuviax.app`
- creează contul dacă nu există
- setează `is_admin=TRUE`
- verifică login API
- afișează pașii finali de acces

După rulare:
1. logout,
2. login cu email-ul final,
3. refresh hard,
4. deschide `https://nuviax.app/admin`.

Dacă apare încă „Eroare internă” în admin:

```bash
bash scripts/apply_migrations.sh
```

Motiv tipic: migrații incomplete (ex: lipsește `completion_ceremonies` / view-uri admin).

---

## 4) Working rules for edits

- Orice schimbare la auth/admin trebuie să atingă și documentația relevantă.
- Pentru schimbări framework (scoring/SRM), actualizezi obligatoriu:
  - `PLAN.md`
  - `ROADMAP.md`
  - `docs/testing/test-plan.md`
- Nu lăsa statusuri "implemented" dacă nu sunt verificabile în cod.

---

## 5) Definition of Done

Task-ul este DONE doar dacă:
1. schimbarea este implementată,
2. root cause este demonstrat cu referințe de cod,
3. test/check minim este rulat,
4. docs principale sunt sincronizate,
5. commit + PR message sunt create.

---

## 6) Security invariants

Nu expune în API:
- metrici interne (weights/drift/chaos/formule)
- praguri interne sensibile

Admin routes:
- JWT valid + `is_admin=TRUE`
- non-admin -> răspuns 404 (fără leak funcționalitate)
