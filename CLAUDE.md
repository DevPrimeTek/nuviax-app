# CLAUDE.md — NuviaX: Context Master

> **Citește acest fișier la ÎNCEPUTUL oricărei sesiuni noi. Confirmă versiunea și branch-ul înainte de orice task.**
> Documentație extinsă → `docs/` (nu e necesară la fiecare sesiune).

---

## 0. Protocol Sesiune Nouă

**Pași obligatorii la start:**

```powershell
# 1. Verifică starea repo
git status
git log --oneline -5
git branch --show-current

# 2. Confirmă înainte de a începe task-ul:
# "Am citit CLAUDE.md. Versiune curentă: X.X.X. Branch activ: Y. Task: Z."
```

**Reguli token:**
- Citește **maxim 3 fișiere** per task — nu explora global
- Specifică întotdeauna path exact + funcția/linia care te interesează
- Un singur task per sesiune — commit, apoi sesiune nouă
- Folosește `/compact` după fiecare subtask major (nu la final)
- **NU citi:** `node_modules/`, `vendor/`, `.next/`, `build/`, `dist/`

---

## 1. Ce este NuviaX

**NuviaX** este o platformă SaaS de management al obiectivelor personale și profesionale, bazată pe **NUViaX Growth Framework REV 5.6** — sistem proprietar cu 5 niveluri (Layer 0 + Level 1–5) și 40 componente (C1–C40).

**Principiu fundamental:** Toate calculele rulează exclusiv server-side. Clientul primește doar rezultate opace (%, grade A/B/C/D, liste sarcini). Nicio formulă nu iese din engine-ul Go.

| Produs | URL |
|--------|-----|
| Aplicație | `nuviax.app` (Next.js) |
| Landing | `nuviaxapp.com` (Next.js static) |
| API | `api.nuviax.app` (Go) |

**Proprietar:** DevPrimeTek — `github.com/DevPrimeTek/nuviax-app`
**Versiune curentă:** `10.4.1`
**Branch development:** `claude/*` → PR → `main`

---

## 2. Stack Tehnic

| Layer | Tehnologie | Versiune |
|-------|-----------|---------|
| Backend API | Go + Fiber v2 | Go 1.22, Fiber 2.52 |
| Database | PostgreSQL | 16 (Docker) |
| Cache/Sessions | Redis | 7 (Docker) |
| Frontend App | Next.js + TypeScript + Tailwind | Next.js 14, React 18 |
| Frontend Landing | Next.js (static) | Next.js 14 |
| Auth | JWT RS256 (RSA 4096-bit) | access: 15min, refresh: 7 zile |
| Email | Resend.com (tranzacțional) | — |
| AI | Anthropic Claude Haiku 4.5 | `claude-haiku-4-5-20251001` |
| CI/CD | GitHub Actions → DockerHub → VPS SSH | — |
| Proxy | nginx-proxy + acme-companion (jwilder) | shared |
| Deploy path | `/var/www/wxr-nuviax/` pe VPS | — |

---

## 3. Model Selection

| Task | Model recomandat |
|------|-----------------|
| Implementare cod, bugfix, refactoring de rutină | **Claude Sonnet** |
| Decizii arhitecturale, review critic, design sistem nou | **Claude Opus** |
| Rename, format, editări mici sub 20 linii | **Claude Haiku** |

**Regulă:** Pornește întotdeauna cu **Sonnet**. Treci la Opus doar dacă ești blocat pe ceva arhitectural complex.

---

## 4. Starea Curentă

**Versiune:** `10.4.2` | **Sprint activ:** Sprint 3

| Categorie | Status |
|-----------|--------|
| Framework REV 5.6 (40 componente) | ✅ 40/40 complet |
| P0 Gaps (stress test) | ✅ 5/5 rezolvate (v10.1) |
| P1 Gaps (stress test) | ✅ 11/12 — G-12 rămâne |
| Bug-uri B-2—B-11 | ✅ Toate rezolvate (v10.2) |
| AI Claude Haiku | ✅ Implementat (v10.2) |
| Email Resend | ✅ Implementat (v10.3) |
| Forgot/Reset parolă | ✅ Implementat (v10.3) |
| G-11 Behavior Model | ✅ Implementat (v10.4.2) |
| Traduceri EN/RU | ❌ Neimplementat |
| Monetizare Stripe | 📅 Planificat Sprint 4 |

---

## 5. Sprint Curent — Sprint 3

### Task-uri în ordine de prioritate

