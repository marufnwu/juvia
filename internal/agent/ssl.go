package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func HandleSSLIssue(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	cmd := exec.Command("certbot", "certonly", "--nginx", "-d", req.Domain, "-d", "www."+req.Domain,
		"--non-interactive", "--agree-tos", "-m", "admin@"+req.Domain, "--keep")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("certbot issue: %w: %s", err, string(output))
	}

	certPath := "/etc/letsencrypt/live/" + req.Domain + "/fullchain.pem"
	keyPath := "/etc/letsencrypt/live/" + req.Domain + "/privkey.pem"
	expiry := getCertExpiry(certPath)

	return map[string]interface{}{
		"cert_path": certPath,
		"key_path":  keyPath,
		"expiry": expiry,
	}, nil
}

func HandleSSLRenew(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	cmd := exec.Command("certbot", "renew", "--cert-name", req.Domain, "--non-interactive")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("certbot renew: %w: %s", err, string(output))
	}

	certPath := "/etc/letsencrypt/live/" + req.Domain + "/fullchain.pem"
	expiry := getCertExpiry(certPath)

	return map[string]interface{}{
		"expiry": expiry,
	}, nil
}

func HandleSSLCheck(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	cmd := exec.Command("certbot", "certificates", "--cert-name", req.Domain)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{"ready": false, "message": "no certificate found"}, nil
	}

	return map[string]interface{}{
		"ready":   true,
		"message": "certificate found",
	}, nil
}

func HandleSSLRemove(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	cmd := exec.Command("certbot", "delete", "--cert-name", req.Domain, "--non-interactive")
	cmd.Run()

	return map[string]interface{}{"removed": true}, nil
}

func getCertExpiry(certPath string) time.Time {
	cmd := exec.Command("openssl", "x509", "-enddate", "-noout", "-in", certPath)
	output, _ := cmd.Output()
	dateStr := strings.ReplaceAll(string(output), "notAfter=", "")
	dateStr = strings.TrimSpace(dateStr)
	t, _ := time.Parse("Jan 2 15:04:05 2006 MST", dateStr)
	if t.IsZero() {
		return time.Now().Add(90 * 24 * time.Hour)
	}
	return t
}
