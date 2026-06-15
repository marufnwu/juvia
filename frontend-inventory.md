# Juvia Frontend Feature Inventory

## Executive Summary

The Juvia frontend is a React 18 + TypeScript single-page application built with Vite, Tailwind CSS, and a custom monochrome shadcn/ui-inspired component layer. It communicates with the panel backend over REST (`/api/v1`) and a single WebSocket stream (`/ws/v1/metrics`).

State management is split between a Zustand auth store, a Zustand UI store, and copious local `useState`. Although React Query is configured in `QueryProvider`, no page currently uses `useQuery`/`useMutation`; every data fetch is hand-rolled inside `useEffect` with local loading/error state. The result is a functional but fragile UI with many half-implemented flows, disconnected forms, and several static/mock screens.

**Overall completion estimate:** ~60% of the visible UI is wired to real API endpoints; the remainder is static, stubbed, or incomplete.

---

## Tech Stack

| Layer | Technology | Notes |
|-------|------------|-------|
| Build tool | Vite | `web/vite.config.ts` (not inspected, inferred from `package.json`) |
| Framework | React 18 | Functional components + hooks |
| Router | React Router v6 | Declared in `web/src/App.tsx` |
| Language | TypeScript | Strict-ish usage, several `any` typed errors |
| Styling | Tailwind CSS | Custom design tokens (`primary`, `surface`, `border`, etc.) |
| State | Zustand | `authStore.ts`, `uiStore.ts` |
| Server state | @tanstack/react-query | Configured but **unused** in pages |
| HTTP client | Axios | `web/src/lib/api.ts` with JWT + refresh interceptor |
| Real-time | Custom `WSClient` | Only used for metrics |
| Icons | lucide-react | — |
| Charts | recharts | Used in `Metrics.tsx` |

---

## Route Matrix

Defined in `web/src/App.tsx:48-80`.

| Route | Component | Protected | Layout | Notes |
|-------|-----------|-----------|--------|-------|
| `/login` | `Login` | No | No | Username/password + 2FA step |
| `/` | `Layout` | Yes | Yes | Redirects to `/dashboard` |
| `/dashboard` | `Dashboard` | Yes | Yes | Server overview + quick actions |
| `/websites` | `WebsiteList` | Yes | Yes | List/search/filter websites |
| `/websites/create` | `CreateWebsite` | Yes | Yes | Multi-step website creation wizard |
| `/websites/:id` | `WebsiteDetail` | Yes | Yes | Tabs: Overview, SSL, DNS, Apps, Git, Logs |
| `/websites/:id/logs` | `LogViewer` | Yes | Yes | Reuses system log viewer for website route |
| `/websites/trash` | `Trash` | Yes | Yes | Soft-deleted websites |
| `/databases` | `DatabaseManager` | Yes | Yes | DB browser, SQL editor, user list |
| `/databases/create` | `CreateDatabase` | Yes | Yes | Create MySQL/PostgreSQL DB |
| `/email` | `EmailDashboard` | Yes | Yes | Mailboxes, aliases, forwarders, deliverability |
| `/email/webmail` | `Webmail` | Yes | Yes | Roundcube install/status |
| `/files` | `FilesIndex` | Yes | Yes | Website picker before file manager |
| `/files/:websiteId` | `FileManager` | Yes | Yes | Browse/edit/upload/delete files |
| `/firewall` | `Firewall` | Yes | Yes | UFW-style rule management |
| `/backups` | `Backups` | Yes | Yes | Backup list + schedule list |
| `/cron` | `CronJobs` | Yes | Yes | Cron job management |
| `/alerts` | `Alerts` | Yes | Yes | Alert list + settings modal |
| `/metrics` | `Metrics` | Yes | Yes | Real-time charts |
| `/terminal` | `Terminal` | Yes | Yes | Session list + root warning |
| `/terminal/recordings` | `Recordings` | Yes | Yes | Static empty state |
| `/settings` | `Settings` | Yes | Yes | Mostly static configuration tabs |

**Route gaps / mismatches:**
- `/users` route exists in backend but not in frontend router.
- `/alerts` route uses `/settings/security` breadcrumb links in `TopBar.tsx:161` but that route does not exist (Settings tabs are local state, not routes).
- Sidebar does not link to `/websites/trash`, `/terminal/recordings`, `/email/webmail`, or `/users`.

---

