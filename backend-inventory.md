# Juvia Backend Feature Inventory

Generated from source code audit of `D:\Server panel`.

---

## Backend Module Matrix

| Module | Description | Key Files | Dependencies |
|---|---|---|---|
| **Panel Entry** | HTTP server bootstrap, DB, JWT, agent socket, event bus, static UI | `cmd/panel/main.go` | SQLite, Gin, agent socket, embedded React |
| **Agent Entry** | Root Unix-socket server registering all privileged JSON-RPC methods | `cmd/agent/main.go` | `internal/agent/*`, `internal/socket` |
| **Router** | Gin route registration for `/api/v1/*` and static asset fallthrough | `internal/api/router.go` | All API handler packages |
| **Middleware** | Request logging, CORS (dev), JWT Bearer auth, admin role gate | `internal/api/middleware.go`, `helpers.go` | `internal/auth` |
| **Auth API** | Login, refresh, logout, me, 2FA enable/verify, session revocation | `internal/api/auth.go` | JWT, session store, bcrypt, TOTP |
| **Setup API** | First-run wizard status and initial admin creation | `internal/api/setup.go` | bcrypt |
| **Users API** | Admin user CRUD | `internal/api/users.go` | bcrypt, `users` table |
| **Websites API** | Website lifecycle, DNS/files/apps/git shortcuts, logs | `internal/api/websites.go`, `dns.go`, `files.go`, `ssl.go`, `logs.go` | Agent website/nginx/ssl/dns/files |
| **Databases API** | MySQL/PostgreSQL DB and user management, export, query | `internal/api/databases.go` | Agent database |
| **Email API** | Mailboxes, aliases, forwarders, catch-all, deliverability | `internal/api/email.go` | Agent email |
| **Firewall API** | UFW rule CRUD | `internal/api/firewall.go` | Agent firewall |
| **Backups API** | Backup and schedule CRUD/restore | `internal/api/backups.go` | Agent backup |
| **Cron API** | Cron job CRUD, enable/disable, logs | `internal/api/cron.go` | Agent cron |
| **Alerts API** | Alert list/ack/delete and settings | `internal/api/alerts.go` | `alerts` table, settings |
| **Metrics API** | Current metrics; history is stubbed | `internal/api/metrics.go` | `internal/metrics` |
| **Apps API** | One-click app list/install/update per website | `internal/api/apps.go` | Agent apps |
| **Git API** | Git deploy config, pull, webhook URL | `internal/api/git.go`, `git_webhook.go` | Agent git |
| **Webmail API** | Roundcube status/install/uninstall/URL | `internal/api/webmail.go` | Agent email/webmail |
| **Terminal API** | PTY session create/list/close, recording list/playback | `internal/api/terminal.go` | Agent terminal |
| **Updates API** | Check/download/apply/rollback self-updates | `internal/api/updater.go` | `internal/updater` |
| **Settings API** | Key/value settings, audit log, export/import | `internal/api/settings.go` | settings, audit_log |
| **Services API** | System service status/restart registry | `internal/api/services.go` | Agent services |
| **Database Layer** | SQLite models, migrations, CRUD query helpers | `internal/db/*` | modernc.org/sqlite |
| **Auth Layer** | JWT, sessions, bcrypt, TOTP | `internal/auth/*` | golang-jwt, pquerna/otp, x/crypto |
| **Agent Layer** | JSON-RPC handlers for privileged OS operations | `internal/agent/*` | system tools (nginx, certbot, ufw, mysql, etc.) |
| **Tasks** | Async task creation/progress/complete/fail | `internal/tasks/runner.go` | `tasks` table |
| **Events** | In-memory pub-sub for metrics/task/alert events | `internal/events/bus.go` | WebSocket server |
| **WebSocket** | `/ws/v1/metrics`, `/ws/v1/tasks/:taskId` streaming | `internal/ws/server.go` | events, metrics |
| **Metrics** | Linux `/proc` metric collector | `internal/metrics/*` | `/proc` (Linux only) |
| **Alerts Engine** | Background 60s checks: disk, services, websites, SSL, backups | `internal/alerts/engine.go` | DB, metrics, events |
| **Updater** | HTTP download, GPG verify, binary swap/rollback | `internal/updater/updater.go` | systemctl |
| **Socket IPC** | JSON-RPC over Unix socket client/server | `internal/socket/*` | net/unix |
| **Utils** | `SafePath` path-traversal prevention | `internal/utils/safe.go` | - |
| **API Test Runner** | Stateful integration test harness for endpoints | `cmd/api-test-runner`, `internal/apitest/runner/*` | HTTP panel API |

---

## Backend Endpoint Matrix

Base path: `/api/v1`

