# Plan: Make Juvia Install Robust for Real Users

## Problem Statement

The current install process is fragile:

1. **Strict `.deb` dependencies cause conflicts**
   - `Depends: mysql-server` conflicts with existing `mariadb-server`
   - `Depends: nginx, apache2` forces both web servers even if user only wants one
   - `apt-get install -f -y` may **remove** existing packages to satisfy conflicts

2. **`install.sh` doesn't handle real-world server states**
   - No handling for apt locks held by `unattended-upgrades`
   - No detection of existing services (Caddy, MariaDB, Nginx, Apache)
   - No retry logic for package installation
   - No verification that installed services actually work

3. **No firewall auto-configuration**
   - Panel port `18473` is not opened automatically
   - User must manually run `ufw allow 18473/tcp`

4. **Install verification is manual**
   - No automated check that panel responds after install
   - No automated check that agent socket exists

## Goal

Make `install.sh` work reliably on:
- Fresh Ubuntu 22.04/24.04 servers
- Servers with existing MariaDB/MySQL/PostgreSQL
- Servers with existing Nginx/Apache/Caddy
- Servers where apt is temporarily locked

## Proposed Changes

### 1. Fix `.deb` Dependencies (`scripts/build-deb.sh`)

Change from strict `Depends:` to a mix of `Depends:`, `Recommends:`, and `Suggests:`:

```
Depends: sqlite3, sudo, ufw, curl, wget, rsync, unzip, git, net-tools
Recommends: nginx | apache2, php-fpm, php-cli, php-mysql, php-pgsql, 
             php-curl, php-gd, php-mbstring, php-xml, php-zip, php-intl, 
             php-bcmath, php-soap, php-opcache, php-apcu, 
             mysql-server | mariadb-server, postgresql, postfix, 
             dovecot-imapd, dovecot-pop3d, rspamd, bind9, certbot, 
             python3-certbot-nginx | python3-certbot-apache, htop, tree
Suggests: python3-certbot-apache
```

**Why:**
- `Depends:` — only packages absolutely required for panel to run
- `Recommends:` — services the panel manages, but with alternatives (`|`)
- `Suggests:` — nice-to-have extras

### 2. Make `install.sh` Robust

#### 2.1 Wait for apt lock
```bash
wait_for_apt_lock() {
    local max_attempts=30
    local attempt=0
    while fuser /var/lib/apt/lists/lock /var/cache/apt/archives/lock /var/lib/dpkg/lock-frontend >/dev/null 2>&1; do
        attempt=$((attempt + 1))
        if [ $attempt -ge $max_attempts ]; then
            echo "Error: apt is locked by another process. Please wait and retry."
            exit 1
        fi
        echo "Waiting for apt lock... ($attempt/$max_attempts)"
        sleep 10
    done
}
```

#### 2.2 Detect existing services and skip conflicts
```bash
detect_existing_services() {
    [ -x "$(command -v mysqld)" ] || [ -x "$(command -v mariadbd)" ] && HAS_MYSQL=1
    [ -x "$(command -v postgres)" ] && HAS_POSTGRES=1
    [ -x "$(command -v nginx)" ] && HAS_NGINX=1
    [ -x "$(command -v apache2)" ] && HAS_APACHE=1
    [ -x "$(command -v caddy)" ] && HAS_CADDY=1
}
```

#### 2.3 Install only missing services
Install web server only if none exists:
```bash
if [ "$HAS_NGINX" != "1" ] && [ "$HAS_APACHE" != "1" ] && [ "$HAS_CADDY" != "1" ]; then
    apt-get install -y nginx
fi
```

Install database only if none exists:
```bash
if [ "$HAS_MYSQL" != "1" ]; then
    apt-get install -y mysql-server || apt-get install -y mariadb-server
fi
```

### 3. Auto-Configure Firewall

After installing Juvia, detect the panel port from settings or use default `18473`:
```bash
PANEL_PORT=$(sqlite3 /var/lib/juvia/juvia.db "SELECT value FROM settings WHERE key='panel_port'" 2>/dev/null || echo "18473")
ufw allow "$PANEL_PORT/tcp"
```

### 4. Post-Install Verification

Add a `verify_install()` function that checks:
```bash
verify_install() {
    local errors=0

    systemctl is-active --quiet juvia || { echo "ERROR: juvia service not running"; errors=$((errors+1)); }
    systemctl is-active --quiet juvia-agent || { echo "ERROR: juvia-agent service not running"; errors=$((errors+1)); }
    [ -S /var/run/juvia/agent.sock ] || { echo "ERROR: agent socket not found"; errors=$((errors+1)); }
    curl -sf "http://127.0.0.1:${PANEL_PORT}/api/v1/health" >/dev/null || { echo "ERROR: panel not responding on port ${PANEL_PORT}"; errors=$((errors+1)); }

    if [ $errors -eq 0 ]; then
        echo "SUCCESS: Juvia is installed and running on port ${PANEL_PORT}"
    else
        echo "FAILURE: ${errors} verification checks failed"
        exit 1
    fi
}
```

