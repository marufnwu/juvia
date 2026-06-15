package ssh

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const sshdConfigPath = "/etc/ssh/sshd_config"

type sshdSetting struct {
	Key   string
	Value string
}

func HandleSSHConfigure(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Port       int    `json:"port"`
		RootLogin  string `json:"root_login"`
		TaskID     string `json:"task_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.Port < 1 || req.Port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535")
	}

	if req.RootLogin == "" {
		req.RootLogin = "prohibit-password"
	}

	validRootLogin := map[string]bool{
		"yes":               true,
		"no":                true,
		"prohibit-password": true,
		"without-password":  true,
	}
	if !validRootLogin[req.RootLogin] {
		return nil, fmt.Errorf("invalid root_login value: must be yes, no, prohibit-password, or without-password")
	}

	if err := writeSSHDConfig(req.Port, req.RootLogin); err != nil {
		return nil, fmt.Errorf("write sshd_config: %w", err)
	}

	if err := validateSSHDConfig(); err != nil {
		os.Remove(sshdConfigPath + ".tmp")
		return nil, fmt.Errorf("validate sshd_config: %w", err)
	}

	if err := reloadSSHD(); err != nil {
		return nil, fmt.Errorf("reload sshd: %w", err)
	}

	slog.Info("ssh configured", "port", req.Port, "root_login", req.RootLogin)

	return map[string]interface{}{
		"port":       req.Port,
		"root_login": req.RootLogin,
	}, nil
}

func HandleSSHReload(ctx context.Context, params json.RawMessage) (interface{}, error) {
	if err := reloadSSHD(); err != nil {
		return nil, fmt.Errorf("reload sshd: %w", err)
	}

	return map[string]interface{}{"reloaded": true}, nil
}

func writeSSHDConfig(port int, rootLogin string) error {
	settings, err := parseSSHDConfig(sshdConfigPath)
	if err != nil {
		slog.Warn("could not read existing sshd_config, using defaults", "error", err)
		settings = defaultSSHDSettings()
	}

	for i, s := range settings {
		switch s.Key {
		case "Port":
			settings[i].Value = strconv.Itoa(port)
		case "PermitRootLogin":
			settings[i].Value = rootLogin
		}
	}

	hasPort := false
	hasRootLogin := false
	for _, s := range settings {
		if s.Key == "Port" {
			hasPort = true
		}
		if s.Key == "PermitRootLogin" {
			hasRootLogin = true
		}
	}
	if !hasPort {
		settings = append(settings, sshdSetting{Key: "Port", Value: strconv.Itoa(port)})
	}
	if !hasRootLogin {
		settings = append(settings, sshdSetting{Key: "PermitRootLogin", Value: rootLogin})
	}

	var sb strings.Builder
	for _, s := range settings {
		sb.WriteString(s.Key)
		sb.WriteString(" ")
		sb.WriteString(s.Value)
		sb.WriteString("\n")
	}

	tmpPath := sshdConfigPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(sb.String()), 0600); err != nil {
		return fmt.Errorf("write tmp config: %w", err)
	}

	if err := os.Rename(tmpPath, sshdConfigPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("atomic rename: %w", err)
	}

	return nil
}

func parseSSHDConfig(path string) ([]sshdSetting, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var settings []sshdSetting
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			settings = append(settings, sshdSetting{Key: line, Value: ""})
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			settings = append(settings, sshdSetting{
				Key:   parts[0],
				Value: parts[1],
			})
		} else {
			settings = append(settings, sshdSetting{Key: line, Value: ""})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return settings, nil
}

func defaultSSHDSettings() []sshdSetting {
	return []sshdSetting{
		{Key: "Port", Value: "22"},
		{Key: "AddressFamily", Value: "any"},
		{Key: "ListenAddress", Value: "0.0.0.0"},
		{Key: "Protocol", Value: "2"},
		{Key: "PermitRootLogin", Value: "prohibit-password"},
		{Key: "PubkeyAuthentication", Value: "yes"},
		{Key: "PasswordAuthentication", Value: "yes"},
		{Key: "ChallengeResponseAuthentication", Value: "no"},
		{Key: "UsePAM", Value: "yes"},
		{Key: "X11Forwarding", Value: "no"},
		{Key: "PrintMotd", Value: "no"},
		{Key: "AcceptEnv", Value: "LANG LC_*"},
		{Key: "Subsystem", Value: "sftp /usr/lib/openssh/sftp-server"},
	}
}

func validateSSHDConfig() error {
	cmd := exec.Command("sshd", "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sshd -t failed: %s", string(output))
	}
	return nil
}

func reloadSSHD() error {
	cmd := exec.Command("systemctl", "restart", "sshd")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl restart sshd: %w: %s", err, string(output))
	}
	return nil
}
