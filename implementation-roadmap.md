# Juvia Implementation Roadmap

Derived from `integration-gap-report.md`.

This roadmap groups integration fixes into 5 phases, ordered by risk (low-risk foundational work first, high-risk architectural work later). Each task includes priority, affected files, dependencies, estimated complexity, breaking-change risk, acceptance criteria, and suggested testing.

---

## Executive Summary

| Phase | Theme | Task Count | Focus |
|-------|-------|------------|-------|
| Phase 1 | Foundation & Stability | 8 | UI/UX foundations, route guards, error handling, small fixes |
| Phase 2 | CRUD Completion | 14 | Finish missing create/read/update/delete flows across core modules |
| Phase 3 | Security & Multi-tenancy | 5 | Auth hardening, RBAC, sessions, 2FA, audit log |
| Phase 4 | Real-Time & Observability | 5 | Terminal PTY, recordings, task progress, metrics history, logs |
| Phase 5 | Advanced Features | 6 | Settings, updates, apps, backups, file manager enhancements |

**Critical blockers:**
- `GAP-002` Terminal PTY requires backend agent changes before frontend work can be tested end-to-end.
- `GAP-041` CSRF/rate limiting and `GAP-042` per-site RBAC require backend middleware first.
- `GAP-001` Settings requires all settings-related backend endpoints to be stable.

---

## Dependency Graph

```
Foundation
├── AUTH-001  Token refresh / logout hardening
├── ROUTE-001 404 + error boundaries
├── UI-001    Theme toggle DOM fix
├── NAV-001   Navigation sync
├── FORM-001  Form validation framework
├── ERR-001   Global error/toast pattern
├── LOAD-001  Loading-state standard
└── PAG-001   Pagination components

CRUD Completion
├── DB-001    Database export method fix
├── DB-002    Database user CRUD
├── WEB-001   Website edit / PHP change
├── WEB-002   SSL removal
├── DNS-001   DNS record CRUD
├── MAIL-001  Mailbox edit
├── MAIL-002  Email catch-all
├── CRON-001  Cron edit/logs
├── FW-001    Firewall rule edit
├── APP-001   App install/update
├── FILE-001  File upload binary fix
├── FILE-002  File rename
├── FILE-003  File create folder
└── SERV-001  Service restart

Security & Multi-tenancy
├── SEC-001   CSRF + rate limiting
├── SEC-002   Per-site RBAC
├── AUTH-002  Session revocation
├── AUTH-003  2FA enrollment
└── AUDIT-001 Audit log page

Real-Time & Observability
├── TERM-001  Terminal PTY + WebSocket
├── TERM-002  Session recordings
├── MET-001   Metrics history
├── LOG-001   Per-website logs
└── TASK-001  Task progress WebSocket

Advanced Features
├── SET-001   Settings API integration
├── SET-002   Settings export/import
├── SET-003   Self-update UI
├── BAK-001   Backup schedule CRUD
├── BAK-002   Backup options (type/storage)
└── SETUP-001 First-run wizard
```

---

## Phase 1: Foundation & Stability

Goal: Fix low-risk bugs, establish shared patterns, and prevent regressions before adding major features.

### Phase 1 Task Breakdown

- [ ] **UI-001 Fix theme toggle DOM application**
  - Description: Ensure `dark`/`light` class is applied to `document.documentElement` when theme changes.
  - Priority: Low
  - Files affected: `web/src/stores/uiStore.ts`, `web/src/components/layout/TopBar.tsx`
  - Dependencies: None
  - Estimated complexity: 1 day
  - Breaking change risk: None
  - Acceptance criteria:
    - [ ] Toggling theme updates `<html class="dark">` or `<html class="light">`.
    - [ ] Theme persists across reloads (read from store/localStorage).
    - [ ] Tailwind `dark:` classes render correctly.
  - Suggested testing: Manual visual check in both themes; verify no flash on reload.

- [ ] **NAV-001 Sync sidebar and command palette with router**
  - Description: Add missing routes (Trash, Recordings, Webmail, Users, Audit Log placeholders) to sidebar and command palette; consider generating nav from route config.
  - Priority: Low
  - Files affected: `web/src/components/layout/Sidebar.tsx`, `web/src/components/CommandPalette.tsx`, `web/src/App.tsx`
  - Dependencies: None
  - Estimated complexity: 2 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Sidebar links to all existing routes.
    - [ ] Command palette can navigate to all major features.
    - [ ] Collapsed sidebar tooltips show labels.
  - Suggested testing: Click every sidebar item; verify `Ctrl+K` finds every top-level feature.

