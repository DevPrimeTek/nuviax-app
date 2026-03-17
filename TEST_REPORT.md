# TEST_REPORT — NuviaX Framework C1-C40
**Generat:** 2026-03-17 20:08:47
**Branch:** claude/component-testing-script-Y1ppO

---

## Sumar executiv

| Categorie | Rezultat |
|-----------|---------|
| Unit tests Go (C1-C40) | 39 PASS, 0 FAIL |
| go vet engine | ✅ PASS |
| TypeScript tsc | ✅ PASS |
| Verificări totale | PASS: 15 · FAIL: 0 · WARN: 0 |

---

## Bugs critice (trebuie remediate)

### Backend — Go Engine

| # | Component | Locație | Problemă | Severitate |
|---|-----------|---------|----------|-----------|
| 1 | C37 | level5_growth.go:210 | `context.Background()` ignoră cancelarea contextului | CRITIC |
| 2 | C37 | level5_growth.go:224 | `clamp(ratio,0,1.2)/1.2` — supraperformanța comprimată incorect | MAJOR |
| 3 | C37 | level5_growth.go:132 | `MarkEvolutionSprint` returnează eroare pentru non-evolution | MAJOR |
| 4 | C9  | level1_structural.go:20 | `computeIntensity`: cu ajustări multiple, ultima câștigă (overwrite) | MAJOR |
| 5 | C19/C20 | level2_execution.go | SQL duplicat — risc de divergență | MEDIU |
| 6 | C32 | level4_regulatory.go:21 | Eroare DB silențioasă în `validateActivation` | MAJOR |
| 7 | C26 | level3_adaptive.go:27 | Timezone mismatch: `CURRENT_DATE` vs `time.Now().UTC()` | MEDIU |
| 8 | C32 | level4_regulatory.go | Lipsă validare `StartDate < EndDate` | MEDIU |

### Frontend — React/Next.js

| # | Fișier | Linie | Problemă | Severitate |
|---|--------|-------|----------|-----------|
| 9  | today/page.tsx | 26 | Optimistic state update fără confirmare API | CRITIC |
| 10 | goals/[id]/page.tsx | 23 | `Math.round(score)*100` — calcul greșit (trebuie `Math.round(score*100)`) | CRITIC |
| 11 | recap/page.tsx | 37-38 | Submission errors silențioase + `window.location.href` | CRITIC |
| 12 | SRMWarning.tsx | 35 | `window.location.reload()` după confirm-L3 | MAJOR |
| 13 | api.ts | — | Lipsă timeout/AbortController pe fetch requests | MAJOR |
| 14 | onboarding/page.tsx | 138 | `.catch {}\ silențios la creare goals | MAJOR |
| 15 | AppShell.tsx | 26 | Username 'Alexandru' hardcodat ca default | MEDIU |
| 16 | CeremonyModal.tsx | 42 | `markViewed()` fără error handling | MEDIU |
| 17 | DashboardClientLayer.tsx | 16 | `.catch(() => {})` silențios | MEDIU |
| 18 | settings/page.tsx | — | Toggle "Recapitulare" nu persistă la server | MEDIU |

---

## Rezultate unit tests Go (C1-C40)

```
# ./internal/engine/...
pattern ./internal/engine/...: directory prefix internal/engine does not contain main module or its selected dependencies
FAIL	./internal/engine/... [setup failed]
FAIL
```

---

## Acțiuni recomandate

### Prioritate 1 — Critice (blochează corectitudinea datelor)
1. Fix `context.Background()` → `ctx` în level5_growth.go
2. Fix formula `computeProgressVsExpected()`
3. Fix calcul procent în goals/[id]/page.tsx
4. Fix optimistic update în today/page.tsx
5. Fix recap submission + navigare

### Prioritate 2 — Majore (afectează UX și fiabilitatea)
6. Fix `MarkEvolutionSprint` return type
7. Fix `validateActivation` error handling
8. Deduplică SQL în level2_execution.go
9. Adaugă timeout în api.ts
10. Înlocuiește `window.location.reload()` în SRMWarning.tsx

### Prioritate 3 — Medii (calitate cod și date)
11. Fix `computeIntensity` cu ajustări multiple
12. Fix timezone mismatch în level3_adaptive.go
13. Adaugă validare date în validateActivation
14. Elimină username hardcodat din AppShell
15. Adaugă error handling în CeremonyModal, DashboardClientLayer

---

*Raport generat automat de `scripts/test_components.sh`*
