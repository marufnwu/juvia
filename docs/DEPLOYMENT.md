# Deployment Guide

## Installation

### One-liner
```bash
curl -fsSL https://get.juvia.dev | bash
```

### Debian Package
```bash
apt install ./juvia_1.0.0_amd64.deb
```

Both methods:
1. Install dependencies (nginx, php-fpm, certbot, etc.)
2. Create `juvia` system user
3. Register two systemd services
4. Apply default firewall rules
5. Open port 8080

## Systemd Services

### juvia.service
```ini
[Unit]
Description=juvia HTTP Server
After=network.target

[Service]
Type=simple
User=juvia
Group=juvia
ExecStart=/usr/local/bin/juvia run
Restart=on-failure
RestartSec=5s
TimeoutStopSec=30s

[Install]
WantedBy=multi-user.target
```

### juvia-agent.service
```ini
[Unit]
Description=juvia Agent (privileged operations)
After=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/juvia agent run
Restart=on-failure
RestartSec=5s
TimeoutStopSec=30s

[Install]
WantedBy=multi-user.target
```

## Directory Structure

```
/etc/juvia/           # All generated configs
/var/lib/juvia/       # Data
  ├── juvia.db # SQLite database
  └── metrics/        # VictoriaMetrics data
/var/log/juvia/       # Logs
  ├── panel.log
  ├── agent.log
  └── terminal/       # Session recordings (.cast files)
/var/run/juvia/       # Runtime
  └── agent.sock      # Unix socket (660, root:juvia)
```

## First-Run Setup Wizard

On first access to port 8080, user completes 4-step wizard.

## Default Firewall Rules (UFW)

```
Status: active
---
To                         Action      From
---
22                         LIMIT       Anywhere
80                         ALLOW       Anywhere
443                        ALLOW       Anywhere
8080                       ALLOW       Anywhere
```

SSH rate-limited by default.

## User and Group Creation

```bash
# Create juvia user
useradd -r -s /usr/sbin/nologin -d /var/lib/juvia -M juvia

# Create juvia group (for socket access)
groupadd -f juvia
usermod -aG juvia juvia
```

## Self-Update Mechanism

1. Background goroutine checks GitHub Releases API daily
2. New version found banner in UI
3. User clicks Update:
   - Download new binary to temp location
   - Verify GPG signature
   - Backup current binary to `.backup`
   - Replace binary atomically
   - Restart both systemd services
   - Back online in ~10 seconds

### Rollback on Failure
If new binary fails to start:
1. Rename `.backup` back to active binary
2. Restart services
3. Alert user via email

## Building .deb Package

```bash
./scripts/build-deb.sh
```

Creates `juvia_1.0.0_amd64.deb` with:
- Binary (combined panel + agent)
- Systemd services
- User/group creation scripts
- Default firewall rules

## GPG Signing

Releases signed with GPG key. Verification steps:
1. Download release binary
2. Import maintainer GPG key
3. Verify signature

## Remote Server Deployment

Target: `maruf@192.168.0.211`

Development workflow:
1. Build binary locally
2. Copy to remote server via SSH/SCP
3. Install with `dpkg -i` or `systemctl restart`

```bash
# Build
go build -o juvia ./cmd/...

# Copy to remote
scp juvia maruf@192.168.0.211:/tmp/

# Install on remote
ssh maruf@192.168.0.211 "sudo dpkg -i /tmp/juvia"
```

## Port Configuration

Default: 8080 (HTTP)

User configures their own HTTPS reverse proxy if needed.

## Dependencies

Installed by installer script:
- nginx
- apache2 (switchable)
- php-fpm (8.1, 8.2, 8.3)
- certbot
- mysql-server or postgresql
- postfix
- dovecot
- rspamd
- bind9
- ufw
- victoria-metrics

## Monitoring Setup

Prometheus metrics endpoint: `GET /api/v1/metrics`

Grafana dashboard JSON provided in `scripts/grafana-dashboard.json`.

## Disaster Recovery

### Full System Backup
Backup `/var/lib/juvia/` and `/etc/juvia/` regularly.

### Restore Procedure
1. Reinstall juvia package
2. Restore SQLite database: `sqlite3 /var/lib/juvia/juvia.db < backup.sql`
3. Restore configs: `cp -r /backup/etc/juvia/* /etc/juvia/`
4. Restart services

## Uninstall

```bash
# Remove packages
dpkg -r juvia

# Remove user and group
userdel juvia
groupdel juvia

# Remove configs and data (optional)
rm -rf /etc/juvia
rm -rf /var/lib/juvia
rm -rf /var/log/juvia
```

## See Also

- [ARCHITECTURE.md](../docs/ARCHITECTURE.md) — System architecture
- [ENVIRONMENT.md](../docs/ENVIRONMENT.md) — Environment setup
