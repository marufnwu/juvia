#!/bin/bash
BASE="http://127.0.0.1:18473/api/v1"
TIMESTAMP=$(date +%s)

echo "=== Juvia E2E Test on Clean Server ==="
echo ""

LOGIN_RESP=$(curl -sf -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"TestPass123!"}')
TOKEN=$(echo "$LOGIN_RESP" | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"//;s/"//')

if [ -z "$TOKEN" ]; then
  echo "Login failed - running first-run setup..."
  SETUP_RESP=$(curl -sf -X POST "$BASE/setup/first-run" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"TestPass123!","email":"admin@test.local"}')
  TOKEN=$(echo "$SETUP_RESP" | grep -o '"token":"[^"]*"' | sed 's/"token":"//;s/"//')
fi

echo "Token: ${TOKEN:0:30}..."

DOMAIN="test${TIMESTAMP}.com"
echo ""
echo "Creating website: $DOMAIN"

WEBSITE_RESP=$(curl -sf -w "\nHTTP:%{http_code}" -X POST "$BASE/websites" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"domain\":\"$DOMAIN\",\"php_version\":\"8.2\",\"web_server\":\"nginx\"}")

BODY=$(echo "$WEBSITE_RESP" | grep -v HTTP:)
HTTP=$(echo "$WEBSITE_RESP" | grep HTTP: | cut -d: -f2)

echo "HTTP: $HTTP"
echo "Response: $BODY"

if [ "$HTTP" = "202" ]; then
  echo ""
  echo "=== Checking system ==="
  echo "Nginx status:"
  systemctl status nginx --no-pager -n 2

  echo ""
  echo "PHP-FPM status:"
  systemctl status php8.2-fpm --no-pager -n 2

  echo ""
  echo "Nginx sites-enabled:"
  ls -la /etc/nginx/sites-enabled/ 2>/dev/null | head -10

  echo ""
  echo "PHP-FPM pool.d:"
  ls -la /etc/php/8.2/fpm/pool.d/ 2>/dev/null | head -10

  echo ""
  echo "User home:"
  ls -la /home/ | grep test
fi