- [ ] **ROUTE-001 Add 404 page and error boundary**
  - Description: Add catch-all route and a React error boundary to prevent blank screens and crashes.
  - Priority: Low
  - Files affected: `web/src/App.tsx`, new `web/src/pages/Errors/NotFound.tsx`, new `web/src/components/ErrorBoundary.tsx`
  - Dependencies: None
  - Estimated complexity: 1 day
  - Breaking change risk: None
  - Acceptance criteria:
    - [ ] Unknown URLs show 404 page.
    - [ ] Rendering errors are caught and show fallback UI.
    - [ ] Error boundary logs to console/error reporting.
  - Suggested testing: Navigate to `/unknown`; manually throw in a component.

- [ ] **UI-002 Fix SearchInput debounce timer leak**
  - Description: Refactor debounce to use `useEffect` cleanup instead of returning cleanup from `onChange`.
  - Priority: Low
  - Files affected: `web/src/components/ui/SearchInput.tsx`
  - Dependencies: None
  - Estimated complexity: 0.5 day
  - Breaking change risk: None
  - Acceptance criteria:
    - [ ] Rapid typing does not fire multiple overlapping searches.
    - [ ] Component unmount cancels pending timeout.
  - Suggested testing: Type quickly and verify only final value triggers search; test unmount.

- [ ] **UI-003 Consolidate toast state**
  - Description: Remove orphaned `toasts` array from `uiStore` or migrate `toastStore` into `uiStore`.
  - Priority: Low
  - Files affected: `web/src/stores/uiStore.ts`, `web/src/components/ui/Toast.tsx`
  - Dependencies: None
  - Estimated complexity: 1 day
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Single source of truth for toasts.
    - [ ] `toast()` still works everywhere.
  - Suggested testing: Trigger toasts from multiple pages; verify no duplicates or lost toasts.

- [ ] **ERR-001 Standardize error handling pattern**
  - Description: Replace `alert()` calls with toast + inline error messages; create a reusable `useApiError` hook or helper.
  - Priority: Medium
  - Files affected: All page files, `web/src/lib/api.ts`, `web/src/components/ui/Toast.tsx`
  - Dependencies: UI-003
  - Estimated complexity: 3 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] No `alert()` remains in page code.
    - [ ] API errors display user-friendly messages from `error.user_message`.
    - [ ] Network errors show generic retry message.
  - Suggested testing: Force 401/403/500 responses in dev tools; verify UX.

- [ ] **LOAD-001 Standardize loading states**
  - Description: Ensure all async buttons disable during action and show loading indicator; introduce `useAsyncAction` hook.
  - Priority: Medium
  - Files affected: All page files, `web/src/components/ui/Button.tsx`
  - Dependencies: None
  - Estimated complexity: 3 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Every mutation button has `loading`/`disabled` bound to async state.
    - [ ] No double submissions possible.
    - [ ] Tables show skeleton while loading.
  - Suggested testing: Throttle network; click buttons rapidly; verify disabled state.

- [ ] **FORM-001 Introduce form validation framework**
  - Description: Add Zod schemas for create/edit forms and wire to `Input` error states.
  - Priority: Medium
  - Files affected: `web/package.json`, `web/src/pages/Websites/CreateWebsite.tsx`, `web/src/pages/Databases/CreateDatabase.tsx`, `web/src/pages/Firewall/Firewall.tsx`, `web/src/pages/Cron/CronJobs.tsx`
  - Dependencies: None
  - Estimated complexity: 3 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Domain, email, cron expression, port, CIDR inputs validated client-side.
    - [ ] Submit blocked until valid.
    - [ ] Backend validation still enforced.
  - Suggested testing: Submit invalid forms; verify inline errors and no API call.

---

## Phase 2: CRUD Completion

Goal: Make all existing list/detail pages fully functional for create, read, update, delete.

### Phase 2 Task Breakdown

