package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

func TestApplyMigrations(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)

	migrationContent := `
CREATE TABLE IF NOT EXISTS test_table (
	id INTEGER PRIMARY KEY,
	name TEXT
);
`
	os.WriteFile(filepath.Join(migrationsDir, "001_test.sql"), []byte(migrationContent), 0644)

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer db.Close()

	if err := db.ApplyMigrations(migrationsDir); err != nil {
		t.Fatalf("ApplyMigrations failed: %v", err)
	}

	var count int
	row := db.QueryRow(`SELECT COUNT(*) FROM test_table`)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("query test_table failed: %v", err)
	}
}

func TestWithTx(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	migrationsDir := filepath.Join(dir, "migrations")
	os.MkdirAll(migrationsDir, 0755)
	os.WriteFile(filepath.Join(migrationsDir, "001_test.sql"), []byte(`CREATE TABLE test_tx (id INTEGER PRIMARY KEY);`), 0644)

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB failed: %v", err)
	}
	defer db.Close()
	db.ApplyMigrations(migrationsDir)

	err = db.WithTx(context.Background(), func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO test_tx (id) VALUES (1)`)
		return err
	})
	if err != nil {
		t.Fatalf("WithTx failed: %v", err)
	}
}
