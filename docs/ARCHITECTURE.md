# Architecture

## Overview

juvia is a two-process system for managing a Linux server via a web UI.

```
┌─────────────────────────────────────────────────────────────┐
│  PROCESS 1: Panel Server (runs as unprivileged user: juvia)│
│  ┌──────────────────────────────────────────────────────┐   │
│  │  HTTP Server (port 8080)                             │   │
│  │  React UI (embedded via Go embed)                       │   │
│  │  REST API handlers                                     │   │
│  │  WebSocket (metrics stream, terminal proxy)            │   │
│  │  Auth / JWT + refresh tokens │   │
│  │  Task manager (tracks job status, streams progress)    │   │
│  └──────────────────────┬───────────────────────────────┘   │
│                         │ Unix socket (chmod 660)            │
│                         │ /var/run/juvia/agent.sock         │
└─────────────────────────┼───────────────────────────────────┘
                          │
┌─────────────────────────┼───────────────────────────────────┐
│  PROCESS 2: Agent (runs as root via systemd)                │
│  ┌──────────────────────▼───────────────────────────────┐   │
│  │  Unix socket server                                   │   │
│  │  nginx / apache config management                     │   │
│  │  SSL (certbot wrapper)                                │   │
│  │  DNS (BIND9 zone management)                         │   │
│  │  Email (Postfix + Dovecot + Rspamd)                   │   │
│  │  Database (MySQL + PostgreSQL ops)                    │   │
│  │  Firewall (UFW wrapper)                               │   │
│  │  Backup (tar + rsync + S3)                            │   │
│  │  Cron management                                      │   │
│  │  File operations (path-safe)                          │   │
│  │  System metrics reader (/proc)                        │   │
│  │  Terminal PTY spawner                                 │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Process Responsibilities

### Panel Process (unprivileged)
- Serves React UI on port 8080
- Handles all HTTP/REST API requests
- Manages JWT authentication and sessions
- Streams metrics and task progress via WebSocket
- Forwards privileged operations to Agent via Unix socket
- Never runs as root

### Agent Process (root)
- Listens on Unix socket for commands from Panel
- Executes all system operations (nginx, certbot, postfix, etc.)
- Manages config files (atomic swap pattern)
- Reads system metrics (/proc)
- Spawns PTY sessions for terminal
- Runs as root to have permission for all operations

## IPC: JSON-RPC over Unix Socket

Protocol: Simple JSON-RPC over Unix socket.

Request:
```json
{
  "id": "req_abc123",
  "method": "website.create",
  "params": {
    "domain": "mydomain.com",
    "php_version": "8.2"
  }
}
```

Response:
```json
{
  "id": "req_abc123",
  "result": {
    "task_id": "task_xyz789",
    "status": "running"
  },
  "error": null
}
```

Error:
```json
{
  "id": "req_abc123",
  "result": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Domain already exists"
  }
}
```

## Circuit Breaker Pattern

If the agent socket is unreachable:
1. Panel retries with exponential backoff (1s, 2s, 4s, 8s, max 30s)
2. After 3 consecutive failures, mark agent as "offline"
3. Show "Agent offline" indicator in UI
4. Queue non-critical operations for when agent returns
5. Critical operations (SSL issuance) return error immediately

## Task/Job System

Every non-trivial operation is async:

1. User action triggers task creation
2. Panel returns HTTP 202 + task ID immediately
3. Agent executes steps, updating status in SQLite
4. Panel streams progress to browser via WebSocket
5. On completion, toast notification fires

Task states: `pending` → `running` → `done` | `failed`

## Data Storage

|Data|Storage|Location|
|----|-------|--------|
|Users, configs, jobs, tasks, audit log|SQLite (WAL mode)|`/var/lib/juvia/juvia.db`|
|Metrics time-series|VictoriaMetrics|`/var/lib/juvia/metrics/`|
|System and service logs|Plain files|`/var/log/juvia/`|
|Terminal session recordings|Plain files|`/var/log/juvia/terminal/`|

> **SQLite WAL Mode:** Always use `PRAGMA journal_mode=WAL` for concurrent reads during writes.

## Caching Strategy

For read-heavy endpoints:

|Endpoint|Cache Strategy|TTL|
|--------|--------------|---|
|`GET /api/v1/services`|In-memory|60 seconds|
|`GET /api/v1/metrics/current`|In-memory|15 seconds|
|`GET /api/v1/websites`|React Query (stale-while-revalidate)|30 seconds|

Cache invalidation:
- Mutations invalidate related caches
- WebSocket pushes invalidate real-time data

## Event Bus / Pub-Sub

For broadcasting events to WebSocket connections:

```go
// Event types
type EventType string

