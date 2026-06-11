package email

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

var cryptoRand = rand.Reader

const roundcubeVersion = "1.6.1"
const roundcubeURL = "https://github.com/roundcube/roundcubemail/releases/download/%s/roundcubemail-%s-complete.tar.gz"

func HandleWebmailInstall(ctx context.Context, params json.RawMessage) (interface{}, error) {
	version := roundcubeVersion
	url := fmt.Sprintf(roundcubeURL, version, version)

	destPath := filepath.Join(os.TempDir(), "roundcube.tar.gz")

	if err := downloadFile(url, destPath); err != nil {
		return nil, fmt.Errorf("download roundcube: %w", err)
	}

	installPath := "/var/www/webmail"
	os.MkdirAll(installPath, 0755)

	cmd := exec.Command("tar", "-xzf", destPath, "-C", installPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("extract roundcube: %w: %s", err, string(output))
	}

	roundcubeFiles := filepath.Join(installPath, "roundcubemail-"+version)
	files, _ := os.ReadDir(roundcubeFiles)
	for _, f := range files {
		os.Rename(filepath.Join(roundcubeFiles, f.Name()), filepath.Join(installPath, f.Name()))
	}
	os.RemoveAll(roundcubeFiles)
	os.Remove(destPath)

	dbName := "roundcube"
	dbUser := "roundcube"
	dbPass := randomString(16)

	setupSQL := fmt.Sprintf(`
CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s';
GRANT ALL PRIVILEGES ON %s.* TO '%s'@'localhost';
FLUSH PRIVILEGES;
`, dbName, dbUser, dbPass, dbName, dbUser)

	mysqlCmd := exec.Command("mysql", "-e", setupSQL)
	if err := mysqlCmd.Run(); err != nil {
		return nil, fmt.Errorf("create roundcube database: %w", err)
	}

	configPath := filepath.Join(installPath, "config/config.inc.php")
	configContent := fmt.Sprintf(`<?php
$config = array();
$config['db_dsnw'] = 'mysql://%s:%s@localhost/%s';
$config['imap_host'] = 'localhost:143';
$config['smtp_host'] = 'localhost:587';
$config['smtp_user'] = '';
$config['smtp_pass'] = '';
$config['support_url'] = '';
$config['product_name'] = 'Juvia Webmail';
`, dbUser, dbPass, dbName)

	os.WriteFile(configPath, []byte(configContent), 0644)

	cmd = exec.Command("chown", "-R", "www-data:www-data", installPath)
	cmd.Run()

	cmd = exec.Command("systemctl", "reload", "nginx")
	cmd.Run()

	return map[string]interface{}{
		"installed":  true,
		"db_name":   dbName,
		"db_user":   dbUser,
		"db_pass":    dbPass,
		"install_path": installPath,
	}, nil
}

func HandleWebmailUninstall(ctx context.Context, params json.RawMessage) (interface{}, error) {
	installPath := "/var/www/webmail"

	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("webmail not installed")
	}

	if err := os.RemoveAll(installPath); err != nil {
		return nil, fmt.Errorf("remove webmail files: %w", err)
	}

	mysqlCmd := exec.Command("mysql", "-e", "DROP DATABASE IF EXISTS roundcube; DROP USER IF EXISTS 'roundcube'@'localhost';")
	mysqlCmd.Run()

	cmd := exec.Command("systemctl", "reload", "nginx")
	cmd.Run()

	return map[string]interface{}{
		"uninstalled": true,
	}, nil
}

func HandleWebmailStatus(ctx context.Context, params json.RawMessage) (interface{}, error) {
	installPath := "/var/www/webmail"
	installed, err := func() (bool, error) {
		if _, err := os.Stat(installPath); os.IsNotExist(err) {
			return false, nil
		} else if err != nil {
			return false, err
		}
		return true, nil
	}()

	return map[string]interface{}{
		"installed": installed,
		"path":      installPath,
	}, err
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
	_, err := cryptoRand.Read(bytes)
	if err != nil {
		for i := range bytes {
			bytes[i] = byte(i % 256)
		}
	}
	hexChars := "0123456789abcdef"
	result := make([]byte, length)
	for i := range result {
		result[i] = hexChars[int(bytes[i/2])%16]
	}
	return string(result)
}