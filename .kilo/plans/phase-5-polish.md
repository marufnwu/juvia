# Phase 5 — Polish Implementation Plan

**Goal:** App installs, git deploy, webmail, terminal recording, self-updater, packaging, command palette, settings. Success = can install WordPress, git deploy, use webmail, view terminal recordings, and update the panel.

**Rule:** Follow steps in order. Each step is a single atomic commit that passes `go build ./...` and `go test ./...`.

---

## Step 5.1: One-Click Apps

### Backend

**`internal/db/app_queries.go`** — SQLite operations for `apps` table:
- `ListAppsByWebsite(ctx, websiteID) ([]App, error)`
- `CreateApp(ctx, appType, name string, websiteID int64, version string) (*App, error)`
- `UpdateAppVersion(ctx, id int64, version string, updateAvailable bool) error`
- `DeleteApp(ctx, id int64) error`
- `MarkUpdateAvailable(ctx, id int64, available bool) error`

**`internal/api/apps.go`** — REST handlers:
- `GET /api/v1/websites/:id/apps` — list installed apps
- `POST /api/v1/websites/:id/apps/install` — install one-click app
  - Body: `{ "app_type": "wordpress", "admin_username": "admin", "admin_password": "...", "db_name": "wp_site" }`
  - Supported: `wordpress`, `laravel`, `nextjs`, `ghost`, `drupal`, `joomla`
  - Call agent `apps.install` with website path, app type, and credentials
  - Insert into `apps` table
- `POST /api/v1/websites/:id/apps/:app_id/update` — update app
  - Check for updates via agent `apps.check-update`
  - If available, mandatory backup first, then agent `apps.update`

**`internal/agent/apps/installer.go`**:
```go
func HandleAppInstall(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleAppUpdate(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleAppCheckUpdate(ctx context.Context, params json.RawMessage) (interface{}, error)
```

Implementation details:
- **WordPress**: Download latest WordPress ZIP, extract to document root, create `wp-config.php` with DB credentials, run WordPress CLI install (if available) or SQL init
- **Laravel**: Composer create-project, set up `.env`, run migrations
- **Next.js**: Node.js `npx create-next-app`, build and start
- Each installer sets file ownership to website's Linux user

**Register agent methods in `cmd/agent/main.go`**.

**Update `internal/api/router.go`** with app routes (already documented).

---

## Step 5.2: Git Deploy

### Backend

**No DB queries needed** — git config stored as JSON in `website.git_config` column (add via migration).

**Migration `002_git_config.sql`**:
```sql
ALTER TABLE websites ADD COLUMN git_config JSON;
```

**`internal/api/git.go`** — REST handlers:
- `GET /api/v1/websites/:id/git` — get git deploy config
- `POST /api/v1/websites/:id/git` — setup git deploy
  - Body: `{ "repo_url": "https://github.com/user/repo", "branch": "main", "deploy_key": "...", "auto_deploy": true }`
  - Validate repo URL format
  - Store in `website.git_config`
  - Call agent `git.setup` to clone repo
- `POST /api/v1/websites/:id/git/pull` — trigger git pull
  - Call agent `git.pull`
- `GET /api/v1/websites/:id/git/webhook` — get webhook URL
  - Return URL: `https://panel.domain/api/v1/webhooks/git/:token`

**`internal/api/git_webhook.go`** — Webhook receiver:
- `POST /api/v1/webhooks/git/:token` — webhook receiver
  - Verify HMAC signature if secret configured
  - Parse payload (GitHub/GitLab)
  - Find matching website by token
  - Trigger agent `git.pull`
  - Return 200 OK immediately (async pull)

**`internal/agent/git/git.go`**:
```go
func HandleGitSetup(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleGitPull(ctx context.Context, params json.RawMessage) (interface{}, error)
```

Implementation details:
- **Setup**: Clone repo into website document root, set ownership
- **Pull**: `git fetch && git reset --hard origin/{branch}`, run post-deploy hooks if configured
- **Webhook token**: Generate random 32-char hex token per website, store in git_config

**Update `internal/api/router.go`** with git routes.

---

## Step 5.3: Webmail

### Backend

**`internal/api/webmail.go`** — REST handlers:
- `GET /api/v1/webmail/status` — check if Roundcube is installed
- `POST /api/v1/webmail/install` — install Roundcube
  - Call agent `webmail.install`
- `POST /api/v1/webmail/uninstall` — uninstall Roundcube
  - Call agent `webmail.uninstall`
- `GET /api/v1/webmail/url` — get webmail URL

**`internal/agent/email/webmail.go`**:
```go
func HandleWebmailInstall(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleWebmailUninstall(ctx context.Context, params json.RawMessage) (interface{}, error)
```

Implementation details:
- **Install**: Download Roundcube, extract to `/var/www/webmail/`, configure DB connection, create tables
- **Uninstall**: Remove `/var/www/webmail/`, drop Roundcube DB
- **URL**: Return `https://webmail.{panel-domain}` or `/webmail`

No DB table needed — webmail is a system-wide service, not per-website.

**Update `internal/api/router.go`** with webmail routes.

