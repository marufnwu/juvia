package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func webmailStatusHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := cfg.AgentClient.Call(c.Request.Context(), "webmail.status", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to check webmail status"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		result := resp.Result.(map[string]interface{})

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"installed": result["installed"],
			},
		})
	}
}

func installWebmailHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := cfg.AgentClient.Call(c.Request.Context(), "webmail.install", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to install webmail"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Installed webmail", c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "webmail installed"},
		})
	}
}

func uninstallWebmailHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := cfg.AgentClient.Call(c.Request.Context(), "webmail.uninstall", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to uninstall webmail"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Uninstalled webmail", c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "webmail uninstalled"},
		})
	}
}

func webmailURLHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := cfg.AgentClient.Call(c.Request.Context(), "webmail.status", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to check webmail status"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		result := resp.Result.(map[string]interface{})
		if !result["installed"].(bool) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"installed": false,
					"message":   "webmail is not installed",
				},
			})
			return
		}

		url := "/webmail"

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"url": url,
			},
		})
	}
}