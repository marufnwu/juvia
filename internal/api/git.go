package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type gitConfig struct {
	RepoURL    string `json:"repo_url"`
	Branch     string `json:"branch"`
	DeployKey  string `json:"deploy_key,omitempty"`
	AutoDeploy bool   `json:"auto_deploy"`
	WebhookToken string `json:"webhook_token"`
}

func getGitConfigHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := parseID(c.Param("id"))
		if id == 0 {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		var gitCfg gitConfig
		if website.GitConfig != "" {
			json.Unmarshal([]byte(website.GitConfig), &gitCfg)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gitCfg,
		})
	}
}

type setupGitRequest struct {
	RepoURL   string `json:"repo_url" binding:"required"`
	Branch    string `json:"branch" binding:"required"`
	DeployKey string `json:"deploy_key"`
	AutoDeploy bool `json:"auto_deploy"`
}

func setupGitHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := parseID(c.Param("id"))
		if id == 0 {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		var req setupGitRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		webhookToken := generateToken(32)

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "git.setup", map[string]interface{}{
			"website_id":   id,
			"website_path": website.DocumentRoot,
			"repo_url":    req.RepoURL,
			"branch":      req.Branch,
			"deploy_key": req.DeployKey,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to setup git"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		gitCfg := gitConfig{
			RepoURL:      req.RepoURL,
			Branch:       req.Branch,
			DeployKey:    req.DeployKey,
			AutoDeploy:   req.AutoDeploy,
			WebhookToken: webhookToken,
		}
		gitCfgJSON, _ := json.Marshal(gitCfg)

		cfg.DB.UpdateWebsiteGitConfig(c.Request.Context(), id, string(gitCfgJSON))

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Set up git deploy for "+website.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"message":       "git deploy configured",
				"webhook_token": webhookToken,
			},
		})
	}
}

func gitPullHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := parseID(c.Param("id"))
		if id == 0 {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "git.pull", map[string]interface{}{
			"website_id":   id,
			"website_path": website.DocumentRoot,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to pull"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "pull completed"},
		})
	}
}

func getWebhookURLHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := parseID(c.Param("id"))
		if id == 0 {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		var gitCfg gitConfig
		if website.GitConfig != "" {
			json.Unmarshal([]byte(website.GitConfig), &gitCfg)
		}

		if gitCfg.WebhookToken == "" {
			c.JSON(http.StatusBadRequest, fail("NOT_CONFIGURED", "git deploy not configured"))
			return
		}

		protocol := "https"
		if c.Request.TLS == nil {
			protocol = "http"
		}
		webhookURL := protocol + "://" + c.Request.Host + "/api/v1/webhooks/git/" + gitCfg.WebhookToken

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"webhook_url": webhookURL,
			},
		})
	}
}

func generateToken(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}