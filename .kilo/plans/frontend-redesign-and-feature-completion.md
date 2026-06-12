# Frontend Redesign & Feature Completion Plan

## Goal

Replace the current minimal light-themed UI with a dense, dark-mode-first developer control panel using the design system below, and implement all missing backend + frontend features required by `docs/API.md` and `docs/MODULES.md`.

## Design System (Authoritative)

| Element | Specification |
|---|---|
| Primary | `#2563EB` (Blue 600) — actions, links, active states |
| Success | `#16A34A` (Green 600) — healthy, deployed, online |
| Warning | `#D97706` (Amber 600) — deploying, pending, attention |
| Danger | `#DC2626` (Red 600) — failed, error, offline, destructive |
| Neutral | `#6B7280` (Gray 500) — inactive, secondary text |
| Background | `#0F172A` (Slate 900) — dark mode default |
| Surface | `#1E293B` (Slate 800) — cards, panels, inputs |
| Border | `#334155` (Slate 700) — dividers, borders |
| Text primary | `#F8FAFC` (Slate 50) — headings, primary content |
| Text secondary | `#94A3B8` (Slate 400) — descriptions, metadata |
| Font UI | Inter |
| Font code/logs/terminal | JetBrains Mono |
| Border radius | Cards 8px, buttons/inputs 6px, tags 4px |
| Shadows | Minimal; subtle glow for focused elements only |

### Principles

1. **Information density over whitespace** — compact tables, tight spacing, monospace for technical content.
2. **Status at a glance** — every service, website, and metric has a color-coded health indicator.
3. **Progressive disclosure** — simple defaults upfront; advanced settings behind collapsible sections.
4. **Action proximity** — the action button is within one click of the data it operates on.
5. **No dead ends** — every empty state has a primary action; every error has a recovery path.
6. **Dark mode default** — light mode optional, not primary.
7. **Motion** — 150ms hover transitions, 200ms modal/drawer transitions; skeletons for initial load, spinners for actions, progress bars for deployments.

---

## Phase 1 — Design System Foundation

**Goal:** establish the new visual language and shared component layer before touching pages.

### Tasks

1. **Tailwind / CSS overhaul**
   - Update `web/tailwind.config.js` with the new color palette, fonts, and border-radius values.
   - Replace `web/src/index.css` with CSS variables for the dark theme and an optional light theme.
   - Add `Inter` and `JetBrains Mono` via Google Fonts or local font files.
   - Add animation utilities for status pulse, skeleton shimmer, and toast slide-up.

2. **Shared component library** (`web/src/components/ui/`)
   - `Button` — primary, secondary, ghost, danger, loading state, icon-only.
   - `Card` — surface card with optional header, footer, and hover state.
   - `Table` — compact table with hover actions, sorting headers, pagination.
   - `Modal` / `Drawer` — square-ish (8px radius) with focus trap and ESC close.
   - `Input`, `Select`, `Textarea`, `Checkbox`, `Switch` — dark theme styles, focus ring in primary blue.
   - `Badge` / `StatusBadge` — success, warning, danger, neutral, pending states.
   - `Tooltip` — black background, white text, 12px font.
   - `Toast` — bottom-right stack, success/error/warning/info variants.
   - `Skeleton` — shimmer for tables, cards, lists.
   - `EmptyState` — icon, title, description, primary action.
   - `PageHeader` — title, subtitle, breadcrumbs, actions.
   - `Tabs` — underline or pill style with active indicator.
   - `SearchInput` — debounced search with clear button.
   - `CopyButton` — copy to clipboard with temporary success state.
   - `ProgressBar` — for tasks and deployments.
   - `HealthIndicator` — dot/pulse + label for live status.

3. **Global state setup**
   - Zustand stores:
     - `authStore` — user, access token, 2FA status, sessions.
     - `uiStore` — sidebar collapsed, theme, command palette open, active toast queue.
   - React Query setup in `web/src/lib/queryClient.ts` with default stale time, retry logic (3 retries, exponential backoff), and global error handling.

4. **API client improvements**
   - Keep axios base in `web/src/lib/api.ts`.
   - Ensure CSRF token extraction works and refresh-token flow redirects to `/login` cleanly.
   - Add typed wrapper: `api.get<T>`, `api.post<T>`, etc.

5. **Error handling**
   - Global error boundary component.
   - 404 page.
   - Toast notifications on mutation errors.

---

## Phase 2 — Core Layout Redesign

**Goal:** rebuild the application shell.

### Tasks

1. **Sidebar** (`web/src/components/layout/Sidebar.tsx`)
   - Dark surface, 240px width, collapsible to 64px.
   - Section grouping: Server (Dashboard, Metrics, Terminal), Hosting (Websites, Databases, Email, Files), Security (Firewall, Backups, Cron, Alerts), System (Settings).
   - Status badges on nav items (e.g., unread alerts count).
   - Active state with blue left border / background tint.
   - Hover tooltips when collapsed.

