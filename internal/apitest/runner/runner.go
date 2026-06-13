package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type APIClient struct {
	BaseURL      string
	Username     string
	Password     string
	Token        string
	HTTPClient   *http.Client
	Cookies      []*http.Cookie
	Registry     *ResourceRegistry
	Timestamp    string
	CreatedDomain string
}

type Result struct {
	Method         string `json:"method"`
	Path           string `json:"path"`
	ResolvedPath   string `json:"resolved_path,omitempty"`
	Status         int    `json:"status"`
	DurationMS     int64  `json:"duration_ms"`
	Success        bool   `json:"success"`
	Body           string `json:"body"`
	ErrorMessage   string `json:"error_message,omitempty"`
	AgentDependent bool   `json:"agent_dependent,omitempty"`
	Phase          string `json:"phase,omitempty"`
	Category       string `json:"category,omitempty"`
}

func NewAPIClient(baseURL, username, password string) *APIClient {
	return &APIClient{
		BaseURL:    strings.TrimSuffix(baseURL, "/"),
		Username:   username,
		Password:   password,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Cookies:    make([]*http.Cookie, 0),
		Registry:   NewResourceRegistry(),
		Timestamp:  fmt.Sprintf("%d", time.Now().Unix()),
	}
}

func (c *APIClient) Login() error {
	body := map[string]string{
		"username": c.Username,
		"password": c.Password,
	}
	resp, err := c.DoRequest(http.MethodPost, "/api/v1/auth/login", body, nil, false)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	if resp.Status != 200 {
		return fmt.Errorf("login failed with status %d: %s", resp.Status, resp.Body)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Body), &result); err != nil {
		return fmt.Errorf("parse login response: %w", err)
	}
	if data, ok := result["data"].(map[string]interface{}); ok {
		if token, ok := data["access_token"].(string); ok {
			c.Token = token
		}
	}

	return nil
}

func (c *APIClient) DoRequest(method, path string, body interface{}, queryParams map[string]string, authenticated bool) (*Result, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	fullURL := c.BaseURL + path
	if len(queryParams) > 0 {
		q := url.Values{}
		for k, v := range queryParams {
			if v == "" && c.CreatedDomain != "" {
				v = c.CreatedDomain
			}
			q.Set(k, v)
		}
		fullURL = fullURL + "?" + q.Encode()
	}
	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if authenticated && c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	if len(c.Cookies) > 0 {
		req.Header.Set("Cookie", c.cookiesToString())
	}

	start := time.Now()
	resp, err := c.HTTPClient.Do(req)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return &Result{
			Method:       method,
			Path:         path,
			DurationMS:   duration,
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		c.Cookies = append(c.Cookies, cookie)
	}

	respBody, _ := io.ReadAll(resp.Body)
	bodyStr := string(respBody)
	if len(bodyStr) > 2000 {
		bodyStr = bodyStr[:2000] + "... [truncated]"
	}

	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return &Result{
		Method:     method,
		Path:       path,
		Status:     resp.StatusCode,
		DurationMS: duration,
		Success:    success,
		Body:       bodyStr,
	}, nil
}

func (c *APIClient) cookiesToString() string {
	var parts []string
	for _, cookie := range c.Cookies {
		parts = append(parts, cookie.Name+"="+cookie.Value)
	}
	return strings.Join(parts, "; ")
}

func (c *APIClient) CheckSetupStatus() (bool, error) {
	result, err := c.DoRequest(http.MethodGet, "/api/v1/setup/status", nil, nil, false)
	if err != nil {
		return false, err
	}
	if result.Status != 200 {
		return false, fmt.Errorf("setup status check failed: %s", result.Body)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result.Body), &data); err != nil {
		return false, err
	}
	if d, ok := data["data"].(map[string]interface{}); ok {
		if setupRequired, ok := d["setup_required"].(bool); ok {
			return setupRequired, nil
		}
	}
	return false, nil
}

func (c *APIClient) FirstRunSetup(username, password, email, serverName string) error {
	body := map[string]string{
		"username":    username,
		"password":    password,
		"email":       email,
		"server_name": serverName,
	}
	result, err := c.DoRequest(http.MethodPost, "/api/v1/setup/first-run", body, nil, false)
	if err != nil {
		return err
	}
	if result.Status != 200 {
		return fmt.Errorf("first run setup failed: %s", result.Body)
	}
	return nil
}

func (c *APIClient) HealthCheck() (*Result, error) {
	return c.DoRequest(http.MethodGet, "/api/v1/health", nil, nil, false)
}

func (c *APIClient) extractID(body string) int64 {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return 0
	}
	if d, ok := data["data"].(map[string]interface{}); ok {
		if id, ok := d["id"].(float64); ok {
			return int64(id)
		}
		if id, ok := d["ID"].(float64); ok {
			return int64(id)
		}
		if id, ok := d["website_id"].(float64); ok {
			return int64(id)
		}
		if id, ok := d["task_id"].(float64); ok {
			return int64(id)
		}
	}
	return 0
}

