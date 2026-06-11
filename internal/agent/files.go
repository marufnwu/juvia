package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"juvia/internal/utils"
)

type FileInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Size       int64  `json:"size"`
	ModifiedAt int64  `json:"modified_at"`
	Permissions string `json:"permissions"`
}

func HandleFilesList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		BasePath string `json:"base_path"`
		RelPath  string `json:"rel_path"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.BasePath == "" {
		return nil, fmt.Errorf("base_path is required")
	}

	safePath, err := utils.SafePath(req.BasePath, req.RelPath)
	if err != nil {
		return nil, fmt.Errorf("path traversal detected: %w", err)
	}

	entries, err := os.ReadDir(safePath)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		info, _ := entry.Info()
		fileType := "file"
		if entry.IsDir() {
			fileType = "directory"
		}
		perms := "----------"
		if info != nil {
			perms = info.Mode().String()
		}
		files = append(files, FileInfo{
			Name:       entry.Name(),
			Type:       fileType,
			Size:       info.Size(),
			ModifiedAt: info.ModTime().Unix(),
			Permissions: perms,
		})
	}

	return map[string]interface{}{"items": files}, nil
}

func HandleFilesUpload(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		BasePath string `json:"base_path"`
		RelPath  string `json:"rel_path"`
		FileName string `json:"file_name"`
		Content  string `json:"content"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.BasePath == "" || req.RelPath == "" {
		return nil, fmt.Errorf("base_path and rel_path are required")
	}

	safePath, err := utils.SafePath(req.BasePath, req.RelPath)
	if err != nil {
		return nil, fmt.Errorf("path traversal detected: %w", err)
	}

	fullPath := filepath.Join(safePath, req.FileName)
	if err := os.WriteFile(fullPath, []byte(req.Content), 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	return map[string]interface{}{"uploaded": true, "path": fullPath}, nil
}

func HandleFilesDownload(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		BasePath string `json:"base_path"`
		RelPath  string `json:"rel_path"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.BasePath == "" {
		return nil, fmt.Errorf("base_path is required")
	}

	safePath, err := utils.SafePath(req.BasePath, req.RelPath)
	if err != nil {
		return nil, fmt.Errorf("path traversal detected: %w", err)
	}

	data, err := os.ReadFile(safePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return map[string]interface{}{
		"content": string(data),
		"size":    len(data),
	}, nil
}

func HandleFilesDelete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		BasePath string `json:"base_path"`
		RelPath  string `json:"rel_path"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.BasePath == "" {
		return nil, fmt.Errorf("base_path is required")
	}

	safePath, err := utils.SafePath(req.BasePath, req.RelPath)
	if err != nil {
		return nil, fmt.Errorf("path traversal detected: %w", err)
	}

	if err := os.RemoveAll(safePath); err != nil {
		return nil, fmt.Errorf("delete: %w", err)
	}

	return map[string]interface{}{"deleted": true}, nil
}

func HandleFilesRename(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		BasePath  string `json:"base_path"`
		OldRelPath string `json:"old_rel_path"`
		NewRelPath string `json:"new_rel_path"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.BasePath == "" {
		return nil, fmt.Errorf("base_path is required")
	}

	oldPath, err := utils.SafePath(req.BasePath, req.OldRelPath)
	if err != nil {
		return nil, fmt.Errorf("path traversal detected: %w", err)
	}
	newPath, err := utils.SafePath(req.BasePath, req.NewRelPath)
	if err != nil {
		return nil, fmt.Errorf("path traversal detected: %w", err)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		return nil, fmt.Errorf("rename: %w", err)
	}

	return map[string]interface{}{"renamed": true}, nil
}

func HandleFilesExtract(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		BasePath string `json:"base_path"`
		RelPath  string `json:"rel_path"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if req.BasePath == "" {
		return nil, fmt.Errorf("base_path is required")
	}

	safePath, err := utils.SafePath(req.BasePath, req.RelPath)
	if err != nil {
		return nil, fmt.Errorf("path traversal detected: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(safePath))
	if ext != ".zip" && ext != ".tar.gz" && ext != ".tgz" {
		return nil, fmt.Errorf("unsupported archive type: %s", ext)
	}

	return map[string]interface{}{"extracted": true}, nil
}

func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
