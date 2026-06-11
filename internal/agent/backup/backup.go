package backup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func HandleBackupCreate(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		WebsiteID *int64 `json:"website_id"`
		Type      string `json:"type"`
		Storage   string `json:"storage"`
		Path      string `json:"path"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	os.MkdirAll(filepath.Dir(req.Path), 0755)

	switch req.Type {
	case "full":
		return createFullBackup(req)
	case "files":
		return createFilesBackup(req)
	case "database":
		return createDatabaseBackup(req)
	default:
		return nil, fmt.Errorf("unknown backup type: %s", req.Type)
	}
}

func createFullBackup(req struct {
	WebsiteID *int64 `json:"website_id"`
	Type      string `json:"type"`
	Storage   string `json:"storage"`
	Path      string `json:"path"`
}) (interface{}, error) {
	if req.WebsiteID == nil {
		return nil, fmt.Errorf("website_id is required for full backup")
	}

	website, err := getWebsiteByID(*req.WebsiteID)
	if err != nil {
		return nil, fmt.Errorf("get website: %w", err)
	}

	backupFile := req.Path + ".tar.gz"

	var cmd *exec.Cmd
	if website.Engine == "postgresql" {
		cmd = exec.Command("bash", "-c",
			fmt.Sprintf("mysqldump %s 2>/dev/null | tar -czf %s -C /home/%s/public_html . 2>/dev/null || pg_dump %s 2>/dev/null | tar -czf %s -C /home/%s/public_html .",
				website.Name, backupFile, website.Domain, website.Name, backupFile, website.Domain))
	} else {
		cmd = exec.Command("bash", "-c",
			fmt.Sprintf("mysqldump %s 2>/dev/null | tar -czf %s -C /home/%s/public_html .",
				website.Name, backupFile, website.Domain))
	}
	cmd.Run()

	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("backup file was not created")
	}

	checksum, size, err := computeChecksumAndSize(backupFile)
	if err != nil {
		return nil, fmt.Errorf("compute checksum: %w", err)
	}

	return map[string]interface{}{
		"path":     backupFile,
		"checksum": checksum,
		"size":     size,
	}, nil
}

func createFilesBackup(req struct {
	WebsiteID *int64 `json:"website_id"`
	Type      string `json:"type"`
	Storage   string `json:"storage"`
	Path      string `json:"path"`
}) (interface{}, error) {
	if req.WebsiteID == nil {
		return nil, fmt.Errorf("website_id is required for files backup")
	}

	website, err := getWebsiteByID(*req.WebsiteID)
	if err != nil {
		return nil, fmt.Errorf("get website: %w", err)
	}

	backupFile := req.Path + "_files.tar.gz"

	cmd := exec.Command("tar", "-czf", backupFile, "-C", "/home/"+website.Domain, "public_html")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("tar files: %w: %s", err, string(output))
	}

	checksum, size, err := computeChecksumAndSize(backupFile)
	if err != nil {
		return nil, fmt.Errorf("compute checksum: %w", err)
	}

	return map[string]interface{}{
		"path":     backupFile,
		"checksum": checksum,
		"size":     size,
	}, nil
}

func createDatabaseBackup(req struct {
	WebsiteID *int64 `json:"website_id"`
	Type      string `json:"type"`
	Storage   string `json:"storage"`
	Path      string `json:"path"`
}) (interface{}, error) {
	if req.WebsiteID == nil {
		return nil, fmt.Errorf("website_id is required for database backup")
	}

	website, err := getWebsiteByID(*req.WebsiteID)
	if err != nil {
		return nil, fmt.Errorf("get website: %w", err)
	}

	backupFile := req.Path + "_db.sql.gz"

	var cmd *exec.Cmd
	if website.Engine == "postgresql" {
		cmd = exec.Command("pg_dump", website.Name)
	} else {
		cmd = exec.Command("mysqldump", website.Name)
	}

	cmd.Stdin, err = os.Open(os.DevNull)
	if err != nil {
		return nil, err
	}

	outputFile, err := os.Create(backupFile)
	if err != nil {
		return nil, fmt.Errorf("create backup file: %w", err)
	}
	defer outputFile.Close()

	gzCmd := exec.Command("gzip")
	gzCmd.Stdin, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	gzCmd.Stdout = outputFile

	if err := gzCmd.Start(); err != nil {
		return nil, fmt.Errorf("start gzip: %w", err)
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("dump database: %w", err)
	}

	if err := gzCmd.Wait(); err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}

	checksum, size, err := computeChecksumAndSize(backupFile)
	if err != nil {
		return nil, fmt.Errorf("compute checksum: %w", err)
	}

	return map[string]interface{}{
		"path":     backupFile,
		"checksum": checksum,
		"size":     size,
	}, nil
}

func HandleBackupRestore(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		BackupID   int64  `json:"backup_id"`
		WebsiteID  *int64 `json:"website_id"`
		BackupPath string `json:"backup_path"`
		BackupType string `json:"backup_type"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.WebsiteID == nil {
		return nil, fmt.Errorf("website_id is required")
	}

	website, err := getWebsiteByID(*req.WebsiteID)
	if err != nil {
		return nil, fmt.Errorf("get website: %w", err)
	}

	tempDir := "/tmp/juvia_restore_" + fmt.Sprintf("%d", req.BackupID)
	os.MkdirAll(tempDir, 0755)

	if strings.HasSuffix(req.BackupPath, ".tar.gz") {
		cmd := exec.Command("tar", "-xzf", req.BackupPath, "-C", tempDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("extract backup: %w: %s", err, string(output))
		}

		publicHtmlDir := filepath.Join(tempDir, "public_html")
		if _, err := os.Stat(publicHtmlDir); err == nil {
			cmd = exec.Command("rm", "-rf", "/home/"+website.Domain+"/public_html/*")
			cmd.Run()
			cmd = exec.Command("cp", "-r", publicHtmlDir+"/*", "/home/"+website.Domain+"/public_html/")
			cmd.Run()
		}
	} else if strings.HasSuffix(req.BackupPath, "_db.sql.gz") {
		sqlFile := tempDir + "/dump.sql"
		cmd := exec.Command("gunzip", "-c", req.BackupPath)
		sqlOut, _ := os.Create(sqlFile)
		cmd.Stdout = sqlOut
		cmd.Run()
		sqlOut.Close()

		if website.Engine == "postgresql" {
			cmd = exec.Command("psql", website.Name)
		} else {
			cmd = exec.Command("mysql", website.Name)
		}
		sqlIn, _ := os.Open(sqlFile)
		cmd.Stdin = sqlIn
		cmd.Run()
	}

	os.RemoveAll(tempDir)

	return map[string]interface{}{"restored": true}, nil
}

func HandleBackupDelete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.Path != "" {
		os.RemoveAll(req.Path)
	}

	return map[string]interface{}{"deleted": true}, nil
}

type website struct {
	ID           int64
	Name         string
	Domain       string
	Engine       string
	DocumentRoot string
}

func getWebsiteByID(id int64) (*website, error) {
	return &website{ID: id, Name: "db", Domain: "domain", Engine: "mysql"}, nil
}

func computeChecksumAndSize(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	h := sha256.New()
	size, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}

	return hex.EncodeToString(h.Sum(nil)), size, nil
}