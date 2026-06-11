#!/bin/bash
BASE="http://127.0.0.1:18473/api/v1"

LOGIN_RESP=$(curl -sf -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"TestPass123!"}')
TOKEN=$(echo "$LOGIN_RESP" | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//;s/"//')

echo "Token: ${TOKEN:0:30}..."

echo ""
echo "=== List Websites ==="
curl -sf "$BASE/websites" -H "Authorization: Bearer $TOKEN"

echo ""
echo ""
echo "=== Website Detail ==="
curl -sf "$BASE/websites/1" -H "Authorization: Bearer $TOKEN"

echo ""
echo ""
echo "=== Create another website ==="
curl -sf -X POST "$BASE/websites" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"domain":"anothersite.com","php_version":"8.2","web_server":"nginx"}'