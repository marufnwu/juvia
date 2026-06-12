package db

import "time"

// User represents a panel user.
type User struct {
	ID                 int64      `json:"id" db:"id"`
	Username           string     `json:"username" db:"username"`
	PasswordHash       string     `json:"-" db:"password_hash"`
	Role               string     `json:"role" db:"role"`
	Email              *string    `json:"email,omitempty" db:"email"`
	Active             bool       `json:"active" db:"active"`
	TwoFAEnabled       bool       `json:"2fa_enabled" db:"totp_enabled"`
	TwoFASecret        *string    `json:"-" db:"totp_secret"`
	TwoFARecoveryCodes *string    `json:"-" db:"totp_recovery_codes"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	LastLogin          *time.Time `json:"last_login,omitempty" db:"last_login"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// UserSite maps users to websites for per-site roles.
type UserSite struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	WebsiteID int64     `json:"website_id" db:"website_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Session represents a user refresh token session.
type Session struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	TokenHash string    `json:"-" db:"token_hash"`
	IPAddress *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string   `json:"user_agent,omitempty" db:"user_agent"`
	Revoked   bool      `json:"revoked" db:"revoked"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Website represents a hosted website.
type Website struct {
	ID           int64      `json:"id" db:"id"`
	Domain       string     `json:"domain" db:"domain"`
	DocumentRoot *string    `json:"document_root,omitempty" db:"document_root"`
	PHPVersion   *string    `json:"php_version,omitempty" db:"php_version"`
	WebServer    string     `json:"web_server" db:"web_server"`
	SSLEnabled   bool       `json:"ssl_enabled" db:"ssl_enabled"`
	SSLExpiry    *time.Time `json:"ssl_expiry,omitempty" db:"ssl_expiry"`
	Status       string     `json:"status" db:"status"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	UserID       *int64     `json:"user_id,omitempty" db:"user_id"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	GitConfig    string     `json:"git_config,omitempty" db:"git_config"`
}

// Domain represents an additional domain for a website.
type Domain struct {
	ID        int64     `json:"id" db:"id"`
	WebsiteID int64     `json:"website_id" db:"website_id"`
	Domain    string    `json:"domain" db:"domain"`
	Type      string    `json:"type" db:"type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Zone represents a DNS zone.
type Zone struct {
	ID        int64     `json:"id" db:"id"`
	WebsiteID *int64    `json:"website_id,omitempty" db:"website_id"`
	Domain    string    `json:"domain" db:"domain"`
	Serial    int64     `json:"serial" db:"serial"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DNSRecord represents a single DNS record.
type DNSRecord struct {
	ID        int64     `json:"id" db:"id"`
	ZoneID    int64     `json:"zone_id" db:"zone_id"`
	Type      string    `json:"type" db:"type"`
	Name      string    `json:"name" db:"name"`
	Value     string    `json:"value" db:"value"`
	Priority  *int      `json:"priority,omitempty" db:"priority"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Database represents a database.
type Database struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Engine    string    `json:"engine" db:"engine"`
	UserID    *int64    `json:"user_id,omitempty" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// DBUser represents a database user.
type DBUser struct {
	ID           int64     `json:"id" db:"id"`
	DatabaseID   int64     `json:"database_id" db:"database_id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Host         *string   `json:"host,omitempty" db:"host"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Mailbox represents an email mailbox.
type Mailbox struct {
	ID                    int64      `json:"id" db:"id"`
	Email                 string     `json:"email" db:"email"`
	PasswordHash          string     `json:"-" db:"password_hash"`
	Quota                 int64      `json:"quota" db:"quota"`
	DisplayName           *string    `json:"display_name,omitempty" db:"display_name"`
	ForwardTo             *string    `json:"forward_to,omitempty" db:"forward_to"`
	Status                string     `json:"status" db:"status"`
	AutoresponderEnabled  bool       `json:"autoresponder_enabled" db:"autoresponder_enabled"`
	AutoresponderMessage  *string    `json:"autoresponder_message,omitempty" db:"autoresponder_message"`
	AutoresponderStartDate *time.Time `json:"autoresponder_start_date,omitempty" db:"autoresponder_start_date"`
	AutoresponderEndDate   *time.Time `json:"autoresponder_end_date,omitempty" db:"autoresponder_end_date"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
}

// EmailAlias represents an email alias.
type EmailAlias struct {
	ID          int64     `json:"id" db:"id"`
	Domain      string    `json:"domain" db:"domain"`
	Source      string    `json:"source" db:"source"`
	Destination string    `json:"destination" db:"destination"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// EmailForwarder represents an email forwarder.
type EmailForwarder struct {
	ID          int64     `json:"id" db:"id"`
	Domain      string    `json:"domain" db:"domain"`
	Source      string    `json:"source" db:"source"`
	Destination string    `json:"destination" db:"destination"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// EmailCatchAll represents a catch-all configuration.
type EmailCatchAll struct {
	ID        int64     `json:"id" db:"id"`
	Domain    string    `json:"domain" db:"domain"`
	ForwardTo string    `json:"forward_to" db:"forward_to"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Task represents an asynchronous job.
type Task struct {
	ID          int64      `json:"id" db:"id"`
	TaskID      string     `json:"task_id" db:"task_id"`
	Type        string     `json:"type" db:"type"`
	Status      string     `json:"status" db:"status"`
	Progress    int        `json:"progress" db:"progress"`
	Steps       *string    `json:"steps,omitempty" db:"steps"`
	Result      *string    `json:"result,omitempty" db:"result"`
	CreatedBy   int64      `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// AuditLog represents an audit log entry.
type AuditLog struct {
	ID        int64     `json:"id" db:"id"`
	UserID    *int64    `json:"user_id,omitempty" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Details   string    `json:"details" db:"details"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// FirewallRule represents a firewall rule.
type FirewallRule struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	Action      string    `json:"action" db:"action"`
	Port        string    `json:"port" db:"port"`
	Protocol    string    `json:"protocol" db:"protocol"`
	Source      string    `json:"source" db:"source"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// CronJob represents a cron job.
type CronJob struct {
	ID         int64      `json:"id" db:"id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	Schedule   string     `json:"schedule" db:"schedule"`
	Command    string     `json:"command" db:"command"`
	RunAs      string     `json:"run_as" db:"run_as"`
	Type       string     `json:"type" db:"type"`
	Enabled    bool       `json:"enabled" db:"enabled"`
	LastRun    *time.Time `json:"last_run,omitempty" db:"last_run"`
	LastStatus *string    `json:"last_status,omitempty" db:"last_status"`
	LastOutput *string    `json:"last_output,omitempty" db:"last_output"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// CronJobLog represents a cron job execution log.
type CronJobLog struct {
	ID         int64     `json:"id" db:"id"`
	CronJobID  int64     `json:"cron_job_id" db:"cron_job_id"`
	RunAt      time.Time `json:"run_at" db:"run_at"`
	DurationMs int64     `json:"duration_ms" db:"duration_ms"`
	Status     string    `json:"status" db:"status"`
	Output     string    `json:"output" db:"output"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Backup represents a backup.
type Backup struct {
	ID          int64      `json:"id" db:"id"`
	WebsiteID   *int64     `json:"website_id,omitempty" db:"website_id"`
	Type        string     `json:"type" db:"type"`
	Status      string     `json:"status" db:"status"`
	Storage     string     `json:"storage" db:"storage"`
	Path        string     `json:"path" db:"path"`
	SizeBytes   int64       `json:"size_bytes" db:"size_bytes"`
	Checksum    *string     `json:"checksum,omitempty" db:"checksum"`
	Verified    bool        `json:"verified" db:"verified"`
	VerifiedAt  *time.Time `json:"verified_at,omitempty" db:"verified_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

// BackupSchedule represents a backup schedule.
type BackupSchedule struct {
	ID            int64     `json:"id" db:"id"`
	WebsiteID     *int64    `json:"website_id,omitempty" db:"website_id"`
	Schedule      string    `json:"schedule" db:"schedule"`
	RetentionDays int       `json:"retention_days" db:"retention_days"`
	Storage       string    `json:"storage" db:"storage"`
	Enabled       bool      `json:"enabled" db:"enabled"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Alert represents a system alert.
type Alert struct {
	ID           int64     `json:"id" db:"id"`
	Type         string    `json:"type" db:"type"`
	Severity     string    `json:"severity" db:"severity"`
	Message      string    `json:"message" db:"message"`
	ResourceID   *int64    `json:"resource_id,omitempty" db:"resource_id"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	Acknowledged bool      `json:"acknowledged" db:"acknowledged"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Service represents a system service.
type Service struct {
	ID         int64     `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Version    *string   `json:"version,omitempty" db:"version"`
	Installed  bool      `json:"installed" db:"installed"`
	Running    bool      `json:"running" db:"running"`
	Health     string    `json:"health" db:"health"`
	LastError  *string   `json:"last_error,omitempty" db:"last_error"`
	CheckedAt  time.Time `json:"checked_at" db:"checked_at"`
}

// Setting represents a key-value setting.
type Setting struct {
	ID        int64     `json:"id" db:"id"`
	Key       string    `json:"key" db:"key"`
	Value     *string   `json:"value,omitempty" db:"value"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// App represents a one-click installed application.
type App struct {
	ID              int64     `json:"id" db:"id"`
	AppType         string    `json:"app_type" db:"app_type"`
	Name            string    `json:"name" db:"name"`
	WebsiteID       int64     `json:"website_id" db:"website_id"`
	Version         *string   `json:"version,omitempty" db:"version"`
	InstalledAt     time.Time `json:"installed_at" db:"installed_at"`
	UpdateAvailable bool      `json:"update_available" db:"update_available"`
}
