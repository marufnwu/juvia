package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)

	srcDir := "internal/db/migrations"
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		t.Skipf("migrations dir not found at %s: %v", srcDir, err)
		return nil, func() {}
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(srcDir, entry.Name()))
		if err != nil {
			t.Fatalf("read migration %s: %v", entry.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(migrationsDir, entry.Name()), data, 0644); err != nil {
			t.Fatalf("write migration %s: %v", entry.Name(), err)
		}
	}

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}

	if err := db.ApplyMigrations(migrationsDir); err != nil {
		db.Close()
		t.Fatalf("failed to apply migrations: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func TestCreateApp(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	app, err := db.CreateApp(ctx, "wordpress", "My WordPress Site", 1, "6.0")
	if err != nil {
		t.Fatalf("CreateApp failed: %v", err)
	}

	if app.ID == 0 {
		t.Error("expected non-zero app ID")
	}
	if app.AppType != "wordpress" {
		t.Errorf("expected app_type 'wordpress', got '%s'", app.AppType)
	}
	if app.Name != "My WordPress Site" {
		t.Errorf("expected name 'My WordPress Site', got '%s'", app.Name)
	}
	if app.WebsiteID != 1 {
		t.Errorf("expected website_id 1, got %d", app.WebsiteID)
	}
	if app.Version == nil || *app.Version != "6.0" {
		t.Errorf("expected version '6.0', got '%v'", app.Version)
	}
}

func TestListAppsByWebsite(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := db.CreateApp(ctx, "wordpress", "Site 1", 1, "6.0")
	if err != nil {
		t.Fatalf("CreateApp failed: %v", err)
	}

	_, err = db.CreateApp(ctx, "laravel", "Site 2", 1, "10.0")
	if err != nil {
		t.Fatalf("CreateApp failed: %v", err)
	}

	_, err = db.CreateApp(ctx, "wordpress", "Other Site", 2, "6.0")
	if err != nil {
		t.Fatalf("CreateApp failed: %v", err)
	}

	apps, err := db.ListAppsByWebsite(ctx, 1)
	if err != nil {
		t.Fatalf("ListAppsByWebsite failed: %v", err)
	}

	if len(apps) != 2 {
		t.Errorf("expected 2 apps for website 1, got %d", len(apps))
	}

	for _, app := range apps {
		if app.WebsiteID != 1 {
			t.Errorf("expected website_id 1, got %d", app.WebsiteID)
		}
	}
}

func TestUpdateAppVersion(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	app, _ := db.CreateApp(ctx, "wordpress", "Test", 1, "6.0")

	err := db.UpdateAppVersion(ctx, app.ID, "6.1", true)
	if err != nil {
		t.Fatalf("UpdateAppVersion failed: %v", err)
	}

	updated, err := db.GetAppByID(ctx, app.ID)
	if err != nil {
		t.Fatalf("GetAppByID failed: %v", err)
	}

	if updated.Version == nil || *updated.Version != "6.1" {
		t.Errorf("expected version '6.1', got '%v'", updated.Version)
	}
	if !updated.UpdateAvailable {
		t.Error("expected update_available to be true")
	}
}

func TestDeleteApp(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	app, _ := db.CreateApp(ctx, "wordpress", "Test", 1, "6.0")

	err := db.DeleteApp(ctx, app.ID)
	if err != nil {
		t.Fatalf("DeleteApp failed: %v", err)
	}

	_, err = db.GetAppByID(ctx, app.ID)
	if err == nil {
		t.Error("expected error when getting deleted app")
	}
}

func TestMarkUpdateAvailable(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	app, _ := db.CreateApp(ctx, "wordpress", "Test", 1, "6.0")

	err := db.MarkUpdateAvailable(ctx, app.ID, true)
	if err != nil {
		t.Fatalf("MarkUpdateAvailable failed: %v", err)
	}

	updated, _ := db.GetAppByID(ctx, app.ID)
	if !updated.UpdateAvailable {
		t.Error("expected update_available to be true")
	}

	err = db.MarkUpdateAvailable(ctx, app.ID, false)
	if err != nil {
		t.Fatalf("MarkUpdateAvailable failed: %v", err)
	}

	updated, _ = db.GetAppByID(ctx, app.ID)
	if updated.UpdateAvailable {
		t.Error("expected update_available to be false")
	}
}

func TestListAppsByWebsite_Empty(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	apps, err := db.ListAppsByWebsite(ctx, 999)
	if err != nil {
		t.Fatalf("ListAppsByWebsite failed: %v", err)
	}

	if len(apps) != 0 {
		t.Errorf("expected 0 apps, got %d", len(apps))
	}
}