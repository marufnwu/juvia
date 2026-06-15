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
	Installed   bool   `json:"installed"`
	ErrorDetail string `json:"error_detail,omitempty"`
	FixHint     string `json:"fix_hint,omitempty"`
}

type serviceInfo struct {
	Description string
	Command     string
	FixInstall  string
	FixStart    string
	FixConfig   string
}

var serviceRegistry = map[string]serviceInfo{
	"nginx": {
		Description: "Web server · Serves your websites to visitors",
		Command:     "nginx",
		FixInstall:  "sudo apt install -y nginx",
		FixStart:    "sudo systemctl start nginx",
		FixConfig:   "sudo nginx -t",
	},
	"php-fpm": {
		Description: "PHP processor · Runs your website's PHP code",
		Command:     "php-fpm",
		FixInstall:  "sudo apt install -y php-fpm",
		FixStart:    "sudo systemctl start php*-fpm",
		FixConfig:   "",
	},
	"mysql": {
		Description: "Database server · Stores website data",
		Command:     "mysqld",
		FixInstall:  "sudo apt install -y mysql-server",
		FixStart:    "sudo systemctl start mysql",
		FixConfig:   "",
	},
	"postfix": {
		Description: "Outgoing mail · Sends email from your server",
		Command:     "postfix",
		FixInstall:  "sudo apt install -y postfix",
		FixStart:    "sudo systemctl start postfix",
		FixConfig:   "",
	},
	"dovecot": {
		Description: "Mail storage · Handles incoming email & IMAP",
		Command:     "dovecot",
		FixInstall:  "sudo apt install -y dovecot-imapd",
		FixStart:    "sudo systemctl start dovecot",
		FixConfig:   "",
	},
	"named": {
		Description: "DNS server · Makes your server authoritative for your domains",
		Command:     "named",
		FixInstall:  "sudo apt install -y bind9 bind9utils",
		FixStart:    "sudo systemctl start bind9",
		FixConfig:   "sudo named-checkconf",
	},
	"ufw": {
		Description: "Firewall · Controls which ports are open",
		Command:     "ufw",
		FixInstall:  "sudo apt install -y ufw",
		FixStart:    "sudo ufw enable",
		FixConfig:   "",
	},
	"certbot": {
		Description: "SSL certificates · Encrypts your websites with HTTPS",
		Command:     "certbot",
		FixInstall:  "sudo snap install certbot --classic",
		FixStart:    "",
		FixConfig:   "",
	},
	"rspamd": {
		Description: "Spam filter · Protects against email spam",
		Command:     "rspamd",
		FixInstall:  "sudo apt install -y rspamd",
		FixStart:    "sudo systemctl start rspamd",
		FixConfig:   "",
	},
}

func listServicesHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var services []Service

		for name, info := range serviceRegistry {
			svc := Service{
				Name:        name,
				Description: info.Description,
				Status:      "unknown",
				Installed:   false,
			}

			status, err := getServiceStatus(cfg.AgentClient, name)
			if err == nil {
				svc.Status = status
			}

			installed, _ := checkServiceInstalled(cfg.AgentClient, name, info.Command)
			svc.Installed = installed

			if !installed {
				svc.ErrorDetail = info.Command + " is not installed"
				svc.FixHint = info.FixInstall
			} else if status == "inactive" {
				svc.ErrorDetail = info.Command + " is installed but not running"
				svc.FixHint = info.FixStart
			} else if status == "failed" {
				svc.ErrorDetail = info.Command + " has crashed or has a config error"
				svc.FixHint = info.FixConfig
			}

			services = append(services, svc)
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

		info, ok := serviceRegistry[serviceName]
		if !ok {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "service not found"))
			return
		}

		status, _ := getServiceStatus(cfg.AgentClient, serviceName)
		installed, _ := checkServiceInstalled(cfg.AgentClient, serviceName, info.Command)

		svc := Service{
			Name:        serviceName,
			Description: info.Description,
			Status:      status,
			Installed:   installed,
		}

		if !installed {
			svc.ErrorDetail = info.Command + " is not installed"
			svc.FixHint = info.FixInstall
		} else if status == "inactive" {
			svc.ErrorDetail = info.Command + " is installed but not running"
			svc.FixHint = info.FixStart
		} else if status == "failed" {
			svc.ErrorDetail = info.Command + " has crashed or has a config error"
			svc.FixHint = info.FixConfig
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    svc,
		})
	}
}

func restartServiceHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceName := c.Param("name")

		if _, ok := serviceRegistry[serviceName]; !ok {
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

func checkServiceInstalled(agentClient *socket.Client, serviceName string, command string) (bool, error) {
	resp, err := agentClient.Call(context.Background(), "services.check_installed", map[string]interface{}{
		"service": serviceName,
		"command": command,
	})
	if err != nil || resp.Error != nil {
		return false, err
	}
	if result, ok := resp.Result.(map[string]interface{}); ok {
		if installed, ok := result["installed"].(bool); ok {
			return installed, nil
		}
	}
	return false, nil
}