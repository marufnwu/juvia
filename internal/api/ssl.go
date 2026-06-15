package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"juvia/internal/tasks"
)

func sslCheckHandler(cfg RouterConfig) gin.HandlerFunc {
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

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userID.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "ssl.check", map[string]interface{}{
			"domain": website.Domain,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to check SSL"))
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

func issueSSLHandler(cfg RouterConfig) gin.HandlerFunc {
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

		ownerID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && *website.UserID != ownerID.(int64) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		userID := getUserID(c)
		task, _ := tasks.NewRunner(cfg.DB.DB, cfg.Log).CreateTask(c.Request.Context(), "issue_ssl", userID)

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "ssl.issue", map[string]interface{}{
			"domain": website.Domain,
		})
		if err != nil {
			tasks.NewRunner(cfg.DB.DB, cfg.Log).FailTask(c.Request.Context(), task.TaskID, err.Error())
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to issue SSL"))
			return
		}
		if resp.Error != nil {
			tasks.NewRunner(cfg.DB.DB, cfg.Log).FailTask(c.Request.Context(), task.TaskID, resp.Error.Message)
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		resultMap, _ := resp.Result.(map[string]interface{})
		expiryVal, _ := resultMap["expiry"].(time.Time)
		if expiryVal.IsZero() {
			expiryVal = time.Now().Add(90 * 24 * time.Hour)
		}

		if err := cfg.DB.UpdateWebsiteSSL(c.Request.Context(), id, true, &expiryVal); err != nil {
			tasks.NewRunner(cfg.DB.DB, cfg.Log).FailTask(c.Request.Context(), task.TaskID, err.Error())
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		tasks.NewRunner(cfg.DB.DB, cfg.Log).CompleteTask(c.Request.Context(), task.TaskID, `{}`)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Issued SSL certificate for "+website.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		data := gin.H{
			"message": "SSL certificate issued",
			"expiry":  expiryVal,
		}
		if certType, ok := resultMap["cert_type"].(string); ok {
			data["cert_type"] = certType
		}
		if warning, ok := resultMap["warning"].(string); ok {
			data["warning"] = warning
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    data,
		})
	}
}

func renewSSLHandler(cfg RouterConfig) gin.HandlerFunc {
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

		ownerID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && *website.UserID != ownerID.(int64) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		userID := getUserID(c)
		task, _ := tasks.NewRunner(cfg.DB.DB, cfg.Log).CreateTask(c.Request.Context(), "renew_ssl", userID)

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "ssl.renew", map[string]interface{}{
			"domain": website.Domain,
		})
		if err != nil {
			tasks.NewRunner(cfg.DB.DB, cfg.Log).FailTask(c.Request.Context(), task.TaskID, err.Error())
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to renew SSL"))
			return
		}
		if resp.Error != nil {
			tasks.NewRunner(cfg.DB.DB, cfg.Log).FailTask(c.Request.Context(), task.TaskID, resp.Error.Message)
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		resultMap, _ := resp.Result.(map[string]interface{})
		expiryVal, _ := resultMap["expiry"].(time.Time)
		if expiryVal.IsZero() {
			expiryVal = time.Now().Add(90 * 24 * time.Hour)
		}

		if err := cfg.DB.UpdateWebsiteSSL(c.Request.Context(), id, true, &expiryVal); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		tasks.NewRunner(cfg.DB.DB, cfg.Log).CompleteTask(c.Request.Context(), task.TaskID, `{}`)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Renewed SSL certificate for "+website.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		data := gin.H{
			"message": "SSL certificate renewed",
			"expiry":  expiryVal,
		}
		if certType, ok := resultMap["cert_type"].(string); ok {
			data["cert_type"] = certType
		}
		if warning, ok := resultMap["warning"].(string); ok {
			data["warning"] = warning
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    data,
		})
	}
}

func removeSSLHandler(cfg RouterConfig) gin.HandlerFunc {
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

		ownerID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && *website.UserID != ownerID.(int64) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		userID := getUserID(c)
		_, _ = cfg.AgentClient.Call(c.Request.Context(), "ssl.remove", map[string]interface{}{
			"domain": website.Domain,
		})

		if err := cfg.DB.UpdateWebsiteSSL(c.Request.Context(), id, false, nil); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		cfg.DB.LogAudit(c.Request.Context(), &userID, "Removed SSL certificate from "+website.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "SSL certificate removed"},
		})
	}
}
