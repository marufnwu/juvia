package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func listDNSRecordsHandler(cfg RouterConfig) gin.HandlerFunc {
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

		records, err := cfg.DB.ListDNSRecordsByZone(c.Request.Context(), zone.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    records,
		})
	}
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

		zone, err := cfg.DB.GetZoneByWebsite(c.Request.Context(), id)
		if err != nil || zone == nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "DNS zone not found"))
			return
		}

		website, _ := cfg.DB.GetWebsiteByID(c.Request.Context(), id)

		record, err := cfg.DB.CreateDNSRecord(c.Request.Context(), zone.ID, req.Type, req.Name, req.Value, req.Priority)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		_, _ = cfg.AgentClient.Call(c.Request.Context(), "dns.zone.update", map[string]interface{}{
			"domain": website.Domain,
		})

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Added "+req.Type+" record to "+website.Domain, c.ClientIP(), c.Request.UserAgent(), "")

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

		if err := cfg.DB.UpdateDNSRecord(c.Request.Context(), id, name, value, req.Priority); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

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

		_, err = cfg.DB.GetDNSRecordByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "DNS record not found"))
			return
		}

		if err := cfg.DB.DeleteDNSRecord(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
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
				"value":       "v=DMARC1; p=quarantine; rua=mailto:dmarc@example.com",
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
