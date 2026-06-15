package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const cronDir = "/etc/cron.d"

func HandleCronCreate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		ID       int64  `json:"id"`
		Schedule string `json:"schedule"`
		Command  string `json:"command"`
		RunAs    string `json:"run_as"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if err := validateCronExpression(req.Schedule); err != nil {
		return nil, err
	}

	if req.RunAs == "" {
		req.RunAs = "root"
	}

	os.MkdirAll(cronDir, 0755)

	runAs := req.RunAs
	cronLine := fmt.Sprintf("%s %s %s # juvia:%d\n", req.Schedule, runAs, req.Command, req.ID)
	cronFile := filepath.Join(cronDir, "juvia")

	f, err := os.OpenFile(cronFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open cron file: %w", err)
	}
	f.WriteString(cronLine)
	f.Close()

	reloadCron()

	return map[string]interface{}{
		"id":       req.ID,
		"schedule": req.Schedule,
		"command":  req.Command,
		"run_as":   runAs,
	}, nil
}

func HandleCronUpdate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		ID       int64  `json:"id"`
		Schedule string `json:"schedule"`
		Command  string `json:"command"`
		RunAs    string `json:"run_as"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.Schedule != "" && validateCronExpression(req.Schedule) != nil {
		return nil, fmt.Errorf("invalid cron expression: %s", req.Schedule)
	}

	cronFile := filepath.Join(cronDir, "juvia")
	data, err := os.ReadFile(cronFile)
	if err != nil {
		return nil, fmt.Errorf("read cron file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var newLines []string
	found := false
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("juvia:%d", req.ID)) {
			found = true
			parts := strings.Split(line, "# juvia:")
			if len(parts) >= 1 {
				existingParts := strings.Fields(parts[0])
				if len(existingParts) >= 5 {
					schedule := req.Schedule
					if schedule == "" {
						schedule = existingParts[0]
					}
					runAs := req.RunAs
					if runAs == "" {
						runAs = existingParts[4]
					}
					command := req.Command
					if command == "" {
						command = strings.Join(existingParts[5:], " ")
					}
					newLines = append(newLines, fmt.Sprintf("%s %s %s # juvia:%d", schedule, runAs, command, req.ID))
				}
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		return nil, fmt.Errorf("cron job %d not found", req.ID)
	}

	if err := os.WriteFile(cronFile, []byte(strings.Join(newLines, "\n")+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("write cron file: %w", err)
	}

	reloadCron()

	return map[string]interface{}{
		"id":       req.ID,
		"schedule": req.Schedule,
		"command":  req.Command,
		"run_as":   req.RunAs,
	}, nil
}

func HandleCronEnable(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	cronFile := filepath.Join(cronDir, "juvia")
	data, err := os.ReadFile(cronFile)
	if err != nil {
		return nil, fmt.Errorf("read cron file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var disabledLine string
	var newLines []string
	found := false

	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("juvia:%d", req.ID)) {
			found = true
			if strings.HasPrefix(line, "# DISABLED: ") {
				disabledLine = strings.TrimPrefix(line, "# DISABLED: ")
				newLines = append(newLines, disabledLine)
			} else {
				return map[string]interface{}{"id": req.ID, "enabled": true}, nil
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found || disabledLine == "" {
		return nil, fmt.Errorf("disabled cron job %d not found", req.ID)
	}

	if err := os.WriteFile(cronFile, []byte(strings.Join(newLines, "\n")+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("write cron file: %w", err)
	}

	reloadCron()

	return map[string]interface{}{"id": req.ID, "enabled": true}, nil
}

func HandleCronDisable(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	cronFile := filepath.Join(cronDir, "juvia")
	data, err := os.ReadFile(cronFile)
	if err != nil {
		return nil, fmt.Errorf("read cron file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var newLines []string
	found := false

	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf("juvia:%d", req.ID)) {
			found = true
			if !strings.HasPrefix(line, "# DISABLED: ") {
				newLines = append(newLines, "# DISABLED: "+line)
			} else {
				newLines = append(newLines, line)
			}
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		return nil, fmt.Errorf("cron job %d not found", req.ID)
	}

	if err := os.WriteFile(cronFile, []byte(strings.Join(newLines, "\n")+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("write cron file: %w", err)
	}

	reloadCron()

	return map[string]interface{}{"id": req.ID, "enabled": false}, nil
}

func HandleCronDelete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	cronFile := filepath.Join(cronDir, "juvia")
	data, err := os.ReadFile(cronFile)
	if err != nil {
		return nil, nil
	}

	lines := strings.Split(string(data), "\n")
	var newLines []string
	for _, line := range lines {
		if !strings.Contains(line, fmt.Sprintf("juvia:%d", req.ID)) {
			newLines = append(newLines, line)
		}
	}

	if err := os.WriteFile(cronFile, []byte(strings.Join(newLines, "\n")+"\n"), 0644); err != nil {
		return nil, fmt.Errorf("write cron file: %w", err)
	}

	reloadCron()

	return map[string]interface{}{"deleted": true}, nil
}

func HandleCronRun(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		ID      int64  `json:"id"`
		Command string `json:"command"`
		RunAs   string `json:"run_as"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	start := time.Now()

	var cmd *exec.Cmd
	if req.RunAs == "root" || req.RunAs == "" {
		cmd = exec.Command("bash", "-c", req.Command)
	} else {
		cmd = exec.Command("su", "-", req.RunAs, "-c", req.Command)
	}

	output, err := cmd.CombinedOutput()
	duration := time.Since(start).Milliseconds()

	status := "success"
	if err != nil {
		status = "failed"
	}

	return map[string]interface{}{
		"id":          req.ID,
		"status":      status,
		"output":      string(output),
		"duration_ms": duration,
	}, nil
}

func validateCronExpression(expr string) error {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return fmt.Errorf("cron expression must have exactly 5 fields")
	}
	return nil
}

func reloadCron() {
	exec.Command("systemctl", "reload", "cron").Run()
	exec.Command("systemctl", "reload", "crond").Run()
}
