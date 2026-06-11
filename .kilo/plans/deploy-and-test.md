# Plan: Automated Deploy and Test on Remote Server

## Problem with Previous Plan
Manual step-by-step install and clicking through UI doesn't prove anything. We need:
1. **One-command install** via `install.sh` (proves the full dependency chain)
2. **Automated API test script** (proves every endpoint works, no UI needed)

## Target
- **Server:** `maruf@192.168.0.211`
- **Port:** `18473` (default for IP access, no hostname configured)

## Phase 1 — Build and Publish Release (Local)

### Step 1.1: Build Frontend
```bash
cd web && npm ci && npm run build && cd ..
```

### Step 1.2: Copy to Embed Directory
```bash
mkdir -p internal/web && rm -rf internal/web/dist && cp -r web/dist internal/web/dist
```

### Step 1.3: Build .deb Package
```bash
bash scripts/build-deb.sh 0.1.2 amd64
```

### Step 1.4: Upload Release Asset
Push the .deb to GitHub releases so `install.sh` can download it:
```bash
# Create release with asset (or use GitHub Actions release workflow)
gh release upload v0.1.2 juvia_0.1.2_amd64.deb --clobber
```

## Phase 2 — One-Command Install (Remote)

### Step 2.1: Run install.sh
```bash
ssh maruf@192.168.0.211 "curl -fsSL https://raw.githubusercontent.com/marufnwu/juvia/main/scripts/install.sh | sudo bash"
```

This single command should:
- Install ALL system dependencies (nginx, php-fpm, mysql, postfix, etc.)
- Download and install the .deb from GitHub releases
- Start `juvia` and `juvia-agent` services
- Print the access URL

### Step 2.2: Open Firewall
```bash
ssh maruf@192.168.0.211 "sudo ufw allow 18473/tcp"
```

## Phase 3 — Automated API Test Script

Create `scripts/test-e2e.sh` that runs on the remote server (or from local via SSH) and exercises every endpoint:

### Script Structure
```bash
#!/bin/bash
set -e

BASE="http://localhost:18473/api/v1"

echo "=== E2E Test Suite ==="

# 1. Health check
echo "[1/15] Health check..."
curl -sf "$BASE/health" | grep '"success":true'

# 2. Setup status — must be first_run=true
echo "[2/15] Setup status..."
curl -sf "$BASE/setup/status" | grep '"first_run":true'

# 3. First-run setup
echo "[3/15] First-run setup..."
ADMIN_TOKEN=$(curl -sf -X POST "$BASE/setup/first-run" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"testpass123","email":"admin@test.local"}' \
  | jq -r '.data.token')
[ -n "$ADMIN_TOKEN" ] && echo "  Admin token received"

# 4. Auth — login
echo "[4/15] Login..."
TOKEN=$(curl -sf -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"testpass123"}' \
  | jq -r '.data.access_token')
[ -n "$TOKEN" ] && echo "  JWT token received"

AUTH="Authorization: Bearer $TOKEN"

# 5. Auth — me
echo "[5/15] Get current user..."
curl -sf "$BASE/auth/me" -H "$AUTH" | grep '"username":"admin"'

# 6. Website — create
echo "[6/15] Create website..."
curl -sf -X POST "$BASE/websites" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"domain":"test.local","php_version":"8.2","web_server":"nginx"}' \
  | grep '"success":true'

# 7. Website — list
echo "[7/15] List websites..."
curl -sf "$BASE/websites" -H "$AUTH" | grep '"test.local"'

# 8. DNS — create record
echo "[8/15] Create DNS record..."
WEBSITE_ID=$(curl -sf "$BASE/websites" -H "$AUTH" | jq -r '.data[0].id')
curl -sf -X POST "$BASE/websites/$WEBSITE_ID/dns/records" \
  -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"type":"A","name":"test.local","value":"192.168.0.211"}' \
  | grep '"success":true'

# 9. Database — create
echo "[9/15] Create database..."
curl -sf -X POST "$BASE/databases" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"name":"testdb","engine":"mysql"}' | grep '"success":true'

# 10. Database — list
echo "[10/15] List databases..."
curl -sf "$BASE/databases" -H "$AUTH" | grep '"testdb"'

# 11. Email — create mailbox
echo "[11/15] Create mailbox..."
curl -sf -X POST "$BASE/email/mailboxes" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"email":"admin@test.local","password":"testpass123","quota":1000}' \
  | grep '"success":true'

# 12. Firewall — create rule
echo "[12/15] Create firewall rule..."
curl -sf -X POST "$BASE/firewall/rules" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"name":"Test","action":"allow","port":"3000","protocol":"tcp","source":"any"}' \
  | grep '"success":true'

# 13. Cron — create job
echo "[13/15] Create cron job..."
curl -sf -X POST "$BASE/cron" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"schedule":"* * * * *","command":"echo test","run_as":"root","type":"custom"}' \
  | grep '"success":true'

# 14. Settings — update
echo "[14/15] Update settings..."
curl -sf -X PUT "$BASE/settings" -H "$AUTH" -H "Content-Type: application/json" \
  -d '{"server_name":"Test Server"}' | grep '"success":true'

# 15. Metrics — current
echo "[15/15] Get current metrics..."
curl -sf "$BASE/metrics/current" -H "$AUTH" | grep '"cpu"'

echo ""
echo "=== ALL 15 TESTS PASSED ==="
```

