package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"juvia/internal/db"
)

func listDomainsHandler(cfg RouterConfig) gin.HandlerFunc {
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

		if !canAccessWebsite(c, website) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		domains, err := cfg.DB.ListDomainsByWebsite(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    domains,
		})
	}
}

type createDomainRequest struct {
	Domain string `json:"domain" binding:"required"`
}

func createDomainHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		var req createDomainRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		req.Domain = strings.ToLower(strings.TrimSpace(req.Domain))
		if !isValidDomainOrWildcard(req.Domain) {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid domain format"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		if !canAccessWebsite(c, website) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		if err := checkDomainConflict(cfg.DB, c.Request.Context(), id, req.Domain); err != nil {
			c.JSON(http.StatusConflict, fail("CONFLICT", err.Error()))
			return
		}

		domainType := classifyDomain(req.Domain, website.Domain)

		if err := cfg.DB.CreateDomain(c.Request.Context(), id, req.Domain, domainType); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		domains, _ := cfg.DB.ListDomainsByWebsite(c.Request.Context(), id)
		var domainList []string
		for _, d := range domains {
			if d.Type != "primary" {
				domainList = append(domainList, d.Domain)
			}
		}

		serverIP, _ := cfg.DB.GetSetting(c.Request.Context(), "server_ip")
		resp, err := cfg.AgentClient.Call(c.Request.Context(), "dns.zone.update", map[string]interface{}{
			"domain":    website.Domain,
			"server_ip": serverIP,
		})
		if err != nil {
			cfg.Log.ErrorContext(c.Request.Context(), "dns.zone.update failed after add domain: "+err.Error())
		} else if resp.Error != nil {
			cfg.Log.ErrorContext(c.Request.Context(), "dns.zone.update agent error after add domain: "+resp.Error.Message)
		}

		cfg.AgentClient.Call(c.Request.Context(), "website.update", map[string]interface{}{
			"domain":       website.Domain,
			"php_version":  website.PHPVersion,
			"web_server":   website.WebServer,
			"extra_domains": domainList,
		})

		uid := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &uid, "Added domain "+req.Domain+" to "+website.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data": gin.H{
				"domain": req.Domain,
				"type":   domainType,
			},
		})
	}
}

func deleteDomainHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		domainID, err := strconv.ParseInt(c.Param("domainId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid domain id"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		if !canAccessWebsite(c, website) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		domain, err := cfg.DB.GetDomainByID(c.Request.Context(), domainID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}
		if domain == nil || domain.WebsiteID != id {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "domain not found"))
			return
		}

		if domain.Type == "primary" {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "cannot delete primary domain"))
			return
		}

		if err := cfg.DB.DeleteDomain(c.Request.Context(), domainID); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		domains, _ := cfg.DB.ListDomainsByWebsite(c.Request.Context(), id)
		var domainList []string
		for _, d := range domains {
			if d.Type != "primary" {
				domainList = append(domainList, d.Domain)
			}
		}

		serverIP, _ := cfg.DB.GetSetting(c.Request.Context(), "server_ip")
		resp, err := cfg.AgentClient.Call(c.Request.Context(), "dns.zone.update", map[string]interface{}{
			"domain":    website.Domain,
			"server_ip": serverIP,
		})
		if err != nil {
			cfg.Log.ErrorContext(c.Request.Context(), "dns.zone.update failed after delete domain: "+err.Error())
		} else if resp.Error != nil {
			cfg.Log.ErrorContext(c.Request.Context(), "dns.zone.update agent error after delete domain: "+resp.Error.Message)
		}

		cfg.AgentClient.Call(c.Request.Context(), "website.update", map[string]interface{}{
			"domain":        website.Domain,
			"php_version":   website.PHPVersion,
			"web_server":    website.WebServer,
			"extra_domains": domainList,
		})

		uid := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &uid, "Removed domain "+domain.Domain+" from "+website.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "domain deleted"},
		})
	}
}

func canAccessWebsite(c *gin.Context, website *db.Website) bool {
	userIDVal, _ := c.Get("user_id")
	role, _ := c.Get("role")
	if role == "admin" {
		return true
	}
	return website.UserID != nil && *website.UserID == userIDVal.(int64)
}

func isValidDomainOrWildcard(domain string) bool {
	if domain == "" || len(domain) > 253 {
		return false
	}
	if strings.HasPrefix(domain, "*.") {
		return isValidDomain(domain[2:])
	}
	return isValidDomain(domain)
}

func isValidDomain(domain string) bool {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" || len(part) > 63 {
			return false
		}
		for i, ch := range part {
			if i == 0 || i == len(part)-1 {
				if !isAlphaNum(ch) && ch != '_' {
					return false
				}
			} else {
				if !isAlphaNum(ch) && ch != '-' && ch != '_' {
					return false
				}
			}
		}
	}
	return true
}

