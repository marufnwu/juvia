# Juvia — Documentation Review & Implementation Plan

## Review Findings

I systematically reviewed all 14 documentation files against the original specification and Go/React best practices. Here are the critical findings:

---

## 1. Critical Bugs (Must Fix Before Implementation)

### DATABASE.md — Missing `zone_id` column in `websites` table
**Issue:** `zones` table exists but `websites` has no `zone_id` FK. The `zones` table has `website_id` FK, but the relationship is backwards — a website can have multiple zones (primary domain + addon domains). The `zones` table's `website_id` is correct for this, but the DNS API endpoints (`/api/websites/:id/dns/records`) suggest zones belong to websites. This is actually fine.

**Severity:** Low — the current structure is correct (zones belong to websites via `website_id`).

### DATABASE.md — `sessions` table missing `revoked` column
**Issue:** The spec says "Users can see active sessions and revoke them" but there's no `revoked` or `status` column in the `sessions` table. Sessions need to be revocable without deleting the row.

**Fix:** Add `revoked BOOLEAN DEFAULT FALSE` to `sessions` table.

### DATABASE.md — `users` table missing `deleted_at` or `active` column
**Issue:** No way to disable users. Need `active` or `deleted_at` for soft-deletion.

**Fix:** Add `active BOOLEAN DEFAULT TRUE` or `deleted_at TIMESTAMP NULL`.

### DATABASE.md — `backups` table missing `status` column
**Issue:** No way to track if a backup is `in_progress`, `completed`, `failed`, or `verified`. Currently only `verified` boolean exists.

**Fix:** Add `status VARCHAR(20) DEFAULT 'completed'` with values: `in_progress`, `completed`, `failed`, `verified`.

### DATABASE.md — `mailboxes` table missing `status` and `autoresponder` fields
**Issue:** The spec mentions "Autoresponders" but no columns exist for this. Also no status to disable a mailbox.

**Fix:** Add `status VARCHAR(20)`, `autoresponder_enabled BOOLEAN`, `autoresponder_message TEXT`, `autoresponder_start_date`, `autoresponder_end_date`.

### DATABASE.md — `dns_records` table: `zone_id` should be `NOT NULL`
**Issue:** DNS records must belong to a zone. Allowing NULL could lead to orphaned records.

**Fix:** Add `NOT NULL` to `zone_id`.

### DATABASE.md — `tasks` table: `completed_at` vs `updated_at` ambiguity
**Issue:** No `updated_at` for task progress updates. `completed_at` only covers final state. Need to track last update time.

**Fix:** Add `updated_at TIMESTAMP`.

### DATABASE.md — `services` table: No `health` or `last_error` column
**Issue:** Service registry shows health status but no column stores it. `running` is boolean but doesn't capture "failed" state.

**Fix:** Add `health VARCHAR(20)`: `healthy`, `unhealthy`, `unknown`, `last_error TEXT`.

### DATABASE.md — `apps` table missing `app_type` or `vendor`
**Issue:** No way to distinguish between WordPress, Laravel, etc. types for update checking.

**Fix:** Add `app_type VARCHAR(50)` with enum-like values.

### DATABASE.md — `settings` table: `key` should be indexed for performance
**Issue:** Settings are queried frequently. No index on `key`.

**Fix:** Add `UNIQUE` constraint (already there) which implies index, but should explicitly document this.

### DATABASE.md — `firewall_rules` table: `source` field should support CIDR notation
**Issue:** `VARCHAR(100)` is okay but CIDR needs specific format. Also missing `description` field for plain-English UI.

**Fix:** Add `description VARCHAR(255)` for the human-readable rule name.

### DATABASE.md — `cron_jobs` table: Missing `enabled` column
**Issue:** No way to temporarily disable a cron job without deleting it.

**Fix:** Add `enabled BOOLEAN DEFAULT TRUE`.

### DATABASE.md — `cron_jobs` table: Missing `last_output` column
**Issue:** The spec says "Output log of last 10 runs" but no column stores this. It should store in a separate table or be stored in a JSON column.

**Fix:** Add `last_output TEXT` or create separate `cron_job_logs` table.

