package runner

import "net/http"

type Phase string

const (
	PhaseAnytime    Phase = "anytime"
	PhaseAfterCreate Phase = "after-create"
	PhaseCleanup    Phase = "cleanup"
)

type Endpoint struct {
	Method         string
	Path           string
	Auth           AuthLevel
	Phase          Phase
	ResourceType   string
	Body           interface{}
	QueryParams    map[string]string
	AgentDependent bool
	SkipCleanup    bool
	Description    string
}

type AuthLevel int

const (
	AuthNone AuthLevel = iota
	AuthUser
	AuthAdmin
)

var Endpoints = []Endpoint{
	// === PHASE 0: Auth & Setup ===
	{Method: http.MethodGet, Path: "/api/v1/health", Auth: AuthNone, Phase: PhaseAnytime, Description: "Health check"},
	{Method: http.MethodGet, Path: "/api/v1/setup/status", Auth: AuthNone, Phase: PhaseAnytime, Description: "Check if setup is required"},
	{Method: http.MethodPost, Path: "/api/v1/setup/first-run", Auth: AuthNone, Phase: PhaseAnytime, Description: "First run setup"},
	{Method: http.MethodPost, Path: "/api/v1/auth/login", Auth: AuthNone, Phase: PhaseAnytime, Description: "Login"},
	{Method: http.MethodPost, Path: "/api/v1/auth/refresh", Auth: AuthNone, Phase: PhaseAnytime, Description: "Refresh token"},
	{Method: http.MethodPost, Path: "/api/v1/auth/logout", Auth: AuthUser, Phase: PhaseAnytime, Description: "Logout"},
	{Method: http.MethodGet, Path: "/api/v1/auth/me", Auth: AuthUser, Phase: PhaseAnytime, Description: "Get current user"},
	{Method: http.MethodPost, Path: "/api/v1/auth/2fa/enable", Auth: AuthUser, Phase: PhaseAnytime, Description: "Enable 2FA"},
	{Method: http.MethodPost, Path: "/api/v1/auth/2fa/verify", Auth: AuthUser, Phase: PhaseAnytime, Description: "Verify 2FA"},
	{Method: http.MethodDelete, Path: "/api/v1/auth/sessions/:id", Auth: AuthUser, Phase: PhaseAnytime, Description: "Revoke session"},

	// === PHASE 1: Create Resources ===
	{Method: http.MethodPost, Path: "/api/v1/websites", Auth: AuthUser, Phase: PhaseAnytime, ResourceType: "website", Description: "Create website"},
	{Method: http.MethodPost, Path: "/api/v1/databases", Auth: AuthUser, Phase: PhaseAnytime, ResourceType: "database", Description: "Create database"},
	{Method: http.MethodPost, Path: "/api/v1/email/mailboxes", Auth: AuthUser, Phase: PhaseAnytime, ResourceType: "mailbox", AgentDependent: true, Description: "Create mailbox"},
	{Method: http.MethodPost, Path: "/api/v1/email/aliases", Auth: AuthUser, Phase: PhaseAnytime, ResourceType: "alias", AgentDependent: true, Description: "Create alias"},
	{Method: http.MethodPost, Path: "/api/v1/email/forwarders", Auth: AuthUser, Phase: PhaseAnytime, ResourceType: "forwarder", AgentDependent: true, Description: "Create forwarder"},
	{Method: http.MethodPost, Path: "/api/v1/firewall/rules", Auth: AuthUser, Phase: PhaseAnytime, ResourceType: "firewall", AgentDependent: true, Description: "Create firewall rule"},
	{Method: http.MethodPost, Path: "/api/v1/cron", Auth: AuthUser, Phase: PhaseAnytime, ResourceType: "cron", AgentDependent: true, Description: "Create cron job"},
	{Method: http.MethodPost, Path: "/api/v1/backup-schedules", Auth: AuthUser, Phase: PhaseAnytime, ResourceType: "backup-schedule", Description: "Create backup schedule"},

	// === PHASE 2: Test with Real IDs ===
	{Method: http.MethodGet, Path: "/api/v1/users", Auth: AuthAdmin, Phase: PhaseAnytime, Description: "List users"},
	{Method: http.MethodPost, Path: "/api/v1/users", Auth: AuthAdmin, Phase: PhaseAnytime, Description: "Create user"},
	{Method: http.MethodGet, Path: "/api/v1/users/:id", Auth: AuthAdmin, Phase: PhaseAfterCreate, ResourceType: "user", Description: "Get user"},
	{Method: http.MethodPut, Path: "/api/v1/users/:id", Auth: AuthAdmin, Phase: PhaseAfterCreate, ResourceType: "user", Description: "Update user"},
	{Method: http.MethodDelete, Path: "/api/v1/users/:id", Auth: AuthAdmin, Phase: PhaseAfterCreate, ResourceType: "user", Description: "Delete user"},

	{Method: http.MethodGet, Path: "/api/v1/websites", Auth: AuthUser, Phase: PhaseAnytime, Description: "List websites"},
	{Method: http.MethodGet, Path: "/api/v1/websites/trash", Auth: AuthUser, Phase: PhaseAnytime, Description: "List trashed websites"},
	{Method: http.MethodDelete, Path: "/api/v1/websites/trash/purge", Auth: AuthUser, Phase: PhaseAnytime, Description: "Purge trash"},
	{Method: http.MethodGet, Path: "/api/v1/websites/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "Get website"},
	{Method: http.MethodPut, Path: "/api/v1/websites/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "Update website"},
	{Method: http.MethodDelete, Path: "/api/v1/websites/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "Delete website"},
	{Method: http.MethodPost, Path: "/api/v1/websites/:id/suspend", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, Description: "Suspend website"},
	{Method: http.MethodPost, Path: "/api/v1/websites/:id/restore", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, Description: "Restore website"},
	{Method: http.MethodDelete, Path: "/api/v1/websites/:id/permanent", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, Description: "Permanently delete website"},
	{Method: http.MethodPut, Path: "/api/v1/websites/:id/php", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "Change PHP version"},
	{Method: http.MethodGet, Path: "/api/v1/websites/:id/ssl/check", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, Description: "Check SSL"},
	{Method: http.MethodPost, Path: "/api/v1/websites/:id/ssl", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, SkipCleanup: true, Description: "Issue SSL"},
	{Method: http.MethodPost, Path: "/api/v1/websites/:id/ssl-renew", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, SkipCleanup: true, Description: "Renew SSL"},
	{Method: http.MethodDelete, Path: "/api/v1/websites/:id/ssl", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, SkipCleanup: true, Description: "Remove SSL"},
	{Method: http.MethodGet, Path: "/api/v1/websites/:id/dns/records", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "List DNS records"},
	{Method: http.MethodPost, Path: "/api/v1/websites/:id/dns/records", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "Create DNS record"},
	{Method: http.MethodPost, Path: "/api/v1/websites/:id/dns/suggest", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "Suggest DNS"},
	{Method: http.MethodGet, Path: "/api/v1/websites/:id/files", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "List files"},
	{Method: http.MethodPost, Path: "/api/v1/websites/:id/files/upload", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, Description: "Upload files"},
	{Method: http.MethodGet, Path: "/api/v1/websites/:id/files/download", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, Description: "Download files"},
	{Method: http.MethodPut, Path: "/api/v1/websites/:id/files/rename", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, Description: "Rename files"},
	{Method: http.MethodDelete, Path: "/api/v1/websites/:id/files/delete", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, Description: "Delete files"},
	{Method: http.MethodPost, Path: "/api/v1/websites/:id/files/extract", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, Description: "Extract files"},
	{Method: http.MethodGet, Path: "/api/v1/websites/:id/files/edit", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "Get file content"},
	{Method: http.MethodPut, Path: "/api/v1/websites/:id/files/edit", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, Description: "Save file content"},
	{Method: http.MethodGet, Path: "/api/v1/websites/:id/logs/access", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "Access logs"},
	{Method: http.MethodGet, Path: "/api/v1/websites/:id/logs/error", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "Error logs"},

	{Method: http.MethodPut, Path: "/api/v1/dns/records/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "dns-record", Description: "Update DNS record"},
	{Method: http.MethodDelete, Path: "/api/v1/dns/records/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "dns-record", Description: "Delete DNS record"},

	{Method: http.MethodGet, Path: "/api/v1/databases", Auth: AuthUser, Phase: PhaseAnytime, Description: "List databases"},
	{Method: http.MethodGet, Path: "/api/v1/databases/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "database", Description: "Get database"},
	{Method: http.MethodDelete, Path: "/api/v1/databases/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "database", AgentDependent: true, Description: "Delete database"},
	{Method: http.MethodPost, Path: "/api/v1/databases/:id/users", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "database", AgentDependent: true, Description: "Create DB user"},
	{Method: http.MethodDelete, Path: "/api/v1/databases/:id/users/:user_id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "database-user", AgentDependent: true, Description: "Delete DB user"},
	{Method: http.MethodPost, Path: "/api/v1/databases/:id/export", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "database", AgentDependent: true, Description: "Export database"},
	{Method: http.MethodGet, Path: "/api/v1/databases/:id/tables", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "database", Description: "List tables"},
	{Method: http.MethodGet, Path: "/api/v1/databases/:id/tables/:table/rows", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "database", Description: "Get table rows"},
	{Method: http.MethodPost, Path: "/api/v1/databases/:id/query", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "database", Description: "Query database"},

	{Method: http.MethodGet, Path: "/api/v1/email/mailboxes", Auth: AuthUser, Phase: PhaseAnytime, Description: "List mailboxes"},
	{Method: http.MethodGet, Path: "/api/v1/email/mailboxes/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "mailbox", Description: "Get mailbox"},
	{Method: http.MethodDelete, Path: "/api/v1/email/mailboxes/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "mailbox", AgentDependent: true, Description: "Delete mailbox"},
	{Method: http.MethodPut, Path: "/api/v1/email/mailboxes/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "mailbox", Description: "Update mailbox"},
	{Method: http.MethodGet, Path: "/api/v1/email/aliases", Auth: AuthUser, Phase: PhaseAnytime, Description: "List aliases"},
	{Method: http.MethodDelete, Path: "/api/v1/email/aliases/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "alias", AgentDependent: true, Description: "Delete alias"},
	{Method: http.MethodGet, Path: "/api/v1/email/forwarders", Auth: AuthUser, Phase: PhaseAnytime, Description: "List forwarders"},
	{Method: http.MethodDelete, Path: "/api/v1/email/forwarders/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "forwarder", AgentDependent: true, Description: "Delete forwarder"},
	{Method: http.MethodGet, Path: "/api/v1/email/catch-all", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", QueryParams: map[string]string{"domain": ""}, Description: "Get catch-all"},
	{Method: http.MethodPut, Path: "/api/v1/email/catch-all", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "Set catch-all"},
	{Method: http.MethodGet, Path: "/api/v1/email/deliverability", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", QueryParams: map[string]string{"domain": ""}, Description: "Check deliverability"},

	{Method: http.MethodGet, Path: "/api/v1/firewall/rules", Auth: AuthUser, Phase: PhaseAnytime, Description: "List firewall rules"},
	{Method: http.MethodGet, Path: "/api/v1/firewall/rules/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "firewall", Description: "Get firewall rule"},
	{Method: http.MethodPut, Path: "/api/v1/firewall/rules/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "firewall", AgentDependent: true, Description: "Update firewall rule"},
	{Method: http.MethodDelete, Path: "/api/v1/firewall/rules/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "firewall", AgentDependent: true, Description: "Delete firewall rule"},

	{Method: http.MethodGet, Path: "/api/v1/backups", Auth: AuthUser, Phase: PhaseAnytime, Description: "List backups"},
	{Method: http.MethodPost, Path: "/api/v1/backups", Auth: AuthUser, Phase: PhaseAnytime, ResourceType: "backup", AgentDependent: true, SkipCleanup: true, Description: "Create backup"},
	{Method: http.MethodGet, Path: "/api/v1/backups/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "backup", Description: "Get backup"},
	{Method: http.MethodPost, Path: "/api/v1/backups/:id/restore", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "backup", AgentDependent: true, SkipCleanup: true, Description: "Restore backup"},
	{Method: http.MethodDelete, Path: "/api/v1/backups/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "backup", AgentDependent: true, Description: "Delete backup"},

	{Method: http.MethodGet, Path: "/api/v1/backup-schedules", Auth: AuthUser, Phase: PhaseAnytime, Description: "List backup schedules"},
	{Method: http.MethodDelete, Path: "/api/v1/backup-schedules/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "backup-schedule", Description: "Delete backup schedule"},

	{Method: http.MethodGet, Path: "/api/v1/cron", Auth: AuthUser, Phase: PhaseAnytime, Description: "List cron jobs"},
	{Method: http.MethodGet, Path: "/api/v1/cron/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "cron", Description: "Get cron job"},
	{Method: http.MethodPut, Path: "/api/v1/cron/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "cron", AgentDependent: true, Description: "Update cron job"},
	{Method: http.MethodDelete, Path: "/api/v1/cron/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "cron", AgentDependent: true, Description: "Delete cron job"},
	{Method: http.MethodPost, Path: "/api/v1/cron/:id/enable", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "cron", AgentDependent: true, Description: "Enable cron job"},
	{Method: http.MethodPost, Path: "/api/v1/cron/:id/disable", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "cron", AgentDependent: true, Description: "Disable cron job"},
	{Method: http.MethodGet, Path: "/api/v1/cron/:id/logs", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "cron", Description: "Get cron job logs"},

	{Method: http.MethodGet, Path: "/api/v1/logs/system", Auth: AuthUser, Phase: PhaseAnytime, Description: "System logs"},

	{Method: http.MethodGet, Path: "/api/v1/alerts", Auth: AuthUser, Phase: PhaseAnytime, Description: "List alerts"},
	{Method: http.MethodPost, Path: "/api/v1/alerts/:id/acknowledge", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "alert", Description: "Acknowledge alert"},
	{Method: http.MethodDelete, Path: "/api/v1/alerts/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "alert", Description: "Delete alert"},
	{Method: http.MethodGet, Path: "/api/v1/alerts/settings", Auth: AuthUser, Phase: PhaseAnytime, Description: "Get alert settings"},
	{Method: http.MethodPut, Path: "/api/v1/alerts/settings", Auth: AuthUser, Phase: PhaseAnytime, Description: "Update alert settings"},

	{Method: http.MethodGet, Path: "/api/v1/metrics/current", Auth: AuthUser, Phase: PhaseAnytime, Description: "Current metrics"},
	{Method: http.MethodGet, Path: "/api/v1/metrics/history", Auth: AuthUser, Phase: PhaseAnytime, Description: "Historical metrics"},

	{Method: http.MethodGet, Path: "/api/v1/websites/:id/apps", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "List apps"},
	{Method: http.MethodPost, Path: "/api/v1/websites/:id/apps/install", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, SkipCleanup: true, Description: "Install app"},
	{Method: http.MethodPost, Path: "/api/v1/websites/:id/apps/:app_id/update", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, SkipCleanup: true, Description: "Update app"},

	{Method: http.MethodGet, Path: "/api/v1/websites/:id/git", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "Get git config"},
	{Method: http.MethodPost, Path: "/api/v1/websites/:id/git", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, Description: "Setup git"},
	{Method: http.MethodPost, Path: "/api/v1/websites/:id/git/pull", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", AgentDependent: true, Description: "Git pull"},
	{Method: http.MethodGet, Path: "/api/v1/websites/:id/git/webhook", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "website", Description: "Get webhook URL"},

	{Method: http.MethodGet, Path: "/api/v1/webmail/status", Auth: AuthUser, Phase: PhaseAnytime, Description: "Webmail status"},
	{Method: http.MethodPost, Path: "/api/v1/webmail/install", Auth: AuthUser, Phase: PhaseAnytime, AgentDependent: true, SkipCleanup: true, Description: "Install webmail"},
	{Method: http.MethodPost, Path: "/api/v1/webmail/uninstall", Auth: AuthUser, Phase: PhaseAnytime, AgentDependent: true, SkipCleanup: true, Description: "Uninstall webmail"},
	{Method: http.MethodGet, Path: "/api/v1/webmail/url", Auth: AuthUser, Phase: PhaseAnytime, Description: "Get webmail URL"},

	{Method: http.MethodPost, Path: "/api/v1/terminal/session", Auth: AuthUser, Phase: PhaseAnytime, Description: "Create terminal session"},
	{Method: http.MethodGet, Path: "/api/v1/terminal/sessions", Auth: AuthUser, Phase: PhaseAnytime, Description: "List terminal sessions"},
	{Method: http.MethodDelete, Path: "/api/v1/terminal/sessions/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "terminal-session", Description: "Close terminal session"},
	{Method: http.MethodGet, Path: "/api/v1/terminal/recordings", Auth: AuthUser, Phase: PhaseAnytime, Description: "List terminal recordings"},
	{Method: http.MethodGet, Path: "/api/v1/terminal/recordings/:id", Auth: AuthUser, Phase: PhaseAfterCreate, ResourceType: "terminal-recording", Description: "Get terminal recording"},

	{Method: http.MethodGet, Path: "/api/v1/updates/check", Auth: AuthUser, Phase: PhaseAnytime, AgentDependent: true, SkipCleanup: true, Description: "Check for updates"},
	{Method: http.MethodPost, Path: "/api/v1/updates/download", Auth: AuthUser, Phase: PhaseAnytime, AgentDependent: true, SkipCleanup: true, Description: "Download update"},
	{Method: http.MethodPost, Path: "/api/v1/updates/apply", Auth: AuthUser, Phase: PhaseAnytime, AgentDependent: true, SkipCleanup: true, Description: "Apply update"},
	{Method: http.MethodPost, Path: "/api/v1/updates/rollback", Auth: AuthUser, Phase: PhaseAnytime, AgentDependent: true, SkipCleanup: true, Description: "Rollback update"},

	{Method: http.MethodGet, Path: "/api/v1/settings", Auth: AuthUser, Phase: PhaseAnytime, Description: "Get settings"},
	{Method: http.MethodPut, Path: "/api/v1/settings", Auth: AuthUser, Phase: PhaseAnytime, Description: "Update settings"},
	{Method: http.MethodGet, Path: "/api/v1/settings/export", Auth: AuthUser, Phase: PhaseAnytime, Description: "Export config"},
	{Method: http.MethodPost, Path: "/api/v1/settings/import", Auth: AuthUser, Phase: PhaseAnytime, Description: "Import config"},

	{Method: http.MethodGet, Path: "/api/v1/audit-log", Auth: AuthUser, Phase: PhaseAnytime, Description: "Get audit log"},

	{Method: http.MethodGet, Path: "/api/v1/services", Auth: AuthUser, Phase: PhaseAnytime, Description: "List services"},
	{Method: http.MethodGet, Path: "/api/v1/services/:name", Auth: AuthUser, Phase: PhaseAnytime, Description: "Get service"},
	{Method: http.MethodPost, Path: "/api/v1/services/:name/restart", Auth: AuthUser, Phase: PhaseAnytime, AgentDependent: true, SkipCleanup: true, Description: "Restart service"},

	{Method: http.MethodGet, Path: "/api/v1/version", Auth: AuthNone, Phase: PhaseAnytime, Description: "Get version"},
}

var CreatePayloads = map[string]interface{}{
	"/api/v1/websites": map[string]interface{}{
		"domain":      "test-example.com",
		"php_version": "8.2",
		"web_server":  "nginx",
	},
	"/api/v1/databases": map[string]interface{}{
		"name":   "testdb",
		"engine": "mysql",
	},
	"/api/v1/databases/:id/users": map[string]interface{}{
		"username": "testuser",
		"password": "TestPass123!",
		"host":     "localhost",
	},
	"/api/v1/databases/:id/query": map[string]interface{}{
		"query": "SELECT 1",
	},
	"/api/v1/email/mailboxes": map[string]interface{}{
		"email":       "test@example.com",
		"password":    "TestPass123!",
		"quota":       10737418240,
		"display_name": "Test User",
	},
	"/api/v1/email/mailboxes/:id": map[string]interface{}{
		"display_name": "Updated User",
		"forward_to":   "",
		"quota":        10737418240,
	},
	"/api/v1/email/aliases": map[string]interface{}{
		"domain":      "example.com",
		"source":      "alias@example.com",
		"destination": "test@example.com",
	},
	"/api/v1/email/forwarders": map[string]interface{}{
		"domain":      "example.com",
		"source":      "forward@example.com",
		"destination": "test@example.com",
	},
	"/api/v1/email/catch-all": map[string]interface{}{
		"domain":    "example.com",
		"forward_to": "test@example.com",
	},
	"/api/v1/firewall/rules": map[string]interface{}{
		"name":        "test-rule",
		"description": "Test firewall rule",
		"action":      "allow",
		"port":        "8080/tcp",
		"protocol":    "tcp",
		"source":      "any",
	},
	"/api/v1/firewall/rules/:id": map[string]interface{}{
		"name":        "updated-rule",
		"description": "Updated rule",
		"action":      "allow",
		"port":        "9090/tcp",
		"protocol":    "tcp",
		"source":      "any",
	},
	"/api/v1/backups": map[string]interface{}{
		"type":    "full",
		"storage": "local",
	},
	"/api/v1/backup-schedules": map[string]interface{}{
		"schedule":       "0 2 * * *",
		"retention_days":  7,
		"storage":         "local",
	},
	"/api/v1/cron": map[string]interface{}{
		"schedule": "0 * * * *",
		"command":  "echo test",
		"run_as":   "root",
		"type":     "shell",
	},
	"/api/v1/cron/:id": map[string]interface{}{
		"schedule": "0 0 * * *",
		"command":  "echo updated",
		"run_as":   "root",
	},
	"/api/v1/alerts/settings": map[string]interface{}{
		"email_enabled": true,
		"email":         "admin@example.com",
	},
	"/api/v1/settings": map[string]interface{}{
		"server_name": "test-panel",
	},
	"/api/v1/settings/import": map[string]interface{}{
		"version": "1.0.0",
		"config":  "{}",
	},
	"/api/v1/users": map[string]interface{}{
		"username": "newuser",
		"password": "NewPass123!",
		"role":     "user",
		"email":    "newuser@example.com",
	},
	"/api/v1/users/:id": map[string]interface{}{
		"role":  "user",
		"email": "updated@example.com",
	},
	"/api/v1/websites/:id/php": map[string]interface{}{
		"php_version": "8.1",
	},
	"/api/v1/websites/:id/dns/records": map[string]interface{}{
		"type":  "A",
		"name":  "@",
		"value": "192.168.0.1",
	},
	"/api/v1/websites/:id/dns/suggest": map[string]interface{}{
		"subdomain": "www",
	},
	"/api/v1/websites/:id/files/rename": map[string]interface{}{
		"old_path": "/old.txt",
		"new_path": "/new.txt",
	},
	"/api/v1/websites/:id/files/extract": map[string]interface{}{
		"path":         "/archive.tar.gz",
		"destination": "/",
	},
	"/api/v1/websites/:id/apps/install": map[string]interface{}{
		"app": "wordpress",
	},
	"/api/v1/websites/:id/git": map[string]interface{}{
		"repository": "https://github.com/example/repo.git",
		"branch":     "main",
	},
	"/api/v1/updates/download": map[string]interface{}{
		"url": "https://example.com/update.tar.gz",
	},
	"/api/v1/updates/apply": map[string]interface{}{
		"path": "/tmp/update.tar.gz",
	},
}