## Component Matrix

### Layout Components (`web/src/components/layout/`)

| Component | File | Props/Key Logic | Consumers |
|-----------|------|-----------------|-----------|
| `Layout` | `Layout.tsx:8` | Renders `Sidebar`, `TopBar`, `<Outlet />`, `CommandPalette` | `App.tsx:54` |
| `Sidebar` | `Sidebar.tsx:60` | Collapsible nav sections; consumes `useAuthStore.user` | `Layout.tsx:13` |
| `TopBar` | `TopBar.tsx:41` | Breadcrumbs, command-palette trigger, theme toggle, notifications, user menu | `Layout.tsx:15` |

### Global Components

| Component | File | Purpose |
|-----------|------|---------|
| `CommandPalette` | `CommandPalette.tsx:17` | `Ctrl+K` search modal; 9 hard-coded commands |
| `ToastContainer` | `ui/Toast.tsx:88` | Global toast renderer (imperative `toast()` API) |

### UI Primitives (`web/src/components/ui/`)

| Component | File | Exports | Notes |
|-----------|------|---------|-------|
| `Button` | `Button.tsx:11` | `Button` | Variants: primary/secondary/ghost/danger/outline; loading spinner |
| `Input` | `Input.tsx:9` | `Input`, `Textarea`, `Select`, `Checkbox`, `Switch`, `Label`, `FormGroup`, `FormError` | Form primitives; `Switch` is a styled checkbox |
| `Card` | `Card.tsx:11` | `Card`, `CardHeader`, `CardTitle`, `CardDescription`, `CardContent`, `CardFooter` | Padding + hover variants |
| `Table` | `Table.tsx:23` | `Table` | Sortable columns, skeleton loading state |
| `Modal` | `Modal.tsx:15` | `Modal`, `Drawer`, `ConfirmModal` | Escape-to-close, scroll lock |
| `Badge` | `Badge.tsx:31` | `Badge`, `StatusBadge` | Variants + status mapping |
| `Tabs` | `Tabs.tsx:10` | `Tabs`, `TabPanel` | Horizontal tab bar |
| `Skeleton` | `Skeleton.tsx:7` | `Skeleton`, `TableSkeleton`, `CardSkeleton`, `MetricSkeleton`, `ListSkeleton` | Loading placeholders |
| `SearchInput` | `SearchInput.tsx:13` | `SearchInput` | Debounced search with clear button (has timer leak — see Issues) |
| `Toast` | `Toast.tsx:88` | `toast`, `useToast`, `ToastContainer` | External imperative store, not tied to `uiStore.toasts` |
| `Misc` | `Misc.tsx` | `CopyButton`, `ProgressBar`, `HealthIndicator`, `EmptyState`, `PageHeader` | Shared helpers |

---

## API Dependency Matrix

HTTP client: `web/src/lib/api.ts` (`axios.create({ baseURL: '/api/v1' })`).

