# TEAM-ROSTER — NuviaX Fix Phase F8

> **Versiune:** 1.0.0
> **Data:** 2026-05-10
> **Owner:** Solution Architect

---

## Filosofie

NuviaX folosește un model de echipă „virtuală" cu Claude Code: fiecare rol se rulează ca o sesiune dedicată cu prompt specializat. Nu există rolling permanent — un specialist există atât timp cât sesiunea sa rulează. Tracking-ul cross-sesiune se face prin documentele governance (CLAUDE.md, ROADMAP.md, BACKLOG.md).

**Principiu cheie:** roluri **separate** pentru a evita bias confirmation. Cine implementează NU este același cu cine testează. Cine planifică NU este același cu cine execută.

---

## Roster (10 roluri)

### 1. Solution Architect
- **Model:** Claude Opus 4.7
- **Responsabilități:**
  - Decizii arhitecturale, design contracts între componente
  - Validare gate-uri între faze
  - Aprobare schimbări la plan (`00-PLAN-MASTER.md`)
  - Actualizare ROADMAP.md la finalizare fază
  - Risk owner pentru riscurile architecturale
- **Apare în:** F8.1 (review), F8.2 (lead), F8.4 (review), F8.7 (gate), F8.8 (sign-off)
- **Output:** decizii documentate, ADR-uri (Architecture Decision Records) când e cazul

### 2. Senior Backend Engineer (Go)
- **Model:** Claude Sonnet 4.6
- **Responsabilități:**
  - Implementare fix-uri backend (handlers, scheduler, engine)
  - Cod conform standardelor Go (effective Go, golint, gosec)
  - Unit tests pentru cod nou (coverage ≥ 80% pe engine, ≥ 70% pe handlers)
  - Documentare endpoints noi în `docs/integrations.md`
- **Apare în:** F8.2 (lead), F8.4 (lead), F8.5 (lead)
- **Output:** PR-uri cu cod + tests + actualizări docs

### 3. Senior Frontend Engineer (Next.js / TypeScript)
- **Model:** Claude Sonnet 4.6
- **Responsabilități:**
  - Implementare fix-uri frontend (pagini, componente)
  - Aliniere cu noile contracte API (envelope, errors)
  - Build curat (zero TS errors, zero warnings critice)
  - Anti-flash, accessibility minim AA
- **Apare în:** F8.6 (lead), F8.7 (review pentru E2E)
- **Output:** PR cu pagini/componente + build success

### 4. DBA / Database Engineer
- **Model:** Claude Sonnet 4.6
- **Responsabilități:**
  - Reconciliere schema DB (F8.1 lead)
  - Migrări versionate idempotente
  - Schema check script
  - Performance analysis pe queries critice (după F8.4)
  - Backup & recovery proceduri (în F8.8)
- **Apare în:** F8.1 (lead), F8.4 (review queries), F8.8 (perf check)
- **Output:** migrări SQL, schema diff reports

### 5. Senior QA Tester
- **Model:** Claude Sonnet 4.6
- **Responsabilități:**
  - Plan de test detaliat per fază (vezi `03-TESTING-STRATEGY.md`)
  - Automatizare TS-01..TS-12
  - Verificare gate-uri F8.x
  - Backlog: deschide DEV-XX noi descoperite în testing
  - Coverage report
- **Apare în:** F8.1–F8.6 (gate verifier), F8.7 (lead), F8.8 (lead pentru E2E)
- **Output:** test plans, test scripts, coverage reports, gap reports

### 6. Security Engineer
- **Model:** Claude Opus 4.7
- **Responsabilități:**
  - Audit API opacity (F8.3 lead)
  - Threat modeling pentru endpoints noi
  - Dependency scanning
  - Verificare reguli securitate (CLAUDE.md §7): drift/chaos_index/weights NU în răspunsuri
  - JWT, bcrypt cost, timing-safe verifications
- **Apare în:** F8.3 (lead), F8.7 (security tests), F8.8 (final scan)
- **Output:** opacity scan results, security report

### 7. Senior Project Manager
- **Model:** Claude Sonnet 4.6
- **Responsabilități:**
  - Owner BACKLOG.md
  - Status updates după fiecare sesiune
  - Verificare gate-uri (cu QA)
  - Escalare riscuri către Architect
  - Comunicare cu stakeholders (rapoarte sumar în română)
  - Decizii pe priorități când backlog-ul depășește capacitatea
- **Apare în:** TOATE fazele (gate verifier + status keeper)
- **Output:** status reports, backlog updates, decision logs

