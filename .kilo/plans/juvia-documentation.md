# juvia — Documentation Preparation Plan

## Goal
Analyze the `juvia-revised-specification (1).md` and prepare all required markdown documentation files to serve as the implementation blueprint before any code is written.

## Analysis Summary

The specification describes a **free, open-source server control panel** for Ubuntu/Debian, built for non-technical small business owners. It is a large-scale project with:

- **Two-process architecture** (unprivileged Panel + root Agent via Unix socket JSON-RPC)
- **Go backend** with embedded React/Vite frontend
- **20+ modules** across 5 build phases
- **Security-first design** (atomic config swaps, path validation, 2FA, session recording)
- **Modern SaaS UX** inspired by Stripe Dashboard (monochrome, sharp, plain-English)

## Required Markdown Files

The following files will be created in the `docs/` directory to provide the complete implementation reference:

### 1. `README.md`
Project overview, quick start, tech stack summary, and contribution guidelines. One-page summary for anyone joining the project.

### 2. `AGENTS.md`
Critical context for AI agents / developers:
- Build steps and conventions
- Two-process model constraints
- Security rules (never run panel as root, atomic config swaps, path safety)
- Plain-English UI rule enforcement
- Code style (Go + React)
- Testing strategy

### 3. `ARCHITECTURE.md`
Deep-dive into the system architecture:
- Process model (Panel + Agent + systemd)
- IPC protocol (JSON-RPC over Unix socket)
- Data flow (User → Panel → Agent → System)
- Task/job system async flow
- Config management strategy (atomic swap + validation)
- Data storage (SQLite + VictoriaMetrics + plain files)
- Project directory structure

### 4. `DATABASE.md`
Complete SQLite schema for all relational data:
- `users` (with 2FA, roles, sessions)
- `websites` (with soft-delete, suspension)
- `domains`, `dns_records`
- `databases`, `db_users`
- `mailboxes`, `email_forwarders`, `aliases`
- `tasks` (async job tracking)
- `audit_log`
- `firewall_rules`, `cron_jobs`
- `backups`, `backup_schedules`
- `alerts`, `notifications`
- `services` (registry)
- `settings`
- `apps` (one-click installs)
- Indexes and foreign key constraints

### 5. `API.md`
REST API + WebSocket protocol specification:
- Authentication flow (JWT + refresh tokens)
- Endpoint definitions per module (methods, paths, request/response schemas)
- WebSocket channels (metrics stream, task progress, terminal proxy)
- JSON-RPC methods for Panel ↔ Agent
- Error codes and plain-English translation mapping
- Rate limiting rules

### 6. `FRONTEND.md`
Frontend architecture and implementation guide:
- React + Vite + Tailwind + shadcn/ui setup
- Design system enforcement (Geist font, monochrome palette, 0px radius)
- Routing structure (sidebar navigation, detail pages, modals)
- State management (Zustand / React Query)
- Component library overrides (shadcn/ui customizations)
- WebSocket client integration
- Terminal (xterm.js) integration
- File manager (Monaco Editor) integration
- Responsive breakpoints

### 7. `MODULES.md`
Module-by-module implementation notes derived from the spec:
- Website Hosting (Nginx/Apache, PHP-FPM, SSL, caching, deployment)
- Email Hosting (Postfix, Dovecot, Rspamd, deliverability)
- DNS Management (BIND9, zone auto-creation, validation)
- Database Management (MySQL, PostgreSQL, visual manager)
- File Manager (path safety, Monaco, uploads)
- Firewall (UFW wrapper, geo-blocking)
- Backups & Restore (local + S3, integrity verification)
- Cron Jobs (visual builder, execution history)
- Metrics & Monitoring (VictoriaMetrics integration)
- Web Terminal (PTY, session recording, security)
- Alerts & Notifications (proactive checks, delivery channels)
- One-Click App Installs (WordPress, Laravel, etc.)
- Settings & Panel Management (audit log, command palette)

### 8. `SECURITY.md`
Security model and hardening requirements:
- Two-process privilege separation
- Unix socket permissions (660, root:panel)
- Path traversal prevention (`safePath` algorithm)
- Zip bomb protection
- Atomic config validation
- Authentication (JWT lifecycle, 2FA, rate limiting, auto-block)
- Terminal security (session recording, idle timeout, re-auth, IP restriction)
- Backup integrity (SHA-256, dry-run restore)
- Self-update GPG verification
- CSP headers
- Executable upload protection

### 9. `DEPLOYMENT.md`
Installation and deployment guide:
- Installer script (`install.sh`) design
- Debian package (`build-deb.sh`) design
- Systemd service definitions (`juvia.service`, `juvia-agent.service`)
- First-run wizard flow
- Default firewall rules
- User/group creation (`panel` user)
- Directory structure (`/etc/juvia/`, `/var/lib/juvia/`, `/var/log/juvia/`)
- Self-update mechanism
- GPG signing process

