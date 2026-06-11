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

	os.MkdirAll(cronDir, 0755)

	cronLine := fmt.Sprintf("%s root %s # juvia:%d\n", req.Schedule, req.Command, req.ID)
	cronFile := filepath.Join(cronDir, "juvia")

	f, err := os.OpenFile(cronFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open cron file: %w", err)
	}
	f.WriteString(cronLine)
	f.Close()

	return map[string]interface{}{
		"id":       req.ID,
		"schedule": req.Schedule,
		"command":  req.Command,
	}, nil
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

	os.WriteFile(cronFile, []byte(strings.Join(newLines, "\n")), 0644)

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
}