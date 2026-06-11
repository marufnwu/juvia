package git

import (
	"context"
	"testing"
)

func TestHandleGitSetup_InvalidParams(t *testing.T) {
	_, err := HandleGitSetup(context.Background(), []byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleGitSetup_EmptyParams(t *testing.T) {
	params := []byte(`{}`)
	_, err := HandleGitSetup(context.Background(), params)
	if err == nil {
		t.Error("expected error for empty params")
	}
}

func TestHandleGitPull_InvalidParams(t *testing.T) {
	_, err := HandleGitPull(context.Background(), []byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleGitPull_EmptyWebsitePath(t *testing.T) {
	params := []byte(`{"website_id": 1, "website_path": ""}`)
	_, err := HandleGitPull(context.Background(), params)
	if err == nil {
		t.Error("expected error for empty website path")
	}
}

func TestHandleGitSetup_EmptyRepoURL(t *testing.T) {
	params := []byte(`{"website_id": 1, "website_path": "/tmp/test", "repo_url": "", "branch": "main"}`)
	_, err := HandleGitSetup(context.Background(), params)
	if err == nil {
		t.Error("expected error for empty repo URL")
	}
}

func TestHandleGitSetup_EmptyWebsitePath(t *testing.T) {
	params := []byte(`{"website_id": 1, "website_path": "", "repo_url": "https://github.com/user/repo.git", "branch": "main"}`)
	_, err := HandleGitSetup(context.Background(), params)
	if err == nil {
		t.Error("expected error for empty website path")
	}
}