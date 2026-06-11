# Implementation Roadmap

## Overview

5 phases, building from foundation to full-featured control panel.

|Phase|Duration|Focus|
|-----|--------|-----|
|1 — Foundation|~4 weeks|Architecture, auth, dashboard, metrics, alerts|
|2 — Web Hosting|~4 weeks|Website management, SSL, DNS, trash|
|3 — Services|~4 weeks|Email, databases, file manager|
|4 — Ops|~3 weeks|Firewall, backups, cron, logs|
|5 — Polish|~3 weeks|App installs, git deploy, webmail, terminal recording, packaging|

Total: ~18 weeks

---

## Phase 1 — Foundation

**Goal**: Running panel, secure, useful from day one.

### Deliverables

- [ ] Two-process architecture (Panel + Agent)
- [ ] Systemd service registration
- [ ] JWT + refresh token authentication
- [ ] 2FA (TOTP) support
- [ ] First-run setup wizard
- [ ] Overview dashboard
- [ ] Service registry (nginx, mysql, postfix, etc.)
- [ ] Metrics collection (VictoriaMetrics)
- [ ] Proactive alert system
- [ ] Command palette (⌘K)

### Tasks

**Architecture**
- Set up Go project structure
- Implement Unix socket client (Panel side)
- Implement Unix socket server (Agent side)
- JSON-RPC protocol implementation
- Systemd service files

**Auth**
- User model and bcrypt hashing
- JWT access token generation (15 min)
- Refresh token in httpOnly cookie
- Session management in SQLite
- Login rate limiting (5 attempts/10 min)
- Auto-block after 20 failed attempts
- 2FA TOTP setup and verification
- Recovery codes generation

**First-Run Wizard**
- 4-step flow (welcome, server, notifications, 2FA)
- Settings storage in SQLite
- SMTP configuration for alerts

**Dashboard**
- Server health widgets (CPU, RAM, disk, network)
- Alert banner (critical warnings)
- Quick actions (add website, etc.)
- Recent activity feed
- Website list with status dots

**Service Registry**
- Detect installed services (nginx, apache, mysql, postfix, etc.)
- Version detection
- Health check every 60 seconds
- Setup prompts for missing services

**Metrics**
- VictoriaMetrics integration
- CPU, RAM, disk, network collection
- WebSocket streaming to frontend
- Historical query API

**Alerts**
- Alert engine (check every 60 seconds)
- Website down detection (HTTP 5xx or timeout)
- Website slow detection (>3s response)
- SSL expiring detection (<14 days)
- Disk space warnings (>80%)
- Unusual login detection
- Service down detection
- In-panel notification center

**Command Palette**
- Global search (⌘K)
- Action shortcuts (add website, renew ssl, etc.)
- Resource search (show all resources for domain)

### Dependencies
- Ubuntu 22.04+ or Debian 11+ server
- Go 1.21+
- SQLite

---

## Phase 2 — Web Hosting

**Goal**: Full website hosting with SSL and DNS management.

### Deliverables

- [ ] Website CRUD (create, read, update, suspend, delete)
- [ ] Nginx + Apache configuration
- [ ] PHP-FPM pool per site
- [ ] Linux user isolation per site
- [ ] SSL issuance (certbot)
- [ ] Auto-renewal
- [ ] DNS zone management (BIND9)
- [ ] Zone auto-creation on website creation
- [ ] Config validation before every reload
- [ ] Soft-delete trash (30-day recovery)
- [ ] Website detail page

### Tasks

**Website Management**
- Create website flow (user, document root, pool, config)
- Nginx site config generation
- Apache site config generation (alternative)
- PHP-FPM pool per site
- Linux user creation per site
- Website list with status (live, suspended, trash)
- Website detail page (SSL, web server, PHP version, caching)

**SSL**
- DNS check before issuance
- Certbot wrapper
- Auto-renewal cron registration
- SSL pending banner when DNS not ready
- Background retry every 10 minutes for 24 hours
- One-click renewal

