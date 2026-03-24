# NuviaX — Roadmap de Dezvoltare

> Acest document reflectă starea curentă a proiectului și pașii următori în ordine de prioritate.
> Se actualizează la fiecare versiune majoră.
> **Ultima actualizare:** v10.1.0 — 2026-03-24

---

## Stare Curentă: v10.2.0

| Categorie | Status |
|-----------|--------|
| Backend Go — NUViaX Framework | ✅ 40/40 componente |
| Admin Panel (backend + frontend) | ✅ Complet |
| Critical Gaps Stress Test (P0) | ✅ 5/5 rezolvate |
| Medium Gaps Stress Test (P1) | ⏳ 0/12 rezolvate |
| Bug-uri UI/UX critice (B-3,B-7,B-8) | ✅ Rezolvate |
| Bug-uri UI/UX majore (B-5,B-6,B-9,B-11) | ✅ Rezolvate |
| Bug-uri medii (B-2,B-4,B-10) | ✅ Rezolvate |
| Integrare AI (Claude Haiku 4.5) | ✅ Implementat (cu graceful fallback) |
| Integrare Email (Resend) | ❌ Neimplementat |
| Tema Light CSS | ✅ Implementat (variabile + bloc light) |
| Traduceri EN/RU | ❌ Neimplementat |
| Monetizare (Stripe) | 📅 Planificat târziu |

---

## Sprint 1 — Bug Fixes Critice + AI + Email ✅ PARȚIAL COMPLET
*v10.2.0 — toate bug-urile rezolvate; email Resend rămâne*

### ✅ Bug-uri Critice (Rezolvate în v10.2.0)

**B-7 — Pagina Obiective goală** ✅
- `GetGoals` acum returnează `{goals:[], waiting:[]}` în loc de array plat

**B-8 — Pagina Recap returnează 404** ✅
- `GET /api/v1/recap/current` + `POST /api/v1/goals/:id/recap` implementate

**B-3 — Sprint afișează 89 zile în loc de 30** ✅
- `daysLeft` folosește `currentSprint.EndDate` când sprint activ există

### ✅ Bug-uri Majore (Rezolvate în v10.2.0)

**B-5 — "Cum mă simt" nu salvează energia** ✅
- Endpoint corectat la `/context/energy`; nivel `mid→normal`, `hi→high`; `goal_id` auto-detectat

**B-6 — Nu se pot adăuga sarcini personale** ✅
- Input + buton "+" adăugat în `today/page.tsx` → `POST /today/personal`

**B-11 — CSS variabile lipsă** ✅
- `--ff-h` adăugat; `[data-theme="light"]` există și e complet

**B-9 — Settings parțial conectate** ✅
- Export date (JSON download), schimbare parolă (modal complet)

### ✅ Bug-uri Medii (Rezolvate în v10.2.0)

**B-4 — Activități zilnice generice** ✅
- Claude Haiku 4.5 cu fallback pe template-uri statice

**B-2 — Analiza GO fără AI** ✅
- Claude Haiku 4.5 cu fallback pe analiza rule-based

**B-10 — Profil fără foto** ✅
- Avatar clickabil, upload local, previzualizare în timp real

### ✉️ Integrare Email (Resend.com)

**Pași necesari:**
1. Creare cont Resend.com
2. Adaugă domeniu `nuviax.app` → obții DNS records
3. Configurezi pe name.com: TXT (SPF), TXT (DKIM), CNAME (tracking)
4. Variabile `RESEND_API_KEY` + `EMAIL_FROM` în `.env` și GitHub Secrets
5. Implementare Go `pkg/email/email.go`
6. Email-uri necesare:
   - Confirmare înregistrare (cu link activare)
   - Reset parolă
   - Notificare sprint completat
   - Notificare ceremony generată

### 🤖 Integrare Claude Haiku

**Pași necesari:**
1. Variabilă `ANTHROPIC_API_KEY` în `.env` și GitHub Secrets
2. Dependință Go: `go get github.com/anthropics/anthropic-sdk-go`
3. Implementare `internal/ai/ai.go` cu client Haiku
4. Înlocuiește `generateTaskTexts` în `level1_structural.go`
5. Upgrade `AnalyzeGO` în `handlers.go`

