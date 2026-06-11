package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func (db *DB) CreateMailbox(ctx context.Context, email, passwordHash string, quota int64, displayName string) (*Mailbox, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO mailboxes (email, password_hash, quota, display_name, status) VALUES (?, ?, ?, ?, 'active')`,
		email, passwordHash, quota, displayName)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, fmt.Errorf("mailbox already exists")
		}
		return nil, fmt.Errorf("create mailbox: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetMailboxByID(ctx, id)
}

func (db *DB) GetMailboxByID(ctx context.Context, id int64) (*Mailbox, error) {
	var m Mailbox
	row := db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, quota, display_name, forward_to, status, autoresponder_enabled,
		autoresponder_message, autoresponder_start_date, autoresponder_end_date, created_at
		FROM mailboxes WHERE id = ?`, id)
	if err := row.Scan(&m.ID, &m.Email, &m.PasswordHash, &m.Quota, &m.DisplayName, &m.ForwardTo, &m.Status,
		&m.AutoresponderEnabled, &m.AutoresponderMessage, &m.AutoresponderStartDate, &m.AutoresponderEndDate, &m.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mailbox not found")
		}
		return nil, fmt.Errorf("get mailbox: %w", err)
	}
	return&m, nil
}

func (db *DB) GetMailboxByEmail(ctx context.Context, email string) (*Mailbox, error) {
	var m Mailbox
	row := db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, quota, display_name, forward_to, status, autoresponder_enabled,
		autoresponder_message, autoresponder_start_date, autoresponder_end_date, created_at
		FROM mailboxes WHERE email = ?`, email)
	if err := row.Scan(&m.ID, &m.Email, &m.PasswordHash, &m.Quota, &m.DisplayName, &m.ForwardTo, &m.Status,
		&m.AutoresponderEnabled, &m.AutoresponderMessage, &m.AutoresponderStartDate, &m.AutoresponderEndDate, &m.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mailbox not found")
		}
		return nil, fmt.Errorf("get mailbox: %w", err)
	}
	return &m, nil
}

type ListMailboxesResult struct {
	Mailboxes []Mailbox
	Total     int
}

func (db *DB) ListMailboxes(ctx context.Context, search string, page, limit int) (*ListMailboxesResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	whereClause := "1=1"
	args := []interface{}{}
	if search != "" {
		whereClause += " AND (email LIKE ? OR display_name LIKE ?)"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	var total int
	countRow := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mailboxes WHERE "+whereClause, args...)
	if err := countRow.Scan(&total); err != nil {
		return nil, fmt.Errorf("count mailboxes: %w", err)
	}

	query := fmt.Sprintf(`SELECT id, email, password_hash, quota, display_name, forward_to, status,
		autoresponder_enabled, autoresponder_message, autoresponder_start_date, autoresponder_end_date, created_at
		FROM mailboxes WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereClause)
	args = append(args, limit, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list mailboxes: %w", err)
	}
	defer rows.Close()

	var mailboxes []Mailbox
	for rows.Next() {
		var m Mailbox
		if err := rows.Scan(&m.ID, &m.Email, &m.PasswordHash, &m.Quota, &m.DisplayName, &m.ForwardTo, &m.Status,
			&m.AutoresponderEnabled, &m.AutoresponderMessage, &m.AutoresponderStartDate, &m.AutoresponderEndDate, &m.CreatedAt); err != nil {
			continue
		}
		mailboxes = append(mailboxes, m)
	}
	return&ListMailboxesResult{Mailboxes: mailboxes, Total: total}, nil
}

func (db *DB) UpdateMailbox(ctx context.Context, id int64, displayName string, forwardTo string, quota int64) error {
	_, err := db.ExecContext(ctx,
		`UPDATE mailboxes SET display_name = ?, forward_to = ?, quota = ? WHERE id = ?`,
		displayName, forwardTo, quota, id)
	if err != nil {
		return fmt.Errorf("update mailbox: %w", err)
	}
	return nil
}

func (db *DB) DeleteMailbox(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM mailboxes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete mailbox: %w", err)
	}
	return nil
}

