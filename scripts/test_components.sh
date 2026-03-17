#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════════════════════════
# NuviaX Framework — Script de testare C1-C40
# Testează toate componentele backend (Go) și frontend (TypeScript/Next.js)
# Generează TEST_REPORT.md cu analiza completă a problemelor găsite.
# ═══════════════════════════════════════════════════════════════════════════════

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND="$ROOT/backend"
FRONTEND="$ROOT/frontend/app"
REPORT="$ROOT/TEST_REPORT.md"
TIMESTAMP="$(date '+%Y-%m-%d %H:%M:%S')"

# Culori pentru output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Contoare
PASS=0
FAIL=0
WARN=0

log_pass()  { echo -e "${GREEN}✓${NC} $1"; PASS=$((PASS+1)); }
log_fail()  { echo -e "${RED}✗${NC} $1"; FAIL=$((FAIL+1)); }
log_warn()  { echo -e "${YELLOW}⚠${NC} $1"; WARN=$((WARN+1)); }
log_info()  { echo -e "${BLUE}→${NC} $1"; }
log_title() { echo -e "\n${BOLD}$1${NC}"; echo "$(printf '%.0s─' {1..60})"; }

# Salvăm output-ul pentru raport
REPORT_LINES=()
add_report() { REPORT_LINES+=("$1"); }

echo ""
echo -e "${BOLD}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║       NuviaX Framework — Test Suite C1-C40               ║${NC}"
echo -e "${BOLD}║       $(date '+%Y-%m-%d %H:%M')                               ║${NC}"
echo -e "${BOLD}╚══════════════════════════════════════════════════════════╝${NC}"
echo ""

# ─────────────────────────────────────────────────────────────────────────────
# SECȚIUNEA 1: BACKEND GO — C1-C40 Engine Tests
# ─────────────────────────────────────────────────────────────────────────────

log_title "SECȚIUNEA 1: Backend Go Engine (C1-C40)"
add_report "## Secțiunea 1: Backend Go Engine (C1-C40)"
add_report ""

cd "$BACKEND"

# 1.1 — go vet (analiză statică)
log_info "Rulăm go vet pe engine package..."
VET_OUTPUT=$(go vet ./internal/engine/... 2>&1 || true)
if [ -z "$VET_OUTPUT" ]; then
    log_pass "go vet ./internal/engine/... — fără probleme"
    add_report "- ✅ **go vet engine**: PASS — nicio problemă detectată"
else
    log_fail "go vet a găsit probleme:"
    echo "$VET_OUTPUT"
    add_report "- ❌ **go vet engine**: FAIL"
    add_report '```'
    add_report "$VET_OUTPUT"
    add_report '```'
fi

# 1.2 — go vet modele
log_info "Rulăm go vet pe models package..."
VET_MODELS=$(go vet ./internal/models/... 2>&1 || true)
if [ -z "$VET_MODELS" ]; then
    log_pass "go vet ./internal/models/... — fără probleme"
    add_report "- ✅ **go vet models**: PASS"
else
    log_fail "go vet models a găsit probleme"
    add_report "- ❌ **go vet models**: FAIL"
    add_report '```'
    add_report "$VET_MODELS"
    add_report '```'
fi

# 1.3 — Unit tests C1-C40
log_info "Rulăm unit tests pentru C1-C40..."
TEST_OUTPUT=$(go test ./internal/engine/... -v -count=1 2>&1 || true)
TESTS_PASS=$(echo "$TEST_OUTPUT" | grep -c "^--- PASS:" || true)
TESTS_FAIL=$(echo "$TEST_OUTPUT" | grep -c "^--- FAIL:" || true)
TEST_OK=$(echo "$TEST_OUTPUT" | grep -c "^ok" || true)
TEST_FAILED=$(echo "$TEST_OUTPUT" | grep -c "^FAIL" || true)

if [ "$TESTS_FAIL" -eq 0 ] && [ "$TEST_FAILED" -eq 0 ]; then
    log_pass "Unit tests engine: $TESTS_PASS PASS, 0 FAIL"
    add_report "- ✅ **Unit tests C1-C40**: $TESTS_PASS teste PASS, 0 FAIL"