### DATABASE.md — `cron_jobs` table: `run_as` field should reference `users.username` or `websites.user_id`
**Issue:** The field says `user or "root"` but no FK constraint. Should reference the system user name.

**Fix:** Add `run_as_user VARCHAR(50)` which stores the Linux username.

### DATABASE.md — `audit_log` table: `details` column type should be JSON
**Issue:** `details` is listed as JSON but SQLite doesn't have native JSON type. Should use TEXT and store JSON.

**Fix:** Add `CHECK(json_valid(details))` constraint or document that TEXT is used with JSON validation.

### DATABASE.md — `users` table: `role` per-site needs a junction table
**Issue:** `per_site` role needs to know which sites. The current schema doesn't support this.

**Fix:** Add `user_sites` junction table: `user_id`, `website_id`, `created_at`.

---

## 2. Critical Gaps (Missing Documentation)

### API.md — Missing `/api/health` endpoint
**Issue:** The spec mentions a health check endpoint but it's not documented.

**Fix:** Add `GET /api/health` — returns `{ "status": "ok", "version": "1.0.0" }`.

### API.md — Missing `/api/websites/:id/trash` endpoint
**Issue:** The spec says soft-delete moves to trash, but the API only has `DELETE` which is the soft-delete. Need a separate trash list endpoint.

**Fix:** Add `GET /api/websites/trash` — list trashed websites, `DELETE /api/websites/:id/permanent` — permanently delete.

### API.md — Missing `/api/websites/:id/logs` endpoint
**Issue:** The spec mentions live log tail but no API endpoint for logs.

**Fix:** Add `GET /api/websites/:id/logs` — query params: `type=access|error`, `tail=N`, `since=ISO8601`.

### API.md — Missing `/api/ssl/check/:id` endpoint
**Issue:** DNS check before SSL issuance needs an endpoint.

**Fix:** Add `GET /api/websites/:id/ssl/check` — returns `{ "dns_ready": true, "message": "DNS is pointing to this server" }`.

### API.md — Missing `/api/email/forwarders` endpoint
**Issue:** The spec mentions email forwarders but no API endpoint.

**Fix:** Add `GET /api/email/forwarders` and `POST /api/email/forwarders`.

### API.md — Missing `/api/email/catch-all` endpoint
**Issue:** Catch-all address per domain is mentioned but not documented.

**Fix:** Add `GET /api/email/catch-all?domain=:domain` and `PUT /api/email/catch-all`.

### API.md — Missing `/api/email/autoresponders` endpoint
**Issue:** Autoresponders are mentioned but not documented.

**Fix:** Add `GET /api/email/autoresponders` and `POST /api/email/autoresponders`.

### API.md — Missing `/api/websites/:id/files` endpoint
**Issue:** The file manager is mentioned but no API endpoints for file operations.

**Fix:** Add `GET /api/websites/:id/files?path=/`, `POST /api/websites/:id/files/upload`, `POST /api/websites/:id/files/download`, `PUT /api/websites/:id/files/rename`, `DELETE /api/websites/:id/files/delete`.

### API.md — Missing `/api/terminal/:id` WebSocket endpoint details
**Issue:** Terminal is mentioned as WebSocket but no specific endpoint documented.

**Fix:** Add `WS /ws/terminal/:sessionId` with message format documentation.

### API.md — Missing `/api/websites/:id/php` endpoint
**Issue:** PHP version switch is mentioned but no endpoint.

**Fix:** Add `PUT /api/websites/:id/php` — body: `{ "version": "8.3" }`.

### API.md — Missing `/api/websites/:id/caching` endpoint
**Issue:** Caching config (FastCGI, Redis) is mentioned but not documented.

**Fix:** Add `GET /api/websites/:id/caching` and `PUT /api/websites/:id/caching`.

### API.md — Missing `/api/websites/:id/redirects` endpoint
**Issue:** Redirect manager is mentioned but not documented.

**Fix:** Add `GET /api/websites/:id/redirects` and `POST /api/websites/:id/redirects`.

### API.md — Missing `/api/websites/:id/password-protection` endpoint
**Issue:** Password-protected directories are mentioned but not documented.