const (
    EventWebsiteCreated EventType = "website.created"
    EventWebsiteDeleted EventType = "website.deleted"
    EventAlertFired     EventType = "alert.fired"
    EventMetricsUpdate  EventType = "metrics.update"
)

// Pub-sub via Go channels
type EventBus struct {
    subscribers map[string][]chan Event
    mu          sync.RWMutex
}
```

## Config Management: Atomic Swap

All service configs are generated by the Agent and live under `/etc/juvia/`:

```
/etc/juvia/
├── nginx/sites/          # One file per website
├── apache/sites/
├── php-fpm/pools/         # One pool config per website
├── postfix/
├── dovecot/
├── bind/zones/
└── ufw/
```

**Write strategy:**
1. Write new config to `mydomain.com.tmp`
2. Validate: `nginx -t` or `named-checkzone`
3. If valid: atomic `rename()` to `mydomain.com`, then `systemctl reload`
4. If invalid: delete tmp file, return translated error

This ensures bad configs never replace working ones.

## Project Structure

```
juvia/
├── cmd/
│   ├── panel/                    # Panel entry point
│   └── agent/                    # Agent entry point
├── internal/
│   ├── agent/
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
│   ├── api/                      # REST handlers
│   ├── ws/                       # WebSocket handlers
│   ├── tasks/                    # Task runner
│   ├── socket/                   # Unix socket client/server
│   ├── db/                       # SQLite models + migrations
│   ├── auth/                     # JWT + refresh tokens + 2FA
│   ├── alerts/                   # Proactive monitoring
│   ├── registry/                  # Service registry
│   ├── cache/                     # In-memory cache
│   ├── events/                    # Event bus
│   └── updater/                   # Self-update + GPG verification
├── web/                          # React frontend (Go embed)
└── scripts/
    ├── install.sh
    └── build-deb.sh
```

## Directory Structure (installed)

```
/etc/juvia/           # Config files (nginx, php-fpm, etc.)
/var/lib/juvia/      # Data (SQLite DB, metrics)
  ├── juvia.db
  └── metrics/
/var/log/juvia/      # Logs and recordings
  ├── panel.log
  ├── agent.log
  └── terminal/           # Session recordings (.cast files)
/var/run/juvia/      # Runtime files
  └── agent.sock         # Unix socket
```

## Reverse Proxy Configuration

For production SSL termination:

```nginx
# /etc/nginx/sites-available/juvia
server {
    listen 443 ssl;
    server_name panel.example.com;

    ssl_certificate /etc/ssl/certs/panel.crt;
    ssl_certificate_key /etc/ssl/private/panel.key;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Graceful Shutdown

Both panel and agent handle SIGTERM gracefully:

1. Stop accepting new connections
2. Wait for active tasks to complete (max 30 seconds)
3. Close socket connections
4. Flush logs
5. Exit cleanly

```go
// Graceful shutdown with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := server.Shutdown(ctx); err != nil {
    log.Error("Shutdown error:", err)
}
```

## Log Rotation

Logs are rotated using logrotate:

```
/var/log/juvia/*.log {
    daily
    rotate 14
    compress
    delaycompress
    missingok
    notifempty
    create0640 juvia juvia
    postrotate
        systemctl reload juvia > /dev/null 2>&1 || true
    endscript
}
```

## Health Check Endpoint

`GET /api/v1/health` checks:
- SQLite connectivity
- Agent socket connectivity
- VictoriaMetrics connectivity

Returns degraded status if any component fails.

## First-Run Setup Wizard

On first login, users complete a 4-step wizard:
1. Welcome
2. Server name + hostname
3. Notification email + SMTP settings
4. 2FA setup (optional)

After completion, user lands on dashboard with contextual nudge to add first website.

## Future: Multi-Server Support

When multi-server support is added in v2, the Unix socket can be promoted to gRPC/TLS over the network. The interface contract stays the same — only the transport changes.