| Page / Feature | GET endpoints | POST endpoints | PUT endpoints | DELETE endpoints | WebSocket |
|----------------|---------------|----------------|---------------|------------------|-----------|
| `Dashboard` | `/metrics/current`, `/services`, `/alerts`, `/websites` | — | — | — | `ws://host/ws/v1/metrics` |
| `Login` | — | `/auth/login`, `/auth/2fa/verify` | — | — | — |
| `WebsiteList` | `/websites` | `/websites/:id/suspend`, `/websites/:id/restore` | — | `/websites/:id` | — |
| `WebsiteDetail` | `/websites/:id`, `/websites/:id/dns/records`, `/websites/:id/apps`, `/websites/:id/git` | `/websites/:id/ssl`, `/websites/:id/ssl-renew`, `/websites/:id/suspend`, `/websites/:id/git`, `/websites/:id/git/pull` | — | `/websites/:id` | — |
| `CreateWebsite` | — | `/websites` | — | — | — |
| `Trash` | `/websites/trash` | `/websites/:id/restore` | — | `/websites/:id/permanent`, `/websites/trash/purge` | — |
| `DatabaseManager` | `/databases`, `/databases/:id`, `/databases/:id/tables`, `/databases/:id/tables/:name/rows` | `/databases/:id/query` | — | — | — |
| `CreateDatabase` | — | `/databases` | — | — | — |
| `EmailDashboard` | `/email/mailboxes`, `/email/aliases`, `/email/forwarders`, `/email/deliverability?domain=` | `/email/mailboxes`, `/email/aliases`, `/email/forwarders` | — | `/email/mailboxes/:id`, `/email/aliases/:id`, `/email/forwarders/:id` | — |
| `FilesIndex` | `/websites` | — | — | — | — |
| `FileManager` | `/websites/:websiteId/files?path=` | `/websites/:websiteId/files/upload`, `/websites/:websiteId/files/extract` | `/websites/:websiteId/files/edit` | `/websites/:websiteId/files/delete` | — |
| `Firewall` | `/firewall/rules` | `/firewall/rules` | `/firewall/rules/:id` | `/firewall/rules/:id` | — |
| `Backups` | `/backups`, `/backup-schedules` | `/backups`, `/backups/:id/restore` | — | `/backups/:id` | — |
| `CronJobs` | `/cron` | `/cron`, `/cron/:id/enable`, `/cron/:id/disable` | — | `/cron/:id` | — |
| `Alerts` | `/alerts` | `/alerts/:id/acknowledge` | — | `/alerts/:id` | — |
| `LogViewer` | `/logs/system` | — | — | — | — |
| `Webmail` | `/webmail/status` | `/webmail/install`, `/webmail/uninstall` | — | — | — |
| `Terminal` | `/terminal/sessions` | `/terminal/session` | — | `/terminal/sessions/:id` | — |
| `Recordings` | — | — | — | — | — |
| `Metrics` | — | — | — | — | `ws://host/ws/v1/metrics` |
| `Settings` | — | — | — | — | — |
| `authStore` | `/auth/me` | `/auth/logout` | — | — | — |

**API consistency notes:**
- Most pages expect `res.data.data` and fall back to empty arrays. A few expect `res.data.success`.
- `WebsiteDetail.tsx:93` normalizes `website` from `res.data.data.website || res.data.data`.
- `Login.tsx:33` treats either `access_token` or `requires_2fa` as valid first-step responses.

---

## State Management Matrix

| Store | File | State | Actions | Consumers |
|-------|------|-------|---------|-----------|
| `useAuthStore` | `stores/authStore.ts:32` | `user`, `accessToken`, `isAuthenticated`, `sessions` | `setUser`, `setAccessToken`, `logout`, `refreshUser` | `App.tsx`, `Login.tsx`, `TopBar.tsx`, `Sidebar.tsx`, `CommandPalette.tsx` |
| `useUIStore` | `stores/uiStore.ts:22` | `sidebarCollapsed`, `theme`, `commandPaletteOpen`, `toasts` | `toggleSidebar`, `setTheme`, `toggleCommandPalette`, `setCommandPaletteOpen` | `Layout.tsx`, `Sidebar.tsx`, `TopBar.tsx`, `CommandPalette.tsx` |
| Local component state | every page | `loading`, `error`, lists, modals, forms | ad-hoc | every page |
| `toastStore` | `components/ui/Toast.tsx:66` | `toasts[]` | `add`, `remove`, `subscribe` | `ToastContainer.tsx`, `useToast` |

**State issues:**
- `uiStore.toasts` is declared but never populated; `Toast` uses its own isolated store.
- `authStore.sessions` is typed with `ip`, `user_agent`, etc., but `refreshUser` maps `res.data.data.sessions` directly without shape validation.
- `TopBar.tsx:44` reads `sessions` from auth store but never renders them.
- No global error boundary or loading boundary.

---

## Feature Matrix