### 10. `ROADMAP.md`
Phase-by-phase implementation roadmap with milestones:
- **Phase 1 — Foundation**: Architecture, auth, dashboard, metrics, alerts, registry, command palette
- **Phase 2 — Web Hosting**: Website management, SSL, DNS, config validation, trash
- **Phase 3 — Services**: Email, databases, file manager
- **Phase 4 — Ops**: Firewall, backups, cron, logs, email alerts
- **Phase 5 — Polish**: App installs, git deploy, webmail, terminal recording, packaging, updater

Each phase includes:
- Deliverables
- Module checklist
- Key technical decisions
- Estimated complexity
- Dependencies on previous phases

## Key Technical Decisions

1. **No gRPC for v1** — JSON-RPC over Unix socket is sufficient and has zero dependencies
2. **VictoriaMetrics for metrics** — Single binary, far better than SQLite for time-series
3. **Atomic config swaps** — Every service reload must validate before atomic rename
4. **Soft-delete (trash) for sites** — 30-day recovery window prevents accidents
5. **Plain-English everywhere** — Technical terms must have inline explanations
6. **shadcn/ui + Tailwind with custom overrides** — Matches the sharp monochrome design system
7. **Go embed for frontend** — Single binary deployment

## User Decisions

|Question|Decision|
|--------|--------|
|Scope priority|Full 5-phase roadmap — all phases documented|
|Dev environment|Develop locally; target is remote Linux server `maruf@192.168.0.211` via SSH|
|Doc depth|Concise reference + high-level — readable, exact schemas will live in code/migrations|

## Cross-Check Issues Found

### 1. DATABASE.md — Broken `users` table formatting
Line 17: `|2fa_recovery_codes|JSON||created_at|TIMESTAMP||` — `created_at` is appended to the same row without a newline. Also `created_at` appears twice.

**Fix:** Separate into proper rows with correct column types.

### 2. ARCHITECTURE.md — Inconsistent socket permission
ASCII art says `chmod 600` but spec/AGENTS.md/SECURITY.md/DEPLOYMENT.md all say `660`.

**Fix:** Change `chmod 600` to `660` for `root:panel` access.

### 3. DEPLOYMENT.md — Minor formatting artifact
UFW rules section has `---\n` appearing twice.

**Fix:** Clean up the formatting.

### 4. API.md — Missing endpoints
- Service registry endpoints (`/api/services`) missing
- First-run wizard endpoints missing

**Fix:** Add endpoints for:
- `GET /api/services` (service registry)
- `POST /api/setup/first-run` (wizard step submission)
- `GET /api/setup/status` (check if wizard is needed)

### 5. DATABASE.md — Missing `zones` table
`dns_records` references `zone_id` but no `zones` table defined.

**Fix:** Add `zones` table or change `dns_records` to reference `website_id` instead.

## Best Practices Improvements

1. **Add `docs/CHANGELOG.md` template** — Track changes per phase. Good for accountability.

2. **Add `.kilo/commands/` shortcuts** — Add common commands like `build`, `test`, `lint`, `deploy` for the developer workflow.

3. **Add a `docs/TESTING.md` file** — Testing strategy is mentioned in AGENTS.md but deserves its own doc. Include: unit testing conventions, integration test setup, CI/CD pipeline expectations, manual testing checklist.

4. **Add `docs/CONTRIBUTING.md`** — For an open-source project, contributor guidelines are essential. Include: issue reporting, PR process, code review checklist, commit message conventions.

5. **Cross-reference docs** — Add "See also" links between related docs (e.g., SECURITY.md should link to AGENTS.md for the security rules, DEPLOYMENT.md should link to ARCHITECTURE.md for directory structure).

6. **Add `docs/ENVIRONMENT.md`** — Document the target environment (Ubuntu 22.04+, remote server at `192.168.0.211`, SSH setup, how to build and deploy for that environment).

7. **API.md — Add rate limiting section** — The spec mentions 5 attempts per 10 minutes, but the API doc doesn't document this. Add a section showing which endpoints are rate-limited and how responses look.

8. **ROADMAP.md — Add risk assessment** — For each phase, identify the highest-risk items and mitigation strategies. This helps the team focus on blockers early.

9. **SECURITY.md — Add `ErrPathTraversal` definition** — The error is referenced but not defined. Add a small Go snippet showing the error type.

10. **MODULES.md — Add module dependency graph** — A visual/text representation of which modules depend on which other modules (e.g., Email depends on DNS, which depends on Website).

## Next Steps

1. Fix all 5 cross-check issues
2. Apply best practices improvements (items 1-10)
3. Ensure consistency across all docs
4. Proceed to implementation

## File Output Structure

```
D:\Server panel\
├── docs/
│   ├── README.md
│   ├── AGENTS.md
│   ├── ARCHITECTURE.md
│   ├── DATABASE.md
│   ├── API.md
│   ├── FRONTEND.md
│   ├── MODULES.md
│   ├── SECURITY.md
│   ├── DEPLOYMENT.md
│   ├── ROADMAP.md
│   ├── TESTING.md          (new)
│   ├── ENVIRONMENT.md       (new)
│   └── CONTRIBUTING.md      (new)
├── juvia-revised-specification (1).md
└── .kilo/
    ├── plans/
    │   └── juvia-documentation.md
    └── commands/
        ├── build.md         (new)
        ├── test.md          (new)
        └── deploy.md        (new)
```
