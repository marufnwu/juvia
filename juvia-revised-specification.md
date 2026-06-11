# juvia — Revised Complete Product Specification

> A free, open-source, modern server control panel for Ubuntu/Debian.  
> Built for non-technical small business owners. Inspired by Stripe Dashboard.  
> **Revision 2** — incorporates security hardening, architectural improvements, and UX refinements.

-----

## Table of Contents

1. [Vision & Philosophy](#1-vision--philosophy)
1. [Target User](#2-target-user)
1. [Core Architecture](#3-core-architecture)
1. [Tech Stack](#4-tech-stack)
1. [Installer & First-Run Setup](#5-installer--first-run-setup)
1. [Authentication & Users](#6-authentication--users)
1. [Service Registry](#7-service-registry)
1. [Module: Website Hosting](#8-module-website-hosting)
1. [Module: Email Hosting](#9-module-email-hosting)
1. [Module: DNS Management](#10-module-dns-management)
1. [Module: Database Management](#11-module-database-management)
1. [Module: File Manager](#12-module-file-manager)
1. [Module: Firewall](#13-module-firewall)
1. [Module: Backups & Restore](#14-module-backups--restore)
1. [Module: Cron Jobs](#15-module-cron-jobs)
1. [Module: Metrics & Monitoring](#16-module-metrics--monitoring)
1. [Module: Web Terminal](#17-module-web-terminal)
1. [Module: Alerts & Notifications](#18-module-alerts--notifications)
1. [Module: One-Click App Installs](#19-module-one-click-app-installs)
1. [Module: Settings & Panel Management](#20-module-settings--panel-management)
1. [UI Structure & Design System](#21-ui-structure--design-system)
1. [Build Phases](#22-build-phases)
1. [What Changed from v1 Spec](#23-what-changed-from-v1-spec)

-----

## 1. Vision & Philosophy

### What We’re Building

A **free, open-source server control panel** that feels like a modern SaaS product — not a tool from 2005. Built for non-technical small business owners who just want their website, email, and server to work — without needing to understand what nginx, SSL, or DNS means under the hood.

**The Stripe Dashboard of server management.**

### The Problem We’re Solving

|Old Way (cPanel / Plesk)                |juvia Way                         |
|----------------------------------------|----------------------------------------|
|Show everything, overwhelm the user     |Show what matters, hide what doesn’t    |
|Technical jargon everywhere             |Plain English + explanation side by side|
|User finds problems after they happen   |Panel warns before problems occur       |
|Click through 10 screens to do one thing|One-click fixes, smart action nudges    |
|Ugly icon grids, dense forms            |Clean sidebar, Stripe-style layout      |
|Manual, reactive management             |Proactive, intelligent suggestions      |
|Security as an afterthought             |Security-first architecture             |

### Language Design Rule

Every technical term in the UI must follow this pattern:

> **PHP Version** · *The programming language your website’s code runs on*  
> **DKIM Record** · *A security signature that proves your emails are genuine*  
> **Cron Job** · *A task that runs automatically on a schedule*

Plain English for progress steps too — never show *“Creating PHP-FPM pool”*, always show *“Setting up PHP for your website…”*

-----

## 2. Target User

**Primary user:** A non-technical small business owner running a VPS or dedicated server on Ubuntu/Debian. They want their website and email running reliably. They get anxious when they see a terminal. They don’t know what PHP-FPM is and shouldn’t need to.

**Supported servers:** VPS and dedicated servers running Ubuntu 22.04+ or Debian 11+.

**Business model:** Free and open-source. Self-hosted, no license fee. GPG-signed releases via GitHub.

-----

## 3. Core Architecture

### Process Model (Revised)

The original spec ran everything in a single process with root privileges. This is a critical security risk — an RCE vulnerability in the HTTP layer (which handles file uploads, runs a file manager, serves user content) would give an attacker full root access.

**Revised model: Two separate processes.**

```
┌─────────────────────────────────────────────────────────────┐
│  PROCESS 1: Panel Server (runs as unprivileged user: panel) │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  HTTP Server (port 8080)                             │   │
│  │  React UI (embedded via Go embed)                    │   │
│  │  REST API handlers                                   │   │
│  │  WebSocket (metrics stream, terminal proxy)          │   │
│  │  Auth / JWT + refresh tokens                         │   │
│  │  Task manager (tracks job status, streams progress)  │   │
│  └──────────────────────┬───────────────────────────────┘   │
│                         │ Unix socket (chmod 600)            │
│                         │ /var/run/juvia/agent.sock    │
└─────────────────────────┼───────────────────────────────────┘
                          │ Go interface calls (no gRPC)
┌─────────────────────────┼───────────────────────────────────┐
│  PROCESS 2: Agent (runs as root via systemd)                │
│  ┌──────────────────────▼───────────────────────────────┐   │
│  │  Unix socket server                                  │   │
│  │  nginx / apache config management                    │   │
│  │  SSL (certbot wrapper)                               │   │
│  │  DNS (BIND9 zone management)                         │   │
│  │  Email (Postfix + Dovecot + Rspamd)                  │   │
│  │  Database (MySQL + PostgreSQL ops)                   │   │
│  │  Firewall (UFW wrapper)                              │   │
│  │  Backup (tar + rsync + S3)                           │   │
│  │  Cron management                                     │   │
│  │  File operations (path-safe)                         │   │
│  │  System metrics reader (/proc)                       │   │
│  │  Terminal PTY spawner                                │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
          │                              │
          ▼                              ▼
       SQLite (relational)       VictoriaMetrics (metrics)
   /var/lib/juvia/         /var/lib/juvia/metrics/
   panel.db                      (single binary, embedded)
```

Both processes ship in one binary and are registered as two separate systemd services:

- `juvia.service` — HTTP panel, runs as user `panel`
- `juvia-agent.service` — privileged agent, runs as root

The installer sets up both. Users interact with only the panel service.

### Internal Communication: Go Interface over Unix Socket

No gRPC. The panel and agent communicate via a **simple JSON-RPC protocol over a Unix socket**. This is fast, typed, and has zero external dependencies. The socket is owned by root, group `panel`, mode `660` — only the panel process can talk to the agent.

When multi-server support is added in v2, this interface can be promoted to a proper gRPC/TLS channel over the network. The interface contract stays the same — only the transport changes.

### Task / Job System

Every non-trivial action becomes an async **task** with a unique ID. This is required from day one — long operations (backup, SSL issuance, app install, git clone) must never block the HTTP response.

```
User clicks "Add Website"
      │
      ▼
Panel creates task: { id: "task_abc123", type: "create_website", status: "running" }
Returns HTTP 202 Accepted + task ID immediately
      │
      ▼
Agent executes steps, updating task status in SQLite after each step
      │
      ▼
Panel streams task progress to browser via WebSocket
      │
      ▼
Task completes → status: "done" → toast notification fires
```

Task history is stored in SQLite. Users can review past task results including any error output.

### Config File Management Strategy

The agent **never edits distro-managed config files directly**. All generated configs live under `/etc/juvia/` and are included by the main service configs.

```
/etc/juvia/
├── nginx/
│   ├── sites/          ← one file per website
│   └── includes/       ← shared snippets
├── apache/
│   └── sites/
├── php-fpm/
│   └── pools/          ← one pool config per website
├── postfix/
├── dovecot/
├── bind/
│   └── zones/
└── ufw/
```

**Write strategy — atomic swap:**

```
1. Write new config to /etc/juvia/nginx/sites/mydomain.com.tmp
2. Validate: nginx -t (or equivalent)
3. If valid:   rename() tmp → mydomain.com  (atomic on Linux)
               reload service
4. If invalid: delete tmp file
               return error with exact validation output
               translate to plain English for the UI
```

This ensures a bad config never replaces a working one and never takes sites down.

### Data Storage

|Data                                  |Storage                                       |Why                                                  |
|--------------------------------------|----------------------------------------------|-----------------------------------------------------|
|Users, configs, jobs, tasks, audit log|SQLite (`panel.db`)                           |Zero config, relational, embedded                    |
|Metrics time-series                   |VictoriaMetrics (single binary)               |Purpose-built, handles millions of points efficiently|
|System and service logs               |Plain files (`/var/log/juvia/`)         |Simple, inspectable, greppable                       |
|Terminal session recordings           |Plain files (`/var/log/juvia/terminal/`)|Append-only, replay-friendly                         |

### Project Structure

```
juvia/
├── cmd/
│   ├── panel/                  # HTTP panel entry point
│   └── agent/                  # Privileged agent entry point
├── internal/
│   ├── agent/                  # Agent implementation
│   │   ├── nginx/
│   │   ├── apache/
│   │   ├── ssl/
│   │   ├── dns/
│   │   ├── email/
│   │   ├── database/
│   │   ├── firewall/
│   │   ├── backup/
│   │   ├── cron/
│   │   ├── files/
│   │   └── system/
│   ├── api/                    # REST API handlers
│   ├── ws/                     # WebSocket handlers
│   ├── tasks/                  # Task runner + status tracking
│   ├── socket/                 # Unix socket client/server
│   ├── db/                     # SQLite models + migrations
│   ├── auth/                   # JWT + refresh tokens + 2FA
│   ├── alerts/                 # Proactive monitoring engine
│   ├── registry/               # Service registry
│   └── updater/                # Self-update + GPG verification
├── web/                        # React frontend (embedded)
└── scripts/
    ├── install.sh
    └── build-deb.sh
```

-----

## 4. Tech Stack

|Layer             |Choice                                        |Reason                                                       |
|------------------|----------------------------------------------|-------------------------------------------------------------|
|Language          |Go                                            |Single binary, low memory, great concurrent shell execution  |
|Frontend          |React + Vite + Tailwind + shadcn/ui           |Modern, fast, customisable                                   |
|Frontend ↔ Backend|REST + WebSocket                              |REST for actions, WebSocket for tasks and real-time data     |
|Panel ↔ Agent     |JSON-RPC over Unix socket                     |Simple, fast, no dependencies, upgradeable to gRPC in v2     |
|Relational DB     |SQLite (embedded)                             |Zero config, file-based                                      |
|Metrics DB        |VictoriaMetrics (single binary)               |Purpose-built time-series, far better than SQLite for metrics|
|Web server        |Nginx (default) + Apache (switchable per site)|Industry standard                                            |
|SSL               |certbot / Let’s Encrypt                       |Free, auto-renewal                                           |
|Spam filter       |Rspamd                                        |Modern, faster than SpamAssassin, better accuracy            |
|Panel access      |HTTP on port 8080                             |User configures HTTPS themselves                             |
|Updates           |In-panel one-click, GPG-signed binaries       |Secure supply chain                                          |

-----

## 5. Installer & First-Run Setup

### Installation

**One-liner:**

```bash
curl -fsSL https://get.juvia.dev | bash
```

**Debian package:**

```bash
apt install ./juvia_1.0.0_amd64.deb
```

Both: install dependencies, create `panel` system user, register two systemd services, apply default firewall rules, open port 8080.

### First-Run Setup Wizard

On first login, the user is shown a mandatory setup wizard before reaching the dashboard. This is a critical UX moment — non-technical users landing on a blank dashboard with no guidance will not know what to do.

```
Step 1 of 4 — Welcome
─────────────────────────────────────────────────────
  Welcome to juvia.

  Let's get your server set up in a few quick steps.
  This should take about 2 minutes.

                                      [Get Started →]

─────────────────────────────────────────────────────
Step 2 of 4 — Your Server
─────────────────────────────────────────────────────
  Server name (shown in the panel)
  ┌──────────────────────────────┐
  │ My Production Server         │
  └──────────────────────────────┘

  Server hostname / IP
  ┌──────────────────────────────┐
  │ 192.168.1.1                  │
  └──────────────────────────────┘
  · This is the IP address your websites will point to

                          [Back]  [Continue →]

─────────────────────────────────────────────────────
Step 3 of 4 — Notifications
─────────────────────────────────────────────────────
  Where should we send alerts?
  (SSL expiry, site downtime, disk space warnings)

  Email address
  ┌──────────────────────────────┐
  │ john@mybusiness.com          │
  └──────────────────────────────┘

  SMTP settings (to send alert emails)
  ┌──────────────────────────────┐
  │ smtp.gmail.com               │  Host
  │ 587                          │  Port
  │ john@gmail.com               │  Username
  │ ••••••••••••••               │  Password
  └──────────────────────────────┘

  [ ] Skip for now — I'll configure this later

                          [Back]  [Continue →]

─────────────────────────────────────────────────────
Step 4 of 4 — Secure Your Panel
─────────────────────────────────────────────────────
  Two-factor authentication
  · Strongly recommended — your panel has full control
    over your server.

  [ ] Enable two-factor authentication (recommended)
      Scan this QR code with Google Authenticator or Authy

  [ ] Skip for now (you can enable this later in Settings)

                          [Back]  [Finish Setup →]
─────────────────────────────────────────────────────
```

After completing the wizard, the user lands on the dashboard with a contextual nudge: *“Add your first website to get started →”*

### Self-Update Flow

1. Background goroutine checks GitHub Releases API daily
1. New version found → banner: *“juvia v1.2.0 is available — what’s new · Update now”*
1. User clicks Update → panel downloads binary, **verifies GPG signature**
1. If signature valid → replaces binary, restarts both systemd services
1. Back online in ~10 seconds
1. If signature invalid → aborts, alerts user, logs the attempt

Users can pin to a specific version in Settings.

-----

## 6. Authentication & Users

### Login

Username + password. Sessions use **short-lived JWT access tokens (15 minutes)** + **long-lived refresh tokens stored in httpOnly cookies (7 days)**. This prevents token theft without constant logouts.

### Two-Factor Authentication (2FA)

TOTP-based (Google Authenticator, Authy, 1Password). Configurable per user. Strongly encouraged during first-run setup. Recovery codes generated on setup (download and store safely).

### Login Security

- **Rate limiting:** 5 failed attempts per IP per 10 minutes
- **Auto-block:** After 20 failed attempts, IP is added to UFW deny list
- **Login alerts:** Notify user on login from a new IP address
- **Session list:** Users can see active sessions and revoke them

### Roles

|Role         |Access                                    |
|-------------|------------------------------------------|
|**Admin**    |Full access to everything                 |
|**Read-only**|Can view all sections, cannot make changes|
|**Per-site** |Access scoped to specific websites only   |

Per-site role is useful for giving a client access to just their website without exposing the rest of the server.

-----

## 7. Service Registry

The panel maintains a **service registry** — a record of which services are installed, their version, and their health status. This solves the problem of the panel not knowing what’s actually available on the server.

```
Service Registry
─────────────────────────────────────────────────────────
Nginx          ✓ Installed   v1.24.0    ● Running
Apache         ✗ Not installed                    [Install]
MySQL          ✓ Installed   v8.0.35    ● Running
PostgreSQL     ✓ Installed   v15.4      ● Running
Postfix        ✓ Installed   v3.7.4     ● Running
Dovecot        ✓ Installed   v2.3.20    ● Running
Rspamd         ✓ Installed   v3.6       ● Running
BIND9          ✓ Installed   v9.18      ● Running
Certbot        ✓ Installed   v2.7.0     —
Roundcube      ✗ Not installed                    [Install]
VictoriaMetrics ✓ Installed  v1.95      ● Running
─────────────────────────────────────────────────────────
```

Features that depend on an uninstalled service show a setup prompt rather than a broken UI. Example: if MySQL is not installed, the Databases section shows: *“MySQL is not set up yet — install it to start managing databases.”* with an Install button.

The registry checks service health every 60 seconds and surfaces issues in the alert system.

-----

## 8. Module: Website Hosting

### Adding a Website

User types a domain name and clicks **Add Website**. The panel does everything automatically, showing plain-English progress steps.

```
User types "mydomain.com" → clicks "Add Website"
         │
         ▼
Panel checks: is DNS pointing to this server?
         │
    ┌────┴────┐
   YES        NO
    │          └─► "DNS not pointed yet — your site will be
    │               created now and SSL will be issued
    │               automatically once DNS is ready.
    │               Point your domain to: 203.0.113.1 ▶"
    ▼
Task created → progress streamed via WebSocket:

  ✓ Creating a secure space for your website...
  ✓ Configuring your web server...
  ✓ Setting up PHP for your website...
  ✓ Securing your website with HTTPS...
  ✓ Your website is live!

         │
         ▼
User lands on Website Detail Page ✓
```

### Supported Site Types

|Type                    |How It Runs                       |Nginx Role                |
|------------------------|----------------------------------|--------------------------|
|Static HTML             |Files served directly             |Nginx serves files        |
|PHP (WordPress, Laravel)|PHP-FPM pool per site             |Nginx → PHP-FPM socket    |
|Node.js                 |systemd service per site          |Nginx reverse proxy → port|
|Python (Django, Flask)  |systemd service (gunicorn/uvicorn)|Nginx reverse proxy → port|

### File Structure

```
/home/mydomain/
├── public_html/        ← web root
├── logs/
│   ├── access.log
│   └── error.log
├── ssl/                ← symlinks to certbot certs
├── tmp/
└── .env                ← per-site environment vars
```

### OS Isolation

Each website gets its own **Linux system user** with an isolated **PHP-FPM pool**. A compromised site cannot read another site’s files.

### Domain Structure

- Primary domain per website
- Subdomains (blog.mydomain.com)
- Addon domains (multiple unrelated domains on one server)

### PHP Management

```
Per-site PHP config:
├── PHP Version          8.1 / 8.2 / 8.3 (switchable, zero downtime)
├── memory_limit         default: 256MB
├── upload_max_filesize  default: 64MB
├── post_max_size        default: 64MB
├── max_execution_time   default: 60s
├── display_errors       default: Off
└── OPcache              default: On
```

### SSL

```
DNS check → passes?
  YES → certbot issues SSL → HTTPS forced → auto-renewal cron registered
        Config validated before reload (nginx -t)
  NO  → Site created without SSL
        Banner: "SSL pending — will issue automatically once DNS is ready"
        Background checker retries every 10 mins for 24hrs
        On success → auto-issues SSL → fires alert "SSL issued ✓"
```

HTTPS forced by default. User can disable with warning tooltip.

### Caching Stack

|Layer              |What It Does                   |Best For               |
|-------------------|-------------------------------|-----------------------|
|Nginx FastCGI Cache|Full page cache at server level|WordPress, any PHP site|
|Redis Object Cache |DB queries and app objects     |WordPress, Laravel     |
|Memcached          |Alternative object caching     |Legacy PHP apps        |
|Built-in Page Cache|Panel-managed static cache     |Simple sites           |

### Deployment Options

**ZIP Upload:** Drag and drop. Panel extracts, sets ownership/permissions, validates no path traversal.

**Git Deploy:**

```
User pastes: https://github.com/user/mysite.git
Panel: clones repo, stores URL, shows "Pull latest" button
Optional: auto-deploy webhook URL generated (secured with HMAC signature)
```

### Advanced Per-Site Features

- Redirect manager (301/302 visual table, no .htaccess)
- Password-protected directories (nginx auth_basic)
- Hotlink protection (nginx valid_referers)
- Custom error pages (404, 500, 403)

### Logs

- **Live tail** — WebSocket stream in xterm.js, filterable by level
- **Download** — one-click download of access.log or error.log

### Suspension & Soft Delete (Revised)

**Suspend:** Vhost swapped to branded “Site Suspended” page. Files untouched. PHP-FPM pool stopped. Instantly reversible.

**Delete:** Sites are **soft-deleted** — moved to trash, not immediately destroyed.

```
Trash system:
- Deleted sites kept for 30 days
- Files, database, and config all preserved
- Site is taken offline immediately
- Restore available with one click during retention period
- After 30 days: permanent deletion with final warning email
```

This prevents catastrophic accidents for non-technical users.

### Config Validation

Before every nginx/Apache reload:

```
nginx -t → output parsed
  Valid   → atomic rename() → reload service
  Invalid → temp file deleted
             error shown: "Web server config has an error"
             plain-English translation of the nginx error
             no site goes down
```

-----

## 9. Module: Email Hosting

### Stack: Postfix + Dovecot + Rspamd

> **Important caveat:** Self-hosted email is the most complex module in this panel. IP reputation, blacklisting, and deliverability issues are genuinely hard to diagnose. Email should be treated as a v2 feature or offered alongside a “use an external provider” option (Mailgun, Postmark, Resend) for users who want reliability without the complexity.

### Features

- Mailboxes ([user@domain.com](mailto:user@domain.com))
- Email forwarders (forward to any address)
- Catch-all address per domain
- Aliases (multiple addresses → one mailbox)
- Autoresponders (out-of-office, custom replies)
- Webmail via Roundcube (optional, toggle per domain)

### Deliverability Health Card

```
Email Deliverability — mydomain.com
──────────────────────────────────────────────────────────
✓ SPF      Your server is authorised to send email
✓ DKIM     Emails are cryptographically signed
✓ DMARC    Unauthorised emails will be rejected
⚠ PTR      Reverse DNS does not match your server IP
           → This must be set with your hosting provider.
             Your PTR record should be: mail.mydomain.com
             [Copy value]  [Learn how to set this ↗]
──────────────────────────────────────────────────────────
```

**PTR records are set by the hosting provider, not the server.** The panel can check and display the correct value, but never implies it can set this automatically. The UI directs users to their hosting provider with the exact value to enter.

### Spam Filtering: Rspamd

Per-domain. User sees a threshold slider: Strict / Balanced / Relaxed. No raw config exposed.

### Mailbox Storage

- Usage bar per mailbox
- Per-mailbox quota configurable
- Alert at 80% usage

### Mail Deliverability Test Tool

A built-in tool (accessible from the Email section) that:

1. Sends a test email to a test address
1. Reports SPF pass/fail, DKIM pass/fail, spam score
1. Checks if the server IP is on common blacklists (Spamhaus, Barracuda, etc.)
1. Shows results in plain English with fix instructions

-----

## 10. Module: DNS Management

### Engine: Self-Hosted BIND9

### Zone Auto-Creation

When a website is added, a DNS zone is auto-created:

```
mydomain.com    A       → server IP
www             CNAME   → mydomain.com
mail            A       → server IP
@               MX      → mail.mydomain.com (priority 10)
@               TXT     → SPF record (auto-generated)
_dmarc          TXT     → DMARC policy (auto-generated)
```

### Supported Record Types

A, AAAA, CNAME, MX, TXT, NS, SRV, CAA — all manageable via a clean records table.

### Smart Suggestions

Panel detects missing records and surfaces nudges:

- *“No DKIM record — add it to improve email delivery”* → [Add automatically]
- *“No CAA record — add one to restrict SSL issuance”* → [Add automatically]

Panel writes the correct value on click.

### Config Validation

Every zone change runs `named-checkzone` before reload. Invalid zones never go live.

-----

## 11. Module: Database Management

### Engines: MySQL + PostgreSQL

### Features

- Create / delete databases
- Create / delete database users
- Assign granular permissions per database
- Remote access toggle (adds UFW rule automatically)
- Import via .sql file upload (with progress + size limit)
- Export as .sql download (one click)

### Built-In Visual DB Manager

Modern replacement for phpMyAdmin — no separate install.

- Browse tables, view rows with pagination
- Run SQL queries (Monaco editor)
- Table structure viewer
- Insert / edit / delete rows

Accessible via **Manage →** button per database.

-----

## 12. Module: File Manager

### Security Model (Revised)

The file manager is the highest-risk surface in the panel. All file operations go through the agent which enforces strict path safety:

```go
// All paths resolved and validated before any operation
func safePath(base, userInput string) (string, error) {
    resolved := filepath.Clean(filepath.Join(base, userInput))
    if !strings.HasPrefix(resolved, base) {
        return "", ErrPathTraversal  // blocks ../../etc/passwd attacks
    }
    return resolved, nil
}
```

Additional restrictions:

- **Zip extraction:** max 1GB extracted size, max 10,000 entries, no absolute paths in archive
- **Upload limits:** configurable max file size (default 512MB)
- **Executable protection:** uploaded files in public_html cannot be executed by the web server unless explicitly a PHP/Node/Python site
- **CSP headers:** panel itself served with strict Content-Security-Policy

### Operations

Browse, upload, download, create, rename, delete, move, copy, chmod, extract zip/tar — all supported.

### In-Browser Editor

**Monaco Editor** (VS Code engine) — full syntax highlighting. Opens in a full-screen overlay.

### Uploads

- Drag and drop onto any folder
- Multi-file selection
- Full folder upload
- Progress bar per file

### UX Details

- Breadcrumb navigation
- Right-click context menu
- Hidden files toggle
- Bulk select with checkboxes

-----

## 13. Module: Firewall

### Engine: UFW (Ubuntu-native)

### Default Rules on Install

```
Allow 22    (SSH)
Allow 80    (HTTP)
Allow 443   (HTTPS)
Allow 8080  (Panel)
Deny all others (default)
```

### Features

- Open / close ports (with plain-English service names)
- Allow / block specific IPs or CIDR ranges
- Rate limiting — brute force protection (auto-configured on install)
- Geo-blocking — block countries via dropdown

### UI

```
Rule              Type    Port    Source        Status
─────────────────────────────────────────────────────
SSH               Allow   22      Anywhere      ● Active
Web (HTTP)        Allow   80      Anywhere      ● Active
Web (HTTPS)       Allow   443     Anywhere      ● Active
Panel             Allow   8080    Anywhere      ● Active
Block Russia      Deny    All     RU            ● Active
```

Rules are never shown as raw UFW commands. Each has a plain-English description field.

-----

## 14. Module: Backups & Restore

### Storage

Both local disk and remote simultaneously.

**Remote providers:** Amazon S3, Cloudflare R2, Backblaze B2, any S3-compatible endpoint, SFTP.

### What Gets Backed Up

- Full server
- Per-website (files + database bundled)
- Database only
- Files only

### Scheduling

- Daily / weekly / monthly
- Custom interval (every X hours)
- Manual on-demand at any time

### Restore Granularity

- Entire website to a point in time
- Individual files browsed from a backup
- Database only

### Retention Policy

User sets limits (e.g. last 7 daily, last 4 weekly). Older backups auto-deleted.

### Backup Integrity Verification (New)

After every backup:

1. Checksum (SHA-256) computed and stored
1. Archive opened and entry count verified (not corrupted)
1. “Last verified” timestamp shown per backup

Periodic dry-run restore test (configurable, e.g. weekly) — extracts the backup to a temp directory, verifies all files present, deletes temp. Result shown in UI.

```
mydomain.com    Jun 11 03:00    1.2GB    Local + S3    ✓ Verified    [Restore]
mydomain.com    Jun 10 03:00    1.1GB    Local + S3    ✓ Verified    [Restore]
mydomain.com    Jun 9 03:00     1.1GB    S3 only       ⚠ Unverified  [Restore]
```

-----

## 15. Module: Cron Jobs

### Job Types

- Shell command
- URL hit (HTTP GET/POST)
- PHP script execution

### Creator UI

Visual builder with raw expression toggle:

```
Run every  [Day ▾]  at  [03:00 ▾]

[ ] Use cron expression
Expression: 0 3 * * *   ← auto-synced

Action: [ Run command ▾ ]
Command: php /home/mydomain/wp-cron.php
Run as user: mydomain
```

### Execution History

- Last run time + duration
- Status (✓ Success / ✗ Failed)
- Output log of last 10 runs
- Email alert on failure (toggle per job)

-----

## 16. Module: Metrics & Monitoring

### What’s Tracked

CPU, RAM, disk, network in/out, per-website requests/sec, running processes.

### Storage: VictoriaMetrics (Revised)

Replaces SQLite for metrics storage. VictoriaMetrics is a single Go binary, uses far less disk than SQLite for time-series data, handles millions of data points efficiently, and has a built-in query API. Metrics are retained for 30 days. Sampled every 15 seconds (more granular than the original 60s).

### Dashboard

```
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐
│ CPU      │ │ RAM      │ │ Disk     │ │ Network      │
│ 23%      │ │ 1.2/4GB  │ │ 42/80GB  │ │ ↑ 2.1 MB/s   │
│ ▁▂▃▂▁▂  │ │ ▄▄▄▄▄▄  │ │ ████░░░  │ │ ↓ 8.4 MB/s   │
└──────────┘ └──────────┘ └──────────┘ └──────────────┘
```

Live via WebSocket every 2 seconds. Historical view via VictoriaMetrics query API.

### Configurable Alert Thresholds

- CPU > X% for Y minutes
- RAM > X%
- Disk > X%
- Website down or slow
- Unusual login

-----

## 17. Module: Web Terminal

### Access Levels

- **Root terminal** — full access, prominent warning label
- **Per-site terminal** — scoped to that site’s Linux user

### UI

Multiple tabs simultaneously, each an independent PTY session via WebSocket + xterm.js.

### Security (Revised)

- **Session recording:** Full input + output recorded to `/var/log/juvia/terminal/` in a replayable format. Each session stored as `YYYY-MM-DD_HH-MM-SS_user_sessionid.cast` (asciinema-compatible format).
- **Idle timeout:** Auto-close after 15 minutes of inactivity (configurable in Settings).
- **Re-authentication:** Opening a root terminal requires re-entering the panel password.
- **Audit label:** Every root terminal tab shows a persistent banner: *“Root Terminal — all actions are recorded”*
- **IP restriction:** Root terminal can be restricted to specific IPs in Settings.

-----

## 18. Module: Alerts & Notifications

### Delivery Channels

- In-panel bell icon (notification centre)
- Email notifications

All alerts go to one unified place.

### Alert Centre

```
🔴  mydomain.com is not responding          2 min ago   [Investigate]
🟡  SSL expires in 9 days — mydomain.com    1 hour ago  [Renew Now]
✓   Disk space back to normal               3 hours ago  Resolved
🔴  Unusual login from 185.23.x.x           Yesterday   [Review]
```

### Timing

Alerts fire immediately. 2-minute grace period prevents false alarms for transient blips.

### Proactive Checks (every 60 seconds)

|Check         |Trigger                 |Shown As          |
|--------------|------------------------|------------------|
|Website down  |HTTP 5xx or timeout     |🔴 Not responding  |
|Website slow  |Response > 3s           |🟡 Running slow    |
|SSL expiring  |< 14 days               |🟡 Expires soon    |
|SSL expired   |Past expiry             |🔴 Expired         |
|Disk > 80%    |Usage threshold         |🟡 Disk filling up |
|Backup overdue|No backup in 7 days     |🟡 Backup overdue  |
|Backup failed |Last backup errored     |🔴 Backup failed   |
|Unusual login |New IP or odd hours     |🔴 Suspicious login|
|Service down  |nginx/mysql/postfix etc.|🔴 Service stopped |

-----

## 19. Module: One-Click App Installs

### Apps in v1

WordPress, WooCommerce, Laravel, Next.js, Ghost, Drupal, Joomla.

### Install Flow

```
User picks app → picks existing site or creates new one
      │
      ▼
Install form (plain English labels):
  - Your site's name
  - Admin username + password
  - Database name (auto-suggested, editable)
      │
      ▼
Plain-English progress steps:
  ✓ Creating your database...
  ✓ Downloading the latest version of WordPress...
  ✓ Installing WordPress on your website...
  ✓ WordPress is ready!
  [Open your site ↗]  [Go to WordPress admin ↗]
```

### Update Management

Background check daily. Badge on installed app: *“WordPress 6.8 available”*. One-click update: **backs up first**, then updates. User cannot skip the backup step.

-----

## 20. Module: Settings & Panel Management

### Sections

```
Settings
├── General        Panel port, hostname, server name
├── Email          SMTP config for panel notifications
├── Security       2FA, session list, login history, terminal IP restrictions
├── Updates        Current version, changelog, version pinning
├── Users          Add/remove users, assign roles
├── Backups        Remote storage credentials, retention policy
├── Notifications  Alert email, thresholds per alert type
├── Audit Log      Full history of every panel action
└── Export/Import  Download/restore full panel config as JSON
```

### Audit Log

```
Jun 11 14:32  admin (192.168.1.5)   Created website mydomain.com
Jun 11 14:35  admin (192.168.1.5)   Issued SSL for mydomain.com
Jun 11 15:01  john  (192.168.1.8)   Changed PHP version → 8.3 (shop.mydomain.com)
Jun 11 16:44  admin (192.168.1.5)   Deleted mailbox info@mydomain.com
Jun 11 18:22  admin (203.0.113.42)  Opened root terminal [session recorded]
```

IP addresses logged with every action.

### Command Palette (⌘K)

Global search is a full **command palette** — not just a search box. Users can type actions directly:

```
⌘K → type "add website"    → opens Add Website form
⌘K → type "renew ssl"      → shows SSL renewal for all sites
⌘K → type "open terminal"  → opens terminal tab
⌘K → type "view nginx log" → opens nginx error log live tail
⌘K → type "backup now"     → triggers manual backup
⌘K → type "mydomain"       → shows all resources for mydomain.com
```

This is the highest-leverage UX feature for users who become comfortable with the panel.

-----

## 21. UI Structure & Design System

### Design Philosophy

Sharp, minimal, monochrome. Stripe-level polish with zero visual noise. Color is reserved exclusively for meaning. Everything else is black, white, and gray.

### Typography

**Font: Geist** (by Vercel). Sharp, technical, designed for developer-facing SaaS. Geist Mono for code and terminal areas.

```
Page titles:      Geist 18px, weight 600, #111111
Section labels:   Geist 13px, weight 500, #555555, uppercase, letter-spaced
Body text:        Geist 14px, weight 400, #333333
Secondary text:   Geist 13px, weight 400, #888888
Code / terminal:  Geist Mono 13px
```

### Color Palette

```
Background:       #FFFFFF
Surface:          #FAFAFA
Border:           #E5E5E5
Text primary:     #111111
Text secondary:   #888888
Text disabled:    #BBBBBB

Status Live:      #16A34A  (green)
Status Error:     #DC2626  (red)
Status Warning:   #D97706  (amber)
Status Pending:   #2563EB  (blue)

Accent / Primary: #111111  (buttons, active nav)
Hover:            #F5F5F5
Focus ring:       2px #111111 offset
```

No brand color. No gradients. Primary buttons are black on white.

### Corner Radius

**Sharp / fully square** — 0px on all components.

### Layout Shell

```
┌─────────────────────────────────────────────────────────────┐
│ SIDEBAR (240px / 64px collapsed)  │ TOP BAR (52px)          │
│                                   │ Breadcrumb  Search  🔔  │
│ ◈ juvia                     ├─────────────────────────│
│                                   │                         │
│ ── HOSTING ──────────             │   PAGE CONTENT          │
│   🌐 Websites         ● ○ ●       │                         │
│   📧 Email                        │                         │
│   🗄️  Databases                    │                         │
│   📁 Files                        │                         │
│                                   │                         │
│ ── SERVER ───────────             │                         │
│   🔒 SSL                          │                         │
│   🌍 DNS                          │                         │
│   🛡️  Firewall                     │                         │
│   💾 Backups                      │                         │
│   ⏰ Cron Jobs                    │                         │
│                                   │                         │
│ ── MONITOR ──────────             │                         │
│   📊 Metrics                      │                         │
│   📋 Logs                         │                         │
│   💻 Terminal                     │                         │
│                                   │                         │
│   ⚙️  Settings                     │                         │
│                                   │                         │
│ ┌─────────────────────────┐       │                         │
│ │ ● my-server.com         │       │                         │
│ │ CPU ████░░░░ 42%        │       │                         │
│ │ RAM ██░░░░░░ 28%        │       │                         │
│ ├─────────────────────────┤       │                         │
│ │ 👤 John Smith   Admin   │       │                         │
│ └─────────────────────────┘       │                         │
└─────────────────────────────────────────────────────────────┘
```

### Sidebar Details

- **Collapsed:** 64px, icons + tooltips only
- **Section headings:** 11px uppercase, #BBBBBB
- **Nav items:** 14px, 36px tall, hover #F5F5F5, active: 2px left border #111 + #F5F5F5
- **Website status dots:** 6px circles next to site names (green/red/gray). Tooltip on hover.
- **Server health widget:** Mini CPU/RAM bars. Click → Metrics page.
- **User profile:** Initials + name + role. Click → Profile / Change Password / Sign Out.

### Top Bar

```
Websites › mydomain.com          [⌘K Search...]   🔔(3)   J▾
```

- Breadcrumb: 13px, #888888 parents, #111111 current
- Search / Command palette: ⌘K, results grouped by type + actions
- Bell: dropdown with last 10 alerts + one-click actions
- Avatar: 28px square, initials, #111 background

### Overview Dashboard

```
── ALERTS (hidden when none) ────────────────────────────────
  🔴 mydomain.com is not responding              [Investigate]
  🟡 SSL expires in 9 days                          [Renew Now]

── SERVER HEALTH ─────────────────────────────────────────────
  [ CPU 23% ]  [ RAM 1.2/4GB ]  [ Disk 42/80GB ]  [ Network ]

── QUICK ACTIONS ─────────────────────────────────────────────
  [+ Add Website]  [+ Create Mailbox]  [+ New Database]
  [+ Run Backup]   [+ Open Terminal]

── WEBSITES ──────────────────────────────────────────────────
  mydomain.com        ● Live    SSL 84d    PHP 8.2
  shop.mydomain.com   ● Live    SSL 9d ⚠   PHP 8.1
  oldsite.com         ○ Susp.   SSL 44d    PHP 8.2

── RECENT ACTIVITY ───────────────────────────────────────────
  Jun 11 14:32  admin   Website created — mydomain.com
  Jun 11 14:35  admin   SSL issued — mydomain.com
```

### Page Structure

```
Page Title                                    [Primary Action]
Short plain-English description
──────────────────────────────────────────────────────────────
(filter / search bar if applicable)
──────────────────────────────────────────────────────────────
Content — separated by 1px #E5E5E5 dividers
No nested card boxes — whitespace + dividers only
```

### Tables

```
Domain              Status    SSL        PHP    Actions
────────────────────────────────────────────────────────
mydomain.com        ● Live    84 days    8.2    ···
shop.mydomain.com   ● Live    9 days ⚠   8.1    ···
oldsite.com         ○ Susp.   44 days    8.2    ···
```

Rows: 44px tall, hover #FAFAFA, click → detail page. `···` → contextual action popover.

### Detail Pages

Full page navigation (URL changes, browser back works).

```
← Back to Websites

mydomain.com                                          ● Live
Added Jun 8, 2026 · Nginx · PHP 8.2

[Open Site ↗]  [File Manager]  [Suspend]  [···]

──────────────────────────────────────────────────────────────
SSL CERTIFICATE
Valid · Expires in 84 days (Sep 3, 2026)           [Renew Now]
──────────────────────────────────────────────────────────────
WEB SERVER
Nginx · PHP 8.2 · PHP-FPM                [Change PHP Version]
──────────────────────────────────────────────────────────────
CACHING
FastCGI Cache  ● Enabled                          [Configure]
Redis Cache    ○ Disabled                            [Enable]
──────────────────────────────────────────────────────────────
```

### Modal Dialogs

```
┌──────────────────────────────────┐
│ Change PHP Version               │
│                                  │
│ Current: PHP 8.2                 │
│                                  │
│ New version                      │
│ ┌──────────────────────────┐     │
│ │ PHP 8.3              ▾  │     │
│ └──────────────────────────┘     │
│                                  │
│ ℹ Changing version reloads       │
│   PHP briefly — no downtime.     │
│                                  │
│ [Cancel]            [Save]       │
└──────────────────────────────────┘
```

Destructive modals require typing the resource name to confirm.

### Toast Notifications

Bottom-right, stacks upward. Success: black bg, 4s auto-dismiss. Error: red bg, stays until dismissed.

### Loading States

Skeleton placeholders with animated shimmer. No spinners except inline on clicked buttons.

### Motion

No page transitions — instant navigation. Only allowed motion: sidebar collapse (150ms), modal fade-in (100ms), toast slide-up (150ms), skeleton shimmer.

### Responsive

- Below 1024px: sidebar icon-only
- Below 768px: sidebar hidden, hamburger in top bar
- Tables: horizontal scroll on mobile
- Modals: full-screen on mobile

### Component Library

**shadcn/ui + Tailwind**, all overridden to match the sharp monochrome system:

|Component|Style                                 |
|---------|--------------------------------------|
|Button   |Square, black primary, ghost secondary|
|Input    |Square, 1px #E5E5E5, #111 focus ring  |
|Select   |Square, consistent with inputs        |
|Table    |No outer border, row hover only       |
|Modal    |Square, 480px max, 40% black backdrop |
|Toast    |Square, bottom-right stack            |
|Badge    |Square, small, semantic colors only   |
|Tooltip  |Square, black bg, white text, 12px    |
|Tabs     |Underline style only                  |

-----

## 22. Build Phases

|Phase                    |Modules                                                                                                                                                                                  |Deliverable                               |
|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------|
|**Phase 1 — Foundation** |Two-process architecture, systemd setup, auth (JWT + refresh + 2FA), first-run wizard, overview dashboard, metrics (VictoriaMetrics), proactive alerts, service registry, command palette|Running panel, secure, useful from day one|
|**Phase 2 — Web Hosting**|Website management (Nginx + Apache + task system), SSL (certbot), DNS (BIND9), config validation, soft-delete / trash                                                                    |Full website hosting                      |
|**Phase 3 — Services**   |Email (Postfix + Dovecot + Rspamd + deliverability tools), MySQL + PostgreSQL (+ visual manager), file manager (Monaco + path safety)                                                    |Complete hosting stack                    |
|**Phase 4 — Ops**        |Firewall (UFW + geo-blocking), backups (local + S3 + integrity verification), cron manager, logs viewer, alert delivery (email)                                                          |Production-ready operations               |
|**Phase 5 — Polish**     |One-click app installs, git deploy + webhooks, webmail (Roundcube), terminal session recording, .deb package, install script, GPG-signed self-updater                                    |Public release ready                      |

-----

## 23. What Changed from v1 Spec

|Area                     |v1 Spec                                               |v2 Revised                                                            |
|-------------------------|------------------------------------------------------|----------------------------------------------------------------------|
|**Process model**        |Single binary, single process, everything runs as root|Two processes: unprivileged HTTP server + separate root agent         |
|**IPC**                  |gRPC over Unix socket                                 |Simple JSON-RPC over Unix socket (gRPC added in v2 for multi-server)  |
|**Metrics storage**      |SQLite (all data)                                     |SQLite for relational + VictoriaMetrics for time-series               |
|**Auth**                 |JWT sessions (unspecified lifetime)                   |Short-lived JWT (15min) + httpOnly refresh tokens (7 days) + 2FA      |
|**Login security**       |Not specified                                         |Rate limiting, auto-block, login alerts, session management           |
|**File manager security**|Not specified                                         |Path traversal prevention, zip bomb protection, CSP headers           |
|**Config management**    |Edit files directly                                   |Atomic swap with validation before every reload                       |
|**Site deletion**        |Suspend only                                          |Suspend + soft-delete trash (30-day recovery)                         |
|**Backups**              |Create and restore                                    |Create + SHA-256 integrity check + dry-run restore verification       |
|**Terminal security**    |Audit log only                                        |Session recording + idle timeout + re-auth for root + IP restriction  |
|**Email PTR records**    |Implied panel can set them                            |Correctly marked as hosting-provider-only, UI shows exact value to set|
|**First-run experience** |Blank dashboard                                       |Mandatory 4-step setup wizard                                         |
|**Updates**              |Download and apply binary                             |GPG signature verification before applying                            |
|**Task system**          |Not defined                                           |Async task runner with live progress streaming from day one           |
|**Service registry**     |Not defined                                           |Tracks installed services, surfaces setup prompts for uninstalled ones|
|**Command palette**      |Basic search                                          |Full action-capable command palette (⌘K)                              |
|**Progress messaging**   |Technical step names                                  |Plain-English descriptions for every operation                        |

-----

*juvia — The server panel that non-technical people actually want to use.*  
*Revision 2 — Security-first, task-driven, plain-English throughout.*