| Feature | Page(s) | Status | Backend endpoint exists | Notes |
|---------|---------|--------|------------------------|-------|
| Login / logout | `Login`, `TopBar` | Implemented | Yes | JWT in `localStorage`; refresh via cookie |
| 2FA verification | `Login` | Implemented | Needs verification | Frontend handles `requires_2fa` flow |
| 2FA setup | `Settings > Security` | **Not implemented** | Unknown | Button is static |
| Dashboard overview | `Dashboard` | Implemented | Yes | Polls every 30s + WebSocket metrics |
| Website list + CRUD | `WebsiteList`, `CreateWebsite`, `Trash` | Mostly implemented | Yes | Soft delete, restore, suspend |
| Website detail | `WebsiteDetail` | Partial | Yes | Tabs render; DNS add not wired; app install modal stubbed |
| SSL issue/renew | `WebsiteDetail` | Implemented | Yes | `/ssl`, `/ssl-renew` |
| Git deployment | `WebsiteDetail` | Implemented | Yes | Config + pull + webhook URL |
| DNS records view | `WebsiteDetail` | Implemented | Yes | Add-record button is a no-op |
| Database CRUD + SQL editor | `DatabaseManager`, `CreateDatabase` | Mostly implemented | Yes | Add/delete DB user is UI-only |
| Email mailboxes/aliases/forwarders | `EmailDashboard` | Implemented | Yes | Catch-all tab is UI-only |
| Email deliverability check | `EmailDashboard` | Implemented | Yes | SPF/DKIM/DMARC/PTR |
| Webmail install | `Webmail` | Implemented | Yes | Roundcube install/uninstall |
| File manager | `FilesIndex`, `FileManager` | Partial | Yes | Create-folder modal no-op; upload reads `.text()` (binary broken); extract icon wrong |
| Firewall rules | `Firewall` | Implemented | Yes | Toggle + create + delete |
| Backups | `Backups` | Partial | Yes | Schedules list rendered; add schedule no-op |
| Cron jobs | `CronJobs` | Implemented | Yes | Enable/disable/delete/create |
| Alerts | `Alerts` | Partial | Yes | Acknowledge/delete; settings modal no-op |
| System logs | `LogViewer` | Implemented | Yes | `/logs/system` |
| Metrics | `Metrics` | Partial | Yes | Real-time only; time-range selector does not fetch history |
| Terminal sessions | `Terminal` | Partial | Yes | Create/close sessions; **connect button is no-op** |
| Session recordings | `Recordings` | **Not implemented** | Unknown | Empty static page |
| Settings | `Settings` | **Mostly static** | No visible API | All tabs use `defaultValue`/static data |
| User management | `Settings > Users` | **Not implemented** | Yes (backend) | Static mock user card |
| Theme toggle | `TopBar` | Implemented | N/A | Toggles `dark`/`light` class in store only (no `<html>` class application inspected) |
| Command palette | `CommandPalette` | Implemented | N/A | Missing many routes |

---

## Missing / Incomplete Features

### Critical gaps
1. **`Settings.tsx` is 99% static.** No API calls. General/Security/Email/Updates/Users/Backup/Notifications tabs all use `defaultValue` and static buttons.
2. **`Recordings.tsx` is an empty placeholder.** No recordings API consumed; `mockRecordings` is an empty array.
3. **Terminal session connection is not implemented.** `Terminal.tsx:157` "Connect" button has no `onClick` and no WebSocket terminal component.
4. **`FileManager.tsx` create-folder modal does not call the API.** `Modal` confirm button just closes the modal (`setShowNewFolderModal(false)`).
5. **React Query is configured but unused.** Every page re-implements fetching, caching, loading, error, and optimistic updates.

### Functional bugs / rough edges
6. **`SearchInput.tsx:26`** debounce creates a timeout but the cleanup function is returned from `onChange`, not from `useEffect`, so timers leak.
7. **`FileManager.tsx:81`** uploads use `await e.target.files[0].text()`, which corrupts binary files.
8. **`FileManager.tsx:150-152`** range selection with Shift is commented/empty.
9. **`FileManager.tsx:291-298**` extract action uses `<Download>` icon instead of an archive/extract icon.
10. **`WebsiteDetail.tsx:393-396`** "Add Record" button is a no-op.
11. **`WebsiteDetail.tsx:430-482`** Install App modal is opened but never rendered (`showInstallModal` state unused in JSX).
12. **`EmailDashboard.tsx:290-314`** Catch-all form has no submit handler.
13. **`Alerts.tsx:170-228`** Alert settings modal sliders are uncontrolled and not saved.
14. **`Backups.tsx:216-258`** "Add Schedule" button is a no-op.
15. **`Metrics.tsx:74-93`** Time range buttons (`1h`, `6h`, etc.) only change local state; no historical data fetch.
16. **`TopBar.tsx:161`** Links to `/settings/security` which is not a route.
17. **`ProtectedRoute.tsx:29-35`** Only checks authentication, not role/permissions.
18. **`api.ts:16-19`** Reads CSRF token from a `<meta>` tag that the backend HTML may not inject.

