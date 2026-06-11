# Deploy Command

## Description
Deploy juvia to the remote server at `maruf@192.168.0.211`.

## Prerequisites
- SSH access to `maruf@192.168.0.211`
- Sudo privileges on remote server

## Usage

### 1. Build the binary
```bash
go build -o juvia ./cmd/...
```

### 2. Copy to remote server
```bash
scp juvia maruf@192.168.0.211:/tmp/juvia
```

### 3. Install on remote
```bash
ssh maruf@192.168.0.211 "sudo dpkg -i /tmp/juvia"
```

### 4. Restart services
```bash
ssh maruf@192.168.0.211 "sudo systemctl restart juvia juvia-agent"
```

## Quick Deploy (Single Command)

```bash
go build -o juvia ./cmd/... && \
scp juvia maruf@192.168.0.211:/tmp/ && \
ssh maruf@192.168.0.211 "sudo dpkg -i /tmp/juvia && sudo systemctl restart juvia juvia-agent"
```

## Rollback

If something goes wrong:

```bash
# View recent logs
ssh maruf@192.168.0.211 "sudo journalctl -u juvia -n 50"

# Check service status
ssh maruf@192.168.0.211 "sudo systemctl status juvia"

# If needed, reinstall previous version
# (you would need to have kept the previous binary)
ssh maruf@192.168.0.211 "sudo dpkg -i /path/to/previous/juvia.deb"
```

## Direct Restart (No Reinstall)

For quick iteration without reinstalling:

```bash
scp juvia maruf@192.168.0.211:/usr/local/bin/ && \
ssh maruf@192.168.0.211 "sudo systemctl restart juvia juvia-agent"
```

## Verify Deployment

```bash
# Check panel is responding
curl http://192.168.0.211:8080/api/health

# Check agent is running
ssh maruf@192.168.0.211 "sudo systemctl status juvia-agent"
```

## See Also
- [ENVIRONMENT.md](../docs/ENVIRONMENT.md) — Environment setup
- [DEPLOYMENT.md](../docs/DEPLOYMENT.md) — Installation guide
- [AGENTS.md](../AGENTS.md) — Developer context