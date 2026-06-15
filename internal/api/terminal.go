package api

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type TerminalSession struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type TerminalRecording struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	UserID    int64     `json:"user_id"`
	StartedAt time.Time `json:"started_at"`
	Duration  int64      `json:"duration_seconds"`
	SizeBytes int64     `json:"size_bytes"`
}

func createTerminalSessionHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := cfg.AgentClient.Call(c.Request.Context(), "terminal.start", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to create session"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		result := resp.Result.(map[string]interface{})
		sessionID, _ := result["session_id"].(string)
		recording, _ := result["recording"].(string)

		data := gin.H{"session_id": sessionID}
		if recording != "" {
			data["recording"] = recording
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    data,
		})
	}
}

func listTerminalSessionsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []TerminalSession{},
		})
	}
}

func closeTerminalSessionHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "terminal.stop", map[string]interface{}{
			"session_id": sessionID,
		})

		data := gin.H{"message": "session closed"}
		if err == nil && resp != nil && resp.Error == nil {
			if result, ok := resp.Result.(map[string]interface{}); ok {
				if recording, ok := result["recording"].(string); ok && recording != "" {
					data["recording"] = recording
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    data,
		})
	}
}

func listTerminalRecordingsHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		recordingsDir := "/var/log/juvia/terminal"

		entries, err := os.ReadDir(recordingsDir)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    []TerminalRecording{},
			})
			return
		}

		var recordings []TerminalRecording
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".cast" {
				continue
			}

			info, _ := entry.Info()
			recordings = append(recordings, TerminalRecording{
				ID:         entry.Name(),
				Filename:   entry.Name(),
				Duration:   0,
				SizeBytes:  info.Size(),
				StartedAt:  info.ModTime(),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    recordings,
		})
	}
}

func getTerminalRecordingHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		recordingsDir := "/var/log/juvia/terminal"
		filePath := filepath.Join(recordingsDir, id)

		data, err := os.ReadFile(filePath)
		if err != nil {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "recording not found"))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"filename": id,
				"content":  string(data),
				"size":    len(data),
			},
		})
	}
}

func deleteTerminalRecordingHandler(cfg RouterConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		recordingsDir := "/var/log/juvia/terminal"
		filePath := filepath.Join(recordingsDir, id)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, fail("NOT_FOUND", "recording not found"))
			return
		}

		if err := os.Remove(filePath); err != nil {
			c.JSON(http.StatusInternalServerError, fail("DELETE_FAILED", "failed to delete recording"))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "recording deleted"},
		})
	}
}