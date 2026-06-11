# Juvia — Complete Implementation Roadmap

This document is the single source of truth for AI coding agents implementing the Juvia server control panel.

**Rule:** Follow the phases in order. Do not skip ahead. Each phase's output is the next phase's input.

---

## Phase 1 — Foundation (Weeks 1-4)

**Goal:** A running Go binary that serves an authenticated API and a React frontend.

### Step 1.1: Go Project Structure

Create the following directories and files:

```
cmd/
  panel/main.go         # HTTP panel entry point
  agent/main.go         # Privileged agent entry point
internal/
  db/
    db.go               # SQLite connection, migrations
    migrations/001_initial_schema.sql
    models.go           # Go struct definitions for all tables
  api/
    router.go           # HTTP router setup
    middleware.go       # Auth middleware, rate limiting
    auth.go             # Login, refresh, logout handlers
    health.go           # Health check handler
  socket/
    client.go           # Panel -> Agent Unix socket client
    server.go           # Agent Unix socket listener
    protocol.go         # JSON-RPC request/response types
  auth/
    jwt.go              # JWT generate/verify
    bcrypt.go           # Password hashing
    session.go          # Session management
    totp.go             # 2FA TOTP
  tasks/
    runner.go           # Async task execution
    status.go           # Task status tracking
web/
  src/
    App.tsx
    main.tsx
    index.css
  package.json
  vite.config.ts
  tsconfig.json
scripts/
  install.sh
  build-deb.sh
```

**Constraints:**
- Use `slog` for structured logging everywhere
- Use `context.Context` for all operations
- Wrap errors: `fmt.Errorf("context: %w", err)`
- Go version: 1.21+

### Step 1.2: SQLite Database + Migrations

**File:** `internal/db/db.go`

Implement:
1. `OpenDB(path string) (*sql.DB, error)` — opens SQLite, enables WAL mode
2. `ApplyMigrations(db *sql.DB) error` — runs migrations from `migrations/*.sql`
3. `CloseDB(db *sql.DB) error` — graceful close

**File:** `internal/db/migrations/001_initial_schema.sql`

Write the complete initial schema from `DATABASE.md`. All tables: users, sessions, websites, domains, zones, dns_records, databases, db_users, mailboxes, email_aliases, email_forwarders, email_catchall, tasks, audit_log, firewall_rules, cron_jobs, cron_job_logs, backups, backup_schedules, alerts, services, settings, apps, user_sites.

**File:** `internal/db/models.go`

Define Go structs for every table. Example:
```go
type User struct {
    ID                int64     `db:"id" json:"id"`
    Username          string    `db:"username" json:"username"`
    PasswordHash      string    `db:"password_hash" json:"-"`
    Role              string    `db:"role" json:"role"`
    Email             string    `db:"email" json:"email"`
    Active            bool      `db:"active" json:"active"`
    TwoFAEnabled      bool      `db:"2fa_enabled" json:"2fa_enabled"`
    TwoFASecret       string    `db:"2fa_secret" json:"-"`
    TwoFARecoveryCodes string   `db:"2fa_recovery_codes" json:"-"`
    CreatedAt         time.Time `db:"created_at" json:"created_at"`
    LastLogin         time.Time `db:"last_login" json:"last_login"`
    UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}
```

### Step 1.3: JSON-RPC Protocol Over Unix Socket

**File:** `internal/socket/protocol.go`

Define types:
```go
type Request struct {
    ID     string      `json:"id"`
    Method string      `json:"method"`
    Params interface{} `json:"params,omitempty"`
}

type Response struct {
    ID      string      `json:"id"`
    Result  interface{} `json:"result,omitempty"`
    Error   *Error      `json:"error,omitempty"`
}

type Error struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    UserMessage string `json:"user_message"`
}
```

**File:** `internal/socket/server.go`

Implement agent-side socket server:
1. `NewServer(path string) (*Server, error)` — creates Unix socket with `660` permissions
2. `Listen()` — accept connections
3. `Handle(method string, handler HandlerFunc)` — register method handlers
4. Graceful shutdown on SIGTERM