else
    log_fail "Unit tests engine: $TESTS_PASS PASS, $TESTS_FAIL FAIL"
    add_report "- ❌ **Unit tests C1-C40**: $TESTS_PASS PASS, $TESTS_FAIL FAIL"
fi

# Extragere teste individuale
add_report ""
add_report "### Rezultate individuale unit tests:"
add_report '```'
echo "$TEST_OUTPUT" | grep -E "^(=== RUN|--- (PASS|FAIL)|PASS|FAIL|ok)" | while IFS= read -r line; do
    add_report "$line"
done
add_report "$TEST_OUTPUT" | grep -E "^(=== RUN|--- (PASS|FAIL)|PASS|FAIL|ok)" >> /dev/null || true
# Salvare directă
TEST_RESULTS=$(echo "$TEST_OUTPUT" | grep -E "^(=== RUN|--- (PASS|FAIL)|ok|FAIL)")
add_report "$TEST_RESULTS"
add_report '```'

# 1.4 — Analiză statică manuală a bug-urilor cunoscute (C1-C40)
add_report ""
add_report "### Bugs identificate în codul engine (analiză statică):"
add_report ""

log_title "SECȚIUNEA 1.4: Verificare manuală bug-uri cunoscute C1-C40"

# Bug 1: context.Background() în level5_growth.go
BUG1=$(grep -n "context.Background()" "$BACKEND/internal/engine/level5_growth.go" 2>/dev/null || true)
if [ -n "$BUG1" ]; then
    log_fail "C37 — context.Background() în loc de ctx parametru"
    add_report "- ❌ **C37 [CRITIC]**: \`context.Background()\` folosit în loc de \`ctx\` parametru (level5_growth.go:$(echo $BUG1 | grep -oP '^\d+'))"
    add_report "  - Impact: operațiunile DB ignoră cancelarea contextului (timeouts, request cancellation)"
else
    log_pass "C37 — fără context.Background() detectat"
    add_report "- ✅ **C37**: context propagation OK"
fi

# Bug 2: ratio / 1.2 în computeProgressVsExpected
BUG2=$(grep -n "/ 1.2" "$BACKEND/internal/engine/level5_growth.go" 2>/dev/null || true)
if [ -n "$BUG2" ]; then
    log_fail "C37 — formula ratio/1.2 distorsionează scorul"
    add_report "- ❌ **C37 [MAJOR]**: \`clamp(ratio,0,1.2)/1.2\` — supraperformanța e comprimată la 1.0 (level5_growth.go)"
    add_report "  - Impact: un obiectiv finalizat 100% în 50% din timp are același scor ca unul exact pe plan"
else
    log_pass "C37 — formula progress OK"
fi

# Bug 3: MarkEvolutionSprint returnează eroare pentru delta < 0.05
BUG3=$(grep -n 'evolution delta.*below threshold' "$BACKEND/internal/engine/level5_growth.go" 2>/dev/null || true)
if [ -n "$BUG3" ]; then
    log_fail "C37 — MarkEvolutionSprint returnează error pentru non-evolution (confuzie API)"
    add_report "- ❌ **C37 [MAJOR]**: \`MarkEvolutionSprint\` returnează \`fmt.Errorf\` când delta < 0.05"
    add_report "  - Impact: clientul nu poate distinge 'nu e evolution sprint' de 'eroare reală'"
else
    log_pass "C37 — MarkEvolutionSprint return type OK"
fi

# Bug 4: computeIntensity procesează ajustări cu prioritate corectă
# Fix: verificăm că există logică hasLow/hasHigh în loc de overwrite
BUG4_FIXED=$(grep -c "hasLow\|hasHigh" "$BACKEND/internal/engine/level1_structural.go" 2>/dev/null || echo 0)
if [ "$BUG4_FIXED" -ge 2 ]; then
    log_pass "C9 — computeIntensity: prioritate corectă pentru ajustări multiple"
    add_report "- ✅ **C9**: \`computeIntensity\` folosește prioritate corectă (Low > High)"
