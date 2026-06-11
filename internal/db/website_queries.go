package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (db *DB) CreateWebsite(ctx context.Context, domain, documentRoot, phpVersion, webServer string, userID int64) (*Website, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO websites (domain, document_root, php_version, web_server, status, user_id) VALUES (?, ?, ?, ?, 'active', ?)`,
		domain, documentRoot, phpVersion, webServer, userID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, fmt.Errorf("domain already exists")
		}
		return nil, fmt.Errorf("create website: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetWebsiteByID(ctx, id)
}

func (db *DB) GetWebsiteByID(ctx context.Context, id int64) (*Website, error) {
	var w Website
	row := db.QueryRowContext(ctx,
		`SELECT id, domain, document_root, php_version, web_server, ssl_enabled, ssl_expiry, status, deleted_at, user_id, created_at, updated_at, COALESCE(git_config, '')
		FROM websites WHERE id = ?`, id)
	if err := row.Scan(&w.ID, &w.Domain, &w.DocumentRoot, &w.PHPVersion, &w.WebServer, &w.SSLEnabled, &w.SSLExpiry, &w.Status, &w.DeletedAt, &w.UserID, &w.CreatedAt, &w.UpdatedAt, &w.GitConfig); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("website not found")
		}
		return nil, fmt.Errorf("get website: %w", err)
	}
	return &w, nil
}

func (db *DB) GetWebsiteByDomain(ctx context.Context, domain string) (*Website, error) {
	var w Website
	row := db.QueryRowContext(ctx,
		`SELECT id, domain, document_root, php_version, web_server, ssl_enabled, ssl_expiry, status, deleted_at, user_id, created_at, updated_at, COALESCE(git_config, '')
		FROM websites WHERE domain = ?`, domain)
	if err := row.Scan(&w.ID, &w.Domain, &w.DocumentRoot, &w.PHPVersion, &w.WebServer, &w.SSLEnabled, &w.SSLExpiry, &w.Status, &w.DeletedAt, &w.UserID, &w.CreatedAt, &w.UpdatedAt, &w.GitConfig); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("website not found")
		}
		return nil, fmt.Errorf("get website by domain: %w", err)
	}
	return&w, nil
}

type ListWebsitesResult struct {
	Websites []Website
	Total    int
}

func (db *DB) ListWebsites(ctx context.Context, status string, search string, sort string, order string, page int, limit int) (*ListWebsitesResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	validSorts := map[string]bool{"domain": true, "created_at": true, "status": true}
	if !validSorts[sort] {
		sort = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	whereClause := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	if status != "" {
		whereClause += " AND status = ?"
		args = append(args, status)
	}
	if search != "" {
		whereClause += " AND domain LIKE ?"
		args = append(args, "%"+search+"%")
	}

	var total int
	countRow := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM websites "+whereClause, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, fmt.Errorf("count websites: %w", err)
	}

	query := fmt.Sprintf(`SELECT id, domain, document_root, php_version, web_server, ssl_enabled, ssl_expiry, status, deleted_at, user_id, created_at, updated_at, COALESCE(git_config, '')
		FROM websites %s ORDER BY %s %s LIMIT ? OFFSET ?`, whereClause, sort, order)
	args = append(args, limit, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list websites: %w", err)
	}
	defer rows.Close()

	var websites []Website
	for rows.Next() {
		var w Website
		if err := rows.Scan(&w.ID, &w.Domain, &w.DocumentRoot, &w.PHPVersion, &w.WebServer, &w.SSLEnabled, &w.SSLExpiry, &w.Status, &w.DeletedAt, &w.UserID, &w.CreatedAt, &w.UpdatedAt, &w.GitConfig); err != nil {
			continue
		}
		websites = append(websites, w)
	}
	return &ListWebsitesResult{Websites: websites, Total: total}, nil
}

func (db *DB) UpdateWebsite(ctx context.Context, id int64, phpVersion, webServer string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE websites SET php_version = ?, web_server = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		phpVersion, webServer, id)
	if err != nil {
		return fmt.Errorf("update website: %w", err)
	}
	return nil
}

func (db *DB) SoftDeleteWebsite(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx,
		`UPDATE websites SET status = 'deleted', deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id)
	if err != nil {
		return fmt.Errorf("soft delete website: %w", err)
	}
	return nil
}

