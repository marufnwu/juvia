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

func HandleServicesCheckInstalled(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Service string `json:"service"`
		Command string `json:"command"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	// Use full PATH to check if command exists
	cmd := exec.Command("bash", "-c", "command -v "+req.Command)
	if output, err := cmd.CombinedOutput(); err == nil && strings.TrimSpace(string(output)) != "" {
		return map[string]interface{}{"installed": true}, nil
	}

	// For php-fpm, also check versioned service names
	if req.Service == "php-fpm" {
		for _, v := range []string{"php8.4-fpm", "php8.3-fpm", "php8.2-fpm", "php8.1-fpm", "php8.0-fpm", "php7.4-fpm"} {
			cmd := exec.Command("bash", "-c", "command -v "+v)
			if output, err := cmd.CombinedOutput(); err == nil && strings.TrimSpace(string(output)) != "" {
				return map[string]interface{}{"installed": true}, nil
			}
		}
		// Also check systemctl for php-fpm service
		out, _ := exec.Command("bash", "-c", "systemctl list-unit-files --type=service --no-pager 2>/dev/null").CombinedOutput()
		s := string(out)
		for _, v := range []string{"php8.4-fpm", "php8.3-fpm", "php8.2-fpm", "php8.1-fpm", "php8.0-fpm", "php7.4-fpm"} {
			if strings.Contains(s, v+".service") {
				return map[string]interface{}{"installed": true}, nil
			}
		}
	}

	// For named, check if the systemd unit exists
	if req.Service == "named" {
		out, _ := exec.Command("bash", "-c", "systemctl list-unit-files --type=service --no-pager 2>/dev/null").CombinedOutput()
		s := string(out)
		if strings.Contains(s, "named.service") || strings.Contains(s, "bind9.service") {
			return map[string]interface{}{"installed": true}, nil
		}
	}

	// For certbot, check if binary exists anywhere
	if req.Service == "certbot" {
		cmd := exec.Command("bash", "-c", "command -v certbot || find /usr -name certbot -type f 2>/dev/null | head -1")
		if output, err := cmd.CombinedOutput(); err == nil && strings.TrimSpace(string(output)) != "" {
			return map[string]interface{}{"installed": true}, nil
		}
	}

	return map[string]interface{}{"installed": false}, nil
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
		"ufw": true, "rspamd": true,
	}

	if !validServices[req.Service] {
		return nil, fmt.Errorf("unknown service: %s", req.Service)
	}

	resolvedName := resolveServiceName(req.Service)
	cmd := exec.Command("systemctl", "restart", resolvedName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("restart %s failed: %w: %s", req.Service, err, string(output))
	}

	return map[string]interface{}{"message": "service restarted"}, nil
}

func getServiceStatus(serviceName string) (string, error) {
	serviceName = resolveServiceName(serviceName)
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

func resolveServiceName(name string) string {
	if name == "php-fpm" {
		// Find the actual versioned service name
		out, _ := exec.Command("systemctl", "list-unit-files", "--type=service", "--no-pager").CombinedOutput()
		s := string(out)
		for _, v := range []string{"php8.4-fpm", "php8.3-fpm", "php8.2-fpm", "php8.1-fpm", "php8.0-fpm", "php7.4-fpm"} {
			if strings.Contains(s, v) {
				return v
			}
		}
	}
	if name == "named" {
		// Check if bind9.service is the actual name
		out, _ := exec.Command("systemctl", "list-unit-files", "--type=service", "--no-pager").CombinedOutput()
		s := string(out)
		if strings.Contains(s, "bind9.service") {
			return "bind9"
		}
	}
	return name
}