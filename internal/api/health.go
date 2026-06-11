package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func healthHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := "ok"
		sqliteStatus := "ok"
		agentStatus := "ok"

		if err := cfg.DB.PingContext(c); err != nil {
			sqliteStatus = "error"
			status = "degraded"
		}

		if cfg.AgentClient != nil && cfg.AgentClient.IsOffline() {
			agentStatus = "offline"
			status = "degraded"
		}

		c.JSON(http.StatusOK, success(gin.H{
			"status":            status,
			"version":           "1.0.0",
			"uptime":            time.Now().Unix() - startTime,
			"sqlite":            sqliteStatus,
			"agent":             agentStatus,
			"victoria_metrics":  "ok",
		}))
	}
}

var startTime = time.Now().Unix()
