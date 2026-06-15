package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getServerIPHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		cachedIP, err := cfg.DB.GetSetting(ctx, "server_ip")
		if err == nil && cachedIP != "" {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"ip":      cachedIP,
					"cached":  true,
					"message": "Using cached server IP",
				},
			})
			return
		}

		resp, err := cfg.AgentClient.Call(ctx, "network.get_public_ip", map[string]interface{}{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to discover server IP: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		result, ok := resp.Result.(map[string]interface{})
		if !ok {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "invalid response from agent"))
			return
		}

		ip, _ := result["ip"].(string)
		if ip == "" {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "agent returned empty IP"))
			return
		}

		if err := cfg.DB.SetSetting(ctx, "server_ip", ip); err != nil {
			cfg.Log.Warn("failed to cache server ip", "error", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"ip":      ip,
				"cached":  false,
				"message": "IP discovered and cached",
			},
		})
	}
}

func refreshServerIPHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		resp, err := cfg.AgentClient.Call(ctx, "network.get_public_ip", map[string]interface{}{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to discover server IP: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		result, ok := resp.Result.(map[string]interface{})
		if !ok {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "invalid response from agent"))
			return
		}

		ip, _ := result["ip"].(string)
		if ip == "" {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "agent returned empty IP"))
			return
		}

		if err := cfg.DB.SetSetting(ctx, "server_ip", ip); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "failed to cache IP: "+err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"ip":      ip,
				"message": "IP refreshed and cached",
			},
		})
	}
}
