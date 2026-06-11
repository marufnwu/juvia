#!/bin/bash
BASE="http://127.0.0.1:18473/api/v1"
TIMESTAMP=$(date +%s)

LOGIN_RESP=$(curl -sf -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"TestPass123!"}')
TOKEN=$(echo "$LOGIN_RESP" | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//;s/"//')

DOMAIN="test${TIMESTAMP}.com"
echo "Creating $DOMAIN..."

WEBSITE_RESP=$(curl -sf -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE/websites" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"domain\":\"$DOMAIN\",\"php_version\":\"8.2\",\"web_server\":\"nginx\"}")

HTTP_CODE=$(echo "$WEBSITE_RESP" | grep HTTP_CODE | cut -d: -f2)
BODY=$(echo "$WEBSITE_RESP" | grep -v HTTP_CODE)

echo "HTTP: $HTTP_CODE"
echo "Body: $BODY"

WEBSITE_ID=$(echo "$BODY" | grep -o '"website_id":[0-9]*' | cut -d: -f2)

echo ""
echo "Checking symlinks for $DOMAIN..."

ssh maruf@192.168.0.211 "ls -la /etc/nginx/sites-enabled/" | grep "$DOMAIN"
echo "---"
ssh maruf@192.168.0.211 "ls -la /etc/php/8.2/fpm/pool.d/" | grep "$DOMAIN"