#!/bin/bash
BASE=http://127.0.0.1:18473/api/v1

echo "=== Testing Juvia API ==="

echo ""
echo "[1] Health check..."
curl -sf "$BASE/health"

echo ""
echo ""
echo "[2] Login..."
RESP=$(curl -sf -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"TestPass123!"}')
echo "Login response: $RESP"

TOKEN=$(echo "$RESP" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
if [ -n "$TOKEN" ]; then
  echo "Token: ${TOKEN:0:30}..."
else
  echo "No token received!"
  exit 1
fi

AUTH="Authorization: Bearer $TOKEN"

echo ""
echo "[3] Get current user..."
curl -sf "$BASE/auth/me" -H "$AUTH"

echo ""
echo ""
echo "[4] Create website..."
curl -sf -X POST "$BASE/websites" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"domain":"test.local","php_version":"8.2","web_server":"nginx"}'

echo ""
echo ""
echo "[5] List websites..."
curl -sf "$BASE/websites" -H "$AUTH"

echo ""
echo ""
echo "[6] Get current metrics..."
curl -sf "$BASE/metrics/current" -H "$AUTH"

echo ""
echo ""
echo "=== Tests complete ==="