func (db *DB) RestoreWebsite(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx,
		`UPDATE websites SET status = 'active', deleted_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id)
	if err != nil {
		return fmt.Errorf("restore website: %w", err)
	}
	return nil
}

func (db *DB) PermanentlyDeleteWebsite(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM websites WHERE id = ? AND status = 'deleted'`,
		id)
	if err != nil {
		return fmt.Errorf("permanently delete website: %w", err)
	}
	return nil
}

func (db *DB) SuspendWebsite(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx,
		`UPDATE websites SET status = 'suspended', updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id)
	if err != nil {
		return fmt.Errorf("suspend website: %w", err)
	}
	return nil
}

func (db *DB) ListTrashedWebsites(ctx context.Context) ([]Website, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, domain, document_root, php_version, web_server, ssl_enabled, ssl_expiry, status, deleted_at, user_id, created_at, updated_at
		FROM websites WHERE deleted_at IS NOT NULL ORDER BY deleted_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list trashed websites: %w", err)
	}
	defer rows.Close()

	var websites []Website
	for rows.Next() {
		var w Website
		if err := rows.Scan(&w.ID, &w.Domain, &w.DocumentRoot, &w.PHPVersion, &w.WebServer, &w.SSLEnabled, &w.SSLExpiry, &w.Status, &w.DeletedAt, &w.UserID, &w.CreatedAt, &w.UpdatedAt); err != nil {
			continue
		}
		websites = append(websites, w)
	}
	return websites, nil
}

func (db *DB) PurgeTrash(ctx context.Context) error {
	_, err := db.ExecContext(ctx, `DELETE FROM websites WHERE deleted_at IS NOT NULL AND deleted_at < datetime('now', '-30 days')`)
	if err != nil {
		return fmt.Errorf("purge trash: %w", err)
	}
	return nil
}

func (db *DB) CreateDomain(ctx context.Context, websiteID int64, domain, domainType string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO domains (website_id, domain, type) VALUES (?, ?, ?)`,
		websiteID, domain, domainType)
	if err != nil {
		return fmt.Errorf("create domain: %w", err)
	}
	return nil
}

func (db *DB) ListDomainsByWebsite(ctx context.Context, websiteID int64) ([]Domain, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, website_id, domain, type, created_at FROM domains WHERE website_id = ?`, websiteID)
	if err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}
	defer rows.Close()

	var domains []Domain
	for rows.Next() {
		var d Domain
		if err := rows.Scan(&d.ID, &d.WebsiteID, &d.Domain, &d.Type, &d.CreatedAt); err != nil {
			continue
		}
		domains = append(domains, d)
	}
	return domains, nil
}

func (db *DB) UpdateWebsiteSSL(ctx context.Context, id int64, enabled bool, expiry *time.Time) error {
	_, err := db.ExecContext(ctx,
		`UPDATE websites SET ssl_enabled = ?, ssl_expiry = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		enabled, expiry, id)
	if err != nil {
		return fmt.Errorf("update website ssl: %w", err)
	}
	return nil
}

func (db *DB) CreateZone(ctx context.Context, websiteID int64, domain string) (*Zone, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO zones (website_id, domain, serial) VALUES (?, ?, 1)`,
		websiteID, domain)
	if err != nil {
		return nil, fmt.Errorf("create zone: %w", err)
	}
	id, _ := result.LastInsertId()
	var z Zone
	row := db.QueryRowContext(ctx, `SELECT id, website_id, domain, serial, created_at, updated_at FROM zones WHERE id = ?`, id)
	if err := row.Scan(&z.ID, &z.WebsiteID, &z.Domain, &z.Serial, &z.CreatedAt, &z.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get zone: %w", err)
	}
	return&z, nil
}

func (db *DB) GetZoneByWebsite(ctx context.Context, websiteID int64) (*Zone, error) {
	var z Zone
	row := db.QueryRowContext(ctx, `SELECT id, website_id, domain, serial, created_at, updated_at FROM zones WHERE website_id = ?`, websiteID)
	if err := row.Scan(&z.ID, &z.WebsiteID, &z.Domain, &z.Serial, &z.CreatedAt, &z.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get zone: %w", err)
	}
	return &z, nil
}

func (db *DB) CreateDNSRecord(ctx context.Context, zoneID int64, recType, name, value string, priority *int) (*DNSRecord, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO dns_records (zone_id, type, name, value, priority) VALUES (?, ?, ?, ?, ?)`,
		zoneID, recType, name, value, priority)
	if err != nil {
		return nil, fmt.Errorf("create dns record: %w", err)
	}
	id, _ := result.LastInsertId()
	var r DNSRecord
	row := db.QueryRowContext(ctx, `SELECT id, zone_id, type, name, value, priority, created_at, updated_at FROM dns_records WHERE id = ?`, id)
	if err := row.Scan(&r.ID, &r.ZoneID, &r.Type, &r.Name, &r.Value, &r.Priority, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get dns record: %w", err)
	}
	return&r, nil
}

