# DB ↔ Agent Consistency Audit & Fix Plan

## Problem

Several API handlers write state to the SQLite database but never propagate those changes to the privileged agent. This causes the panel to show "success" while the actual server configuration remains unchanged — the exact inconsistency that produced the DNS nameserver/glue-record issue.

## Goal

Make every mutating API operation that affects server configuration either:
1. Call the agent **before** committing to the DB (validate + apply, then persist), or
2. Call the agent **after** committing and roll back on failure.

Operations that are purely panel state (users, sessions, alert acks, audit log) do not require agent calls.

---

## Findings

### Critical — DB write with no agent call at all

| File | Handler | DB Operation | Missing Agent Action | Impact |
|------|---------|--------------|---------------------|--------|
| `internal/api/websites.go` | `updateWebsiteHandler` | `UpdateWebsite` (PHP/web server) | No `website.update` call | User changes PHP version / web server; config files unchanged |
| `internal/api/websites.go` | `changePHPHandler` | `UpdateWebsite` (PHP only) | No `website.update` call | PHP version changed in UI but PHP-FPM pool + Nginx config unchanged |
| `internal/api/websites.go` | `deleteWebsiteHandler` (trash) | `SoftDeleteWebsite` | No call to disable site | Website appears deleted but still serves traffic |
| `internal/api/email.go` | `updateMailboxHandler` | `UpdateMailbox` | No `email.update` call (method exists but is a no-op) | Mailbox quota/forward/display name never applied to Dovecot/Postfix |
| `internal/api/email.go` | `setCatchAllHandler` | `SetCatchAll` | No catch-all agent method exists | Catch-all stored in DB but Postfix never configured |
| `internal/api/cron.go` | `updateCronJobHandler` | `UpdateCronJob` | No `cron.update` method exists | Cron job changes never written to `/etc/cron.d/juvia` |
| `internal/api/cron.go` | `enableCronJobHandler` | `EnableCronJob` | No enable agent call | Job marked enabled but still commented/missing in cron file |
| `internal/api/cron.go` | `disableCronJobHandler` | `DisableCronJob` | No disable agent call | Job marked disabled but still runs from cron file |
| `internal/api/backups.go` | `createBackupScheduleHandler` | `CreateBackupSchedule` | No schedule agent method exists | Schedules stored but crontab/systemd timer never created |
| `internal/api/backups.go` | `deleteBackupScheduleHandler` | `DeleteBackupSchedule` | No schedule agent method exists | Schedule removed from DB but still runs on server |

### High — Agent called but with wrong or incomplete parameters

| File | Handler | Issue |
|------|---------|-------|
| `internal/api/databases.go` | `deleteDBUserHandler` | Calls `database.user.delete` with only `id`; agent expects `name`, `engine`, `username`, `host`. DB user deleted from SQLite but MySQL/PostgreSQL user remains. |

### Medium — DB committed before agent; inconsistency possible on agent failure

| File | Handler | Current Order | Risk |
|------|---------|---------------|------|
| `internal/api/backups.go` | `createBackupHandler` | DB create → agent | Backup row exists with "pending" then "failed"; acceptable but should be created after agent starts or status handled cleanly |
| `internal/api/cron.go` | `createCronJobHandler` | DB create → agent | Row deleted on failure, but race/bug could leave orphan |
| `internal/api/domains.go` | `createDomainHandler` | DB create → agent | Domain exists in DB but not in Nginx/BIND if agent fails |
| `internal/api/dns.go` | `createZoneHandler` | DB create → agent | Zone exists in DB but not in BIND if agent fails |

### Already correct or panel-only

- `firewall.go`, `ssh.go`, `server.go`, `services.go`, `ssl.go`, `apps.go`, `git.go`, `webmail.go`, `files.go`, `terminal.go`, `auth.go`, `alerts.go`, `users.go`, `updater.go` are either agent-first/agent-always or pure panel state.
- The nameserver brand-domain fix (`dns.brand.setup`) added in the previous commit resolves the original inconsistency.

---

## Required Agent Changes

### 1. `internal/agent/cron/manager.go`

