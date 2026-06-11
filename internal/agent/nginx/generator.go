package nginx

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type SiteConfig struct {
	Domain       string
	DocumentRoot string
	PHPVersion   string
	LinuxUser    string
	SSLEnabled   bool
	SSLCertPath  string
	SSLKeyPath   string
}

type PoolConfig struct {
	Domain       string
	PHPVersion   string
	LinuxUser    string
	DocumentRoot string
}

var siteTemplate = `server {
    listen 80;
    server_name {{.Domain}} www.{{.Domain}};
    root {{.DocumentRoot}};
    index index.php index.html;

    access_log /home/{{.LinuxUser}}/logs/access.log;
    error_log /home/{{.LinuxUser}}/logs/error.log;

    client_body_temp_path /home/{{.LinuxUser}}/tmp;
    proxy_temp_path /home/{{.LinuxUser}}/tmp;
    fastcgi_temp_path /home/{{.LinuxUser}}/tmp;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/run/php/php{{.PHPVersion}}-fpm-{{.LinuxUser}}.sock;
    }

    location ~ /\.ht {
        deny all;
    }
}
`

var sslSiteTemplate = `server {
    listen 80;
    server_name {{.Domain}} www.{{.Domain}};
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name {{.Domain}} www.{{.Domain}};
    root {{.DocumentRoot}};
    index index.php index.html;

    ssl_certificate {{.SSLCertPath}};
    ssl_certificate_key {{.SSLKeyPath}};

    access_log /home/{{.LinuxUser}}/logs/access.log;
    error_log /home/{{.LinuxUser}}/logs/error.log;

    client_body_temp_path /home/{{.LinuxUser}}/tmp;
    proxy_temp_path /home/{{.LinuxUser}}/tmp;
    fastcgi_temp_path /home/{{.LinuxUser}}/tmp;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/run/php/php{{.PHPVersion}}-fpm-{{.LinuxUser}}.sock;
    }

    location ~ /\.ht {
        deny all;
    }
}
`

var suspendedTemplate = `server {
    listen 80;
    server_name {{.Domain}} www.{{.Domain}};
    root {{.DocumentRoot}};
    index index.html;

    location / {
        return 503;
    }

    location = / {
        return 503;
    }

    error_page 503 /suspended.html;
    location = /suspended.html {
        internal;
    }
}
`

var poolTemplate = `[{{.Domain}}]
user = {{.LinuxUser}}
group = {{.LinuxUser}}
listen = /run/php/php{{.PHPVersion}}-fpm-{{.LinuxUser}}.sock
listen.owner = www-data
listen.group = www-data
listen.mode = 0660
pm = dynamic
pm.max_children = 5
pm.start_servers = 2
pm.min_spare_servers = 1
pm.max_spare_servers = 3
chdir = {{.DocumentRoot}}
php_admin_value[open_basedir] = {{.DocumentRoot}}:/tmp
php_admin_value[upload_tmp_dir] = {{.DocumentRoot}}/../tmp
php_admin_value[session.save_path] = {{.DocumentRoot}}/../tmp
`

var suspendedHTML = `<!DOCTYPE html>
<html>
<head><title>Site Suspended</title></head>
<body>
<h1>This website has been suspended</h1>
<p>Please contact the server administrator for more information.</p>
</body>
</html>
`

const (
	configDir = "/etc/juvia/nginx/sites"
	poolDir    = "/etc/juvia/php-fpm/pools"
	snippetsDir = "/etc/juvia/nginx/snippets"
)

func GenerateAndWriteSiteConfig(cfg SiteConfig) error {
	content, err := generateSiteConfig(cfg)
	if err != nil {
		return err
	}
	return writeSiteConfigAtomic(cfg.Domain, content)
}

func generateSiteConfig(cfg SiteConfig) (string, error) {
	tmpl := siteTemplate
	if cfg.SSLEnabled {
		tmpl = sslSiteTemplate
	}
	t, err := template.New("site").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, cfg); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func writeSiteConfigAtomic(domain, content string) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	tmpPath := filepath.Join(configDir, domain+".conf.tmp")
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write config tmp: %w", err)
	}
	if err := ValidateConfig(tmpPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("invalid config: %w", err)
	}
	if err := os.Rename(tmpPath, filepath.Join(configDir, domain+".conf")); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("atomic rename: %w", err)
	}
	return nil
}

func GenerateAndWritePoolConfig(cfg PoolConfig) error {
	t, err := template.New("pool").Parse(poolTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, cfg); err != nil {
		return err
	}
	return writePoolConfigAtomic(cfg.Domain, buf.String())
}

func writePoolConfigAtomic(domain, content string) error {
	if err := os.MkdirAll(poolDir, 0755); err != nil {
		return fmt.Errorf("create pool dir: %w", err)
	}
	tmpPath := filepath.Join(poolDir, domain+".conf.tmp")
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write pool tmp: %w", err)
	}
	if err := ValidatePoolConfig(tmpPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("invalid pool config: %w", err)
	}
	if err := os.Rename(tmpPath, filepath.Join(poolDir, domain+".conf")); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("atomic rename: %w", err)
	}
	return nil
}

func WriteSuspendedConfig(domain, documentRoot string) error {
	t, err := template.New("suspended").Parse(suspendedTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	cfg := map[string]string{"Domain": domain, "DocumentRoot": documentRoot}
	if err := t.Execute(&buf, cfg); err != nil {
		return err
	}
	return writeSiteConfigAtomic(domain, buf.String())
}

func RemoveSiteConfig(domain string) error {
	path := filepath.Join(configDir, domain+".conf")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

func RemovePoolConfig(domain string) error {
	path := filepath.Join(poolDir, domain+".conf")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

func ValidateConfig(path string) error {
	cmd := exec.Command("nginx", "-t", "-c", "/etc/nginx/nginx.conf", "-q", "-p", filepath.Dir(path), "-t")
	cmd.Env = append(os.Environ(), "NGINX_CONF_STRING="+path)
	output, _ := cmd.CombinedOutput()
	if len(output) > 0 && strings.Contains(string(output), "error") {
		return fmt.Errorf("nginx validation failed: %s", string(output))
	}
	return nil
}

func ValidatePoolConfig(path string) error {
	cmd := exec.Command("php-fpm", "-t", "-y", path)
	output, _ := cmd.CombinedOutput()
	if len(output) > 0 && strings.Contains(string(output), "ERROR") {
		return fmt.Errorf("php-fpm validation failed: %s", string(output))
	}
	return nil
}

func ReloadNginx() error {
	cmd := exec.Command("systemctl", "reload", "nginx")
	if output, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(output), "No such file or directory") || strings.Contains(string(output), "not-found") {
			cmd = exec.Command("nginx", "-s", "reload")
			output, err = cmd.CombinedOutput()
		}
		if err != nil {
			return fmt.Errorf("reload nginx: %w", err)
		}
	}
	return nil
}

func StopPHPFPM(domain string) error {
	socketPath := fmt.Sprintf("/run/php/php-fpm-%s.sock", strings.ReplaceAll(domain, ".", "_"))
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil
	}
	cmd := exec.Command("systemctl", "stop", "php"+strings.ReplaceAll(domain, ".", "-")+"-fpm")
	cmd.Run()
	return nil
}

func WriteSuspendedHTML(documentRoot string) error {
	path := filepath.Join(documentRoot, "suspended.html")
	return os.WriteFile(path, []byte(suspendedHTML), 0644)
}