### UX / navigation gaps
19. Sidebar omits `/websites/trash`, `/terminal/recordings`, `/email/webmail`, `/users`, `/logs`.
20. Command palette only has 9 commands and omits Files, Firewall, Backups, Cron, Alerts, Metrics, Recordings, Trash, Webmail.
21. No global 404 / catch-all route in `App.tsx`.
22. No error boundaries.

---

## Architecture Summary

```
web/src/
├── App.tsx                 # Router + ProtectedRoute + auth refresh
├── main.tsx                # React mount (assumed)
├── lib/
│   ├── api.ts              # Axios client, JWT, refresh, CSRF
│   ├── queryClient.tsx     # React Query setup (unused by pages)
│   ├── utils.ts            # cn, formatBytes, formatDate, etc.
│   └── ws.ts               # Generic WSClient with reconnect
├── stores/
│   ├── authStore.ts        # Auth state + user/session actions
│   └── uiStore.ts          # Theme, sidebar, command palette (orphaned toasts field)
├── components/
│   ├── layout/             # Layout, Sidebar, TopBar
│   ├── CommandPalette.tsx  # Global search/commands
│   └── ui/                 # 10 primitive component files
└── pages/                  # 21 page components
```

### Auth flow
1. `Login.tsx` POST `/auth/login`.
2. On `access_token`, `setAccessToken` writes `localStorage` and navigates to `/dashboard`.
3. On `requires_2fa`, switches to 2FA form and POST `/auth/2fa/verify`.
4. `api.ts` attaches `Bearer` token and refreshes on 401 via `/auth/refresh`.
5. `AppRoutes.tsx:41-45` calls `refreshUser()` when authenticated, which GET `/auth/me`.

### Data-fetching pattern
Every page follows the same pattern:
```ts
const [data, setData] = useState([])
const [loading, setLoading] = useState(true)
useEffect(() => { api.get('/x').then(...) }, [])
```
No shared loading/error components, no retry UI, no stale-while-revalidate.

### Real-time
Only `Dashboard.tsx:58` and `Metrics.tsx:38` create `WSClient` for `/ws/v1/metrics`. Terminal sessions do not use WebSocket despite the presence of `ws.ts`.

---

## Machine-Readable JSON Inventory