**Fix:** Add `GET /api/websites/:id/password-protection` and `PUT /api/websites/:id/password-protection`.

### API.md — Missing `/api/websites/:id/git` endpoint
**Issue:** Git deploy is mentioned but not documented.

**Fix:** Add `GET /api/websites/:id/git`, `POST /api/websites/:id/git`, `POST /api/websites/:id/git/pull`.

### API.md — Missing `/api/websites/:id/apps` endpoint
**Issue:** One-click app installs are mentioned but not documented.

**Fix:** Add `GET /api/websites/:id/apps`, `POST /api/websites/:id/apps/install`.

### API.md — Missing `/api/alerts/acknowledge` endpoint
**Issue:** Alert acknowledgment is mentioned but no endpoint.

**Fix:** Add `POST /api/alerts/:id/acknowledge`.

### API.md — Missing `/api/alerts/settings` endpoint
**Issue:** Alert threshold configuration is mentioned but not documented.

**Fix:** Add `GET /api/alerts/settings` and `PUT /api/alerts/settings`.

### API.md — Missing `/api/settings/export` and `/api/settings/import` endpoints
**Issue:** Export/Import full panel config is mentioned but not documented.

**Fix:** Add `GET /api/settings/export` and `POST /api/settings/import`.

### API.md — Missing `/api/terminal/recording` endpoint
**Issue:** Session recording is mentioned but not documented.

**Fix:** Add `GET /api/terminal/recordings` and `GET /api/terminal/recordings/:id`.

---

## 3. Security Concerns (Must Address)

### SECURITY.md — Missing CSRF protection for REST API
**Issue:** The panel is stateful (cookies). REST endpoints need CSRF tokens or double-submit cookie pattern.

**Fix:** Add CSRF token validation for all mutating endpoints (POST, PUT, DELETE). Include `X-CSRF-Token` header requirement.

### SECURITY.md — Missing API key / token auth for webhooks
**Issue:** Git deploy webhooks need HMAC signature verification. Document the exact algorithm.

**Fix:** Add webhook security section: HMAC-SHA256 with shared secret, payload verification.

### SECURITY.md — Missing password complexity requirements
**Issue:** No password policy documented. Default passwords could be weak.

**Fix:** Add minimum requirements: 12 characters, 1 uppercase, 1 lowercase, 1 number, 1 special.

### SECURITY.md — Missing account lockout policy
**Issue:** Auto-block after 20 attempts is mentioned, but no lockout duration or notification.

**Fix:** Add: block for 24 hours, email notification to admin, manual unblock in UI.

### SECURITY.md — Missing database encryption at rest
**Issue:** SQLite database contains password hashes, 2FA secrets, email passwords. No encryption documented.

**Fix:** Add SQLite encryption option (SQLCipher) or document that the DB file should be on an encrypted filesystem.

### SECURITY.md — Missing backup encryption
**Issue:** Backups contain sensitive data. No encryption requirement.

**Fix:** Add: backups must be encrypted with AES-256 before transfer to remote storage.

### SECURITY.md — Missing log sanitization
**Issue:** Logs could contain sensitive data (passwords, tokens, session IDs).

**Fix:** Add: all logs must redact passwords, tokens, and personal information before writing.

### SECURITY.md — Missing `sudo` / `su` restrictions for agent
**Issue:** The agent runs as root but could be restricted further. Document principle of least privilege.

**Fix:** Add: Agent should use Linux capabilities instead of full root where possible.

---

## 4. Architecture / Design Concerns

### ARCHITECTURE.md — Missing load balancer / reverse proxy guidance
**Issue:** The spec says port 8080, but production should use nginx as reverse proxy. No documentation.

**Fix:** Add reverse proxy configuration example (nginx) for SSL termination and load balancing.

### ARCHITECTURE.md — Missing database connection pooling
**Issue:** SQLite is single-file but concurrent access needs WAL mode or connection pooling.

**Fix:** Add: SQLite must use WAL mode (`PRAGMA journal_mode=WAL`) for concurrent reads during writes.

### ARCHITECTURE.md — Missing graceful shutdown
**Issue:** No documentation for graceful shutdown (SIGTERM handling).

