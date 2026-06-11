#!/bin/bash
set -e

BASE="http://localhost:18473/api/v1"
FAILED=0

pass() { echo "  PASS: $1"; }
fail() { echo "  FAIL: $1"; FAILED=1; }

echo "=== Juvia E2E Test Suite ==="
echo ""

echo "[1/15] Health check..."
result=$(curl -sf "$BASE/health" 2>/dev/null) && pass "Health check" || fail "Health check"

echo "[2/15] Setup status..."
result=$(curl -sf "$BASE/setup/status" 2>/dev/null) && pass "Setup status" || fail "Setup status"

echo "[3/15] First-run setup..."
result=$(curl -sf -X POST "$BASE/setup/first-run" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"TestPass123!","email":"admin@test.local"}' 2>/dev/null)
if [ -n "$result" ]; then
  ADMIN_TOKEN=$(echo "$result" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
  if [ -n "$ADMIN_TOKEN" ]; then
    pass "First-run setup"
  else
    fail "First-run setup - no token"
  fi
else
  fail "First-run setup - no response"
fi

echo "[4/15] Login..."
result=$(curl -sf -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"TestPass123!"}' 2>/dev/null)
if [ -n "$result" ]; then
  TOKEN=$(echo "$result" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
  if [ -n "$TOKEN" ]; then
    pass "Login"
  else
    fail "Login - no token"
  fi
else
  fail "Login - no response"
fi

AUTH="Authorization: Bearer $TOKEN"

echo "[5/15] Get current user..."
result=$(curl -sf "$BASE/auth/me" -H "$AUTH" 2>/dev/null) && pass "Get current user" || fail "Get current user"

echo "[6/15] Create website..."
result=$(curl -sf -X POST "$BASE/websites" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"domain":"test.local","php_version":"8.2","web_server":"nginx"}' 2>/dev/null)
if echo "$result" | grep -q '"success":true'; then
  pass "Create website"
else
  fail "Create website"
fi

echo "[7/15] List websites..."
result=$(curl -sf "$BASE/websites" -H "$AUTH" 2>/dev/null)
if echo "$result" | grep -q "test.local"; then
  pass "List websites"
else
  fail "List websites"
fi

echo "[8/15] Create DNS record..."
WEBSITE_ID=$(curl -sf "$BASE/websites" -H "$AUTH" 2>/dev/null | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
if [ -n "$WEBSITE_ID" ]; then
  result=$(curl -sf -X POST "$BASE/websites/$WEBSITE_ID/dns/records" \
    -H "$AUTH" -H "Content-Type: application/json" \
    -d '{"type":"A","name":"test.local","value":"192.168.0.211"}' 2>/dev/null)
  if echo "$result" | grep -q '"success":true'; then
    pass "Create DNS record"
  else
    fail "Create DNS record"
  fi
else
  fail "Create DNS record - no website ID"
fi

echo "[9/15] Create database..."
result=$(curl -sf -X POST "$BASE/databases" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"name":"testdb","engine":"mysql"}' 2>/dev/null)
if echo "$result" | grep -q '"success":true'; then
  pass "Create database"
else
  fail "Create database"
fi

echo "[10/15] List databases..."
result=$(curl -sf "$BASE/databases" -H "$AUTH" 2>/dev/null)
if echo "$result" | grep -q "testdb"; then
  pass "List databases"
else
  fail "List databases"
fi

echo "[11/15] Create mailbox..."
result=$(curl -sf -X POST "$BASE/email/mailboxes" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"email":"admin@test.local","password":"TestPass123!","quota":1000}' 2>/dev/null)
if echo "$result" | grep -q '"success":true'; then
  pass "Create mailbox"
else
  fail "Create mailbox"
fi

echo "[12/15] Create firewall rule..."
result=$(curl -sf -X POST "$BASE/firewall/rules" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"name":"Test","action":"allow","port":"3000","protocol":"tcp","source":"any"}' 2>/dev/null)
if echo "$result" | grep -q '"success":true'; then
  pass "Create firewall rule"
else
  fail "Create firewall rule"
fi

echo "[13/15] Create cron job..."
result=$(curl -sf -X POST "$BASE/cron" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"schedule":"* * * * *","command":"echo test","run_as":"root","type":"custom"}' 2>/dev/null)
if echo "$result" | grep -q '"success":true'; then
  pass "Create cron job"
else
  fail "Create cron job"
fi

echo "[14/15] Update settings..."
result=$(curl -sf -X PUT "$BASE/settings" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"server_name":"Test Server"}' 2>/dev/null)
if echo "$result" | grep -q '"success":true'; then
  pass "Update settings"
else
  fail "Update settings"
fi

echo "[15/15] Get current metrics..."
result=$(curl -sf "$BASE/metrics/current" -H "$AUTH" 2>/dev/null) && pass "Get current metrics" || fail "Get current metrics"

echo ""
if [ $FAILED -eq 0 ]; then
  echo "=== ALL 15 TESTS PASSED ==="
  exit 0
else
  echo "=== SOME TESTS FAILED ==="
  exit 1
fi