#!/bin/bash
BASE="http://127.0.0.1:18473/api/v1"
TIMESTAMP=$(date +%s)

echo "=== Juvia Website Creation E2E Test ==="
echo "Timestamp: $TIMESTAMP"
echo ""

LOGIN_RESP=$(curl -sf -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"TestPass123!"}')
TOKEN=$(echo "$LOGIN_RESP" | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//;s/"//')

if [ -z "$TOKEN" ]; then
  echo "FAIL: Login failed"
  exit 1
fi
echo "OK: Login successful"

DOMAIN="test${TIMESTAMP}.com"
echo "=== Creating website $DOMAIN ==="
WEBSITE_RESP=$(curl -sf -w "\nHTTP_CODE:%{http_code}" -X POST "$BASE/websites" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"domain\":\"$DOMAIN\",\"php_version\":\"8.2\",\"web_server\":\"nginx\"}")

HTTP_CODE=$(echo "$WEBSITE_RESP" | grep HTTP_CODE | cut -d: -f2)
BODY=$(echo "$WEBSITE_RESP" | grep -v HTTP_CODE)

if [ "$HTTP_CODE" != "202" ]; then
  echo "FAIL: Website creation returned $HTTP_CODE"
  echo "$BODY"
  exit 1
fi

WEBSITE_ID=$(echo "$BODY" | grep -o '"website_id":[0-9]*' | cut -d: -f2)
echo "OK: Website created with ID $WEBSITE_ID"

echo ""
echo "=== Verifying nginx config ==="
NGINX_SYMLINK=$(ssh maruf@192.168.0.211 "ls -la /etc/nginx/sites-enabled/ 2>/dev/null | grep '$DOMAIN'" | grep -c '\->')
if [ "$NGINX_SYMLINK" -gt 0 ]; then
  echo "OK: Nginx symlink exists"
else
  echo "FAIL: Nginx symlink not found"
  exit 1
fi

echo ""
echo "=== Verifying PHP-FPM pool config ==="
PHP_SYMLINK=$(ssh maruf@192.168.0.211 "ls -la /etc/php/8.2/fpm/pool.d/ 2>/dev/null | grep '$DOMAIN'" | grep -c '\->')
if [ "$PHP_SYMLINK" -gt 0 ]; then
  echo "OK: PHP-FPM pool symlink exists"
else
  echo "FAIL: PHP-FPM pool symlink not found"
  exit 1
fi

echo ""
echo "=== All tests passed! ==="