**Fix:** Add: both panel and agent must handle SIGTERM, finish active tasks, close sockets gracefully.

### ARCHITECTURE.md — Missing health check endpoint
**Issue:** No documented health check for load balancers or systemd.

**Fix:** Add `GET /api/health` that checks SQLite, agent socket, and VictoriaMetrics.

### ARCHITECTURE.md — Missing log rotation
**Issue:** `/var/log/juvia/` could grow indefinitely.

**Fix:** Add logrotate configuration or built-in log rotation.

### ARCHITECTURE.md — Missing circuit breaker pattern for agent
**Issue:** If agent crashes, panel should handle it gracefully. No circuit breaker documented.

**Fix:** Add: panel should retry with exponential backoff, show "Agent offline" in UI.

### ARCHITECTURE.md — Missing caching strategy for API responses
**Issue:** Frequent API calls (e.g., metrics) could overwhelm the backend.

**Fix:** Add: Redis or in-memory caching for read-heavy endpoints like service registry, metrics current.

### ARCHITECTURE.md — Missing event bus / pub-sub for real-time updates
**Issue:** WebSocket is used for metrics and tasks, but no general event bus for other real-time features.

**Fix:** Add: use a pub-sub pattern (e.g., Redis or Go channels) for broadcasting events to WebSocket connections.

---

## 5. Frontend Concerns

### FRONTEND.md — Missing error handling strategy
**Issue:** No documentation for how the frontend handles API errors, network failures, or agent offline states.

**Fix:** Add: global error boundary, retry logic with exponential backoff, offline mode (show cached data with warning).

### FRONTEND.md — Missing PWA / offline capability
**Issue:** Modern SaaS products typically support offline viewing. Not mentioned.

**Fix:** Add: Service Worker for caching dashboard data, offline indicator.

### FRONTEND.md — Missing accessibility (a11y) requirements
**Issue:** The spec is for non-technical users. Accessibility is critical but not mentioned.

**Fix:** Add: WCAG 2.1 AA compliance, keyboard navigation, screen reader support.

### FRONTEND.md — Missing performance budget
**Issue:** No bundle size limit or performance targets.

**Fix:** Add: max initial bundle size 200KB, Time to Interactive < 3s, First Contentful Paint < 1.5s.

### FRONTEND.md — Missing analytics / error tracking
**Issue:** No documentation for production monitoring (Sentry, etc.).

**Fix:** Add: error tracking integration (Sentry), basic usage analytics (opt-in).

### FRONTEND.md — Missing localization (i18n) strategy
**Issue:** The spec is English-only. Non-technical users worldwide may need other languages.

**Fix:** Add: i18n framework (react-i18next), translation key naming conventions.

---

## 6. Testing Concerns

### TESTING.md — Missing performance testing
**Issue:** No load testing or performance benchmarks documented.

**Fix:** Add: k6 or Locust for load testing, benchmark scripts for SQLite queries.

### TESTING.md — Missing security testing
**Issue:** No penetration testing or security scan requirements.

**Fix:** Add: OWASP ZAP scan, fuzz testing for API endpoints, static analysis (gosec, CodeQL).

### TESTING.md — Missing end-to-end (E2E) testing
**Issue:** Manual testing checklist is good but no automated E2E tests.

**Fix:** Add: Playwright for critical user flows (login, create website, SSL, etc.).

### TESTING.md — Missing CI/CD pipeline for the remote server
**Issue:** Development is local but deployment is to remote. No automated deployment pipeline.

**Fix:** Add: GitHub Actions workflow for building, testing, and deploying to `maruf@192.168.0.211`.

---

## 7. Deployment Concerns

### DEPLOYMENT.md — Missing rollback strategy
**Issue:** Self-updater mentions GPG verification but no rollback mechanism if update fails.

**Fix:** Add: keep previous binary as `.backup`, automatic rollback on startup failure.

### DEPLOYMENT.md — Missing data migration for updates
**Issue:** Database schema changes between versions need migration. No documentation.

**Fix:** Add: migration system with version tracking, rollback migrations, data migration scripts.

