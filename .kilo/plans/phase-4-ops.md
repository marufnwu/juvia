# Phase 4 — Ops Implementation Plan

**Goal:** Firewall, backups, cron, logs, alerts. Success = can set firewall rules, run backups, create cron jobs, view logs, and receive alerts.

**Rule:** Follow steps in order. Do not skip ahead. Each step is a single atomic commit that passes `go build ./...` and `go test ./...`.

---

## Step 4.1: Firewall API + Agent Methods

### Backend

**`internal/db/firewall_queries.go`** — SQLite operations:
- `CreateFirewallRule(ctx, name, description, action, port, protocol, source) (*FirewallRule, error)`
- `GetFirewallRuleByID(ctx, id) (*FirewallRule, error)`
- `ListFirewallRules(ctx, page, limit) ([]FirewallRule, int, error)`
- `UpdateFirewallRule(ctx, id, name, description, action, port, protocol, source) error`
- `DeleteFirewallRule(ctx, id) error`

**`internal/api/firewall.go`** — REST handlers:
- `GET /api/v1/firewall/rules` — list rules with pagination
- `POST /api/v1/firewall/rules` — create rule
  - Body: `{ "name": "SSH", "description": "Allow SSH access", "action": "allow", "port": "22", "protocol": "tcp", "source": "any" }`
  - Validate action is `allow` or `deny`
  - Validate protocol is `tcp`, `udp`, or `both`
  - Call agent `firewall.rule.create`
  - Insert into DB, audit log
- `GET /api/v1/firewall/rules/:id` — get rule
- `PUT /api/v1/firewall/rules/:id` — update rule
  - Call agent `firewall.rule.update`
- `DELETE /api/v1/firewall/rules/:id` — delete rule
  - Call agent `firewall.rule.delete`

**`internal/agent/firewall/ufw.go`**:
```go
func HandleRuleCreate(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleRuleUpdate(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleRuleDelete(ctx context.Context, params json.RawMessage) (interface{}, error)
```

Implementation details:
- **Create rule**: `ufw {action} {port}/{protocol} from {source} comment '{name}'`
  - If source is `any`, omit `from` clause
  - Reload UFW
- **Delete rule**: `ufw delete {action} {port}/{protocol} from {source}`
- **UFW not installed**: Return clear error message
- **Plain-English descriptions**: Each rule stored in DB has a user-provided description

**Register agent methods in `cmd/agent/main.go`**:
```go
server.RegisterMethod("firewall.rule.create", firewall.HandleRuleCreate)
server.RegisterMethod("firewall.rule.update", firewall.HandleRuleUpdate)
server.RegisterMethod("firewall.rule.delete", firewall.HandleRuleDelete)
```

**Update `internal/api/router.go`** with firewall routes.

---

## Step 4.2: Backup API + Agent Methods

### Backend

**`internal/db/backup_queries.go`** — SQLite operations:
- `CreateBackup(ctx, websiteID, backupType, storage, path) (*Backup, error)`
- `GetBackupByID(ctx, id) (*Backup, error)`
- `ListBackups(ctx, websiteID, page, limit) ([]Backup, int, error)`
- `UpdateBackupStatus(ctx, id, status, sizeBytes, checksum) error`
- `DeleteBackup(ctx, id) error`
- `CreateBackupSchedule(ctx, websiteID, schedule, retentionDays, storage) (*BackupSchedule, error)`
- `ListBackupSchedules(ctx, websiteID) ([]BackupSchedule, error)`
- `DeleteBackupSchedule(ctx, id) error`

**`internal/api/backups.go`** — REST handlers:
- `GET /api/v1/backups` — list backups with pagination
- `POST /api/v1/backups` — create manual backup
  - Body: `{ "website_id": 1, "type": "full", "storage": "local" }`
  - Types: `full`, `files`, `database`
  - Storage: `local`, `s3`
  - Call agent `backup.create`
  - Insert DB record with status `in_progress`
  - Poll or wait for agent response, update status
- `GET /api/v1/backups/:id` — get backup details
- `POST /api/v1/backups/:id/restore` — restore backup
  - Call agent `backup.restore`
  - Warning: this is destructive