func (db *DB) ListDNSRecordsByZone(ctx context.Context, zoneID int64) ([]DNSRecord, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, zone_id, type, name, value, priority, created_at, updated_at FROM dns_records WHERE zone_id = ?`, zoneID)
	if err != nil {
		return nil, fmt.Errorf("list dns records: %w", err)
	}
	defer rows.Close()

	var records []DNSRecord
	for rows.Next() {
		var r DNSRecord
		if err := rows.Scan(&r.ID, &r.ZoneID, &r.Type, &r.Name, &r.Value, &r.Priority, &r.CreatedAt, &r.UpdatedAt); err != nil {
			continue
		}
		records = append(records, r)
	}
	return records, nil
}

func (db *DB) GetDNSRecordByID(ctx context.Context, id int64) (*DNSRecord, error) {
	var r DNSRecord
	row := db.QueryRowContext(ctx, `SELECT id, zone_id, type, name, value, priority, created_at, updated_at FROM dns_records WHERE id = ?`, id)
	if err := row.Scan(&r.ID, &r.ZoneID, &r.Type, &r.Name, &r.Value, &r.Priority, &r.CreatedAt, &r.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("dns record not found")
		}
		return nil, fmt.Errorf("get dns record: %w", err)
	}
	return &r, nil
}

func (db *DB) UpdateDNSRecord(ctx context.Context, id int64, name, value string, priority *int) error {
	_, err := db.ExecContext(ctx,
		`UPDATE dns_records SET name = ?, value = ?, priority = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		name, value, priority, id)
	if err != nil {
		return fmt.Errorf("update dns record: %w", err)
	}
	return nil
}

func (db *DB) DeleteDNSRecord(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM dns_records WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete dns record: %w", err)
	}
	return nil
}

func (db *DB) LogAudit(ctx context.Context, userID *int64, action, ipAddress, userAgent, details string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO audit_log (user_id, action, ip_address, user_agent, details) VALUES (?, ?, ?, ?, ?)`,
		userID, action, ipAddress, userAgent, details)
	if err != nil {
		return fmt.Errorf("log audit: %w", err)
	}
	return nil
}

func (db *DB) UpdateWebsiteGitConfig(ctx context.Context, websiteID int64, gitConfig string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE websites SET git_config = ? WHERE id = ?`,
		gitConfig, websiteID)
	if err != nil {
		return fmt.Errorf("update website git config: %w", err)
	}
	return nil
}

func (db *DB) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	row := db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key)
	if err := row.Scan(&value); err != nil {
		return "", fmt.Errorf("get setting: %w", err)
	}
	return value, nil
}

func (db *DB) SetSetting(ctx context.Context, key, value string) error {
	_, err := db.ExecContext(ctx,
		`INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)`,
		key, value)
	if err != nil {
		return fmt.Errorf("set setting: %w", err)
	}
	return nil
}

type AuditLogEntry struct {
	ID        int64     `json:"id" db:"id"`
	UserID    *int64    `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Details   string    `json:"details" db:"details"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (db *DB) ListAuditLog(ctx context.Context, page, limit int) ([]AuditLogEntry, error) {
	offset := (page - 1) * limit
	rows, err := db.QueryContext(ctx,
		`SELECT id, user_id, action, ip_address, user_agent, details, created_at
		FROM audit_log ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list audit log: %w", err)
	}
	defer rows.Close()

	var entries []AuditLogEntry
	for rows.Next() {
		var e AuditLogEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.Action, &e.IPAddress, &e.UserAgent, &e.Details, &e.CreatedAt); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}