func (db *DB) CreateAlias(ctx context.Context, domain, source, destination string) (*EmailAlias, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO email_aliases (domain, source, destination) VALUES (?, ?, ?)`,
		domain, source, destination)
	if err != nil {
		return nil, fmt.Errorf("create alias: %w", err)
	}
	id, _ := result.LastInsertId()
	var a EmailAlias
	row := db.QueryRowContext(ctx, `SELECT id, domain, source, destination, created_at FROM email_aliases WHERE id = ?`, id)
	if err := row.Scan(&a.ID, &a.Domain, &a.Source, &a.Destination, &a.CreatedAt); err != nil {
		return nil, fmt.Errorf("get alias: %w", err)
	}
	return &a, nil
}

func (db *DB) ListAliasesByDomain(ctx context.Context, domain string) ([]EmailAlias, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, domain, source, destination, created_at FROM email_aliases WHERE domain = ?`, domain)
	if err != nil {
		return nil, fmt.Errorf("list aliases: %w", err)
	}
	defer rows.Close()

	var aliases []EmailAlias
	for rows.Next() {
		var a EmailAlias
		if err := rows.Scan(&a.ID, &a.Domain, &a.Source, &a.Destination, &a.CreatedAt); err != nil {
			continue
		}
		aliases = append(aliases, a)
	}
	return aliases, nil
}

func (db *DB) ListAllAliases(ctx context.Context) ([]EmailAlias, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, domain, source, destination, created_at FROM email_aliases ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list all aliases: %w", err)
	}
	defer rows.Close()

	var aliases []EmailAlias
	for rows.Next() {
		var a EmailAlias
		if err := rows.Scan(&a.ID, &a.Domain, &a.Source, &a.Destination, &a.CreatedAt); err != nil {
			continue
		}
		aliases = append(aliases, a)
	}
	return aliases, nil
}

func (db *DB) DeleteAlias(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM email_aliases WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete alias: %w", err)
	}
	return nil
}

func (db *DB) CreateForwarder(ctx context.Context, domain, source, destination string) (*EmailForwarder, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO email_forwarders (domain, source, destination) VALUES (?, ?, ?)`,
		domain, source, destination)
	if err != nil {
		return nil, fmt.Errorf("create forwarder: %w", err)
	}
	id, _ := result.LastInsertId()
	var f EmailForwarder
	row := db.QueryRowContext(ctx, `SELECT id, domain, source, destination, created_at FROM email_forwarders WHERE id = ?`, id)
	if err := row.Scan(&f.ID, &f.Domain, &f.Source, &f.Destination, &f.CreatedAt); err != nil {
		return nil, fmt.Errorf("get forwarder: %w", err)
	}
	return&f, nil
}

func (db *DB) ListForwardersByDomain(ctx context.Context, domain string) ([]EmailForwarder, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, domain, source, destination, created_at FROM email_forwarders WHERE domain = ?`, domain)
	if err != nil {
		return nil, fmt.Errorf("list forwarders: %w", err)
	}
	defer rows.Close()

	var forwarders []EmailForwarder
	for rows.Next() {
		var f EmailForwarder
		if err := rows.Scan(&f.ID, &f.Domain, &f.Source, &f.Destination, &f.CreatedAt); err != nil {
			continue
		}
		forwarders = append(forwarders, f)
	}
	return forwarders, nil
}

func (db *DB) ListAllForwarders(ctx context.Context) ([]EmailForwarder, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, domain, source, destination, created_at FROM email_forwarders ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list all forwarders: %w", err)
	}
	defer rows.Close()

	var forwarders []EmailForwarder
	for rows.Next() {
		var f EmailForwarder
		if err := rows.Scan(&f.ID, &f.Domain, &f.Source, &f.Destination, &f.CreatedAt); err != nil {
			continue
		}
		forwarders = append(forwarders, f)
	}
	return forwarders, nil
}

func (db *DB) DeleteForwarder(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM email_forwarders WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete forwarder: %w", err)
	}
	return nil
}

func (db *DB) GetCatchAll(ctx context.Context, domain string) (*EmailCatchAll, error) {
	var c EmailCatchAll
	row := db.QueryRowContext(ctx, `SELECT id, domain, forward_to, created_at FROM email_catchall WHERE domain = ?`, domain)
	if err := row.Scan(&c.ID, &c.Domain, &c.ForwardTo, &c.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get catchall: %w", err)
	}
	return &c, nil
}

func (db *DB) SetCatchAll(ctx context.Context, domain, forwardTo string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO email_catchall (domain, forward_to) VALUES (?, ?)
		ON CONFLICT(domain) DO UPDATE SET forward_to = ?`,
		domain, forwardTo, forwardTo)
	if err != nil {
		return fmt.Errorf("set catchall: %w", err)
	}
	return nil
}

func (db *DB) DeleteCatchAll(ctx context.Context, domain string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM email_catchall WHERE domain = ?`, domain)
	if err != nil {
		return fmt.Errorf("delete catchall: %w", err)
	}
	return nil
}