2. **TopBar** (`web/src/components/layout/TopBar.tsx`)
   - Breadcrumb path for current page.
   - Global search trigger (Cmd+K).
   - Notification bell with dropdown.
   - User avatar menu: profile, sessions, dark/light toggle, logout.
   - Server health summary dot.

3. **Layout** (`web/src/components/layout/Layout.tsx`)
   - Sticky top bar, scrollable main area with consistent padding.
   - Skip-to-content link for accessibility.

4. **Command Palette** (`web/src/components/CommandPalette.tsx`)
   - Cmd+K / Ctrl+K activation.
   - Search actions: "Create website", "Open terminal", "Renew SSL", "Restart nginx", etc.
   - Recent commands and quick navigation.

---

## Phase 3 — Authentication & User Pages

**Goal:** complete login, 2FA, setup, and user management UI.

### Tasks

1. **Login page** (`web/src/pages/Auth/Login.tsx`)
   - Dark themed, centered card, branded header.
   - Support 2FA code input after password step.
   - Show user-friendly error messages.

2. **2FA setup flow**
   - New page/modal for enabling TOTP: show QR code URI, manual secret, verification code input.
   - Recovery codes display.

3. **Setup wizard**
   - First-run setup page: admin user creation, server name, hostname, timezone.

4. **User management** (Settings > Users)
   - List users, create user, edit roles, delete user, force password reset.

---

## Phase 4 — Dashboard Redesign

**Goal:** turn the dashboard into a dense server command center.

### Tasks

1. **Metrics cards**
   - CPU, RAM, Disk, Network, Load average, Processes.
   - Mini sparkline charts using Recharts.
   - Color-coded thresholds (amber >70%, red >90%).

2. **Services status grid**
   - Cards for nginx, php-fpm, mysql, postfix, dovecot, bind9, ssh, juvia-agent.
   - Green / amber / red indicator with restart action.

3. **Alerts panel**
   - Latest unresolved alerts with severity and timestamp.
   - Link to full alerts page.

4. **Websites summary**
   - Count of active / suspended / trashed sites.
   - Quick list of recently added sites.

5. **Quick actions**
   - Create website, open terminal, run backup, add firewall rule.

6. **Recent activity**
   - Audit log preview (last 10 events).

---

## Phase 5 — Websites Module (Redesign + Completion)

**Goal:** fully functional website hosting UI with all detail tabs.

### Frontend Tasks

1. **WebsiteList** (`web/src/pages/Websites/WebsiteList.tsx`)
   - Compact data table with sorting, pagination, search, status filter.
   - Bulk actions: suspend, delete.
   - Inline status badge, SSL expiry countdown, PHP version.
   - Quick action menu (view, suspend, delete, issue SSL).

2. **CreateWebsite** (`web/src/pages/Websites/CreateWebsite.tsx`)
   - Step form: domain, web server, PHP version, advanced options.
   - Advanced section: custom document root, create database, enable SSL later.
   - Real-time domain validation feedback.

3. **WebsiteDetail** (`web/src/pages/Websites/WebsiteDetail.tsx`)
   - Header with domain, status, actions (suspend/restore, delete, visit site).
   - Tabs:
     - **Overview** — status card, runtime info, domains list, quick links to files/terminal/database.
     - **SSL** — certificate status, issuer, expiry, issue/renew/remove buttons, DNS readiness check.
     - **DNS** — full record editor with add/edit/delete, smart suggestions, named-checkzone validation feedback.
     - **Files** — link to file manager.
     - **Apps** — install/update one-click apps (WordPress, Laravel, etc.) with progress.
     - **Git** — setup repo, branch, deploy key, webhook URL copy, pull now, auto-deploy toggle.
     - **Logs** — live tail / download access and error logs.
     - **Advanced** — redirects table, password protection, caching toggles, PHP config values, custom error pages.

4. **Trash** (`web/src/pages/Websites/Trash.tsx`)
   - List trashed sites with days remaining, restore and permanent delete actions, purge all.

### Backend Tasks

1. Implement missing endpoints in `internal/api/websites.go`:
   - `GET /websites/:id/caching` and `PUT /websites/:id/caching`
   - `GET /websites/:id/redirects`, `POST /websites/:id/redirects`, `DELETE /websites/:id/redirects/:rule_id`
   - `GET /websites/:id/password-protection`, `PUT /websites/:id/password-protection`
   - `GET /websites/:id/logs` (unified) if not present
2. Add corresponding agent JSON-RPC methods in `internal/agent/`:
   - `caching.update`
   - `redirect.create/delete`
   - `password_protection.update`
