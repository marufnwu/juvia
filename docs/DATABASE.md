# Database Schema

SQLite database at `/var/lib/juvia/juvia.db`.

> **Note:** SQLite must use WAL mode for concurrent reads during writes:
> ```sql
> PRAGMA journal_mode=WAL;
> ```

## Core Tables

### users
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|username|VARCHAR(255) UNIQUE||
|password_hash|VARCHAR(255)|bcrypt|
|role|VARCHAR(50)|admin, read_only, per_site|
|email|VARCHAR(255)||
|active|BOOLEAN DEFAULT TRUE||soft-disable without deletion|
|2fa_enabled|BOOLEAN||
|2fa_secret|VARCHAR(255)|encrypted TOTP secret|
|2fa_recovery_codes|JSON||
|created_at|TIMESTAMP||
|last_login|TIMESTAMP||
|updated_at|TIMESTAMP||

### user_sites (junction table for per-site roles)
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|user_id|INTEGER FK||
|website_id|INTEGER FK||
|created_at|TIMESTAMP||

### sessions
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|user_id|INTEGER FK||
|token_hash|VARCHAR(255)|hashed refresh token|
|ip_address|VARCHAR(45)||
|user_agent|TEXT||
|revoked|BOOLEAN DEFAULT FALSE||for session revocation|
|expires_at|TIMESTAMP||
|created_at|TIMESTAMP||

### websites
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|domain|VARCHAR(255) UNIQUE||
|document_root|VARCHAR(500)|/home/mydomain/public_html|
|php_version|VARCHAR(10)|8.1, 8.2, 8.3|
|web_server|VARCHAR(20)|nginx or apache|
|ssl_enabled|BOOLEAN||
|ssl_expiry|TIMESTAMP||
|status|VARCHAR(20)|active, suspended, deleted|
|deleted_at|TIMESTAMP||soft-delete for trash|
|user_id|INTEGER FK||linux user owner|
|created_at|TIMESTAMP||
|updated_at|TIMESTAMP||

### domains
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|website_id|INTEGER FK||
|domain|VARCHAR(255)||
|type|VARCHAR(20)|primary, subdomain, addon|
|created_at|TIMESTAMP||

### zones
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|website_id|INTEGER FK||associated website|
|domain|VARCHAR(255)||e.g. mydomain.com|
|serial|INTEGER||zone serial number|
|created_at|TIMESTAMP||
|updated_at|TIMESTAMP||

### dns_records
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|zone_id|INTEGER FK NOT NULL||
|type|VARCHAR(10)|A, AAAA, CNAME, MX, TXT, etc.|
|name|VARCHAR(255)||mydomain.com, www, mail|
|value|TEXT||
|priority|INTEGER||for MX records|
|created_at|TIMESTAMP||
|updated_at|TIMESTAMP||

### databases
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|name|VARCHAR(255)||
|engine|VARCHAR(20)|mysql or postgresql|
|user_id|INTEGER FK||owner|
|created_at|TIMESTAMP||

### db_users
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|database_id|INTEGER FK||
|username|VARCHAR(255)||
|password_hash|VARCHAR(255)||
|host|VARCHAR(50)|localhost|
|created_at|TIMESTAMP||

### mailboxes
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|email|VARCHAR(255) UNIQUE||
|password_hash|VARCHAR(255)||
|quota|INTEGER||bytes|
|display_name|VARCHAR(255)||
|forward_to|VARCHAR(500)||comma-separated|
|status|VARCHAR(20)|active, suspended||
|autoresponder_enabled|BOOLEAN DEFAULT FALSE||
|autoresponder_message|TEXT||
|autoresponder_start_date|TIMESTAMP||
|autoresponder_end_date|TIMESTAMP||
|created_at|TIMESTAMP||

### email_aliases
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|domain|VARCHAR(255)||
|source|VARCHAR(255)||alias address|
|destination|VARCHAR(255)||forwards to|
|created_at|TIMESTAMP||

### email_forwarders
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|domain|VARCHAR(255)||
|source|VARCHAR(255)||forwarder address|
|destination|VARCHAR(255)||receives forwarded email|
|created_at|TIMESTAMP||

### email_catchall
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|domain|VARCHAR(255) UNIQUE||
|forward_to|VARCHAR(500)||comma-separated destinations|
|created_at|TIMESTAMP||

### tasks
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|task_id|VARCHAR(100) UNIQUE||e.g. task_abc123|
|type|VARCHAR(50)||create_website, issue_ssl, etc.|
|status|VARCHAR(20)|pending, running, done, failed|
|progress|INTEGER||0-100|
|steps|JSON||array of step objects|
|result|JSON||final result or error|
|created_by|INTEGER FK||user who initiated|
|created_at|TIMESTAMP||
|updated_at|TIMESTAMP||
|completed_at|TIMESTAMP||