---

## Step 5.4: Terminal Recording

### Backend

**No DB changes needed** — recordings stored as files at `/var/log/juvia/terminal/`.

**`internal/api/terminal.go`** — REST handlers:
- `POST /api/v1/terminal/session` — create PTY session (returns session ID)
  - WebSocket upgrade for interactive terminal
  - Call agent `terminal.start` to create PTY
  - Stream I/O via WebSocket
- `GET /api/v1/terminal/sessions` — list active sessions
- `DELETE /api/v1/terminal/sessions/:id` — close session
- `GET /api/v1/terminal/recordings` — list session recordings
  - Read `/var/log/juvia/terminal/` directory
  - Return metadata: filename, timestamp, user, duration
- `GET /api/v1/terminal/recordings/:id` — get recording playback data
  - Read `.cast` file and return JSON

**`internal/agent/terminal/record.go`**:
```go
func HandleTerminalStart(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleTerminalStop(ctx context.Context, params json.RawMessage) (interface{}, error)
```

Implementation details:
- **Recording format**: asciinema `.cast` format (JSON lines)
- **File naming**: `YYYY-MM-DD_HH-MM-SS_user_sessionid.cast`
- **Agent**: Uses `script` command or Go PTY library to capture terminal I/O
- **Playback**: Frontend parses `.cast` format and replays character-by-character

**Update `cmd/agent/main.go`** with terminal methods.

### Frontend

**`web/src/pages/Terminal/Recordings.tsx`**:
- Route: `/terminal/recordings`
- Table of recordings with date, user, duration
- Click to open playback player
- Player shows terminal output at original speed with play/pause/scrubber

---

## Step 5.5: Self-Updater

### Backend

**`internal/updater/updater.go`**:
```go
func CheckForUpdates(ctx context.Context) (*UpdateInfo, error)
func DownloadUpdate(ctx context.Context, url string) (string, error)
func VerifyGPGSignature(binaryPath, sigPath string) error
func ApplyUpdate(binaryPath string) error
func Rollback() error
```

**`internal/api/updater.go`** — REST handlers:
- `GET /api/v1/updates/check` — check for updates
  - Fetch latest release from update server (or local version file)
  - Return current version, latest version, changelog
- `POST /api/v1/updates/download` — download update
  - Download binary to `/tmp/juvia-update-{version}`
  - Download GPG signature
- `POST /api/v1/updates/apply` — apply update
  - Verify GPG signature
  - Backup current binary to `.backup`
  - Atomic rename new binary
  - Schedule restart
- `POST /api/v1/updates/rollback` — rollback
  - Restore `.backup` binary
  - Restart

Implementation details:
- **Version source**: GitHub releases API or self-hosted update.json
- **GPG verification**: Use embedded public key, verify detached signature
- **Atomic swap**: Write to `.new`, verify, rename to final (Linux rename is atomic)
- **Restart**: Send SIGTERM to self, systemd will restart

**Update `internal/api/router.go`** with updater routes.

---

## Step 5.6: Packaging

### Scripts

**`scripts/build-deb.sh`**:
- Build both binaries (`go build`)
- Create DEB package structure:
  ```
  juvia_{version}_amd64.deb
  ├── DEBIAN/control
  ├── DEBIAN/postinst
  ├── DEBIAN/prerm
  ├── usr/bin/juvia
  ├── usr/bin/juvia-agent
  ├── etc/systemd/system/juvia.service
  ├── etc/systemd/system/juvia-agent.service
  └── etc/juvia/
  ```
- Package metadata: name, version, description, dependencies (ufw, nginx, certbot)
- postinst: create user `juvia`, create directories, enable systemd services

**`scripts/install.sh`** — One-liner install:
```bash
#!/bin/bash
curl -fsSL https://get.juvia.io/install.sh | bash
```
- Detect architecture (amd64/arm64)
- Download latest .deb from GitHub releases
- `dpkg -i` the package
- Output success message with panel URL

No Go code changes needed for this step.

---

## Step 5.7: Command Palette

### Frontend

**`web/src/components/CommandPalette.tsx`**:
- Global component mounted in `Layout`
- Trigger: `Cmd+K` (Mac) / `Ctrl+K` (Windows/Linux)
- Features:
  - Search across all actions: "add website", "renew ssl", "open terminal"
  - Fuzzy matching
  - Recent commands
  - Keyboard navigation (↑↓ Enter Esc)

Actions registry:
```ts
const actions = [
  { id: 'add-website', label: 'Add Website', shortcut: '→ Websites', action: () => navigate('/websites/create') },
  { id: 'view-databases', label: 'View Databases', shortcut: '→ Databases', action: () => navigate('/databases') },
  { id: 'view-firewall', label: 'View Firewall Rules', shortcut: '→ Firewall', action: () => navigate('/firewall') },
  { id: 'create-backup', label: 'Create Backup', shortcut: '→ Backups', action: () => navigate('/backups') },
  { id: 'view-alerts', label: 'View Alerts', shortcut: '→ Alerts', action: () => navigate('/alerts') },
  // ... all routes
]
```