| Module | Endpoint | Method | Auth | Controller | Service / Agent Method | Request Fields | Response Fields | Permission | Business Purpose |
|---|---|---|---|---|---|---|---|---|---|
| Health | `/health` | GET | None | `healthHandler` | `DB.Ping`, `AgentClient.IsOffline` | - | `status`, `version`, `uptime`, `sqlite`, `agent`, `victoria_metrics` | None | Panel + agent liveness |
| Setup | `/setup/status` | GET | None | `setupStatusHandler` | `COUNT(users)` | - | `setup_required` | None | First-run detection |
| Setup | `/setup/first-run` | POST | None | `firstRunHandler` | bcrypt + `INSERT users` | `username`, `password`, `email`, `server_name` | `message` | None | Create admin account |
| Auth | `/auth/login` | POST | None | `loginHandler` | JWT + SessionStore | `username`, `password`, `code` | `access_token`, `expires_in`, `user` | None | Authenticate, issue tokens |
| Auth | `/auth/refresh` | POST | Cookie | `refreshHandler` | SessionStore + JWT | `refresh_token` cookie | `access_token`, `expires_in` | None | Refresh access token |
| Auth | `/auth/logout` | POST | User | `logoutHandler` | SessionStore.RevokeAllUserSessions | - | `message` | Own sessions | Invalidate refresh sessions |
| Auth | `/auth/me` | GET | User | `meHandler` | DB users, SessionStore.ListUserSessions | - | `user`, `sessions`, `X-CSRF-Token` header | Own profile | Current user + sessions |
| Auth | `/auth/2fa/enable` | POST | User | `enable2FAHandler` | TOTPManager.GenerateSecret | - | `secret`, `uri` | Own account | Start TOTP setup |
| Auth | `/auth/2fa/verify` | POST | User | `verify2FAHandler` | TOTPManager.VerifyCode | `secret`, `code` | `recovery_codes` | Own account | Enable 2FA |
| Auth | `/auth/sessions/:id` | DELETE | User | `revokeSessionHandler` | SessionStore.RevokeSession | `id` (uri) | `message` | Own sessions | Revoke specific session |
| Users | `/users` | GET | Admin | `listUsersHandler` | `ListUsers` query | - | `[]User` | admin | List panel users |
| Users | `/users` | POST | Admin | `createUserHandler` | bcrypt + `INSERT users` | `username`, `password`, `role`, `email` | `id` | admin | Create user |
| Users | `/users/:id` | GET | Admin | `getUserHandler` | `SELECT users` | `id` | `User` | admin | Get user |
| Users | `/users/:id` | PUT | Admin | `updateUserHandler` | `UPDATE users` | `role`, `email`, `active` | `message` | admin | Update user |
| Users | `/users/:id` | DELETE | Admin | `deleteUserHandler` | `DELETE users` | `id` | `message` | admin | Delete user |
| Websites | `/websites` | GET | User | `listWebsitesHandler` | `ListWebsites` | `status`, `search`, `sort`, `order`, `page`, `limit` | `data[]`, `meta` | admin/user/per-site | List active sites |
| Websites | `/websites` | POST | User | `createWebsiteHandler` | `CreateWebsite` + agent `website.create` | `domain`, `php_version`, `web_server` | `id`, `domain`, `status`, `task_id` | admin/user | Create site + Linux user + DNS zone |
| Websites | `/websites/trash` | GET | User | `listTrashedWebsitesHandler` | `ListTrashedWebsites` | - | `[]Website` | admin/user | List soft-deleted sites |
| Websites | `/websites/trash/purge` | DELETE | User | `purgeTrashHandler` | `PurgeTrash` | - | `message` | admin/user | Delete trashed sites >30d |
| Websites | `/websites/:id` | GET | User | `getWebsiteHandler` | `GetWebsiteByID`, `ListDomainsByWebsite` | `id` | `website`, `domains` | admin/user/per-site | Site details |
| Websites | `/websites/:id` | PUT | User | `updateWebsiteHandler` | `UpdateWebsite` | `php_version`, `web_server` | `message` | admin/user/per-site | Update site config |
| Websites | `/websites/:id` | DELETE | User | `deleteWebsiteHandler` | `SoftDeleteWebsite` | `id` | `message` | admin/user/per-site | Move to trash |
| Websites | `/websites/:id/permanent` | DELETE | User | `permanentDeleteWebsiteHandler` | agent `website.delete` + `PermanentlyDeleteWebsite` | `id` | `message` | admin/user/per-site | Hard delete site files + DB |
| Websites | `/websites/:id/suspend` | POST | User | `suspendWebsiteHandler` | agent `website.suspend` + `SuspendWebsite` | `id` | `message` | admin/user/per-site | Suspend web+vhost |
| Websites | `/websites/:id/restore` | POST | User | `restoreWebsiteHandler` | `RestoreWebsite` | `id` | `message` | admin/user/per-site | Restore from trash |
| Websites | `/websites/:id/php` | PUT | User | `changePHPHandler` | `UpdateWebsite` | `php_version` | `message` | admin/user/per-site | Change PHP version |
| Websites | `/websites/:id/ssl/check` | GET | User | `sslCheckHandler` | agent `ssl.check` | `id` | `ready`, `message` | admin/user/per-site | Check cert readiness |
| Websites | `/websites/:id/ssl` | POST | User | `issueSSLHandler` | agent `ssl.issue` + `UpdateWebsiteSSL` | `id` | `message`, `expiry`, `cert_type`, `warning` | admin/user/per-site | Issue Let's Encrypt / self-signed |
| Websites | `/websites/:id/ssl-renew` | POST | User | `renewSSLHandler` | agent `ssl.renew` + `UpdateWebsiteSSL` | `id` | `message`, `expiry`, `cert_type` | admin/user/per-site | Renew certificate |
| Websites | `/websites/:id/ssl` | DELETE | User | `removeSSLHandler` | agent `ssl.remove` + `UpdateWebsiteSSL` | `id` | `message` | admin/user/per-site | Remove certificate |
| Websites | `/websites/:id/dns/records` | GET | User | `listDNSRecordsHandler` | `GetZoneByWebsite`, `ListDNSRecordsByZone` | `id` | `[]DNSRecord` | admin/user/per-site | List DNS records |
| Websites | `/websites/:id/dns/records` | POST | User | `createDNSRecordHandler` | `CreateDNSRecord` + agent `dns.zone.update` | `type`, `name`, `value`, `priority` | `DNSRecord` | admin/user/per-site | Add DNS record |
| Websites | `/websites/:id/dns/suggest` | POST | User | `suggestDNSHandler` | `ListDNSRecordsByZone` | `id` | suggestions[] | admin/user/per-site | Suggest missing records |
| DNS | `/dns/records/:id` | PUT | User | `updateDNSRecordHandler` | `UpdateDNSRecord` | `name`, `value`, `priority` | `message` | admin/user/per-site | Update DNS record |
| DNS | `/dns/records/:id` | DELETE | User | `deleteDNSRecordHandler` | `DeleteDNSRecord` | `id` | `message` | admin/user/per-site | Delete DNS record |
| Files | `/websites/:id/files` | GET | User | `listFilesHandler` | agent `files.list` | `path` query | `items[]` | admin/user/per-site | Browse files |
| Files | `/websites/:id/files/upload` | POST | User | `uploadFilesHandler` | agent `files.upload` | `path`, `file_name`, `content` | agent result | admin/user/per-site | Upload text file |
| Files | `/websites/:id/files/download` | GET | User | `downloadFilesHandler` | agent `files.download` | `path` query/body | `content`, `size` | admin/user/per-site | Download file |
| Files | `/websites/:id/files/rename` | PUT | User | `renameFilesHandler` | agent `files.rename` | `old_path`, `new_path` | agent result | admin/user/per-site | Rename file/folder |
| Files | `/websites/:id/files/delete` | DELETE | User | `deleteFilesHandler` | agent `files.delete` | `path` | agent result | admin/user/per-site | Delete file/folder |
| Files | `/websites/:id/files/extract` | POST | User | `extractFilesHandler` | agent `files.extract` | `path` | agent result | admin/user/per-site | Extract archive |
| Files | `/websites/:id/files/edit` | GET | User | `getFileContentHandler` | agent `files.download` | `path` | `content`, `size` | admin/user/per-site | Read file for editor |
| Files | `/websites/:id/files/edit` | PUT | User | `saveFileContentHandler` | agent `files.upload` | `path`, `content` | `message` | admin/user/per-site | Save edited file |
| Logs | `/websites/:id/logs/access` | GET | User | `accessLogHandler` | read `/var/log/juvia/nginx/{domain}.access.log` | `lines` | `lines[]`, `path` | admin/user/per-site | Site access log |
| Logs | `/websites/:id/logs/error` | GET | User | `errorLogHandler` | read `/var/log/juvia/nginx/{domain}.error.log` | `lines` | `lines[]`, `path` | admin/user/per-site | Site error log |
| Logs | `/logs/system` | GET | User | `systemLogsHandler` | read panel/agent log | `type`, `lines` | `lines[]`, `path` | admin/user | System logs |
| Databases | `/databases` | GET | User | `listDatabasesHandler` | `ListDatabases` | `engine`, `page`, `limit` | `data[]`, `meta` | admin/user | List DBs |
| Databases | `/databases` | POST | User | `createDatabaseHandler` | agent `database.create` + `CreateDatabase` | `name`, `engine` | `Database` | admin/user | Create DB |
| Databases | `/databases/:id` | GET | User | `getDatabaseHandler` | `GetDatabaseByID`, `ListDBUsers` | `id` | `database`, `users` | admin/user | DB details |
| Databases | `/databases/:id` | DELETE | User | `deleteDatabaseHandler` | agent `database.delete` + `DeleteDatabase` | `id` | `message` | admin/user | Delete DB |
| Databases | `/databases/:id/users` | POST | User | `createDBUserHandler` | agent `database.user.create` + `CreateDBUser` | `username`, `password`, `host` | `DBUser` | admin/user | Create DB user |
| Databases | `/databases/:id/users/:user_id` | DELETE | User | `deleteDBUserHandler` | agent `database.user.delete` + `DeleteDBUser` | `id`, `user_id` | `message` | admin/user | Delete DB user |
| Databases | `/databases/:id/export` | POST | User | `exportDatabaseHandler` | agent `database.export` | `id` | agent result | admin/user | Export to .sql |
| Databases | `/databases/:id/tables` | GET | User | `listTablesHandler` | agent `database.tables` | `id` | `tables[]` | admin/user | List tables |
| Databases | `/databases/:id/tables/:table/rows` | GET | User | `getTableRowsHandler` | agent `database.rows` | `id`, `table`, `page`, `limit` | `rows[]` | admin/user | Paginated rows |
| Databases | `/databases/:id/query` | POST | User | `queryDatabaseHandler` | agent `database.query` | `query` (SELECT only) | `rows[]` | admin/user | Run SELECT query |
| Email | `/email/mailboxes` | GET | User | `listMailboxesHandler` | `ListMailboxes` | `search`, `page`, `limit` | `data[]`, `meta` | admin/user | List mailboxes |
| Email | `/email/mailboxes` | POST | User | `createMailboxHandler` | agent `email.create` + `CreateMailbox` | `email`, `password`, `quota`, `display_name` | `Mailbox` | admin/user | Create mailbox |
| Email | `/email/mailboxes/:id` | GET | User | `getMailboxHandler` | `GetMailboxByID` | `id` | `Mailbox` | admin/user | Mailbox details |
| Email | `/email/mailboxes/:id` | PUT | User | `updateMailboxHandler` | `UpdateMailbox` | `display_name`, `forward_to`, `quota` | `message` | admin/user | Update mailbox |
| Email | `/email/mailboxes/:id` | DELETE | User | `deleteMailboxHandler` | agent `email.delete` + `DeleteMailbox` | `id` | `message` | admin/user | Delete mailbox |
| Email | `/email/aliases` | GET | User | `listAliasesHandler` | `ListAllAliases` | - | `[]EmailAlias` | admin/user | List aliases |
| Email | `/email/aliases` | POST | User | `createAliasHandler` | agent `email.alias.create` + `CreateAlias` | `domain`, `source`, `destination` | `EmailAlias` | admin/user | Create alias |
| Email | `/email/aliases/:id` | DELETE | User | `deleteAliasHandler` | agent `email.alias.delete` + `DeleteAlias` | `id`, `domain`, `source` | `message` | admin/user | Delete alias |
| Email | `/email/forwarders` | GET | User | `listForwardersHandler` | `ListAllForwarders` | - | `[]EmailForwarder` | admin/user | List forwarders |
| Email | `/email/forwarders` | POST | User | `createForwarderHandler` | agent `email.forwarder.create` + `CreateForwarder` | `domain`, `source`, `destination` | `EmailForwarder` | admin/user | Create forwarder |
| Email | `/email/forwarders/:id` | DELETE | User | `deleteForwarderHandler` | agent `email.forwarder.delete` + `DeleteForwarder` | `id`, `domain`, `source` | `message` | admin/user | Delete forwarder |
| Email | `/email/catch-all` | GET | User | `getCatchAllHandler` | `GetCatchAll` | `domain` | `EmailCatchAll` | admin/user | Get catch-all |
| Email | `/email/catch-all` | PUT | User | `setCatchAllHandler` | `SetCatchAll` | `domain`, `forward_to` | `message` | admin/user | Set catch-all |
| Email | `/email/deliverability` | GET | User | `deliverabilityHandler` | agent `email.deliverability` | `domain` | `spf`, `dkim`, `dmarc`, `ptr` | admin/user | DNS deliverability check |
| Firewall | `/firewall/rules` | GET | User | `listFirewallRulesHandler` | `ListFirewallRules` | `page`, `limit` | `data[]`, `meta` | admin/user | List rules |
| Firewall | `/firewall/rules` | POST | User | `createFirewallRuleHandler` | agent `firewall.rule.create` + `CreateFirewallRule` | `name`, `description`, `action`, `port`, `protocol`, `source` | `FirewallRule` | admin/user | Create UFW rule |
| Firewall | `/firewall/rules/:id` | GET | User | `getFirewallRuleHandler` | `GetFirewallRuleByID` | `id` | `FirewallRule` | admin/user | Rule details |
| Firewall | `/firewall/rules/:id` | PUT | User | `updateFirewallRuleHandler` | agent `firewall.rule.update` + `UpdateFirewallRule` | same as create | `message` | admin/user | Update rule |
| Firewall | `/firewall/rules/:id` | DELETE | User | `deleteFirewallRuleHandler` | agent `firewall.rule.delete` + `DeleteFirewallRule` | `id` | `message` | admin/user | Delete rule |
| Backups | `/backups` | GET | User | `listBackupsHandler` | `ListBackups` | `website_id`, `page`, `limit` | `data[]`, `meta` | admin/user | List backups |
| Backups | `/backups` | POST | User | `createBackupHandler` | `CreateBackup` + agent `backup.create` | `website_id`, `type`, `storage` | `Backup` | admin/user | Manual backup |
| Backups | `/backups/:id` | GET | User | `getBackupHandler` | `GetBackupByID` | `id` | `Backup` | admin/user | Backup details |
| Backups | `/backups/:id/restore` | POST | User | `restoreBackupHandler` | agent `backup.restore` | `website_id` (optional) | `message` | admin/user | Restore backup |
| Backups | `/backups/:id` | DELETE | User | `deleteBackupHandler` | agent `backup.delete` + `DeleteBackup` | `id` | `message` | admin/user | Delete backup |
| Backup Schedules | `/backup-schedules` | GET | User | `listBackupSchedulesHandler` | `ListBackupSchedules` | `website_id` | `[]BackupSchedule` | admin/user | List schedules |
| Backup Schedules | `/backup-schedules` | POST | User | `createBackupScheduleHandler` | `CreateBackupSchedule` | `website_id`, `schedule`, `retention_days`, `storage` | `BackupSchedule` | admin/user | Create schedule |
| Backup Schedules | `/backup-schedules/:id` | DELETE | User | `deleteBackupScheduleHandler` | `DeleteBackupSchedule` | `id` | `message` | admin/user | Delete schedule |
| Cron | `/cron` | GET | User | `listCronJobsHandler` | `ListCronJobs` | `page`, `limit` | `data[]`, `meta` | Own jobs / admin | List cron jobs |
| Cron | `/cron` | POST | User | `createCronJobHandler` | `CreateCronJob` + agent `cron.create` | `schedule`, `command`, `run_as`, `type` | `CronJob` | admin/user | Create cron job |
| Cron | `/cron/:id` | GET | User | `getCronJobHandler` | `GetCronJobByID`, `ListCronJobLogs` | `id` | `job`, `logs` | own/admin | Job details |
| Cron | `/cron/:id` | PUT | User | `updateCronJobHandler` | `UpdateCronJob` | `schedule`, `command`, `run_as` | `message` | own/admin | Update job |
| Cron | `/cron/:id` | DELETE | User | `deleteCronJobHandler` | agent `cron.delete` + `DeleteCronJob` | `id` | `message` | own/admin | Delete job |
| Cron | `/cron/:id/enable` | POST | User | `enableCronJobHandler` | `EnableCronJob` | `id` | `message` | own/admin | Enable job |
| Cron | `/cron/:id/disable` | POST | User | `disableCronJobHandler` | `DisableCronJob` | `id` | `message` | own/admin | Disable job |
| Cron | `/cron/:id/logs` | GET | User | `getCronJobLogsHandler` | `ListCronJobLogs` | `limit` | `[]CronJobLog` | own/admin | Execution logs |
| Alerts | `/alerts` | GET | User | `listAlertsHandler` | `ListAlerts` | `acknowledged`, `page`, `limit` | `data[]`, `meta` | admin/user | List alerts |
| Alerts | `/alerts/:id/acknowledge` | POST | User | `acknowledgeAlertHandler` | `AcknowledgeAlert` | `id` | `message` | admin/user | Acknowledge alert |
| Alerts | `/alerts/:id` | DELETE | User | `deleteAlertHandler` | `DeleteAlert` | `id` | `message` | admin/user | Delete alert |
| Alerts | `/alerts/settings` | GET | User | `getAlertSettingsHandler` | settings + defaults | - | `AlertSettings` | admin/user | Get alert config |
| Alerts | `/alerts/settings` | PUT | User | `updateAlertSettingsHandler` | `SetAlertSetting` | `AlertSettings` fields | `message` | admin/user | Update alert config |
| Metrics | `/metrics/current` | GET | User | `currentMetricsHandler` | metrics.Collector.Collect | - | `Metrics` | admin/user | Current server metrics |
| Metrics | `/metrics/history` | GET | User | `historyMetricsHandler` | stub | `range`, `interval` | `message` | admin/user | Historical metrics placeholder |
| Apps | `/websites/:id/apps` | GET | User | `listAppsHandler` | `ListAppsByWebsite` | `id` | `[]App` | admin/user/per-site | List installed apps |
| Apps | `/websites/:id/apps/install` | POST | User | `installAppHandler` | agent `apps.install` + `CreateApp` | `app_type`, admin/db fields | `App` | admin/user/per-site | Install app |
| Apps | `/websites/:id/apps/:app_id/update` | POST | User | `updateAppHandler` | agent `apps.check-update`, `apps.update` | `id`, `app_id` | `message`, `version` | admin/user/per-site | Update app |
| Git | `/websites/:id/git` | GET | User | `getGitConfigHandler` | `GetWebsiteByID` | `id` | `gitConfig` | admin/user/per-site | Get git deploy config |
| Git | `/websites/:id/git` | POST | User | `setupGitHandler` | agent `git.setup` + `UpdateWebsiteGitConfig` | `repo_url`, `branch`, `deploy_key`, `auto_deploy` | `message`, `webhook_token` | admin/user/per-site | Setup git deploy |
| Git | `/websites/:id/git/pull` | POST | User | `gitPullHandler` | agent `git.pull` | `id` | `message` | admin/user/per-site | Trigger git pull |
| Git | `/websites/:id/git/webhook` | GET | User | `getWebhookURLHandler` | `GetWebsiteByID` | `id` | `webhook_url` | admin/user/per-site | Get webhook URL |
| Webhooks | `/webhooks/git/:token` | POST | None | `gitWebhookHandler` | find website by token, HMAC verify, agent `git.pull` | payload body | `message` | None (token auth) | GitHub webhook auto-deploy |
| Webmail | `/webmail/status` | GET | User | `webmailStatusHandler` | agent `webmail.status` | - | `installed` | admin/user | Roundcube status |
| Webmail | `/webmail/install` | POST | User | `installWebmailHandler` | agent `webmail.install` | - | `message` | admin/user | Install Roundcube |
| Webmail | `/webmail/uninstall` | POST | User | `uninstallWebmailHandler` | agent `webmail.uninstall` | - | `message` | admin/user | Uninstall Roundcube |
| Webmail | `/webmail/url` | GET | User | `webmailURLHandler` | agent `webmail.status` | - | `url` | admin/user | Get webmail URL |
| Terminal | `/terminal/session` | POST | User | `createTerminalSessionHandler` | agent `terminal.start` | - | `session_id`, `recording` | admin/user | Create PTY session |
| Terminal | `/terminal/sessions` | GET | User | `listTerminalSessionsHandler` | in-memory stub | - | `[]TerminalSession` | admin/user | List sessions |
| Terminal | `/terminal/sessions/:id` | DELETE | User | `closeTerminalSessionHandler` | agent `terminal.stop` | `id` | `message` | admin/user | Close session |
| Terminal | `/terminal/recordings` | GET | User | `listTerminalRecordingsHandler` | read `/var/log/juvia/terminal` | - | `[]TerminalRecording` | admin/user | List recordings |
| Terminal | `/terminal/recordings/:id` | GET | User | `getTerminalRecordingHandler` | read `.cast` file | `id` | `filename`, `content`, `size` | admin/user | Playback recording |
| Updates | `/updates/check` | GET | User | `checkUpdateHandler` | `updater.CheckForUpdates` | - | `UpdateInfo` | admin/user | Check for update |
| Updates | `/updates/download` | POST | User | `downloadUpdateHandler` | `updater.DownloadUpdate` | `url` | `path` | admin/user | Download binary |
| Updates | `/updates/apply` | POST | User | `applyUpdateHandler` | `updater.VerifyGPGSignature`, `ApplyUpdate` | `path`, `sig_url` | `message` | admin/user | Apply update |
| Updates | `/updates/rollback` | POST | User | `rollbackHandler` | `updater.Rollback` | - | `message` | admin/user | Rollback update |
| Settings | `/settings` | GET | User | `getSettingsHandler` | `GetSetting` for known keys | - | `map[string]string` | admin/user | Get panel settings |
| Settings | `/settings` | PUT | User | `updateSettingsHandler` | `SetSetting` | key/value map | `message`, `warning` | admin/user | Update settings |
| Settings | `/settings/export` | GET | User | `exportConfigHandler` | DB list queries | - | JSON export | admin/user | Export config |
| Settings | `/settings/import` | POST | User | `importConfigHandler` | parsing stub | `version`, config arrays | `imported_count`, `skipped_count` | admin/user | Import config |
| Audit Log | `/audit-log` | GET | User | `getAuditLogHandler` | `ListAuditLog` | `page`, `limit` | `[]AuditLogEntry`, `meta` | admin/user | View audit log |
| Services | `/services` | GET | User | `listServicesHandler` | agent `services.status` | - | `[]Service` | admin/user | List service statuses |
| Services | `/services/:name` | GET | User | `getServiceHandler` | agent `services.status` | `name` | `Service` | admin/user | Service details |
| Services | `/services/:name/restart` | POST | User | `restartServiceHandler` | agent `services.restart` | `name` | `message` | admin/user | Restart service |
| Version | `/version` | GET | None | `getVersionHandler` | `updater.GetCurrentVersion` | - | `version` | None | Panel version |