**Cost estimat:** $4-5/lună la 1.000 utilizatori activi
**Model recomandat:** `claude-haiku-4-5-20251001` — $0.25/1M tokens input

---

## Sprint 2 — Medium Gaps Stress Test (P1)
*12 gap-uri din simulare — valori de referință din `NUVIAX_Stress_Test_Simulation.docx`*

| Gap | Titlu | Valoare referință |
|-----|-------|-----------------|
| #1 | Deadline recalcul după pauză | Sprint curent + zile pauză |
| #2 | Focus Rotation algorithm | Redirect atenție spre GO cu stagnare |
| #3 | Chaos Index formula exactă | 0.40 threshold → SRM L2 |
| #4 | GORI calculation complet | Formula cu ponderi per sprint |
| #5 | Stagnation detection granular | 5 zile consecutive fără progres |
| #6 | Velocity Control activare | ALI_projected > 1.15 |
| #7 | Reactivation Protocol pași | 7 zile stabilitate → propunere reactivare |
| #8 | Sprint score formula completă | Ponderi: 40% completion, 25% consistency, 25% progress, 10% energy |
| #9 | Annual Relevance Recalibration | La 90 zile per GO, 180 zile Vault |
| #10 | Future Vault cu recalibration | Max 3 active, restul în Vault |
| #11 | Behavior Model dominance | EVOLVE override la GO-uri hibride |
| #12 | SRM L1 auto + L2 manual flow | L1: automat, L2: confirmare, L3: confirmare dublă |

---

## Sprint 3 — UX Completare + Traduceri
*Funcționalitate completă, experiență de utilizare șlefuită*

- [ ] **Tema Light CSS** — bloc complet `[data-theme="light"]`
- [ ] **Traduceri EN** — toate textele din RO → EN (framework localizare există)
- [ ] **Traduceri RU** — RO → RU
- [ ] **Upload foto profil** — UI + backend + stocare
- [ ] **Onboarding îmbunătățit** — AI suggestions la clasificare GO
- [ ] **Notificări push** — PWA web push pentru reminders zilnice
- [ ] **Pagina profil** — statistici personale avansate, calendar activitate
- [ ] **Export date** — GDPR compliance (endpoint există, UI lipsă)

---

## Sprint 4 — Monetizare (Planificat)
*Doar după ce aplicația funcționează complet și are utilizatori reali*

- [ ] **Stripe integration** — subscripție lunară/anuală
- [ ] **Free tier limits** — max 1 GO, fără vizualizare, fără achievements
- [ ] **Pro tier** — 3 GO, toate funcționalitățile
- [ ] **Trial 14 zile** — acces complet fără card
- [ ] **Billing portal** — upgrade/downgrade/cancel
- [ ] **Webhook Stripe** — actualizare status subscripție în timp real

---

## Decizii Tehnice Deschise

| Decizie | Opțiuni | Recomandare |
|---------|---------|------------|
| Stocare foto profil | Local VPS / AWS S3 / Cloudflare R2 | R2 — cel mai ieftin, S3-compatible |
| Notificări push | Web Push API / OneSignal | Web Push API nativ (FOSS) |
| Mobile | PWA / React Native / Expo | PWA pentru v1, React Native mai târziu |
| Căutare full-text | PostgreSQL tsvector / Meilisearch | PostgreSQL tsvector (fără infra extra) |
| Analytics | Plausible / PostHog / self-hosted | Plausible (GDPR compliant, ieftin) |

---

## Decizii Confirmate

| Decizie | Alegere | Motiv |
|---------|---------|-------|
| AI provider | Anthropic Claude Haiku 4.5 | $4-5/lună la 1K users, context nativ NuviaX |
| Email tranzacțional | Resend.com | Setup 15 min, 3K email/lună gratis |
| Email business | Microsoft 365 (existent) | Deja configurat pe domeniu proprietar |
| DNS | name.com (existent) | Deja configurat |
| CI/CD | GitHub Actions → DockerHub → SSH | Configurat și funcțional |
| Proxy | nginx-proxy + acme-companion | Shared cu alte proiecte pe același VPS |

---

*Actualizare roadmap: la fiecare versiune majoră sau decizie arhitecturală importantă*
