package db

import (
	"context"
	"fmt"
	"time"
)

func (db *DB) CreateAlert(ctx context.Context, alertType, severity, message string, resourceID *int64, resourceType string) (*Alert, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO alerts (type, severity, message, resource_id, resource_type) VALUES (?, ?, ?, ?, ?)`,
		alertType, severity, message, resourceID, resourceType)
	if err != nil {
		return nil, fmt.Errorf("create alert: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetAlertByID(ctx, id)
}

func (db *DB) GetAlertByID(ctx context.Context, id int64) (*Alert, error) {
	var a Alert
	row := db.QueryRowContext(ctx,
		`SELECT id, type, severity, message, resource_id, resource_type, acknowledged, created_at
		FROM alerts WHERE id = ?`, id)
	if err := row.Scan(&a.ID, &a.Type, &a.Severity, &a.Message, &a.ResourceID, &a.ResourceType, &a.Acknowledged, &a.CreatedAt); err != nil {
		return nil, fmt.Errorf("get alert: %w", err)
	}
	return &a, nil
}

type ListAlertsResult struct {
	Alerts []Alert
	Total  int
}

func (db *DB) ListAlerts(ctx context.Context, acknowledged *bool, page, limit int) (*ListAlertsResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	whereClause := "1=1"
	args := []interface{}{}
	if acknowledged != nil {
		whereClause += " AND acknowledged = ?"
		args = append(args, *acknowledged)
	}

	var total int
	countRow := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM alerts WHERE "+whereClause, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, fmt.Errorf("count alerts: %w", err)
	}

	query := fmt.Sprintf(`SELECT id, type, severity, message, resource_id, resource_type, acknowledged, created_at
		FROM alerts WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereClause)
	args = append(args, limit, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var a Alert
		if err := rows.Scan(&a.ID, &a.Type, &a.Severity, &a.Message, &a.ResourceID, &a.ResourceType, &a.Acknowledged, &a.CreatedAt); err != nil {
			continue
		}
		alerts = append(alerts, a)
	}
	return &ListAlertsResult{Alerts: alerts, Total: total}, nil
}

func (db *DB) AcknowledgeAlert(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `UPDATE alerts SET acknowledged = TRUE WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("acknowledge alert: %w", err)
	}
	return nil
}

func (db *DB) DeleteAlert(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM alerts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete alert: %w", err)
	}
	return nil
}

func (db *DB) GetRecentAlertByType(ctx context.Context, alertType string, within time.Duration) (*Alert, error) {
	var a Alert
	row := db.QueryRowContext(ctx,
		`SELECT id, type, severity, message, resource_id, resource_type, acknowledged, created_at
		FROM alerts WHERE type = ? AND created_at > ? AND acknowledged = FALSE
		ORDER BY created_at DESC LIMIT 1`,
		alertType, time.Now().Add(-within))
	if err := row.Scan(&a.ID, &a.Type, &a.Severity, &a.Message, &a.ResourceID, &a.ResourceType, &a.Acknowledged, &a.CreatedAt); err != nil {
		return nil, err
	}
	return &a, nil
}

func (db *DB) GetAlertSetting(ctx context.Context, key string) (string, error) {
	var value string
	row := db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key)
	if err := row.Scan(&value); err != nil {
		return "", nil
	}
	return value, nil
}

func (db *DB) SetAlertSetting(ctx context.Context, key, value string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = ?`,
		key, value, value)
	if err != nil {
		return fmt.Errorf("set alert setting: %w", err)
	}
	return nil
}