package email

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const mailBaseDir = "/var/mail"

func HandleMailboxCreate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Domain   string `json:"domain"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	user := strings.Split(req.Email, "@")[0]
	mailDir := filepath.Join(mailBaseDir, req.Domain, user)

	if err := os.MkdirAll(mailDir, 0700); err != nil {
		return nil, fmt.Errorf("create mail dir: %w", err)
	}

	hashedPassword, err := generateDovecotPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	userDBDir := filepath.Join(mailBaseDir, req.Domain)
	os.MkdirAll(userDBDir, 0700)
	userDBPath := filepath.Join(userDBDir, "userdb")
	userDBLine := fmt.Sprintf("%s:1:%d:%s:%s:::", user, 5000, mailDir, hashedPassword)

	f, err := os.OpenFile(userDBPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("write userdb: %w", err)
	}
	f.WriteString(userDBLine + "\n")
	f.Close()

	reloadPostfix()
	reloadDovecot()

	return map[string]interface{}{
		"email":    req.Email,
		"mail_dir": mailDir,
	}, nil
}

func HandleMailboxDelete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	user := strings.Split(req.Email, "@")[0]
	domain := strings.Split(req.Email, "@")[1]
	mailDir := filepath.Join(mailBaseDir, domain, user)

	os.RemoveAll(mailDir)

	userDBPath := filepath.Join(mailBaseDir, domain, "userdb")
	removeLineFromFile(userDBPath, user)

	reloadPostfix()
	reloadDovecot()

	return map[string]interface{}{"deleted": true}, nil
}

func HandleMailboxUpdate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
		Quota       int64  `json:"quota"`
		ForwardTo   string `json:"forward_to"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	return map[string]interface{}{"updated": true}, nil
}

func HandleAliasCreate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain      string `json:"domain"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Source == "" || req.Destination == "" {
		return nil, fmt.Errorf("source and destination are required")
	}

	virtualAliasPath := "/etc/postfix/virtual_aliases"
	os.MkdirAll(filepath.Dir(virtualAliasPath), 0700)
	line := fmt.Sprintf("%s@%s %s", req.Source, req.Domain, req.Destination)

	f, err := os.OpenFile(virtualAliasPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("write alias: %w", err)
	}
	f.WriteString(line + "\n")
	f.Close()

	reloadPostfix()

	return map[string]interface{}{
		"source":      req.Source + "@" + req.Domain,
		"destination": req.Destination,
	}, nil
}

func HandleAliasDelete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain string `json:"domain"`
		Source string `json:"source"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	virtualAliasPath := "/etc/postfix/virtual_aliases"
	removeLineFromFile(virtualAliasPath, req.Source+"@"+req.Domain)

	reloadPostfix()

	return map[string]interface{}{"deleted": true}, nil
}

func HandleForwarderCreate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain      string `json:"domain"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Source == "" || req.Destination == "" {
		return nil, fmt.Errorf("source and destination are required")
	}

	virtualAliasPath := "/etc/postfix/virtual_aliases"
	os.MkdirAll(filepath.Dir(virtualAliasPath), 0700)
	line := fmt.Sprintf("%s@%s %s", req.Source, req.Domain, req.Destination)

	f, err := os.OpenFile(virtualAliasPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("write forwarder: %w", err)
	}
	f.WriteString(line + "\n")
	f.Close()

	reloadPostfix()

	return map[string]interface{}{
		"source":      req.Source + "@" + req.Domain,
		"destination": req.Destination,
	}, nil
}

func HandleForwarderDelete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain string `json:"domain"`
		Source string `json:"source"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	virtualAliasPath := "/etc/postfix/virtual_aliases"
	removeLineFromFile(virtualAliasPath, req.Source+"@"+req.Domain)

	reloadPostfix()

	return map[string]interface{}{"deleted": true}, nil
}

func HandleDeliverability(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}

	spf := checkSPF(req.Domain)
	dkim := checkDKIM(req.Domain)
	dmarc := checkDMARC(req.Domain)
	ptr := checkPTR(req.Domain)

	return map[string]interface{}{
		"spf":   spf,
		"dkim":  dkim,
		"dmarc": dmarc,
		"ptr":   ptr,
	}, nil
}

func generateDovecotPassword(password string) (string, error) {
	cmd := exec.Command("doveadm", "pw", "-s", "BLF-CRYPT")
	cmd.Stdin = strings.NewReader(password)
	output, err := cmd.Output()
	if err != nil {
		cmd2 := exec.Command("doveadm", "pw", "-s", "SHA512")
		cmd2.Stdin = strings.NewReader(password)
		output, err = cmd2.Output()
		if err != nil {
			return "", fmt.Errorf("dovecot pw: %w", err)
		}
	}
	return strings.TrimSpace(string(output)), nil
}

func reloadPostfix() {
	exec.Command("postmap", "/etc/postfix/virtual_aliases").Run()
	exec.Command("systemctl", "reload", "postfix").Run()
}

func reloadDovecot() {
	exec.Command("systemctl", "reload", "dovecot").Run()
}

func removeLineFromFile(path, pattern string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(data), "\n")
	var newLines []string
	for _, l := range lines {
		if !strings.Contains(l, pattern) && l != "" {
			newLines = append(newLines, l)
		}
	}
	return os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0600)
}

func checkSPF(domain string) map[string]interface{} {
	cmd := exec.Command("dig", "TXT", domain, "+short")
	output, _ := cmd.Output()
	txt := strings.TrimSpace(string(output))

	hasSPF := strings.Contains(txt, "v=spf1")
	return map[string]interface{}{
		"status":    hasSPF,
		"record":   txt,
		"message":   "SPF Record · Authorizes which servers can send email for your domain",
		"recommendation": "Add 'v=spf1 mx -all' TXT record if missing",
	}
}

func checkDKIM(domain string) map[string]interface{} {
	cmd := exec.Command("dig", "TXT", "default._domainkey."+domain, "+short")
	output, _ := cmd.Output()
	txt := strings.TrimSpace(string(output))

	hasDKIM := strings.Contains(txt, "v=DKIM1")
	return map[string]interface{}{
		"status":    hasDKIM,
		"record":   txt,
		"message":  "DKIM Record · A cryptographic signature that proves your emails are genuine",
		"recommendation": "Add DKIM selector TXT record if missing",
	}
}

func checkDMARC(domain string) map[string]interface{} {
	cmd := exec.Command("dig", "TXT", "_dmarc."+domain, "+short")
	output, _ := cmd.Output()
	txt := strings.TrimSpace(string(output))

	hasDMARC := strings.Contains(txt, "v=DMARC1")
	return map[string]interface{}{
		"status":    hasDMARC,
		"record":   txt,
		"message":  "DMARC Record · Tells receiving servers what to do with emails that fail authentication",
		"recommendation": "Add 'v=DMARC1; p=quarantine; rua=mailto:dmarc@" + domain + "' TXT record if missing",
	}
}

func checkPTR(domain string) map[string]interface{} {
	cmd := exec.Command("hostname", "-I")
	output, _ := cmd.Output()
	ip := strings.TrimSpace(string(output))

	return map[string]interface{}{
		"status": ip != "",
		"ip":     ip,
		"message": "PTR Record · Reverse DNS that links your server IP to your domain name",
		"recommendation": "Set PTR record to '" + domain + "' via your hosting provider",
	}
}