### audit_log
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|user_id|INTEGER FK||
|action|VARCHAR(255)||e.g. "Created website mydomain.com"|
|ip_address|VARCHAR(45)||
|user_agent|TEXT||
|details|TEXT||JSON string (use json_valid() for checking)|
|created_at|TIMESTAMP||

### firewall_rules
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|name|VARCHAR(255)||plain-English name|
|description|VARCHAR(255)||plain-English description|
|action|VARCHAR(10)|allow, deny|
|port|VARCHAR(20)||22, 80, 443, or "All"|
|protocol|VARCHAR(10)|tcp, udp, all|
|source|VARCHAR(100)||IP, CIDR, or country code|
|created_at|TIMESTAMP||

### cron_jobs
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|user_id|INTEGER FK||
|schedule|VARCHAR(100)||cron expression|
|command|TEXT||
|run_as|VARCHAR(50)||linux username or "root"|
|type|VARCHAR(20)||shell, url, php|
|enabled|BOOLEAN DEFAULT TRUE||
|last_run|TIMESTAMP||
|last_status|VARCHAR(20)||success, failed|
|last_output|TEXT||last 10 runs stored in cron_job_logs|
|created_at|TIMESTAMP||

### cron_job_logs
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|cron_job_id|INTEGER FK||
|run_at|TIMESTAMP||
|duration_ms|INTEGER||
|status|VARCHAR(20)||success, failed|
|output|TEXT||stdout/stderr|
|created_at|TIMESTAMP||

### backups
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|website_id|INTEGER FK||nullable for full server|
|type|VARCHAR(20)||full, website, database, files|
|status|VARCHAR(20)|in_progress, completed, failed, verified||
|storage|VARCHAR(20)||local, s3, r2, b2|
|path|TEXT||backup file path or bucket|
|size_bytes|INTEGER||
|checksum|VARCHAR(64)||SHA-256|
|verified|BOOLEAN||
|verified_at|TIMESTAMP||
|created_at|TIMESTAMP||

### backup_schedules
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|website_id|INTEGER FK||nullable|
|schedule|VARCHAR(100)||cron expression|
|retention_days|INTEGER||
|storage|VARCHAR(20)||
|enabled|BOOLEAN||
|created_at|TIMESTAMP||

### alerts
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|type|VARCHAR(50)||ssl_expiring, disk_full, etc.|
|severity|VARCHAR(10)||critical, warning, info|
|message|TEXT||
|resource_id|INTEGER||website_id, etc.|
|resource_type|VARCHAR(50)||website, server, etc.|
|acknowledged|BOOLEAN||
|created_at|TIMESTAMP||

### services
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|name|VARCHAR(100)||nginx, mysql, postfix, etc.|
|version|VARCHAR(50)||
|installed|BOOLEAN||
|running|BOOLEAN||
|health|VARCHAR(20)||healthy, unhealthy, unknown|
|last_error|TEXT||
|checked_at|TIMESTAMP||

### settings
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|key|VARCHAR(255) UNIQUE||
|value|TEXT||
|updated_at|TIMESTAMP||

### apps (one-click installs)
|Column|Type|Notes|
|------|----|-----|
|id|INTEGER PRIMARY KEY||
|app_type|VARCHAR(50)||wordpress, laravel, nextjs, ghost, drupal, joomla|
|name|VARCHAR(100)||display name|
|website_id|INTEGER FK||
|version|VARCHAR(20)||
|installed_at|TIMESTAMP||
|update_available|BOOLEAN||

## Indexes

Key indexes for performance:
- `websites(domain)` — fast lookup
- `tasks(task_id)` — unique task tracking
- `audit_log(user_id, created_at)` — user activity history
- `alerts(acknowledged, created_at)` — alert feed queries
- `mailboxes(email)` — fast email lookup
- `sessions(user_id, revoked)` — session lookup
- `cron_job_logs(cron_job_id, run_at)` — log lookup

## Migrations

Migrations stored in `internal/db/migrations/`. Run on startup.

Naming: `001_initial_schema.sql`, `002_add_soft_delete.sql`, etc.

Apply with:
```go
db.ApplyMigrations(ctx)
```

## Environment Variables

|Variable|Default|Description|
|--------|-------|-----------|
|`JUVIA_DB`|/var/lib/juvia/juvia.db|SQLite database path|
|`JUVIA_LOG`|/var/log/juvia/panel.log|Log file path|
|`JUVIA_PORT`|8080|HTTP port for panel|
|`JUVIA_SOCKET`|/var/run/juvia/agent.sock|Unix socket path|