else
    log_warn "C9 — computeIntensity: cu ajustări multiple, ultima câștigă (overwrite)"
    add_report "- ⚠️  **C9 [MEDIU]**: \`computeIntensity\` suprascrie \`base\` pentru fiecare ajustare"
    add_report "  - Impact: ordinea ajustărilor din DB determină rezultatul (non-deterministic)"
fi

# Bug 5: cod duplicat în level2_execution.go
DUP_SQL=$(grep -c "COUNT.*FILTER.*WHERE completed" "$BACKEND/internal/engine/level2_execution.go" 2>/dev/null || echo 0)
if [ "$DUP_SQL" -ge 2 ]; then
    log_warn "C19/C20 — SQL duplicat în computeCompletionRate și computeSprintInternal"
    add_report "- ⚠️  **C19/C20 [MEDIU]**: SQL identic duplicat în \`computeCompletionRate\` și \`computeSprintInternal\`"
    add_report "  - Impact: risc de divergență la modificări viitoare"
else
    log_pass "C19/C20 — fără cod duplicat detectat"
fi

# Bug 6: validateActivation tratează eroarea DB ca limit-reached
BUG6=$(grep -n "err != nil || activeCount" "$BACKEND/internal/engine/level4_regulatory.go" 2>/dev/null || true)
if [ -n "$BUG6" ]; then
    log_fail "C32 — eroarea DB tratată identic cu limita de obiective atinsă"
    add_report "- ❌ **C32 [MAJOR]**: \`validateActivation\` tratează \`err != nil\` identic cu \`activeCount >= 3\`"
    add_report "  - Impact: eroarea DB e silențioasă, utilizatorul vede mesaj greșit"
else
    log_pass "C32 — validateActivation error handling OK"
fi

# Bug 7: timezone mismatch CURRENT_DATE vs time.Now().UTC()
# Verificăm că CURRENT_DATE nu apare în cod SQL (ignorăm comentariile // ...)
BUG7=$(grep -n "CURRENT_DATE" "$BACKEND/internal/engine/level3_adaptive.go" 2>/dev/null | grep -v "^\s*//" | grep -v "// " || true)
if [ -n "$BUG7" ]; then
    log_warn "C26 — CURRENT_DATE în SQL vs time.Now().UTC() în Go — potențial timezone mismatch"
    add_report "- ⚠️  **C26 [MEDIU]**: \`CURRENT_DATE\` (timezone server DB) vs \`time.Now().UTC()\` (UTC Go)"
    add_report "  - Impact: consistența calculată poate fi off-by-one la schimbarea zilei"
else
    log_pass "C26 — timezone mismatch fix aplicat (data parametrizată)"
    add_report "- ✅ **C26**: data curentă transmisă ca parametru SQL explicit"
fi

# Bug 8: validateActivation nu validează StartDate > EndDate
BUG8=$(grep -n "duration.Hours()/24 > 365" "$BACKEND/internal/engine/level4_regulatory.go" 2>/dev/null || true)
if [ -n "$BUG8" ]; then
    # Verificăm dacă există validare StartDate < EndDate
    HAS_DATE_VAL=$(grep -c "StartDate.*Before\|Before.*EndDate\|EndDate.*After" "$BACKEND/internal/engine/level4_regulatory.go" 2>/dev/null || echo 0)
    if [ "$HAS_DATE_VAL" -eq 0 ]; then
        log_warn "C32 — nu există validare că StartDate < EndDate în validateActivation"
        add_report "- ⚠️  **C32 [MEDIU]**: Lipsă validare \`StartDate < EndDate\` — durată negativă posibilă"
    fi
fi

echo ""

# ─────────────────────────────────────────────────────────────────────────────
# SECȚIUNEA 2: FRONTEND TypeScript/Next.js
# ─────────────────────────────────────────────────────────────────────────────