**G-11 — Behavior Model dominance** ✅ COMPLET (v10.4.2)
- ✅ Câmp nou: `dominant_behavior_model VARCHAR(20)` pe `global_objectives`
- ✅ Migration: `011_behavior_model.sql`
- ✅ `level5_growth.go`: ApplyEvolveOverride() pentru GO hibride
- ✅ `handlers.go`: CreateGoal cu dominant_behavior_model opțional
- ✅ Models, DB queries: dominant_behavior_model support

**Traduceri i18n EN + RU** ← URMĂTOR
- Creează `lib/i18n.ts` cu `useTranslation()` hook
- Detectare limbă din `settings.language` (câmp JSONB pe `users`)
- Fișiere: `public/locales/ro.json`, `en.json`, `ru.json`
- Migrează mai întâi doar `today/page.tsx` ca proof of concept

**Onboarding AI îmbunătățit**
- La clasificare GO nouă → Claude Haiku sugerează categoria
- Inspiră-te din patternul existent: `backend/internal/ai/ai.go`
- Fallback: dacă Haiku nu răspunde în 2s → utilizatorul alege manual

### Sprint 4 (mai târziu)
- Stripe: subscripție Pro ($9.99/lună) + Free tier limits + Trial 14 zile
- PWA + Notificări push
- Export PDF raport lunar
- Statistici avansate heatmap

---

## 6. Workflow Dezvoltare

```powershell
# Start sesiune
git checkout -b claude/feature-name-XXXXX

# Final sesiune
git add [fișiere specifice — NU git add .]
git commit -m "feat/fix/docs: descriere clară"
git push -u origin claude/feature-name-XXXXX
# → deschide PR spre main pe GitHub
```

**Convenții commit:**
```
feat:     funcționalitate nouă
fix:      corectare bug
docs:     documentație
refactor: restructurare fără funcționalitate nouă
chore:    configurare, dependențe
```

**NU comite niciodată:** `.env`, `.env.*`, `.keys/`, `node_modules/`, `vendor/`

---

## 7. Regula README.md

> **OBLIGATORIU:** Actualizează `README.md` la FIECARE sesiune care modifică:

| Eveniment | Ce actualizezi |
|-----------|---------------|
| Versiune nouă | Linia `**Versiune curentă:**` + tabelul Changelog |
| Funcționalitate nouă | Secțiunea API Endpoints |
| Migration nouă | Secțiunea Database (număr migrări, tabele) |
| Job scheduler nou | Tabelul Scheduler Jobs |
| Endpoint nou/eliminat | Secțiunea API Endpoints |
| Structură modificată | Secțiunea Structura Proiectului |
| Env variable nouă | Secțiunea Environment Variables |

```powershell
# Verificare rapidă la finalul sesiunii
Select-String "Versiune curentă" README.md
Select-String "Versiune curentă" CLAUDE.md
# Trebuie să coincidă
```

---

## 8. Deployment rapid

**Flow:** push pe `main` → GitHub Actions → DockerHub → SSH VPS → health check

```bash
# Health check după deploy
curl https://api.nuviax.app/health
# Expected: {"status":"ok","db":true,"redis":true}
```

**Detalii complete:** `docs/deployment.md`

---

## 9. Reguli Critice

```
NICIODATĂ nu expune în API:
❌ drift, chaos_index, continuity, weights, factors, penalties
❌ thresholds (0.25, 0.40, 0.60), formule, componente scor

EXPUNE DOAR:
✅ Progress % (0-100)
✅ Grade (A+, A, B, C, D)
✅ Ceremony (tier, message, badge)
✅ Achievements (ID, name, icon)
```

**Admin 404:** panel-ul admin returnează 404 (nu 403) pentru non-admini.
**Timing-safe:** `forgot-password` returnează mereu 200.
**Graceful degradation:** AI și Email funcționează fără cheile respective.

---

## 10. Referințe

| Resursă | Locație |
|---------|---------|
| Structura repo + design system | `docs/project-structure.md` |
| Schema DB completă + migrări | `docs/database-reference.md` |
| AI (Haiku) + Email (Resend) | `docs/integrations.md` |
| VPS, Docker, CI/CD, secrets | `docs/deployment.md` |
| Bug-uri + gap-uri stress test | `docs/history-bugs-gaps.md` |
| Formule framework (C1-C40) | `FORMULAS_QUICK_REFERENCE.md` |
| Plan dezvoltare | `ROADMAP.md` |
| Changelog detaliat | `CHANGES.md` |
| GitHub Secrets guide | `infra/GITHUB_SECRETS.md` |

---

*Ultima actualizare: 2026-03-29 — v10.4.2 — G-11 Behavior Model dominance completat, migration 011*
