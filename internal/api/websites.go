package api

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/idna"
	"juvia/internal/tasks"
)

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

func validateDomain(domain string) error {
	if domain == "" || len(domain) > 255 {
		return &validationError{"domain is required and must be 255 characters or less"}
	}
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return &validationError{"domain cannot start or end with a dot"}
	}
	if _, err := idna.New().ToUnicode(domain); err != nil {
		return&validationError{"invalid domain name"}
	}
	if !domainRegex.MatchString(domain) {
		return &validationError{"invalid domain format"}
	}
	return nil
}

type validationError struct {
	msg string
}

func (e *validationError) Error() string {
	return e.msg
}

type listWebsitesHandlerParams struct {
	Status string `form:"status"`
	Search  string `form:"search"`
	Sort    string `form:"sort"`
	Order   string `form:"order"`
	Page    int    `form:"page"`
	Limit   int    `form:"limit"`
}

func listWebsitesHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p listWebsitesHandlerParams
		if err := c.ShouldBindQuery(&p); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", err.Error()))
			return
		}
		if p.Page < 1 {
			p.Page = 1
		}
		if p.Limit < 1 || p.Limit > 100 {
			p.Limit = 50
		}

		result, err := cfg.DB.ListWebsites(c.Request.Context(), p.Status, p.Search, p.Sort, p.Order, p.Page, p.Limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result.Websites,
			"meta": gin.H{
				"total":  result.Total,
				"page":   p.Page,
				"limit":  p.Limit,
				"pages":  (result.Total + p.Limit - 1) / p.Limit,
			},
		})
	}
}

type createWebsiteRequest struct {
	Domain     string `json:"domain" binding:"required"`
	PHPVersion string `json:"php_version" binding:"required"`
	WebServer  string `json:"web_server" binding:"required"`
}

func createWebsiteHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createWebsiteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		if err := validateDomain(req.Domain); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", err.Error()))
			return
		}

		validPHP := map[string]bool{"8.1": true, "8.2": true, "8.3": true}
		if !validPHP[req.PHPVersion] {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "PHP version must be 8.1, 8.2, or 8.3"))
			return
		}

		validServer := map[string]bool{"nginx": true, "apache": true}
		if !validServer[req.WebServer] {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "web server must be nginx or apache"))
			return
		}

		existing, _ := cfg.DB.GetWebsiteByDomain(c.Request.Context(), req.Domain)
		if existing != nil {
			c.JSON(http.StatusConflict, fail("CONFLICT", "a website with this domain already exists"))
			return
		}

		userID := getUserID(c)
		task, err := tasks.NewRunner(cfg.DB.DB, cfg.Log).CreateTask(c.Request.Context(), "create_website", userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "failed to create task"))
			return
		}

		documentRoot := "/home/" + req.Domain + "/public_html"
		website, err := cfg.DB.CreateWebsite(c.Request.Context(), req.Domain, documentRoot, req.PHPVersion, req.WebServer, userID)
		if err != nil {
			tasks.NewRunner(cfg.DB.DB, cfg.Log).FailTask(c.Request.Context(), task.TaskID, err.Error())
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}
		_ = cfg.DB.CreateDomain(c.Request.Context(), website.ID, req.Domain, "primary")

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "website.create", map[string]interface{}{
			"domain":      req.Domain,
			"php_version": req.PHPVersion,
			"web_server":  req.WebServer,
			"task_id":     task.TaskID,
		})
		if err != nil {
			cfg.DB.UpdateWebsiteStatus(c.Request.Context(), website.ID, "failed")
			tasks.NewRunner(cfg.DB.DB, cfg.Log).FailTask(c.Request.Context(), task.TaskID, err.Error())
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to communicate with agent: "+err.Error()))
			return
		}
		if resp.Error != nil {
			cfg.DB.UpdateWebsiteStatus(c.Request.Context(), website.ID, "failed")
			tasks.NewRunner(cfg.DB.DB, cfg.Log).FailTask(c.Request.Context(), task.TaskID, resp.Error.Message)
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		resultMap, ok := resp.Result.(map[string]interface{})
		if !ok {
			cfg.DB.UpdateWebsiteStatus(c.Request.Context(), website.ID, "failed")
			tasks.NewRunner(cfg.DB.DB, cfg.Log).FailTask(c.Request.Context(), task.TaskID, "invalid agent response")
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "invalid agent response"))
			return
		}

		if respDocRoot, ok := resultMap["document_root"].(string); ok && respDocRoot != "" {
			documentRoot = respDocRoot
		}

		if err := cfg.DB.UpdateWebsiteDocumentRoot(c.Request.Context(), website.ID, documentRoot); err != nil {
		}

		cfg.DB.UpdateWebsiteStatus(c.Request.Context(), website.ID, "active")
		tasks.NewRunner(cfg.DB.DB, cfg.Log).CompleteTask(c.Request.Context(), task.TaskID, `{"website_id":`+strconv.FormatInt(website.ID, 10)+`}`)

		cfg.DB.LogAudit(c.Request.Context(), &userID, "Created website "+req.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data": gin.H{
				"id":          website.ID,
				"domain":      website.Domain,
				"status":      "active",
				"php_version": req.PHPVersion,
				"web_server":  req.WebServer,
				"task_id":     task.TaskID,
			},
		})
	}
}

