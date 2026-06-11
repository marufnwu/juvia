package db

import (
	"context"
	"fmt"
)

func (db *DB) CreateFirewallRule(ctx context.Context, name, description, action, port, protocol, source string) (*FirewallRule, error) {
	result, err := db.ExecContext(ctx,
		`INSERT INTO firewall_rules (name, description, action, port, protocol, source) VALUES (?, ?, ?, ?, ?, ?)`,
		name, description, action, port, protocol, source)
	if err != nil {
		return nil, fmt.Errorf("create firewall rule: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetFirewallRuleByID(ctx, id)
}

func (db *DB) GetFirewallRuleByID(ctx context.Context, id int64) (*FirewallRule, error) {
	var r FirewallRule
	row := db.QueryRowContext(ctx,
		`SELECT id, name, description, action, port, protocol, source, created_at
		FROM firewall_rules WHERE id = ?`, id)
	if err := row.Scan(&r.ID, &r.Name, &r.Description, &r.Action, &r.Port, &r.Protocol, &r.Source, &r.CreatedAt); err != nil {
		return nil, fmt.Errorf("get firewall rule: %w", err)
	}
	return &r, nil
}

type ListFirewallRulesResult struct {
	Rules []FirewallRule
	Total int
}

func (db *DB) ListFirewallRules(ctx context.Context, page, limit int) (*ListFirewallRulesResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	var total int
	countRow := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM firewall_rules")
	if err := countRow.Scan(&total); err != nil {
		return nil, fmt.Errorf("count firewall rules: %w", err)
	}

	query := `SELECT id, name, description, action, port, protocol, source, created_at
		FROM firewall_rules ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list firewall rules: %w", err)
	}
	defer rows.Close()

	var rules []FirewallRule
	for rows.Next() {
		var r FirewallRule
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Action, &r.Port, &r.Protocol, &r.Source, &r.CreatedAt); err != nil {
			continue
		}
		rules = append(rules, r)
	}
	return &ListFirewallRulesResult{Rules: rules, Total: total}, nil
}

func (db *DB) UpdateFirewallRule(ctx context.Context, id int64, name, description, action, port, protocol, source string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE firewall_rules SET name = ?, description = ?, action = ?, port = ?, protocol = ?, source = ? WHERE id = ?`,
		name, description, action, port, protocol, source, id)
	if err != nil {
		return fmt.Errorf("update firewall rule: %w", err)
	}
	return nil
}

func (db *DB) DeleteFirewallRule(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM firewall_rules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete firewall rule: %w", err)
	}
	return nil
}

func (db *DB) GetFirewallRuleByPortAndSource(ctx context.Context, port, source string) (*FirewallRule, error) {
	var r FirewallRule
	query := `SELECT id, name, description, action, port, protocol, source, created_at
		FROM firewall_rules WHERE port = ? AND source = ?`
	row := db.QueryRowContext(ctx, query, port, source)
	if err := row.Scan(&r.ID, &r.Name, &r.Description, &r.Action, &r.Port, &r.Protocol, &r.Source, &r.CreatedAt); err != nil {
		return nil, err
	}
	return &r, nil
}