package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"juvia/internal/db"
)

func gitWebhookHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")

		website, err := findWebsiteByWebhookToken(c.Request.Context(), cfg.DB, token)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "webhook token not found"))
			return
		}

		if website.GitConfig == "" {
			c.JSON(http.StatusBadRequest, fail("NOT_CONFIGURED", "git not configured"))
			return
		}

		var gitCfg gitConfig
		if err := json.Unmarshal([]byte(website.GitConfig), &gitCfg); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "invalid git config"))
			return
		}

		body, _ := io.ReadAll(c.Request.Body)
		if gitCfg.DeployKey != "" && len(body) > 0 {
			signature := c.GetHeader("X-Hub-Signature-256")
			if signature != "" {
				parts := strings.SplitN(signature, "=", 2)
				if len(parts) == 2 && parts[0] == "sha256" {
					mac := hmac.New(sha256.New, []byte(gitCfg.DeployKey))
					mac.Write(body)
					expectedMAC := mac.Sum(nil)
					providedMAC, err := hex.DecodeString(parts[1])
					if err == nil && !hmac.Equal(expectedMAC, providedMAC) {
						c.JSON(http.StatusUnauthorized, fail("UNAUTHORIZED", "invalid signature"))
						return
					}
				}
			}
		}

		go func() {
			cfg.AgentClient.Call(c.Request.Context(), "git.pull", map[string]interface{}{
				"website_id":   website.ID,
				"website_path": website.DocumentRoot,
			})
		}()

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "webhook received, pull triggered"},
		})
	}
}

func findWebsiteByWebhookToken(ctx context.Context, db *db.DB, token string) (*db.Website, error) {
	websites, err := db.ListWebsites(ctx, "", "", "id", "asc", 1, 1000)
	if err != nil {
		return nil, err
	}

	for _, site := range websites.Websites {
		if site.GitConfig == "" {
			continue
		}
		var cfg gitConfig
		if err := json.Unmarshal([]byte(site.GitConfig), &cfg); err != nil {
			continue
		}
		if cfg.WebhookToken == token {
			return &site, nil
		}
	}

	return nil, fmt.Errorf("website not found")
}