# Juvia — Implementation Roadmap

**This is a concise reference. Detailed specifications are in `docs/` files.**

**Rule:** Follow phases in order. Do not skip ahead. Each phase's output is the next phase's input.

---

## Implementation Rules (MUST FOLLOW)

1. **REAL CODE ONLY** — No mock data, no placeholder functions, no TODO-only implementations. Every function must have real logic.
2. **Never skip ahead** — Complete Phase 1 before Phase 2, Phase 2 before Phase 3, etc.
3. **Always write tests** — Every Go function and React component must have tests.
4. **Security first** — Every file operation uses `safePath()`. Every config uses atomic swap.
5. **Plain-English UI** — Every technical term has a plain-English explanation.
6. **Atomic commits** — Each step is a single commit.
7. **Deploy often** — Push to `maruf@192.168.0.211` after each step.
8. **Build must succeed** — Every commit must pass `go build ./...` and `go test ./...` before deployment.

---

## Phase 1 — Foundation (Weeks 1-4)

**Goal:** Running Go binary with authenticated API and React frontend.

### Step 1.1: Project Structure + go.mod
- Create `cmd/panel/main.go`, `cmd/agent/main.go`
- Create `internal/` packages (db, api, socket, auth, tasks, utils, metrics)
- Create `web/` React app scaffold
- Reference: `docs/ARCHITECTURE.md` for project structure

### Step 1.2: SQLite Database + Models
- Implement `internal/db/db.go` — OpenDB, ApplyMigrations, CloseDB
- Write `internal/db/migrations/001_initial_schema.sql` from `docs/DATABASE.md`
- Define Go structs in `internal/db/models.go` for all tables

### Step 1.3: JSON-RPC over Unix Socket
- `internal/socket/protocol.go` — Request/Response types
- `internal/socket/server.go` — Agent socket server (root)
- `internal/socket/client.go` — Panel socket client (retry, circuit breaker)
- Reference: `docs/ARCHITECTURE.md` IPC section

### Step 1.4: Authentication Layer
- `internal/auth/jwt.go` — Generate/Verify JWT access + refresh tokens
- `internal/auth/bcrypt.go` — Password hashing (bcrypt cost 12)
- `internal/auth/session.go` — Session management in SQLite
- `internal/auth/totp.go` — 2FA TOTP (secret, QR URI, verify, recovery codes)
- Reference: `docs/SECURITY.md` auth section

### Step 1.5: HTTP Router + Middleware
- `internal/api/router.go` — Gin router with route groups
- `internal/api/middleware.go` — Auth, CSRF, RateLimit, SetupRequired
- Reference: `docs/API.md` for endpoint list

### Step 1.6: Auth API + Setup Wizard + User Management
- `internal/api/auth.go` — Login, refresh, logout, me, 2FA, sessions
- `internal/api/setup.go` — First-run wizard (status, first-run)
- `internal/api/users.go` — Admin CRUD for users

### Step 1.7: Task System
- `internal/tasks/runner.go` — Create/Update/Complete/Fail tasks, background worker
- `internal/tasks/status.go` — GetTask, ListTasks

### Step 1.8: Metrics Collection
- `internal/metrics/collector.go` — CPU, RAM, Disk, Network, LoadAvg, Processes
- `internal/metrics/types.go` — Metrics and Process structs
- Reference: `docs/ARCHITECTURE.md` metrics section

### Step 1.9: WebSocket Server
- `internal/ws/server.go` — Metrics stream, task progress, terminal proxy

### Step 1.10: Frontend Scaffold
- `web/package.json`, `web/src/main.tsx`, `web/src/App.tsx`
- Layout components (Sidebar, TopBar)
- API client (`web/src/lib/api.ts`)
- WebSocket client (`web/src/lib/ws.ts`)
- Reference: `docs/FRONTEND.md` for design system and routing

### Step 1.11: Dashboard Page
- `web/src/pages/Dashboard/Dashboard.tsx` — Server health, alerts, quick actions, website list, activity

### Step 1.12: Agent Stub
- `cmd/agent/main.go` — Entry point, socket server, `system.ping` handler

### Step 1.13: Panel Entry Point
- `cmd/panel/main.go` — HTTP server, Go embed for React static files

### Step 1.14: Systemd Services + Build + Deploy
- `scripts/juvia.service`, `scripts/juvia-agent.service`
- `scripts/build.sh`, `scripts/deploy.sh`

---

## Phase 2 — Web Hosting (Weeks 5-8)

**Goal:** Website CRUD with SSL and DNS.

### Step 2.1: Website API + Agent Methods
- `internal/api/websites.go` — CRUD, suspend, restore, trash
- `internal/agent/website.go` — create Linux user, document root, permissions

### Step 2.2: Nginx Config Generation
- `internal/agent/nginx/generator.go` — site config, PHP-FPM pool, validate, reload

### Step 2.3: SSL (Certbot)
- `internal/agent/ssl/certbot.go` — IssueSSL, RenewSSL, CheckDNS, AutoRenewCron
- `internal/api/ssl.go` — SSL check and renew endpoints

### Step 2.4: DNS (BIND9)
- `internal/agent/dns/bind.go` — GenerateZoneFile, ValidateZoneFile, ReloadBind
- `internal/api/dns.go` — DNS record CRUD + suggest

### Step 2.5: Website Frontend
- `web/src/pages/Websites/WebsiteList.tsx`
- `web/src/pages/Websites/WebsiteDetail.tsx`
- `web/src/pages/Websites/CreateWebsite.tsx`
- `web/src/pages/Websites/Trash.tsx`