- [ ] **DB-001 Fix database export HTTP method**
  - Description: Change database export from `window.open GET` to `fetch POST + Blob download` (or change backend to accept GET).
  - Priority: High
  - Files affected: `web/src/pages/Databases/DatabaseManager.tsx:115`, `internal/api/databases.go` (if backend changes)
  - Dependencies: None
  - Estimated complexity: 1 day
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Export button downloads `.sql` file successfully.
    - [ ] Large exports do not block UI.
  - Suggested testing: Export a database; verify file content and name.

- [ ] **DB-002 Implement database user CRUD**
  - Description: Wire "Add User" and "Delete" buttons in DatabaseManager to backend endpoints.
  - Priority: High
  - Files affected: `web/src/pages/Databases/DatabaseManager.tsx`, `internal/api/databases.go`
  - Dependencies: FORM-001, ERR-001, LOAD-001
  - Estimated complexity: 2 days
  - Breaking change risk: None
  - Acceptance criteria:
    - [ ] Add User modal creates DB user via `POST /databases/:id/users`.
    - [ ] Delete button removes user via `DELETE /databases/:id/users/:user_id`.
    - [ ] User list refreshes after mutation.
  - Suggested testing: Create and delete DB user; verify DB reflects changes.

- [ ] **WEB-001 Implement website edit / PHP change**
  - Description: Add controls to update website PHP version and web server; call `PUT /websites/:id` or `PUT /websites/:id/php`.
  - Priority: Medium
  - Files affected: `web/src/pages/Websites/WebsiteDetail.tsx`, `internal/api/websites.go`
  - Dependencies: FORM-001, ERR-001, LOAD-001
  - Estimated complexity: 2 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] User can change PHP version from detail page.
    - [ ] Backend regenerates config (blocked until backend update propagates to agent).
    - [ ] Success/error feedback shown.
  - Suggested testing: Change PHP version; verify agent config and website response.

- [ ] **WEB-002 Add SSL removal action**
  - Description: Add "Remove SSL" button in SSL tab calling `DELETE /websites/:id/ssl`.
  - Priority: Medium
  - Files affected: `web/src/pages/Websites/WebsiteDetail.tsx`, `internal/api/websites.go`
  - Dependencies: ERR-001, LOAD-001
  - Estimated complexity: 1 day
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Button visible only when SSL enabled.
    - [ ] Confirmation modal shown.
    - [ ] SSL status updates after removal.
  - Suggested testing: Issue SSL then remove; verify nginx config no longer serves HTTPS.

- [ ] **DNS-001 Implement DNS record CRUD**
  - Description: Add create/edit/delete UI for DNS records.
  - Priority: High
  - Files affected: `web/src/pages/Websites/WebsiteDetail.tsx`, `internal/api/websites.go`, `internal/api/dns.go`
  - Dependencies: FORM-001, ERR-001, LOAD-001
  - Estimated complexity: 3 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Add Record button opens form and calls `POST /websites/:id/dns/records`.
    - [ ] Each record row has edit/delete actions.
    - [ ] Record list refreshes after mutation.
  - Suggested testing: Add A/MX records; verify zone file and DNS responses.

- [ ] **MAIL-001 Implement mailbox edit**
  - Description: Add edit modal for mailbox quota, display name, and forward_to.
  - Priority: Medium
  - Files affected: `web/src/pages/Email/EmailDashboard.tsx`, `internal/api/email.go`
  - Dependencies: FORM-001, ERR-001, LOAD-001
  - Estimated complexity: 2 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Edit action opens pre-filled modal.
    - [ ] PUT `/email/mailboxes/:id` called on save.
    - [ ] Password change excluded (backend does not support it).
  - Suggested testing: Edit quota/display name; verify DB update.

- [ ] **MAIL-002 Wire email catch-all form**
  - Description: Load and save catch-all configuration.
  - Priority: High
  - Files affected: `web/src/pages/Email/EmailDashboard.tsx`, `internal/api/email.go`
  - Dependencies: ERR-001, LOAD-001
  - Estimated complexity: 1 day
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Catch-all tab loads current config via `GET /email/catch-all`.
    - [ ] Submit calls `PUT /email/catch-all`.
    - [ ] Success/error feedback shown.
  - Suggested testing: Set catch-all; verify DB and Postfix config (when agent sync implemented).

