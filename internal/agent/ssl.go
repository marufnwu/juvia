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

func HandleSSLIssueDomain(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain      string `json:"domain"`
		DNSProvider string `json:"dns_provider"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	isWildcard := strings.HasPrefix(req.Domain, "*.")
	certName := strings.ReplaceAll(req.Domain, "*.", "wildcard-")
	certName = strings.ReplaceAll(certName, ".", "-")

	certDir := "/etc/juvia/ssl/" + certName
	os.MkdirAll(certDir, 0750)

	certPath := filepath.Join(certDir, "fullchain.pem")
	keyPath := filepath.Join(certDir, "privkey.pem")

	if isWildcard || req.DNSProvider != "" {
		args := []string{
			"certonly",
			"--dns-" + req.DNSProvider,
			"-d", req.Domain,
			"--non-interactive",
			"--agree-tos",
			"-m", "admin@" + strings.TrimPrefix(req.Domain, "*."),
			"--cert-name", certName,
			"--keep",
		}
		output, err := exec.CommandContext(ctx, "certbot", args...).CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("dns-01 challenge failed: %w: %s", err, string(output))
		}
		livePath := "/etc/letsencrypt/live/" + certName
		leCert := filepath.Join(livePath, "fullchain.pem")
		leKey := filepath.Join(livePath, "privkey.pem")
		if _, err := os.Stat(leCert); err == nil {
			expiry := getCertExpiry(leCert)
			return map[string]interface{}{
				"cert_type":  "letsencrypt",
				"cert_path":  leCert,
				"key_path":   leKey,
				"expiry":     expiry,
				"cert_name":  certName,
			}, nil
		}
	}

	args := []string{
		"certonly", "--nginx",
		"-d", req.Domain,
	}
	if !isWildcard {
		args = append(args, "-d", "www."+strings.TrimPrefix(req.Domain, "www."))
	}
	args = append(args,
		"--non-interactive", "--agree-tos",
		"-m", "admin@"+strings.TrimPrefix(req.Domain, "*."),
		"--cert-name", certName, "--keep",
	)

	_, err := exec.CommandContext(ctx, "certbot", args...).CombinedOutput()
	if err == nil {
		livePath := "/etc/letsencrypt/live/" + certName
		leCert := filepath.Join(livePath, "fullchain.pem")
		leKey := filepath.Join(livePath, "privkey.pem")
		expiry := getCertExpiry(leCert)
		return map[string]interface{}{
			"cert_type": "letsencrypt",
			"cert_path": leCert,
			"key_path":  leKey,
			"expiry":    expiry,
			"cert_name": certName,
		}, nil
	}

	opensslCmd := exec.Command("openssl", "req", "-x509", "-nodes", "-days", "90",
		"-newkey", "rsa:2048",
		"-keyout", keyPath,
		"-out", certPath,
		"-subj", "/CN="+strings.TrimPrefix(req.Domain, "*."),
		"-addext", "subjectAltName=DNS:"+req.Domain)
	if _, err := opensslCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("self-signed cert failed: %w", err)
	}

	expiry := time.Now().Add(90 * 24 * time.Hour)
	return map[string]interface{}{
		"cert_type": "self_signed",
		"cert_path": certPath,
		"key_path":  keyPath,
		"expiry":    expiry,
		"cert_name": certName,
		"warning":   "Could not obtain a Let's Encrypt certificate; generated a self-signed certificate for testing.",
	}, nil
}

func HandleSSLRenewDomain(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		CertName string `json:"cert_name"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.CertName == "" {
		return nil, fmt.Errorf("cert_name is required")
	}

	cmd := exec.Command("certbot", "renew", "--cert-name", req.CertName, "--non-interactive")
	output, err := cmd.CombinedOutput()
	if err == nil {
		livePath := "/etc/letsencrypt/live/" + req.CertName
		certPath := filepath.Join(livePath, "fullchain.pem")
		expiry := getCertExpiry(certPath)
		return map[string]interface{}{
			"expiry": expiry,
		}, nil
	}

	selfSignedPath := "/etc/juvia/ssl/" + req.CertName + "/fullchain.pem"
	if _, statErr := os.Stat(selfSignedPath); statErr == nil {
		expiry := getCertExpiry(selfSignedPath)
		return map[string]interface{}{
			"expiry":    expiry,
			"cert_type": "self_signed",
			"warning":   "Self-signed certificates cannot be renewed via certbot.",
		}, nil
	}

	return nil, fmt.Errorf("renew failed: %w: %s", err, string(output))
}