### 8. DevOps Engineer
- **Model:** Claude Sonnet 4.6
- **Responsabilități:**
  - CI pipeline (F8.7) — unit + integration + opacity + schema check
  - Deploy staging (F8.8)
  - Observability stack (logs, metrics, error tracking)
  - Secrets management verificare
  - Rollback procedure
- **Apare în:** F8.7 (lead pentru CI), F8.8 (lead pentru deploy)
- **Output:** `.github/workflows/*.yml`, deploy script, runbook staging

### 9. Product Manager / Framework Owner
- **Model:** Claude Opus 4.7
- **Responsabilități:**
  - Validare că fix-urile reflectă spec Framework Rev 5.6
  - Decide ce e MVP-required vs POST_MVP (când Architect și Backend au păreri diferite)
  - Update `MVP_SCOPE.md` dacă scope-ul se schimbă
  - Owner pentru `docs/user-workflow.md` updates
- **Apare în:** F8.2 (review formule), F8.4 (review semantică), F8.6 (review UX flow), F8.8 (sign-off)
- **Output:** decizii scope, validări semantice

### 10. UX Lead
- **Model:** Claude Sonnet 4.6
- **Responsabilități:**
  - Review flow onboarding (F8.6)
  - Verificare că mesajele utilizator sunt clare, în română corectă
  - Accessibility (contrast, keyboard nav, screen reader basic)
  - Empty states și error states
- **Apare în:** F8.6 (lead-co cu Frontend), F8.8 (E2E review)
- **Output:** UX review notes, copy adjustments

---

## Maparea responsabilităților pe DEV-XX

| DEV-ID | Owner principal | Reviewer | Faza |
|--------|-----------------|----------|------|
| DEV-01 | Security Engineer | Backend Senior + Architect | F8.3 |
| DEV-02 | DBA | Architect | F8.1 |
| DEV-03 | DBA | Architect | F8.1 |
| DEV-04 | DBA | Architect | F8.1 |
| DEV-05 | Backend Senior | Architect | F8.4 |
| DEV-06 | Backend Senior | QA | F8.5 |
| DEV-07 | Backend Senior | QA | F8.4 |
| DEV-08 | Backend Senior | Architect | F8.4 |
| DEV-09 | Backend Senior | PM (scope decision) | F8.4 |
| DEV-10 | Backend Senior | PM (scope decision) | F8.4 |
| DEV-11 | Architect + Backend Senior | PM | F8.2 |
| DEV-12 | Backend Senior | Architect | F8.5 |
| DEV-13 | Backend Senior | QA | F8.5 |
| DEV-14 | Backend Senior + Frontend Senior | UX | F8.5 + F8.6 |
| DEV-15 | Backend Senior | Architect | F8.5 |
| DEV-16 | Backend Senior | PM (Framework alignment) | F8.4 |
| DEV-17 | DBA | Backend Senior | F8.1 |
| DEV-18 | Backend Senior | PM (semantic alignment) | F8.4 |
| DEV-19 | Backend Senior | QA | F8.4 |
| DEV-20 | Backend Senior | QA | F8.5 |
| DEV-21 | Frontend Senior | UX + PM | F8.6 |
| DEV-22 | Backend Senior | QA | F8.5 |
| DEV-23 | Backend Senior | Architect | F8.5 |
| DEV-24 | — (already RESOLVED) | — | — |
| DEV-25 | Backend Senior | QA | F8.5 |
| DEV-26 | Backend Senior + Frontend Senior | UX | F8.5 + F8.6 |
| DEV-27 | Backend Senior + Frontend Senior | UX | F8.5 + F8.6 |
| DEV-28 | Backend Senior | Architect | F8.5 |

---

## Comunicarea între roluri

- **Sesiune-la-sesiune:** prin BACKLOG.md (status updates) + ROADMAP.md (faze)
- **PR comments:** doar comentarii substanțiale (per CLAUDE.md, "Be frugal about posting replies")
- **Escalări:** PM → Architect prin decision logs în `docs/fix-plan/decisions/` (creat la nevoie)
- **Status raport săptămânal:** PM agregă din BACKLOG + ROADMAP, postează în `docs/fix-plan/status/W<num>.md`

---

## Selecția modelului

**Sonnet 4.6** este default pentru execuție de cod, teste, fixe scopate.
**Opus 4.7** este folosit doar pentru:
- Decizii arhitecturale
- Security review-uri
- Framework alignment
- Debug complex multi-fișier
- Gate-uri majore (F8.7, F8.8)

Conform CLAUDE.md §4: „Regulă implicită: Sonnet. Opus doar când task-ul necesită raționament, nu implementare."