- `DELETE /api/v1/backups/:id` — delete backup
  - Call agent `backup.delete` to remove file
- `GET /api/v1/backup-schedules` — list schedules
- `POST /api/v1/backup-schedules` — create schedule
  - Body: `{ "website_id": 1, "schedule": "0 3 * * *", "retention_days": 7, "storage": "local" }`
- `DELETE /api/v1/backup-schedules/:id` — delete schedule

**`internal/agent/backup/backup.go`**:
```go
func HandleBackupCreate(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleBackupRestore(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleBackupDelete(ctx context.Context, params json.RawMessage) (interface{}, error)
```

Implementation details:
- **Create backup**: 
  - `type=full`: Tar website files + mysqldump database
  - `type=files`: Tar website files only
  - `type=database`: mysqldump only
  - Destination: `/var/backups/juvia/{website_id}/{timestamp}.tar.gz`
  - Compute SHA-256 checksum
  - Return path and checksum
- **Restore**: Extract tar to temp, verify, move to document root
- **Delete**: Remove file from disk
- **Local storage only** for Phase 4. S3/R2/B2 in Phase 5.

**Register agent methods in `cmd/agent/main.go`**:
```go
server.RegisterMethod("backup.create", backup.HandleBackupCreate)
server.RegisterMethod("backup.restore", backup.HandleBackupRestore)
server.RegisterMethod("backup.delete", backup.HandleBackupDelete)
```

**Update `internal/api/router.go`** with backup routes.

---

## Step 4.3: Cron API + Agent Methods

### Backend

**`internal/db/cron_queries.go`** — SQLite operations:
- `CreateCronJob(ctx, userID, schedule, command, runAs, jobType string) (*CronJob, error)`
- `GetCronJobByID(ctx, id) (*CronJob, error)`
- `ListCronJobs(ctx, userID, page, limit) ([]CronJob, int, error)`
- `UpdateCronJob(ctx, id, schedule, command, runAs) error`
- `DeleteCronJob(ctx, id) error`
- `EnableCronJob(ctx, id) error`
- `DisableCronJob(ctx, id) error`
- `CreateCronJobLog(ctx, cronJobID, durationMs int64, status, output string) (*CronJobLog, error)`
- `ListCronJobLogs(ctx, cronJobID, limit int) ([]CronJobLog, error)`

**`internal/api/cron.go`** — REST handlers:
- `GET /api/v1/cron` — list cron jobs
- `POST /api/v1/cron` — create cron job
  - Body: `{ "schedule": "0 3 * * *", "command": "/usr/local/bin/backup.sh", "run_as": "root", "type": "shell" }`
  - Types: `shell`, `url`, `php`
  - Validate cron expression format (5 fields)
  - Call agent `cron.create`
  - Insert into DB
- `GET /api/v1/cron/:id` — get cron job with last 10 logs
- `PUT /api/v1/cron/:id` — update cron job
- `DELETE /api/v1/cron/:id` — delete cron job
- `POST /api/v1/cron/:id/enable` — enable
- `POST /api/v1/cron/:id/disable` — disable
- `GET /api/v1/cron/:id/logs` — get execution history

**`internal/agent/cron/manager.go`**:
```go
func HandleCronCreate(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleCronDelete(ctx context.Context, params json.RawMessage) (interface{}, error)
func HandleCronRun(ctx context.Context, params json.RawMessage) (interface{}, error)
```

Implementation details:
- **Create**: Write to system crontab at `/etc/cron.d/juvia`
  - Format: `{schedule} {run_as} {command} # juvia:{id}`
- **Delete**: Remove line from `/etc/cron.d/juvia`
- **Run now**: Execute command via `su - {run_as} -c '{command}'`, capture output
- **Execution logging**: Agent writes to cron_job_logs table via direct DB or panel callback

**Register agent methods in `cmd/agent/main.go`**:
```go
server.RegisterMethod("cron.create", cron.HandleCronCreate)
server.RegisterMethod("cron.delete", cron.HandleCronDelete)
server.RegisterMethod("cron.run", cron.HandleCronRun)
```

**Update `internal/api/router.go`** with cron routes.

---

## Step 4.4: Logs API

