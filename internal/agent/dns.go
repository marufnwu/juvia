package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"juvia/internal/utils"
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
const namedConfLocal = "/etc/bind/named.conf.local"
const bindZoneIncludePrefix = "// Juvia managed zones - do not edit manually\n"
const bindZoneIncludeSuffix = "// End Juvia managed zones\n"

type ZoneData struct {
	Domain   string
	ServerIP string
	Serial   int64
}

type DNSRecordData struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Priority int    `json:"priority"`
	TTL      int    `json:"ttl"`
}

type ZoneSyncResult struct {
	ZonePath  string `json:"zone_path"`
	Serial   int64  `json:"serial"`
	Reloaded bool   `json:"reloaded"`
	BindError string `json:"bind_error,omitempty"`
}

func serialFromTime(t time.Time, prev int64) int64 {
	datePart := t.Format("20060102")
	seq := 1
	if prev > 0 {
		prevStr := strconv.FormatInt(prev, 10)
		if len(prevStr) == 10 {
			prevDate := prevStr[:8]
			if prevDate == datePart {
				seqStr := prevStr[8:]
				if p, err := strconv.Atoi(seqStr); err == nil {
					seq = p + 1
				}
			}
		}
	}
	serial, _ := strconv.ParseInt(datePart+fmt.Sprintf("%02d", seq), 10, 64)
	return serial
}

func nextSerial(currentSerial int64) int64 {
	return serialFromTime(time.Now(), currentSerial)
}

func HandleDNSVerify(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain      string `json:"domain"`
		ServerIP    string `json:"server_ip"`
		NS1Hostname string `json:"ns1_hostname"`
		NS2Hostname string `json:"ns2_hostname"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}
	if err := validateDomainName(req.Domain); err != nil {
		return nil, fmt.Errorf("invalid domain: %w", err)
	}
	if req.ServerIP == "" {
		req.ServerIP = getServerIP(ctx)
	}
	if req.NS1Hostname == "" {
		req.NS1Hostname = "ns1." + req.Domain
	}
	if req.NS2Hostname == "" {
		req.NS2Hostname = "ns2." + req.Domain
	}

	results := map[string]interface{}{
		"domain":    req.Domain,
		"server_ip": req.ServerIP,
		"records":   []map[string]interface{}{},
	}

	var records []map[string]interface{}

	aIPs := digAllRecords(req.Domain, "A")
	aOK := false
	for _, ip := range aIPs {
		if ip == req.ServerIP {
			aOK = true
			break
		}
	}
	records = append(records, map[string]interface{}{
		"type":     "A",
		"name":     "@",
		"expected": req.ServerIP,
		"actual":   strings.Join(aIPs, ","),
		"ok":       aOK,
	})

	wwwIPs := digAllRecords("www."+req.Domain, "A")
	wwwCNAMEs := digAllRecords("www."+req.Domain, "CNAME")
	wwwOK := false
	for _, ip := range wwwIPs {
		if ip == req.ServerIP {
			wwwOK = true
			break
		}
	}
	if !wwwOK {
		for _, cname := range wwwCNAMEs {
			if cname == req.Domain {
				wwwOK = true
				break
			}
		}
	}
	records = append(records, map[string]interface{}{
		"type":     "CNAME",
		"name":     "www",
		"expected": req.Domain,
		"actual":   joinFirstNonEmpty(wwwCNAMEs, wwwIPs),
		"ok":       wwwOK,
	})

	mxRecords := digAllRecords(req.Domain, "MX")
	mxOK := len(mxRecords) > 0
	records = append(records, map[string]interface{}{
		"type":     "MX",
		"name":     "@",
		"expected": "mail." + req.Domain,
		"actual":   strings.Join(mxRecords, ","),
		"ok":       mxOK,
	})

	txtRecords := digAllRecords(req.Domain, "TXT")
	spfOK := false
	for _, txt := range txtRecords {
		for _, part := range splitTXTRecords(txt) {
			if strings.Contains(strings.ToLower(part), "spf") {
				spfOK = true
				break
			}
		}
		if spfOK {
			break
		}
	}
	records = append(records, map[string]interface{}{
		"type":     "TXT",
		"name":     "@",
		"expected": "v=spf1 ...",
		"actual":   strings.Join(txtRecords, "; "),
		"ok":       spfOK,
	})

	nsRecords := digAllRecords(req.Domain, "NS")
	nsOK := false
	for _, ns := range nsRecords {
		if strings.Contains(ns, req.NS1Hostname) || strings.Contains(ns, req.NS2Hostname) {
			nsOK = true
			break
		}
	}
	records = append(records, map[string]interface{}{
		"type":     "NS",
		"name":     "@",
		"expected": req.NS1Hostname + " / " + req.NS2Hostname,
		"actual":   strings.Join(nsRecords, ", "),
		"ok":       nsOK,
	})

	results["records"] = records

	allOK := true
	for _, r := range records {
		if !r["ok"].(bool) {
			allOK = false
			break
		}
	}
	results["all_ok"] = allOK

	return results, nil
}

func joinFirstNonEmpty(ss, fallback []string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	for _, s := range fallback {
		if s != "" {
			return s
		}
	}
	return ""
}

func HandleNameserverHealth(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		ServerIP    string `json:"server_ip"`
		NS1Hostname string `json:"ns1_hostname"`
		NS2Hostname string `json:"ns2_hostname"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.ServerIP == "" {
		req.ServerIP = getServerIP(ctx)
	}

	bindActive := isBindActive()
	results := map[string]interface{}{
		"bind_active": bindActive,
		"server_ip":   req.ServerIP,
		"nameservers": []map[string]interface{}{},
	}

	var nsResults []map[string]interface{}

	for _, ns := range []struct {
		hostname string
		label    string
	}{
		{req.NS1Hostname, "ns1"},
		{req.NS2Hostname, "ns2"},
	} {
		if ns.hostname == "" {
			continue
		}
		nsIPs := digAllRecords(ns.hostname, "A")
		resolvesToThisServer := false
		for _, ip := range nsIPs {
			if ip == req.ServerIP {
				resolvesToThisServer = true
				break
			}
		}
		nsResults = append(nsResults, map[string]interface{}{
			"label":              ns.label,
			"hostname":           ns.hostname,
			"a_record":           strings.Join(nsIPs, ", "),
			"resolves_to_server": resolvesToThisServer,
			"ok":                 len(nsIPs) > 0 && resolvesToThisServer,
		})
	}

	results["nameservers"] = nsResults

	allNSOK := true
	for _, ns := range nsResults {
		if !ns["ok"].(bool) {
			allNSOK = false
			break
		}
	}
	results["all_ok"] = bindActive && allNSOK

	return results, nil
}

