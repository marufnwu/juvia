# API Reference

**Base URL:** `/api/v1`

All endpoints return a standard response envelope:
```json
{
  "success": true,
  "data": {},
  "error": null,
  "meta": {
    "request_id": "req_abc123",
    "timestamp": "2026-06-11T14:08:50Z"
  }
}
```

Error response:
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input",
    "user_message": "The domain name is invalid. Please check and try again."
  },
  "meta": {
    "request_id": "req_abc123",
    "timestamp": "2026-06-11T14:08:50Z"
  }
}
```

## Authentication

### Login
```
POST /api/v1/auth/login
```
Request:
```json
{
  "username": "admin",
  "password": "••••••"
}
```
Response:
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGc...",
    "expires_in": 900,
    "user": { "id": 1, "username": "admin", "role": "admin" }
  }
}
```
- Access token: 15 minutes lifetime
- Refresh token: stored in httpOnly cookie, 7 days

### Refresh Token
```
POST /api/v1/auth/refresh
```
Uses httpOnly cookie. Returns new access token.

### Logout
```
POST /api/v1/auth/logout
```
Invalidates refresh token.

### Enable2FA
```
POST /api/v1/auth/2fa/enable
```
Returns QR code URI for authenticator app.

### Verify 2FA
```
POST /api/v1/auth/2fa/verify
```
```json
{ "code": "123456" }
```

### Get Current User
```
GET /api/v1/auth/me
```
Returns current user info and session list.

### Revoke Session
```
DELETE /api/v1/auth/sessions/:id
```
Revoke a specific session.

## Health

### Health Check
```
GET /api/v1/health
```
Returns system health status. No authentication required.

```json
{
  "success": true,
  "data": {
    "status": "ok",
    "version": "1.0.0",
    "uptime": 3600,
    "sqlite": "ok",
    "agent": "ok",
    "victoria_metrics": "ok"
  }
}
```

## REST Endpoints

All list endpoints support:
- **Pagination:** `?page=1&limit=50`
- **Sorting:** `?sort=domain&order=asc` (or `desc`)
- **Filtering:** `?status=active&search=mydomain`

Response format for lists:
```json
{
  "success": true,
  "data": [],
  "meta": {
    "total": 100,
    "page": 1,
    "limit": 50,
    "pages": 2
  }
}
```

### Websites

|Method|Path|Description|
|------|----|-----------|
|GET|/api/v1/websites|List all websites|
|POST|/api/v1/websites|Create website|
|GET|/api/v1/websites/:id|Get website details|
|PUT|/api/v1/websites/:id|Update website|
|DELETE|/api/v1/websites/:id|Soft-delete (move to trash)|
|POST|/api/v1/websites/:id/suspend|Suspend website|
|POST|/api/v1/websites/:id/restore|Restore from trash|
|POST|/api/v1/websites/:id/ssl-renew|Renew SSL certificate|
|GET|/api/v1/websites/:id/ssl/check|Check if DNS is ready for SSL|
|GET|/api/v1/websites/:id/logs|Get website logs|
|PUT|/api/v1/websites/:id/php|Change PHP version|
|GET|/api/v1/websites/:id/caching|Get caching config|
|PUT|/api/v1/websites/:id/caching|Update caching config|
|GET|/api/v1/websites/:id/redirects|Get redirect rules|
|POST|/api/v1/websites/:id/redirects|Create redirect rule|
|DELETE|/api/v1/websites/:id/redirects/:rule_id|Delete redirect rule|
|GET|/api/v1/websites/:id/password-protection|Get password protection|
|PUT|/api/v1/websites/:id/password-protection|Update password protection|
|GET|/api/v1/websites/:id/git|Get git deploy config|
|POST|/api/v1/websites/:id/git|Setup git deploy|
|POST|/api/v1/websites/:id/git/pull|Trigger git pull|
|GET|/api/v1/websites/:id/git/webhook|Get webhook URL|
|GET|/api/v1/websites/:id/files|List files|
|POST|/api/v1/websites/:id/files/upload|Upload file(s)|
|GET|/api/v1/websites/:id/files/download|Download file/folder|
|PUT|/api/v1/websites/:id/files/rename|Rename file/folder|
|DELETE|/api/v1/websites/:id/files/delete|Delete file/folder|
|POST|/api/v1/websites/:id/files/extract|Extract archive|
|GET|/api/v1/websites/:id/files/edit|Get file content for editing|
|PUT|/api/v1/websites/:id/files/edit|Save edited file|
|GET|/api/v1/websites/:id/apps|List installed apps|
|POST|/api/v1/websites/:id/apps/install|Install one-click app|
|POST|/api/v1/websites/:id/apps/:app_id/update|Update app|

### Website Trash

|Method|Path|Description|
|------|----|-----------|
|GET|/api/v1/websites/trash|List trashed websites|
|DELETE|/api/v1/websites/:id/permanent|Permanently delete|
|DELETE|/api/v1/websites/trash/purge|Purge all trashed sites|