### Backend

**`internal/api/logs.go`** — REST handlers:
- `GET /api/v1/websites/:id/logs/access` — get website access log
  - Query: `?lines=100`
  - Read from `/var/log/juvia/nginx/{domain}.access.log`
- `GET /api/v1/websites/:id/logs/error` — get website error log
  - Read from `/var/log/juvia/nginx/{domain}.error.log`
- `GET /api/v1/logs/system` — get system logs
  - Read from `/var/log/juvia/panel.log` or `/var/log/juvia/agent.log`
- `GET /api/v1/logs/download` — download log file

No agent methods needed — panel reads log files directly (must use `safePath()` validation).

**Update `internal/api/router.go`** with log routes.

---

## Step 4.5: Logs Frontend

### React Components

**`web/src/pages/Logs/LogViewer.tsx`**:
- Route: `/websites/:id/logs`
- Tabs: Access Log, Error Log
- Features:
  - Live tail (poll every 5 seconds)
  - Line count selector (50, 100, 500, 1000)
  - Filter/search within log
  - Download button
  - Plain-English: "Access Log · Records every visit to your website"

**`web/src/App.tsx`** — Add route:
```tsx
<Route path="websites/:id/logs" element={<LogViewer />} />
```

---

## Step 4.6: Alert Engine

### Backend

**`internal/db/alert_queries.go`** — SQLite operations:
- `CreateAlert(ctx, alertType, severity, message string, resourceID *int64, resourceType string) (*Alert, error)`
- `ListAlerts(ctx, acknowledged bool, page, limit int) ([]Alert, int, error)`
- `AcknowledgeAlert(ctx, id int64) error`
- `DeleteAlert(ctx, id int64) error`

**`internal/api/alerts.go`** — REST handlers:
- `GET /api/v1/alerts` — list alerts
  - Query: `?acknowledged=false`
- `POST /api/v1/alerts/:id/acknowledge` — acknowledge alert
- `DELETE /api/v1/alerts/:id` — delete alert
- `GET /api/v1/alerts/settings` — get alert settings
- `PUT /api/v1/alerts/settings` — update alert settings

**`internal/alerts/engine.go`** — Background checker:
```go
type Engine struct {
    db     *db.DB
    client *socket.Client
    interval time.Duration
}

func (e *Engine) Start(ctx context.Context)
func (e *Engine) checkWebsites()
func (e *Engine) checkSSLExpiry()
func (e *Engine) checkDiskUsage()
func (e *Engine) checkServices()
```

Checks every 60 seconds:
- **Website down**: HTTP GET to each website, alert on 5xx or timeout > 5s
- **Website slow**: Response time > 3s
- **SSL expiring**: Check `ssl_expiry` date, alert if < 14 days
- **Disk > 80%**: Use metrics collector
- **Service down**: Check nginx, mysql, postfix, dovecot via `systemctl is-active`
- **Backup overdue**: No backup in 7 days

Alert delivery:
- Insert into `alerts` table
- Emit `EventAlertFired` on event bus (for WebSocket push)

**Start engine in `cmd/panel/main.go`** alongside existing services.

**Update `internal/api/router.go`** with alert routes.

---

## Frontend Pages for Ops

### Firewall (`web/src/pages/Firewall/Firewall.tsx`)
- Route: `/firewall`
- Table of rules with plain-English descriptions
- Create rule modal (name, description, action, port, protocol, source)
- Delete action per row
- "Firewall Rule · Controls which network traffic can reach your server"

### Backups (`web/src/pages/Backups/Backups.tsx`)
- Route: `/backups`
- List of backups with status, size, verified timestamp
- "Create Backup" button (manual)
- Restore and delete actions
- Schedules tab for automated backups
- "Backup · A snapshot of your website you can restore if something goes wrong"

### Cron Jobs (`web/src/pages/Cron/CronJobs.tsx`)
- Route: `/cron`
- Table of jobs with schedule, command, last run, status
- Create job modal with visual builder (every day at time)
- Enable/disable toggle
- View logs button per job
- "Cron Job · A task that runs automatically on a schedule"

