# Build Command

## Description
Build the juvia binary (combined panel + agent).

## Usage
```bash
go build -o juvia ./cmd/...
```

## Build Targets

### Development (local)
```bash
go build -o juvia ./cmd/...
```

### Cross-compile for Linux AMD64
```bash
GOOS=linux GOARCH=amd64 go build -o juvia-linux-amd64 ./cmd/...
```

### With version info
```bash
ldflags="-X main.version=1.0.0 -X main.commit=$(git rev-parse HEAD)"
go build -ldflags "$ldflags" -o juvia ./cmd/...
```

## Output
- `juvia` — Single binary containing both panel and agent
- Runs as panel by default; use `juvia run` for panel, `juvia agent run` for agent

## Notes
- The agent is built into the same binary and runs as a separate process
- Systemd services determine which mode each binary runs in

## See Also
- [AGENTS.md](../AGENTS.md) — Developer context
- [DEPLOYMENT.md](../docs/DEPLOYMENT.md) — Installation and deployment