**File:** `internal/socket/client.go`

Implement panel-side client:
1. `NewClient(path string) *Client`
2. `Call(ctx context.Context, method string, params interface{}, result interface{}) error`
3. Connection pooling with retry (exponential backoff)
4. Circuit breaker: after 3 failures, mark agent offline

### Step 1.4: Authentication Layer

**File:** `internal/auth/jwt.go`

Implement:
1. `GenerateAccessToken(userID int64, role string) (string, error)` — 15 min expiry
2. `GenerateRefreshToken(userID int64) (string, string, error)` — returns token and hash
3. `VerifyAccessToken(token string) (*Claims, error)` — validate JWT
4. `VerifyRefreshToken(token string) (*Claims, error)` — check hash in DB

**File:** `internal/auth/bcrypt.go`

Implement:
1. `HashPassword(password string) (string, error)` — bcrypt cost 12
2. `VerifyPassword(hash, password string) bool`

**File:** `internal/auth/session.go`

Implement:
1. `CreateSession(db *sql.DB, userID int64, ip, userAgent string) (string, error)` — store refresh token hash
2. `RevokeSession(db *sql.DB, token string) error`
3. `ListSessions(db *sql.DB, userID int64) ([]Session, error)`
4. `CleanupExpiredSessions(db *sql.DB) error` — background job

**File:** `internal/auth/totp.go`

Implement:
1. `GenerateSecret() (string, error)`
2. `GenerateQRCodeURI(secret, username, issuer string) string`
3. `VerifyCode(secret, code string) bool`
4. `GenerateRecoveryCodes() ([]string, error)` — 10 single-use codes

### Step 1.5: HTTP Router + Middleware

**File:** `internal/api/router.go`

Implement:
1. `NewRouter(db *sql.DB, socketClient *socket.Client) *gin.Engine` (or `chi.Mux`)
2. Mount all route groups:
   - `/api/v1/auth/*` — public
   - `/api/v1/*` — requires `AuthMiddleware` + `CSRFMiddleware`
   - `/api/v1/setup/*` — public (only before setup complete)
   - `/api/v1/health` — public
3. Attach standard middleware:
   - Request ID injection
   - Structured logging (slog)
   - Rate limiting (`RateLimitMiddleware`)
   - Panic recovery

**File:** `internal/api/middleware.go`

Implement:
1. `AuthMiddleware(db *sql.DB) Middleware` — verify JWT from `Authorization: Bearer <token>` header
2. `CSRFMiddleware() Middleware` — verify `X-CSRF-Token` header
3. `RateLimitMiddleware(store limiter.Store) Middleware` — 5/min for login, 100/min for API
4. `SetupRequiredMiddleware(db *sql.DB) Middleware` — redirect to setup if `settings.setup_complete` is false

### Step 1.6: Auth API Endpoints

**File:** `internal/api/auth.go`

Implement handlers:
1. `POST /api/v1/auth/login` — Login, return access token + set refresh httpOnly cookie
2. `POST /api/v1/auth/refresh` — Read refresh cookie, verify, return new access token
3. `POST /api/v1/auth/logout` — Revoke refresh token, clear cookie
4. `GET /api/v1/auth/me` — Return current user, set `X-CSRF-Token` header
5. `POST /api/v1/auth/2fa/enable` — Generate secret, return QR URI
6. `POST /api/v1/auth/2fa/verify` — Verify TOTP code
7. `DELETE /api/v1/auth/sessions/:id` — Revoke specific session

### Step 1.7: Task System

**File:** `internal/tasks/runner.go`

Implement:
1. `CreateTask(db *sql.DB, taskType string, userID int64) (*Task, error)` — create task, return task ID
2. `UpdateTask(db *sql.DB, taskID string, status string, progress int, step string) error`
3. `CompleteTask(db *sql.DB, taskID string, result json.RawMessage) error`
4. `FailTask(db *sql.DB, taskID string, err error) error`
5. Background worker: `StartWorker(db *sql.DB, agentClient *socket.Client)` — process pending tasks

