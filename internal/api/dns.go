package api

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func listDNSRecordsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		userIDVal, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userIDVal.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		zone, err := cfg.DB.GetZoneByWebsite(c.Request.Context(), id)
		if err != nil || zone == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    []interface{}{},
				"zone":    nil,
			})
			return
		}

		records, err := cfg.DB.ListDNSRecordsByZone(c.Request.Context(), zone.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    records,
			"zone":    zone,
		})
	}
}

func getWebsiteZoneHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		userIDVal, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userIDVal.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		zone, err := cfg.DB.GetZoneByWebsite(c.Request.Context(), id)
		if err != nil || zone == nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "DNS zone not found for this website"))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    zone,
		})
	}
}

func validateDNSRecordValue(recordType, value, name string) error {
	switch recordType {
	case "A":
		if net.ParseIP(value) == nil || net.ParseIP(value).To4() == nil {
			return fmt.Errorf("invalid IPv4 address: %s", value)
		}
	case "AAAA":
		if net.ParseIP(value) == nil || net.ParseIP(value).To4() != nil {
			return fmt.Errorf("invalid IPv6 address: %s", value)
		}
	case "CNAME":
		if net.ParseIP(value) != nil {
			return fmt.Errorf("CNAME record cannot point to an IP address")
		}
		if !isValidRecordHostname(value) {
			return fmt.Errorf("invalid hostname for CNAME: %s", value)
		}
	case "MX":
		if !isValidRecordHostname(value) {
			return fmt.Errorf("invalid hostname for MX: %s", value)
		}
	case "NS":
		if !isValidRecordHostname(value) {
			return fmt.Errorf("invalid hostname for NS: %s", value)
		}
	case "TXT":
		if len(value) > 65535 {
			return fmt.Errorf("TXT record too long (max 65535 bytes)")
		}
	case "SRV":
		parts := strings.Fields(value)
		if len(parts) != 4 {
			return fmt.Errorf("SRV record must have format: priority weight port target")
		}
	}
	if strings.ContainsAny(value, "\n\r") {
		return fmt.Errorf("record value cannot contain newlines")
	}
	if name != "@" && name != "" && strings.ContainsAny(name, " \t\n\r") {
		return fmt.Errorf("record name cannot contain spaces")
	}
	return nil
}

func isValidRecordHostname(h string) bool {
	if net.ParseIP(h) != nil {
		return false
	}
	h = strings.TrimSuffix(h, ".")
	if len(h) == 0 || len(h) > 253 {
		return false
	}
	labels := strings.Split(h, ".")
	if len(labels) < 1 {
		return false
	}
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
	}
	return true
}

