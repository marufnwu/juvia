package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func (db *DB) CreateDatabase(ctx context.Context, name, engine string, userID int64) (*Database, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO databases (name, engine, user_id) VALUES (?, ?, ?)`,
		name, engine, userID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, fmt.Errorf("database already exists")
		}
		return nil, fmt.Errorf("create database: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetDatabaseByID(ctx, id)
}

func (db *DB) GetDatabaseByID(ctx context.Context, id int64) (*Database, error) {
	var d Database
	row := db.QueryRowContext(ctx,
		`SELECT id, name, engine, user_id, created_at FROM databases WHERE id = ?`, id)
	if err := row.Scan(&d.ID, &d.Name, &d.Engine, &d.UserID, &d.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("database not found")
		}
		return nil, fmt.Errorf("get database: %w", err)
	}
	return &d, nil
}

type ListDatabasesResult struct {
	Databases []Database
	Total     int
}

func (db *DB) ListDatabases(ctx context.Context, engine string, page, limit int) (*ListDatabasesResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	whereClause := "1=1"
	args := []interface{}{}
	if engine != "" {
		whereClause += " AND engine = ?"
		args = append(args, engine)
	}

	var total int
	countRow := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM databases WHERE "+whereClause, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, fmt.Errorf("count databases: %w", err)
	}

	query := fmt.Sprintf(`SELECT id, name, engine, user_id, created_at
		FROM databases WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereClause)
	args = append(args, limit, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}
	defer rows.Close()

	var databases []Database
	for rows.Next() {
		var d Database
		if err := rows.Scan(&d.ID, &d.Name, &d.Engine, &d.UserID, &d.CreatedAt); err != nil {
			continue
		}
		databases = append(databases, d)
	}
	return&ListDatabasesResult{Databases: databases, Total: total}, nil
}

func (db *DB) DeleteDatabase(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM databases WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete database: %w", err)
	}
	return nil
}

func (db *DB) CreateDBUser(ctx context.Context, databaseID int64, username, passwordHash, host string) (*DBUser, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO db_users (database_id, username, password_hash, host) VALUES (?, ?, ?, ?)`,
		databaseID, username, passwordHash, host)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, fmt.Errorf("user already exists")
		}
		return nil, fmt.Errorf("create db user: %w", err)
	}
	id, _ := result.LastInsertId()
	var u DBUser
	row := db.QueryRowContext(ctx, `SELECT id, database_id, username, password_hash, host, created_at FROM db_users WHERE id = ?`, id)
	if err := row.Scan(&u.ID, &u.DatabaseID, &u.Username, &u.PasswordHash, &u.Host, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("get db user: %w", err)
	}
	return&u, nil
}

func (db *DB) ListDBUsers(ctx context.Context, databaseID int64) ([]DBUser, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, database_id, username, password_hash, host, created_at FROM db_users WHERE database_id = ?`, databaseID)
	if err != nil {
		return nil, fmt.Errorf("list db users: %w", err)
	}
	defer rows.Close()

	var users []DBUser
	for rows.Next() {
		var u DBUser
		if err := rows.Scan(&u.ID, &u.DatabaseID, &u.Username, &u.PasswordHash, &u.Host, &u.CreatedAt); err != nil {
			continue
		}
		users = append(users, u)
	}
	return users, nil
}

func (db *DB) DeleteDBUser(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM db_users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete db user: %w", err)
	}
	return nil
}
