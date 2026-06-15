package api

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"juvia/internal/auth"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func validateEmail(email string) error {
	if email == "" {
		return &validationError{"email is required"}
	}
	if !emailRegex.MatchString(email) {
		return &validationError{"invalid email format"}
	}
	return nil
}

func listMailboxesHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		search := c.Query("search")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

		result, err := cfg.DB.ListMailboxes(c.Request.Context(), search, page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result.Mailboxes,
			"meta": gin.H{
				"total":  result.Total,
				"page":   page,
				"limit":  limit,
				"pages":  (result.Total + limit - 1) / limit,
			},
		})
	}
}

type createMailboxRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Quota       int64  `json:"quota"`
	DisplayName string `json:"display_name"`
}

func createMailboxHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createMailboxRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		if err := validateEmail(req.Email); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", err.Error()))
			return
		}

		if len(req.Password) < 8 {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "password must be at least 8 characters"))
			return
		}

		existing, _ := cfg.DB.GetMailboxByEmail(c.Request.Context(), req.Email)
		if existing != nil {
			c.JSON(http.StatusConflict, fail("CONFLICT", "mailbox already exists"))
			return
		}

		passwordHash, err := auth.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "failed to hash password"))
			return
		}

		if req.Quota == 0 {
			req.Quota = 10737418240
		}

		domain := strings.Split(req.Email, "@")[1]

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "email.create", map[string]interface{}{
			"email":    req.Email,
			"password": req.Password,
			"domain":   domain,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to create mailbox"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		mailbox, err := cfg.DB.CreateMailbox(c.Request.Context(), req.Email, passwordHash, req.Quota, req.DisplayName)
		if err != nil {
			cfg.AgentClient.Call(c.Request.Context(), "email.delete", map[string]interface{}{"email": req.Email})
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Created mailbox "+req.Email, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    mailbox,
		})
	}
}

func getMailboxHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid mailbox id"))
			return
		}

		mailbox, err := cfg.DB.GetMailboxByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "mailbox not found"))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    mailbox,
		})
	}
}

type updateMailboxRequest struct {
	DisplayName string `json:"display_name"`
	ForwardTo   string `json:"forward_to"`
	Quota       int64  `json:"quota"`
}

func updateMailboxHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid mailbox id"))
			return
		}

		var req updateMailboxRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		mailbox, err := cfg.DB.GetMailboxByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "mailbox not found"))
			return
		}

		displayName := req.DisplayName
		if displayName == "" && mailbox.DisplayName != nil {
			displayName = *mailbox.DisplayName
		}
		forwardTo := req.ForwardTo
		if forwardTo == "" && mailbox.ForwardTo != nil {
			forwardTo = *mailbox.ForwardTo
		}
		quota := req.Quota
		if quota == 0 {
			quota = mailbox.Quota
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "email.update", map[string]interface{}{
			"email":       mailbox.Email,
			"display_name": displayName,
			"quota":       quota,
			"forward_to":   forwardTo,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to update mailbox: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		if err := cfg.DB.UpdateMailbox(c.Request.Context(), id, displayName, forwardTo, quota); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "mailbox updated"},
		})
	}
}

func deleteMailboxHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid mailbox id"))
			return
		}

		mailbox, err := cfg.DB.GetMailboxByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "mailbox not found"))
			return
		}

		cfg.AgentClient.Call(c.Request.Context(), "email.delete", map[string]interface{}{
			"email": mailbox.Email,
		})

		if err := cfg.DB.DeleteMailbox(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Deleted mailbox "+mailbox.Email, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "mailbox deleted"},
		})
	}
}

func listAliasesHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		aliases, err := cfg.DB.ListAllAliases(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    aliases,
		})
	}
}

type createAliasRequest struct {
	Domain      string `json:"domain" binding:"required"`
	Source      string `json:"source" binding:"required"`
	Destination string `json:"destination" binding:"required"`
}

func createAliasHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createAliasRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "email.alias.create", map[string]interface{}{
			"domain":      req.Domain,
			"source":      req.Source,
			"destination": req.Destination,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to create alias"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		alias, err := cfg.DB.CreateAlias(c.Request.Context(), req.Domain, req.Source, req.Destination)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Created alias "+req.Source+" -> "+req.Destination, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    alias,
		})
	}
}

func deleteAliasHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid alias id"))
			return
		}

		cfg.AgentClient.Call(c.Request.Context(), "email.alias.delete", map[string]interface{}{
			"domain": c.Query("domain"),
			"source": c.Query("source"),
		})

		if err := cfg.DB.DeleteAlias(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "alias deleted"},
		})
	}
}

func listForwardersHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		forwarders, err := cfg.DB.ListAllForwarders(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    forwarders,
		})
	}
}

type createForwarderRequest struct {
	Domain      string `json:"domain" binding:"required"`
	Source      string `json:"source" binding:"required"`
	Destination string `json:"destination" binding:"required"`
}

func createForwarderHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createForwarderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "email.forwarder.create", map[string]interface{}{
			"domain":      req.Domain,
			"source":      req.Source,
			"destination": req.Destination,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to create forwarder"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		forwarder, err := cfg.DB.CreateForwarder(c.Request.Context(), req.Domain, req.Source, req.Destination)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Created forwarder "+req.Source+" -> "+req.Destination, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    forwarder,
		})
	}
}

func deleteForwarderHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid forwarder id"))
			return
		}

		cfg.AgentClient.Call(c.Request.Context(), "email.forwarder.delete", map[string]interface{}{
			"domain": c.Query("domain"),
			"source": c.Query("source"),
		})

		if err := cfg.DB.DeleteForwarder(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "forwarder deleted"},
		})
	}
}

type catchAllRequest struct {
	Domain    string `json:"domain" binding:"required"`
	ForwardTo string `json:"forward_to" binding:"required"`
}

func getCatchAllHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		domain := c.Query("domain")
		if domain == "" {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "domain is required"))
			return
		}

		catchAll, err := cfg.DB.GetCatchAll(c.Request.Context(), domain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    catchAll,
		})
	}
}

func setCatchAllHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req catchAllRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "email.catchall.set", map[string]interface{}{
			"domain":     req.Domain,
			"forward_to": req.ForwardTo,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to set catch-all: "+err.Error()))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		if err := cfg.DB.SetCatchAll(c.Request.Context(), req.Domain, req.ForwardTo); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Set catch-all for "+req.Domain+" -> "+req.ForwardTo, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "catch-all set"},
		})
	}
}

func deliverabilityHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		domain := c.Query("domain")
		if domain == "" {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "domain is required"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "email.deliverability", map[string]interface{}{
			"domain": domain,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to check deliverability"))
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