### DNS

|Method|Path|Description|
|------|----|-----------|
|GET|/api/v1/websites/:id/dns/records|List DNS records|
|POST|/api/v1/websites/:id/dns/records|Create record|
|PUT|/api/v1/dns/records/:id|Update record|
|DELETE|/api/v1/dns/records/:id|Delete record|
|POST|/api/v1/websites/:id/dns/suggest|Check for missing records and suggest|

### Databases

|Method|Path|Description|
|------|----|-----------|
|GET|/api/v1/databases|List databases|
|POST|/api/v1/databases|Create database|
|GET|/api/v1/databases/:id|Get database details|
|DELETE|/api/v1/databases/:id|Delete database|
|POST|/api/v1/databases/:id/users|Create DB user|
|DELETE|/api/v1/databases/:id/users/:user_id|Delete DB user|
|POST|/api/v1/databases/:id/export|Export database|
|POST|/api/v1/databases/:id/import|Import SQL file|
|GET|/api/v1/databases/:id/tables|List tables|
|GET|/api/v1/databases/:id/tables/:table/rows|Get table rows (paginated)|
|POST|/api/v1/databases/:id/query|Run SQL query|

### Email

|Method|Path|Description|
|------|----|-----------|
|GET|/api/v1/email/mailboxes|List mailboxes|
|POST|/api/v1/email/mailboxes|Create mailbox|
|GET|/api/v1/email/mailboxes/:id|Get mailbox details|
|DELETE|/api/v1/email/mailboxes/:id|Delete mailbox|
|PUT|/api/v1/email/mailboxes/:id|Update mailbox|
|GET|/api/v1/email/aliases|List aliases|
|POST|/api/v1/email/aliases|Create alias|
|DELETE|/api/v1/email/aliases/:id|Delete alias|
|GET|/api/v1/email/forwarders|List forwarders|
|POST|/api/v1/email/forwarders|Create forwarder|
|DELETE|/api/v1/email/forwarders/:id|Delete forwarder|
|GET|/api/v1/email/catch-all|Get catch-all for domain|
|PUT|/api/v1/email/catch-all|Set catch-all for domain|
|GET|/api/v1/email/autoresponders|List autoresponders|
|POST|/api/v1/email/autoresponders|Create autoresponder|
|PUT|/api/v1/email/autoresponders/:id|Update autoresponder|
|DELETE|/api/v1/email/autoresponders/:id|Delete autoresponder|
|GET|/api/v1/email/deliverability|Domain deliverability status|

### Firewall

|Method|Path|Description|
|------|----|-----------|
|GET|/api/v1/firewall/rules|List rules|
|POST|/api/v1/firewall/rules|Create rule|
|GET|/api/v1/firewall/rules/:id|Get rule details|
|PUT|/api/v1/firewall/rules/:id|Update rule|
|DELETE|/api/v1/firewall/rules/:id|Delete rule|

### Backups

|Method|Path|Description|
|------|----|-----------|
|GET|/api/v1/backups|List backups|
|POST|/api/v1/backups|Create backup (manual)|
|GET|/api/v1/backups/:id|Get backup details|
|POST|/api/v1/backups/:id/restore|Restore backup|
|DELETE|/api/v1/backups/:id|Delete backup|
|GET|/api/v1/backup-schedules|List schedules|
|POST|/api/v1/backup-schedules|Create schedule|
|GET|/api/v1/backup-schedules/:id|Get schedule details|
|PUT|/api/v1/backup-schedules/:id|Update schedule|
|DELETE|/api/v1/backup-schedules/:id|Delete schedule|

### Cron Jobs

|Method|Path|Description|
|------|----|-----------|
|GET|/api/v1/cron|List cron jobs|
|POST|/api/v1/cron|Create cron job|
|GET|/api/v1/cron/:id|Get cron job details|
|PUT|/api/v1/cron/:id|Update cron job|
|DELETE|/api/v1/cron/:id|Delete cron job|
|POST|/api/v1/cron/:id/enable|Enable cron job|
|POST|/api/v1/cron/:id/disable|Disable cron job|
|GET|/api/v1/cron/:id/logs|Get execution history|

### Metrics

|Method|Path|Description|
|------|----|-----------|
|GET|/api/v1/metrics/current|Get current server metrics|
|GET|/api/v1/metrics/history|Get historical metrics|

Query params: `?range=24h&interval=5m`

### Alerts

|Method|Path|Description|
|------|----|-----------|
|GET|/api/v1/alerts|List alerts|
|POST|/api/v1/alerts/:id/acknowledge|Acknowledge alert|
|DELETE|/api/v1/alerts/:id|Delete alert|
|GET|/api/v1/alerts/settings|Get alert settings|
|PUT|/api/v1/alerts/settings|Update alert settings|