**DNS**
- BIND9 zone file generation
- Zone auto-creation on website creation
- SOA and NS records
- A, AAAA, CNAME, MX, TXT, SRV, CAA record management
- named-checkzone validation
- Smart suggestions (add DKIM, CAA, etc.)

**Soft-Delete Trash**
- Deleted sites moved to trash (not destroyed)
- 30-day retention
- Restore with one click
- Permanent deletion after 30 days
- Final warning email before deletion

**Config Validation**
- nginx -t before reload
- apachectl configtest before reload
- named-checkzone before zone reload
- Error translation to plain English

### Dependencies
- Phase 1 complete
- nginx, apache, php-fpm installed

---

## Phase 3 — Services

**Goal**: Complete hosting stack — email, databases, file management.

### Deliverables

- [ ] Email mailboxes and forwarders
- [ ] Postfix + Dovecot configuration
- [ ] Rspamd spam filtering
- [ ] Deliverability checks (SPF, DKIM, DMARC, PTR)
- [ ] MySQL + PostgreSQL management
- [ ] Visual database manager
- [ ] File manager with Monaco Editor
- [ ] Path safety implementation

### Tasks

**Email**
- Mailbox creation (postfix + dovecot)
- Aliases and forwarders
- Catch-all address per domain
- Autoresponders
- SPF, DKIM, DMARC record generation
- Rspamd integration (per-domain spam threshold)
- Mailbox quota management
- Deliverability test tool (check SPF/DKIM, blacklist lookup)
- PTR record display with provider instructions

**Databases**
- MySQL database and user creation
- PostgreSQL database and user creation
- Permission management per database
- Remote access toggle (UFW rule)
- SQL file import (with progress)
- SQL export (one-click download)
- Visual DB manager (Monaco-based SQL editor, table browser)

**File Manager**
- Directory browsing with breadcrumb nav
- Upload (drag-and-drop, multi-file, folder)
- Download
- Create, rename, delete, move, copy
- chmod (permissions editor)
- ZIP/TAR extract
- Monaco Editor for file editing
- Path safety (safePath() algorithm)
- Zip bomb protection (size + entry limits)
- Hidden files toggle
- Bulk select

### Dependencies
- Phase 2 complete
- postfix, dovecot, rspamd, mysql, postgresql installed

---

## Phase 4 — Ops

**Goal**: Production-ready operations — firewall, backups, cron, logs.

### Deliverables

- [ ] Firewall management (UFW)
- [ ] Geo-blocking (country-based)
- [ ] Backup system (local + S3)
- [ ] Backup scheduling
- [ ] Backup integrity verification
- [ ] Cron job manager
- [ ] Log viewer (access + error logs)
- [ ] Email alert delivery

### Tasks

**Firewall**
- UFW wrapper
- Rule list (allow/deny, port, source)
- Open/close ports
- Allow/block IP or CIDR
- Rate limiting (brute force protection)
- Geo-blocking (country dropdown)
- Plain-English rule descriptions

**Backups**
- Backup types (full, website, database, files)
- Local storage
- Remote storage (S3, R2, B2, S3-compatible, SFTP)
- Backup scheduling (daily/weekly/monthly/custom)
- Retention policy
- SHA-256 integrity checksum
- Archive verification (entry count)
- Dry-run restore test (weekly)
- One-click restore

**Cron Jobs**
- Visual cron builder (every day at time, etc.)
- Cron expression toggle
- Job types (shell, URL, PHP)
- Run as user selection
- Execution history (last 10 runs)
- Status (success/failed)
- Output log display
- Email alert on failure

**Logs**
- Live log tail via WebSocket
- Access log viewer (filterable)
- Error log viewer (filterable)
- Log download
- Log rotation configuration

**Email Alerts**
- SMTP configuration
- Alert email delivery
- Per-alert-type toggle

### Dependencies
- Phase 3 complete

---

## Phase 5 — Polish

**Goal**: Public release ready — app installs, git deploy, webmail, packaging.

### Deliverables