- [ ] **CRON-001 Add cron job edit and logs**
  - Description: Add detail/edit modal and execution log viewer.
  - Priority: Medium
  - Files affected: `web/src/pages/Cron/CronJobs.tsx`, `internal/api/cron.go`
  - Dependencies: FORM-001, ERR-001, LOAD-001
  - Estimated complexity: 2 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Row click opens detail/edit modal.
    - [ ] PUT `/cron/:id` updates job.
    - [ ] Logs tab calls `GET /cron/:id/logs`.
  - Suggested testing: Edit cron expression; run job; view logs.

- [ ] **FW-001 Add firewall rule edit**
  - Description: Add edit modal for existing firewall rules.
  - Priority: Medium
  - Files affected: `web/src/pages/Firewall/Firewall.tsx`, `internal/api/firewall.go`
  - Dependencies: FORM-001, ERR-001, LOAD-001
  - Estimated complexity: 1 day
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Edit action opens pre-filled modal.
    - [ ] PUT `/firewall/rules/:id` called on save.
    - [ ] Rule list refreshes.
  - Suggested testing: Edit a rule port; verify UFW update.

- [ ] **APP-001 Implement app install and update**
  - Description: Render app install modal and add update action for installed apps.
  - Priority: High
  - Files affected: `web/src/pages/Websites/WebsiteDetail.tsx`, `internal/api/apps.go`
  - Dependencies: ERR-001, LOAD-001
  - Estimated complexity: 3 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Install App modal lists available apps and submits `POST /websites/:id/apps/install`.
    - [ ] Update available badge has update action calling `POST /websites/:id/apps/:app_id/update`.
    - [ ] App list refreshes after install/update.
  - Suggested testing: Install WordPress; verify files/DB; trigger update.

- [ ] **FILE-001 Fix binary file upload**
  - Description: Convert file to base64 before upload or implement multipart support.
  - Priority: High
  - Files affected: `web/src/pages/Files/FileManager.tsx`, `internal/api/files.go`
  - Dependencies: ERR-001, LOAD-001
  - Estimated complexity: 2 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Binary files (images, PDFs, zip) upload without corruption.
    - [ ] Large files handled with progress indication.
    - [ ] Backend receives valid content.
  - Suggested testing: Upload binary file; download and compare checksum.

- [ ] **FILE-002 Add file rename action**
  - Description: Add context-menu item and modal for renaming files/folders.
  - Priority: Medium
  - Files affected: `web/src/pages/Files/FileManager.tsx`, `internal/api/files.go`
  - Dependencies: FORM-001, ERR-001, LOAD-001
  - Estimated complexity: 1 day
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Rename action calls `PUT /websites/:id/files/rename`.
    - [ ] File list refreshes.
    - [ ] Validation prevents empty/invalid names.
  - Suggested testing: Rename file/folder; verify filesystem.

- [ ] **FILE-003 Add file create folder endpoint + UI**
  - Description: Add backend endpoint and wire frontend modal.
  - Priority: Medium
  - Files affected: `web/src/pages/Files/FileManager.tsx`, `internal/api/files.go`, `internal/agent/files.go`
  - Dependencies: New backend endpoint
  - Estimated complexity: 2 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Backend exposes `POST /websites/:id/files/mkdir`.
    - [ ] Frontend modal calls endpoint.
    - [ ] Folder appears in file list.
  - Suggested testing: Create folder; verify on filesystem.

- [ ] **SERV-001 Wire dashboard service restart**
  - Description: Add confirmation and POST `/services/:name/restart` to dashboard service restart button.
  - Priority: Medium
  - Files affected: `web/src/pages/Dashboard/Dashboard.tsx`, `internal/api/services.go`
  - Dependencies: ERR-001, LOAD-001
  - Estimated complexity: 1 day
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Restart button triggers confirmation modal.
    - [ ] POST call made with service name.
    - [ ] Service status refreshes after delay.
  - Suggested testing: Restart nginx; verify service state changes.

---

## Phase 3: Security & Multi-tenancy

Goal: Harden auth, enforce roles, and add audit visibility.

### Phase 3 Task Breakdown

