package nginx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

func TestGenerateSiteConfig(t *testing.T) {
	cfg := SiteConfig{
		Domain:       "example.com",
		DocumentRoot: "/home/example/public_html",
		PHPVersion:   "8.2",
		LinuxUser:    "example",
	}

	content, err := generateSiteConfig(cfg)
	if err != nil {
		t.Fatalf("generate site config: %v", err)
	}

	if !strings.Contains(content, "example.com") {
		t.Errorf("expected domain in config")
	}
	if !strings.Contains(content, "/home/example/public_html") {
		t.Errorf("expected document root in config")
	}
	if !strings.Contains(content, "8.2") {
		t.Errorf("expected PHP version in config")
	}
	if !strings.Contains(content, "example") {
		t.Errorf("expected linux user in config")
	}
}

func TestGenerateSiteConfigWithSSL(t *testing.T) {
	cfg := SiteConfig{
		Domain:       "example.com",
		DocumentRoot: "/home/example/public_html",
		PHPVersion:   "8.2",
		LinuxUser:    "example",
		SSLEnabled:   true,
		SSLCertPath:  "/etc/ssl/certs/example.crt",
		SSLKeyPath:   "/etc/ssl/private/example.key",
	}

	content, err := generateSiteConfig(cfg)
	if err != nil {
		t.Fatalf("generate site config: %v", err)
	}

	if !strings.Contains(content, "listen 443 ssl") {
		t.Errorf("expected SSL listen directive")
	}
	if !strings.Contains(content, "ssl_certificate") {
		t.Errorf("expected SSL certificate directive")
	}
}

func TestGeneratePoolConfig(t *testing.T) {
	cfg := PoolConfig{
		Domain:       "example.com",
		PHPVersion:   "8.2",
		LinuxUser:    "example",
		DocumentRoot: "/home/example/public_html",
	}

	tmpl := poolTemplate
	tpl, err := template.New("pool").Parse(tmpl)
	if err != nil {
		t.Fatalf("parse pool template: %v", err)
	}

	var buf strings.Builder
	if err := tpl.Execute(&buf, cfg); err != nil {
		t.Fatalf("execute pool template: %v", err)
	}

	content := buf.String()
	if !strings.Contains(content, "example.com") {
		t.Errorf("expected domain in pool config")
	}
	if !strings.Contains(content, "8.2") {
		t.Errorf("expected PHP version in pool config")
	}
	if !strings.Contains(content, "example") {
		t.Errorf("expected linux user in pool config")
	}
}

func TestWriteSiteConfigAtomicCreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	testConfigDir := filepath.Join(tmpDir, "sites")
	os.MkdirAll(testConfigDir, 0755)

	content := `server {
    listen 80;
    server_name example.com;
}`

	tmpPath := filepath.Join(testConfigDir, "example.com.conf.tmp")
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}

	actualPath := filepath.Join(testConfigDir, "example.com.conf")
	if err := os.Rename(tmpPath, actualPath); err != nil {
		t.Fatalf("atomic rename: %v", err)
	}

	data, _ := os.ReadFile(actualPath)
	if string(data) != content {
		t.Errorf("expected config content to match")
	}
}

func TestRemoveSiteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	testConfigDir := filepath.Join(tmpDir, "sites")
	os.MkdirAll(testConfigDir, 0755)

	path := filepath.Join(testConfigDir, "example.com.conf")
	os.WriteFile(path, []byte("test"), 0644)

	os.Remove(path)

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected config file to be removed")
	}
}

func TestRemoveSiteConfigNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	testConfigDir := filepath.Join(tmpDir, "sites")
	os.MkdirAll(testConfigDir, 0755)

	os.Remove(filepath.Join(testConfigDir, "nonexistent.com.conf"))
}

func TestGenerateSuspendedConfig(t *testing.T) {
	cfg := map[string]string{
		"Domain":       "example.com",
		"DocumentRoot": "/home/example/public_html",
	}

	tmpl := suspendedTemplate
	tpl, err := template.New("suspended").Parse(tmpl)
	if err != nil {
		t.Fatalf("parse suspended template: %v", err)
	}

	var buf strings.Builder
	if err := tpl.Execute(&buf, cfg); err != nil {
		t.Fatalf("execute suspended template: %v", err)
	}

	content := buf.String()
	if !strings.Contains(content, "example.com") {
		t.Errorf("expected domain in suspended config")
	}
	if !strings.Contains(content, "return 503") {
		t.Errorf("expected503 status in suspended config")
	}
}