- [ ] One-click app installs (WordPress, etc.)
- [ ] Git deployment + webhooks
- [ ] Webmail (Roundcube)
- [ ] Terminal session recording
- [ ] .deb package build
- [ ] Install script
- [ ] GPG-signed self-updater

### Tasks

**One-Click Apps**
- App catalog (WordPress, WooCommerce, Laravel, Next.js, Ghost, Drupal, Joomla)
- Install form (site name, admin credentials, database name)
- Plain-English progress steps
- Database auto-creation
- Admin URL generation
- Update available badge
- One-click update (with mandatory backup first)

**Git Deploy**
- Git clone to website
- Pull latest button
- Webhook URL generation (HMAC secured)
- Auto-deploy on webhook trigger

**Webmail**
- Roundcube installation
- Per-domain toggle
- Webmail access link

**Terminal Recording**
- asciinema-compatible recording format
- Session playback in UI
- Recording storage management

**Packaging**
- build-deb.sh script
- .deb package with:
  - Binary (combined)
  - Systemd services
  - User/group creation
  - Default firewall rules
  - Dependency installation
- Install script (get.juvia.dev)

**Self-Updater**
- GitHub Releases API check (daily)
- Update banner in UI
- GPG signature verification
- Binary replacement
- Service restart
- Version pinning option

### Dependencies
- Phases 1-4 complete

---

## Priority Order

If scope needs to be reduced, prioritize:

1. **Phase 1** (always required — foundation)
2. **Phase 2** (core hosting functionality)
3. **Phase 4** (ops essentials: backups, firewall, cron)
4. **Phase 3** (email, databases, file manager)
5. **Phase 5** (nice-to-have: app installs, webmail)

## Milestones

|Phase|Milestone|
|-----|---------|
|Phase 1 MVP|Running panel with dashboard and metrics|
|Phase 2 MVP|Working website with SSL|
|Phase 3 MVP|Databases and email working|
|Phase 4 MVP|Backups and firewall operational|
|Phase 5|Production release|

## Risk Assessment

### Phase 1 — Foundation
|Item|Risk|Mitigation|
|----|----|----------|
|Unix socket JSON-RPC|Protocol bugs can cause crashes|Write protocol tests first|
|JWT + refresh token security|Token leakage could compromise sessions|httpOnly cookies, short access token lifetime|
|2FA implementation|Buggy TOTP could lock users out|Recovery codes, test thoroughly|

### Phase 2 — Web Hosting
|Item|Risk|Mitigation|
|----|----|----------|
|Config generation (nginx/php-fpm)|Bad configs could take sites down|Atomic swap + nginx -t validation|
|SSL issuance timing|DNS not propagated, SSL fails|Background retry with clear UI messaging|
|PHP-FPM pool isolation|Vulnerability could cross-site|Strict chroot, separate users|

### Phase 3 — Services
|Item|Risk|Mitigation|
|----|----|----------|
|Email deliverability|IP reputation issues|Warn users, suggest external provider|
|Database import|Large SQL files could timeout|Chunked processing, progress bar|
|File manager path safety|Bypass could expose system files|Thorough safePath() testing|

### Phase 4 — Ops
|Item|Risk|Mitigation|
|----|----|----------|
|Backup to S3|Credential exposure|Use IAM roles, not access keys|
|UFW rule conflicts|Blocking SSH accidentally|Systemd pre-stop check, require confirmation for port 22|
|Cron job security|Malicious commands could be injected|Sanitize command input, run as limited user|

### Phase 5 — Polish
|Item|Risk|Mitigation|
|----|----|----------|
|GPG signature verification|Bypass could install malware|Sign releases with separate offline key|
|Self-update atomicity|Interrupted update could brick panel|Atomic binary replacement, rollback mechanism|

## See Also

- [ARCHITECTURE.md](../docs/ARCHITECTURE.md) — System architecture
- [TESTING.md](../docs/TESTING.md) — Testing strategy
- [SECURITY.md](../docs/SECURITY.md) — Security model