log_title "SECȚIUNEA 2: Frontend TypeScript/Next.js"
add_report ""
add_report "## Secțiunea 2: Frontend TypeScript/Next.js"
add_report ""

cd "$FRONTEND"

# 2.1 — TypeScript type check
log_info "Rulăm TypeScript type check (tsc --noEmit)..."
if [ -d "node_modules" ]; then
    TSC_OUTPUT=$(npx tsc --noEmit 2>&1 || true)
    TSC_ERRORS=$(echo "$TSC_OUTPUT" | grep -c "error TS" || true)
    if [ "$TSC_ERRORS" -eq 0 ]; then
        log_pass "TypeScript: 0 erori de tip"
        add_report "- ✅ **TypeScript (tsc --noEmit)**: 0 erori"
    else
        log_fail "TypeScript: $TSC_ERRORS erori"
        add_report "- ❌ **TypeScript (tsc --noEmit)**: $TSC_ERRORS erori de tip"
        add_report '```'
        echo "$TSC_OUTPUT" | head -30 | while IFS= read -r line; do
            add_report "$line"
        done
        add_report '```'
    fi
else
    log_warn "node_modules lipsă — rulăm npm install..."
    npm install --silent 2>&1 | tail -2
    TSC_OUTPUT=$(npx tsc --noEmit 2>&1 || true)
    TSC_ERRORS=$(echo "$TSC_OUTPUT" | grep -c "error TS" || true)
    if [ "$TSC_ERRORS" -eq 0 ]; then
        log_pass "TypeScript: 0 erori (după npm install)"
        add_report "- ✅ **TypeScript**: 0 erori (node_modules instalate)"
    else
        log_fail "TypeScript: $TSC_ERRORS erori după npm install"
        add_report "- ❌ **TypeScript**: $TSC_ERRORS erori"
    fi
fi

# 2.2 — Analiză statică manuală frontend
add_report ""
add_report "### Bugs identificate în frontend (analiză statică):"
add_report ""

log_title "SECȚIUNEA 2.2: Verificare manuală bug-uri frontend"

# Bug FE-1: Optimistic state update în today/page.tsx
# Verificăm dacă setTasks vine ÎNAINTE de await fetch (pattern buggy)
FE1=$(grep -n "setTasks" "$FRONTEND/app/today/page.tsx" 2>/dev/null | head -3 || true)
FE1_OPTIMISTIC=$(awk '/async function toggleTask/,/^  }/' "$FRONTEND/app/today/page.tsx" 2>/dev/null | grep -c "setTasks" || true)
FE1_FETCH_FIRST=$(awk '/async function toggleTask/,/^  }/' "$FRONTEND/app/today/page.tsx" 2>/dev/null | grep -n "await fetch\|setTasks" | head -2 | awk -F: '{print $2}' | tr '\n' ' ' || true)
# Verificăm că fetch vine ÎNAINTEA setTasks (fix corect)
FETCH_LINE=$(awk '/async function toggleTask/,/^  }/' "$FRONTEND/app/today/page.tsx" 2>/dev/null | grep -n "await fetch" | head -1 | cut -d: -f1 || echo 0)
SETTASKS_LINE=$(awk '/async function toggleTask/,/^  }/' "$FRONTEND/app/today/page.tsx" 2>/dev/null | grep -n "setTasks" | head -1 | cut -d: -f1 || echo 0)
if [ -n "$FETCH_LINE" ] && [ -n "$SETTASKS_LINE" ] && [ "$FETCH_LINE" -gt "$SETTASKS_LINE" ] 2>/dev/null; then
    log_fail "today/page.tsx — optimistic update fără confirmare API"
    add_report "- ❌ **today/page.tsx [CRITIC]**: State actualizat înainte de confirmarea API"
    add_report "  - Impact: dacă API-ul eșuează, UI-ul rămâne în stare greșită"
else
    log_pass "today/page.tsx — toggleTask actualizează state după confirmare API"
    add_report "- ✅ **today/page.tsx**: optimistic update fix aplicat"
fi