### 5. Handle Package Install Failures

If `dpkg -i` fails due to dependencies, run `apt-get install -f -y` **without** allowing removals:
```bash
dpkg -i juvia.deb || {
    echo "Installing missing dependencies..."
    apt-get install -f -y --no-remove
    dpkg -i juvia.deb
}
```

### 6. Add Uninstall Safety

Update `scripts/uninstall.sh` to:
- Stop services first
- Offer to keep or remove databases
- Offer to keep or remove websites
- Not remove system packages that other services may use

## Files to Modify

| File | Change |
|------|--------|
| `scripts/build-deb.sh` | Change `Depends:` to use `Depends:`/`Recommends:`/`Suggests:` with alternatives |
| `scripts/install.sh` | Add apt lock wait, existing service detection, firewall config, verification |
| `scripts/uninstall.sh` | Add safety prompts, don't remove shared system packages |
| `scripts/test-e2e.sh` | Add pre-install environment checks |

## Testing Strategy

| Test | Environment | Expected Result |
|------|-------------|-----------------|
| Fresh install | Clean Ubuntu 24.04 VM | All services installed, panel responds |
| MariaDB pre-installed | Ubuntu with `mariadb-server` | Juvia installs without removing MariaDB |
| Caddy pre-installed | Ubuntu with `caddy` | Juvia installs without forcing Nginx |
| apt locked | VM running `unattended-upgrades` | Installer waits up to 5 minutes then succeeds |
| Re-install | Existing Juvia install | Upgrades cleanly without data loss |

## Success Criteria

1. `install.sh` completes successfully on fresh Ubuntu without manual intervention
2. `install.sh` does not remove existing packages (MariaDB, Caddy, etc.)
3. Panel port is automatically opened in UFW
4. Post-install verification confirms panel responds on correct port
5. E2E test suite passes after automated install

## Implementation Order

0. **Clean the server** — Option 2 fresh slate cleanup on `maruf@192.168.0.211`
1. Fix `build-deb.sh` dependency declarations
2. Update `install.sh` with apt lock handling, service detection, firewall config, and verification
3. Update `uninstall.sh` with safety prompts
4. Build, tag, and release v0.1.3
5. Run `install.sh` on the cleaned server
6. Run `test-e2e.sh` and verify all 15 tests pass

## Server Cleanup (Prerequisite) — Option 2: Fresh Slate

The user selected **Option 2**: purge the server of all Juvia-related services and data before installing Juvia's own stack from scratch.

**Warning:** This will delete all websites, databases, emails, DNS zones, and configs managed by Juvia or its dependent services. Only run this on a server dedicated to Juvia.

1. Stop all Juvia-related services:
   ```bash
   sudo systemctl stop juvia juvia-agent nginx apache2 php*-fpm mysql mariadb postgresql postfix dovecot rspamd bind9 certbot 2>/dev/null || true
   ```

2. Remove the Juvia package and binaries:
   ```bash
   sudo dpkg -r juvia 2>/dev/null || true
   sudo dpkg --purge juvia 2>/dev/null || true
   sudo rm -f /usr/bin/juvia /usr/bin/juvia-agent
   sudo rm -f /etc/systemd/system/juvia.service /etc/systemd/system/juvia-agent.service
   sudo systemctl daemon-reload
   ```

3. Remove Juvia data and config directories:
   ```bash
   sudo rm -rf /var/lib/juvia /var/log/juvia /var/run/juvia /etc/juvia
   ```

4. Remove Juvia-managed system packages (fresh slate):
   ```bash
   sudo apt-get remove --purge -y nginx apache2 php-fpm php-cli php-mysql php-pgsql php-curl php-gd php-mbstring php-xml php-zip php-intl php-bcmath php-soap php-opcache php-apcu mysql-server mariadb-server postgresql postfix dovecot-imapd dovecot-pop3d rspamd bind9 certbot python3-certbot-nginx python3-certbot-apache 2>/dev/null || true
   sudo apt-get autoremove -y
   ```

5. Remove the `juvia` user and group:
   ```bash
   sudo userdel juvia 2>/dev/null || true
   sudo groupdel juvia 2>/dev/null || true
   ```

6. Clean up temporary files:
   ```bash
   sudo rm -f /tmp/juvia.deb /tmp/install.sh /tmp/test-e2e.sh
   ```

7. Clear stale apt locks:
   ```bash
   sudo rm -f /var/lib/apt/lists/lock /var/cache/apt/archives/lock /var/lib/dpkg/lock-frontend /var/lib/dpkg/lock
   sudo dpkg --configure -a
   sudo apt-get update
   ```

8. Reboot the server.

## Open Questions

1. Should the panel prefer `mariadb-server` over `mysql-server` on Ubuntu 24.04 (where MySQL is less common)?
2. Should we provide a `--minimal` install flag that skips all optional services and only installs the panel + agent?
3. Should the panel auto-detect the server's public IP and show that in the install completion message?
