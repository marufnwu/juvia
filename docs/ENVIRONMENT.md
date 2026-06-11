# Environment Setup

## Target Environment

|Component|Value|
|---------|-----|
|Remote Server|`maruf@192.168.0.211`|
|SSH Port|22 (default)|
|OS|Ubuntu 22.04+ or Debian 11+|
|Panel Port|8080|
|Agent Socket|`/var/run/juvia/agent.sock`|

## Local Development Setup

### Prerequisites

- **Go 1.21+** — For building the backend
- **Node.js 20+** — For frontend development
- **Git** — Version control
- **SSH access** to remote server

### Initial Setup

1. **Clone the repository**
```bash
git clone <repo-url>
cd juvia
```

2. **Build the binary**
```bash
go build -o juvia ./cmd/...
```

3. **Copy to remote server**
```bash
scp juvia maruf@192.168.0.211:/tmp/
```

4. **Install on remote**
```bash
ssh maruf@192.168.0.211 "sudo dpkg -i /tmp/juvia"
# or for development, run directly:
ssh maruf@192.168.0.211 "sudo /tmp/juvia run"
```

### Frontend Development

For frontend development, you can run the React app separately:

```bash
cd web
npm install
npm run dev
```

This runs the frontend on `http://localhost:5173` and proxies API requests to the panel at port 8080.

**Note:** The Go binary embeds the frontend via `go:embed`. For development, the frontend runs separately and connects to the backend API. In production, everything is in one binary.

### Remote Server Access

```bash
# SSH to remote server
ssh maruf@192.168.0.211

# Check panel status
ssh maruf@192.168.0.211 "systemctl status juvia"

# View panel logs
ssh maruf@192.168.0.211 "tail -f /var/log/juvia/panel.log"

# Restart services
ssh maruf@192.168.0.211 "sudo systemctl restart juvia"
ssh maruf@192.168.0.211 "sudo systemctl restart juvia-agent"
```

### File Editing Workflow

1. Edit files locally
2. Build binary
3. Copy to remote: `scp juvia maruf@192.168.0.211:/tmp/`
4. Install: `ssh maruf@192.168.0.211 "sudo dpkg -i /tmp/juvia"`
5. Or for quick iteration: `scp juvia maruf@192.168.0.211:/usr/local/bin/ && ssh maruf@192.168.0.211 "sudo systemctl restart juvia"`

### Environment Variables

For testing specific features:

```bash
# On remote server, create env file
ssh maruf@192.168.0.211 "sudo nano /etc/juvia/env"
```

Common environment variables:
|Variable|Default|Description|
|--------|-------|-----------|
|`JUVIA_PORT`|8080|HTTP port for panel|
|`JUVIA_DB`|/var/lib/juvia/juvia.db|SQLite database path|
|`JUVIA_LOG`|/var/log/juvia/panel.log|Log file path|
|`JUVIA_SOCKET`|/var/run/juvia/agent.sock|Unix socket path|

### Database Access

```bash
# Connect to SQLite on remote
ssh maruf@192.168.0.211 "sqlite3 /var/lib/juvia/juvia.db"

# View tables
sqlite3> .tables

# View schema
sqlite3> .schema users
```

### Service Management

```bash
# Check both services
ssh maruf@192.168.0.211 "systemctl status juvia juvia-agent"

# View logs
ssh maruf@192.168.0.211 "journalctl -u juvia -f"
ssh maruf@192.168.0.211 "journalctl -u juvia-agent -f"

# Restart both
ssh maruf@192.168.0.211 "sudo systemctl restart juvia juvia-agent"
```

## See Also

- [AGENTS.md](../AGENTS.md) — Build commands and developer context
- [DEPLOYMENT.md](../DEPLOYMENT.md) — Installation and service setup
- [ROADMAP.md](../ROADMAP.md) — Implementation phases