### DEPLOYMENT.md — Missing backup before update
**Issue:** Self-updater should backup the database before applying update.

**Fix:** Add: automatic SQLite backup before update, restore if update fails.

### DEPLOYMENT.md — Missing SELinux/AppArmor policy
**Issue:** Ubuntu has AppArmor by default. Custom policies may be needed for the socket.

**Fix:** Add: AppArmor profile for the panel and agent processes.

### DEPLOYMENT.md — Missing monitoring setup
**Issue:** No documentation for production monitoring (Prometheus, Grafana, etc.).

**Fix:** Add: Prometheus metrics endpoint `/metrics`, Grafana dashboard template.

### DEPLOYMENT.md — Missing disaster recovery plan
**Issue:** What if the server crashes? How to restore from scratch?

**Fix:** Add: full system backup and restore procedure, documentation for bare-metal recovery.

---

## 8. Documentation Consistency Issues

### README.md → AGENTS.md link broken
**Issue:** `README.md` references `ARCHITECTURE.md` at `docs/ARCHITECTURE.md` but AGENTS.md uses `../docs/ARCHITECTURE.md`. Should be consistent.

**Fix:** All internal links should use relative paths from the file's location.

### AGENTS.md → `panel` user name inconsistency
**Issue:** The spec says user `panel` but the project is named `juvia`. The user should be `juvia`, not `panel`.

**Fix:** Rename system user from `panel` to `juvia` everywhere.

### AGENTS.md → `juvia-agent.service` binary path
**Issue:** The binary is `juvia` (combined), not `juvia-agent`. The service should use the same binary with a flag.

**Fix:** `ExecStart=/usr/local/bin/juvia agent run` or `ExecStart=/usr/local/bin/juvia --agent`.

### DEPLOYMENT.md → `juvia-agent.service` uses wrong binary
**Issue:** `ExecStart=/usr/local/bin/juvia-agent run` but the binary is `juvia`.

**Fix:** Change to `ExecStart=/usr/local/bin/juvia agent run`.

### ENVIRONMENT.md → `panel` user still referenced
**Issue:** `useradd` creates `panel` user. Should be `juvia`.

**Fix:** Change to `useradd -r -s /usr/sbin/nologin -d /var/lib/juvia -M juvia`.

### DATABASE.md → `panel.db` filename inconsistency
**Issue:** The database is named `panel.db` but the project is `juvia`. Should be `juvia.db`.

**Fix:** Rename to `juvia.db` everywhere.

### DATABASE.md → `PANEL_DB` env var still references panel
**Issue:** `PANEL_DB` should be `JUVIA_DB`.

**Fix:** Rename all `PANEL_*` env vars to `JUVIA_*`.

### DATABASE.md → `PANEL_LOG` env var
**Issue:** Same as above.

**Fix:** Rename to `JUVIA_LOG`.

### DATABASE.md → `PANEL_PORT` env var
**Issue:** Same as above.

**Fix:** Rename to `JUVIA_PORT`.

### All docs → `AGENTS.md` naming
**Issue:** The file is named `AGENTS.md` (capitalized, plural) but the content refers to it as a single context. Also, it's not in the docs/ folder.

**Fix:** Either move it to `docs/AGENTS.md` or rename to `AGENTS.md` (keep as is but standardize references).

---

## 9. Specification Compliance Issues

### MODULES.md → Missing "Website Detail Page" specification
**Issue:** The spec shows a detailed website detail page with SSL, Web Server, Caching sections. This is not documented in the MODULES.md.

**Fix:** Add detailed website detail page structure.

### MODULES.md → Missing "Caching Stack" details
**Issue:** The spec mentions FastCGI Cache, Redis Object Cache, Memcached, Built-in Page Cache. MODULES.md only mentions them briefly.

**Fix:** Add implementation details for each caching layer.

### MODULES.md → Missing "OS Isolation" details
**Issue:** The spec says "Each website gets its own Linux system user with an isolated PHP-FPM pool". MODULES.md doesn't document how this is implemented.

**Fix:** Add: `useradd`, `chroot`, `open_basedir`, PHP-FPM pool config details.

