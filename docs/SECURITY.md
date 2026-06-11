# Security Model

## Two-Process Privilege Separation

**Critical**: The panel process runs as unprivileged user `juvia`. Only the agent process runs as root.

This prevents an RCE vulnerability in the HTTP layer (file uploads, file manager, user content) from giving attackers full root access.

```
Panel (user: juvia) → handles HTTP, UI, API
Agent (user: root)       → handles all system operations
```

## Unix Socket Security

Socket: `/var/run/juvia/agent.sock`

Permissions: `660` (read/write for owner and group only)
Owner: `root:juvia`

Only the panel process can communicate with the agent.

## Authentication

### JWT + Refresh Tokens
- Access token: 15 minutes lifetime, stored in memory
- Refresh token: httpOnly cookie, 7 days, stored server-side in SQLite

### Password Storage
bcrypt hashing with cost factor 12.

### Password Complexity Requirements
- Minimum 12 characters
- At least 1 uppercase letter
- At least 1 lowercase letter
- At least 1 number
- At least 1 special character

### 2FA (TOTP)
- Google Authenticator, Authy, 1Password support
- Recovery codes generated on setup (10 codes, single-use)

### Login Security
- Rate limiting: 5 failed attempts per IP per 10 minutes
- Auto-block: After 20 failed attempts, IP added to UFW deny list for 24 hours
- Login alerts: Notify user on login from new IP
- Session list: Users can see active sessions and revoke them
- Account lockout: Admin notified of lockout events

### Session Management
- Sessions are revocable (revoked column in sessions table)
- All sessions listed in user profile with IP, user agent, last active
- One-click revoke individual sessions

## CSRF Protection

All mutating endpoints (POST, PUT, DELETE) require:
- `X-CSRF-Token` header with valid token
- Token obtained from `GET /api/v1/auth/me` response header `X-CSRF-Token`

## Path Traversal Prevention

All file operations use `safePath()`:
```go
func safePath(base, userInput string) (string, error) {
    resolved := filepath.Clean(filepath.Join(base, userInput))
    if !strings.HasPrefix(resolved, base) {
        return "", ErrPathTraversal
    }
    return resolved, nil
}
```

This blocks attacks like `../../etc/passwd`.

### Error Definition

```go
// ErrPathTraversal is returned when a path traversal attack is detected
var ErrPathTraversal = errors.New("path traversal attempt detected")
```

## Config Management: Atomic Swaps

Never edit configs directly. Always:
1. Write to `filename.tmp`
2. Validate (`nginx -t`, `named-checkzone`, etc.)
3. If valid: atomic `rename()` to final name
4. If invalid: delete `.tmp`, return error

Bad configs never replace working ones.

## Upload Security

### ZIP Extraction Limits
- Max extracted size: 1GB
- Max entries: 10,000
- No absolute paths in archive
- All extracted paths validated with `safePath()`

### Upload Limits
- Configurable max file size (default 512MB)
- Uploads go to staging first, then moved to final location

### Executable Protection
Uploaded files in `public_html` cannot be executed unless the site is explicitly a PHP/Node/Python site.

## Webhook Security

Git deploy webhooks use HMAC-SHA256 signature verification:

1. Webhook URL includes secret: `https://example.com/webhook/abc123?secret=xyz`
2. GitHub sends `X-Hub-Signature-256` header with HMAC
3. Panel verifies: `HMAC-SHA256(secret, payload) == signature`

## Terminal Security

### Session Recording
All terminal sessions recorded to `/var/log/juvia/terminal/` in asciinema-compatible format (`YYYY-MM-DD_HH-MM-SS_user_sessionid.cast`).

### Idle Timeout
Auto-close after 15 minutes of inactivity (configurable in Settings).

### Re-authentication
Opening a root terminal requires re-entering the panel password.

### Root Terminal Banner
Persistent banner: "Root Terminal — all actions are recorded"

### IP Restriction
Root terminal can be restricted to specific IPs in Settings.

## Backup Integrity

After every backup:
1. SHA-256 checksum computed and stored
2. Archive entry count verified
3. "Last verified" timestamp shown

Periodic dry-run restore test extracts to temp directory, verifies all files present, deletes temp. Result shown in UI.

### Backup Encryption
Backups are encrypted with AES-256 before transfer to remote storage.

## Self-Update Security

1. Background goroutine checks GitHub Releases API daily
2. New version found → banner shown
3. User clicks Update → binary downloaded
4. **GPG signature verified** before applying
5. If signature valid → replace binary, restart services
6. If signature invalid → abort, alert user, log attempt

### Rollback on Failure
- Previous binary kept as `.backup`
- Automatic rollback if new binary fails to start
- Email notification of rollback

## Database Encryption at Rest

SQLite database contains sensitive data. Enable encryption with SQLCipher:
```bash
# Environment variable
SQLITE_KEY=your-encryption-key
```

Alternatively, store database on an encrypted filesystem (LUKS).

## Log Sanitization

All logs must redact sensitive data before writing:
- Passwords (replaced with `***`)
- Tokens (replaced with `***`)
- API keys (replaced with `***`)
- Personal information (emails, IPs hashed for privacy)

## Agent Privilege Restrictions

The agent runs as root but should use Linux capabilities where possible:
- `CAP_NET_BIND_SERVICE` for binding privileged ports
- `CAP_DAC_OVERRIDE` for file access checks
- Avoid full root where a capability suffices

## Content Security Policy

Panel UI served with strict CSP headers:
```
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self' wss://*:8080; img-src 'self' data:; font-src 'self'
```

## Service Health Monitoring

Service registry checks health every 60 seconds. Issues surfaced via alert system.

## Audit Logging

Every action logged with:
- User ID
- Action description
- IP address
- Timestamp
- User agent

Viewable in Settings → Audit Log.

## Security Checklist

- [ ] Panel process never runs as root
- [ ] Socket permissions are 660, owner root:juvia
- [ ] All paths validated with safePath() before operations
- [ ] All config changes use atomic swap pattern
- [ ] Upload limits enforced (size + entry count)
- [ ] ZIP extraction validated for path traversal
- [ ] Terminal sessions recorded
- [ ] Idle timeout on terminal sessions
- [ ] Re-auth for root terminal
- [ ] 2FA available and encouraged
- [ ] Rate limiting on login
- [ ] Auto-block after 20 failed login attempts
- [ ] Backup integrity checksums computed
- [ ] Backup encryption before remote transfer
- [ ] GPG verification on self-update
- [ ] CSP headers on panel UI
- [ ] Audit log for all actions
- [ ] CSRF protection on mutating endpoints
- [ ] Webhook HMAC verification
- [ ] Password complexity enforced
- [ ] Database encryption at rest
- [ ] Log sanitization

## See Also

- [AGENTS.md](../AGENTS.md) — Security rules summary
- [ARCHITECTURE.md](../docs/ARCHITECTURE.md) — Two-process model details
- [DEPLOYMENT.md](../docs/DEPLOYMENT.md) — Socket permissions and directory structure