# Bug FE-2: window.location.href în recap/page.tsx (nu în context de auth redirect)
FE2=$(grep -n "window.location.href" "$FRONTEND/app/recap/page.tsx" 2>/dev/null || true)
if [ -n "$FE2" ]; then
    log_fail "recap/page.tsx — window.location.href în loc de router.push()"
    add_report "- ❌ **recap/page.tsx [MAJOR]**: \`window.location.href\` — full page reload în loc de Next.js router"
    add_report "  - Impact: performanță slabă, state-ul aplicației se pierde"
else
    log_pass "recap/page.tsx — folosește router.push() în loc de window.location.href"
    add_report "- ✅ **recap/page.tsx**: navigare corectă cu router.push()"
fi

# Bug FE-3: recap submission fără error handling
FE3=$(grep -n "await fetch.*recap.*catch.*{}" "$FRONTEND/app/recap/page.tsx" 2>/dev/null || true)
FE3B=$(grep -A2 "await fetch.*recap" "$FRONTEND/app/recap/page.tsx" 2>/dev/null | grep -c "catch.*{}" || true)
if [ "$FE3B" -gt 0 ] || grep -q "\.catch(() => {})" "$FRONTEND/app/recap/page.tsx" 2>/dev/null; then
    log_fail "recap/page.tsx — submission errors silențioase"
    add_report "- ❌ **recap/page.tsx [CRITIC]**: \`.catch(() => {})\` — utilizatorul nu știe dacă recapitularea s-a salvat"
fi

# Bug FE-4: Calcul incorect procente în goals/[id]/page.tsx
# Pattern buggy: Math.round(score) * 100 — înmulțire DUPĂ Math.round, last char before * 100 is ')'
# Pattern corect: Math.round(score * 100) — înmulțire ÎNAINTE, last char before ')' is '0' sau ')'
# Detectăm buggul: Math.round() urmat direct de ' * 100' (nu în interiorul parantezelor)
FE4=$(grep -nP "Math\.round\([^)]+\)\s*\*\s*100" "$FRONTEND/app/goals/[id]/page.tsx" 2>/dev/null || true)
if [ -n "$FE4" ]; then
    log_fail "goals/[id]/page.tsx — calcul greșit procent: Math.round(progress_score) * 100"
    add_report "- ❌ **goals/[id]/page.tsx [CRITIC]**: \`Math.round(score ?? 0) * 100\` — calcul greșit"
    add_report "  - Corect: \`Math.round((score ?? 0) * 100)\`"
else
    log_pass "goals/[id]/page.tsx — calcul procent corect: Math.round(score * 100)"
    add_report "- ✅ **goals/[id]/page.tsx**: calcul procent fix aplicat"
fi

# Bug FE-5: window.location.reload() în SRMWarning.tsx
FE5=$(grep -n "window.location.reload()" "$FRONTEND/components/SRMWarning.tsx" 2>/dev/null || true)
if [ -n "$FE5" ]; then
    log_warn "SRMWarning.tsx — window.location.reload() după confirm-L3"
    add_report "- ⚠️  **SRMWarning.tsx [MAJOR]**: \`window.location.reload()\` — full reload după confirmarea L3"
    add_report "  - Corect: re-fetch sau setState null pentru a actualiza UI"
fi

# Bug FE-6: .catch(() => {}) silențios în DashboardClientLayer.tsx
FE6=$(grep -n "catch.*{}" "$FRONTEND/components/DashboardClientLayer.tsx" 2>/dev/null || true)
if [ -n "$FE6" ]; then
    log_warn "DashboardClientLayer.tsx — erori silențioase"
    add_report "- ⚠️  **DashboardClientLayer.tsx [MEDIU]**: \`.catch(() => {})\` — erori la fetch ceremonies sunt silențioase"
fi

# Bug FE-7: username hardcodat în AppShell
FE7=$(grep -n "userName='Alexandru'" "$FRONTEND/components/layout/AppShell.tsx" 2>/dev/null || true)
if [ -n "$FE7" ]; then
    log_warn "AppShell.tsx — username hardcodat 'Alexandru'"
    add_report "- ⚠️  **AppShell.tsx [MEDIU]**: \`userName='Alexandru'\` hardcodat ca default prop"
    add_report "  - Impact: utilizatori noi văd un alt nume în UI"
