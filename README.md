# Juvia — Server Control Panel

A modern, security-first server control panel for managing websites, email, databases, DNS, firewall, backups, and more.

## Features

- **Websites** — Create and manage PHP/Node.js/Python websites with automatic Nginx/Apache configuration
- **SSL Certificates** — Free Let's Encrypt certificates with automatic renewal via Certbot
- **DNS Management** — Full DNS zone management with BIND9 integration
- **Email Hosting** — Mailboxes, aliases, forwarders, and SPF/DKIM/DMARC with Postfix and Dovecot
- **Databases** — MySQL and PostgreSQL database management
- **File Manager** — Browser-based file manager with code editor
- **Firewall** — UFW-based firewall with country blocking
- **Backups** — Automated backups to local storage
- **Cron Jobs** — Visual cron job editor with execution history
- **Logs** — Real-time log viewing with search and filtering
- **One-Click Apps** — Install WordPress, Laravel, Next.js, and more with one click
- **Git Deploy** — Automatic deployments from GitHub repositories
- **Terminal Recording** — Record and playback SSH-like terminal sessions
- **Two-Factor Authentication** — TOTP-based 2FA with recovery codes

## Architecture

Juvia uses a two-process model for security:

1. **Panel process** (`juvia`) — Runs as unprivileged user `juvia`, handles HTTP API
2. **Agent process** (`juvia-agent`) — Runs as root, handles privileged operations

Communication is via JSON-RPC over a Unix socket (`/var/run/juvia/agent.sock`).

## Requirements

- Ubuntu 22.04+ or Debian-based Linux
- Nginx (for web serving)
- SQLite3 (for database)
- 512MB RAM minimum
- 10GB disk space

## Installation

```bash
curl -sSL https://raw.githubusercontent.com/marufnwu/juvia/main/scripts/install.sh | bash
```

This will download and install the latest .deb package from GitHub releases.

After installation, access the control panel at `http://your-server-ip:8080`.

## Manual Installation

```bash
# Download the .deb package
wget https://github.com/marufnwu/juvia/releases/download/v0.1.0/juvia_0.1.0_amd64.deb

# Install it
sudo dpkg -i juvia_0.1.0_amd64.deb
sudo apt-get install -f  # Install dependencies if needed
```

## Building from Source

### Prerequisites

- Go 1.22+
- Node.js 20+
- npm

### Build

```bash
# Clone the repository
git clone https://github.com/marufnwu/juvia.git
cd juvia

# Run the build script
./scripts/build.sh

# This produces:
#   juvia        - Panel binary
#   juvia-agent  - Agent binary
```

### Build .deb Package

```bash
./scripts/build-deb.sh 0.1.0 amd64
```

This produces `juvia_0.1.0_amd64.deb`.

## Updating

```bash
# Using the update script
sudo ./scripts/update.sh

# Or manually download and install
wget https://github.com/marufnwu/juvia/releases/download/vX.Y.Z/juvia_X.Y.Z_amd64.deb
sudo dpkg -i juvia_X.Y.Z_amd64.deb
sudo systemctl restart juvia juvia-agent
```

## Uninstalling

```bash
sudo ./scripts/uninstall.sh
```

Type `yes` when prompted to confirm removal.

## Configuration

Configuration files are stored in `/etc/juvia/`:

- `nginx/sites/` — Website Nginx configurations
- `php-fpm/pools/` — PHP-FPM pool configurations
- `postfix/` — Postfix configuration
- `dovecot/` — Dovecot configuration
- `bind/zones/` — DNS zone files
- `ufw/rules` — Firewall rules

Data is stored in:

- `/var/lib/juvia/` — SQLite database and website data
- `/var/log/juvia/` — Log files
- `/var/run/juvia/` — Runtime files (including agent socket)

## Security

- Panel runs as unprivileged user `juvia`
- Agent runs as root for privileged operations
- JWT-based authentication with refresh tokens
- TOTP 2FA support with recovery codes
- Session recording for terminal sessions
- Atomic config file updates (write to `.tmp`, validate, rename)
- Path traversal prevention via `safePath()` validation

## API

The REST API is available at `http://localhost:8080/api/v1/`. All endpoints require JWT authentication except:

- `POST /api/v1/auth/login` — Login
- `POST /api/v1/auth/refresh` — Refresh token
- `GET /api/v1/setup/status` — Check if setup is required
- `POST /api/v1/setup` — Initial setup wizard

WebSocket is available at `ws://localhost:8080/ws` for:

- Real-time metrics streaming
- Task progress updates
- Terminal session proxy

Agent methods are available via JSON-RPC over Unix socket at `/var/run/juvia/agent.sock`.

## Development

```bash
# Build for development
go build -o juvia ./cmd/panel
go build -o juvia-agent ./cmd/agent

# Run tests
go test ./...

# Lint
golangci-lint run
```

## License

MIT License

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.