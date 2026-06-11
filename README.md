# Juvia

A free, open-source server control panel for Ubuntu/Debian. Built for non-technical small business owners. Inspired by Stripe Dashboard.

## Quick Start

```bash
# Install on Ubuntu/Debian
curl -fsSL https://get.juvia.dev | bash

# Or install from .deb package
apt install ./juvia_1.0.0_amd64.deb
```

After installation, access the panel at `http://your-server:8080` and complete the first-run setup wizard.

## Tech Stack

|Layer|Technology|
|-----|----------|
|Language|Go (single binary)|
|Frontend|React + Vite + Tailwind + shadcn/ui|
|Panel ↔ Agent|JSON-RPC over Unix socket|
|Database|SQLite (relational) + VictoriaMetrics (metrics)|
|Web Server|Nginx (default) + Apache (switchable)|
|SSL|certbot / Let's Encrypt|
|Spam Filter|Rspamd|

## Architecture

Two-process model:
- **Panel** — HTTP server (port 8080), runs as unprivileged user `juvia`
- **Agent** — Privileged operations, runs as root via systemd

Communication via JSON-RPC over Unix socket at `/var/run/juvia/agent.sock`.

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for full details.

## Development

Target server: `maruf@192.168.0.211` (SSH)

```bash
# Build
go build -o juvia ./cmd/panel
go build -o juvia-agent ./cmd/agent

# Run tests
go test ./...

# Lint
golangci-lint run
```

## Phases

|Phase|Focus|
|-----|-----|
|1|Foundation — architecture, auth, dashboard, metrics, alerts|
|2|Web Hosting — websites, SSL, DNS, config validation|
|3|Services — email, databases, file manager|
|4|Ops — firewall, backups, cron, logs|
|5|Polish — app installs, git deploy, webmail, terminal recording|

See [ROADMAP.md](docs/ROADMAP.md) for details.

## License

MIT — free, open-source, GPG-signed releases via GitHub.