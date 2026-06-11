# juvia — Developer Context

This file provides critical context for AI agents and developers working on juvia.

**IMPORTANT:** The complete implementation roadmap is in `.kilo/rules/juvia-implementation-roadmap.md`.
AI agents MUST follow the phases in order. Do not skip ahead. Each phase's output is the next phase's input.
See the roadmap for: step-by-step tasks, file paths, constraints, success criteria, and implementation rules.

**CRITICAL RULE: REAL CODE ONLY**
- This is a production server control panel. Implement actual working code.
- NO mock data, NO placeholder functions, NO TODO-only implementations.
- Every function must have real logic. Every API endpoint must work.
- If a feature cannot be fully implemented, flag it for clarification — do NOT create stubs.
- Every commit must pass `go build ./...` and `go test ./...` before deployment.

## Build Commands

```bash
# Build both binaries
go build -o juvia ./cmd/panel
go build -o juvia-agent ./cmd/agent

# Build combined single binary (panel + agent)
go build -o juvia ./cmd/...

# Run tests
go test ./...

# Lint
golangci-lint run

# Build frontend (requires Node.js)
cd web && npm install && npm run build && cd ..

# Full build script (Linux/macOS)
./scripts/build.sh

# Full build script (Windows)
powershell -File scripts/build.ps1

# Build .deb package (on Linux)
./scripts/build-deb.sh 0.1.0 amd64

# Install package
dpkg -i juvia_0.1.0_amd64.deb

# Uninstall package
./scripts/uninstall.sh

# Update package
./scripts/update.sh
```

## Two-Process Model Rules

1. **Panel process** runs as unprivileged user `juvia` — NEVER run as root
2. **Agent process** runs as root — handles all privileged operations
3. Communication is JSON-RPC over Unix socket (`/var/run/juvia/agent.sock`)
4. Socket permissions: `660`, owner `root`, group `juvia`
5. Never expose the agent socket to the network

## Security Rules

- **Path traversal prevention**: All file paths must be validated with `safePath()` before operations
- **Atomic config swaps**: Never edit configs directly — write to `.tmp`, validate, then atomic rename
- **Input validation**: Never trust user input — validate everything server-side
- **No direct root access**: Panel process must never run with root privileges
- **Session recording**: All terminal sessions must be recorded to `/var/log/juvia/terminal/`
- **2FA enforcement**: Admin accounts should be encouraged to enable 2FA

## Plain-English UI Rule

Every technical term in the UI must include a plain-English explanation:

```
PHP Version · The programming language your website's code runs on
DKIM Record · A security signature that proves your emails are genuine
Cron Job    · A task that runs automatically on a schedule
```

Never show technical step names like "Creating PHP-FPM pool" — always show "Setting up PHP for your website..."

## Code Style

### Go
- Standard Go formatting (`gofmt`)
- Error wrapping with `fmt.Errorf("context: %w", err)`
- Context propagation for all operations
- Structured logging with `slog`

### React/TypeScript
- Functional components with hooks
- Tailwind for styling (no CSS files)
- shadcn/ui components (customized to monochrome system)
- Zod for validation
- React Query for server state

## Project Structure

```
juvia/
├── cmd/
│   ├── panel/          # HTTP panel entry point
│   └── agent/          # Privileged agent entry point
├── internal/
│   ├── agent/          # Agent implementations (nginx, ssl, email, etc.)
│   ├── api/            # REST API handlers
│   ├── ws/             # WebSocket handlers
│   ├── tasks/          # Async task runner
│   ├── socket/         # Unix socket client/server
│   ├── db/             # SQLite models + migrations
│   ├── auth/           # JWT + 2FA
│   └── ...
├── web/                # React frontend (embedded via Go embed)
└── scripts/
    ├── install.sh
    └── build-deb.sh
```

## Config Directory Structure

All generated configs live under `/etc/juvia/`:
- `nginx/sites/` — one file per website
- `php-fpm/pools/` — one pool config per website
- `postfix/`, `dovecot/`, `bind/zones/`, `ufw/`

Configs are never edited directly — always use atomic swap:
1. Write to `filename.tmp`
2. Validate (`nginx -t`, `named-checkzone`, etc.)
3. If valid: `rename()` to final name (atomic on Linux)
4. If invalid: delete `.tmp` and return error

## Testing Strategy

- Unit tests for all business logic
- Integration tests for API endpoints
- Manual testing on target server (Ubuntu 22.04+)
- Test on both Nginx and Apache configurations

## Common Operations

```bash
# Check panel status
systemctl status juvia

# Check agent status
systemctl status juvia-agent

# View logs
tail -f /var/log/juvia/panel.log
tail -f /var/log/juvia/agent.log

# Restart services
systemctl restart juvia
systemctl restart juvia-agent

# Database location
/var/lib/juvia/juvia.db
```