package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"juvia/internal/updater"
)

func checkUpdateHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		info, err := updater.CheckForUpdates()
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    info,
		})
	}
}

type downloadUpdateRequest struct {
	URL string `json:"url"`
}

func downloadUpdateHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req downloadUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "url is required"))
			return
		}

		path, err := updater.DownloadUpdate(req.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("DOWNLOAD_FAILED", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"path": path,
			},
		})
	}
}

type applyUpdateRequest struct {
	Path string `json:"path"`
	SigURL string `json:"sig_url"`
}

func applyUpdateHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req applyUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "path is required"))
			return
		}

		if req.SigURL != "" {
			if err := updater.VerifyGPGSignature(req.Path, req.SigURL); err != nil {
				c.JSON(http.StatusBadRequest, fail("SIGNATURE_INVALID", err.Error()))
				return
			}
		}

		if err := updater.ApplyUpdate(req.Path); err != nil {
			c.JSON(http.StatusInternalServerError, fail("UPDATE_FAILED", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"message": "update applied, restarting",
			},
		})
	}
}

func rollbackHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := updater.Rollback(); err != nil {
			c.JSON(http.StatusInternalServerError, fail("ROLLBACK_FAILED", err.Error()))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"message": "rollback applied, restarting",
			},
		})
	}
}

func getVersionHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"version": updater.GetCurrentVersion(),
			},
		})
	}
}