### Settings

|Method|Path|Description|
|------|----|-----------|
|GET|/api/v1/settings|Get all settings|
|PUT|/api/v1/settings|Update settings|
|GET|/api/v1/audit-log|Get audit log entries|
|GET|/api/v1/settings/export|Export full config as JSON|
|POST|/api/v1/settings/import|Import config from JSON|

### Terminal

|Method|Path|Description|
|------|----|-----------|
|POST|/api/v1/terminal/session|Create PTY session|
|GET|/api/v1/terminal/sessions|List active sessions|
|DELETE|/api/v1/terminal/sessions/:id|Close session|
|GET|/api/v1/terminal/recordings|List session recordings|
|GET|/api/v1/terminal/recordings/:id|Get recording playback URL|

### Services (Registry)

|Method|Path|Description|
|------|----|-----------|
|GET|/api/v1/services|List all services and their status|
|GET|/api/v1/services/:name|Get specific service details|
|POST|/api/v1/services/:name/install|Install service|
|POST|/api/v1/services/:name/restart|Restart service|

### Setup

|Method|Path|Description|
|------|----|-----------|
|GET|/api/v1/setup/status|Check if first-run wizard is needed|
|POST|/api/v1/setup/first-run|Submit first-run wizard data|

## WebSocket Channels

### /ws/v1/metrics
Streams live server metrics every 2 seconds.
```json
{
  "type": "metrics",
  "data": { "cpu": 23, "ram": 1288490188, "disk": 45000000000 }
}
```

### /ws/v1/tasks/:taskId
Streams task progress.
```json
{
  "type": "task_progress",
  "data": { "task_id": "task_abc123", "progress": 50, "step": "Configuring SSL" }
}
```
Final message:
```json
{
  "type": "task_complete",
  "data": { "task_id": "task_abc123", "status": "done", "result": {...} }
}
```

### /ws/v1/terminal/:sessionId
Terminal I/O. Send input, receive output.

Message format (client → server):
```json
{
  "type": "input",
  "data": { "data": "ls\r\n" }
}
```

Message format (server → client):
```json
{
  "type": "output",
  "data": { "data": "total 4\ndrwxr-xr-x 2 user user 4096 Jun 11 14:00 .\n" }
}
```

Resize:
```json
{
  "type": "resize",
  "data": { "rows": 24, "cols": 80 }
}
```

## Agent JSON-RPC Methods

Panel → Agent communication over Unix socket.

### website.create
```json
{
  "method": "website.create",
  "params": {
    "domain": "mydomain.com",
    "php_version": "8.2",
    "web_server": "nginx"
  }
}
```

### website.delete
```json
{
  "method": "website.delete",
  "params": { "id": 1, "force": false }
}
```

### ssl.issue
```json
{
  "method": "ssl.issue",
  "params": { "domain": "mydomain.com" }
}
```

### ssl.check
```json
{
  "method": "ssl.check",
  "params": { "domain": "mydomain.com" }
}
```

### config.validate
```json
{
  "method": "config.validate",
  "params": { "type": "nginx", "path": "/etc/juvia/nginx/sites/mydomain.com" }
}
```

### backup.create
```json
{
  "method": "backup.create",
  "params": { "website_id": 1, "type": "full" }
}
```

### terminal.spawn
```json
{
  "method": "terminal.spawn",
  "params": { "user": "mydomain", "rows": 24, "cols": 80 }
}
```

## Rate Limiting

|Endpoint|Limit|Window|
|--------|-----|------|
|`POST /api/v1/auth/login`|5 attempts|per IP per 10 minutes|
|All other API endpoints|100 requests|per IP per minute|
|WebSocket connections|10 connections|per IP per minute|

When rate limited:
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMITED",
    "message": "Too many requests. Please slow down.",
    "retry_after": 60
  }
}
```

After 20 failed login attempts, the IP is automatically added to the UFW deny list.

## CORS Policy

Development mode allows CORS for frontend development server:

```
Access-Control-Allow-Origin: http://localhost:5173
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, X-CSRF-Token
Access-Control-Allow-Credentials: true
```

Production: Panel UI is served from the same origin, no CORS needed.

## CSRF Protection

All mutating endpoints (POST, PUT, DELETE) require:
- `X-CSRF-Token` header with valid token
- Token obtained from `GET /api/v1/auth/me` response header `X-CSRF-Token`

## Error Codes

|Code|Meaning|
|----|-------|
|VALIDATION_ERROR|Invalid input|
|NOT_FOUND|Resource not found|
|UNAUTHORIZED|Session expired or invalid|
|FORBIDDEN|Insufficient permissions|
|CONFLICT|Resource already exists|
|TASK_FAILED|An async task failed|
|RATE_LIMITED|Too many requests|
|SERVER_ERROR|Internal server error|

All errors include a `user_message` field with plain-English translation.