**File:** `internal/tasks/status.go`

Implement:
1. `GetTask(db *sql.DB, taskID string) (*Task, error)`
2. `ListTasks(db *sql.DB, userID int64, limit int) ([]Task, error)`

### Step 1.8: WebSocket Server

**File:** `internal/ws/server.go`

Implement:
1. `NewServer(db *sql.DB) *Server` — upgrade HTTP to WebSocket
2. `HandleMetrics(conn *websocket.Conn, userID int64)` — push metrics every 2s
3. `HandleTasks(conn *websocket.Conn, userID int64)` — push task updates
4. `HandleTerminal(conn *websocket.Conn, sessionID string)` — PTY proxy
5. Pub-sub: `Subscribe(userID int64, eventType string) chan Event` and `Broadcast(event Event)`

### Step 1.9: Frontend Scaffold

**File:** `web/package.json`

Dependencies:
- react, react-dom, typescript
- react-router-dom, react-query, zustand
- tailwindcss, shadcn/ui
- xterm, monaco-editor
- recharts

**File:** `web/src/main.tsx`

Setup: React Query client, Zustand store, React Router.

**File:** `web/src/App.tsx`

Implement route structure from `FRONTEND.md`:
- `/login` — Login page
- `/setup/*` — First-run wizard (no auth required)
- `/dashboard` — Dashboard with sidebar layout
- `/websites/*` — Website list + detail
- `/email/*` — Email management
- `/databases/*` — Database management
- `/files/*` — File manager
- `/firewall/*` — Firewall rules
- `/backups/*` — Backup management
- `/cron/*` — Cron jobs
- `/metrics/*` — Server metrics
- `/terminal/*` — Web terminal
- `/settings/*` — Settings pages

**File:** `web/src/components/layout/Sidebar.tsx`

Implement the sidebar from the spec:
- 240px wide, 0px border-radius
- Section headings: 11px uppercase, #BBBBBB
- Nav items: 14px, 36px tall, hover #F5F5F5, active: 2px left border #111 + #F5F5F5
- Server health widget (mini CPU/RAM bars)
- User profile (initials + name + role)
- Collapsible: 64px icon-only mode

**File:** `web/src/components/layout/TopBar.tsx`