func (c *APIClient) extractDomain(body string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return ""
	}
	if d, ok := data["data"].(map[string]interface{}); ok {
		if domain, ok := d["domain"].(string); ok {
			return domain
		}
	}
	return ""
}

func (c *APIClient) ResolvePath(path string) string {
	result := path

	// Determine which resource :id refers to based on path pattern
	var id int64
	switch {
	case strings.Contains(path, "/dns/records/"):
		id = c.Registry.GetDNSRecord()
	case strings.Contains(path, "/databases/"):
		id = c.Registry.GetDatabase()
	case strings.Contains(path, "/email/mailboxes/"):
		id = c.Registry.GetMailbox()
	case strings.Contains(path, "/firewall/rules/"):
		id = c.Registry.GetFirewallRule()
	case strings.Contains(path, "/cron/"):
		id = c.Registry.GetCronJob()
	case strings.Contains(path, "/backups/"):
		id = c.Registry.GetBackupSchedule()
	default:
		id = c.Registry.GetWebsite()
	}
	result = strings.ReplaceAll(result, ":id", strconv.FormatInt(id, 10))
	result = strings.ReplaceAll(result, ":user_id", strconv.FormatInt(c.Registry.GetDatabaseUser(c.Registry.GetDatabase()), 10))
	result = strings.ReplaceAll(result, ":table", "test_table")
	result = strings.ReplaceAll(result, ":app_id", "1")
	result = strings.ReplaceAll(result, ":name", "nginx")

	return result
}

func (c *APIClient) getCategory(path string) string {
	if strings.Contains(path, "/auth/") {
		return "auth"
	}
	if strings.Contains(path, "/users") {
		return "users"
	}
	if strings.Contains(path, "/websites") {
		return "websites"
	}
	if strings.Contains(path, "/dns/") {
		return "dns"
	}
	if strings.Contains(path, "/databases") {
		return "databases"
	}
	if strings.Contains(path, "/email/") {
		return "email"
	}
	if strings.Contains(path, "/firewall") {
		return "firewall"
	}
	if strings.Contains(path, "/backups") {
		return "backups"
	}
	if strings.Contains(path, "/cron") {
		return "cron"
	}
	if strings.Contains(path, "/logs") {
		return "logs"
	}
	if strings.Contains(path, "/alerts") {
		return "alerts"
	}
	if strings.Contains(path, "/metrics") {
		return "metrics"
	}
	if strings.Contains(path, "/apps") {
		return "apps"
	}
	if strings.Contains(path, "/git") {
		return "git"
	}
	if strings.Contains(path, "/webmail") {
		return "webmail"
	}
	if strings.Contains(path, "/terminal") {
		return "terminal"
	}
	if strings.Contains(path, "/updates") {
		return "updates"
	}
	if strings.Contains(path, "/settings") {
		return "settings"
	}
	if strings.Contains(path, "/audit-log") {
		return "audit"
	}
	if strings.Contains(path, "/services") {
		return "services"
	}
	if strings.Contains(path, "/setup") {
		return "setup"
	}
	if strings.Contains(path, "/version") {
		return "version"
	}
	return "misc"
}

func (c *APIClient) Run(destructive bool) ([]Result, error) {
	var results []Result

	for _, ep := range Endpoints {
		if ep.AgentDependent && !destructive && strings.HasPrefix(ep.Path, "/api/v1/updates") {
			results = append(results, Result{
				Method:         ep.Method,
				Path:           ep.Path,
				Success:        true,
				DurationMS:     0,
				Phase:          string(ep.Phase),
				Category:       c.getCategory(ep.Path),
				ErrorMessage:   "skipped (non-destructive mode)",
			})
			continue
		}

		resolvedPath := c.ResolvePath(ep.Path)
		body := c.getPayload(ep.Path, resolvedPath)

		authRequired := ep.Auth != AuthNone
		result, err := c.DoRequest(ep.Method, resolvedPath, body, ep.QueryParams, authRequired)
		if err != nil {
			results = append(results, Result{
				Method:       ep.Method,
				Path:         ep.Path,
				Success:      false,
				ErrorMessage: err.Error(),
				Phase:        string(ep.Phase),
				Category:     c.getCategory(ep.Path),
			})
			continue
		}

		result.ResolvedPath = resolvedPath
		result.AgentDependent = ep.AgentDependent
		result.Phase = string(ep.Phase)
		result.Category = c.getCategory(ep.Path)

		if result.Success && ep.ResourceType != "" && (ep.Phase == PhaseAnytime || ep.Phase == PhaseAfterCreate) {
			id := c.extractID(result.Body)
			if id > 0 {
				switch ep.ResourceType {
				case "website":
					c.Registry.AddWebsite(id)
					if c.CreatedDomain == "" {
						c.CreatedDomain = c.extractDomain(result.Body)
					}
				case "database":
					c.Registry.AddDatabase(id)
				case "mailbox":
					c.Registry.AddMailbox(id)
				case "alias":
					c.Registry.AddAlias(id)
				case "forwarder":
					c.Registry.AddForwarder(id)
				case "firewall":
					c.Registry.AddFirewallRule(id)
				case "cron":
					c.Registry.AddCronJob(id)
				case "backup-schedule":
					c.Registry.AddBackupSchedule(id)
				case "dns-record":
					c.Registry.AddDNSRecord(id)
				case "backup":
				case "database-user":
					c.Registry.AddDatabaseUser(c.Registry.GetDatabase(), id)
				}
			}
		}

		results = append(results, *result)
	}

	return results, nil
}