- [ ] **SEC-001 Implement CSRF validation and rate limiting**
  - Description: Backend middleware to validate `X-CSRF-Token` and rate-limit auth endpoints; frontend handles 429.
  - Priority: High
  - Files affected: `internal/api/middleware.go`, `internal/api/router.go`, `web/src/lib/api.ts`, `web/src/pages/Auth/Login.tsx`
  - Dependencies: None
  - Estimated complexity: 4 days
  - Breaking change risk: Medium
  - Acceptance criteria:
    - [ ] Mutating requests without CSRF token return 403.
    - [ ] Login endpoint rate-limited by IP.
    - [ ] Frontend retries/renders 429 message.
  - Suggested testing: Omit CSRF header; brute-force login; verify blocking.

- [ ] **SEC-002 Implement per-site RBAC**
  - Description: Enforce `user_sites` ownership in backend handlers; frontend filters lists/actions by role.
  - Priority: High
  - Files affected: `internal/api/middleware.go`, all website handlers, `web/src/App.tsx`, `web/src/pages/Websites/*`, `web/src/stores/authStore.ts`
  - Dependencies: None (backend-first)
  - Estimated complexity: 5 days
  - Breaking change risk: High
  - Acceptance criteria:
    - [ ] `per_site` users only see assigned websites.
    - [ ] Admin users see all websites.
    - [ ] Frontend hides admin-only actions for non-admins.
  - Suggested testing: Create per_site user; verify filtering on backend and UI.

- [ ] **AUTH-002 Add session revocation UI**
  - Description: Display active sessions in Settings > Security and allow revoke.
  - Priority: Medium
  - Files affected: `web/src/pages/Settings/Settings.tsx`, `web/src/stores/authStore.ts`, `internal/api/auth.go`
  - Dependencies: SET-001 (settings API integration) or standalone Security tab
  - Estimated complexity: 2 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Sessions listed with IP, user agent, last active.
    - [ ] Revoke calls `DELETE /auth/sessions/:id`.
    - [ ] Current session cannot be revoked without re-auth.
  - Suggested testing: Log in from two browsers; revoke one; verify logout.

- [ ] **AUTH-003 Implement 2FA enrollment flow**
  - Description: Build QR-code setup using `/auth/2fa/enable` and `/auth/2fa/verify`.
  - Priority: Medium
  - Files affected: `web/src/pages/Settings/Settings.tsx`, `internal/api/auth.go`
  - Dependencies: SET-001 or standalone Security tab
  - Estimated complexity: 3 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Enable 2FA shows QR code and secret.
    - [ ] User verifies code to activate.
    - [ ] Recovery codes displayed once.
  - Suggested testing: Enroll 2FA; log out/in with TOTP code.

- [ ] **AUDIT-001 Build audit log viewer**
  - Description: Add `/audit-log` page with pagination.
  - Priority: Medium
  - Files affected: New `web/src/pages/Audit/AuditLog.tsx`, `web/src/App.tsx`, `web/src/components/layout/Sidebar.tsx`, `internal/api/settings.go`
  - Dependencies: PAG-001, ERR-001, LOAD-001
  - Estimated complexity: 2 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Page lists audit entries with action, user, IP, time.
    - [ ] Pagination uses `page`/`limit` query params.
    - [ ] Accessible to admin only.
  - Suggested testing: Perform mutating actions; verify audit entries.

---

## Phase 4: Real-Time & Observability

Goal: Make terminal, recordings, task progress, metrics history, and logs work end-to-end.

### Phase 4 Task Breakdown

- [ ] **TERM-001 Implement terminal PTY + WebSocket**
  - Description: Backend agent must create real PTY; panel needs `/ws/v1/terminal/:sessionId`; frontend builds xterm.js terminal.
  - Priority: Critical
  - Files affected: `internal/agent/terminal/record.go`, `internal/ws/server.go`, `internal/api/terminal.go`, `web/src/pages/Terminal/Terminal.tsx`, new `web/src/components/terminal/XTermTerminal.tsx`
  - Dependencies: Backend agent PTY implementation
  - Estimated complexity: 10 days
  - Breaking change risk: High
  - Acceptance criteria:
    - [ ] Connect button opens interactive shell.
    - [ ] Input/output streamed over WebSocket.
    - [ ] Sessions recorded to `.cast` files.
    - [ ] Root sessions require explicit warning.
  - Suggested testing: Open site and root terminals; run commands; verify recording playback.