Implement:
- Breadcrumb: 13px, #888888 parents, #111111 current
- Search / Command palette (⌘K) — placeholder for now
- Bell icon with alert dropdown
- Avatar (28px square, initials, #111 background)

**File:** `web/src/lib/api.ts`

Implement API client:
1. `api.get<T>(path, params?)` — typed GET
2. `api.post<T>(path, body)` — typed POST
3. `api.put<T>(path, body)` — typed PUT
4. `api.delete<T>(path)` — typed DELETE
5. Automatic JWT attachment, refresh on 401, CSRF header injection
6. Standard response envelope parsing

**File:** `web/src/lib/ws.ts`

Implement WebSocket client:
1. `connect()` — open WebSocket, auto-reconnect with exponential backoff
2. `subscribe(channel, callback)` — subscribe to channels
3. `send(channel, data)` — send data
4. Heartbeat ping/pong every 30s

### Step 1.10: Dashboard Page

**File:** `web/src/pages/Dashboard/Dashboard.tsx`

Implement the overview dashboard from the spec:
1. **Alerts section** — hidden when none, shows critical/warning alerts
2. **Server health** — CPU/RAM/Disk/Network widgets with mini sparklines
3. **Quick actions** — buttons: Add Website, Create Mailbox, New Database, Run Backup, Open Terminal
4. **Website list** — domain, status dot, SSL days, PHP version
5. **Recent activity** — audit log entries

### Step 1.11: Agent Stub

**File:** `cmd/agent/main.go`

Implement the agent entry point:
1. `func main()` — parse flags, set up slog
2. `system.run()` — run as root, create socket at `/var/run/juvia/agent.sock`
3. Register handler: `system.ping` — returns `pong`
4. Graceful shutdown on SIGTERM

### Step 1.12: Panel Entry Point

**File:** `cmd/panel/main.go`

Implement:
1. `func main()` — parse flags (`-port`, `-db`, `-socket`, `-env`)
2. `panel.run()` — open DB, apply migrations, connect to agent socket, start HTTP server
3. `web.go` — embed `web/dist` via `go:embed`, serve as static files
4. Fallback: any unmatched route serves `index.html` (SPA routing)

### Step 1.13: Systemd Services

**File:** `scripts/juvia.service`

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

**File:** `scripts/juvia-agent.service`

```ini
[Unit]
Description=juvia Agent (privileged)
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/juvia agent run
Restart=on-failure
RestartSec=5s
TimeoutStopSec=30s

[Install]
WantedBy=multi-user.target
```

### Step 1.14: Build Script

**File:** `scripts/build.sh`

Implement:
1. `npm ci && npm run build` in `web/` directory
2. `go build -o juvia ./cmd/...` — single binary with embedded frontend
3. `go build -o juvia-agent ./cmd/agent` (or same binary with subcommand)
4. `go test ./...` — run tests

### Step 1.15: Deployment to Remote

**File:** `scripts/deploy.sh`

Implement:
1. `scp juvia maruf@192.168.0.211:/tmp/`
2. `ssh maruf@192.168.0.211 "sudo dpkg -i /tmp/juvia"`
3. `ssh maruf@192.168.0.211 "sudo systemctl restart juvia juvia-agent"`

---

## Phase 2 — Web Hosting (Weeks 5-8)

**Goal:** Create, manage, suspend, and delete websites with SSL and DNS.

### Step 2.1: Website API Endpoints

**File:** `internal/api/websites.go`

Implement handlers:
1. `GET /api/v1/websites` — list all websites (with pagination)
2. `POST /api/v1/websites` — create website (triggers async task)
3. `GET /api/v1/websites/:id` — get website details
4. `PUT /api/v1/websites/:id` — update website
5. `DELETE /api/v1/websites/:id` — soft-delete (move to trash)
6. `POST /api/v1/websites/:id/suspend` — suspend
7. `POST /api/v1/websites/:id/restore` — restore from trash
8. `GET /api/v1/websites/trash` — list trashed

### Step 2.2: Website Agent Methods

**File:** `internal/agent/website.go`

Implement agent methods:
1. `website.create` — create Linux user, document root, set permissions
2. `website.delete` — soft-delete (move to trash dir)
3. `website.suspend` — swap nginx vhost to suspended page
4. `website.restore` — move from trash back to live

### Step 2.3: Nginx Config Generation

**File:** `internal/agent/nginx/generator.go`

Implement:
1. `GenerateSiteConfig(site Website) ([]byte, error)` — generate nginx vhost
2. `GeneratePHPFPMPool(site Website) ([]byte, error)` — generate PHP-FPM pool
3. `ValidateConfig(path string) error` — run `nginx -t`
4. `ReloadNginx() error` — `systemctl reload nginx`

### Step 2.4: SSL (Certbot)

**File:** `internal/agent/ssl/certbot.go`

Implement:
1. `IssueSSL(domain string) error` — run `certbot certonly`
2. `RenewSSL(domain string) error` — run `certbot renew`
3. `CheckDNS(domain string) (bool, error)` — check if domain resolves to server IP
4. `AutoRenewCron()` — register cron job for auto-renewal

**File:** `internal/api/ssl.go`

Implement:
1. `GET /api/v1/websites/:id/ssl/check` — check DNS readiness
2. `POST /api/v1/websites/:id/ssl-renew` — trigger SSL renewal task

### Step 2.5: DNS (BIND9)

**File:** `internal/agent/dns/bind.go`

Implement:
1. `GenerateZoneFile(zone Zone, records []DNSRecord) ([]byte, error)`
2. `ValidateZoneFile(zoneFile string) error` — `named-checkzone`
3. `ReloadBind() error` — `systemctl reload bind9`

**File:** `internal/api/dns.go`

Implement:
1. `GET /api/v1/websites/:id/dns/records` — list DNS records
2. `POST /api/v1/websites/:id/dns/records` — create record
3. `PUT /api/v1/dns/records/:id` — update record
4. `DELETE /api/v1/dns/records/:id` — delete record
5. `POST /api/v1/websites/:id/dns/suggest` — check missing records and suggest

### Step 2.6: Website Frontend

**File:** `web/src/pages/Websites/WebsiteList.tsx`

Implement:
1. Table with columns: Domain, Status, SSL, PHP, Actions
2. Status dot: green (live), red (suspended), gray (trash)
3. `···` action menu: suspend, restore, delete
4. Search + filter + pagination

**File:** `web/src/pages/Websites/WebsiteDetail.tsx`

Implement sections:
1. SSL Certificate — validity, expiry, renew button
2. Web Server — nginx/apache, PHP version
3. PHP Config — memory_limit, upload_max_filesize, etc.
4. Caching — FastCGI, Redis, Memcached toggles
5. Logs — live tail (WebSocket), download
6. Danger zone — suspend, delete buttons

**File:** `web/src/pages/Websites/CreateWebsite.tsx`

Implement form:
1. Domain input with validation
2. PHP version selector (8.1, 8.2, 8.3)
3. Web server selector (nginx/apache)
4. Submit triggers task, shows progress via WebSocket

### Step 2.7: File Manager API

**File:** `internal/api/files.go`

Implement handlers:
1. `GET /api/v1/websites/:id/files` — list files with `path` query param
2. `POST /api/v1/websites/:id/files/upload` — multipart upload
3. `GET /api/v1/websites/:id/files/download` — download file/folder
4. `PUT /api/v1/websites/:id/files/rename` — rename
5. `DELETE /api/v1/websites/:id/files/delete` — delete
6. `POST /api/v1/websites/:id/files/extract` — extract zip/tar

All paths validated with `safePath()` before any operation.

**File:** `internal/agent/files/operations.go`

Implement:
1. `ListFiles(basePath, subPath string) ([]FileInfo, error)`
2. `UploadFile(destPath string, reader io.Reader) error`
3. `DownloadFile(path string) (io.ReadCloser, error)`
4. `DeleteFile(path string) error`
5. `RenameFile(oldPath, newPath string) error`
6. `ExtractArchive(archivePath, destPath string) error` — with zip bomb protection

### Step 2.8: Website Trash Frontend

**File:** `web/src/pages/Websites/Trash.tsx`

Implement:
1. List of trashed websites with delete date
2. Restore button per site
3. Permanent delete button (with confirmation)
4. Purge all button (with confirmation)

---

## Phase 3 — Services (Weeks 9-12)

**Goal:** Email hosting, database management, file manager UI.

### Step 3.1: Email API

**File:** `internal/api/email.go`

Implement handlers:
1. `GET /api/v1/email/mailboxes` — list mailboxes
2. `POST /api/v1/email/mailboxes` — create mailbox
3. `GET /api/v1/email/mailboxes/:id` — get mailbox details
4. `DELETE /api/v1/email/mailboxes/:id` — delete mailbox
5. `PUT /api/v1/email/mailboxes/:id` — update mailbox (quota, autoresponder)
6. `GET /api/v1/email/aliases` — list aliases
7. `POST /api/v1/email/aliases` — create alias
8. `DELETE /api/v1/email/aliases/:id` — delete alias
9. `GET /api/v1/email/forwarders` — list forwarders
10. `POST /api/v1/email/forwarders` — create forwarder
11. `GET /api/v1/email/catch-all` — get catch-all
12. `PUT /api/v1/email/catch-all` — set catch-all
13. `GET /api/v1/email/deliverability` — deliverability status

### Step 3.2: Email Agent Methods

**File:** `internal/agent/email/postfix.go`

Implement:
1. `email.createMailbox` — create system user for mailbox
2. `email.deleteMailbox` — delete mailbox user
3. `email.updateMailbox` — update quota, autoresponder
4. `email.createAlias` — add to aliases table
5. `email.generateSPF(domain string) string` — generate SPF record
6. `email.generateDKIM(domain string) (string, string)` — generate DKIM keypair
7. `email.generateDMARC(domain string) string` — generate DMARC record

### Step 3.3: Database API

**File:** `internal/api/databases.go`

Implement handlers:
1. `GET /api/v1/databases` — list databases
2. `POST /api/v1/databases` — create database
3. `GET /api/v1/databases/:id` — get details
4. `DELETE /api/v1/databases/:id` — delete database
5. `POST /api/v1/databases/:id/users` — create DB user
6. `DELETE /api/v1/databases/:id/users/:user_id` — delete DB user
7. `POST /api/v1/databases/:id/export` — export SQL
8. `POST /api/v1/databases/:id/import` — import SQL
9. `GET /api/v1/databases/:id/tables` — list tables
10. `GET /api/v1/databases/:id/tables/:table/rows` — get rows
11. `POST /api/v1/databases/:id/query` — run SQL

### Step 3.4: Database Agent Methods

**File:** `internal/agent/database/mysql.go`

Implement:
1. `database.createMySQLDB` — `CREATE DATABASE`
2. `database.createMySQLUser` — `CREATE USER` + `GRANT`
3. `database.deleteMySQLDB` — `DROP DATABASE`
4. `database.exportMySQLDB` — `mysqldump`
5. `database.importMySQLDB` — `mysql < file.sql`
6. `database.queryMySQL` — execute query, return results

**File:** `internal/agent/database/postgresql.go`

Same as above but for PostgreSQL (`createdb`, `createuser`, `pg_dump`, `psql`).

### Step 3.5: File Manager Frontend

**File:** `web/src/pages/Files/FileManager.tsx`

Implement:
1. Breadcrumb navigation
2. File list with icons, permissions, size, date
3. Drag-and-drop upload
4. Right-click context menu: edit, rename, delete, chmod, extract
5. Monaco Editor overlay for file editing
6. Bulk select with checkboxes
7. Hidden files toggle

### Step 3.6: Visual Database Manager

**File:** `web/src/pages/Databases/DatabaseManager.tsx`

Implement:
1. Table list sidebar
2. Table browser: rows with pagination, sorting
3. Monaco-based SQL query editor
4. Insert/edit row modal
5. Table structure view

### Step 3.7: Email Frontend

**File:** `web/src/pages/Email/EmailDashboard.tsx`

Implement:
1. Mailbox list with quota bars
2. Alias list
3. Forwarder list
4. Deliverability health card (SPF, DKIM, DMARC, PTR)
5. Catch-all configuration
6. Mail deliverability test tool

---

## Phase 4 — Ops (Weeks 13-15)

**Goal:** Firewall, backups, cron, logs, alerts.

### Step 4.1: Firewall API

**File:** `internal/api/firewall.go`

Implement handlers:
1. `GET /api/v1/firewall/rules` — list rules
2. `POST /api/v1/firewall/rules` — create rule
3. `GET /api/v1/firewall/rules/:id` — get details
4. `PUT /api/v1/firewall/rules/:id` — update
5. `DELETE /api/v1/firewall/rules/:id` — delete

### Step 4.2: Firewall Agent

**File:** `internal/agent/firewall/ufw.go`

Implement:
1. `firewall.addRule` — `ufw allow/deny`
2. `firewall.deleteRule` — `ufw delete`
3. `firewall.listRules` — `ufw status`
4. `firewall.blockCountry` — use geoip module

### Step 4.3: Backup API

**File:** `internal/api/backups.go`

Implement handlers:
1. `GET /api/v1/backups` — list backups
2. `POST /api/v1/backups` — create backup
3. `GET /api/v1/backups/:id` — get details
4. `POST /api/v1/backups/:id/restore` — restore
5. `DELETE /api/v1/backups/:id` — delete
6. `GET /api/v1/backup-schedules` — list schedules
7. `POST /api/v1/backup-schedules` — create schedule
8. `PUT /api/v1/backup-schedules/:id` — update
9. `DELETE /api/v1/backup-schedules/:id` — delete

### Step 4.4: Backup Agent

**File:** `internal/agent/backup/backup.go`

Implement:
1. `backup.create` — tar/rsync files + mysqldump
2. `backup.restore` — extract + import SQL
3. `backup.verify` — SHA-256 checksum + entry count
4. `backup.uploadS3` — upload to S3/R2/B2
5. `backup.dryRun` — extract to temp, verify

### Step 4.5: Cron API

**File:** `internal/api/cron.go`

Implement handlers:
1. `GET /api/v1/cron` — list jobs
2. `POST /api/v1/cron` — create job
3. `GET /api/v1/cron/:id` — get details
4. `PUT /api/v1/cron/:id` — update
5. `DELETE /api/v1/cron/:id` — delete
6. `POST /api/v1/cron/:id/enable` — enable
7. `POST /api/v1/cron/:id/disable` — disable
8. `GET /api/v1/cron/:id/logs` — execution history

### Step 4.6: Cron Agent

**File:** `internal/agent/cron/manager.go`

Implement:
1. `cron.create` — write to crontab or system cron
2. `cron.delete` — remove from crontab
3. `cron.run` — execute job, capture output, log result
4. `cron.history` — read execution logs

### Step 4.7: Logs API

**File:** `internal/api/logs.go`

Implement handlers:
1. `GET /api/v1/websites/:id/logs` — get logs (access/error)
2. `GET /api/v1/logs/system` — system logs
3. `GET /api/v1/logs/download` — download log file

### Step 4.8: Logs Frontend

**File:** `web/src/pages/Logs/LogViewer.tsx`

Implement:
1. Live log tail via WebSocket
2. Filter by level (INFO, ERROR, WARN)
3. Search within logs
4. Download button
5. Log rotation indicator

### Step 4.9: Alerts API

**File:** `internal/api/alerts.go`

Implement handlers:
1. `GET /api/v1/alerts` — list alerts
2. `POST /api/v1/alerts/:id/acknowledge` — acknowledge
3. `DELETE /api/v1/alerts/:id` — delete
4. `GET /api/v1/alerts/settings` — alert settings
5. `PUT /api/v1/alerts/settings` — update settings

### Step 4.10: Alert Engine

**File:** `internal/alerts/engine.go`

Implement background checker (every 60s):
1. Check website down (HTTP GET, timeout 5s)
2. Check website slow (response > 3s)
3. Check SSL expiry (< 14 days)
4. Check disk usage (> 80%)
5. Check service status (systemctl)
6. Fire alerts if thresholds exceeded

---

## Phase 5 — Polish (Weeks 16-18)

**Goal:** App installs, git deploy, webmail, terminal recording, packaging.

### Step 5.1: One-Click App Installs

**File:** `internal/api/apps.go`

Implement handlers:
1. `GET /api/v1/websites/:id/apps` — list installed apps
2. `POST /api/v1/websites/:id/apps/install` — install app
3. `POST /api/v1/websites/:id/apps/:app_id/update` — update app

**File:** `internal/agent/apps/installer.go`

Implement:
1. `apps.installWordPress` — download, create DB, configure wp-config.php
2. `apps.installLaravel` — composer install, create DB, configure .env
3. `apps.installNextJS` — npx create-next-app, configure
4. `apps.checkUpdates` — check latest version via API

### Step 5.2: Git Deploy

**File:** `internal/api/git.go`

Implement handlers:
1. `GET /api/v1/websites/:id/git` — get config
2. `POST /api/v1/websites/:id/git` — setup git
3. `POST /api/v1/websites/:id/git/pull` — trigger pull
4. `GET /api/v1/websites/:id/git/webhook` — get webhook URL

**File:** `internal/api/git_webhook.go`

Implement:
1. `POST /webhook/:token` — webhook receiver
2. Verify HMAC-SHA256 signature
3. Trigger git pull

### Step 5.3: Webmail

**File:** `internal/api/webmail.go`

Implement:
1. `GET /api/v1/email/webmail` — get webmail URL
2. `POST /api/v1/email/webmail` — enable/disable webmail

**File:** `internal/agent/email/webmail.go`

Implement:
1. `email.installRoundcube` — download, configure
2. `email.uninstallRoundcube` — remove

### Step 5.4: Terminal Recording

**File:** `internal/agent/terminal/record.go`

Implement:
1. `terminal.startRecording` — open `.cast` file
2. `terminal.writeFrame` — write timestamp + data
3. `terminal.stopRecording` — close file

**File:** `web/src/pages/Terminal/Recordings.tsx`

Implement:
1. List of recordings
2. Playback in browser
3. Download `.cast` file

### Step 5.5: Self-Updater

**File:** `internal/updater/updater.go`

Implement:
1. `CheckForUpdates()` — query GitHub Releases API
2. `DownloadUpdate(url string) error` — download binary
3. `VerifyGPGSignature(binary, signature []byte) error` — verify GPG
4. `ApplyUpdate()` — backup old binary, replace, restart
5. `Rollback()` — restore backup if update fails

### Step 5.6: Packaging

**File:** `scripts/build-deb.sh`

Implement:
1. Build binary
2. Create `DEBIAN/control`
3. Install systemd services
4. Create `juvia` user/group
5. Set up directories (`/etc/juvia`, `/var/lib/juvia`, `/var/log/juvia`)
6. Package with `dpkg-deb`

**File:** `scripts/install.sh`

Implement:
1. Check OS (Ubuntu 22.04+, Debian 11+)
2. Install dependencies (nginx, php-fpm, certbot, etc.)
3. Download latest `.deb` from GitHub
4. Install package
5. Start services
6. Print success message + URL

### Step 5.7: Command Palette

**File:** `web/src/components/CommandPalette.tsx`

Implement:
1. `Cmd+K` shortcut
2. Search input with instant results
3. Actions: add website, renew SSL, open terminal, etc.
4. Resource search: type "mydomain" to show all resources for that domain
5. Keyboard navigation (arrow keys, enter)

### Step 5.8: Settings Pages

**File:** `web/src/pages/Settings/Settings.tsx`

Implement sections:
1. General — panel port, hostname, server name
2. Email — SMTP config for alerts
3. Security — 2FA, sessions, login history, terminal IP restrictions
4. Updates — version, changelog, version pinning
5. Users — add/remove, role assignment
6. Backups — remote storage credentials, retention
7. Notifications — alert thresholds
8. Audit Log — full history with IP
9. Export/Import — download/restore config as JSON

---

## Implementation Rules

1. **Never skip ahead.** Complete Phase 1 before starting Phase 2.
2. **Always write tests.** Every Go function and React component must have tests.
3. **Security first.** Every file operation uses `safePath()`. Every config uses atomic swap.
4. **Plain-English UI.** Every technical term has a plain-English explanation in the UI.
5. **Atomic commits.** Each step should be a single commit.
6. **Deploy often.** Push to `maruf@192.168.0.211` after each step.

## Success Criteria

- **Phase 1 done:** `curl http://192.168.0.211:8080/api/v1/health` returns `200 OK`
- **Phase 2 done:** Can create a website, issue SSL, and see it in a browser
- **Phase 3 done:** Can create a mailbox, database, and browse files
- **Phase 4 done:** Can set firewall rules, run backups, and create cron jobs
- **Phase 5 done:** Can install WordPress, use webmail, and update via the panel
