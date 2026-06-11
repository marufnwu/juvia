package db

import (
	"context"
	"fmt"
	"time"
)

func (db *DB) ListAppsByWebsite(ctx context.Context, websiteID int64) ([]App, error) {
	query := `SELECT id, app_type, name, website_id, version, installed_at, update_available
		FROM apps WHERE website_id = ? ORDER BY installed_at DESC`
	rows, err := db.QueryContext(ctx, query, websiteID)
	if err != nil {
		return nil, fmt.Errorf("list apps: %w", err)
	}
	defer rows.Close()

	var apps []App
	for rows.Next() {
		var a App
		if err := rows.Scan(&a.ID, &a.AppType, &a.Name, &a.WebsiteID, &a.Version, &a.InstalledAt, &a.UpdateAvailable); err != nil {
			continue
		}
		apps = append(apps, a)
	}
	return apps, nil
}

func (db *DB) CreateApp(ctx context.Context, appType, name string, websiteID int64, version string) (*App, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO apps (app_type, name, website_id, version, installed_at) VALUES (?, ?, ?, ?, ?)`,
		appType, name, websiteID, version, time.Now())
	if err != nil {
		return nil, fmt.Errorf("create app: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetAppByID(ctx, id)
}

func (db *DB) GetAppByID(ctx context.Context, id int64) (*App, error) {
	var a App
	row := db.QueryRowContext(ctx,
		`SELECT id, app_type, name, website_id, version, installed_at, update_available
		FROM apps WHERE id = ?`, id)
	if err := row.Scan(&a.ID, &a.AppType, &a.Name, &a.WebsiteID, &a.Version, &a.InstalledAt, &a.UpdateAvailable); err != nil {
		return nil, fmt.Errorf("get app: %w", err)
	}
	return &a, nil
}

func (db *DB) UpdateAppVersion(ctx context.Context, id int64, version string, updateAvailable bool) error {
	_, err := db.ExecContext(ctx,
		`UPDATE apps SET version = ?, update_available = ? WHERE id = ?`,
		version, updateAvailable, id)
	if err != nil {
		return fmt.Errorf("update app version: %w", err)
	}
	return nil
}

func (db *DB) DeleteApp(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM apps WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete app: %w", err)
	}
	return nil
}

func (db *DB) MarkUpdateAvailable(ctx context.Context, id int64, available bool) error {
	_, err := db.ExecContext(ctx,
		`UPDATE apps SET update_available = ? WHERE id = ?`,
		available, id)
	if err != nil {
		return fmt.Errorf("mark update available: %w", err)
	}
	return nil
}