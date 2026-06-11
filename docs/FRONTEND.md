# Frontend Architecture

## Stack

- **React 18** with TypeScript
- **Vite** for build tooling
- **Tailwind CSS** for styling
- **shadcn/ui** for base components (heavily customized)
- **Zustand** for client state
- **React Query** for server state
- **xterm.js** for terminal
- **Monaco Editor** for file editor
- **Recharts** for metrics charts (drop-in replacement for Go Highcharts)

## Design System

### Typography
- Font: Geist (by Vercel)
- Code: Geist Mono
- Page titles: 18px, weight 600, #111111
- Section labels: 13px, weight 500, #555555, uppercase
- Body text: 14px, weight 400, #333333
- Secondary text: 13px, weight 400, #888888

### Color Palette
```
Background:       #FFFFFF
Surface:          #FAFAFA
Border:           #E5E5E5
Text primary:     #111111
Text secondary:   #888888

Status Live:      #16A34A (green)
Status Error:     #DC2626 (red)
Status Warning:   #D97706 (amber)
Status Pending:   #2563EB (blue)

Accent:           #111111 (buttons, active nav)
Hover:            #F5F5F5
```

### Corner Radius
**0px everywhere** — fully square, no rounded corners.

### Component Overrides (shadcn/ui)
- Buttons: Square, black primary, ghost secondary
- Inputs: Square, 1px #E5E5E5 border, #111 focus ring
- Modals: Square, 480px max, 40% black backdrop
- Toasts: Square, bottom-right stack
- Tooltips: Square, black bg, white text, 12px

## Project Structure

```
web/
├── src/
│   ├── components/
│   │   ├── ui/              # shadcn/ui components (modified)
│   │   ├── layout/          # Sidebar, TopBar, PageLayout
│   │   ├── forms/           # Form components per module
│   │   └── shared/          # StatusBadge, InfoTooltip, etc.
│   ├── pages/
│   │   ├── Dashboard/
│   │   ├── Websites/
│   │   ├── Email/
│   │   ├── Databases/
│   │   ├── Files/
│   │   ├── Firewall/
│   │   ├── Backups/
│   │   ├── Cron/
│   │   ├── Metrics/
│   │   ├── Terminal/
│   │   └── Settings/
│   ├── hooks/               # Custom React hooks
│   ├── lib/
│   │   ├── api.ts           # API client
│   │   ├── ws.ts            # WebSocket client
│   │   └── utils.ts         # Helpers
│   ├── stores/              # Zustand stores
│   ├── types/               # TypeScript types
│   └── App.tsx
├── public/
└── index.html
```

## Routing

React Router v6 with nested routes.

```
/                     → Redirect to /dashboard
/dashboard            → Overview dashboard
/websites             → Website list
/websites/:id         → Website detail
/email                → Email management
/email/mailboxes      → Mailbox list
/databases            → Database list
/databases/:id         → Database detail (with visual manager)
/files                → File manager
/firewall              → Firewall rules
/backups               → Backup management
/cron                  → Cron job manager
/metrics               → Server metrics
/terminal              → Web terminal
/settings              → Settings pages
/settings/users        → User management
/settings/security     → Security settings
/settings/updates      → Update management
/settings/backup       → Backup configuration
```

## State Management

### Zustand (client state)
- Auth state (user, token, 2FA status)
- UI state (sidebar collapsed, active modal)
- Command palette open/closed

### React Query (server state)
- Website list, database list, etc.
- Automatic refetching
- Optimistic updates for mutations
- Stale-while-revalidate caching

## Error Handling

### Global Error Boundary
```tsx
class ErrorBoundary extends Component {
  componentDidCatch(error, errorInfo) {
    Sentry.captureException(error, { extra: errorInfo });
    this.setState({ hasError: true });
  }
}
```

### Retry Logic
- API calls retry with exponential backoff (1s, 2s, 4s)
- Max 3 retries before showing error
- WebSocket reconnects automatically

### Offline Mode
- Service Worker caches dashboard data
- Show "Offline" indicator when disconnected
- Queue mutations for when back online

## WebSocket Integration

```typescript
// Connect to WebSocket
const ws = useWebSocket('/ws/v1/metrics')

// Subscribe to task progress
ws.subscribe('/tasks/:taskId', (data) => {
  updateTaskProgress(data)
})

// Auto-reconnection with exponential backoff
ws.onClose(() => {
  setTimeout(() => ws.connect(), Math.min(1000 * 2 ** attempts, 30000))
})
```

## Terminal Integration

Using xterm.js with WebSocket backend:
- Full PTY emulation
- Multiple tabs (independent sessions)
- Session recording to asciinema format
- Resize handling
- Search in terminal

## File Manager

Using Monaco Editor for in-browser file editing:
- Syntax highlighting for common web formats
- Search and replace
- Drag-and-drop upload with progress
- Chunked upload for large files

## Accessibility (a11y)

- WCAG 2.1 AA compliance
- Keyboard navigation for all interactive elements
- Focus management for modals
- Screen reader support (ARIA labels)
- Skip to main content link

## Performance Budget

|Metric|Target|
|------|------|
|Initial bundle size|< 200KB gzipped|
|Time to Interactive|< 3s|
|First Contentful Paint|< 1.5s|
|Largest Contentful Paint|< 2.5s|

### Code Splitting
- Route-based lazy loading
- Heavy components (Monaco, xterm) loaded on demand
- Dynamic imports for modals

## PWA / Offline Support

Service Worker caches:
- App shell (HTML, CSS, JS)
- API responses for dashboard
- Static assets

Offline indicator shown when disconnected.

## Analytics / Error Tracking

- Sentry for error tracking (opt-in)
- Basic usage analytics (page views, feature usage)
- No personal data collected

## Localization (i18n)

Using react-i18next:

```
web/src/locales/
├── en/
│   └── translation.json
├── es/
│   └── translation.json
└── fr/
    └── translation.json
```

Translation key naming: `module.section.key`

## API Client

```typescript
// Typed API calls with response envelope
const websites = await api.get<Website[]>('/websites')

// Pagination
const websites = await api.get<Website[]>('/websites', {
  params: { page: 1, limit: 50 }
})

// Optimistic update
const updateWebsite = useMutation({
  mutationFn: (data) => api.put(`/websites/${data.id}`, data),
  onMutate: async (newData) => {
    await queryClient.cancelQueries(['websites'])
    const old = queryClient.getQueryData(['websites'])
    queryClient.setQueryData(['websites'], (old) => [...old, newData])
    return { old }
  },
  onError: (err, newData, context) => {
    queryClient.setQueryData(['websites'], context.old)
  }
})
```

JWT token automatically attached to all requests. Refresh token in httpOnly cookie.

## Responsive Breakpoints

|Breakpoint|Behavior|
|----------|--------|
|≥1024px|Full sidebar (240px)|
|768-1023px|Collapsed sidebar (64px)|
|<768px|Hamburger menu, full-screen modals|

## Loading States

Skeleton placeholders with animated shimmer. No spinners except inline on clicked buttons.

## Motion

Minimal — instant navigation. Only allowed:
- Sidebar collapse: 150ms
- Modal fade-in: 100ms
- Toast slide-up: 150ms
- Skeleton shimmer
