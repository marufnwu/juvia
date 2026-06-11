package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func listAlertsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

		var acknowledged *bool
		if ackStr := c.Query("acknowledged"); ackStr != "" {
			ack := ackStr == "true"
			acknowledged = &ack
		}

		result, err := cfg.DB.ListAlerts(c.Request.Context(), acknowledged, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result.Alerts,
			"meta": gin.H{
				"total":  result.Total,
				"page":   page,
				"limit":  limit,
				"pages":  (result.Total + limit - 1) / limit,
			},
		})
	}
}

func acknowledgeAlertHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid alert id"))
			return
		}

		if err := cfg.DB.AcknowledgeAlert(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "alert acknowledged"},
		})
	}
}

func deleteAlertHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid alert id"))
			return
		}

		if err := cfg.DB.DeleteAlert(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "alert deleted"},
		})
	}
}

type AlertSettings struct {
	WebsiteDownEnabled    bool `json:"website_down_enabled"`
	WebsiteDownThreshold  int  `json:"website_down_threshold"`
	SSLExpiryEnabled      bool `json:"ssl_expiry_enabled"`
	SSLExpiryDays         int  `json:"ssl_expiry_days"`
	DiskUsageEnabled      bool `json:"disk_usage_enabled"`
	DiskUsageThreshold    int  `json:"disk_usage_threshold"`
	BackupOverdueEnabled  bool `json:"backup_overdue_enabled"`
	BackupOverdueDays     int  `json:"backup_overdue_days"`
	ServiceDownEnabled    bool `json:"service_down_enabled"`
}

func getAlertSettingsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		settings := AlertSettings{
			WebsiteDownEnabled:   true,
			WebsiteDownThreshold: 5,
			SSLExpiryEnabled:     true,
			SSLExpiryDays:        14,
			DiskUsageEnabled:     true,
			DiskUsageThreshold:   80,
			BackupOverdueEnabled: true,
			BackupOverdueDays:    7,
			ServiceDownEnabled:   true,
		}

		if val, err := cfg.DB.GetAlertSetting(c.Request.Context(), "alert_website_down_enabled"); err == nil && val != "" {
			settings.WebsiteDownEnabled = val == "true"
		}
		if val, err := cfg.DB.GetAlertSetting(c.Request.Context(), "alert_disk_usage_threshold"); err == nil && val != "" {
			settings.DiskUsageThreshold, _ = strconv.Atoi(val)
		}
		if val, err := cfg.DB.GetAlertSetting(c.Request.Context(), "alert_ssl_expiry_days"); err == nil && val != "" {
			settings.SSLExpiryDays, _ = strconv.Atoi(val)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    settings,
		})
	}
}

func updateAlertSettingsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var settings AlertSettings
		if err := c.ShouldBindJSON(&settings); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		cfg.DB.SetAlertSetting(c.Request.Context(), "alert_website_down_enabled", strconv.FormatBool(settings.WebsiteDownEnabled))
		cfg.DB.SetAlertSetting(c.Request.Context(), "alert_website_down_threshold", strconv.Itoa(settings.WebsiteDownThreshold))
		cfg.DB.SetAlertSetting(c.Request.Context(), "alert_ssl_expiry_enabled", strconv.FormatBool(settings.SSLExpiryEnabled))
		cfg.DB.SetAlertSetting(c.Request.Context(), "alert_ssl_expiry_days", strconv.Itoa(settings.SSLExpiryDays))
		cfg.DB.SetAlertSetting(c.Request.Context(), "alert_disk_usage_enabled", strconv.FormatBool(settings.DiskUsageEnabled))
		cfg.DB.SetAlertSetting(c.Request.Context(), "alert_disk_usage_threshold", strconv.Itoa(settings.DiskUsageThreshold))
		cfg.DB.SetAlertSetting(c.Request.Context(), "alert_backup_overdue_enabled", strconv.FormatBool(settings.BackupOverdueEnabled))
		cfg.DB.SetAlertSetting(c.Request.Context(), "alert_backup_overdue_days", strconv.Itoa(settings.BackupOverdueDays))
		cfg.DB.SetAlertSetting(c.Request.Context(), "alert_service_down_enabled", strconv.FormatBool(settings.ServiceDownEnabled))

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "alert settings updated"},
		})
	}
}