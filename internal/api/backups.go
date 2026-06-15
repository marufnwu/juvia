package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func listBackupsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

		var websiteID *int64
		if wid := c.Query("website_id"); wid != "" {
			if id, err := strconv.ParseInt(wid, 10, 64); err == nil {
				websiteID = &id
			}
		}

		result, err := cfg.DB.ListBackups(c.Request.Context(), websiteID, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result.Backups,
			"meta": gin.H{
				"total":  result.Total,
				"page":   page,
				"limit":  limit,
				"pages":  (result.Total + limit - 1) / limit,
			},
		})
	}
}

type createBackupRequest struct {
	WebsiteID *int64 `json:"website_id"`
	Type      string `json:"type" binding:"required"`
	Storage   string `json:"storage"`
}

func createBackupHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createBackupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		if req.Type != "full" && req.Type != "files" && req.Type != "database" {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "type must be 'full', 'files', or 'database'"))
			return
		}

		storage := req.Storage
		if storage == "" {
			storage = "local"
		}

		backupPath := "/var/backups/juvia"
		if req.WebsiteID != nil {
			backupPath = backupPath + "/" + strconv.FormatInt(*req.WebsiteID, 10)
		}
		backupPath = backupPath + "/" + time.Now().Format("20060102_150405")

		backup, err := cfg.DB.CreateBackup(c.Request.Context(), req.WebsiteID, req.Type, storage, backupPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "backup.create", map[string]interface{}{
			"website_id": req.WebsiteID,
			"type":       req.Type,
			"storage":    storage,
			"path":       backupPath,
		})
		if err != nil {
			cfg.DB.UpdateBackupStatus(c.Request.Context(), backup.ID, "failed", 0, "")
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "backup creation failed"))
			return
		}
		if resp.Error != nil {
			cfg.DB.UpdateBackupStatus(c.Request.Context(), backup.ID, "failed", 0, "")
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		result := resp.Result.(map[string]interface{})
		sizeBytes, _ := result["size"].(float64)
		checksum, _ := result["checksum"].(string)

		cfg.DB.UpdateBackupStatus(c.Request.Context(), backup.ID, "completed", int64(sizeBytes), checksum)
		cfg.DB.MarkBackupVerified(c.Request.Context(), backup.ID)

		backup, _ = cfg.DB.GetBackupByID(c.Request.Context(), backup.ID)

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Created backup", c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    backup,
		})
	}
}

func getBackupHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid backup id"))
			return
		}

		backup, err := cfg.DB.GetBackupByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "backup not found"))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    backup,
		})
	}
}

type restoreBackupRequest struct {
	WebsiteID *int64 `json:"website_id"`
}

func restoreBackupHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid backup id"))
			return
		}

		backup, err := cfg.DB.GetBackupByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "backup not found"))
			return
		}

		var req restoreBackupRequest
		if err := c.ShouldBindJSON(&req); err == nil && req.WebsiteID != nil {
			backup.WebsiteID = req.WebsiteID
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "backup.restore", map[string]interface{}{
			"backup_id":   backup.ID,
			"website_id":  backup.WebsiteID,
			"backup_path": backup.Path,
			"backup_type": backup.Type,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "restore failed"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Restored backup "+backup.Path, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "backup restored successfully"},
		})
	}
}

func deleteBackupHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid backup id"))
			return
		}

		backup, err := cfg.DB.GetBackupByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "backup not found"))
			return
		}

		cfg.AgentClient.Call(c.Request.Context(), "backup.delete", map[string]interface{}{
			"path": backup.Path,
		})

		if err := cfg.DB.DeleteBackup(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "backup deleted"},
		})
	}
}

func listBackupSchedulesHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var websiteID *int64
		if wid := c.Query("website_id"); wid != "" {
			if id, err := strconv.ParseInt(wid, 10, 64); err == nil {
				websiteID = &id
			}
		}

		schedules, err := cfg.DB.ListBackupSchedules(c.Request.Context(), websiteID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    schedules,
		})
	}
}

type createBackupScheduleRequest struct {
	WebsiteID     *int64 `json:"website_id"`
	Schedule      string `json:"schedule" binding:"required"`
	RetentionDays int    `json:"retention_days"`
	Storage       string `json:"storage"`
}

func createBackupScheduleHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createBackupScheduleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		retentionDays := req.RetentionDays
		if retentionDays == 0 {
			retentionDays = 7
		}

		storage := req.Storage
		if storage == "" {
			storage = "local"
		}

		schedule, err := cfg.DB.CreateBackupSchedule(c.Request.Context(), req.WebsiteID, req.Schedule, retentionDays, storage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "backup.schedule.create", map[string]interface{}{
			"schedule_id":    schedule.ID,
			"website_id":    req.WebsiteID,
			"schedule":      req.Schedule,
			"storage":       storage,
			"retention_days": retentionDays,
		})
		if err != nil {
			cfg.DB.DeleteBackupSchedule(c.Request.Context(), schedule.ID)
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to create backup schedule: "+err.Error()))
			return
		}
		if resp.Error != nil {
			cfg.DB.DeleteBackupSchedule(c.Request.Context(), schedule.ID)
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    schedule,
		})
	}
}

func deleteBackupScheduleHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid schedule id"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "backup.schedule.delete", map[string]interface{}{
			"schedule_id": id,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to delete backup schedule from server: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		if err := cfg.DB.DeleteBackupSchedule(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "schedule deleted"},
		})
	}
}