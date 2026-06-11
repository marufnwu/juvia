package db

import (
	"context"
	"fmt"
)

func (db *DB) CreateBackup(ctx context.Context, websiteID *int64, backupType, storage, path string) (*Backup, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO backups (website_id, type, status, storage, path) VALUES (?, ?, 'in_progress', ?, ?)`,
		websiteID, backupType, storage, path)
	if err != nil {
		return nil, fmt.Errorf("create backup: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetBackupByID(ctx, id)
}

func (db *DB) GetBackupByID(ctx context.Context, id int64) (*Backup, error) {
	var b Backup
	row := db.QueryRowContext(ctx,
		`SELECT id, website_id, type, status, storage, path, size_bytes, checksum, verified, verified_at, created_at
		FROM backups WHERE id = ?`, id)
	if err := row.Scan(&b.ID, &b.WebsiteID, &b.Type, &b.Status, &b.Storage, &b.Path, &b.SizeBytes, &b.Checksum, &b.Verified, &b.VerifiedAt, &b.CreatedAt); err != nil {
		return nil, fmt.Errorf("get backup: %w", err)
	}
	return &b, nil
}

type ListBackupsResult struct {
	Backups []Backup
	Total   int
}

func (db *DB) ListBackups(ctx context.Context, websiteID *int64, page, limit int) (*ListBackupsResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	whereClause := "1=1"
	args := []interface{}{}
	if websiteID != nil {
		whereClause += " AND website_id = ?"
		args = append(args, *websiteID)
	}

	var total int
	countRow := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM backups WHERE "+whereClause, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, fmt.Errorf("count backups: %w", err)
	}

	query := fmt.Sprintf(`SELECT id, website_id, type, status, storage, path, size_bytes, checksum, verified, verified_at, created_at
		FROM backups WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereClause)
	args = append(args, limit, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list backups: %w", err)
	}
	defer rows.Close()

	var backups []Backup
	for rows.Next() {
		var b Backup
		if err := rows.Scan(&b.ID, &b.WebsiteID, &b.Type, &b.Status, &b.Storage, &b.Path, &b.SizeBytes, &b.Checksum, &b.Verified, &b.VerifiedAt, &b.CreatedAt); err != nil {
			continue
		}
		backups = append(backups, b)
	}
	return &ListBackupsResult{Backups: backups, Total: total}, nil
}

func (db *DB) UpdateBackupStatus(ctx context.Context, id int64, status string, sizeBytes int64, checksum string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE backups SET status = ?, size_bytes = ?, checksum = ? WHERE id = ?`,
		status, sizeBytes, checksum, id)
	if err != nil {
		return fmt.Errorf("update backup status: %w", err)
	}
	return nil
}

func (db *DB) MarkBackupVerified(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx,
		`UPDATE backups SET verified = TRUE, verified_at = CURRENT_TIMESTAMP WHERE id = ?`,
		id)
	if err != nil {
		return fmt.Errorf("mark backup verified: %w", err)
	}
	return nil
}

func (db *DB) DeleteBackup(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM backups WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete backup: %w", err)
	}
	return nil
}

func (db *DB) CreateBackupSchedule(ctx context.Context, websiteID *int64, schedule string, retentionDays int, storage string) (*BackupSchedule, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO backup_schedules (website_id, schedule, retention_days, storage) VALUES (?, ?, ?, ?)`,
		websiteID, schedule, retentionDays, storage)
	if err != nil {
		return nil, fmt.Errorf("create backup schedule: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetBackupScheduleByID(ctx, id)
}

func (db *DB) GetBackupScheduleByID(ctx context.Context, id int64) (*BackupSchedule, error) {
	var s BackupSchedule
	row := db.QueryRowContext(ctx,
		`SELECT id, website_id, schedule, retention_days, storage, enabled, created_at
		FROM backup_schedules WHERE id = ?`, id)
	if err := row.Scan(&s.ID, &s.WebsiteID, &s.Schedule, &s.RetentionDays, &s.Storage, &s.Enabled, &s.CreatedAt); err != nil {
		return nil, fmt.Errorf("get backup schedule: %w", err)
	}
	return &s, nil
}

func (db *DB) ListBackupSchedules(ctx context.Context, websiteID *int64) ([]BackupSchedule, error) {
	query := `SELECT id, website_id, schedule, retention_days, storage, enabled, created_at
		FROM backup_schedules WHERE 1=1`
	var args []interface{}
	if websiteID != nil {
		query += " AND website_id = ?"
		args = append(args, *websiteID)
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list backup schedules: %w", err)
	}
	defer rows.Close()

	var schedules []BackupSchedule
	for rows.Next() {
		var s BackupSchedule
		if err := rows.Scan(&s.ID, &s.WebsiteID, &s.Schedule, &s.RetentionDays, &s.Storage, &s.Enabled, &s.CreatedAt); err != nil {
			continue
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}

func (db *DB) DeleteBackupSchedule(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM backup_schedules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete backup schedule: %w", err)
	}
	return nil
}

func (db *DB) GetLastBackupForWebsite(ctx context.Context, websiteID int64) (*Backup, error) {
	var b Backup
	row := db.QueryRowContext(ctx,
		`SELECT id, website_id, type, status, storage, path, size_bytes, checksum, verified, verified_at, created_at
		FROM backups WHERE website_id = ? AND status = 'completed' ORDER BY created_at DESC LIMIT 1`, websiteID)
	if err := row.Scan(&b.ID, &b.WebsiteID, &b.Type, &b.Status, &b.Storage, &b.Path, &b.SizeBytes, &b.Checksum, &b.Verified, &b.VerifiedAt, &b.CreatedAt); err != nil {
		return nil, err
	}
	return &b, nil
}