---

## Data Model Matrix

| Model | Fields | Relations | Used By |
|---|---|---|---|
| **User** | id, username, password_hash, role, email, active, totp_enabled, totp_secret, totp_recovery_codes, created_at, last_login, updated_at | - | Auth, Users API, sessions, cron jobs, audit log |
| **UserSite** | id, user_id, website_id, created_at | users.id, websites.id | Per-site authorization (schema exists; enforcement not implemented) |
| **Session** | id, user_id, token_hash, ip_address, user_agent, revoked, expires_at, created_at | users.id | Auth refresh/logout/session list |
| **Website** | id, domain, document_root, php_version, web_server, ssl_enabled, ssl_expiry, status, deleted_at, user_id, created_at, updated_at, git_config | users.id (owner) | Websites API, DNS, backups, apps, git, alerts |
| **Domain** | id, website_id, domain, type, created_at | websites.id | Website aliases/addons (currently only primary created) |
| **Zone** | id, website_id, domain, serial, created_at, updated_at | websites.id | DNS API, website creation |
| **DNSRecord** | id, zone_id, type, name, value, priority, created_at, updated_at | zones.id | DNS API, suggestions |
| **Database** | id, name, engine, user_id, created_at | users.id | Databases API, backups |
| **DBUser** | id, database_id, username, password_hash, host, created_at | databases.id | Databases API |
| **Mailbox** | id, email, password_hash, quota, display_name, forward_to, status, autoresponder_* fields, created_at | - | Email API |
| **EmailAlias** | id, domain, source, destination, created_at | - | Email API |
| **EmailForwarder** | id, domain, source, destination, created_at | - | Email API |
| **EmailCatchAll** | id, domain, forward_to, created_at | - | Email API |
| **Task** | id, task_id, type, status, progress, steps, result, created_by, created_at, updated_at, completed_at | users.id | Task runner, website/SSL workflows |
| **AuditLog** | id, user_id, action, ip_address, user_agent, details, created_at | users.id | Audit log API, most mutating handlers |
| **FirewallRule** | id, name, description, action, port, protocol, source, created_at | - | Firewall API, agent UFW |
| **CronJob** | id, user_id, schedule, command, run_as, type, enabled, last_run, last_status, last_output, created_at | users.id | Cron API, agent cron |
| **CronJobLog** | id, cron_job_id, run_at, duration_ms, status, output, created_at | cron_jobs.id | Cron job details |
| **Backup** | id, website_id, type, status, storage, path, size_bytes, checksum, verified, verified_at, created_at | websites.id | Backups API, alert engine |
| **BackupSchedule** | id, website_id, schedule, retention_days, storage, enabled, created_at | websites.id | Backup schedules API |
| **Alert** | id, type, severity, message, resource_id, resource_type, acknowledged, created_at | websites.id (resource) | Alerts API, alert engine |
| **Service** | id, name, version, installed, running, health, last_error, checked_at | - | Service registry table (agent-driven status used instead) |
| **Setting** | id, key, value, updated_at | - | Settings API, alert settings |
| **App** | id, app_type, name, website_id, version, installed_at, update_available | websites.id | Apps API |

