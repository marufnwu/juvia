package api

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var cronExprRegex = regexp.MustCompile(`^(\S+\s+){4}\S+$`)

func validateCronExpression(expr string) error {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return &validationError{"cron expression must have exactly 5 fields"}
	}
	return nil
}

func validateCronCommand(command string) error {
	dangerous := []string{"rm -rf /", "dd if=", "> /dev/", "mkfs", ":(){:|:&};"}
	lower := strings.ToLower(command)
	for _, d := range dangerous {
		if strings.Contains(lower, d) {
			return &validationError{"command contains disallowed pattern"}
		}
	}
	return nil
}

func listCronJobsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := getUserID(c)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

		result, err := cfg.DB.ListCronJobs(c.Request.Context(), userID, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result.CronJobs,
			"meta": gin.H{
				"total":  result.Total,
				"page":   page,
				"limit":  limit,
				"pages":  (result.Total + limit - 1) / limit,
			},
		})
	}
}

type createCronJobRequest struct {
	Schedule string `json:"schedule" binding:"required"`
	Command  string `json:"command" binding:"required"`
	RunAs    string `json:"run_as" binding:"required"`
	Type     string `json:"type"`
}

func createCronJobHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createCronJobRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		if err := validateCronExpression(req.Schedule); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", err.Error()))
			return
		}

		if err := validateCronCommand(req.Command); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", err.Error()))
			return
		}

		jobType := req.Type
		if jobType == "" {
			jobType = "shell"
		}

		userID := getUserID(c)

		job, err := cfg.DB.CreateCronJob(c.Request.Context(), userID, req.Schedule, req.Command, req.RunAs, jobType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "cron.create", map[string]interface{}{
			"id":       job.ID,
			"schedule": req.Schedule,
			"command":  req.Command,
			"run_as":   req.RunAs,
		})
		if err != nil {
			cfg.DB.DeleteCronJob(c.Request.Context(), job.ID)
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to create cron job"))
			return
		}
		if resp.Error != nil {
			cfg.DB.DeleteCronJob(c.Request.Context(), job.ID)
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		cfg.DB.LogAudit(c.Request.Context(), &userID, "Created cron job", c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    job,
		})
	}
}

func getCronJobHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid cron job id"))
			return
		}

		job, err := cfg.DB.GetCronJobByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "cron job not found"))
			return
		}

		logs, _ := cfg.DB.ListCronJobLogs(c.Request.Context(), id, 10)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"job":  job,
				"logs": logs,
			},
		})
	}
}

type updateCronJobRequest struct {
	Schedule string `json:"schedule"`
	Command  string `json:"command"`
	RunAs    string `json:"run_as"`
}

func updateCronJobHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid cron job id"))
			return
		}

		job, err := cfg.DB.GetCronJobByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "cron job not found"))
			return
		}

		var req updateCronJobRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		schedule := req.Schedule
		if schedule == "" {
			schedule = job.Schedule
		}
		command := req.Command
		if command == "" {
			command = job.Command
		}
		runAs := req.RunAs
		if runAs == "" {
			runAs = job.RunAs
		}

		if err := validateCronExpression(schedule); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", err.Error()))
			return
		}
		if err := validateCronCommand(command); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", err.Error()))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "cron.update", map[string]interface{}{
			"id":       id,
			"schedule": schedule,
			"command":  command,
			"run_as":   runAs,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to update cron job: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		if err := cfg.DB.UpdateCronJob(c.Request.Context(), id, schedule, command, runAs); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "cron job updated"},
		})
	}
}

func deleteCronJobHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid cron job id"))
			return
		}

		job, err := cfg.DB.GetCronJobByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "cron job not found"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "cron.delete", map[string]interface{}{
			"id": id,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to delete cron job from server: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		if err := cfg.DB.DeleteCronJob(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		cfg.DB.LogAudit(c.Request.Context(), &job.UserID, "Deleted cron job "+job.Command, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "cron job deleted", "job_id": job.ID},
		})
	}
}

func enableCronJobHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid cron job id"))
			return
		}

		job, err := cfg.DB.GetCronJobByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "cron job not found"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "cron.enable", map[string]interface{}{
			"id": id,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to enable cron job: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		if err := cfg.DB.EnableCronJob(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		cfg.DB.LogAudit(c.Request.Context(), &job.UserID, "Enabled cron job "+job.Command, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "cron job enabled"},
		})
	}
}

func disableCronJobHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid cron job id"))
			return
		}

		job, err := cfg.DB.GetCronJobByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "cron job not found"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "cron.disable", map[string]interface{}{
			"id": id,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to disable cron job: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		if err := cfg.DB.DisableCronJob(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		cfg.DB.LogAudit(c.Request.Context(), &job.UserID, "Disabled cron job "+job.Command, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "cron job disabled"},
		})
	}
}

func getCronJobLogsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid cron job id"))
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		logs, err := cfg.DB.ListCronJobLogs(c.Request.Context(), id, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    logs,
		})
	}
}