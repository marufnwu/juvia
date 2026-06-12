package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

func HandleServicesStatus(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Service string `json:"service"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	status, err := getServiceStatus(req.Service)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"status": status}, nil
}

func HandleServicesRestart(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Service string `json:"service"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	validServices := map[string]bool{
		"nginx": true, "php-fpm": true, "mysql": true,
		"postfix": true, "dovecot": true, "named": true,
	}

	if !validServices[req.Service] {
		return nil, fmt.Errorf("unknown service: %s", req.Service)
	}

	cmd := exec.Command("systemctl", "restart", req.Service)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("restart %s failed: %w: %s", req.Service, err, string(output))
	}

	return map[string]interface{}{"message": "service restarted"}, nil
}

func getServiceStatus(serviceName string) (string, error) {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "unknown", nil
	}
	status := strings.TrimSpace(string(output))
	switch status {
	case "active":
		return "active", nil
	case "inactive":
		return "inactive", nil
	case "failed":
		return "failed", nil
	default:
		return status, nil
	}
}