---

## Event Matrix

| Event | Trigger | Consumer | Notes |
|---|---|---|---|
| `website.created` | Code constant defined | Event bus only; no producer currently | Not emitted in current code |
| `website.deleted` | Code constant defined | Event bus only; no producer currently | Not emitted in current code |
| `alert.fired` | `alerts.Engine.emitAlert` | WebSocket subscribers (potential) | Emitted on alert creation |
| `metrics.update` | Code constant defined | Event bus only; `metricsPump` reads collector directly | Not used via bus |
| `task.progress` | Code constant defined | `/ws/v1/tasks/:taskId` subscribers | Subscribed but never published |
| `task.complete` | Code constant defined | `/ws/v1/tasks/:taskId` subscribers | Subscribed but never published |

---

## Missing/Incomplete Backend Features

| # | Feature | Status | Evidence / Location |
|---|---|---|---|
| 1 | **Per-site RBAC enforcement** | Missing | `users.role = 'per_site'` exists and `user_sites` table exists, but middleware only checks `admin`; no site ownership checks in handlers. |
| 2 | **CSRF protection** | Missing | `generateCSRFToken` exists, `X-CSRF-Token` returned by `/auth/me`, but no middleware validates it on mutating endpoints. |
| 3 | **Rate limiting** | Missing | Docs specify login/IP rate limits; no implementation in `middleware.go` or router. |
| 4 | **Auto-block failed logins via UFW** | Missing | SECURITY.md mentions blocking IP after 20 failures; not implemented. |
| 5 | **Login alerts / new-IP detection** | Missing | SECURITY.md mentions login alerts; not implemented. |
| 6 | **Session cleanup / last active** | Partial | `CleanupExpiredSessions` exists but is never called. `last_login` not updated on login. |
| 7 | **Password complexity validation** | Missing | SECURITY.md requires 12 chars + complexity; no validation in `setup.go`, `users.go`, or login flow. |
| 8 | **2FA recovery codes persistence** | Missing | `verify2FAHandler` generates codes but JSON-marshals and discards them; DB column remains unused. |
| 9 | **TOTP secret encryption** | Missing | `totp_secret` stored plaintext. Docs say encrypted. |
| 10 | **VictoriaMetrics integration** | Missing | Health returns `victoria_metrics: "ok"` hardcoded; `metrics/history` is a stub. |
| 11 | **Website update propagates to agent** | Missing | `updateWebsiteHandler` only updates DB; does not regenerate nginx/php-fpm config. |
| 12 | **Apache support** | Partial | `web_server` accepts `apache` but agent only generates nginx configs. |
| 13 | **Domain aliases/addons** | Missing | `domains` table exists; only primary domain created. No API to manage addon/subdomains. |
| 14 | **SSL auto-redirect nginx config** | Missing | `sslSiteTemplate` exists but `SSLEnabled` never set true in agent `website.create`; issue endpoint does not rewrite vhost to HTTPS. |
| 15 | **File archive extraction** | Stub | `HandleFilesExtract` only validates extension and returns `extracted: true`; no actual extraction. |
| 16 | **File/folder upload binary/multipart** | Missing | Upload endpoint expects base64/text `content`; no multipart file upload. |
| 17 | **Database import endpoint** | Missing | API.md lists `POST /databases/:id/import`; not implemented. |
| 18 | **Database remote access toggle / UFW** | Missing | MODULES.md mentions remote access adds UFW rule; no endpoint exists. |
| 19 | **Email autoresponders** | Missing | Schema fields exist; no API endpoints. |
| 20 | **Email mailbox password change** | Missing | `updateMailboxHandler` does not accept password. |
| 21 | **Catch-all agent synchronization** | Missing | DB catch-all set, but no agent method updates Postfix config. |
| 22 | **DKIM signing generation** | Missing | Deliverability checks DKIM DNS but panel does not generate/sign keys. |
| 23 | **Firewall rule update / delete UFW sync** | Buggy | `ufw.go` `deleteOldRule` parses `ufw status numbered` but never executes deletion by number. |
| 24 | **Firewall country/geo blocking** | Missing | MODULES.md mentions geo-blocking; not implemented. |
| 25 | **Backup schedule execution** | Missing | Schedules stored in DB but no scheduler/agent cron links them. |
| 26 | **Backup remote storage (S3/R2/B2/SFTP)** | Missing | `storage` field accepts values but agent only writes local. |
| 27 | **Backup restore database path** | Buggy | `backup.go` `getWebsiteByID` is a hardcoded stub returning dummy data. |
| 28 | **Cron job execution scheduling** | Partial | Agent writes `/etc/cron.d/juvia` but `reloadCron()` never called; job runs via `HandleCronRun` only. |
| 29 | **Cron job types URL/PHP** | Missing | `type` field stored but all executed as shell. |
| 30 | **Cron job failure email alerts** | Missing | MODULES.md mentions email alert on failure. |
| 31 | **Alert engine start in panel** | Missing | `alerts.NewEngine` defined but not instantiated/started in `cmd/panel/main.go`. |
| 32 | **Task progress WebSocket events** | Missing | Task runner never publishes `EventTaskProgress` / `EventTaskComplete`; clients subscribing get nothing. |
| 33 | **Terminal PTY implementation** | Missing | Agent `terminal.start/stop` create empty `.cast` files; no actual PTY or I/O bridge. WebSocket `/ws/v1/terminal/:sessionId` route missing. |
| 34 | **Terminal idle timeout / root re-auth / IP restriction** | Missing | SECURITY.md mentions these; not implemented. |
| 35 | **Self-update HTTP/GPG real flow** | Partial | `CheckForUpdates` fetches but never parses JSON; `VerifyGPGSignature` implemented but `ApplyUpdate` restarts service blindly. |
| 36 | **Settings import real logic** | Stub | `importConfigHandler` only counts objects; does not insert anything. |
| 37 | **Audit log details JSON** | Missing | All handlers pass empty `details` string; structured details unused. |
| 38 | **Log sanitization (passwords/tokens)** | Missing | SECURITY.md requires redaction; logs use raw user agent/IPs. |
| 39 | **SQLite encryption at rest** | Missing | SECURITY.md mentions SQLCipher; no implementation. |
| 40 | **Agent privilege drop / capabilities** | Missing | Agent runs as root; no Linux capability isolation implemented. |
| 41 | **Path safety on terminal recordings / backup paths** | Missing | `SafePath` used for website files only; recording/backup paths constructed unchecked. |
| 42 | **UFW validation of dangerous defaults** | Missing | User can create deny rule on port 22/8080 and lock themselves out. |
| 43 | **App installs beyond WordPress/Laravel** | Partial | Regex allows ghost/drupal/joomla/nextjs but agent only supports wordpress/laravel. |
| 44 | **Next.js / Node app install** | Missing | Not implemented. |
| 45 | **One-click app database creation** | Missing | App installer does not create DB or DB user; uses placeholder credentials. |
| 46 | **Git webhook HMAC verification bug** | Buggy | `git_webhook.go` logic inverts HMAC equality (`!hmac.Equal` returns success when mismatching). |
| 47 | **Git deploy key actual usage** | Missing | Deploy key stored but clone command ignores it (same branch regardless). |
| 48 | **Roundcube nginx config / vhost** | Missing | Webmail installer extracts files and creates DB but no nginx vhost generated. |
| 49 | **Service registry background health checks** | Missing | `services` table exists but health engine not started. |
| 50 | **CORS production policy / CSP headers** | Missing | Only dev CORS middleware exists; no CSP headers. |

