package api

import (
	"bufio"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func accessLogHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		lines, _ := strconv.Atoi(c.DefaultQuery("lines", "100"))

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), parseID(id))
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		logPath := "/var/log/juvia/nginx/" + website.Domain + ".access.log"

		content, err := readLastLines(logPath, lines)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"lines":  []string{},
					"path":   logPath,
					"message": "log file not found",
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"lines": content,
				"path":  logPath,
			},
		})
	}
}

func errorLogHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		lines, _ := strconv.Atoi(c.DefaultQuery("lines", "100"))

		website, err := cfg.DB.GetWebsiteByID(c.Request.Context(), parseID(id))
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "website not found"))
			return
		}

		logPath := "/var/log/juvia/nginx/" + website.Domain + ".error.log"

		content, err := readLastLines(logPath, lines)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"lines":  []string{},
					"path":   logPath,
					"message": "log file not found",
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"lines": content,
				"path":  logPath,
			},
		})
	}
}

func systemLogsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		lines, _ := strconv.Atoi(c.DefaultQuery("lines", "100"))
		logType := c.DefaultQuery("type", "panel")

		var logPath string
		if logType == "agent" {
			logPath = "/var/log/juvia/agent.log"
		} else {
			logPath = "/var/log/juvia/panel.log"
		}

		content, err := readLastLines(logPath, lines)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"lines":  []string{},
					"path":   logPath,
					"message": "log file not found",
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"lines": content,
				"path":  logPath,
			},
		})
	}
}

func readLastLines(path string, n int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return []string{}, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	return lines, nil
}

func parseID(s string) int64 {
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}