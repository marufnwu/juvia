package api

import (
	"encoding/base64"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func listFilesHandler(cfg RouterConfig) gin.HandlerFunc {
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

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userID.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		relPath := c.Query("path")

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "files.list", map[string]interface{}{
			"base_path": website.DocumentRoot,
			"rel_path":  relPath,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to list files"))
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

func uploadFilesHandler(cfg RouterConfig) gin.HandlerFunc {
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

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userID.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "failed to parse multipart form: "+err.Error()))
			return
		}

		file, header, err := c.Request.FormFile("content")
		if err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "content file is required"))
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("SERVER_ERROR", "failed to read file content"))
			return
		}

		relPath := c.PostForm("path")
		fileName := header.Filename
		if overrideName := c.PostForm("file_name"); overrideName != "" {
			fileName = overrideName
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "files.upload", map[string]interface{}{
			"base_path": website.DocumentRoot,
			"rel_path":  relPath,
			"file_name": fileName,
			"content":   base64.StdEncoding.EncodeToString(data),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to upload file"))
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

func downloadFilesHandler(cfg RouterConfig) gin.HandlerFunc {
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

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userID.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		relPath := c.Query("path")
		if relPath == "" {
			var body struct {
				Path string `json:"path"`
			}
			if err := c.ShouldBindJSON(&body); err == nil {
				relPath = body.Path
			}
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "files.download", map[string]interface{}{
			"base_path": website.DocumentRoot,
			"rel_path":  relPath,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to download file"))
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

type renameFilesRequest struct {
	OldPath string `json:"old_path" binding:"required"`
	NewPath string `json:"new_path" binding:"required"`
}

func renameFilesHandler(cfg RouterConfig) gin.HandlerFunc {
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

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userID.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		var req renameFilesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "files.rename", map[string]interface{}{
			"base_path":   website.DocumentRoot,
			"old_rel_path": req.OldPath,
			"new_rel_path": req.NewPath,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to rename file"))
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

type deleteFilesRequest struct {
	Path string `json:"path" binding:"required"`
}

func deleteFilesHandler(cfg RouterConfig) gin.HandlerFunc {
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

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userID.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		var req deleteFilesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "files.delete", map[string]interface{}{
			"base_path": website.DocumentRoot,
			"rel_path":  req.Path,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to delete file"))
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

type extractFilesRequest struct {
	Path string `json:"path" binding:"required"`
}

func extractFilesHandler(cfg RouterConfig) gin.HandlerFunc {
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

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userID.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		var req extractFilesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "files.extract", map[string]interface{}{
			"base_path": website.DocumentRoot,
			"rel_path":  req.Path,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to extract archive"))
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

func getFileContentHandler(cfg RouterConfig) gin.HandlerFunc {
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

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userID.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		relPath := c.Query("path")
		if relPath == "" {
			var body struct {
				Path string `json:"path"`
			}
			if err := c.ShouldBindJSON(&body); err == nil {
				relPath = body.Path
			}
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "files.download", map[string]interface{}{
			"base_path": website.DocumentRoot,
			"rel_path":  relPath,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to get file content"))
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

type saveFileContentRequest struct {
	Path    string `json:"path" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func saveFileContentHandler(cfg RouterConfig) gin.HandlerFunc {
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

		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if role != "admin" && (website.UserID == nil || *website.UserID != userID.(int64)) {
			c.JSON(http.StatusForbidden, fail("FORBIDDEN", "You do not have access to this website"))
			return
		}

		var req saveFileContentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		resp, err := cfg.AgentClient.Call(c.Request.Context(), "files.upload", map[string]interface{}{
			"base_path": website.DocumentRoot,
			"rel_path":  req.Path,
			"file_name": "",
			"content":   req.Content,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to save file"))
			return
		}
		if resp.Error != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", resp.Error.Message))
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    gin.H{"message": "file saved"},
		})
	}
}

type createFolderRequest struct {
	Path string `json:"path" binding:"required"`
}

func createFolderHandler(cfg RouterConfig) gin.HandlerFunc {
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

		var req createFolderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, fail("VALIDATION_ERROR", "invalid request body"))
			return
		}

		_, err = cfg.AgentClient.Call(c.Request.Context(), "files.mkdir", map[string]interface{}{
			"base_path": website.DocumentRoot,
			"rel_path":  req.Path,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, fail("AGENT_ERROR", "failed to create folder"))
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    gin.H{"message": "folder created"},
		})
	}
}