3. Add database migrations if new tables/columns are required.

---

## Phase 6 — File Manager (Completion + Redesign)

**Goal:** production-grade file manager UI.

### Frontend Tasks

1. **FileManager** (`web/src/pages/Files/FileManager.tsx`)
   - Breadcrumb navigation, path bar.
   - Monaco Editor integration for code editing.
   - Right-click context menu: open, edit, rename, copy, move, delete, extract, chmod, download.
   - Drag-and-drop upload with progress overlay.
   - Bulk select with shift-click support.
   - Create new file / folder.
   - Hidden files toggle.
   - File type icons.

2. **FilesIndex** (`web/src/pages/Files/FilesIndex.tsx`)
   - Choose a website to manage files for.

### Backend Tasks

1. Add missing file operations in `internal/api/files.go` and agent:
   - `files.create_directory`
   - `files.move`
   - `files.copy`
   - `files.chmod`
   - Bulk delete/move/copy.
2. Fix upload to use binary content / multipart properly (currently uses `.text()` which corrupts binary files).
3. Add chunked upload support for files > 100MB.

---

## Phase 7 — Email Module (Completion + Redesign)

**Goal:** complete email hosting UI.

### Frontend Tasks

1. **EmailDashboard** (`web/src/pages/Email/EmailDashboard.tsx`)
   - Redesigned tabs with count badges.
   - **Mailboxes** — create/edit/delete, quota display, password reset.
   - **Aliases** and **Forwarders** — clean tables with inline create.
   - **Catch-All** — working save button calling the API.
   - **Autoresponders** — new tab: list, create, edit, enable/disable, date range.
   - **Deliverability** — SPF/DKIM/DMARC/PTR cards with fix instructions and copy buttons.

2. **Webmail** (`web/src/pages/Webmail/Webmail.tsx`)
   - Show install/uninstall status, open webmail button.

### Backend Tasks

1. Implement autoresponder endpoints in `internal/api/email.go`:
   - `GET /email/autoresponders`
   - `POST /email/autoresponders`
   - `PUT /email/autoresponders/:id`
   - `DELETE /email/autoresponders/:id`
2. Add agent methods for autoresponders in `internal/agent/email/postfix.go`.
3. Verify catch-all endpoint wiring works.

---

## Phase 8 — Databases Module (Completion + Redesign)

**Goal:** visual database manager comparable to phpMyAdmin-lite.

### Frontend Tasks

1. **DatabaseManager** (`web/src/pages/Databases/DatabaseManager.tsx`)
   - Sidebar database list with engine badge.
   - **Tables** tab — tree of tables, row browser with pagination, insert/edit/delete rows, table structure.
   - **Users** tab — create/delete users, grant permissions, remote access toggle.
   - **SQL Editor** tab — Monaco Editor for SQL, run query, result table, query history.
   - **Import/Export** — upload SQL file with progress, one-click export download.

2. **CreateDatabase** (`web/src/pages/Databases/CreateDatabase.tsx`)
   - Choose engine (MySQL/PostgreSQL), name, optional initial user.

### Backend Tasks

1. Add missing endpoints in `internal/api/databases.go`:
   - Row insert/edit/delete via `POST /databases/:id/tables/:table/rows` etc.
   - Table structure endpoint.
   - Import endpoint `POST /databases/:id/import`.
2. Add agent methods for row CRUD and import in `internal/agent/database/`.

---

## Phase 9 — Firewall, Backups, Cron, Logs, Alerts

**Goal:** redesign and complete operational modules.

### Frontend Tasks

1. **Firewall** (`web/src/pages/Firewall/Firewall.tsx`)
   - Rules table with status toggle, plain-English description, protocol/port/source.
   - Quick templates: SSH, HTTP, HTTPS, MySQL remote, custom.
   - Geo-blocking section (future).

2. **Backups** (`web/src/pages/Backups/Backups.tsx`)
   - Backup list with type, size, integrity status, restore/delete actions.
   - Create backup modal: full, website, database, files.
   - Schedules tab: create/edit/delete backup schedules, retention settings.
   - Remote storage config (S3/R2/B2/SFTP).

3. **CronJobs** (`web/src/pages/Cron/CronJobs.tsx`)
   - Visual builder: run every [minute/hour/day/week/month] at [time].
   - Advanced mode: raw cron expression.
   - Job types: shell command, URL hit, PHP script.
   - Execution history table with output log.

4. **LogViewer** (`web/src/pages/Logs/LogViewer.tsx`)
   - Live tail with pause/resume.
   - Search/filter, severity highlight.
   - Download button.
   - Support website logs and system logs.

5. **Alerts** (`web/src/pages/Alerts/Alerts.tsx`)
   - Alert list with severity, source, time, acknowledge/delete.
   - Settings: thresholds for CPU, RAM, disk, SSL expiry, website down/slow.