### Run the Test Script
```bash
scp scripts/test-e2e.sh maruf@192.168.0.211:/tmp/
ssh maruf@192.168.0.211 "chmod +x /tmp/test-e2e.sh && bash /tmp/test-e2e.sh"
```

## Phase 4 — System-Level Verification

After API tests pass, verify the actual system changes:

```bash
# Check agent socket
ssh maruf@192.168.0.211 "ls -la /var/run/juvia/agent.sock"

# Check nginx config was generated
ssh maruf@192.168.0.211 "cat /etc/juvia/nginx/sites/test.local"

# Check php-fpm pool
ssh maruf@192.168.0.211 "cat /etc/juvia/php-fpm/pools/test.local"

# Check MySQL database exists
ssh maruf@192.168.0.211 "sudo mysql -e 'SHOW DATABASES;' | grep testdb"

# Check Postfix mailbox
ssh maruf@192.168.0.211 "sudo postmap -s /etc/postfix/virtual | grep admin@test.local"

# Check UFW rule
ssh maruf@192.168.0.211 "sudo ufw status | grep 3000"

# Check BIND9 zone
ssh maruf@192.168.0.211 "sudo cat /etc/juvia/bind/zones/db.test.local"

# Check logs
ssh maruf@192.168.0.211 "sudo tail -n 20 /var/log/juvia/panel.log"
```

## Phase 5 — Rollback Script

```bash
ssh maruf@192.168.0.211 "
  sudo systemctl stop juvia juvia-agent
  sudo dpkg -r juvia || true
  sudo rm -rf /var/lib/juvia /var/log/juvia /etc/juvia /var/run/juvia
  sudo userdel juvia 2>/dev/null || true
  echo 'Juvia removed'
"
```

## Success Criteria

| # | Check | How Verified |
|---|-------|-------------|
| 1 | `install.sh` runs end-to-end without error | Exit code 0 |
| 2 | Panel responds on port 18473 | HTTP 200 from `/api/v1/health` |
| 3 | First-run setup creates admin | JWT token returned |
| 4 | All 15 API tests pass | `test-e2e.sh` exit code 0 |
| 5 | Nginx config generated for website | File exists at `/etc/juvia/nginx/sites/test.local` |
| 6 | MySQL database created | `SHOW DATABASES` shows `testdb` |
| 7 | Postfix mailbox created | Virtual map has `admin@test.local` |
| 8 | BIND9 zone file updated | Zone file has A record |
| 9 | UFW rule created | `ufw status` shows port 3000 |
| 10 | Agent socket exists | `/var/run/juvia/agent.sock` present |
| 11 | Logs written | `/var/log/juvia/panel.log` has entries |
| 12 | Services restart clean | `systemctl restart` — no errors |
| 13 | Frontend served | `curl /` returns HTML |
| 14 | WebSocket endpoint responds | `/ws/v1/metrics` accepts WS upgrade |
| 15 | Panel port is 18473 (uncommon) | Confirmed in logs |

## Files to Create/Modify

| File | Action |
|------|--------|
| `scripts/test-e2e.sh` | New — automated API test suite |
| `scripts/install.sh` | Already installs all deps |
| `scripts/build-deb.sh` | Already builds .deb |

## Decision Needed

Should the test script run:
1. **On the remote server** (simpler, no SSH tunnel needed)
2. **From local machine** (proves remote accessibility, but needs port 18473 open)

Recommended: Run on remote server via SSH (`ssh maruf@192.168.0.211 "bash /tmp/test-e2e.sh"`). This tests the full stack without firewall complications.
