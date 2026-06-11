# Changelog

All notable changes to Juvia will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.1.0] — Initial Release

### Added

#### Phase 1 — Foundation
- Go-based HTTP server with embedded React frontend
- SQLite database with migrations
- JSON-RPC over Unix socket (panel ↔ agent communication)
- JWT authentication with refresh tokens
- TOTP 2FA with recovery codes
- Session management
- Rate limiting middleware
- CSRF protection
- WebSocket server for real-time features
- Background task runner

#### Phase 2 — Web Hosting
- Website CRUD (create, suspend, restore, trash)
- Nginx configuration generation
- PHP-FPM pool management
- Let's Encrypt SSL via Certbot
- Automatic SSL renewal
- BIND9 DNS zone management
- DNS record CRUD
- File manager API (list, upload, download, rename, delete, extract)

#### Phase 3 — Services
- Email mailbox management (Postfix/Dovecot)
- Email aliases and forwarders
- SPF, DKIM, DMARC record generation
- MySQL database management
- PostgreSQL database management
- Database export/import
- Visual file manager with Monaco editor
- Visual database manager with SQL editor
- Email dashboard

#### Phase 4 — Operations
- UFW firewall management
- Country-based blocking
- Backup system (create, restore, verify)
- Backup schedules
- Cron job CRUD
- Cron execution history
- Log viewer with live tail
- Log search and filtering
- Alert engine (website down/slow, SSL expiry, disk, services)
- Metrics collection (CPU, RAM, disk, network, processes)

#### Phase 5 — Polish
- One-click app installer (WordPress, Laravel, Next.js)
- Git-based deployment
- GitHub webhook receiver with HMAC verification
- Webmail (Roundcube) integration
- Terminal session recording
- Self-updater with GPG signature verification
- .deb packaging
- Command palette (Cmd+K search)
- Settings pages (General, Email, Security, Updates, Users, Backups, Notifications, Audit Log)

### Changed

- Improved error handling across all API endpoints
- Better input validation with Zod schemas
- Optimized database queries with proper indexes

### Security

- All file operations use `safePath()` validation
- Config files use atomic swap (write → validate → rename)
- Path traversal prevention
- Session recording for audit trail
- 2FA enforcement for admin accounts