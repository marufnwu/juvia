package database

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func HandleDatabaseCreate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Name   string `json:"name"`
		Engine string `json:"engine"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	if req.Engine == "mysql" {
		cmd := exec.Command("mysql", "-e", "CREATE DATABASE IF NOT EXISTS `"+req.Name+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;")
		if output, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("mysql create: %w: %s", err, string(output))
		}
	} else if req.Engine == "postgresql" {
		cmd := exec.Command("createdb", "-E", "UTF8", req.Name)
		if output, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("postgresql create: %w: %s", err, string(output))
		}
	}

	return map[string]interface{}{
		"name":   req.Name,
		"engine": req.Engine,
	}, nil
}

func HandleDatabaseDelete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Name   string `json:"name"`
		Engine string `json:"engine"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	if req.Engine == "mysql" {
		cmd := exec.Command("mysql", "-e", "DROP DATABASE IF EXISTS `"+req.Name+"`;")
		cmd.Run()
	} else if req.Engine == "postgresql" {
		cmd := exec.Command("dropdb", req.Name)
		cmd.Run()
	}

	return map[string]interface{}{"deleted": true}, nil
}

func HandleDBUserCreate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Name     string `json:"name"`
		Engine   string `json:"engine"`
		Username string `json:"username"`
		Password string `json:"password"`
		Host     string `json:"host"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Username == "" || req.Password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	if req.Host == "" {
		req.Host = "localhost"
	}

	if req.Engine == "mysql" {
		cmd := exec.Command("mysql", "-e",
			fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%s' IDENTIFIED WITH caching_sha2_password BY '%s'; GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%s'; FLUSH PRIVILEGES;",
				req.Username, req.Host, req.Password, req.Name, req.Username, req.Host))
		if output, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("mysql user create: %w: %s", err, string(output))
		}
	} else if req.Engine == "postgresql" {
		cmd := exec.Command("psql", "-c",
			fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'; GRANT ALL PRIVILEGES ON DATABASE %s TO %s;",
				req.Username, req.Password, req.Name, req.Username))
		if output, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("postgresql user create: %w: %s", err, string(output))
		}
	}

	return map[string]interface{}{
		"username": req.Username,
		"host":     req.Host,
	}, nil
}

func HandleDBUserDelete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Name     string `json:"name"`
		Engine   string `json:"engine"`
		Username string `json:"username"`
		Host     string `json:"host"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.Host == "" {
		req.Host = "localhost"
	}

	if req.Engine == "mysql" {
		cmd := exec.Command("mysql", "-e",
			fmt.Sprintf("DROP USER IF EXISTS '%s'@'%s';", req.Username, req.Host))
		cmd.Run()
	} else if req.Engine == "postgresql" {
		cmd := exec.Command("psql", "-c", fmt.Sprintf("DROP USER IF EXISTS %s;", req.Username))
		cmd.Run()
	}

	return map[string]interface{}{"deleted": true}, nil
}

func HandleExport(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Name   string `json:"name"`
		Engine string `json:"engine"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	var path string
	if req.Engine == "mysql" {
		cmd := exec.Command("mysqldump", req.Name)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("mysqldump: %w", err)
		}
		path = "/tmp/" + req.Name + ".sql"
		os.WriteFile(path, output, 0600)
	} else if req.Engine == "postgresql" {
		cmd := exec.Command("pg_dump", req.Name)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("pg_dump: %w", err)
		}
		path = "/tmp/" + req.Name + ".sql"
		os.WriteFile(path, output, 0600)
	}

	return map[string]interface{}{
		"path": path,
		"size": 0,
	}, nil
}