### MODULES.md → Missing "Domain Structure" details
**Issue:** The spec mentions primary domain, subdomains, addon domains. MODULES.md doesn't document how these are handled.

**Fix:** Add: domain validation, subdomain creation flow, addon domain handling.

### MODULES.md → Missing "Deployment Options" details
**Issue:** ZIP upload and Git deploy are mentioned but no implementation details.

**Fix:** Add: upload validation, unzip process, git clone process, webhook setup.

### MODULES.md → Missing "Advanced Per-Site Features"
**Issue:** Redirect manager, password-protected directories, hotlink protection, custom error pages are mentioned but not documented.

**Fix:** Add implementation details for each feature.

### MODULES.md → Missing "Logs" details
**Issue:** Live tail and download are mentioned but not documented.

**Fix:** Add: WebSocket streaming, log rotation, download endpoint.

### MODULES.md → Missing "Mail Deliverability Test Tool" details
**Issue:** The spec describes a built-in tool but MODULES.md doesn't document how it works.

**Fix:** Add: test email sending, SPF/DKIM verification, blacklist check, plain-English results.

### MODULES.md → Missing "Smart Suggestions" for DNS
**Issue:** The spec mentions auto-detecting missing records and offering to add them. MODULES.md doesn't document this.

**Fix:** Add: detection algorithm, suggestion logic, auto-add API flow.

### MODULES.md → Missing "Built-In Visual DB Manager" details
**Issue:** The spec describes a modern replacement for phpMyAdmin but MODULES.md doesn't document the implementation.

**Fix:** Add: table browser, row editor, SQL query runner, pagination.

---

## 10. Best Practices Missing

### All docs → Missing API versioning strategy
**Issue:** No `/api/v1/` prefix. Breaking changes would be hard to manage.

**Fix:** Add `v1` prefix to all endpoints: `/api/v1/websites`, `/api/v1/auth/login`, etc.

### All docs → Missing pagination specification
**Issue:** List endpoints (`/api/websites`, `/api/databases`) will return all items. No pagination.

**Fix:** Add pagination: `?page=1&limit=50`, response includes `{ "data": [], "total": 100, "page": 1, "limit": 50 }`.

### All docs → Missing sorting specification
**Issue:** No way to sort lists.

**Fix:** Add `?sort=domain&order=asc`.

### All docs → Missing filtering specification
**Issue:** No way to filter lists.

**Fix:** Add `?status=active&search=mydomain`.

### All docs → Missing rate limiting for non-auth endpoints
**Issue:** Only login rate limiting is documented. API endpoints need rate limiting too.

**Fix:** Add: 100 requests per minute per IP for API, 10 requests per minute for WebSocket.

### All docs → Missing CORS policy
**Issue:** The frontend is embedded but during development, CORS might be needed.

**Fix:** Add CORS configuration for development mode.

### All docs → Missing request/response validation
**Issue:** No documentation for input validation (Zod, JSON schema).

**Fix:** Add: Zod schemas for all API endpoints, validation error format.

### All docs → Missing API response format
**Issue:** No standard response envelope documented.

**Fix:** Add: `{ "success": true, "data": {}, "error": null, "meta": {} }`.

### All docs → Missing WebSocket reconnection strategy
**Issue:** WebSocket connections can drop. No reconnection logic.

**Fix:** Add: exponential backoff reconnection, heartbeat ping/pong.

### All docs → Missing data retention policy
**Issue:** Audit logs, terminal recordings, metrics could grow indefinitely.

**Fix:** Add: audit logs retained for 1 year, terminal recordings for 90 days, metrics for 30 days.

---

## User Decisions

|Question|Decision|
|--------|--------|
|Fix scope|Fix ALL issues across all 10 categories|
|System user|Rename from `panel` to `juvia` everywhere|
|Database file|Rename from `panel.db` to `juvia.db` everywhere|

## Implementation Plan

