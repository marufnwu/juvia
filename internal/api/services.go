package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"juvia/internal/socket"
)

type Service struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func listServicesHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		services := []Service{
			{Name: "nginx", Description: "Web server", Status: "unknown"},
			{Name: "php-fpm", Description: "PHP FastCGI Process Manager", Status: "unknown"},
			{Name: "mysql", Description: "MySQL database server", Status: "unknown"},
			{Name: "postfix", Description: "Mail transfer agent", Status: "unknown"},
			{Name: "dovecot", Description: "IMAP/POP3 mail server", Status: "unknown"},
			{Name: "named", Description: "BIND DNS server", Status: "unknown"},
			{Name: "ufw", Description: "Uncomplicated Firewall", Status: "unknown"},
		}

		for i := range services {
			status, err := getServiceStatus(cfg.AgentClient, services[i].Name)
			if err == nil {
				services[i].Status = status
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    services,
		})
	}
}

func getServiceHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := c.Param("name")

		validServices := map[string]string{
			"nginx":   "Web server",
			"php-fpm": "PHP FastCGI Process Manager",
			"mysql":    "MySQL database server",
			"postfix":  "Mail transfer agent",
			"dovecot":  "IMAP/POP3 mail server",
			"named":    "BIND DNS server",
			"ufw":      "Uncomplicated Firewall",
		}

		description, ok := validServices[serviceName]
		if !ok {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "service not found"))
			return
		}

		status, _ := getServiceStatus(cfg.AgentClient, serviceName)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": Service{
				Name:        serviceName,
				Description: description,
				Status:      status,
			},
		})
	}
}

func restartServiceHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := c.Param("name")

		validServices := map[string]bool{
			"nginx": true, "php-fpm": true, "mysql": true,
			"postfix": true, "dovecot": true, "named": true,
		}

		if !validServices[serviceName] {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "service not found"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "services.restart", map[string]interface{}{
			"service": serviceName,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to restart service"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "service restarted"},
		})
	}
}

func getServiceStatus(agentClient *socket.Client, serviceName string) (string, error) {
	resp, err := agentClient.Call(context.Background(), "services.status", map[string]interface{}{
		"service": serviceName,
	})
	if err != nil || resp.Error != nil {
		return "unknown", err
	}
	if result, ok := resp.Result.(map[string]interface{}); ok {
		if status, ok := result["status"].(string); ok {
			return status, nil
		}
	}
	return "unknown", nil
}