### Backend Tasks

1. **Firewall** — ensure UFW rule enable/disable toggles work; add geo-blocking stub if needed.
2. **Backups** — implement restore verification, schedules CRUD, remote storage config persistence.
3. **Cron** — implement execution history logging, URL hit type, PHP script type.
4. **Logs** — implement website access/error log tail and system log streaming.
5. **Alerts** — implement alert engine background checks (website down/slow, SSL expiry, disk, services) in `internal/alerts/engine.go`.

---

## Phase 10 — Terminal, Metrics, Settings, Services, Updates

**Goal:** complete system-level pages.

### Frontend Tasks

1. **Terminal** (`web/src/pages/Terminal/Terminal.tsx`)
   - xterm.js with dark theme, multiple tabs.
   - Root terminal warning banner.
   - Per-site terminal selector.
   - Idle timeout notice.

2. **Recordings** (`web/src/pages/Terminal/Recordings.tsx`)
   - List recordings with playback via asciinema player.

3. **Metrics** (`web/src/pages/Metrics/Metrics.tsx`)
   - Recharts line/area charts for CPU, RAM, disk, network.
   - Time range selector (1h, 6h, 24h, 7d, 30d).
   - Per-website request metrics when available.

4. **Settings** (`web/src/pages/Settings/Settings.tsx`)
   - Split into sub-pages/tabs:
     - General
     - Email (SMTP)
     - Security (2FA, sessions, login history, terminal IP restrictions)
     - Updates (version, changelog, apply update, rollback)
     - Users
     - Backup config
     - Notifications
     - Audit Log
     - Export/Import

5. **Services** page
   - New page listing all services with status, restart action, install action.

### Backend Tasks

1. **Metrics** — replace stub `historyMetricsHandler` with VictoriaMetrics integration or SQLite-backed history as fallback.
2. **Terminal** — verify PTY session recording to asciinema format.
3. **Updates** — implement self-updater backend in `internal/updater/updater.go`.
4. **Services** — ensure service status and restart endpoints work for all services.
5. **Settings** — persist all settings in database; audit log every change.

---

## Phase 11 — Advanced Features

**Goal:** implement remaining roadmap features.

### Tasks

1. **One-click apps**
   - Expand app catalog: WordPress, WooCommerce, Laravel, Next.js, Ghost, Drupal, Joomla.
   - App install wizard with database selection, admin credentials, progress tracking.
   - Update available badge and one-click update with pre-update backup.

2. **Git deploy**
   - HMAC webhook verification in `internal/api/git_webhook.go`.
   - Deploy history log.

3. **Self-updater UI**
   - Check for updates, download, verify GPG signature, apply, rollback.

4. **Backup integrity verification**
   - Show SHA-256 checksum and last verified timestamp in UI.

---

## Phase 12 — Polish, Accessibility & Testing

**Goal:** make the panel production-ready.

### Tasks

1. **Accessibility**
   - WCAG 2.1 AA compliance: keyboard navigation, focus management, ARIA labels, skip link.
   - Ensure modals trap focus and close on ESC.

2. **Responsive design**
   - ≥1024px: full sidebar.
   - 768–1023px: collapsed sidebar.
   - <768px: hamburger menu, full-screen modals, stacked tables.

3. **Loading & empty states**
   - Skeleton screens for all list/detail pages.
   - Empty states with primary action.

4. **Error handling**
   - Toast on every mutation error.
   - Error boundaries per route.
   - Retry buttons on fetch errors.

5. **Light mode** (optional)
   - Add theme toggle; maintain dark as default.

6. **Testing**
   - Unit tests for shared components.
   - API integration tests for new endpoints.
   - Build and lint checks.

7. **Build verification**
   - `cd web && npm install && npm run build` must succeed.
   - `go build ./...` and `go test ./...` must succeed.

---

## Execution Order

1. Phase 1 (foundation) and Phase 2 (layout) must be completed first — they are prerequisites for all page work.
2. Phase 3 (auth) before any protected page redesign.
3. Phase 4 (dashboard) after layout.
4. Phases 5–8 can overlap (Websites, Files, Email, Databases) because they are mostly independent.
5. Phase 9 (Ops) and Phase 10 (System) after core hosting modules.
6. Phase 11 (Advanced) after Websites and Databases.
7. Phase 12 runs continuously and finalizes at the end.

---

## Success Criteria

- The panel renders in the new dark theme with no leftover light/minimal styling.
- All existing pages are redesigned and functional.
- All backend endpoints documented in `docs/API.md` are implemented and wired to working UI.
- `go build ./...`, `go test ./...`, and `cd web && npm run build` all pass.
- A user can create a website, manage DNS/SSL/files/database/email/firewall/backups/cron from the UI without dead ends.