func digAllRecords(domain, rtype string) []string {
	cmd := exec.Command("dig", "+short", domain, rtype)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	result := strings.TrimSpace(string(output))
	if result == "" {
		return nil
	}
	var records []string
	for _, line := range strings.Split(result, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			records = append(records, line)
		}
	}
	return records
}

func splitTXTRecords(s string) []string {
	var results []string
	inQuote := false
	current := ""
	for _, ch := range s {
		if ch == '"' {
			if inQuote {
				results = append(results, current)
				current = ""
			}
			inQuote = !inQuote
		} else if inQuote {
			current += string(ch)
		}
	}
	if current != "" {
		results = append(results, current)
	}
	if len(results) == 0 && s != "" {
		results = append(results, s)
	}
	return results
}

func validateDomainName(domain string) error {
	if len(domain) > 253 {
		return fmt.Errorf("domain name too long (max 253 characters)")
	}
	if len(domain) == 0 {
		return fmt.Errorf("domain name is empty")
	}
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return fmt.Errorf("domain must have at least two labels (e.g. example.com)")
	}
	for _, label := range labels {
		if len(label) == 0 {
			return fmt.Errorf("domain contains empty label")
		}
		if len(label) > 63 {
			return fmt.Errorf("domain label too long (max 63 characters): %s", label)
		}
		for _, ch := range label {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '*') {
				return fmt.Errorf("domain contains invalid character '%c' in label '%s'", ch, label)
			}
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("domain label cannot start or end with hyphen: %s", label)
		}
	}
	return nil
}

func isBindActive() bool {
	serviceName := resolveBindServiceName()
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "active"
}

func resolveBindServiceName() string {
	out, _ := exec.Command("systemctl", "list-unit-files", "--type=service", "--no-pager").CombinedOutput()
	s := string(out)
	if strings.Contains(s, "bind9.service") {
		return "bind9"
	}
	return "named"
}