### Alerts (`web/src/pages/Alerts/Alerts.tsx`)
- Route: `/alerts` (or bell icon dropdown)
- List of alerts with severity badges
- Acknowledge button
- Filter by acknowledged/unacknowledged
- "Alert · A notification about something that needs your attention"

**Update `web/src/App.tsx`** with all new routes.

---

## Metrics Endpoint Enhancement

**`internal/api/metrics.go`** (new file or update existing):
- `GET /api/v1/metrics/current` — return current metrics from collector
- `GET /api/v1/metrics/history` — placeholder for VictoriaMetrics integration (Phase 5)

**Update `internal/api/router.go`** with metrics routes.

---

## Testing Strategy

### Go Tests
- `internal/db/firewall_queries_test.go`
- `internal/db/backup_queries_test.go`
- `internal/db/cron_queries_test.go`
- `internal/db/alert_queries_test.go`
- `internal/api/firewall_test.go`
- `internal/api/backups_test.go`
- `internal/api/cron_test.go`
- `internal/api/alerts_test.go`
- `internal/agent/firewall/ufw_test.go`
- `internal/agent/backup/backup_test.go`
- `internal/agent/cron/manager_test.go`
- `internal/alerts/engine_test.go`

### Frontend Tests
- `web/src/pages/Firewall/Firewall.test.tsx`
- `web/src/pages/Backups/Backups.test.tsx`
- `web/src/pages/Cron/CronJobs.test.tsx`
- `web/src/pages/Alerts/Alerts.test.tsx`
- `web/src/pages/Logs/LogViewer.test.tsx`

---

## File Summary

| Step | New Files | Modified Files |
|------|-----------|----------------|
| 4.1 | `internal/db/firewall_queries.go`, `internal/api/firewall.go`, `internal/agent/firewall/ufw.go`, tests | `cmd/agent/main.go`, `internal/api/router.go` |
| 4.2 | `internal/db/backup_queries.go`, `internal/api/backups.go`, `internal/agent/backup/backup.go`, tests | `cmd/agent/main.go`, `internal/api/router.go` |
| 4.3 | `internal/db/cron_queries.go`, `internal/api/cron.go`, `internal/agent/cron/manager.go`, tests | `cmd/agent/main.go`, `internal/api/router.go` |
| 4.4 | `internal/api/logs.go`, tests | `internal/api/router.go` |
| 4.5 | `web/src/pages/Logs/LogViewer.tsx`, `web/src/pages/Logs/LogViewer.test.tsx` | `web/src/App.tsx` |
| 4.6 | `internal/db/alert_queries.go`, `internal/api/alerts.go`, `internal/alerts/engine.go`, tests | `cmd/panel/main.go`, `internal/api/router.go` |
| Frontend | `web/src/pages/Firewall/Firewall.tsx`, `web/src/pages/Backups/Backups.tsx`, `web/src/pages/Cron/CronJobs.tsx`, `web/src/pages/Alerts/Alerts.tsx`, tests | `web/src/App.tsx` |

---

## Success Criteria

Phase 4 is complete when:
1. Firewall rules can be created/deleted via UI and applied via UFW
2. Backups can be created manually and scheduled automatically
3. Cron jobs can be created with visual builder and execution history viewed
4. Website access/error logs can be viewed live in the browser
5. Alert engine detects website down, SSL expiry, disk usage, and service health
6. Alerts appear in the UI and can be acknowledged
7. All tests pass (`go test ./...`)
8. Frontend builds without errors (`npm run build`)

---

## Important Notes

- **No schema changes needed** — all tables (firewall_rules, backups, backup_schedules, cron_jobs, cron_job_logs, alerts) already exist in `001_initial_schema.sql`.
- **UFW commands must be validated** — never allow arbitrary command injection through rule parameters.
- **Backup paths** — use `safePath()` for all file operations. Backups stored at `/var/backups/juvia/`.
- **Cron security** — validate cron expressions strictly (5 fields). Reject shell metacharacters in commands.
- **Alert deduplication** — don't create duplicate alerts for the same issue within 1 hour.
- **Plain-English UI** — Every technical term must have a plain-English explanation.
- **Agent runs as root** — can execute UFW, crontab, tar, systemctl commands.
