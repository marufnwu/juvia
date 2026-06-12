package api

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	firewallPortRegex = regexp.MustCompile(`^\d{1,5}(/\w+)?$`)
	firewallSourceRegex = regexp.MustCompile(`^(any|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}(/\d{1,2})?)$`)
)

func validateFirewallPort(port string) error {
	if port == "" {
		return &validationError{"port is required"}
	}
	if !firewallPortRegex.MatchString(port) {
		return &validationError{"invalid port format"}
	}
	return nil
}

func validateFirewallSource(source string) error {
	if source == "" {
		source = "any"
	}
	if source != "any" && !firewallSourceRegex.MatchString(source) {
		return &validationError{"invalid source format (use IP or CIDR notation)"}
	}
	return nil
}

func listFirewallRulesHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

		result, err := cfg.DB.ListFirewallRules(c.Request.Context(), page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result.Rules,
			"meta": gin.H{
				"total":  result.Total,
				"page":   page,
				"limit":  limit,
				"pages":  (result.Total + limit - 1) / limit,
			},
		})
	}
}

type createFirewallRuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Action      string `json:"action" binding:"required"`
	Port        string `json:"port" binding:"required"`
	Protocol    string `json:"protocol" binding:"required"`
	Source      string `json:"source"`
}

func createFirewallRuleHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createFirewallRuleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		if req.Action != "allow" && req.Action != "deny" {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "action must be 'allow' or 'deny'"))
			return
		}

		if req.Protocol != "tcp" && req.Protocol != "udp" && req.Protocol != "both" {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "protocol must be 'tcp', 'udp', or 'both'"))
			return
		}

		if err := validateFirewallPort(req.Port); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", err.Error()))
			return
		}

		source := req.Source
		if source == "" {
			source = "any"
		}
		if err := validateFirewallSource(source); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", err.Error()))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "firewall.rule.create", map[string]interface{}{
			"action":   req.Action,
			"port":     req.Port,
			"protocol": req.Protocol,
			"source":   source,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to create firewall rule"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		rule, err := cfg.DB.CreateFirewallRule(c.Request.Context(), req.Name, req.Description, req.Action, req.Port, req.Protocol, source)
		if err != nil {
			cfg.AgentClient.Call(c.Request.Context(), "firewall.rule.delete", map[string]interface{}{
				"port":   req.Port,
				"source": source,
			})
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Created firewall rule: "+req.Name, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    rule,
		})
	}
}

func getFirewallRuleHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid rule id"))
			return
		}

		rule, err := cfg.DB.GetFirewallRuleByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "firewall rule not found"))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    rule,
		})
	}
}

type updateFirewallRuleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Port        string `json:"port"`
	Protocol    string `json:"protocol"`
	Source      string `json:"source"`
}

func updateFirewallRuleHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid rule id"))
			return
		}

		rule, err := cfg.DB.GetFirewallRuleByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "firewall rule not found"))
			return
		}

		var req updateFirewallRuleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		name := req.Name
		if name == "" {
			name = rule.Name
		}
		description := req.Description
		if description == "" && rule.Description != nil {
			description = *rule.Description
		}
		action := req.Action
		if action == "" {
			action = rule.Action
		}
		port := req.Port
		if port == "" {
			port = rule.Port
		}
		protocol := req.Protocol
		if protocol == "" {
			protocol = rule.Protocol
		}
		source := req.Source
		if source == "" {
			source = rule.Source
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "firewall.rule.update", map[string]interface{}{
			"old_port":   rule.Port,
			"old_source": rule.Source,
			"action":     action,
			"port":       port,
			"protocol":   protocol,
			"source":     source,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to update firewall rule"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		if err := cfg.DB.UpdateFirewallRule(c.Request.Context(), id, name, description, action, port, protocol, source); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "firewall rule updated"},
		})
	}
}

func deleteFirewallRuleHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid rule id"))
			return
		}

		rule, err := cfg.DB.GetFirewallRuleByID(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "firewall rule not found"))
			return
		}

		cfg.AgentClient.Call(c.Request.Context(), "firewall.rule.delete", map[string]interface{}{
			"port":   rule.Port,
			"source": rule.Source,
		})

		if err := cfg.DB.DeleteFirewallRule(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		userID := getUserID(c)
		cfg.DB.LogAudit(c.Request.Context(), &userID, "Deleted firewall rule: "+rule.Name, c.ClientIP(), c.Request.UserAgent(), "")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "firewall rule deleted"},
		})
	}
}