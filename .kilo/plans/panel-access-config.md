# Plan: Panel Access Configuration (Bind Address, Port, Security)

## Problem Statement

The panel currently:
1. Binds to **all interfaces** (`:`) — accessible via localhost, LAN IP, and public IP
2. Uses fixed port **8080** — a common, easily scanned port
3. Has a `panel_port` setting in DB but **does not restart the server** when changed
4. Has no `panel_bind_address` setting

This means if a user accesses the panel via public IP, it's exposed on a predictable port with no additional hardening.

## Proposed Changes

### Option A: Uncommon Port by Default (Recommended for IP Access)

Change the default port from `8080` to an uncommon port (e.g., `18473`) when the panel is accessed by raw IP. This is "security through obscurity" but practically reduces automated scanning exposure.

### Option B: Configurable Bind Address + Port

Add `panel_bind_address` and `panel_port` as first-class config options:

| Setting | Default | Description |
|---------|---------|-------------|
| `panel_bind_address` | `0.0.0.0` | `127.0.0.1` = localhost only, `0.0.0.0` = all interfaces |
| `panel_port` | `8080` | Panel HTTP port |
| `panel_hostname` | `""` | If set, redirect IP access to hostname |

### Option C: Access Mode (Simpler UX)

Instead of raw bind address, offer an access mode:

| Mode | Bind Address | Use Case |
|------|-------------|----------|
| `localhost` | `127.0.0.1` | SSH tunnel / reverse proxy only |
| `private` | `0.0.0.0` + firewall whitelist | LAN access |
| `public` | `0.0.0.0` | Direct public access (uncommon port recommended) |

## Implementation Plan

### Step 1: Read Settings at Startup
Modify `cmd/panel/main.go` to read `panel_bind_address` and `panel_port` from the database **before** starting the server. Fall back to env var `JUVIA_PORT` and `0.0.0.0`.

### Step 2: Settings API — Return Restart Warning
When `panel_port` or `panel_bind_address` is updated via `PUT /api/v1/settings`, return a warning:  
`"Panel port changed. Restart the juvia service to apply: sudo systemctl restart juvia"`

### Step 3: Default to Uncommon Port
If `panel_port` is not set and no `panel_hostname` is configured, default to an uncommon port (e.g., `18473`) instead of `8080`. This protects users who access via raw IP without a domain.

### Step 4: systemd Service Update
Update `scripts/juvia.service` to remove the hardcoded `JUVIA_PORT=8080` environment variable. The panel should read from its own settings DB.

### Step 5: Install Script Update
Update `scripts/install.sh` to:
- Detect if running on a public IP
- Suggest an uncommon port if no domain is configured
- Open the chosen port in UFW

## Files to Modify

| File | Change |
|------|--------|
| `cmd/panel/main.go` | Read bind address + port from settings DB at startup |
| `internal/api/settings.go` | Add `panel_bind_address` to settings keys; return restart warning on port change |
| `scripts/juvia.service` | Remove hardcoded `JUVIA_PORT=8080` |
| `scripts/install.sh` | Suggest uncommon port; update UFW rule |
| `README.md` | Document access modes and port configuration |

## Security Notes

- Binding to `127.0.0.1` is the safest option if using an SSH tunnel or reverse proxy (nginx)
- If binding to `0.0.0.0` on a public IP, always use HTTPS (future: auto-SSL for panel itself)
- Uncommon ports don't stop targeted attacks but reduce automated scanner noise

## Decision Needed

**Should the default be:**
1. Port `8080` (common, predictable) — current behavior
2. Uncommon port (e.g., random 5-digit) when accessed by IP
3. `localhost` only by default, require explicit opt-in for public access