---

## Backend Architecture Summary

Juvia is a two-process Linux server control panel:

- **Panel** (`cmd/panel`) runs as unprivileged user `juvia`. It serves an embedded React frontend, exposes a Gin REST API on port `8080`/`18473`, authenticates users with JWT + SQLite-backed refresh sessions, and forwards privileged operations to the Agent over a Unix domain socket.
- **Agent** (`cmd/agent`) runs as root, exposes a JSON-RPC Unix socket (`/var/run/juvia/agent.sock`, mode `660`, group `juvia`), and executes system operations: Linux user creation, nginx/PHP-FPM config generation, certbot SSL, BIND9 DNS zone files, MySQL/PostgreSQL, Postfix/Dovecot email, UFW firewall, backups, cron, git deploy, terminal stubs, and service control.
- **Database** is SQLite with WAL mode (`/var/lib/juvia/juvia.db`). Migrations are applied at startup from `/usr/share/juvia/migrations`.
- **IPC** uses a custom line-delimited JSON-RPC over Unix socket with request/response envelopes.
- **Security model** relies on privilege separation, path traversal checks (`utils.SafePath`), and atomic config writes. Many documented controls (CSRF, rate limiting, RBAC, encryption, CSP, audit details) are not yet implemented.
- **Async tasks** are tracked in the `tasks` table but task progress is not published to the event bus; WebSocket task subscriptions receive no updates.
- **WebSocket** streams raw metrics every 2 seconds and can subscribe to task events, but task events are never emitted.
- **Alert engine** has background check logic but is not wired into the panel startup.