- [ ] **TERM-002 Implement session recordings list/player**
  - Description: Fetch recordings and add asciinema player.
  - Priority: High
  - Files affected: `web/src/pages/Terminal/Recordings.tsx`, `internal/api/terminal.go`
  - Dependencies: TERM-001 (recordings produced)
  - Estimated complexity: 3 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Recordings list fetched from `/terminal/recordings`.
    - [ ] Clicking recording opens player using `/terminal/recordings/:id` content.
    - [ ] Player supports play/pause/seek.
  - Suggested testing: Create terminal session; verify recording appears and plays.

- [ ] **MET-001 Implement metrics history fetch**
  - Description: Fetch historical metrics by time range and render on charts.
  - Priority: High
  - Files affected: `web/src/pages/Metrics/Metrics.tsx`, `internal/api/metrics.go`
  - Dependencies: Backend `/metrics/history` implementation
  - Estimated complexity: 3 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Time-range buttons fetch `/metrics/history?range=...`.
    - [ ] Charts update with historical data.
    - [ ] Fallback to client-side buffer if history unavailable.
  - Suggested testing: Select 24h range; verify data points render.

- [ ] **LOG-001 Implement per-website access/error logs**
  - Description: Update `LogViewer` to call website-specific endpoints.
  - Priority: Medium
  - Files affected: `web/src/pages/Logs/LogViewer.tsx`, `web/src/pages/Websites/WebsiteDetail.tsx`, `internal/api/logs.go`
  - Dependencies: ERR-001, LOAD-001
  - Estimated complexity: 2 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Website logs tab shows access and error logs.
    - [ ] Log lines filterable by level/search.
    - [ ] System logs remain accessible from `/alerts` or new route.
  - Suggested testing: Visit website; verify access log entries appear.

- [ ] **TASK-001 Publish and consume task progress WebSocket**
  - Description: Backend task runner emits progress events; frontend subscribes for long-running operations.
  - Priority: Medium
  - Files affected: `internal/tasks/runner.go`, `internal/events/bus.go`, `internal/ws/server.go`, website/SSL/backup flows
  - Dependencies: Backend event wiring
  - Estimated complexity: 4 days
  - Breaking change risk: Medium
  - Acceptance criteria:
    - [ ] `EventTaskProgress` and `EventTaskComplete` published.
    - [ ] Frontend subscribes to `/ws/v1/tasks/:taskId` during website create/SSL/backup.
    - [ ] Progress shown in UI.
  - Suggested testing: Create website; verify real-time progress updates.

---

## Phase 5: Advanced Features

Goal: Complete settings, updates, backup scheduling, and first-run experience.

### Phase 5 Task Breakdown

- [ ] **SET-001 Wire Settings page to backend APIs**
  - Description: Load and save all settings tabs; integrate `/settings`, `/settings/export`, `/settings/import`.
  - Priority: Critical
  - Files affected: `web/src/pages/Settings/Settings.tsx`, `internal/api/settings.go`
  - Dependencies: ERR-001, LOAD-001, FORM-001
  - Estimated complexity: 6 days
  - Breaking change risk: Medium
  - Acceptance criteria:
    - [ ] Each tab loads real settings via `GET /settings`.
    - [ ] Save buttons persist via `PUT /settings`.
    - [ ] Export/Import buttons trigger downloads/uploads.
    - [ ] Users tab redirects to `/users` (USER-001).
  - Suggested testing: Change timezone/SSH port; verify DB settings updated.

- [ ] **SET-002 Implement self-update UI**
  - Description: Wire `/updates/check`, `/updates/download`, `/updates/apply`, `/updates/rollback` in Settings > Updates.
  - Priority: Medium
  - Files affected: `web/src/pages/Settings/Settings.tsx`, `internal/api/updater.go`
  - Dependencies: SET-001
  - Estimated complexity: 3 days
  - Breaking change risk: Medium
  - Acceptance criteria:
    - [ ] Check for updates shows current and available version.
    - [ ] Download/apply flow with confirmation.
    - [ ] Rollback available after update.
  - Suggested testing: Run update flow in staging; verify binary swap and service restart.

- [ ] **BAK-001 Implement backup schedule CRUD**
  - Description: Add create/delete for backup schedules.
  - Priority: High
  - Files affected: `web/src/pages/Backups/Backups.tsx`, `internal/api/backups.go`
  - Dependencies: FORM-001, ERR-001, LOAD-001
  - Estimated complexity: 2 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Add Schedule modal creates schedule via `POST /backup-schedules`.
    - [ ] Each schedule has delete action via `DELETE /backup-schedules/:id`.
    - [ ] Toggle enabled state if endpoint added.
  - Suggested testing: Create schedule; verify DB entry; delete schedule.

