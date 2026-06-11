#!/bin/bash
BASE="http://127.0.0.1:18473/api/v1"

LOGIN_RESP=$(curl -sf -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"TestPass123!"}')
TOKEN=$(echo "$LOGIN_RESP" | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//;s/"//')

echo "Creating new website (should trigger nginx reload with symlinks)..."
curl -sf -X POST "$BASE/websites" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"domain":"newtest123.com","php_version":"8.2","web_server":"nginx"}'

echo ""
echo "Checking symlinks..."
ssh maruf@192.168.0.211 "ls -la /etc/nginx/sites-enabled/ | grep -E 'newtest|test888|anothersite'"