func (c *APIClient) getPayload(path, resolvedPath string) interface{} {
	if path == "/api/v1/auth/login" {
		return map[string]string{"username": c.Username, "password": c.Password}
	}

	if body, ok := CreatePayloads[path]; ok {
		return c.injectUniqueValues(body, path)
	}
	if body, ok := CreatePayloads[resolvedPath]; ok {
		return c.injectUniqueValues(body, path)
	}
	return nil
}

func (c *APIClient) injectUniqueValues(body interface{}, path string) interface{} {
	if body == nil {
		return nil
	}

	switch v := body.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			switch str := val.(type) {
			case string:
				if k == "domain" {
					result[k] = c.Timestamp + "-" + str
				} else if k == "email" {
					result[k] = c.Timestamp + "-" + str
				} else if k == "source" && (strings.Contains(str, "@") || strings.Contains(str, "example.com")) {
					result[k] = c.Timestamp + "-" + str
				} else if k == "destination" && strings.Contains(str, "@") {
					result[k] = c.Timestamp + "-" + str
				} else if k == "username" {
					result[k] = c.Timestamp + str
				} else {
					result[k] = str
				}
			default:
				result[k] = val
			}
		}
		if isBackupCreate(path) {
			if websiteID := c.Registry.GetWebsite(); websiteID > 0 {
				result["website_id"] = websiteID
			}
		}
		return result
	}
	return body
}

func isBackupCreate(path string) bool {
	return path == "/api/v1/backups"
}

func (c *APIClient) Cleanup() []Result {
	var results []Result

	cleanupEndpoints := []struct {
		Method       string
		Path         string
		ResourceType string
	}{
		{http.MethodDelete, "/api/v1/cron/:id", "cron"},
		{http.MethodDelete, "/api/v1/backup-schedules/:id", "backup-schedule"},
		{http.MethodDelete, "/api/v1/firewall/rules/:id", "firewall"},
		{http.MethodDelete, "/api/v1/email/aliases/:id", "alias"},
		{http.MethodDelete, "/api/v1/email/forwarders/:id", "forwarder"},
		{http.MethodDelete, "/api/v1/email/mailboxes/:id", "mailbox"},
		{http.MethodDelete, "/api/v1/databases/:id/users/:user_id", "database-user"},
		{http.MethodDelete, "/api/v1/databases/:id", "database"},
		{http.MethodDelete, "/api/v1/websites/:id", "website"},
	}

	for _, ep := range cleanupEndpoints {
		id := c.getIDForResourceType(ep.ResourceType)
		if id == 0 {
			continue
		}

		path := ep.Path
		path = strings.ReplaceAll(path, ":id", strconv.FormatInt(id, 10))
		if ep.ResourceType == "database-user" {
			path = strings.ReplaceAll(path, ":user_id", strconv.FormatInt(c.Registry.GetDatabaseUser(c.Registry.GetDatabase()), 10))
		}

		result, err := c.DoRequest(ep.Method, path, nil, nil, true)
		if err != nil {
			result = &Result{
				Method:       ep.Method,
				Path:         ep.Path,
				Success:      false,
				ErrorMessage: err.Error(),
				Phase:        "cleanup",
			}
		}
		result.Phase = "cleanup"
		result.Category = c.getCategory(ep.Path)
		results = append(results, *result)
	}

	return results
}

func (c *APIClient) getIDForResourceType(rt string) int64 {
	switch rt {
	case "website":
		return c.Registry.GetWebsite()
	case "database":
		return c.Registry.GetDatabase()
	case "mailbox":
		return c.Registry.GetMailbox()
	case "alias":
		return c.Registry.GetAlias()
	case "forwarder":
		return c.Registry.GetForwarder()
	case "firewall":
		return c.Registry.GetFirewallRule()
	case "cron":
		return c.Registry.GetCronJob()
	case "backup-schedule":
		return c.Registry.GetBackupSchedule()
	}
	return 0
}