fi

# Bug FE-8: markViewed fără error handling în CeremonyModal
FE8=$(grep -c "try" "$FRONTEND/components/CeremonyModal.tsx" 2>/dev/null || echo 0)
if [ "${FE8:-0}" -eq 0 ]; then
    log_warn "CeremonyModal.tsx — markViewed fără error handling"
    add_report "- ⚠️  **CeremonyModal.tsx [MEDIU]**: \`markViewed()\` fără try/catch — API failure nu e gestionat"
else
    log_pass "CeremonyModal.tsx — markViewed are error handling"
    add_report "- ✅ **CeremonyModal.tsx**: try/catch adăugat în markViewed"
fi

# Bug FE-9: no timeout pe fetch în api.ts
FE9=$(grep -c "AbortController\|timeout" "$FRONTEND/lib/api.ts" 2>/dev/null || true)
if [ "$FE9" -eq 0 ]; then
    log_warn "api.ts — nicio funcție de timeout/abort pentru fetch"
    add_report "- ⚠️  **api.ts [MAJOR]**: Lipsă timeout pe \`fetch\` — request-urile pot atârna la infinit"
fi

# Bug FE-10: settings toggle review fără persistare
FE10=$(grep -n "setReview" "$FRONTEND/app/settings/page.tsx" 2>/dev/null | grep -v "fetch\|API" | head -3 || true)
REVIEW_API=$(grep -A5 "setReview" "$FRONTEND/app/settings/page.tsx" 2>/dev/null | grep -c "fetch\|api" || true)
if [ "$REVIEW_API" -eq 0 ]; then
    log_warn "settings/page.tsx — toggle 'Recapitulare de etapă' nu persistă la server"
    add_report "- ⚠️  **settings/page.tsx [MEDIU]**: Toggle \`review\` nu face API call — preferința nu e salvată"
fi

echo ""

# ─────────────────────────────────────────────────────────────────────────────
# SECȚIUNEA 3: SUMAR GENERAL
# ─────────────────────────────────────────────────────────────────────────────

log_title "SUMAR GENERAL"

TOTAL_CHECKS=$((PASS + FAIL + WARN))
echo ""
echo -e "  ${GREEN}PASS:${NC}     $PASS"
echo -e "  ${RED}FAIL:${NC}     $FAIL"
echo -e "  ${YELLOW}AVERTISMENTE:${NC} $WARN"
echo -e "  ${BOLD}TOTAL:${NC}    $TOTAL_CHECKS"
echo ""

if [ "$FAIL" -eq 0 ]; then
    echo -e "${GREEN}${BOLD}✓ Toate verificările critice au trecut!${NC}"
else
    echo -e "${RED}${BOLD}✗ $FAIL verificări critice au eșuat — necesită remediere.${NC}"
fi

# ─────────────────────────────────────────────────────────────────────────────
# GENERARE RAPORT TEST_REPORT.md
# ─────────────────────────────────────────────────────────────────────────────

cat > "$REPORT" << REPORTEOF
# TEST_REPORT — NuviaX Framework C1-C40
**Generat:** $TIMESTAMP
**Branch:** $(git -C "$ROOT" branch --show-current 2>/dev/null || echo "unknown")

---

## Sumar executiv

| Categorie | Rezultat |
|-----------|---------|
| Unit tests Go (C1-C40) | ${TESTS_PASS:-0} PASS, ${TESTS_FAIL:-0} FAIL |
| go vet engine | $([ -z "${VET_OUTPUT:-}" ] && echo "✅ PASS" || echo "❌ FAIL") |
| TypeScript tsc | $([ "${TSC_ERRORS:-0}" -eq 0 ] && echo "✅ PASS" || echo "❌ ${TSC_ERRORS} erori") |
| Verificări totale | PASS: $PASS · FAIL: $FAIL · WARN: $WARN |

---