func sanitizeForZone(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\t", " ")
	return s
}

func HandleZoneSyncRecords(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain   string           `json:"domain"`
		ServerIP string           `json:"server_ip"`
		Serial   int64            `json:"serial"`
		Records  []DNSRecordData  `json:"records"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}
	if strings.Contains(req.Domain, "..") || strings.Contains(req.Domain, "/") {
		return nil, fmt.Errorf("invalid domain")
	}
	if req.ServerIP == "" {
		req.ServerIP = getServerIP(ctx)
	}
	if req.Serial == 0 {
		req.Serial = nextSerial(0)
	}

	zoneContent := generateZoneFromRecords(req.Domain, req.ServerIP, req.Serial, req.Records)

	if err := os.MkdirAll(zoneDir, 0755); err != nil {
		return nil, fmt.Errorf("create zone dir: %w", err)
	}

	safeZonePath, err := utils.SafePath(zoneDir, req.Domain+".zone")
	if err != nil {
		return nil, fmt.Errorf("invalid zone path: %w", err)
	}

	if err := backupNamedConf(); err != nil {
		return nil, fmt.Errorf("backup named.conf.local: %w", err)
	}

	tmpPath := safeZonePath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(zoneContent), 0600); err != nil {
		return nil, fmt.Errorf("write zone tmp: %w", err)
	}

	if err := validateZone(req.Domain, tmpPath); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("invalid zone: %w", err)
	}

	if err := os.Rename(tmpPath, safeZonePath); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("atomic rename: %w", err)
	}

	if err := addZoneToNamedConf(req.Domain); err != nil {
		return nil, fmt.Errorf("update named.conf.local: %w", err)
	}

	if err := validateNamedConf(); err != nil {
		return nil, fmt.Errorf("named.conf.local invalid: %w", err)
	}

	reloaded := true
	reloadErrStr := ""
	if err := reloadBind(); err != nil {
		reloaded = false
		reloadErrStr = err.Error()
	}

	return ZoneSyncResult{
		ZonePath:  safeZonePath,
		Serial:   req.Serial,
		Reloaded: reloaded,
		BindError: reloadErrStr,
	}, nil
}

func defaultTTL(ttl int) int {
	if ttl <= 0 {
		return 3600
	}
	return ttl
}

func generateZoneFromRecords(domain, serverIP string, serial int64, records []DNSRecordData) string {
	var buf bytes.Buffer

	buf.WriteString("$TTL 3600\n")
	buf.WriteString(fmt.Sprintf("@ IN    SOA    ns1.%s.    admin.%s. (\n", domain, domain))
	buf.WriteString(fmt.Sprintf("              %d    ; Serial\n", serial))
	buf.WriteString("              3600        ; Refresh\n")
	buf.WriteString("              1800        ; Retry\n")
	buf.WriteString("              604800      ; Expire\n")
	buf.WriteString("              86400 )     ; Minimum TTL\n\n")

	hasNS := false
	hasA := false
	hasMX := false
	hasSPF := false
	hasDMARC := false
	hasCAA := false
	hasWWW := false

	for _, r := range records {
		sName := sanitizeForZone(r.Name)
		sValue := sanitizeForZone(r.Value)
		ttl := defaultTTL(r.TTL)
		switch r.Type {
		case "NS":
			hasNS = true
			name := sName
			if name == "" || name == "@" {
				name = "@"
			}
			buf.WriteString(fmt.Sprintf("%-12s %-5d IN    NS     %s\n", name, ttl, sValue))
		case "A":
			if sName == "@" || sName == "" {
				hasA = true
				buf.WriteString(fmt.Sprintf("%-12s %-5d IN    A      %s\n", "@", ttl, sValue))
			} else if sName == "www" {
				hasWWW = true
				buf.WriteString(fmt.Sprintf("%-12s %-5d IN    A      %s\n", "www", ttl, sValue))
			} else {
				buf.WriteString(fmt.Sprintf("%-12s %-5d IN    A      %s\n", sName, ttl, sValue))
			}
		case "AAAA":
			buf.WriteString(fmt.Sprintf("%-12s %-5d IN    AAAA   %s\n", sName, ttl, sValue))
		case "CNAME":
			if sName == "www" {
				hasWWW = true
			}
			buf.WriteString(fmt.Sprintf("%-12s %-5d IN    CNAME  %s\n", sName, ttl, sValue))
		case "MX":
			hasMX = true
			prio := r.Priority
			if prio == 0 {
				prio = 10
			}
			buf.WriteString(fmt.Sprintf("%-12s %-5d IN    MX     %d    %s\n", "@", ttl, prio, sValue))
		case "TXT":
			if strings.Contains(strings.ToLower(sValue), "spf") {
				hasSPF = true
			}
			if strings.Contains(strings.ToLower(sValue), "dmarc") {
				hasDMARC = true
			}
			buf.WriteString(fmt.Sprintf("%-12s %-5d IN    TXT    \"%s\"\n", sName, ttl, sValue))
		case "CAA":
			hasCAA = true
			buf.WriteString(fmt.Sprintf("%-12s %-5d IN    CAA    %s\n", sName, ttl, sValue))
		case "SRV":
			buf.WriteString(fmt.Sprintf("%-12s %-5d IN    SRV    %s\n", sName, ttl, sValue))
		}
	}

	if !hasNS {
		buf.WriteString(fmt.Sprintf("%-12s %-5d IN    NS     ns1.%s.\n", "@", 3600, domain))
		buf.WriteString(fmt.Sprintf("%-12s %-5d IN    NS     ns2.%s.\n", "@", 3600, domain))
	}

	if !hasA {
		buf.WriteString(fmt.Sprintf("%-12s %-5d IN    A      %s\n", "@", 3600, serverIP))
	}

	if !hasWWW {
		buf.WriteString(fmt.Sprintf("%-12s %-5d IN    CNAME  %s\n", "www", 3600, "@"))
	}

	if !hasMX {
		buf.WriteString(fmt.Sprintf("%-12s %-5d IN    MX     10    mail.%s.\n", "@", 3600, domain))
	}

	if !hasSPF {
		buf.WriteString(fmt.Sprintf("%-12s %-5d IN    TXT    \"v=spf1 mx ~all\"\n", "@", 3600))
	}

	if !hasDMARC {
		buf.WriteString(fmt.Sprintf("%-12s %-5d IN    TXT    \"v=DMARC1; p=quarantine; rua=mailto:dmarc@%s\"\n", "_dmarc", 3600, domain))
	}

	if !hasCAA {
		buf.WriteString(fmt.Sprintf("%-12s %-5d IN    CAA    0 issue \"letsencrypt.org\"\n", "@", 3600))
	}

	return buf.String()
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
	if strings.Contains(req.Domain, "..") || strings.Contains(req.Domain, "/") {
		return nil, fmt.Errorf("invalid domain")
	}
	if req.ServerIP == "" {
		req.ServerIP = getServerIP(ctx)
	}

	serial := nextSerial(0)
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

	safeZonePath, err := utils.SafePath(zoneDir, req.Domain+".zone")
	if err != nil {
		return nil, fmt.Errorf("invalid zone path: %w", err)
	}

	if err := backupNamedConf(); err != nil {
		return nil, fmt.Errorf("backup named.conf.local: %w", err)
	}

	tmpPath := safeZonePath + ".tmp"
	if err := os.WriteFile(tmpPath, buf.Bytes(), 0600); err != nil {
		return nil, fmt.Errorf("write zone tmp: %w", err)
	}

	if err := validateZone(req.Domain, tmpPath); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("invalid zone: %w", err)
	}

	if err := os.Rename(tmpPath, safeZonePath); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("atomic rename: %w", err)
	}

	if err := addZoneToNamedConf(req.Domain); err != nil {
		return nil, fmt.Errorf("update named.conf.local: %w", err)
	}

	if err := validateNamedConf(); err != nil {
		return nil, fmt.Errorf("named.conf.local invalid: %w", err)
	}

	reloaded := true
	reloadErrStr := ""
	if err := reloadBind(); err != nil {
		reloaded = false
		reloadErrStr = err.Error()
	}

	return ZoneSyncResult{
		ZonePath:  safeZonePath,
		Serial:   serial,
		Reloaded: reloaded,
		BindError: reloadErrStr,
	}, nil
}

// HandleBrandDNSSetup creates or updates the brand domain zone with NS records
// for the nameservers and A records pointing to the server IP. This is called
// when the user saves nameserver settings.
func HandleBrandDNSSetup(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		BrandDomain string `json:"brand_domain"`
		NS1Hostname string `json:"ns1_hostname"`
		NS2Hostname string `json:"ns2_hostname"`
		ServerIP    string `json:"server_ip"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.BrandDomain == "" {
		return nil, fmt.Errorf("brand_domain is required")
	}
	if err := validateDomainName(req.BrandDomain); err != nil {
		return nil, fmt.Errorf("invalid brand_domain: %w", err)
	}
	if req.ServerIP == "" {
		req.ServerIP = getServerIP(ctx)
	}

	// Default nameserver hostnames if not provided
	if req.NS1Hostname == "" {
		req.NS1Hostname = "ns1." + req.BrandDomain
	}
	if req.NS2Hostname == "" {
		req.NS2Hostname = "ns2." + req.BrandDomain
	}

	serial := nextSerial(0)

	// Build zone content with NS records and glue A records for ns1/ns2
	var buf bytes.Buffer
	buf.WriteString("$TTL 3600\n")
	buf.WriteString(fmt.Sprintf("@ IN    SOA    %s.    admin.%s. (\n", req.NS1Hostname, req.BrandDomain))
	buf.WriteString(fmt.Sprintf("              %d    ; Serial\n", serial))
	buf.WriteString("              3600        ; Refresh\n")
	buf.WriteString("              1800        ; Retry\n")
	buf.WriteString("              604800      ; Expire\n")
	buf.WriteString("              86400 )     ; Minimum TTL\n\n")

	// NS records for the brand domain
	buf.WriteString(fmt.Sprintf("@ IN    NS     %s.\n", req.NS1Hostname))
	buf.WriteString(fmt.Sprintf("@ IN    NS     %s.\n\n", req.NS2Hostname))

	// Glue A records - these are critical for the nameservers to resolve
	buf.WriteString(fmt.Sprintf("%s. IN    A      %s\n", req.NS1Hostname, req.ServerIP))
	buf.WriteString(fmt.Sprintf("%s. IN    A      %s\n", req.NS2Hostname, req.ServerIP))

	zoneContent := buf.String()

	if err := os.MkdirAll(zoneDir, 0755); err != nil {
		return nil, fmt.Errorf("create zone dir: %w", err)
	}

	safeZonePath, err := utils.SafePath(zoneDir, req.BrandDomain+".zone")
	if err != nil {
		return nil, fmt.Errorf("invalid zone path: %w", err)
	}

	if err := backupNamedConf(); err != nil {
		return nil, fmt.Errorf("backup named.conf.local: %w", err)
	}

	tmpPath := safeZonePath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(zoneContent), 0600); err != nil {
		return nil, fmt.Errorf("write zone tmp: %w", err)
	}

	if err := validateZone(req.BrandDomain, tmpPath); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("invalid zone: %w", err)
	}

	if err := os.Rename(tmpPath, safeZonePath); err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("atomic rename: %w", err)
	}

	if err := addZoneToNamedConf(req.BrandDomain); err != nil {
		return nil, fmt.Errorf("update named.conf.local: %w", err)
	}

	if err := validateNamedConf(); err != nil {
		return nil, fmt.Errorf("named.conf.local invalid: %w", err)
	}

	reloaded := true
	reloadErrStr := ""
	if err := reloadBind(); err != nil {
		reloaded = false
		reloadErrStr = err.Error()
	}

	return ZoneSyncResult{
		ZonePath:  safeZonePath,
		Serial:   serial,
		Reloaded: reloaded,
		BindError: reloadErrStr,
	}, nil
}