- Add `HandleCronUpdate` — replace the cron line identified by `id` with the new schedule/command/run_as.
- Add `HandleCronEnable` / `HandleCronDisable` — comment/uncomment the line identified by `id` in `/etc/cron.d/juvia`.
- Fix `HandleCronCreate` to respect `run_as` (currently hard-coded to `root`) and call `reloadCron()` after writes.
- Fix `HandleCronDelete` to call `reloadCron()` after deletion.

### 2. `internal/agent/email/postfix.go`

- Implement `HandleMailboxUpdate` to apply quota, forward, and display-name changes to the Dovecot userdb / Postfix aliases.
- Add `HandleCatchAllSet` / `HandleCatchAllDelete` to manage catch-all entries in `/etc/postfix/virtual_aliases` or a dedicated catch-all map.

### 3. `internal/agent/backup/backup.go`

- Add `HandleBackupScheduleCreate` — write a cron entry or systemd timer for the backup schedule.
- Add `HandleBackupScheduleDelete` — remove the corresponding cron/timer entry.

### 4. `cmd/agent/main.go`

Register new methods:
- `cron.update`
- `cron.enable`
- `cron.disable`
- `email.catchall.set`
- `email.catchall.delete`
- `backup.schedule.create`
- `backup.schedule.delete`

---

## Required API Changes

### 1. `internal/api/websites.go`

- `updateWebsiteHandler`: after DB update, call `website.update` with the new `php_version`, `web_server`, and current domain/extra domains. On agent error, return error and optionally roll back DB.
- `changePHPHandler`: after DB update, call `website.update` with the new PHP version. On agent error, roll back DB change.
- `deleteWebsiteHandler` (trash): call `website.suspend` before soft-delete so the site stops serving.

### 2. `internal/api/email.go`

- `updateMailboxHandler`: call `email.update` before or after DB update; implement real update logic in agent first.
- `setCatchAllHandler`: call `email.catchall.set` after DB update.

### 3. `internal/api/cron.go`

- `updateCronJobHandler`: call `cron.update` after DB update.
- `enableCronJobHandler`: call `cron.enable` after DB update.
- `disableCronJobHandler`: call `cron.disable` after DB update.

### 4. `internal/api/backups.go`

- `createBackupScheduleHandler`: call `backup.schedule.create` after DB insert.
- `deleteBackupScheduleHandler`: call `backup.schedule.delete` before DB delete.

### 5. `internal/api/databases.go`

- `deleteDBUserHandler`: pass `name`, `engine`, `username`, and `host` to `database.user.delete`, matching the agent's expected parameters.

---

## Implementation Order

1. **Cron agent fixes** (`cron/manager.go`) + API wiring (`cron.go`)
2. **Email agent fixes** (`email/postfix.go`) + API wiring (`email.go`)
3. **Website update/trash agent calls** (`websites.go`)
4. **Backup schedule agent methods** (`backup/backup.go`) + API wiring (`backups.go`)
5. **Database user delete parameter fix** (`databases.go`)
6. **Registration** in `cmd/agent/main.go`
7. **Tests** for every new/modified agent and API handler
8. **Build + test + deploy**

Each bullet should be a separate atomic commit per project rules.

---

## Testing Strategy

- Unit tests for each new agent handler (cron update/enable/disable, email update/catch-all, backup schedule create/delete).
- API handler tests verifying that the correct agent method is called and that agent failures produce errors rather than silent DB-only success.
- Run `go build ./...` and `go test ./...` before every commit and deployment.

---

## Files to Modify

- `cmd/agent/main.go`
- `internal/agent/cron/manager.go`
- `internal/agent/email/postfix.go`
- `internal/agent/backup/backup.go`
- `internal/api/websites.go`
- `internal/api/email.go`
- `internal/api/cron.go`
- `internal/api/backups.go`
- `internal/api/databases.go`
- Corresponding `*_test.go` files

---

## Success Criteria

- Every mutating endpoint that affects server configuration results in an actual agent-side change.
- If the agent operation fails, the API returns an error and the DB is not left in a falsely-successful state.
- `go build ./...` and `go test ./...` pass on every commit.
