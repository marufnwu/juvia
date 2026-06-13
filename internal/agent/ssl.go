package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	if err := cmd.Run(); err == nil {
		certPath := "/etc/letsencrypt/live/" + req.Domain + "/fullchain.pem"
		keyPath := "/etc/letsencrypt/live/" + req.Domain + "/privkey.pem"
		expiry := getCertExpiry(certPath)

		return map[string]interface{}{
			"cert_type": "letsencrypt",
			"cert_path": certPath,
			"key_path":  keyPath,
			"expiry":    expiry,
		}, nil
	}

	certDir := "/etc/juvia/ssl/self-signed/" + req.Domain
	os.MkdirAll(certDir, 0750)

	certPath := filepath.Join(certDir, "fullchain.pem")
	keyPath := filepath.Join(certDir, "privkey.pem")

	opensslCmd := exec.Command("openssl", "req", "-x509", "-nodes", "-days", "90",
		"-newkey", "rsa:2048",
		"-keyout", keyPath,
		"-out", certPath,
		"-subj", "/CN="+req.Domain,
		"-addext", "subjectAltName=DNS:"+req.Domain+",DNS:www."+req.Domain)
	opensslOutput, err := opensslCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("self-signed cert generation failed: %w: %s", err, string(opensslOutput))
	}

	expiry := time.Now().Add(90 * 24 * time.Hour)

	return map[string]interface{}{
		"cert_type": "self_signed",
		"cert_path": certPath,
		"key_path":  keyPath,
		"expiry":    expiry,
		"warning":   "Domain could not be validated by Let's Encrypt; a self-signed certificate was generated for testing/internal use.",
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
	if err == nil {
		certPath := "/etc/letsencrypt/live/" + req.Domain + "/fullchain.pem"
		expiry := getCertExpiry(certPath)
		return map[string]interface{}{
			"expiry": expiry,
		}, nil
	}

	selfSignedPath := "/etc/juvia/ssl/self-signed/" + req.Domain + "/fullchain.pem"
	if _, statErr := os.Stat(selfSignedPath); statErr == nil {
		expiry := getCertExpiry(selfSignedPath)
		return map[string]interface{}{
			"expiry":      expiry,
			"cert_type":   "self_signed",
			"warning":     "Self-signed certificates cannot be renewed by certbot; returning existing expiry.",
		}, nil
	}

	return nil, fmt.Errorf("certbot renew: %w: %s", err, string(output))
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

	exec.Command("certbot", "delete", "--cert-name", req.Domain, "--non-interactive").Run()

	selfSignedDir := "/etc/juvia/ssl/self-signed/" + req.Domain
	os.RemoveAll(selfSignedDir)

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
