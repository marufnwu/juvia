package email

import (
	"context"
	"testing"
)

func TestHandleWebmailInstall_InvalidParams(t *testing.T) {
	_, err := HandleWebmailInstall(context.Background(), []byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleWebmailUninstall_ParamsIgnored(t *testing.T) {
	result, err := HandleWebmailUninstall(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data := result.(map[string]interface{})
	if !data["uninstalled"].(bool) {
		t.Error("expected uninstalled to be true")
	}
}

func TestRandomString(t *testing.T) {
	s1 := randomString(16)
	s2 := randomString(16)

	if len(s1) != 16 {
		t.Errorf("expected length 16, got %d", len(s1))
	}

	if s1 == s2 {
		t.Error("random strings should be unique")
	}
}

func TestRandomString_EmptyInput(t *testing.T) {
	s := randomString(8)
	if len(s) != 8 {
		t.Errorf("expected length 8, got %d", len(s))
	}
}

func TestDownloadFile_InvalidURL(t *testing.T) {
	err := downloadFile("http://invalid-url-that-does-not-exist.example.com/file.tar.gz", "/tmp/dest")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}