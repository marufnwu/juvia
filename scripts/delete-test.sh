#!/bin/bash
BASE="http://127.0.0.1:18473/api/v1"

LOGIN_RESP=$(curl -sf -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"TestPass123!"}')
TOKEN=$(echo "$LOGIN_RESP" | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//;s/"//')

echo "Deleting website 4..."
curl -sf -X DELETE "$BASE/websites/4" -H "Authorization: Bearer $TOKEN"
echo ""
echo "Done"