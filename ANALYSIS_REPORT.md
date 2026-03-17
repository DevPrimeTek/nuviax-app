# Raport Analiză Build Curent — NuviaX

**Data:** 2026-03-17
**Branch:** claude/analyze-server-build-HIrNj
**Scope:** Verificare 11 puncte identificate de utilizator față de codul sursă curent

---

## Punct 1 — GO-uri: primul activ, celelalte inactive ✅ Corect

Fluxul `onboarding/page.tsx` trimite `waiting_list: i > 0` pentru fiecare GO:
- GO 1 → `ACTIVE` (Sprint 1 creat, checkpoints create, sarcini generate)
- GO 2, GO 3 → `WAITING`

`handlers.go:354` respectă logica. **Funcționează conform design-ului.**

---

## Punct 2 — Fereastra verificare: lipsă sugestii AI ❌

`AnalyzeGO` (`handlers.go:999`) folosește **reguli lingvistice statice**, nu AI/LLM.
Returnează o singură întrebare identică pentru orice GO vag. Câmpul de răspuns din
`onboarding/page.tsx` este gol — fără sugestii pre-generate.
Comentariul din cod: `// Faza 1: template-uri statice; Faza 2: AI-assisted` — **Faza 2 neimplementată.**

---

## Punct 3 — Sprint card afișează ~89 zile, 8% ❌ Bug calcul

**Cauza** (`handlers.go:276`):
```go
daysLeft := int(time.Until(g.EndDate).Hours() / 24)  // folosește END DATE al GOALULUI (90 zile)
```
Sprintul 1 durează 30 zile, dar `days_left` ia data de sfârșit a **goalului** (90 zile).
Afișajul trebuie să folosească `sprint.EndDate`, nu `goal.EndDate`.

---

## Punct 4 — Activități fără sens ❌ Template static

`level1_structural.go:72-84` generează texte din template-uri fixe:
```go
"Lucrează la: " + cp.Name      // → "Lucrează la: Etapa 1: Fundament"
"Avansează cu: " + cp.Name
"Finalizează o parte din: " + cp.Name
```
Textul GO-ului utilizatorului nu influențează sarcinile. **Faza 2 AI neimplementată.**

---

## Punct 5 — "Cum mă simt" nu funcționează ❌

**Problema 1:** `today/page.tsx:139` — click pe buton actualizează doar state local, nu apelează
`POST /api/proxy/context/energy`. Endpointul backend există dar e neconectat.

**Problema 2:** CSS — `--ul` și `--l2g` sunt **nedefinite** în `globals.css`, cauzând culori invizibile.

---

## Punct 6 — "Ce faci azi": fără creare sarcini noi ❌

Backend: `POST /api/v1/today/personal` există și funcționează (`handlers.go:686`).
Frontend: `today/page.tsx` nu are niciun buton, input sau formular pentru adăugare sarcini.

---

## Punct 7 — Pagina "Obiective": nu afișează niciun obiectiv ❌ Bug critic

**Incompatibilitate tip backend ↔ frontend:**

```go
// handlers.go:343 — returnează ARRAY plat
return c.JSON(goals)  // → [{...}, {...}]
```
```ts
// lib/api.ts:92 — frontend se așteaptă la OBIECT
req<{goals:Goal[]; waiting:Goal[]}>('/v1/goals', ...)
// data.goals = undefined → allGoals = [] → pagina goală
```

---

## Punct 8 — Pagina "Recap": complet goală ❌ Endpoint lipsă

Frontend apelează `GET /api/v1/recap/current`.
**Nu există nicio rută `/recap/*` în `server.go`.** → 404 → "Nu ai nicio recapitulare disponibilă."

Logica de recap există (SaveReflection, CloseSprint) dar nu e expusă ca GET endpoint.

---

## Punct 9 — Setări: funcționalitate parțială ❌

| Element | Stare | Cauza |
|---|---|---|
| Limbă (RO/EN/RU) | ✅ | localStorage + PATCH backend |
| Temă Dark/Light | ⚠️ | `dataset.theme` setat, dar **lipsă CSS `[data-theme="light"]`** |
| Notificări toggle | ❌ | Doar state local, fără API |
| Verificare 2 pași | ❌ | Badge static "Activ", nicio logică |
| Schimbă parola | ❌ | Niciun `onClick`, nicio pagină |
| Descarcă datele | ❌ | Endpointul există (`GET /settings/export`), neconectat |

---

## Punct 10 — Profil: fără upload foto ❌

`profile/page.tsx` afișează inițiale. Lipsesc: `<input type="file">`, preview, handler upload,
endpoint backend pentru imagine.

---

## Punct 11 — Design nu coincide cu mockup-ul ❌

**Variabile CSS nedefinite în `globals.css`:**
- `--ul` — folosit în streak, titlu, scor
- `--l2g` — folosit în badge-uri
- `--ff-h` (font heading) — referit în cod, nedefinit

**Tema Light:** Lipsă bloc `[data-theme="light"] { ... }` în globals.css.
**Fonturi:** `Bricolage Grotesque`, `DM Sans`, `JetBrains Mono` declarate dar neimportate.

---

## Prioritizare

| Prioritate | Punct | Severitate |
|---|---|---|
| 🔴 BLOCKER | #7 Obiective nu apar — mismatch API | Critic |
| 🔴 BLOCKER | #8 Recap rupt — endpoint 404 | Critic |
| 🔴 BLOCKER | #3 Sprint zile incorecte — calcul greșit | Major |
| 🟠 MAJOR | #6 Fără creare sarcini în "Ce faci azi" | Major |
| 🟠 MAJOR | #5 "Cum mă simt" nu salvează energia | Major |
| 🟠 MAJOR | #4 Activități generice fără context | Major |
| 🟠 MAJOR | #9 Setări parțiale (temă, notificări, parolă) | Major |
| 🟡 MEDIU | #2 Fără sugestii AI la verificare GO | Mediu |
| 🟡 MEDIU | #10 Fără upload foto profil | Mediu |
| 🟡 MEDIU | #11 CSS variabile lipsă, fără light theme | Mediu |
| ✅ OK | #1 GO activation funcționează corect | — |
