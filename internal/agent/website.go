package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"juvia/internal/agent/nginx"
)

func HandleWebsiteCreate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain     string `json:"domain"`
		PHPVersion string `json:"php_version"`
		WebServer  string `json:"web_server"`
		TaskID     string `json:"task_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	linuxUser := SanitizeLinuxUser(req.Domain)
	homeDir := "/home/" + linuxUser
	publicHTML := filepath.Join(homeDir, "public_html")
	logsDir := filepath.Join(homeDir, "logs")
	sslDir := filepath.Join(homeDir, "ssl")
	tmpDir := filepath.Join(homeDir, "tmp")

	if err := createLinuxUser(linuxUser, homeDir); err != nil {
		return nil, fmt.Errorf("create linux user: %w", err)
	}

	if err := os.MkdirAll(homeDir, 0755); err != nil {
		return nil, fmt.Errorf("create home dir: %w", err)
	}
	if err := os.MkdirAll(publicHTML, 0755); err != nil {
		return nil, fmt.Errorf("create public_html: %w", err)
	}
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("create logs: %w", err)
	}
	if err := os.MkdirAll(sslDir, 0750); err != nil {
		return nil, fmt.Errorf("create ssl: %w", err)
	}
	if err := os.MkdirAll(tmpDir, 0770); err != nil {
		return nil, fmt.Errorf("create tmp: %w", err)
	}

	if err := os.Chown(homeDir, getUID(linuxUser), getGID(linuxUser)); err != nil {
		return nil, fmt.Errorf("chown home: %w", err)
	}
	if err := os.Chown(publicHTML, getUID(linuxUser), getGID(linuxUser)); err != nil {
		return nil, fmt.Errorf("chown public_html: %w", err)
	}
	if err := os.Chown(logsDir, getUID(linuxUser), getGID(linuxUser)); err != nil {
		return nil, fmt.Errorf("chown logs: %w", err)
	}
	if err := os.Chown(tmpDir, getUID(linuxUser), getGID(linuxUser)); err != nil {
		return nil, fmt.Errorf("chown tmp: %w", err)
	}

	if req.WebServer == "" {
		req.WebServer = "nginx"
	}
	if req.PHPVersion == "" {
		req.PHPVersion = "8.2"
	}

	siteCfg := nginx.SiteConfig{
		Domain:       req.Domain,
		DocumentRoot: publicHTML,
		PHPVersion:   req.PHPVersion,
		LinuxUser:    linuxUser,
	}

	if err := nginx.GenerateAndWriteSiteConfig(siteCfg); err != nil {
		return nil, fmt.Errorf("generate nginx config: %w", err)
	}

	poolCfg := nginx.PoolConfig{
		Domain:       req.Domain,
		PHPVersion:   req.PHPVersion,
		LinuxUser:    linuxUser,
		DocumentRoot: publicHTML,
	}
	if err := nginx.GenerateAndWritePoolConfig(poolCfg); err != nil {
		return nil, fmt.Errorf("generate php-fpm pool: %w", err)
	}

	if err := nginx.ReloadNginx(); err != nil {
		return nil, fmt.Errorf("reload nginx: %w", err)
	}

	return map[string]interface{}{
		"document_root": publicHTML,
		"linux_user":    linuxUser,
	}, nil
}

func HandleWebsiteDelete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain       string `json:"domain"`
		DocumentRoot string `json:"document_root"`
		Force        bool   `json:"force"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	linuxUser := SanitizeLinuxUser(req.Domain)

	nginx.RemoveSiteConfig(req.Domain)
	nginx.RemovePoolConfig(req.Domain)
	nginx.ReloadNginx()

	if req.Force {
		if err := deleteLinuxUser(linuxUser); err != nil {
		}
		if req.DocumentRoot != "" {
			os.RemoveAll(filepath.Dir(req.DocumentRoot))
		}
	}

	return map[string]interface{}{"deleted": true}, nil
}

func HandleWebsiteSuspend(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain       string `json:"domain"`
		DocumentRoot string `json:"document_root"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	nginx.WriteSuspendedConfig(req.Domain, req.DocumentRoot)
	nginx.StopPHPFPM(req.Domain)
	nginx.ReloadNginx()

	return map[string]interface{}{"suspended": true}, nil
}

func SanitizeLinuxUser(domain string) string {
	user := strings.ReplaceAll(domain, ".", "_")
	user = strings.ReplaceAll(user, "-", "_")
	if len(user) > 32 {
		user = user[:32]
	}
	return user
}

func createLinuxUser(username, homeDir string) error {
	cmd := exec.Command("useradd", "-r", "-s", "/usr/sbin/nologin", "-d", homeDir, "-M", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(output), "already exists") {
			return nil
		}
		return err
	}
	return nil
}

func deleteLinuxUser(username string) error {
	cmd := exec.Command("userdel", "-r", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(output), "does not exist") {
			return nil
		}
		return err
	}
	return nil
}

func getUID(username string) int {
	cmd := exec.Command("id", "-u", username)
	output, _ := cmd.Output()
	var uid int
	fmt.Sscanf(string(output), "%d", &uid)
	if uid == 0 {
		return 1000
	}
	return uid
}

func getGID(username string) int {
	cmd := exec.Command("id", "-g", username)
	output, _ := cmd.Output()
	var gid int
	fmt.Sscanf(string(output), "%d", &gid)
	if gid == 0 {
		return 1000
	}
	return gid
}