- [ ] **BAK-002 Expose backup type and storage options**
  - Description: Allow selecting full/database/files backup and storage target.
  - Priority: Medium
  - Files affected: `web/src/pages/Backups/Backups.tsx`, `internal/api/backups.go`
  - Dependencies: BAK-001
  - Estimated complexity: 2 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Create backup modal offers type and storage.
    - [ ] Payload sent to backend.
    - [ ] Storage options validated.
  - Suggested testing: Create database-only backup to local storage.

- [ ] **USER-001 Build user management page**
  - Description: Add `/users` route with list/create/edit/delete for admin user management.
  - Priority: High
  - Files affected: New `web/src/pages/Users/Users.tsx`, `web/src/App.tsx`, `web/src/components/layout/Sidebar.tsx`, `internal/api/users.go`
  - Dependencies: FORM-001, ERR-001, LOAD-001, SEC-002 (role guards)
  - Estimated complexity: 4 days
  - Breaking change risk: Medium
  - Acceptance criteria:
    - [ ] Admin can list/create/edit/delete users.
    - [ ] Role selection (admin/user/per_site).
    - [ ] Per-site users can be assigned websites.
  - Suggested testing: Create user; log in as new user; verify permissions.

- [ ] **SETUP-001 Build first-run setup wizard**
  - Description: On app load, check `/setup/status` and redirect to setup if required.
  - Priority: Medium
  - Files affected: `web/src/App.tsx`, new `web/src/pages/Setup/Setup.tsx`, `internal/api/setup.go`
  - Dependencies: FORM-001, ERR-001
  - Estimated complexity: 2 days
  - Breaking change risk: Low
  - Acceptance criteria:
    - [ ] Fresh install redirects to setup wizard.
    - [ ] Wizard creates admin account via `POST /setup/first-run`.
    - [ ] After setup, user is logged in.
  - Suggested testing: Wipe users table; load app; complete setup.

---

## Recommended Execution Order

1. **Phase 1 first** — establishes patterns and prevents regressions.
2. **Phase 2 next** — delivers the highest number of user-visible fixes with low architectural risk.
3. **Phase 3 in parallel with backend team** — security changes require backend-first implementation.
4. **Phase 4 after backend PTY/task events ready** — blocked on agent/backend work.
5. **Phase 5 last** — depends on stable settings/users endpoints.

### Cross-phase dependencies
- `SEC-002` should be implemented before `USER-001` so the Users page can enforce admin-only access.
- `SET-001` should be done before `AUTH-002`, `AUTH-003`, and `SET-002` so the Settings page framework exists.
- `TERM-001` backend PTY must be functional before `TERM-002` recordings have content.
- `PAG-001` should be implemented before `AUDIT-001` and any large-list feature.

---

## Testing Checklist

### Unit / Integration Tests
- [ ] All new hooks/helpers have unit tests.
- [ ] API client interceptors tested with mocked 401/403/429 responses.
- [ ] Form validation schemas tested with valid/invalid inputs.

### Frontend Manual Tests
- [ ] Every sidebar route loads without error.
- [ ] Every command palette command navigates correctly.
- [ ] Create/read/update/delete flows tested for websites, databases, email, firewall, cron, backups.
- [ ] File upload/download/rename/delete/extract tested with binary and text files.
- [ ] SSL issue/renew/remove flow tested.
- [ ] Terminal connect and recording playback tested.
- [ ] Settings save/export/import tested.
- [ ] First-run wizard tested on fresh install.

### Backend Manual Tests
- [ ] CSRF token rejected on mutating requests.
- [ ] Rate limiting triggers after threshold.
- [ ] Per-site user sees only assigned websites.
- [ ] Audit log records mutating actions.
- [ ] Task progress events received by WebSocket client.
- [ ] Metrics history returns data for all ranges.

### End-to-End Tests
- [ ] Full user journey: setup → login → create website → install app → configure SSL → add DNS → upload file → schedule backup → open terminal → view recording → update settings.
- [ ] Negative cases: invalid inputs, insufficient permissions, network failures, service restarts.

