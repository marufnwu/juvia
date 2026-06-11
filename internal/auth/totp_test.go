package auth

import (
	"testing"
)

func TestTOTPManager_GenerateSecret(t *testing.T) {
	tm := NewTOTPManager()
	secret, uri, err := tm.GenerateSecret("juvia", "admin@example.com")
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}
	if secret == "" {
		t.Fatal("expected secret")
	}
	if uri == "" {
		t.Fatal("expected uri")
	}
}

func TestTOTPManager_GenerateRecoveryCodes(t *testing.T) {
	tm := NewTOTPManager()
	codes, err := tm.GenerateRecoveryCodes()
	if err != nil {
		t.Fatalf("GenerateRecoveryCodes failed: %v", err)
	}
	if len(codes) != 10 {
		t.Fatalf("expected 10 codes, got %d", len(codes))
	}
	for _, c := range codes {
		if c == "" {
			t.Fatal("expected non-empty code")
		}
	}
}