type createDNSRecordRequest struct {
	Type string `json:"type" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Value    string `json:"value" binding:"required"`
	Priority *int   `json:"priority"`
}

func createDNSRecordHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		var req createDNSRecordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		validTypes := map[string]bool{"A": true, "AAAA": true, "CNAME": true, "MX": true, "TXT": true, "NS": true, "CAA": true, "SRV": true}
		if !validTypes[req.Type] {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid record type"))
			return
		}

		if err := validateDNSRecordValue(req.Type, req.Value, req.Name); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", err.Error()))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		userIDVal, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userIDVal.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		zone, err := cfg.DB.GetZoneByWebsite(c.Request.Context(), id)
		if err != nil || zone == nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "DNS zone not found"))
			return
		}

		record, err := cfg.DB.CreateDNSRecord(c.Request.Context(), zone.ID, req.Type, req.Name, req.Value, req.Priority)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		if err := syncZoneRecords(cfg, c, website.Domain, zone.ID); err != nil {
			cfg.Log.ErrorContext(c.Request.Context(), "zone sync failed after record create: "+err.Error())
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "DNS zone sync failed: "+err.Error()))
			return
		}

		uid := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &uid, "Added "+req.Type+" record to "+website.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    record,
		})
	}
}

type updateDNSRecordRequest struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Priority *int   `json:"priority"`
	TTL      *int   `json:"ttl"`
}

func updateDNSRecordHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid record id"))
			return
		}

		var req updateDNSRecordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		record, err := cfg.DB.GetDNSRecordByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "DNS record not found"))
			return
		}

		name := req.Name
		if name == "" {
			name = record.Name
		}
		value := req.Value
		if value == "" {
			value = record.Value
		}

		if err := cfg.DB.UpdateDNSRecord(c.Request.Context(), id, name, value, req.Priority, req.TTL); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		if zone, err := cfg.DB.GetZoneByID(c.Request.Context(), record.ZoneID); err == nil && zone != nil {
			if zone.WebsiteID != nil {
				website, _ := cfg.DB.GetWebsiteByID(c.Request.Context(), *zone.WebsiteID)
				if website != nil {
					if err := syncZoneRecords(cfg, c, website.Domain, zone.ID); err != nil {
						cfg.Log.ErrorContext(c.Request.Context(), "zone sync failed after record update: "+err.Error())
					}
				}
			}
		}

		uid := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &uid, "Updated DNS record on "+record.Name+"."+func() string {
			if zone, err := cfg.DB.GetZoneByID(c.Request.Context(), record.ZoneID); err == nil && zone != nil {
				return zone.Domain
			}
			return "unknown"
		}(), c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "record updated"},
		})
	}
}

func deleteDNSRecordHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid record id"))
			return
		}

		record, err := cfg.DB.GetDNSRecordByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "DNS record not found"))
			return
		}

		zone, err := cfg.DB.GetZoneByID(c.Request.Context(), record.ZoneID)
		if err != nil || zone == nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "DNS zone not found"))
			return
		}

		if err := cfg.DB.DeleteDNSRecord(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		if zone.WebsiteID != nil {
			website, _ := cfg.DB.GetWebsiteByID(c.Request.Context(), *zone.WebsiteID)
			if website != nil {
				if err := syncZoneRecords(cfg, c, website.Domain, zone.ID); err != nil {
					cfg.Log.ErrorContext(c.Request.Context(), "zone sync failed after record delete: "+err.Error())
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "record deleted"},
		})
	}
}

func suggestDNSHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		zone, err := cfg.DB.GetZoneByWebsite(c.Request.Context(), id)
		if err != nil || zone == nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "DNS zone not found"))
			return
		}

		records, _ := cfg.DB.ListDNSRecordsByZone(c.Request.Context(), zone.ID)
		recordMap := make(map[string]bool)
		for _, r := range records {
			recordMap[r.Type+"_"+r.Name] = true
		}

		var suggestions []map[string]string

		if !recordMap["TXT__dmarc"] {
			suggestions = append(suggestions, map[string]string{
				"type":        "TXT",
				"name":        "_dmarc",
				"value":       "v=DMARC1; p=quarantine; rua=mailto:dmarc@" + zone.Domain,
				"description": "DMARC record — helps prevent email spoofing and improves deliverability",
			})
		}
		if !recordMap["CAA_@"] {
			suggestions = append(suggestions, map[string]string{
				"type":        "CAA",
				"name":        "@",
				"value":       "0 issue \"letsencrypt.org\"",
				"description": "CAA record — restricts which Certificate Authorities can issue SSL for your domain",
			})
		}
		if !recordMap["TXT@"] {
			suggestions = append(suggestions, map[string]string{
				"type":        "TXT",
				"name":        "@",
				"value":       "v=spf1 mx ~all",
				"description": "SPF record — authorizes your server to send email for your domain",
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    suggestions,
		})
	}
}

func getNameserversHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		serverIP, _ := cfg.DB.GetSetting(ctx, "server_ip")
		nsBrandDomain, _ := cfg.DB.GetSetting(ctx, "ns_brand_domain")
		ns1Hostname, _ := cfg.DB.GetSetting(ctx, "ns1_hostname")
		ns2Hostname, _ := cfg.DB.GetSetting(ctx, "ns2_hostname")

		if nsBrandDomain != "" {
			if ns1Hostname == "" {
				ns1Hostname = "ns1." + nsBrandDomain
			}
			if ns2Hostname == "" {
				ns2Hostname = "ns2." + nsBrandDomain
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"server_ip":       serverIP,
				"ns_brand_domain": nsBrandDomain,
				"ns1_hostname":    ns1Hostname,
				"ns2_hostname":    ns2Hostname,
				"instructions": gin.H{
					"a_record":    "Create an A record for @ pointing to " + serverIP,
					"nameservers": "Or set your domain's nameservers to " + ns1Hostname + " and " + ns2Hostname,
				},
			},
		})
	}
}

func syncZoneRecords(cfg RouterConfig, c *gin.Context, websiteDomain string, zoneID int64) error {
	ctx := c.Request.Context()

	records, err := cfg.DB.ListDNSRecordsByZone(ctx, zoneID)
	if err != nil {
		return fmt.Errorf("list records: %w", err)
	}

	zone, err := cfg.DB.GetZoneByID(ctx, zoneID)
	if err != nil || zone == nil {
		return fmt.Errorf("zone not found")
	}

	serial := nextSerial(zone.Serial)
	serverIP, _ := cfg.DB.GetSetting(ctx, "server_ip")

	var agentRecords []map[string]interface{}
	for _, r := range records {
		rec := map[string]interface{}{
			"type":  r.Type,
			"name":  r.Name,
			"value": r.Value,
			"ttl":   r.TTL,
		}
		if r.Priority != nil {
			rec["priority"] = *r.Priority
		}
		agentRecords = append(agentRecords, rec)
	}

	resp, err := cfg.AgentClient.Call(ctx, "dns.zone.sync_records", map[string]interface{}{
		"domain":    websiteDomain,
		"server_ip": serverIP,
		"serial":    serial,
		"records":   agentRecords,
	})
	if err != nil {
		return fmt.Errorf("agent call: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("agent error: %s", resp.Error.Message)
	}

	if err := cfg.DB.UpdateZoneSerial(ctx, zoneID, serial); err != nil {
		cfg.Log.WarnContext(ctx, "failed to update zone serial", "zone_id", zoneID, "error", err)
	}

	return nil
}

func listZonesHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		zones, err := cfg.DB.ListAllZones(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		type zoneWithRecords struct {
			ID          int64  `json:"id"`
			WebsiteID   *int64 `json:"website_id,omitempty"`
			Domain      string `json:"domain"`
			Serial      int64  `json:"serial"`
			RecordCount int    `json:"record_count"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
		}

		var result []zoneWithRecords
		for _, z := range zones {
			records, _ := cfg.DB.ListDNSRecordsByZone(ctx, z.ID)
			result = append(result, zoneWithRecords{
				ID:          z.ID,
				WebsiteID:   z.WebsiteID,
				Domain:      z.Domain,
				Serial:      z.Serial,
				RecordCount: len(records),
				CreatedAt:   z.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:   z.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	}
}

func createZoneHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Domain string `json:"domain" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "domain is required"))
			return
		}

		req.Domain = strings.TrimSpace(strings.ToLower(req.Domain))
		if req.Domain == "" {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "domain cannot be empty"))
			return
		}

		zone, err := cfg.DB.CreateZoneManual(c.Request.Context(), req.Domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "failed to create zone: "+err.Error()))
			return
		}

		records, _ := cfg.DB.ListDNSRecordsByZone(c.Request.Context(), zone.ID)
		recordModels := make([]map[string]interface{}, 0)
		for _, r := range records {
			rec := map[string]interface{}{
				"type":  r.Type,
				"name":  r.Name,
				"value": r.Value,
				"ttl":   r.TTL,
			}
			if r.Priority != nil {
				rec["priority"] = *r.Priority
			}
			recordModels = append(recordModels, rec)
		}

		ctx := c.Request.Context()
		resp, err := cfg.AgentClient.Call(ctx, "dns.zone.update", map[string]interface{}{
			"domain":  zone.Domain,
			"records": recordModels,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to sync zone to nameserver: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    zone,
		})
	}
}

func deleteZoneHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid zone id"))
			return
		}

		zone, err := cfg.DB.GetZoneByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}
		if zone == nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "zone not found"))
			return
		}

		if err := cfg.DB.DeleteAllRecordsForZone(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "failed to delete zone records: "+err.Error()))
			return
		}

		if err := cfg.DB.DeleteZone(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "failed to delete zone: "+err.Error()))
			return
		}

		ctx := c.Request.Context()
		resp, err := cfg.AgentClient.Call(ctx, "dns.zone.delete", map[string]interface{}{
			"domain": zone.Domain,
		})
		if err != nil {
			cfg.Log.ErrorContext(ctx, "dns.zone.delete agent error after db delete: "+err.Error())
		} else if resp.Error != nil {
			cfg.Log.ErrorContext(ctx, "dns.zone.delete agent error after db delete: "+resp.Error.Message)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	}
}

func restartBindHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := cfg.AgentClient.Call(c.Request.Context(), "services.restart", map[string]interface{}{
			"name": "named",
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to restart BIND: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		uid := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &uid, "Restarted BIND nameserver", c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "BIND restarted"},
		})
	}
}

func nameserverHealthHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		serverIP, _ := cfg.DB.GetSetting(ctx, "server_ip")
		ns1Hostname, _ := cfg.DB.GetSetting(ctx, "ns1_hostname")
		ns2Hostname, _ := cfg.DB.GetSetting(ctx, "ns2_hostname")
		nsBrandDomain, _ := cfg.DB.GetSetting(ctx, "ns_brand_domain")

		if nsBrandDomain != "" {
			if ns1Hostname == "" {
				ns1Hostname = "ns1." + nsBrandDomain
			}
			if ns2Hostname == "" {
				ns2Hostname = "ns2." + nsBrandDomain
			}
		}

		resp, err := cfg.AgentClient.Call(ctx, "dns.verify_nameservers", map[string]interface{}{
			"server_ip":    serverIP,
			"ns1_hostname": ns1Hostname,
			"ns2_hostname": ns2Hostname,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to check nameservers: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    resp.Result,
		})
	}
}

func checkAllDomainsDNSHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverIP, _ := cfg.DB.GetSetting(c.Request.Context(), "server_ip")
		if serverIP == "" {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "server IP not configured"))
			return
		}

		domains, err := cfg.DB.ListAllDomains(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("DB_ERROR", "failed to list domains"))
			return
		}

		type domainResult struct {
			Domain    string `json:"domain"`
			WebsiteID int64  `json:"website_id"`
			OK        bool   `json:"ok"`
			ARecord   string `json:"a_record,omitempty"`
			Error     string `json:"error,omitempty"`
		}

		var results []domainResult
		seen := make(map[string]bool)

		for _, d := range domains {
			if seen[d.Domain] {
				continue
			}
			seen[d.Domain] = true

			result := domainResult{
				Domain:    d.Domain,
				WebsiteID: d.WebsiteID,
			}

			resp, err := cfg.AgentClient.Call(c.Request.Context(), "dns.verify", map[string]interface{}{
				"domain":    d.Domain,
				"server_ip": serverIP,
			})
			if err != nil {
				result.Error = err.Error()
			} else if resp.Error != nil {
				result.Error = resp.Error.Message
			} else if data, ok := resp.Result.(map[string]interface{}); ok {
				if aRecord, ok := data["a_record"].(string); ok {
					result.ARecord = aRecord
				}
				if allOK, ok := data["all_ok"].(bool); ok {
					result.OK = allOK
				}
			}

			results = append(results, result)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    results,
		})
	}
}

func nextSerial(current int64) int64 {
	if current == 0 {
		t := time.Now()
		serial, _ := strconv.ParseInt(t.Format("20060102")+"01", 10, 64)
		return serial
	}
	s := strconv.FormatInt(current, 10)
	if len(s) != 10 {
		t := time.Now()
		serial, _ := strconv.ParseInt(t.Format("20060102")+"01", 10, 64)
		return serial
	}
	datePart := s[:8]
	seqPart := s[8:]
	t := time.Now()
	today := t.Format("20060102")
	if datePart == today {
		seq, _ := strconv.Atoi(seqPart)
		serial, _ := strconv.ParseInt(today+fmt.Sprintf("%02d", seq+1), 10, 64)
		return serial
	}
	serial, _ := strconv.ParseInt(today+"01", 10, 64)
	return serial
}
