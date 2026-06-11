#!/bin/bash
BASE="http://127.0.0.1:18473/api/v1"

echo "[1] Login..."
LOGIN_RESP=$(curl -sf -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"TestPass123!"}')
TOKEN=$(echo "$LOGIN_RESP" | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//;s/"//')
AUTH="Authorization: Bearer $TOKEN"

echo ""
echo "[2] Create website with headers..."
curl -si -X POST "$BASE/websites" \
  -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"domain":"test888.com","php_version":"8.2","web_server":"nginx"}'