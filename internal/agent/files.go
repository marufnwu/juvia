package agent

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/base64"
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

	data, err := base64.StdEncoding.DecodeString(req.Content)
	if err != nil {
		data = []byte(req.Content)
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
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
		"content": base64.StdEncoding.EncodeToString(data),
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
	if ext == ".gz" {
		if strings.HasSuffix(strings.ToLower(safePath), ".tar.gz") {
			ext = ".tar.gz"
		} else if strings.HasSuffix(strings.ToLower(safePath), ".tgz") {
			ext = ".tgz"
		}
	}
	if ext != ".zip" && ext != ".tar.gz" && ext != ".tgz" {
		return nil, fmt.Errorf("unsupported archive type: %s", ext)
	}

	switch ext {
	case ".zip":
		if err := extractZip(safePath, safePath[:len(safePath)-len(filepath.Ext(safePath))]); err != nil {
			return nil, fmt.Errorf("extract zip: %w", err)
		}
	case ".tar.gz", ".tgz":
		if err := extractTarGz(safePath, safePath[:len(safePath)-len(filepath.Ext(safePath))]); err != nil {
			return nil, fmt.Errorf("extract tar.gz: %w", err)
		}
	}

	return map[string]interface{}{"extracted": true}, nil
}

func extractZip(archivePath, destDir string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		if err := extractZipEntry(file, destDir); err != nil {
			return err
		}
	}
	return nil
}

func extractZipEntry(file *zip.File, destDir string) error {
	path := filepath.Join(destDir, file.Name)

	if file.FileInfo().IsDir() {
		return os.MkdirAll(path, 0755)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	inFile, err := file.Open()
	if err != nil {
		return err
	}
	defer inFile.Close()

	_, err = io.Copy(outFile, inFile)
	return err
}

func extractTarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
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

func HandleFilesMkdir(ctx context.Context, params json.RawMessage) (interface{}, error) {
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

	fullPath, err := utils.SafePath(req.BasePath, req.RelPath)
	if err != nil {
		return nil, fmt.Errorf("path traversal detected: %w", err)
	}

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}

	return map[string]interface{}{"created": true}, nil
}
