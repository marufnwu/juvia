package api

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

var appTypeRegex = regexp.MustCompile(`^(wordpress|laravel|nextjs|ghost|drupal|joomla)$`)

func listAppsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		apps, err := cfg.DB.ListAppsByWebsite(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    apps,
		})
	}
}

type installAppRequest struct {
	AppType       string `json:"app_type" binding:"required"`
	AdminUsername string `json:"admin_username"`
	AdminPassword string `json:"admin_password"`
	DBName        string `json:"db_name"`
	DBPassword    string `json:"db_password"`
}

func installAppHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		websiteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), websiteID)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		var req installAppRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		if !appTypeRegex.MatchString(req.AppType) {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid app type"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "apps.install", map[string]interface{}{
			"website_id":     websiteID,
			"website_path":  website.DocumentRoot,
			"app_type":      req.AppType,
			"admin_username": req.AdminUsername,
			"admin_password": req.AdminPassword,
			"db_name":       req.DBName,
			"db_password":    req.DBPassword,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to install app"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		result := resp.Result.(map[string]interface{})
		version, _ := result["version"].(string)
		appName := req.AppType
		if name, ok := result["name"].(string); ok {
			appName = name
		}

		app, err := cfg.DB.CreateApp(c.Request.Context(), req.AppType, appName, websiteID, version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Installed "+req.AppType+" on "+website.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    app,
		})
	}
}

func updateAppHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		websiteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		appID, err := strconv.ParseInt(c.Param("app_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid app id"))
			return
		}

		app, err := cfg.DB.GetAppByID(c.Request.Context(), appID)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "app not found"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), websiteID)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		checkResp, err := cfg.AgentClient.Call(c.Request.Context(), "apps.check-update", map[string]interface{}{
			"website_id":    websiteID,
			"website_path":  website.DocumentRoot,
			"app_type":      app.AppType,
			"current_version": app.Version,
		})
		if err != nil || checkResp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to check for updates"))
			return
		}

		checkResult := checkResp.Result.(map[string]interface{})
		updateAvailable, _ := checkResult["update_available"].(bool)

		if !updateAvailable {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    gin.H{"message": "no update available"},
			})
			return
		}

		updateResp, err := cfg.AgentClient.Call(c.Request.Context(), "apps.update", map[string]interface{}{
			"website_id":    websiteID,
			"website_path":  website.DocumentRoot,
			"app_type":      app.AppType,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to update app"))
			return
		}
		if updateResp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", updateResp.Error.Message))
			return
		}

		updateResult := updateResp.Result.(map[string]interface{})
		newVersion, _ := updateResult["version"].(string)

		cfg.DB.UpdateAppVersion(c.Request.Context(), appID, newVersion, false)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"message": "app updated",
				"version": newVersion,
			},
		})
	}
}