func getWebsiteHandler(cfg RouterConfig) gin.HandlerFunc {
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

		domains, _ := cfg.DB.ListDomainsByWebsite(c.Request.Context(), id)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"website": website,
				"domains": domains,
			},
		})
	}
}

type updateWebsiteRequest struct {
	PHPVersion string `json:"php_version"`
	WebServer  string `json:"web_server"`
}

func updateWebsiteHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		var req updateWebsiteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		if req.PHPVersion != "" {
			validPHP := map[string]bool{"8.1": true, "8.2": true, "8.3": true}
			if !validPHP[req.PHPVersion] {
				c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "PHP version must be 8.1, 8.2, or 8.3"))
				return
			}
		}
		if req.WebServer != "" {
			validServer := map[string]bool{"nginx": true, "apache": true}
			if !validServer[req.WebServer] {
				c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "web server must be nginx or apache"))
				return
			}
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		phpVersion := req.PHPVersion
		if phpVersion == "" && website.PHPVersion != nil {
			phpVersion = *website.PHPVersion
		}
		webServer := req.WebServer
		if webServer == "" {
			webServer = website.WebServer
		}

		if err := cfg.DB.UpdateWebsite(c.Request.Context(), id, phpVersion, webServer); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Updated website "+website.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "website updated"},
		})
	}
}

func deleteWebsiteHandler(cfg RouterConfig) gin.HandlerFunc {
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

		if err := cfg.DB.SoftDeleteWebsite(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Moved website "+website.Domain+" to trash", c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "website moved to trash"},
		})
	}
}

func suspendWebsiteHandler(cfg RouterConfig) gin.HandlerFunc {
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

		userID := getUserID(c)
		task, _ := tasks.NewRunner(cfg.DB.DB, cfg.Log).CreateTask(c.Request.Context(), "suspend_website", userID)

		_, err = cfg.AgentClient.Call(c.Request.Context(), "website.suspend", map[string]interface{}{
			"domain":       website.Domain,
			"document_root": website.DocumentRoot,
		})
		if err != nil {
			tasks.NewRunner(cfg.DB.DB, cfg.Log).FailTask(c.Request.Context(), task.TaskID, err.Error())
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to suspend website"))
			return
		}

		if err := cfg.DB.SuspendWebsite(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		tasks.NewRunner(cfg.DB.DB, cfg.Log).CompleteTask(c.Request.Context(), task.TaskID, `{}`)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Suspended website "+website.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "website suspended"},
		})
	}
}

func restoreWebsiteHandler(cfg RouterConfig) gin.HandlerFunc {
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

		if err := cfg.DB.RestoreWebsite(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Restored website "+website.Domain+" from trash", c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "website restored"},
		})
	}
}

func permanentDeleteWebsiteHandler(cfg RouterConfig) gin.HandlerFunc {
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

		userID := getUserID(c)
		task, _ := tasks.NewRunner(cfg.DB.DB, cfg.Log).CreateTask(c.Request.Context(), "delete_website", userID)

		_, err = cfg.AgentClient.Call(c.Request.Context(), "website.delete", map[string]interface{}{
			"domain":       website.Domain,
			"document_root": website.DocumentRoot,
			"force":        true,
		})
		if err != nil {
			tasks.NewRunner(cfg.DB.DB, cfg.Log).FailTask(c.Request.Context(), task.TaskID, err.Error())
		}

		if err := cfg.DB.PermanentlyDeleteWebsite(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		tasks.NewRunner(cfg.DB.DB, cfg.Log).CompleteTask(c.Request.Context(), task.TaskID, `{}`)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Permanently deleted website "+website.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "website permanently deleted"},
		})
	}
}

func listTrashedWebsitesHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		websites, err := cfg.DB.ListTrashedWebsites(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    websites,
		})
	}
}

func purgeTrashHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := cfg.DB.PurgeTrash(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "trash purged"},
		})
	}
}

type changePHPRequest struct {
	PHPVersion string `json:"php_version" binding:"required"`
}

func changePHPHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid website id"))
			return
		}

		var req changePHPRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		validPHP := map[string]bool{"8.1": true, "8.2": true, "8.3": true}
		if !validPHP[req.PHPVersion] {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "PHP version must be 8.1, 8.2, or 8.3"))
			return
		}

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		if err := cfg.DB.UpdateWebsite(c.Request.Context(), id, req.PHPVersion, website.WebServer); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Changed PHP version to "+req.PHPVersion+" for "+website.Domain, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "PHP version changed"},
		})
	}
}