Mount in `Layout.tsx` as a modal overlay.

---

## Step 5.8: Settings Pages

### Backend

**`internal/api/settings.go`** — REST handlers:
- `GET /api/v1/settings` — get all settings
  - Query `settings` table, return as key-value map
- `PUT /api/v1/settings` — update settings
  - Body: `{ "panel_port": 8080, "hostname": "panel.example.com", ... }`
  - Validate each setting
  - Update `settings` table
- `GET /api/v1/audit-log` — get audit log entries
  - Pagination, filter by user, date range
- `GET /api/v1/settings/export` — export full config as JSON
  - Include: users, websites, databases, email, DNS, firewall rules, cron, backups
  - Redact passwords and secrets
- `POST /api/v1/settings/import` — import config from JSON
  - Validate structure
  - Import in transaction (rollback on error)

**Update `internal/api/router.go`** with settings routes.

### Frontend

**`web/src/pages/Settings/Settings.tsx`**:
- Route: `/settings`
- Sidebar tabs: General, Email, Security, Updates, Users, Backups, Notifications, Audit Log, Export/Import

**General tab:**
- Panel port, hostname, server name

**Email tab:**
- SMTP configuration (host, port, username, password, from address)

**Security tab:**
- 2FA enforcement toggle
- Session timeout
- Terminal IP restrictions

**Updates tab:**
- Current version, check for updates button
- Auto-update toggle
- Version pinning

**Users tab:**
- User list (reuse existing user management)

**Backups tab:**
- Remote storage config (S3/R2/B2 credentials)
- Default retention

**Notifications tab:**
- Alert thresholds (reuse alert settings)
- Email notification toggle

**Audit Log tab:**
- Table of all actions with user, IP, timestamp
- Filter by action type, date range

**Export/Import tab:**
- Download full config JSON
- Upload and import config JSON
- Warning about overwriting data

---

## Testing Strategy

### Go Tests
- `internal/db/app_queries_test.go`
- `internal/api/apps_test.go`
- `internal/api/git_test.go`
- `internal/api/git_webhook_test.go`
- `internal/api/webmail_test.go`
- `internal/api/terminal_test.go`
- `internal/api/updater_test.go`
- `internal/api/settings_test.go`
- `internal/agent/apps/installer_test.go`
- `internal/agent/git/git_test.go`
- `internal/agent/email/webmail_test.go`
- `internal/updater/updater_test.go`

### Frontend Tests
- `web/src/components/CommandPalette.test.tsx`
- `web/src/pages/Settings/Settings.test.tsx`
- `web/src/pages/Terminal/Recordings.test.tsx`

---

## File Summary

| Step | New Files | Modified Files |
|------|-----------|----------------|
| 5.1 | `internal/db/app_queries.go`, `internal/api/apps.go`, `internal/agent/apps/installer.go`, tests | `cmd/agent/main.go`, `internal/api/router.go` |
| 5.2 | `internal/db/migrations/002_git_config.sql`, `internal/api/git.go`, `internal/api/git_webhook.go`, `internal/agent/git/git.go`, tests | `cmd/agent/main.go`, `internal/api/router.go` |
| 5.3 | `internal/api/webmail.go`, `internal/agent/email/webmail.go`, tests | `cmd/agent/main.go`, `internal/api/router.go` |
| 5.4 | `internal/api/terminal.go`, `internal/agent/terminal/record.go`, `web/src/pages/Terminal/Recordings.tsx`, tests | `cmd/agent/main.go`, `internal/api/router.go`, `web/src/App.tsx` |
| 5.5 | `internal/updater/updater.go`, `internal/api/updater.go`, tests | `internal/api/router.go` |
| 5.6 | `scripts/build-deb.sh`, `scripts/install.sh` | — |
| 5.7 | `web/src/components/CommandPalette.tsx`, tests | `web/src/components/layout/Layout.tsx` |
| 5.8 | `internal/api/settings.go`, `web/src/pages/Settings/Settings.tsx`, tests | `internal/api/router.go`, `web/src/App.tsx` |

---

## Success Criteria

Phase 5 is complete when:
1. Can install WordPress on a website via one-click
2. Can set up git deploy with webhook auto-pull
3. Webmail (Roundcube) can be installed and accessed
4. Terminal sessions are recorded and can be played back
5. Panel can check for updates and self-update
6. `.deb` package builds successfully
7. Command palette works with Cmd+K
8. Settings pages cover all configuration areas
9. All tests pass (`go test ./...`)
10. Frontend builds without errors (`npm run build`)

---

## Important Notes

- **No schema changes needed for most steps** — `apps` and `settings` tables already exist. Only `002_git_config.sql` migration needed.
- **WordPress installer** is the priority — implement WordPress first, others are variations.
- **Git webhook** must verify HMAC signature to prevent unauthorized deploys.
- **Self-updater** must verify GPG signatures before applying updates.
- **Terminal recordings** use asciinema `.cast` format for compatibility.
- **Settings import/export** must redact all passwords and secrets.
- **Plain-English UI** — Every technical term must have a plain-English explanation.
