package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var zoneTemplate = `$TTL 3600
@ IN    SOA    ns1.{{.Domain}}.    admin.{{.Domain}}. (
              {{.Serial}}    ; Serial
              3600        ; Refresh
              1800        ; Retry
              604800      ; Expire
              86400 )     ; Minimum TTL

@    IN    NS     ns1.{{.Domain}}.
@    IN    NS     ns2.{{.Domain}}.
@    IN    A      {{.ServerIP}}
www IN    CNAME  @
mail IN    A      {{.ServerIP}}
@    IN    MX     10    mail.{{.Domain}}.
@    IN    TXT    "v=spf1 mx ~all"
_dmarc   IN    TXT    "v=DMARC1; p=quarantine; rua=mailto:dmarc@{{.Domain}}"
`

const zoneDir = "/etc/juvia/bind/zones"

type ZoneData struct {
	Domain   string
	ServerIP string
	Serial   int64
}

func HandleZoneCreate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain   string `json:"domain"`
		ServerIP string `json:"server_ip"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}
	if req.ServerIP == "" {
		req.ServerIP = getServerIP()
	}

	serial := time.Now().Unix() / 86400
	data := ZoneData{Domain: req.Domain, ServerIP: req.ServerIP, Serial: serial}

	t, err := template.New("zone").Parse(zoneTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	if err := os.MkdirAll(zoneDir, 0755); err != nil {
		return nil, fmt.Errorf("create zone dir: %w", err)
	}

	tmpPath := filepath.Join(zoneDir, req.Domain+".zone.tmp")
	if err := os.WriteFile(tmpPath, buf.Bytes(), 0644); err != nil {
		return nil, fmt.Errorf("write zone tmp: %w", err)
	}

	if err := validateZone(req.Domain, tmpPath); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("invalid zone: %w", err)
	}

	if err := os.Rename(tmpPath, filepath.Join(zoneDir, req.Domain+".zone")); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("atomic rename: %w", err)
	}

	reloadBind()

	return map[string]interface{}{
		"zone_path": filepath.Join(zoneDir, req.Domain+".zone"),
	}, nil
}

func HandleZoneUpdate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain   string `json:"domain"`
		ServerIP string `json:"server_ip"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	return HandleZoneCreate(ctx, params)
}

func HandleZoneDelete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	path := filepath.Join(zoneDir, req.Domain+".zone")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return map[string]interface{}{"deleted": true}, nil
	}
	if err := os.Remove(path); err != nil {
		return nil, fmt.Errorf("remove zone: %w", err)
	}

	reloadBind()

	return map[string]interface{}{"deleted": true}, nil
}

func validateZone(domain, path string) error {
	cmd := exec.Command("named-checkzone", domain, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("named-checkzone: %w: %s", err, string(output))
	}
	return nil
}

func reloadBind() {
	cmd := exec.Command("systemctl", "reload", "named")
	cmd.Run()
}

func getServerIP() string {
	cmd := exec.Command("curl", "-s", "ifconfig.me")
	output, _ := cmd.Output()
	ip := strings.TrimSpace(string(output))
	if ip == "" || strings.Contains(ip, "error") {
		return "127.0.0.1"
	}
	return ip
}