### Step 2.6: File Manager API
- `internal/api/files.go` — list, upload, download, rename, delete, extract
- `internal/agent/files/operations.go` — ListFiles, UploadFile, DownloadFile, DeleteFile, RenameFile, ExtractArchive
- Reference: `docs/MODULES.md` File Manager section

---

## Phase 3 — Services (Weeks 9-12)

**Goal:** Email, databases, file manager UI.

### Step 3.1: Email API + Agent Methods
- `internal/api/email.go` — mailboxes, aliases, forwarders, catch-all, deliverability
- `internal/agent/email/postfix.go` — createMailbox, deleteMailbox, updateMailbox, createAlias, generateSPF/DKIM/DMARC
- Reference: `docs/MODULES.md` Email Hosting section

### Step 3.2: Database API + Agent Methods
- `internal/api/databases.go` — CRUD, users, export/import, tables, rows, query
- `internal/agent/database/mysql.go` — MySQL operations
- `internal/agent/database/postgresql.go` — PostgreSQL operations
- Reference: `docs/MODULES.md` Database Management section

### Step 3.3: File Manager Frontend
- `web/src/pages/Files/FileManager.tsx` — Monaco Editor, drag-drop, context menu

### Step 3.4: Visual Database Manager
- `web/src/pages/Databases/DatabaseManager.tsx` — Table browser, SQL editor, row editor

### Step 3.5: Email Frontend
- `web/src/pages/Email/EmailDashboard.tsx` — Mailbox list, alias, forwarder, deliverability

---

## Phase 4 — Ops (Weeks 13-15)

**Goal:** Firewall, backups, cron, logs, alerts.

### Step 4.1: Firewall API + Agent
- `internal/api/firewall.go` — CRUD for rules
- `internal/agent/firewall/ufw.go` — addRule, deleteRule, listRules, blockCountry

### Step 4.2: Backup API + Agent
- `internal/api/backups.go` — CRUD, schedules, restore
- `internal/agent/backup/backup.go` — create, restore, verify, uploadS3, dryRun

### Step 4.3: Cron API + Agent
- `internal/api/cron.go` — CRUD, enable/disable, logs
- `internal/agent/cron/manager.go` — create, delete, run, history

### Step 4.4: Logs API
- `internal/api/logs.go` — Website logs, system logs, download

### Step 4.5: Logs Frontend
- `web/src/pages/Logs/LogViewer.tsx` — Live tail, filter, search, download

### Step 4.6: Alert Engine
- `internal/alerts/engine.go` — Background checker (website down/slow, SSL expiry, disk, services)

---

## Phase 5 — Polish (Weeks 16-18)

**Goal:** App installs, git deploy, webmail, terminal recording, packaging.

### Step 5.1: One-Click Apps
- `internal/api/apps.go` — list, install, update
- `internal/agent/apps/installer.go` — WordPress, Laravel, Next.js, etc.
- Reference: `docs/MODULES.md` One-Click App Installs section

### Step 5.2: Git Deploy
- `internal/api/git.go` — setup, pull, webhook
- `internal/api/git_webhook.go` — webhook receiver with HMAC verification

### Step 5.3: Webmail
- `internal/api/webmail.go` — enable/disable, get URL
- `internal/agent/email/webmail.go` — Roundcube install/uninstall

### Step 5.4: Terminal Recording
- `internal/agent/terminal/record.go` — startRecording, writeFrame, stopRecording
- `web/src/pages/Terminal/Recordings.tsx` — Playback UI

### Step 5.5: Self-Updater
- `internal/updater/updater.go` — CheckForUpdates, DownloadUpdate, VerifyGPGSignature, ApplyUpdate, Rollback

### Step 5.6: Packaging
- `scripts/build-deb.sh` — .deb package
- `scripts/install.sh` — One-liner install

### Step 5.7: Command Palette
- `web/src/components/CommandPalette.tsx` — Cmd+K search

### Step 5.8: Settings Pages
- `web/src/pages/Settings/Settings.tsx` — General, Email, Security, Updates, Users, Backups, Notifications, Audit Log, Export/Import

---

## Authoritative Sources

|File|Content|
|----|-------|
|`docs/ARCHITECTURE.md`|System architecture, two-process model, IPC, project structure|
|`docs/DATABASE.md`|SQLite schema (all tables, indexes, migrations)|
|`docs/API.md`|All REST endpoints, WebSocket channels, JSON-RPC methods, error codes|
|`docs/SECURITY.md`|Auth model, CSRF, path traversal, atomic swaps, session recording|
|`docs/FRONTEND.md`|Design system, routing, component structure, API client|
|`docs/MODULES.md`|Detailed module implementation notes (website, email, DNS, etc.)|
|`docs/TESTING.md`|Testing strategy, unit/integration/E2E tests|
|`docs/DEPLOYMENT.md`|Installation, systemd services, directory structure|
|`docs/ENVIRONMENT.md`|Local dev setup, remote server access|

---

## Success Criteria

|Phase|Done when|
|------|----------|
|Phase 1|`curl http://192.168.0.211:8080/api/v1/health` returns `200 OK`|
|Phase 2|Can create a website, issue SSL, and see it in a browser|
|Phase 3|Can create a mailbox, database, and browse files|
|Phase 4|Can set firewall rules, run backups, and create cron jobs|
|Phase 5|Can install WordPress, use webmail, and update via the panel|