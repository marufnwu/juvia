package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"juvia/internal/metrics"
)

func currentMetricsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		collector := metrics.NewCollector()
		m, err := collector.Collect()
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    m,
		})
	}
}

func historyMetricsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"message": "historical metrics require VictoriaMetrics (Phase 5)",
			},
		})
	}
}