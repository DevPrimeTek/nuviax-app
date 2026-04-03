# CLAUDE.md — NuviaX Context Master (Architect Sync)

> Versiune: 11.0.0  
> Actualizat: 2026-04-03  
> Citește acest fișier la începutul fiecărei sesiuni.

---

## 0) Protocol start sesiune (obligatoriu)

```bash
git status
git log --oneline -5
git branch --show-current
```

Confirmare explicită înainte de lucru:
- versiune documente (README/CLAUDE/ROADMAP/PLAN)
- task exact
- fișiere țintă

---

## 1) Context produs

NuviaX este platformă SaaS de management al obiectivelor, bazată pe **NuviaX Growth Framework Rev 5.6**.

**Principiu tehnic critic:** engine-ul rămâne opac; formulele interne nu se expun în API.

---

## 2) Prioritate actuală a echipei de arhitectură

Focusul principal NU este adăugarea de features noi, ci alinierea completă la framework:

1. Behavior Model canonic
2. Sezonalitate (`execution_windows` + `SEASONAL_PAUSE`)
3. SRM single-active-level
4. Regression pipeline E2E
5. Temporal validity (A3)
6. Documentație "as-built" + test governance

Referințe obligatorii:
- `PLAN.md`
- `ROADMAP.md`
- `docs/framework_100_percent_implementation_playbook.md`
- `docs/framework_workflow_deviations_stress_test.md`
- `docs/testing/test-plan.md`

---

## 3) Reguli de lucru

- Nu modifica arhitectura fără să actualizezi documentația principală în același PR.
- Orice schimbare backend care afectează scor/SRM trebuie acoperită de test plan.
- Nu marca "DONE" fără criteriu de acceptanță verificabil.
- Evită afirmații de status care nu sunt susținute de codul curent.

---

## 4) Definiție de "Done" pentru task-uri framework

Un task este DONE doar când:
1. cod + migrare (dacă este cazul) sunt implementate,
2. docs principale sunt actualizate,
3. testele relevante sunt executate sau explicit marcate ca blocare de mediu,
4. impactul asupra C1..C40 este declarat.

---

## 5) Ghid rapid de modele

- Implementare standard: Sonnet
- Decizie arhitecturală complexă: Opus
- Editări minore de text/config: Haiku

---

## 6) Reguli de securitate (nemodificate)

Nu expune în API:
- drift/chaos/weights/factori interni
- praguri interne și formule

Expune doar:
- progres %, grade, task-uri, ceremonii, achievements