---

## Rollout Strategy

### Development
- Use feature branches off `main`.
- Each phase merges via PR with passing `go build ./...`, `go test ./...`, and frontend build.

### Staging
- Deploy each phase to staging environment.
- Run manual integration tests against real Ubuntu 22.04+ server.
- Validate agent/panel two-process model.

### Production
- Phase 1 and Phase 2 can be deployed together as "core fixes."
- Phase 3 security changes require security review.
- Phase 4 terminal changes require careful PTY isolation testing.
- Phase 5 settings/updates should be deployed after core stability proven.

### Monitoring
- Watch panel/agent logs for new errors.
- Monitor WebSocket connection rates and errors.
- Track user-reported issues after each phase rollout.

---

## Task Breakdown Summary

| Task ID | Phase | Description | Priority | Complexity | Risk |
|---------|-------|-------------|----------|------------|------|
| UI-001 | 1 | Theme toggle DOM fix | Low | 1d | None |
| NAV-001 | 1 | Navigation sync | Low | 2d | Low |
| ROUTE-001 | 1 | 404 + error boundary | Low | 1d | None |
| UI-002 | 1 | SearchInput debounce fix | Low | 0.5d | None |
| UI-003 | 1 | Toast state consolidation | Low | 1d | Low |
| ERR-001 | 1 | Error handling standardization | Medium | 3d | Low |
| LOAD-001 | 1 | Loading state standardization | Medium | 3d | Low |
| FORM-001 | 1 | Form validation framework | Medium | 3d | Low |
| DB-001 | 2 | Database export method fix | High | 1d | Low |
| DB-002 | 2 | Database user CRUD | High | 2d | Low |
| WEB-001 | 2 | Website edit / PHP change | Medium | 2d | Low |
| WEB-002 | 2 | SSL removal | Medium | 1d | Low |
| DNS-001 | 2 | DNS record CRUD | High | 3d | Low |
| MAIL-001 | 2 | Mailbox edit | Medium | 2d | Low |
| MAIL-002 | 2 | Email catch-all | High | 1d | Low |
| CRON-001 | 2 | Cron edit/logs | Medium | 2d | Low |
| FW-001 | 2 | Firewall rule edit | Medium | 1d | Low |
| APP-001 | 2 | App install/update | High | 3d | Low |
| FILE-001 | 2 | Binary file upload fix | High | 2d | Low |
| FILE-002 | 2 | File rename | Medium | 1d | Low |
| FILE-003 | 2 | File create folder | Medium | 2d | Low |
| SERV-001 | 2 | Service restart | Medium | 1d | Low |
| SEC-001 | 3 | CSRF + rate limiting | High | 4d | Medium |
| SEC-002 | 3 | Per-site RBAC | High | 5d | High |
| AUTH-002 | 3 | Session revocation | Medium | 2d | Low |
| AUTH-003 | 3 | 2FA enrollment | Medium | 3d | Low |
| AUDIT-001 | 3 | Audit log viewer | Medium | 2d | Low |
| TERM-001 | 4 | Terminal PTY + WebSocket | Critical | 10d | High |
| TERM-002 | 4 | Session recordings | High | 3d | Low |
| MET-001 | 4 | Metrics history | High | 3d | Low |
| LOG-001 | 4 | Per-website logs | Medium | 2d | Low |
| TASK-001 | 4 | Task progress WebSocket | Medium | 4d | Medium |
| SET-001 | 5 | Settings API integration | Critical | 6d | Medium |
| SET-002 | 5 | Self-update UI | Medium | 3d | Medium |
| BAK-001 | 5 | Backup schedule CRUD | High | 2d | Low |
| BAK-002 | 5 | Backup options | Medium | 2d | Low |
| USER-001 | 5 | User management page | High | 4d | Medium |
| SETUP-001 | 5 | First-run wizard | Medium | 2d | Low |

---

## Notes for Executing Agents

- Do not modify code until the phase is started.
- Each task should be a separate PR.
- Run `go build ./...` and `go test ./...` before any backend change.
- Run `cd web && npm run build` before any frontend change.
- Prefer adding new endpoints over changing existing endpoint signatures to minimize regressions.
- For any task blocked by a missing backend capability, flag it and move to the next independent task.
