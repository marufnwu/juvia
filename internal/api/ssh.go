package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func getSSHSettingsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		port := "22"
		rootLogin := "prohibit-password"

		if v, err := cfg.DB.GetSetting(c.Request.Context(), "ssh_port"); err == nil && v != "" {
			port = v
		}
		if v, err := cfg.DB.GetSetting(c.Request.Context(), "ssh_root_login"); err == nil && v != "" {
			rootLogin = v
		}

		portInt, _ := strconv.Atoi(port)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"port":       portInt,
				"root_login": rootLogin,
			},
		})
	}
}

type updateSSHSettingsRequest struct {
	Port      int    `json:"port"`
	RootLogin string `json:"root_login"`
}

func updateSSHSettingsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req updateSSHSettingsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		if req.Port < 1 || req.Port > 65535 {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "port must be between 1 and 65535"))
			return
		}

		if req.RootLogin == "" {
			req.RootLogin = "prohibit-password"
		}

		validRootLogin := map[string]bool{
			"yes":               true,
			"no":                true,
			"prohibit-password": true,
			"without-password":  true,
		}
		if !validRootLogin[req.RootLogin] {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid root_login value"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "ssh.configure", map[string]interface{}{
			"port":       req.Port,
			"root_login": req.RootLogin,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to configure SSH: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		if err := cfg.DB.SetSetting(c.Request.Context(), "ssh_port", strconv.Itoa(req.Port)); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "failed to save ssh_port setting: "+err.Error()))
			return
		}
		if err := cfg.DB.SetSetting(c.Request.Context(), "ssh_root_login", req.RootLogin); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "failed to save ssh_root_login setting: "+err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"port":       req.Port,
				"root_login": req.RootLogin,
				"message":    "SSH settings updated. Restart the SSH service if the port was changed.",
			},
		})
	}
}