func isAlphaNum(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

func classifyDomain(domain, primary string) string {
	if strings.HasPrefix(domain, "*.") {
		return "wildcard"
	}
	if strings.HasSuffix(domain, "."+primary) || domain == primary {
		return "subdomain"
	}
	return "alias"
}

func checkDomainConflict(dbConn *db.DB, ctx context.Context, currentWebsiteID int64, newDomain string) error {
	existing, err := dbConn.GetDomainByName(ctx, newDomain)
	if err != nil {
		return err
	}
	if existing != nil && existing.WebsiteID != currentWebsiteID {
		return fmt.Errorf("domain %s is already assigned to another website", existing.Domain)
	}

	allDomains, err := dbConn.ListAllDomains(ctx)
	if err != nil {
		return err
	}

	for _, d := range allDomains {
		if d.WebsiteID == currentWebsiteID {
			continue
		}
		if d.Domain == newDomain {
			return fmt.Errorf("domain %s is already assigned to another website", d.Domain)
		}
		if strings.HasPrefix(d.Domain, "*.") {
			base := d.Domain[2:]
			if newDomain == base || strings.HasSuffix(newDomain, "."+base) {
				return fmt.Errorf("domain %s conflicts with wildcard %s", newDomain, d.Domain)
			}
		}
		if strings.HasPrefix(newDomain, "*.") {
			base := newDomain[2:]
			if d.Domain == base || strings.HasSuffix(d.Domain, "."+base) {
				return fmt.Errorf("wildcard %s conflicts with domain %s", newDomain, d.Domain)
			}
		}
	}

	return nil
}

func issueDomainSSLHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}
		domainID, err := strconv.ParseInt(c.Param("domainId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid domain id"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil || website == nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userID.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		domain, err := cfg.DB.GetDomainByID(c.Request.Context(), domainID)
		if err != nil || domain == nil || domain.WebsiteID != id {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "domain not found"))
			return
		}

		var req struct {
			DNSProvider string `json:"dns_provider"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			req.DNSProvider = ""
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "ssl.issue_domain", map[string]interface{}{
			"domain":       domain.Domain,
			"dns_provider": req.DNSProvider,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to issue SSL: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		result, _ := resp.Result.(map[string]interface{})
		certType, _ := result["cert_type"].(string)
		certPath, _ := result["cert_path"].(string)
		keyPath, _ := result["key_path"].(string)
		expiry, _ := result["expiry"].(string)

		var expiryTime *time.Time
		if expiry != "" {
			if t, parseErr := time.Parse(time.RFC3339, expiry); parseErr == nil {
				expiryTime = &t
			}
		}

		cfg.DB.UpdateDomainSSL(c.Request.Context(), domainID, true, expiryTime, certType, certPath, keyPath)
		cfg.DB.UpdateWebsiteSSL(c.Request.Context(), id, true, expiryTime)

		var uid int64
		if v, ok := c.Get("user_id"); ok {
			uid = v.(int64)
		}
		cfg.DB.LogAudit(c.Request.Context(), &uid, "Issued SSL for "+domain.Domain+" (type: "+certType+")", c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	}
}

func renewDomainSSLHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}
		domainID, err := strconv.ParseInt(c.Param("domainId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid domain id"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil || website == nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userID.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		domain, err := cfg.DB.GetDomainByID(c.Request.Context(), domainID)
		if err != nil || domain == nil || domain.WebsiteID != id {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "domain not found"))
			return
		}

		certName := strings.ReplaceAll(domain.Domain, "*.", "wildcard-")
		certName = strings.ReplaceAll(certName, ".", "-")

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "ssl.renew_domain", map[string]interface{}{
			"cert_name": certName,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to renew SSL: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		result, _ := resp.Result.(map[string]interface{})
		expiry, _ := result["expiry"].(string)

		var expiryTime *time.Time
		if expiry != "" {
			if t, parseErr := time.Parse(time.RFC3339, expiry); parseErr == nil {
				expiryTime = &t
			}
		}

		cfg.DB.UpdateDomainSSL(c.Request.Context(), domainID, true, expiryTime, domain.SSLCertType, domain.SSLCertPath, domain.SSLKeyPath)

		var uid int64
		if v, ok := c.Get("user_id"); ok {
			uid = v.(int64)
		}
		cfg.DB.LogAudit(c.Request.Context(), &uid, "Renewed SSL for "+domain.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	}
}

func verifyDomainDNSHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}
		domainID, err := strconv.ParseInt(c.Param("domainId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid domain id"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil || website == nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userID.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		domain, err := cfg.DB.GetDomainByID(c.Request.Context(), domainID)
		if err != nil || domain == nil || domain.WebsiteID != id {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "domain not found"))
			return
		}

		serverIP, _ := cfg.DB.GetSetting(c.Request.Context(), "server_ip")
		ns1Hostname, _ := cfg.DB.GetSetting(c.Request.Context(), "ns1_hostname")
		ns2Hostname, _ := cfg.DB.GetSetting(c.Request.Context(), "ns2_hostname")
		nsBrandDomain, _ := cfg.DB.GetSetting(c.Request.Context(), "ns_brand_domain")
		if nsBrandDomain != "" {
			if ns1Hostname == "" {
				ns1Hostname = "ns1." + nsBrandDomain
			}
			if ns2Hostname == "" {
				ns2Hostname = "ns2." + nsBrandDomain
			}
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "dns.verify", map[string]interface{}{
			"domain":      domain.Domain,
			"server_ip":   serverIP,
			"ns1_hostname": ns1Hostname,
			"ns2_hostname": ns2Hostname,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to verify DNS: "+err.Error()))
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
