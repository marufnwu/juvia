# Module Implementation Notes

## Website Hosting

### Stack
- Nginx (default) or Apache (switchable per site)
- PHP-FPM pool per website (isolated Linux user)
- Certbot for SSL

### Key Operations
1. **Create website**: Create Linux user, document root, PHP-FPM pool, nginx config
2. **SSL issuance**: DNS check -> certbot -> HTTPS forced
3. **Change PHP version**: Update pool config, graceful reload (zero downtime)
4. **Suspend**: Swap vhost to branded "Site Suspended" page, stop PHP-FPM
5. **Delete (soft)**: Move to trash, 30-day retention

### OS Isolation
Each website gets its own Linux system user:
```bash
useradd -r -s /usr/sbin/nologin -d /home/mydomain -M mydomain
```
PHP-FPM pool runs under this user with chroot and open_basedir restrictions.

### Config Locations
- Nginx: `/etc/juvia/nginx/sites/`
- Apache: `/etc/juvia/apache/sites/`
- PHP-FPM: `/etc/juvia/php-fpm/pools/`

### File Structure
```
/home/mydomain/
  public_html/     # Web root
  logs/
    access.log
    error.log
  ssl/ # Symlinks to certbot certs
  tmp/
  .env              # Per-site environment vars
```

### Domain Structure
- Primary domain per website
- Subdomains (blog.mydomain.com)
- Addon domains (multiple unrelated domains on one server)

### Deployment Options
- ZIP upload (drag-and-drop, max 512MB)
- Git deploy (clone repo, webhook for auto-deploy with HMAC)
- FTP (future)

### Advanced Per-Site Features
- **Redirect manager**: Visual301/302 redirect table
- **Password protection**: nginx auth_basic
- **Hotlink protection**: nginx valid_referers
- **Custom error pages**: 404, 500, 403

### Website Detail Page
Full page navigation with sections:
- SSL Certificate (validity, expiry, renew button)
- Web Server (nginx/apache, PHP version, PHP-FPM)
- Caching (FastCGI, Redis, Memcached toggles)
- Logs (live tail, download)
- PHP Config (memory_limit, upload_max_filesize, etc.)

## Email Hosting

### Stack
- Postfix (MTA)
- Dovecot (IMAP/POP)
- Rspamd (spam filtering)
- Roundcube (optional webmail)

### Features
- Mailboxes with quotas
- Email forwarders
- Catch-all addresses per domain
- Aliases (multiple addresses -> one mailbox)
- Autoresponders (with start/end dates)

### Deliverability Checks
- SPF record: Authorizes sending servers
- DKIM record: Cryptographic email signature
- DMARC record: Policy for unauthorized emails
- PTR record: Reverse DNS (set by hosting provider)

### PTR Records
PTR records are set by the hosting provider, not the panel. Panel displays the correct value and directs users to their provider with instructions.

### Mail Deliverability Test Tool
1. Sends test email to check address
2. Reports SPF/DKIM/DMARC pass/fail
3. Checks if server IP is on blacklists (Spamhaus, Barracuda, etc.)
4. Shows results in plain English with fix instructions

## DNS Management

### Engine: BIND9

### Zone Auto-Creation
When website is created, auto-generate:
```
mydomain.com    A       -> server IP
www             CNAME   -> mydomain.com
mail            A       -> server IP
@               MX      -> mail.mydomain.com (priority 10)
@               TXT -> v=spf1 mx ~all
_dmarc          TXT      -> v=DMARC1; p=quarantine; rua=mailto:dmarc@mydomain.com
```

### Smart Suggestions
Panel detects missing records and offers auto-add:
- "No DKIM record - add it to improve email delivery"
- "No CAA record - add one to restrict SSL issuance"

### Validation
Run `named-checkzone` before any zone reload. Invalid zones never go live.

## Database Management

### Engines
- MySQL 8.0
- PostgreSQL 15

### Features
- Create/delete databases
- Create/delete users with granular permissions
- Remote access toggle (adds UFW rule automatically)
- Import .sql file (with progress + size limit)
- Export .sql download (one click)

### Visual DB Manager
Built-in replacement for phpMyAdmin:
- Browse tables, view rows with pagination
- Run SQL queries (Monaco editor)
- Table structure viewer
- Insert/edit/delete rows

## File Manager

### Security
All operations go through Agent with path safety:
```go
func safePath(base, userInput string) (string, error) {
    resolved := filepath.Clean(filepath.Join(base, userInput))
    if !strings.HasPrefix(resolved, base) {
        return "", ErrPathTraversal
    }
    return resolved, nil
}
```

### Restrictions
- Max 1GB extracted size for ZIP
- Max 10,000 entries per archive
- No absolute paths in archives
- Configurable upload limit (default 512MB)
- Uploaded files in public_html cannot be executed unless site is PHP/Node/Python

