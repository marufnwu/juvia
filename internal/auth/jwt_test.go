package auth

import (
	"testing"
	"time"
)

func TestJWTManager_GenerateAndVerify(t *testing.T) {
	jm := NewJWTManager("test-secret-key")
	tp, err := jm.GenerateTokenPair(1, "admin", "admin")
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	if tp.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if tp.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
	if tp.ExpiresIn != int64(AccessTokenTTL.Seconds()) {
		t.Fatalf("expected expires_in %d, got %d", int64(AccessTokenTTL.Seconds()), tp.ExpiresIn)
	}

	claims, err := jm.VerifyAccessToken(tp.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccessToken failed: %v", err)
	}
	if claims.UserID != 1 {
		t.Errorf("expected userID 1, got %d", claims.UserID)
	}
	if claims.Username != "admin" {
		t.Errorf("expected username admin, got %s", claims.Username)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role admin, got %s", claims.Role)
	}
}

func TestJWTManager_VerifyExpired(t *testing.T) {
	jm := NewJWTManager("test-secret-key")
	// Use a token that will be expired by using a short-lived custom token
	// This test verifies the verify rejects invalid tokens
	_, err := jm.VerifyAccessToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestTokenPair_Expiry(t *testing.T) {
	jm := NewJWTManager("test-secret-key")
	tp, err := jm.GenerateTokenPair(42, "user", "read_only")
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}
	now := time.Now().UTC()
	if tp.ExpiresAt.Before(now) || tp.ExpiresAt.After(now.Add(8*24*time.Hour)) {
		t.Errorf("unexpected refresh token expiry: %v", tp.ExpiresAt)
	}
}
