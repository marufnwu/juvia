package apps

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type appInstallParams struct {
	WebsiteID     int64  `json:"website_id"`
	WebsitePath   string `json:"website_path"`
	AppType       string `json:"app_type"`
	AdminUsername string `json:"admin_username"`
	AdminPassword string `json:"admin_password"`
	DBName        string `json:"db_name"`
	DBPassword    string `json:"db_password"`
}

type appUpdateParams struct {
	WebsiteID   int64  `json:"website_id"`
	WebsitePath string `json:"website_path"`
	AppType    string `json:"app_type"`
}

func HandleAppInstall(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req appInstallParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.WebsitePath == "" {
		return nil, fmt.Errorf("website path is required")
	}

	switch req.AppType {
	case "wordpress":
		return installWordPress(req)
	case "laravel":
		return installLaravel(req)
	default:
		return nil, fmt.Errorf("unsupported app type: %s", req.AppType)
	}
}

func installWordPress(req appInstallParams) (interface{}, error) {
	if req.WebsitePath == "" {
		return nil, fmt.Errorf("website path is required")
	}

	os.MkdirAll(req.WebsitePath, 0755)

	latest := "https://wordpress.org/latest.tar.gz"
	destPath := filepath.Join(os.TempDir(), "wordpress.tar.gz")

	if err := downloadFile(latest, destPath); err != nil {
		return nil, fmt.Errorf("download wordpress: %w", err)
	}

	cmd := exec.Command("tar", "-xzf", destPath, "-C", req.WebsitePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("extract wordpress: %w: %s", err, string(output))
	}

	wpFiles := filepath.Join(req.WebsitePath, "wordpress")
	files, _ := os.ReadDir(wpFiles)
	for _, f := range files {
		os.Rename(filepath.Join(wpFiles, f.Name()), filepath.Join(req.WebsitePath, f.Name()))
	}
	os.RemoveAll(wpFiles)
	os.Remove(destPath)

	wpConfigPath := filepath.Join(req.WebsitePath, "wp-config.php")
	wpConfigSample, _ := os.ReadFile(filepath.Join(req.WebsitePath, "wp-config-sample.php"))

	dbName := req.DBName
	if dbName == "" {
		dbName = "wp_" + randomString(8)
	}
	dbUser := "juvia"
	dbPass := req.DBPassword
	if dbPass == "" {
		dbPass = randomString(16)
	}

	wpConfig := strings.Replace(string(wpConfigSample),
		"define( 'DB_NAME', 'database_name_here' );",
		fmt.Sprintf("define( 'DB_NAME', '%s' );", dbName), 1)
	wpConfig = strings.Replace(wpConfig,
		"define( 'DB_USER', 'username_here' );",
		fmt.Sprintf("define( 'DB_USER', '%s' );", dbUser), 1)
	wpConfig = strings.Replace(wpConfig,
		"define( 'DB_PASSWORD', 'password_here' );",
		fmt.Sprintf("define( 'DB_PASSWORD', '%s' );", dbPass), 1)
	wpConfig = strings.Replace(wpConfig,
		"define( 'DB_HOST', 'localhost' );",
		"define( 'DB_HOST', 'localhost' );", 1)

	wpConfig += "\ndefine('WP_DEBUG', false);\n"
	wpConfig += "define('FS_METHOD', 'direct');\n"

	os.WriteFile(wpConfigPath, []byte(wpConfig), 0644)

	cmd = exec.Command("chown", "-R", "www-data:www-data", req.WebsitePath)
	cmd.Run()

	return map[string]interface{}{
		"version": "latest",
		"name":    "WordPress",
	}, nil
}

func installLaravel(req appInstallParams) (interface{}, error) {
	if req.WebsitePath == "" {
		return nil, fmt.Errorf("website path is required")
	}

	os.MkdirAll(req.WebsitePath, 0755)

	composerCmd := exec.Command("composer", "create-project", "--prefer-dist", "laravel/laravel", req.WebsitePath)
	composerCmd.Env = append(os.Environ(), "COMPOSER_NO_INTERACTION=1")
	if output, err := composerCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("composer create-project: %w: %s", err, string(output))
	}

	envPath := filepath.Join(req.WebsitePath, ".env")
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		envSamplePath := filepath.Join(req.WebsitePath, ".env.example")
		if data, err := os.ReadFile(envSamplePath); err == nil {
			os.WriteFile(envPath, data, 0644)
		}
	}

	cmd := exec.Command("chown", "-R", "www-data:www-data", req.WebsitePath)
	cmd.Run()

	return map[string]interface{}{
		"version": "latest",
		"name":    "Laravel",
	}, nil
}

func HandleAppUpdate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req appUpdateParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	switch req.AppType {
	case "wordpress":
		return updateWordPress(req)
	case "laravel":
		return updateLaravel(req)
	default:
		return nil, fmt.Errorf("unsupported app type: %s", req.AppType)
	}
}

func updateWordPress(req appUpdateParams) (interface{}, error) {
	cmd := exec.Command("wp", "core", "update", "--allow-root")
	cmd.Dir = req.WebsitePath
	cmd.Run()

	return map[string]interface{}{
		"version": "latest",
		"updated": true,
	}, nil
}

func updateLaravel(req appUpdateParams) (interface{}, error) {
	cmd := exec.Command("composer", "update", "--no-interaction", "--prefer-dist")
	cmd.Dir = req.WebsitePath
	cmd.Env = append(os.Environ(), "COMPOSER_NO_INTERACTION=1")
	cmd.Run()

	return map[string]interface{}{
		"version": "latest",
		"updated": true,
	}, nil
}

func HandleAppCheckUpdate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		WebsiteID       int64  `json:"website_id"`
		WebsitePath    string `json:"website_path"`
		AppType        string `json:"app_type"`
		CurrentVersion string `json:"current_version"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	updateAvailable := false
	var latestVersion string

	switch req.AppType {
	case "wordpress":
		latestVersion = getLatestWPVersion()
		updateAvailable = latestVersion != "" && latestVersion != req.CurrentVersion
	case "laravel":
		latestVersion = getLatestLaravelVersion()
		updateAvailable = latestVersion != "" && latestVersion != req.CurrentVersion
	}

	return map[string]interface{}{
		"update_available": updateAvailable,
		"latest_version":  latestVersion,
	}, nil
}

func getLatestWPVersion() string {
	resp, err := http.Get("https://api.wordpress.org/core/version-check/1.7/")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	parts := strings.Split(string(data), "\n")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func getLatestLaravelVersion() string {
	resp, err := http.Get("https://packagist.org/packages/laravel/laravel.json")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(data), `"latest":"`) {
		start := strings.Index(string(data), `"latest":"`) + 9
		end := strings.Index(string(data)[start:], `"`)
		if end > 0 {
			return string(data)[start : start+end]
		}
	}
	return ""
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func randomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}