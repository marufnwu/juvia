package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func copyMigrations(t *testing.T, destDir string) {
	srcDir := "internal/db/migrations"
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		t.Skipf("migrations dir not found at %s: %v", srcDir, err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(srcDir, entry.Name()))
		if err != nil {
			t.Fatalf("read migration %s: %v", entry.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(destDir, entry.Name()), data, 0644); err != nil {
			t.Fatalf("write migration %s: %v", entry.Name(), err)
		}
	}
}

func TestCreateAndGetWebsite(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)
	copyMigrations(t, migrationsDir)

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.ApplyMigrations(migrationsDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	ctx := context.Background()
	w, err := db.CreateWebsite(ctx, "example.com", "/home/example/public_html", "8.2", "nginx", 1)
	if err != nil {
		t.Fatalf("create website: %v", err)
	}
	if w.Domain != "example.com" {
		t.Errorf("expected domain example.com, got %s", w.Domain)
	}
	if w.PHPVersion == nil || *w.PHPVersion != "8.2" {
		t.Errorf("expected php version 8.2, got %v", w.PHPVersion)
	}
	if w.WebServer != "nginx" {
		t.Errorf("expected web server nginx, got %s", w.WebServer)
	}
	if w.Status != "active" {
		t.Errorf("expected status active, got %s", w.Status)
	}

	w2, err := db.GetWebsiteByID(ctx, w.ID)
	if err != nil {
		t.Fatalf("get website: %v", err)
	}
	if w2.Domain != "example.com" {
		t.Errorf("expected domain example.com, got %s", w2.Domain)
	}
}

func TestGetWebsiteByDomain(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)
	copyMigrations(t, migrationsDir)

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.ApplyMigrations(migrationsDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	ctx := context.Background()
	_, err = db.CreateWebsite(ctx, "example.com", "/home/example/public_html", "8.2", "nginx", 1)
	if err != nil {
		t.Fatalf("create website: %v", err)
	}

	w, err := db.GetWebsiteByDomain(ctx, "example.com")
	if err != nil {
		t.Fatalf("get website by domain: %v", err)
	}
	if w.Domain != "example.com" {
		t.Errorf("expected domain example.com, got %s", w.Domain)
	}
}

func TestListWebsites(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)
	copyMigrations(t, migrationsDir)

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.ApplyMigrations(migrationsDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		domain := "site" + string(rune('a'+i)) + ".com"
		_, err := db.CreateWebsite(ctx, domain, "/home/site/public_html", "8.2", "nginx", 1)
		if err != nil {
			t.Fatalf("create website: %v", err)
		}
	}

	result, err := db.ListWebsites(ctx, "", "", "created_at", "desc", 1, 10, 0)
	if err != nil {
		t.Fatalf("list websites: %v", err)
	}
	if len(result.Websites) != 5 {
		t.Errorf("expected 5 websites, got %d", len(result.Websites))
	}
	if result.Total != 5 {
		t.Errorf("expected total 5, got %d", result.Total)
	}
}

func TestSoftDeleteAndRestoreWebsite(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)
	copyMigrations(t, migrationsDir)

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.ApplyMigrations(migrationsDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	ctx := context.Background()
	w, _ := db.CreateWebsite(ctx, "example.com", "/home/example/public_html", "8.2", "nginx", 1)

	err = db.SoftDeleteWebsite(ctx, w.ID)
	if err != nil {
		t.Fatalf("soft delete: %v", err)
	}

	w2, _ := db.GetWebsiteByID(ctx, w.ID)
	if w2.Status != "deleted" {
		t.Errorf("expected status deleted, got %s", w2.Status)
	}
	if w2.DeletedAt == nil {
		t.Errorf("expected deleted_at to be set")
	}

	err = db.RestoreWebsite(ctx, w.ID)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}

	w3, _ := db.GetWebsiteByID(ctx, w.ID)
	if w3.Status != "active" {
		t.Errorf("expected status active, got %s", w3.Status)
	}
}

func TestSuspendWebsite(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)
	copyMigrations(t, migrationsDir)

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.ApplyMigrations(migrationsDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	ctx := context.Background()
	w, _ := db.CreateWebsite(ctx, "example.com", "/home/example/public_html", "8.2", "nginx", 1)

	err = db.SuspendWebsite(ctx, w.ID)
	if err != nil {
		t.Fatalf("suspend: %v", err)
	}

	w2, _ := db.GetWebsiteByID(ctx, w.ID)
	if w2.Status != "suspended" {
		t.Errorf("expected status suspended, got %s", w2.Status)
	}
}

func TestCreateDomain(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)
	copyMigrations(t, migrationsDir)

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.ApplyMigrations(migrationsDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	ctx := context.Background()
	w, _ := db.CreateWebsite(ctx, "example.com", "/home/example/public_html", "8.2", "nginx", 1)

	err = db.CreateDomain(ctx, w.ID, "blog.example.com", "subdomain")
	if err != nil {
		t.Fatalf("create domain: %v", err)
	}

	domains, err := db.ListDomainsByWebsite(ctx, w.ID)
	if err != nil {
		t.Fatalf("list domains: %v", err)
	}
	if len(domains) != 1 {
		t.Errorf("expected 1 domain, got %d", len(domains))
	}
	if domains[0].Domain != "blog.example.com" {
		t.Errorf("expected blog.example.com, got %s", domains[0].Domain)
	}
}

func TestCreateZoneAndDNSRecords(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)
	copyMigrations(t, migrationsDir)

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.ApplyMigrations(migrationsDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	ctx := context.Background()
	w, _ := db.CreateWebsite(ctx, "example.com", "/home/example/public_html", "8.2", "nginx", 1)

	z, err := db.CreateZone(ctx, w.ID, "example.com")
	if err != nil {
		t.Fatalf("create zone: %v", err)
	}
	if z.Domain != "example.com" {
		t.Errorf("expected domain example.com, got %s", z.Domain)
	}

	priority := 10
	r, err := db.CreateDNSRecord(ctx, z.ID, "A", "@", "192.168.1.1", &priority)
	if err != nil {
		t.Fatalf("create dns record: %v", err)
	}
	if r.Type != "A" {
		t.Errorf("expected type A, got %s", r.Type)
	}
	if r.Name != "@" {
		t.Errorf("expected name @, got %s", r.Name)
	}

	records, err := db.ListDNSRecordsByZone(ctx, z.ID)
	if err != nil {
		t.Fatalf("list records: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("expected 1 record, got %d", len(records))
	}
}

func TestUpdateWebsiteSSL(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)
	copyMigrations(t, migrationsDir)

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.ApplyMigrations(migrationsDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	ctx := context.Background()
	w, _ := db.CreateWebsite(ctx, "example.com", "/home/example/public_html", "8.2", "nginx", 1)

	expiry := time.Now().Add(90 * 24 * time.Hour)
	err = db.UpdateWebsiteSSL(ctx, w.ID, true, &expiry)
	if err != nil {
		t.Fatalf("update ssl: %v", err)
	}

	w2, _ := db.GetWebsiteByID(ctx, w.ID)
	if !w2.SSLEnabled {
		t.Errorf("expected ssl enabled")
	}
	if w2.SSLExpiry == nil {
		t.Errorf("expected ssl expiry to be set")
	}
}

func TestLogAudit(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)
	copyMigrations(t, migrationsDir)

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.ApplyMigrations(migrationsDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	ctx := context.Background()
	uid := int64(1)
	err = db.LogAudit(ctx, &uid, "Created website example.com", "127.0.0.1", "TestAgent", "")
	if err != nil {
		t.Fatalf("log audit: %v", err)
	}
}
