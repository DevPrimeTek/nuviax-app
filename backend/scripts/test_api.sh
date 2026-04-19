#!/bin/bash
# NUViaX API Endpoint Tests
# Usage: TOKEN=<jwt> ./test_api.sh [BASE_URL]

BASE_URL="${1:-http://localhost:8080/api/v1}"
TOKEN="${TOKEN:-}"

if [ -z "$TOKEN" ]; then
  echo "⚠️  No TOKEN set. Skipping authenticated tests."
  echo "   Run: TOKEN=<your-jwt> ./test_api.sh"
  echo ""
fi

pass=0
fail=0

check() {
  local label="$1"
  local status="$2"
  local expected="${3:-200}"
  if [ "$status" = "$expected" ]; then
    echo "  ✅ $label (HTTP $status)"
    pass=$((pass + 1))
  else
    echo "  ❌ $label (HTTP $status, expected $expected)"
    fail=$((fail + 1))
  fi
}

echo "========================================"
echo "API Endpoint Tests → $BASE_URL"
echo "========================================"
echo ""

# Health
echo "1. Health check:"
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/../health" 2>/dev/null || echo "000")
check "GET /health" "$STATUS"
echo ""

if [ -n "$TOKEN" ]; then
  AUTH="-H \"Authorization: Bearer $TOKEN\""

  echo "2. Dashboard:"
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/dashboard")
  check "GET /dashboard" "$STATUS"

  echo ""
  echo "3. Achievements:"
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/achievements")
  check "GET /achievements" "$STATUS"

  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/achievements/progress")
  check "GET /achievements/progress" "$STATUS"

  echo ""
  echo "4. Ceremonies:"
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/ceremonies/unviewed")
  check "GET /ceremonies/unviewed" "$STATUS"

  echo ""
  echo "5. Onboarding flow (AI endpoints):"
  # /goals/analyze — SMART check
  ANALYZE_RESP=$(curl -s -H "Authorization: Bearer $TOKEN" -X POST \
    -H "Content-Type: application/json" \
    -d '{"text":"Vreau să slăbesc"}' \
    "$BASE_URL/goals/analyze")
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" -X POST \
    -H "Content-Type: application/json" \
    -d '{"text":"Vreau să slăbesc"}' \
    "$BASE_URL/goals/analyze")
  check "POST /goals/analyze" "$STATUS"
  if echo "$ANALYZE_RESP" | grep -q '"needs_clarification"'; then
    echo "  ✅ analyze response contains needs_clarification"
    pass=$((pass + 1))
  else
    echo "  ❌ analyze response missing needs_clarification field"
    fail=$((fail + 1))
  fi

  # /goals/suggest-category — C2 behavior_model + C9 directions
  SUGGEST_RESP=$(curl -s -H "Authorization: Bearer $TOKEN" -X POST \
    -H "Content-Type: application/json" \
    -d '{"title":"Vreau să alerg 5km în 30 minute","description":""}' \
    "$BASE_URL/goals/suggest-category")
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" -X POST \
    -H "Content-Type: application/json" \
    -d '{"title":"Vreau să alerg 5km în 30 minute","description":""}' \
    "$BASE_URL/goals/suggest-category")
  check "POST /goals/suggest-category" "$STATUS"
  if echo "$SUGGEST_RESP" | grep -q '"behavior_model"'; then
    echo "  ✅ suggest-category returns behavior_model"
    pass=$((pass + 1))
  else
    echo "  ❌ suggest-category missing behavior_model (onboarding will fail to create GO)"
    fail=$((fail + 1))
  fi
  if echo "$SUGGEST_RESP" | grep -q '"directions"'; then
    echo "  ✅ suggest-category returns directions[]"
    pass=$((pass + 1))
  else
    echo "  ❌ suggest-category missing directions[] (user cannot pick variant)"
    fail=$((fail + 1))
  fi

  echo ""
  echo "6. Goals:"
  GOALS_RESP=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/goals")
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/goals")
  check "GET /goals" "$STATUS"

  # API opacity — internal engine fields must never leak
  if echo "$GOALS_RESP" | grep -qiE 'drift|chaos_index|weights|threshold'; then
    echo "  ❌ API OPACITY VIOLATION — internal fields in /goals response"
    fail=$((fail + 1))
  else
    echo "  ✅ API opacity: no drift/chaos/weights/threshold leaked"
    pass=$((pass + 1))
  fi

  GOAL_ID=$(echo "$GOALS_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('goals',[{}])[0].get('id',''))" 2>/dev/null || echo "")
  if [ -n "$GOAL_ID" ]; then
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/srm/status/$GOAL_ID")
    check "GET /srm/status/:goalId" "$STATUS"

    STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/goals/$GOAL_ID/visualize")
    check "GET /goals/:id/visualize" "$STATUS"
  else
    echo "  ⚠️  No goals found — skipping goal-specific tests"
  fi

  echo ""
  echo "7. Admin guard (non-admin user must get 404, not 403):"
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/admin/stats")
  if [ "$STATUS" = "404" ]; then
    echo "  ✅ /admin/stats returns 404 for non-admin"
    pass=$((pass + 1))
  else
    echo "  ❌ /admin/stats returned $STATUS (expected 404 to avoid leaking admin surface)"
    fail=$((fail + 1))
  fi
fi

echo ""
echo "========================================"
echo "Results: $pass passed, $fail failed"
[ $fail -eq 0 ] && echo "✅ All API tests passed!" || echo "❌ Some tests failed."
echo "========================================"