func HandleListTables(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Name   string `json:"name"`
		Engine string `json:"engine"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	var tables []map[string]interface{}

	if req.Engine == "mysql" {
		cmd := exec.Command("mysql", req.Name, "-e", "SHOW TABLES;")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("mysql show tables: %w", err)
		}
		lines := strings.Split(string(output), "\n")
		for i, line := range lines {
			if i == 0 || line == "" || strings.HasPrefix(line, "Tables_in") {
				continue
			}
			tables = append(tables, map[string]interface{}{
				"name": strings.TrimSpace(line),
			})
		}
	} else if req.Engine == "postgresql" {
		cmd := exec.Command("psql", req.Name, "-t", "-c", "SELECT tablename FROM pg_tables WHERE schemaname='public';")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("postgresql show tables: %w", err)
		}
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			name := strings.TrimSpace(line)
			if name != "" {
				tables = append(tables, map[string]interface{}{
					"name": name,
				})
			}
		}
	}

	return map[string]interface{}{"tables": tables}, nil
}

func HandleGetRows(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Name   string `json:"name"`
		Engine string `json:"engine"`
		Table  string `json:"table"`
		Limit  int    `json:"limit"`
		Offset int    `json:"offset"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.Limit == 0 {
		req.Limit = 50
	}

	var rows []map[string]interface{}

	if req.Engine == "mysql" {
		query := fmt.Sprintf("SELECT * FROM `%s` LIMIT %d OFFSET %d", req.Table, req.Limit, req.Offset)
		cmd := exec.Command("mysql", req.Name, "-e", query)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("mysql select: %w", err)
		}
		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			headers := strings.Split(lines[0], "\t")
			for _, line := range lines[1:] {
				if line == "" {
					continue
				}
				cols := strings.Split(line, "\t")
				row := make(map[string]interface{})
				for i, h := range headers {
					if i < len(cols) {
						row[strings.TrimSpace(h)] = strings.TrimSpace(cols[i])
					}
				}
				rows = append(rows, row)
			}
		}
	} else if req.Engine == "postgresql" {
		query := fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", req.Table, req.Limit, req.Offset)
		cmd := exec.Command("psql", req.Name, "-c", query)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("postgresql select: %w", err)
		}
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "(") && strings.HasSuffix(line, ")") {
				row := make(map[string]interface{})
				row["data"] = line
				rows = append(rows, row)
			}
		}
	}

	return map[string]interface{}{
		"rows": rows,
		"meta": map[string]interface{}{
			"limit":  req.Limit,
			"offset": req.Offset,
		},
	}, nil
}

func HandleQuery(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Name   string `json:"name"`
		Engine string `json:"engine"`
		Query  string `json:"query"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	isSelect := strings.HasPrefix(strings.ToUpper(req.Query), "SELECT")
	if !isSelect {
		return nil, fmt.Errorf("only SELECT queries are allowed in the panel")
	}

	var result []map[string]interface{}

	if req.Engine == "mysql" {
		cmd := exec.Command("mysql", req.Name, "-e", req.Query)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("mysql query: %w", err)
		}
		lines := strings.Split(string(output), "\n")
		if len(lines) > 1 {
			headers := strings.Split(lines[0], "\t")
			for _, line := range lines[1:] {
				if line == "" {
					continue
				}
				cols := strings.Split(line, "\t")
				row := make(map[string]interface{})
				for i, h := range headers {
					if i < len(cols) {
						row[strings.TrimSpace(h)] = strings.TrimSpace(cols[i])
					}
				}
				result = append(result, row)
			}
		}
	} else if req.Engine == "postgresql" {
		cmd := exec.Command("psql", req.Name, "-c", req.Query)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("postgresql query: %w", err)
		}
		result = append(result, map[string]interface{}{
			"output": string(output),
		})
	}

	return map[string]interface{}{
		"rows": result,
	}, nil
}

func getMySQLVersion() string {
	cmd := exec.Command("mysql", "--version")
	output, _ := cmd.Output()
	return string(output)
}

func getPostgreSQLVersion() string {
	cmd := exec.Command("psql", "--version")
	output, _ := cmd.Output()
	return string(output)
}

func importSQL(database, engine, path string) error {
	if engine == "mysql" {
		cmd := exec.Command("mysql", database)
		cmd.Stdin, _ = os.Open(path)
		return cmd.Run()
	} else if engine == "postgresql" {
		cmd := exec.Command("psql", database)
		cmd.Stdin, _ = os.Open(path)
		return cmd.Run()
	}
	return nil
}

func formatMySQLRow(line string) map[string]interface{} {
	cols := strings.Split(line, "\t")
	row := make(map[string]interface{})
	for i, c := range cols {
		row[strconv.Itoa(i)] = c
	}
	return row
}
