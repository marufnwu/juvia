package socket

import (
	"testing"
)

func TestNewResponse(t *testing.T) {
	resp := NewResponse("req_123", map[string]bool{"pong": true})
	if resp.ID != "req_123" {
		t.Errorf("expected ID req_123, got %s", resp.ID)
	}
	if resp.Error != nil {
		t.Error("expected no error")
	}
	if resp.Result == nil {
		t.Error("expected result")
	}
}

func TestNewErrorResponse(t *testing.T) {
	resp := NewErrorResponse("req_456", "NOT_FOUND", "resource missing")
	if resp.ID != "req_456" {
		t.Errorf("expected ID req_456, got %s", resp.ID)
	}
	if resp.Result != nil {
		t.Error("expected nil result")
	}
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %s", resp.Error.Code)
	}
	if resp.Error.Message != "resource missing" {
		t.Errorf("expected message 'resource missing', got %s", resp.Error.Message)
	}
}
