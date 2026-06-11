package git

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func HandleGitSetup(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		WebsiteID   int64  `json:"website_id"`
		WebsitePath string `json:"website_path"`
		RepoURL    string `json:"repo_url"`
		Branch     string `json:"branch"`
		DeployKey  string `json:"deploy_key"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.WebsitePath == "" || req.RepoURL == "" {
		return nil, fmt.Errorf("website path and repo URL are required")
	}

	os.MkdirAll(req.WebsitePath, 0755)

	if req.Branch == "" {
		req.Branch = "main"
	}

	var cmd *exec.Cmd
	if req.DeployKey != "" {
		cmd = exec.Command("git", "clone", "-b", req.Branch, "--single-branch", req.RepoURL, req.WebsitePath)
	} else {
		cmd = exec.Command("git", "clone", "-b", req.Branch, "--single-branch", req.RepoURL, req.WebsitePath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git clone failed: %w: %s", err, string(output))
	}

	cmd = exec.Command("chown", "-R", "www-data:www-data", req.WebsitePath)
	cmd.Run()

	return map[string]interface{}{
		"cloned": true,
		"branch": req.Branch,
	}, nil
}

func HandleGitPull(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		WebsiteID   int64  `json:"website_id"`
		WebsitePath string `json:"website_path"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	if req.WebsitePath == "" {
		return nil, fmt.Errorf("website path is required")
	}

	gitDir := filepath.Join(req.WebsitePath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "fetch", "origin")
	cmd.Dir = req.WebsitePath
	cmd.Run()

	cmd = exec.Command("git", "reset", "--hard", "origin/main")
	cmd.Dir = req.WebsitePath
	output, err := cmd.CombinedOutput()
	if err != nil {
		cmd = exec.Command("git", "reset", "--hard", "origin/master")
		cmd.Dir = req.WebsitePath
		output, err = cmd.CombinedOutput()
	}

	cmd = exec.Command("chown", "-R", "www-data:www-data", req.WebsitePath)
	cmd.Run()

	return map[string]interface{}{
		"pulled": true,
		"output": string(output),
	}, nil
}