---

## Machine-Readable Inventory

```json
{
  "modules": [
    {"name": "panel", "type": "entrypoint", "path": "cmd/panel/main.go"},
    {"name": "agent", "type": "entrypoint", "path": "cmd/agent/main.go"},
    {"name": "api", "type": "http_handlers", "path": "internal/api"},
    {"name": "auth", "type": "authentication", "path": "internal/auth"},
    {"name": "db", "type": "database", "path": "internal/db"},
    {"name": "agent_website", "type": "privileged_service", "path": "internal/agent/website.go"},
    {"name": "agent_nginx", "type": "privileged_service", "path": "internal/agent/nginx/generator.go"},
    {"name": "agent_ssl", "type": "privileged_service", "path": "internal/agent/ssl.go"},
    {"name": "agent_dns", "type": "privileged_service", "path": "internal/agent/dns.go"},
    {"name": "agent_files", "type": "privileged_service", "path": "internal/agent/files.go"},
    {"name": "agent_email", "type": "privileged_service", "path": "internal/agent/email/postfix.go"},
    {"name": "agent_webmail", "type": "privileged_service", "path": "internal/agent/email/webmail.go"},
    {"name": "agent_database", "type": "privileged_service", "path": "internal/agent/database/mysql.go"},
    {"name": "agent_firewall", "type": "privileged_service", "path": "internal/agent/firewall/ufw.go"},
    {"name": "agent_backup", "type": "privileged_service", "path": "internal/agent/backup/backup.go"},
    {"name": "agent_cron", "type": "privileged_service", "path": "internal/agent/cron/manager.go"},
    {"name": "agent_apps", "type": "privileged_service", "path": "internal/agent/apps/installer.go"},
    {"name": "agent_git", "type": "privileged_service", "path": "internal/agent/git/git.go"},
    {"name": "agent_terminal", "type": "privileged_service", "path": "internal/agent/terminal/record.go"},
    {"name": "agent_services", "type": "privileged_service", "path": "internal/agent/services/services.go"},
    {"name": "tasks", "type": "async_jobs", "path": "internal/tasks/runner.go"},
    {"name": "events", "type": "pub_sub", "path": "internal/events/bus.go"},
    {"name": "ws", "type": "websocket", "path": "internal/ws/server.go"},
    {"name": "metrics", "type": "monitoring", "path": "internal/metrics/collector_linux.go"},
    {"name": "alerts", "type": "monitoring", "path": "internal/alerts/engine.go"},
    {"name": "updater", "type": "self_update", "path": "internal/updater/updater.go"},
    {"name": "socket", "type": "ipc", "path": "internal/socket"},
    {"name": "utils", "type": "security", "path": "internal/utils/safe.go"},
    {"name": "apitest", "type": "integration_tests", "path": "cmd/api-test-runner"}
  ],
  "routes": [
    {"method": "GET", "path": "/api/v1/health", "auth": "none", "handler": "healthHandler"},
    {"method": "GET", "path": "/api/v1/setup/status", "auth": "none", "handler": "setupStatusHandler"},
    {"method": "POST", "path": "/api/v1/setup/first-run", "auth": "none", "handler": "firstRunHandler"},
    {"method": "POST", "path": "/api/v1/auth/login", "auth": "none", "handler": "loginHandler"},
    {"method": "POST", "path": "/api/v1/auth/refresh", "auth": "cookie", "handler": "refreshHandler"},
    {"method": "POST", "path": "/api/v1/auth/logout", "auth": "user", "handler": "logoutHandler"},
    {"method": "GET", "path": "/api/v1/auth/me", "auth": "user", "handler": "meHandler"},
    {"method": "POST", "path": "/api/v1/auth/2fa/enable", "auth": "user", "handler": "enable2FAHandler"},
    {"method": "POST", "path": "/api/v1/auth/2fa/verify", "auth": "user", "handler": "verify2FAHandler"},
    {"method": "DELETE", "path": "/api/v1/auth/sessions/:id", "auth": "user", "handler": "revokeSessionHandler"},
    {"method": "GET", "path": "/api/v1/users", "auth": "admin", "handler": "listUsersHandler"},
    {"method": "POST", "path": "/api/v1/users", "auth": "admin", "handler": "createUserHandler"},
    {"method": "GET", "path": "/api/v1/users/:id", "auth": "admin", "handler": "getUserHandler"},
    {"method": "PUT", "path": "/api/v1/users/:id", "auth": "admin", "handler": "updateUserHandler"},
    {"method": "DELETE", "path": "/api/v1/users/:id", "auth": "admin", "handler": "deleteUserHandler"},
    {"method": "GET", "path": "/api/v1/websites", "auth": "user", "handler": "listWebsitesHandler"},
    {"method": "POST", "path": "/api/v1/websites", "auth": "user", "handler": "createWebsiteHandler"},
    {"method": "GET", "path": "/api/v1/websites/trash", "auth": "user", "handler": "listTrashedWebsitesHandler"},
    {"method": "DELETE", "path": "/api/v1/websites/trash/purge", "auth": "user", "handler": "purgeTrashHandler"},
    {"method": "GET", "path": "/api/v1/websites/:id", "auth": "user", "handler": "getWebsiteHandler"},
    {"method": "PUT", "path": "/api/v1/websites/:id", "auth": "user", "handler": "updateWebsiteHandler"},
    {"method": "DELETE", "path": "/api/v1/websites/:id", "auth": "user", "handler": "deleteWebsiteHandler"},
    {"method": "POST", "path": "/api/v1/websites/:id/suspend", "auth": "user", "handler": "suspendWebsiteHandler"},
    {"method": "POST", "path": "/api/v1/websites/:id/restore", "auth": "user", "handler": "restoreWebsiteHandler"},
    {"method": "DELETE", "path": "/api/v1/websites/:id/permanent", "auth": "user", "handler": "permanentDeleteWebsiteHandler"},
    {"method": "PUT", "path": "/api/v1/websites/:id/php", "auth": "user", "handler": "changePHPHandler"},
    {"method": "GET", "path": "/api/v1/websites/:id/ssl/check", "auth": "user", "handler": "sslCheckHandler"},
    {"method": "POST", "path": "/api/v1/websites/:id/ssl", "auth": "user", "handler": "issueSSLHandler"},
    {"method": "POST", "path": "/api/v1/websites/:id/ssl-renew", "auth": "user", "handler": "renewSSLHandler"},
    {"method": "DELETE", "path": "/api/v1/websites/:id/ssl", "auth": "user", "handler": "removeSSLHandler"},
    {"method": "GET", "path": "/api/v1/websites/:id/dns/records", "auth": "user", "handler": "listDNSRecordsHandler"},
    {"method": "POST", "path": "/api/v1/websites/:id/dns/records", "auth": "user", "handler": "createDNSRecordHandler"},
    {"method": "POST", "path": "/api/v1/websites/:id/dns/suggest", "auth": "user", "handler": "suggestDNSHandler"},
    {"method": "PUT", "path": "/api/v1/dns/records/:id", "auth": "user", "handler": "updateDNSRecordHandler"},
    {"method": "DELETE", "path": "/api/v1/dns/records/:id", "auth": "user", "handler": "deleteDNSRecordHandler"},
    {"method": "GET", "path": "/api/v1/websites/:id/files", "auth": "user", "handler": "listFilesHandler"},
    {"method": "POST", "path": "/api/v1/websites/:id/files/upload", "auth": "user", "handler": "uploadFilesHandler"},
    {"method": "GET", "path": "/api/v1/websites/:id/files/download", "auth": "user", "handler": "downloadFilesHandler"},
    {"method": "PUT", "path": "/api/v1/websites/:id/files/rename", "auth": "user", "handler": "renameFilesHandler"},
    {"method": "DELETE", "path": "/api/v1/websites/:id/files/delete", "auth": "user", "handler": "deleteFilesHandler"},
    {"method": "POST", "path": "/api/v1/websites/:id/files/extract", "auth": "user", "handler": "extractFilesHandler"},
    {"method": "GET", "path": "/api/v1/websites/:id/files/edit", "auth": "user", "handler": "getFileContentHandler"},
    {"method": "PUT", "path": "/api/v1/websites/:id/files/edit", "auth": "user", "handler": "saveFileContentHandler"},
    {"method": "GET", "path": "/api/v1/websites/:id/logs/access", "auth": "user", "handler": "accessLogHandler"},
    {"method": "GET", "path": "/api/v1/websites/:id/logs/error", "auth": "user", "handler": "errorLogHandler"},
    {"method": "GET", "path": "/api/v1/databases", "auth": "user", "handler": "listDatabasesHandler"},
    {"method": "POST", "path": "/api/v1/databases", "auth": "user", "handler": "createDatabaseHandler"},
    {"method": "GET", "path": "/api/v1/databases/:id", "auth": "user", "handler": "getDatabaseHandler"},
    {"method": "DELETE", "path": "/api/v1/databases/:id", "auth": "user", "handler": "deleteDatabaseHandler"},
    {"method": "POST", "path": "/api/v1/databases/:id/users", "auth": "user", "handler": "createDBUserHandler"},
    {"method": "DELETE", "path": "/api/v1/databases/:id/users/:user_id", "auth": "user", "handler": "deleteDBUserHandler"},
    {"method": "POST", "path": "/api/v1/databases/:id/export", "auth": "user", "handler": "exportDatabaseHandler"},
    {"method": "GET", "path": "/api/v1/databases/:id/tables", "auth": "user", "handler": "listTablesHandler"},
    {"method": "GET", "path": "/api/v1/databases/:id/tables/:table/rows", "auth": "user", "handler": "getTableRowsHandler"},
    {"method": "POST", "path": "/api/v1/databases/:id/query", "auth": "user", "handler": "queryDatabaseHandler"},
    {"method": "GET", "path": "/api/v1/email/mailboxes", "auth": "user", "handler": "listMailboxesHandler"},
    {"method": "POST", "path": "/api/v1/email/mailboxes", "auth": "user", "handler": "createMailboxHandler"},
    {"method": "GET", "path": "/api/v1/email/mailboxes/:id", "auth": "user", "handler": "getMailboxHandler"},
    {"method": "PUT", "path": "/api/v1/email/mailboxes/:id", "auth": "user", "handler": "updateMailboxHandler"},
    {"method": "DELETE", "path": "/api/v1/email/mailboxes/:id", "auth": "user", "handler": "deleteMailboxHandler"},
    {"method": "GET", "path": "/api/v1/email/aliases", "auth": "user", "handler": "listAliasesHandler"},
    {"method": "POST", "path": "/api/v1/email/aliases", "auth": "user", "handler": "createAliasHandler"},
    {"method": "DELETE", "path": "/api/v1/email/aliases/:id", "auth": "user", "handler": "deleteAliasHandler"},
    {"method": "GET", "path": "/api/v1/email/forwarders", "auth": "user", "handler": "listForwardersHandler"},
    {"method": "POST", "path": "/api/v1/email/forwarders", "auth": "user", "handler": "createForwarderHandler"},
    {"method": "DELETE", "path": "/api/v1/email/forwarders/:id", "auth": "user", "handler": "deleteForwarderHandler"},
    {"method": "GET", "path": "/api/v1/email/catch-all", "auth": "user", "handler": "getCatchAllHandler"},
    {"method": "PUT", "path": "/api/v1/email/catch-all", "auth": "user", "handler": "setCatchAllHandler"},
    {"method": "GET", "path": "/api/v1/email/deliverability", "auth": "user", "handler": "deliverabilityHandler"},
    {"method": "GET", "path": "/api/v1/firewall/rules", "auth": "user", "handler": "listFirewallRulesHandler"},
    {"method": "POST", "path": "/api/v1/firewall/rules", "auth": "user", "handler": "createFirewallRuleHandler"},
    {"method": "GET", "path": "/api/v1/firewall/rules/:id", "auth": "user", "handler": "getFirewallRuleHandler"},
    {"method": "PUT", "path": "/api/v1/firewall/rules/:id", "auth": "user", "handler": "updateFirewallRuleHandler"},
    {"method": "DELETE", "path": "/api/v1/firewall/rules/:id", "auth": "user", "handler": "deleteFirewallRuleHandler"},
    {"method": "GET", "path": "/api/v1/backups", "auth": "user", "handler": "listBackupsHandler"},
    {"method": "POST", "path": "/api/v1/backups", "auth": "user", "handler": "createBackupHandler"},
    {"method": "GET", "path": "/api/v1/backups/:id", "auth": "user", "handler": "getBackupHandler"},
    {"method": "POST", "path": "/api/v1/backups/:id/restore", "auth": "user", "handler": "restoreBackupHandler"},
    {"method": "DELETE", "path": "/api/v1/backups/:id", "auth": "user", "handler": "deleteBackupHandler"},
    {"method": "GET", "path": "/api/v1/backup-schedules", "auth": "user", "handler": "listBackupSchedulesHandler"},
    {"method": "POST", "path": "/api/v1/backup-schedules", "auth": "user", "handler": "createBackupScheduleHandler"},
    {"method": "DELETE", "path": "/api/v1/backup-schedules/:id", "auth": "user", "handler": "deleteBackupScheduleHandler"},
    {"method": "GET", "path": "/api/v1/cron", "auth": "user", "handler": "listCronJobsHandler"},
    {"method": "POST", "path": "/api/v1/cron", "auth": "user", "handler": "createCronJobHandler"},
    {"method": "GET", "path": "/api/v1/cron/:id", "auth": "user", "handler": "getCronJobHandler"},
    {"method": "PUT", "path": "/api/v1/cron/:id", "auth": "user", "handler": "updateCronJobHandler"},
    {"method": "DELETE", "path": "/api/v1/cron/:id", "auth": "user", "handler": "deleteCronJobHandler"},
    {"method": "POST", "path": "/api/v1/cron/:id/enable", "auth": "user", "handler": "enableCronJobHandler"},
    {"method": "POST", "path": "/api/v1/cron/:id/disable", "auth": "user", "handler": "disableCronJobHandler"},
    {"method": "GET", "path": "/api/v1/cron/:id/logs", "auth": "user", "handler": "getCronJobLogsHandler"},
    {"method": "GET", "path": "/api/v1/logs/system", "auth": "user", "handler": "systemLogsHandler"},
    {"method": "GET", "path": "/api/v1/alerts", "auth": "user", "handler": "listAlertsHandler"},
    {"method": "POST", "path": "/api/v1/alerts/:id/acknowledge", "auth": "user", "handler": "acknowledgeAlertHandler"},
    {"method": "DELETE", "path": "/api/v1/alerts/:id", "auth": "user", "handler": "deleteAlertHandler"},
    {"method": "GET", "path": "/api/v1/alerts/settings", "auth": "user", "handler": "getAlertSettingsHandler"},
    {"method": "PUT", "path": "/api/v1/alerts/settings", "auth": "user", "handler": "updateAlertSettingsHandler"},
    {"method": "GET", "path": "/api/v1/metrics/current", "auth": "user", "handler": "currentMetricsHandler"},
    {"method": "GET", "path": "/api/v1/metrics/history", "auth": "user", "handler": "historyMetricsHandler"},
    {"method": "GET", "path": "/api/v1/websites/:id/apps", "auth": "user", "handler": "listAppsHandler"},
    {"method": "POST", "path": "/api/v1/websites/:id/apps/install", "auth": "user", "handler": "installAppHandler"},
    {"method": "POST", "path": "/api/v1/websites/:id/apps/:app_id/update", "auth": "user", "handler": "updateAppHandler"},
    {"method": "GET", "path": "/api/v1/websites/:id/git", "auth": "user", "handler": "getGitConfigHandler"},
    {"method": "POST", "path": "/api/v1/websites/:id/git", "auth": "user", "handler": "setupGitHandler"},
    {"method": "POST", "path": "/api/v1/websites/:id/git/pull", "auth": "user", "handler": "gitPullHandler"},
    {"method": "GET", "path": "/api/v1/websites/:id/git/webhook", "auth": "user", "handler": "getWebhookURLHandler"},
    {"method": "POST", "path": "/api/v1/webhooks/git/:token", "auth": "none", "handler": "gitWebhookHandler"},
    {"method": "GET", "path": "/api/v1/webmail/status", "auth": "user", "handler": "webmailStatusHandler"},
    {"method": "POST", "path": "/api/v1/webmail/install", "auth": "user", "handler": "installWebmailHandler"},
    {"method": "POST", "path": "/api/v1/webmail/uninstall", "auth": "user", "handler": "uninstallWebmailHandler"},
    {"method": "GET", "path": "/api/v1/webmail/url", "auth": "user", "handler": "webmailURLHandler"},
    {"method": "POST", "path": "/api/v1/terminal/session", "auth": "user", "handler": "createTerminalSessionHandler"},
    {"method": "GET", "path": "/api/v1/terminal/sessions", "auth": "user", "handler": "listTerminalSessionsHandler"},
    {"method": "DELETE", "path": "/api/v1/terminal/sessions/:id", "auth": "user", "handler": "closeTerminalSessionHandler"},
    {"method": "GET", "path": "/api/v1/terminal/recordings", "auth": "user", "handler": "listTerminalRecordingsHandler"},
    {"method": "GET", "path": "/api/v1/terminal/recordings/:id", "auth": "user", "handler": "getTerminalRecordingHandler"},
    {"method": "GET", "path": "/api/v1/updates/check", "auth": "user", "handler": "checkUpdateHandler"},
    {"method": "POST", "path": "/api/v1/updates/download", "auth": "user", "handler": "downloadUpdateHandler"},
    {"method": "POST", "path": "/api/v1/updates/apply", "auth": "user", "handler": "applyUpdateHandler"},
    {"method": "POST", "path": "/api/v1/updates/rollback", "auth": "user", "handler": "rollbackHandler"},
    {"method": "GET", "path": "/api/v1/settings", "auth": "user", "handler": "getSettingsHandler"},
    {"method": "PUT", "path": "/api/v1/settings", "auth": "user", "handler": "updateSettingsHandler"},
    {"method": "GET", "path": "/api/v1/settings/export", "auth": "user", "handler": "exportConfigHandler"},
    {"method": "POST", "path": "/api/v1/settings/import", "auth": "user", "handler": "importConfigHandler"},
    {"method": "GET", "path": "/api/v1/audit-log", "auth": "user", "handler": "getAuditLogHandler"},
    {"method": "GET", "path": "/api/v1/services", "auth": "user", "handler": "listServicesHandler"},
    {"method": "GET", "path": "/api/v1/services/:name", "auth": "user", "handler": "getServiceHandler"},
    {"method": "POST", "path": "/api/v1/services/:name/restart", "auth": "user", "handler": "restartServiceHandler"},
    {"method": "GET", "path": "/api/v1/version", "auth": "none", "handler": "getVersionHandler"},
    {"method": "GET", "path": "/ws/v1/metrics", "auth": "none", "handler": "wsServer.HandleMetrics"},
    {"method": "GET", "path": "/ws/v1/tasks/:taskId", "auth": "none", "handler": "wsServer.HandleTasks"}
  ],
  "models": [
    "users", "user_sites", "sessions", "websites", "domains", "zones", "dns_records",
    "databases", "db_users", "mailboxes", "email_aliases", "email_forwarders", "email_catchall",
    "tasks", "audit_log", "firewall_rules", "cron_jobs", "cron_job_logs", "backups",
    "backup_schedules", "alerts", "services", "settings", "apps"
  ],
  "events": [
    {"type": "website.created", "trigger": "none (constant only)", "consumer": "event bus"},
    {"type": "website.deleted", "trigger": "none (constant only)", "consumer": "event bus"},
    {"type": "alert.fired", "trigger": "alerts.Engine.emitAlert", "consumer": "event bus / potential WebSocket"},
    {"type": "metrics.update", "trigger": "none (constant only)", "consumer": "event bus"},
    {"type": "task.progress", "trigger": "none (constant only)", "consumer": "WebSocket task subscribers"},
    {"type": "task.complete", "trigger": "none (constant only)", "consumer": "WebSocket task subscribers"}
  ],
  "features": {
    "implemented": [
      "JWT access + refresh cookie sessions", "bcrypt passwords", "TOTP 2FA enable/verify",
      "first-run wizard", "admin user CRUD", "website CRUD with Linux user + nginx + PHP-FPM",
      "website trash/purge/restore/suspend", "DNS zone + record CRUD + suggestions",
      "file manager list/upload/download/rename/delete/edit", "SSL issue/renew/remove (certbot fallback self-signed)",
      "MySQL/PostgreSQL create/delete/user/query/export", "email mailboxes/aliases/forwarders/catch-all",
      "deliverability DNS checks", "UFW firewall rule CRUD", "manual backup + schedule CRUD",
      "cron job CRUD/enable/disable/logs", "alert list/ack/delete/settings", "current metrics collection",
      "one-click WordPress/Laravel install/update", "git deploy setup/pull/webhook", "Roundcube webmail install/uninstall",
      "terminal session/recording API (no PTY)", "self-update check/download/apply/rollback API",
      "settings export/import API", "audit log API", "service status/restart API"
    ],
    "missing_or_incomplete": [
      "CSRF middleware enforcement", "rate limiting / auto-block", "per-site RBAC enforcement",
      "password complexity validation", "2FA recovery code persistence", "TOTP secret encryption",
      "VictoriaMetrics integration", "Apache web server support", "addon/subdomain management",
      "SSL nginx auto-redirect rewrite", "real file archive extraction", "multipart file upload",
      "database import endpoint", "database remote access toggle", "email autoresponders",
      "mailbox password change", "catch-all Postfix sync", "DKIM key generation",
      "firewall rule delete by UFW number", "geo-blocking", "backup schedule executor",
      "remote backup storage engines", "cron reload after file write", "URL/PHP cron execution",
      "alert engine wiring in panel", "task progress event publishing", "real PTY terminal",
      "terminal idle timeout / root re-auth / IP restriction", "self-update JSON parsing",
      "settings import real insertion", "audit log structured details", "log sanitization",
      "SQLite encryption", "agent capability isolation", "CSP headers / production CORS"
    ]
  }
}
```
