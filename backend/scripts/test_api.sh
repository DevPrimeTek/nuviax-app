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
  echo "5. Goals:"
  GOALS_RESP=$(curl -s -H "Authorization: Bearer $TOKEN" "$BASE_URL/goals")
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/goals")
  check "GET /goals" "$STATUS"

  GOAL_ID=$(echo "$GOALS_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('goals',[{}])[0].get('id',''))" 2>/dev/null || echo "")
  if [ -n "$GOAL_ID" ]; then
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/srm/status/$GOAL_ID")
    check "GET /srm/status/:goalId" "$STATUS"

    STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE_URL/goals/$GOAL_ID/visualize")
    check "GET /goals/:id/visualize" "$STATUS"
  else
    echo "  ⚠️  No goals found — skipping goal-specific tests"
  fi
fi

echo ""
echo "========================================"
echo "Results: $pass passed, $fail failed"
[ $fail -eq 0 ] && echo "✅ All API tests passed!" || echo "❌ Some tests failed."
echo "========================================"