## Bugs critice (trebuie remediate)

### Backend — Go Engine

| # | Component | Locație | Problemă | Severitate |
|---|-----------|---------|----------|-----------|
| 1 | C37 | level5_growth.go:210 | \`context.Background()\` ignoră cancelarea contextului | CRITIC |
| 2 | C37 | level5_growth.go:224 | \`clamp(ratio,0,1.2)/1.2\` — supraperformanța comprimată incorect | MAJOR |
| 3 | C37 | level5_growth.go:132 | \`MarkEvolutionSprint\` returnează eroare pentru non-evolution | MAJOR |
| 4 | C9  | level1_structural.go:20 | \`computeIntensity\`: cu ajustări multiple, ultima câștigă (overwrite) | MAJOR |
| 5 | C19/C20 | level2_execution.go | SQL duplicat — risc de divergență | MEDIU |
| 6 | C32 | level4_regulatory.go:21 | Eroare DB silențioasă în \`validateActivation\` | MAJOR |
| 7 | C26 | level3_adaptive.go:27 | Timezone mismatch: \`CURRENT_DATE\` vs \`time.Now().UTC()\` | MEDIU |
| 8 | C32 | level4_regulatory.go | Lipsă validare \`StartDate < EndDate\` | MEDIU |

### Frontend — React/Next.js

| # | Fișier | Linie | Problemă | Severitate |
|---|--------|-------|----------|-----------|
| 9  | today/page.tsx | 26 | Optimistic state update fără confirmare API | CRITIC |
| 10 | goals/[id]/page.tsx | 23 | \`Math.round(score)*100\` — calcul greșit (trebuie \`Math.round(score*100)\`) | CRITIC |
| 11 | recap/page.tsx | 37-38 | Submission errors silențioase + \`window.location.href\` | CRITIC |
| 12 | SRMWarning.tsx | 35 | \`window.location.reload()\` după confirm-L3 | MAJOR |
| 13 | api.ts | — | Lipsă timeout/AbortController pe fetch requests | MAJOR |
| 14 | onboarding/page.tsx | 138 | \`.catch {}\ silențios la creare goals | MAJOR |
| 15 | AppShell.tsx | 26 | Username 'Alexandru' hardcodat ca default | MEDIU |
| 16 | CeremonyModal.tsx | 42 | \`markViewed()\` fără error handling | MEDIU |
| 17 | DashboardClientLayer.tsx | 16 | \`.catch(() => {})\` silențios | MEDIU |
| 18 | settings/page.tsx | — | Toggle "Recapitulare" nu persistă la server | MEDIU |

---

## Rezultate unit tests Go (C1-C40)

\`\`\`
$(go test ./internal/engine/... -v -count=1 2>&1)
\`\`\`

---

## Acțiuni recomandate

### Prioritate 1 — Critice (blochează corectitudinea datelor)
1. Fix \`context.Background()\` → \`ctx\` în level5_growth.go
2. Fix formula \`computeProgressVsExpected()\`
3. Fix calcul procent în goals/[id]/page.tsx
4. Fix optimistic update în today/page.tsx
5. Fix recap submission + navigare

### Prioritate 2 — Majore (afectează UX și fiabilitatea)
6. Fix \`MarkEvolutionSprint\` return type
7. Fix \`validateActivation\` error handling
8. Deduplică SQL în level2_execution.go
9. Adaugă timeout în api.ts
10. Înlocuiește \`window.location.reload()\` în SRMWarning.tsx

### Prioritate 3 — Medii (calitate cod și date)
11. Fix \`computeIntensity\` cu ajustări multiple
12. Fix timezone mismatch în level3_adaptive.go
13. Adaugă validare date în validateActivation
14. Elimină username hardcodat din AppShell
15. Adaugă error handling în CeremonyModal, DashboardClientLayer

---

*Raport generat automat de \`scripts/test_components.sh\`*
REPORTEOF

echo ""
log_pass "Raport generat: $REPORT"
echo ""

# Ieșim cu cod de eroare dacă există FAIL
exit $FAIL
