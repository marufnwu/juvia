package apps

import (
	"context"
	"encoding/json"
	"testing"
)

func TestHandleAppCheckUpdate_InvalidParams(t *testing.T) {
	_, err := HandleAppCheckUpdate(context.Background(), []byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleAppCheckUpdate_EmptyParams(t *testing.T) {
	params := []byte(`{}`)
	result, err := HandleAppCheckUpdate(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result")
	}

	updateAvailable, ok := data["update_available"].(bool)
	if !ok {
		t.Fatal("expected update_available field")
	}

	if updateAvailable {
		t.Error("update_available should be false for empty params")
	}
}

func TestHandleAppUpdate_InvalidParams(t *testing.T) {
	_, err := HandleAppUpdate(context.Background(), []byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleAppUpdate_UnsupportedAppType(t *testing.T) {
	params := []byte(`{"website_id": 1, "website_path": "/tmp", "app_type": "unsupported"}`)
	_, err := HandleAppUpdate(context.Background(), params)
	if err == nil {
		t.Error("expected error for unsupported app type")
	}
}

func TestHandleAppInstall_InvalidParams(t *testing.T) {
	_, err := HandleAppInstall(context.Background(), []byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleAppInstall_EmptyWebsitePath(t *testing.T) {
	params := []byte(`{"website_id": 1, "website_path": "", "app_type": "wordpress"}`)
	_, err := HandleAppInstall(context.Background(), params)
	if err == nil {
		t.Error("expected error for empty website path")
	}
}

func TestHandleAppInstall_UnsupportedAppType(t *testing.T) {
	params := []byte(`{"website_id": 1, "website_path": "/tmp/test", "app_type": "unsupported"}`)
	_, err := HandleAppInstall(context.Background(), params)
	if err == nil {
		t.Error("expected error for unsupported app type")
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

func TestRandomString_DifferentLengths(t *testing.T) {
	s8 := randomString(8)
	s16 := randomString(16)
	s32 := randomString(32)

	if len(s8) != 8 || len(s16) != 16 || len(s32) != 32 {
		t.Error("randomString should return strings of requested length")
	}
}

func TestAppInstallParams_Unmarshal(t *testing.T) {
	params := appInstallParams{
		WebsiteID:     1,
		WebsitePath:   "/var/www/example.com",
		AppType:       "wordpress",
		AdminUsername: "admin",
		AdminPassword: "secret",
		DBName:        "wp_db",
		DBPassword:    "dbpass",
	}

	data, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded appInstallParams
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.WebsiteID != params.WebsiteID {
		t.Error("WebsiteID mismatch")
	}
	if decoded.WebsitePath != params.WebsitePath {
		t.Error("WebsitePath mismatch")
	}
	if decoded.AppType != params.AppType {
		t.Error("AppType mismatch")
	}
}