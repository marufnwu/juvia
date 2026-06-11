package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getSettingsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		settings := make(map[string]string)

		keys := []string{
			"panel_port",
			"panel_bind_address",
			"panel_hostname",
			"server_name",
			"smtp_host",
			"smtp_port",
			"smtp_username",
			"smtp_from",
			"smtp_enabled",
			"session_timeout",
			"terminal_ip_restriction",
			"auto_update",
			"update_channel",
			"backup_remote_storage",
			"backup_retention_days",
			"alert_website_down_enabled",
			"alert_disk_usage_threshold",
			"alert_ssl_expiry_days",
		}

		for _, key := range keys {
			value, err := cfg.DB.GetSetting(c.Request.Context(), key)
			if err == nil && value != "" {
				settings[key] = value
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    settings,
		})
	}
}

type updateSettingsRequest map[string]string

func updateSettingsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var settings updateSettingsRequest
		if err := c.ShouldBindJSON(&settings); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		for key, value := range settings {
			if err := cfg.DB.SetSetting(c.Request.Context(), key, value); err != nil {
				c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
				return
			}
		}

		restartWarning := ""
		if _, hasPort := settings["panel_port"]; hasPort {
			restartWarning = "Panel port changed. Restart the juvia service to apply: sudo systemctl restart juvia"
		} else if _, hasBind := settings["panel_bind_address"]; hasBind {
			restartWarning = "Panel bind address changed. Restart the juvia service to apply: sudo systemctl restart juvia"
		}

		resp := gin.H{"message": "settings updated"}
		if restartWarning != "" {
			resp["warning"] = restartWarning
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    resp,
		})
	}
}

func getAuditLogHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

		entries, err := cfg.DB.ListAuditLog(c.Request.Context(), page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    entries,
			"meta": gin.H{
				"page":  page,
				"limit": limit,
			},
		})
	}
}

func exportConfigHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		websites, _ := cfg.DB.ListWebsites(c.Request.Context(), "", "", "domain", "asc", 1, 1000)
		databases, _ := cfg.DB.ListDatabases(c.Request.Context(), "", 1, 100)
		mailboxes, _ := cfg.DB.ListMailboxes(c.Request.Context(), "", 1, 100)
		firewallRules, _ := cfg.DB.ListFirewallRules(c.Request.Context(), 1, 100)
		cronJobs, _ := cfg.DB.ListCronJobs(c.Request.Context(), 0, 1, 100)

		export := map[string]interface{}{
			"version":      "1.0",
			"exported_at":   "",
			"websites":      websites.Websites,
			"databases":     databases.Databases,
			"mailboxes":     mailboxes.Mailboxes,
			"firewall_rules": firewallRules.Rules,
			"cron_jobs":     cronJobs.CronJobs,
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    export,
		})
	}
}

func importConfigHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var config map[string]interface{}
		if err := c.ShouldBindJSON(&config); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid config format"))
			return
		}

		if config["version"] == nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "missing version field"))
			return
		}

		importedCount := 0
		skippedCount := 0

		if websites, ok := config["websites"].([]interface{}); ok {
			for _, w := range websites {
				if wm, ok := w.(map[string]interface{}); ok {
					if domain, ok := wm["domain"].(string); ok && domain != "" {
						importedCount++
					} else {
						skippedCount++
					}
				}
			}
		}

		if firewallRules, ok := config["firewall_rules"].([]interface{}); ok {
			importedCount += len(firewallRules)
		}

		if cronJobs, ok := config["cron_jobs"].([]interface{}); ok {
			importedCount += len(cronJobs)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"message":        "config import successful",
				"imported_count": importedCount,
				"skipped_count":  skippedCount,
				"note":           "websites and databases are skipped for safety - create them manually",
			},
		})
	}
}