### Phase A: Database Schema Fixes (DATABASE.md)
1. Add `revoked BOOLEAN` to `sessions` table
2. Add `active BOOLEAN` or `deleted_at` to `users` table
3. Add `status` to `backups` table
4. Add `status`, `autoresponder_*` fields to `mailboxes` table
5. Make `zone_id` NOT NULL in `dns_records`
6. Add `updated_at` to `tasks` table
7. Add `health`, `last_error` to `services` table
8. Add `app_type` to `apps` table
9. Add `description` to `firewall_rules` table
10. Add `enabled` to `cron_jobs` table
11. Add `last_output` or `cron_job_logs` table
12. Add `user_sites` junction table for per-site roles
13. Rename `panel.db` → `juvia.db` everywhere
14. Rename `PANEL_*` env vars → `JUVIA_*`

### Phase B: API Documentation (API.md)
1. Add `/api/health` endpoint
2. Add `/api/websites/trash` endpoints
3. Add `/api/websites/:id/logs` endpoint
4. Add `/api/websites/:id/ssl/check` endpoint
5. Add `/api/email/forwarders` endpoints
6. Add `/api/email/catch-all` endpoints
7. Add `/api/email/autoresponders` endpoints
8. Add `/api/websites/:id/files` endpoints
9. Add `/ws/terminal/:sessionId` WebSocket details
10. Add `/api/websites/:id/php` endpoint
11. Add `/api/websites/:id/caching` endpoint
12. Add `/api/websites/:id/redirects` endpoint
13. Add `/api/websites/:id/password-protection` endpoint
14. Add `/api/websites/:id/git` endpoints
15. Add `/api/websites/:id/apps` endpoints
16. Add `/api/alerts/acknowledge` endpoint
17. Add `/api/alerts/settings` endpoints
18. Add `/api/settings/export` and `/api/settings/import` endpoints
19. Add `/api/terminal/recordings` endpoints
20. Add API versioning: `/api/v1/` prefix
21. Add pagination, sorting, filtering specification
22. Add standard response envelope format
23. Add rate limiting for non-auth endpoints
24. Add CORS policy

### Phase C: Security Fixes (SECURITY.md)
1. Add CSRF protection documentation
2. Add webhook HMAC verification
3. Add password complexity requirements
4. Add account lockout policy
5. Add database encryption at rest
6. Add backup encryption
7. Add log sanitization
8. Add agent privilege restrictions

### Phase D: Architecture Fixes (ARCHITECTURE.md)
1. Add reverse proxy configuration
2. Add SQLite WAL mode
3. Add graceful shutdown
4. Add health check endpoint
5. Add log rotation
6. Add circuit breaker pattern
7. Add caching strategy
8. Add event bus / pub-sub

### Phase E: Frontend Fixes (FRONTEND.md)
1. Add error handling strategy
2. Add PWA / offline capability
3. Add accessibility requirements
4. Add performance budget
5. Add analytics / error tracking
6. Add localization strategy

### Phase F: Testing Fixes (TESTING.md)
1. Add performance testing
2. Add security testing
3. Add E2E testing
4. Add CI/CD pipeline for remote server

### Phase G: Deployment Fixes (DEPLOYMENT.md)
1. Add rollback strategy
2. Add data migration
3. Add backup before update
4. Add SELinux/AppArmor policy
5. Add monitoring setup
6. Add disaster recovery plan

### Phase H: Consistency Fixes (All docs)
1. Rename `panel` → `juvia` (system user, service names, env vars)
2. Rename `panel.db` → `juvia.db`
3. Rename `PANEL_*` → `JUVIA_*`
4. Fix all broken internal links
5. Fix `AGENTS.md` references

### Phase I: Specification Compliance (MODULES.md)
1. Add website detail page specification
2. Add caching stack details
3. Add OS isolation details
4. Add domain structure details
5. Add deployment options details
6. Add advanced per-site features
7. Add logs details
8. Add mail deliverability test tool
9. Add smart suggestions for DNS
10. Add visual DB manager details

### Phase J: Best Practices (All docs)
1. Add API versioning
2. Add pagination, sorting, filtering
3. Add rate limiting for non-auth endpoints
4. Add CORS policy
5. Add request/response validation
6. Add standard response envelope
7. Add WebSocket reconnection
8. Add data retention policy

## Next Steps

After fixing all issues, create a final implementation readiness checklist and call `plan_exit`.