### Features
- Browse, upload, download, create, rename, delete, move, copy, chmod
- Extract zip/tar
- Monaco Editor for in-browser editing
- Drag-and-drop uploads
- Right-click context menu
- Hidden files toggle
- Bulk select

## Firewall

### Engine: UFW

### Default Rules
```
Allow 22    (SSH) - rate limited
Allow 80    (HTTP)
Allow 443   (HTTPS)
Allow 8080  (Panel)
Deny all others
```

### Features
- Open/close ports with plain-English names
- Allow/block IPs or CIDR ranges
- Rate limiting (brute force protection)
- Geo-blocking (block by country)

### UI
Never show raw UFW commands. Each rule has a plain-English description field.

## Backups & Restore

### Storage Options
- Local disk
- Amazon S3
- Cloudflare R2
- Backblaze B2
- Any S3-compatible endpoint
- SFTP

### Backup Types
- Full server
- Per-website (files + database bundled)
- Database only
- Files only

### Integrity Verification
After every backup:
1. SHA-256 checksum computed
2. Archive entry count verified
3. "Last verified" timestamp shown

### Restore Testing
Weekly dry-run restore test:
1. Extract to temp directory
2. Verify all files present
3. Delete temp
4. Show result in UI

### Scheduling
- Daily/weekly/monthly
- Custom interval (every X hours)
- Manual on-demand

### Retention
User sets limits (e.g., last 7 daily, last 4 weekly). Older backups auto-deleted.

## Cron Jobs

### Types
- Shell command
- URL hit (HTTP GET/POST)
- PHP script execution

### Visual Builder
```
Run every  [Day v]  at  [03:00 v]
[ ] Use cron expression
Expression: 0 3 * * *
```

### Execution History
- Last run time + duration
- Status (success/failed)
- Output log of last 10 runs
- Email alert on failure (toggle per job)

## Metrics & Monitoring

### Storage: VictoriaMetrics
Single Go binary, embedded. Better than SQLite for time-series.

### Metrics Tracked
- CPU usage
- RAM usage
- Disk usage
- Network in/out
- Per-website requests/sec
- Running processes

### Dashboard
Live via WebSocket every 2 seconds. Historical view via VictoriaMetrics query API.

### Configurable Alerts
- CPU > X% for Y minutes
- RAM > X%
- Disk > X%
- Website down or slow
- Unusual login

## Web Terminal

### Access Levels
- **Root terminal** — full access, prominent warning
- **Per-site terminal** — scoped to site Linux user

### Security
- Session recording to asciinema format
- Idle timeout (15 min, configurable)
- Re-authentication for root terminal
- Persistent "Root Terminal - all actions are recorded" banner
- IP restriction option for root terminal

## Alerts & Notifications

### Proactive Checks (every 60 seconds)
|Alert|Trigger|
|-----|--------|
|Website down|HTTP 5xx or timeout|
|Website slow|Response > 3s|
|SSL expiring|< 14 days|
|Disk > 80%|Usage threshold|
|Backup overdue|No backup in 7 days|
|Unusual login|New IP or odd hours|
|Service down|nginx/mysql/postfix etc.|

### Delivery
- In-panel bell icon
- Email notifications

## One-Click App Installs

### Available Apps
WordPress, WooCommerce, Laravel, Next.js, Ghost, Drupal, Joomla

### Install Flow
1. User picks app + existing site or new site
2. Enter admin credentials, database name
3. Plain-English progress steps
4. Done: links to site and admin panel

### Update Management
Background check daily. Badge shows "WordPress 6.8 available". One-click update with mandatory backup first.

## Settings & Panel Management

### Sections
- General (port, hostname, server name)
- Email (SMTP config)
- Security (2FA, sessions, login history, terminal IP restrictions)
- Updates (version, changelog, version pinning)
- Users (add/remove, roles)
- Backups (remote storage, retention)
- Notifications (alert thresholds)
- Audit Log (full history of every panel action)
- Export/Import (full config as JSON)

### Command Palette (Cmd+K)
Full action-capable search:
```
Cmd+K -> "add website"     -> opens Add Website form
Cmd+K -> "renew ssl"       -> shows SSL renewal for all sites
Cmd+K -> "open terminal"   -> opens terminal tab
```

## Module Dependencies

```
Website Hosting
    ├── DNS Management (BIND9) — zones created with website
    ├── SSL (certbot) — requires DNS to be pointed
    ├── Email Hosting — requires DNS for MX records
    └── Database Management — per-site databases

Email Hosting
    └── DNS Management — SPF/DKIM/DMARC records

File Manager
    └── Website Hosting — operates within site document root

Firewall
    └── Database Management — remote access toggle adds UFW rule

Cron Jobs
    └── Website Hosting — run as site user

Backups
    ├── Website Hosting — backs up site files
    ├── Database Management — backs up databases
    └── Email Hosting — backs up mailboxes (future)

Metrics & Monitoring
    └── Service Registry — monitors service health

Alerts
    └── Metrics, Service Registry, Website Hosting — triggers alerts
```
