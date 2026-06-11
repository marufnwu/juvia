#!/bin/bash
BASE=http://127.0.0.1:18473/api/v1

echo "Getting token..."
LOGIN_RESP=$(curl -sf -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"TestPass123!"}')

echo "Login response: $LOGIN_RESP"

TOKEN=$(echo "$LOGIN_RESP" | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//;s/"//')
echo "Token: ${TOKEN:0:30}..."

if [ -z "$TOKEN" ]; then
  echo "ERROR: No token received"
  exit 1
fi

AUTH="Authorization: Bearer $TOKEN"

echo ""
echo "Creating website..."
WEBSITE_RESP=$(curl -sf -X POST "$BASE/websites" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"domain":"newsite.com","php_version":"8.2","web_server":"nginx"}')

echo "Website response: $WEBSITE_RESP"

echo ""
echo "Checking /etc/juvia..."
ssh maruf@192.168.0.211 "ls -la /etc/juvia/ 2>/dev/null || echo '/etc/juvia not found'; ls -la /etc/juvia/nginx/ 2>/dev/null || echo '/etc/juvia/nginx not found'"