```json
{
  "project": "juvia",
  "inventory_type": "frontend",
  "generated": "2026-06-13",
  "tech_stack": {
    "framework": "React 18",
    "language": "TypeScript",
    "build_tool": "Vite",
    "router": "React Router v6",
    "styling": "Tailwind CSS",
    "state_client": "Zustand",
    "state_server": "@tanstack/react-query (configured, unused)",
    "http": "axios",
    "websocket": "custom WSClient",
    "icons": "lucide-react",
    "charts": "recharts"
  },
  "routes": [
    { "path": "/login", "component": "Login", "protected": false },
    { "path": "/", "component": "Layout", "protected": true },
    { "path": "/dashboard", "component": "Dashboard", "protected": true },
    { "path": "/websites", "component": "WebsiteList", "protected": true },
    { "path": "/websites/create", "component": "CreateWebsite", "protected": true },
    { "path": "/websites/:id", "component": "WebsiteDetail", "protected": true },
    { "path": "/websites/:id/logs", "component": "LogViewer", "protected": true },
    { "path": "/websites/trash", "component": "Trash", "protected": true },
    { "path": "/databases", "component": "DatabaseManager", "protected": true },
    { "path": "/databases/create", "component": "CreateDatabase", "protected": true },
    { "path": "/email", "component": "EmailDashboard", "protected": true },
    { "path": "/email/webmail", "component": "Webmail", "protected": true },
    { "path": "/files", "component": "FilesIndex", "protected": true },
    { "path": "/files/:websiteId", "component": "FileManager", "protected": true },
    { "path": "/firewall", "component": "Firewall", "protected": true },
    { "path": "/backups", "component": "Backups", "protected": true },
    { "path": "/cron", "component": "CronJobs", "protected": true },
    { "path": "/alerts", "component": "Alerts", "protected": true },
    { "path": "/metrics", "component": "Metrics", "protected": true },
    { "path": "/terminal", "component": "Terminal", "protected": true },
    { "path": "/terminal/recordings", "component": "Recordings", "protected": true },
    { "path": "/settings", "component": "Settings", "protected": true }
  ],
  "components": {
    "layout": ["Layout", "Sidebar", "TopBar"],
    "global": ["CommandPalette", "ToastContainer"],
    "ui": [
      "Button", "Input", "Textarea", "Select", "Checkbox", "Switch",
      "Label", "FormGroup", "FormError", "Card", "CardHeader", "CardTitle",
      "CardDescription", "CardContent", "CardFooter", "Table", "Modal",
      "Drawer", "ConfirmModal", "Badge", "StatusBadge", "Tabs", "TabPanel",
      "Skeleton", "TableSkeleton", "CardSkeleton", "MetricSkeleton",
      "ListSkeleton", "SearchInput", "Toast", "CopyButton", "ProgressBar",
      "HealthIndicator", "EmptyState", "PageHeader"
    ]
  },
  "stores": [
    {
      "name": "authStore",
      "file": "web/src/stores/authStore.ts",
      "state": ["user", "accessToken", "isAuthenticated", "sessions"],
      "actions": ["setUser", "setAccessToken", "setSessions", "logout", "refreshUser"]
    },
    {
      "name": "uiStore",
      "file": "web/src/stores/uiStore.ts",
      "state": ["sidebarCollapsed", "theme", "commandPaletteOpen", "toasts"],
      "actions": ["toggleSidebar", "setSidebarCollapsed", "setTheme", "toggleCommandPalette", "setCommandPaletteOpen"]
    }
  ],
  "api_surface": {
    "base_url": "/api/v1",
    "auth": ["/auth/login", "/auth/2fa/verify", "/auth/refresh", "/auth/logout", "/auth/me"],
    "websites": [
      "/websites", "/websites/:id", "/websites/:id/suspend", "/websites/:id/restore",
      "/websites/:id/ssl", "/websites/:id/ssl-renew", "/websites/:id/dns/records",
      "/websites/:id/apps", "/websites/:id/git", "/websites/:id/git/pull",
      "/websites/:id/files", "/websites/:id/files/upload", "/websites/:id/files/edit",
      "/websites/:id/files/delete", "/websites/:id/files/download", "/websites/:id/files/extract",
      "/websites/trash", "/websites/trash/purge", "/websites/:id/permanent"
    ],
    "databases": ["/databases", "/databases/:id", "/databases/:id/tables", "/databases/:id/tables/:name/rows", "/databases/:id/query"],
    "email": ["/email/mailboxes", "/email/aliases", "/email/forwarders", "/email/deliverability", "/webmail/status", "/webmail/install", "/webmail/uninstall"],
    "firewall": ["/firewall/rules"],
    "backups": ["/backups", "/backup-schedules", "/backups/:id/restore"],
    "cron": ["/cron", "/cron/:id/enable", "/cron/:id/disable"],
    "alerts": ["/alerts", "/alerts/:id/acknowledge"],
    "logs": ["/logs/system"],
    "terminal": ["/terminal/sessions", "/terminal/session"],
    "metrics": ["ws://host/ws/v1/metrics"]
  },
  "missing_features": [
    "Settings page API integration",
    "Session recordings list/player",
    "Terminal WebSocket shell connection",
    "Role-based route guards",
    "2FA enrollment/setup UI",
    "User management CRUD",
    "Backup schedule creation",
    "Email catch-all configuration submit",
    "DNS record creation",
    "Website app installation flow",
    "File manager folder creation",
    "Historical metrics by time range",
    "Global error boundary",
    "404 catch-all route"
  ],
  "notable_bugs": [
    "SearchInput debounce timer leak",
    "FileManager binary upload uses File.text()",
    "FileManager extract button uses Download icon",
    "uiStore.toasts is unused (Toast has separate store)",
    "React Query configured but unused; pages use hand-rolled fetch",
    "Settings tabs linked from TopBar via non-existent /settings/security route"
  ]
}
```

---

## Recommendations

1. **Adopt React Query** for server state to remove boilerplate and add caching/refetch/error handling.
2. **Implement Settings API** or remove the page until backend endpoints exist.
3. **Wire Terminal to a real WebSocket shell** using the existing `WSClient`.
4. **Fix file upload** to use `FormData` with the raw `File`/`Blob`, not `.text()`.
5. **Add role-based guards** to `ProtectedRoute` once backend roles are stable.
6. **Consolidate toast state** either into `uiStore` or remove the orphaned `toasts` field.
7. **Add error boundaries** and a 404 route.
8. **Complete navigation parity** between routes, sidebar, and command palette.