func HandleZoneUpdate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Domain   string           `json:"domain"`
		ServerIP string           `json:"server_ip"`
		Serial   int64            `json:"serial"`
		Records  []DNSRecordData  `json:"records"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if len(req.Records) > 0 {
		return HandleZoneSyncRecords(ctx, params)
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
	if strings.Contains(req.Domain, "..") || strings.Contains(req.Domain, "/") {
		return nil, fmt.Errorf("invalid domain")
	}

	safeZonePath, err := utils.SafePath(zoneDir, req.Domain+".zone")
	if err != nil {
		return nil, fmt.Errorf("invalid zone path: %w", err)
	}

	if _, err := os.Stat(safeZonePath); os.IsNotExist(err) {
		_ = removeZoneFromNamedConf(req.Domain)
		return map[string]interface{}{"deleted": true}, nil
	}
	if err := os.Remove(safeZonePath); err != nil {
		return nil, fmt.Errorf("remove zone: %w", err)
	}

	if err := removeZoneFromNamedConf(req.Domain); err != nil {
		return nil, fmt.Errorf("update named.conf.local: %w", err)
	}

	if err := validateNamedConf(); err != nil {
		return nil, fmt.Errorf("named.conf.local invalid: %w", err)
	}

	if err := reloadBind(); err != nil {
		return nil, fmt.Errorf("reload bind: %w", err)
	}

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

func reloadBind() error {
	serviceName := resolveBindServiceName()
	cmd := exec.Command("systemctl", "reload", serviceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("reload %s failed: %w: %s", serviceName, err, string(output))
	}
	if !isBindActive() {
		return fmt.Errorf("%s is not active after reload", serviceName)
	}
	return nil
}

func getServerIP(ctx context.Context) string {
	ip := discoverPublicIP(ctx)
	if ip == "" {
		return "127.0.0.1"
	}
	return ip
}

func backupNamedConf() error {
	if _, err := os.Stat(namedConfLocal); os.IsNotExist(err) {
		return nil
	}
	backupPath := namedConfLocal + ".juvia-backup"
	if _, err := os.Stat(backupPath); err == nil {
		return nil
	}
	cmd := exec.Command("cp", namedConfLocal, backupPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("backup named.conf.local: %w", err)
	}
	return nil
}

func addZoneToNamedConf(domain string) error {
	if err := ensureNamedConfManagedBlock(); err != nil {
		return err
	}

	includeLine := fmt.Sprintf(`zone "%s" { type master; file "%s"; };`+"\n", domain, filepath.Join(zoneDir, domain+".zone"))

	lines, err := readNamedConfLines()
	if err != nil {
		return err
	}

	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf(`zone "%s"`, domain)) {
			return nil
		}
	}

	var newLines []string
	inBlock := false
	inserted := false
	for _, line := range lines {
		if strings.Contains(line, bindZoneIncludePrefix) {
			inBlock = true
		}
		newLines = append(newLines, line)
		if inBlock && !inserted && strings.Contains(line, bindZoneIncludePrefix) {
			newLines = append(newLines, includeLine)
			inserted = true
		}
		if strings.Contains(line, bindZoneIncludeSuffix) {
			inBlock = false
		}
	}

	if !inserted {
		newLines = append([]string{bindZoneIncludePrefix, includeLine, bindZoneIncludeSuffix}, newLines...)
	}

	return writeNamedConfLines(newLines)
}

func removeZoneFromNamedConf(domain string) error {
	lines, err := readNamedConfLines()
	if err != nil {
		return err
	}

	var newLines []string
	for _, line := range lines {
		if !strings.Contains(line, fmt.Sprintf(`zone "%s"`, domain)) {
			newLines = append(newLines, line)
		}
	}

	return writeNamedConfLines(newLines)
}

func ensureNamedConfManagedBlock() error {
	lines, err := readNamedConfLines()
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		lines = []string{}
	}

	hasPrefix := false
	hasSuffix := false
	for _, line := range lines {
		if strings.Contains(line, bindZoneIncludePrefix) {
			hasPrefix = true
		}
		if strings.Contains(line, bindZoneIncludeSuffix) {
			hasSuffix = true
		}
	}

	if hasPrefix && hasSuffix {
		return nil
	}

	if !hasPrefix {
		lines = append(lines, bindZoneIncludePrefix)
	}
	if !hasSuffix {
		lines = append(lines, bindZoneIncludeSuffix)
	}

	return writeNamedConfLines(lines)
}

func readNamedConfLines() ([]string, error) {
	f, err := os.Open(namedConfLocal)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text()+"\n")
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func writeNamedConfLines(lines []string) error {
	tmpPath := namedConfLocal + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create named.conf.tmp: %w", err)
	}
	for _, line := range lines {
		if _, err := f.WriteString(line); err != nil {
			f.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("write named.conf.tmp: %w", err)
		}
	}
	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close named.conf.tmp: %w", err)
	}

	if err := os.Rename(tmpPath, namedConfLocal); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename named.conf: %w", err)
	}

	return nil
}

func validateNamedConf() error {
	cmd := exec.Command("named-checkconf")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("named-checkconf: %w: %s", err, string(output))
	}
	return nil
}
