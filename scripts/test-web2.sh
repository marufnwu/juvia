#!/bin/bash
BASE="http://127.0.0.1:18473/api/v1"

echo "[1] Login..."
LOGIN_RESP=$(curl -sf -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"TestPass123!"}')
echo "Login: $LOGIN_RESP"

TOKEN=$(echo "$LOGIN_RESP" | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//;s/"//')
echo "Token obtained: ${TOKEN:0:20}..."

AUTH="Authorization: Bearer $TOKEN"

echo ""
echo "[2] Create website..."
WEBSITE_RESP=$(curl -sf -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE/websites" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"domain":"test999.com","php_version":"8.2","web_server":"nginx"}')
echo "Response: $WEBSITE_RESP"

echo ""
echo "[3] Check directories..."
ssh maruf@192.168.0.211 "ls -la /etc/juvia/ 2>/dev/null && ls -la /etc/juvia/nginx/ 2>/dev/null || echo 'No /etc/juvia'"