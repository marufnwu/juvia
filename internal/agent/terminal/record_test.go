package terminal

import (
	"context"
	"encoding/json"
	"testing"
)

func TestHandleTerminalStart(t *testing.T) {
	result, err := HandleTerminalStart(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result")
	}

	sessionID, ok := data["session_id"].(string)
	if !ok {
		t.Fatal("expected session_id field")
	}

	if sessionID == "" {
		t.Error("expected non-empty session ID")
	}
}

func TestHandleTerminalStart_TwoSessions(t *testing.T) {
	result1, _ := HandleTerminalStart(context.Background(), nil)
	result2, _ := HandleTerminalStart(context.Background(), nil)

	id1 := result1.(map[string]interface{})["session_id"].(string)
	id2 := result2.(map[string]interface{})["session_id"].(string)

	if id1 == id2 {
		t.Error("session IDs should be unique")
	}
}

func TestHandleTerminalStop(t *testing.T) {
	startResult, _ := HandleTerminalStart(context.Background(), nil)
	sessionID := startResult.(map[string]interface{})["session_id"].(string)

	params := json.RawMessage(`{"session_id": "` + sessionID + `"}`)
	result, err := HandleTerminalStop(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("expected map result")
	}

	stopped, ok := data["stopped"].(bool)
	if !ok {
		t.Fatal("expected stopped field")
	}

	if !stopped {
		t.Error("expected stopped to be true")
	}
}

func TestHandleTerminalStop_InvalidParams(t *testing.T) {
	_, err := HandleTerminalStop(context.Background(), []byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleTerminalStop_NonExistentSession(t *testing.T) {
	params := json.RawMessage(`{"session_id": "nonexistent-session-id"}`)
	result, err := HandleTerminalStop(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := result.(map[string]interface{})
	if !data["stopped"].(bool) {
		t.Error("expected stopped to be true even for non-existent session")
	}
}

func TestGenerateSessionID(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()

	if id1 == id2 {
		t.Error("generated IDs should be unique")
	}

	if len(id1) != 32 {
		t.Errorf("expected length 32, got %d", len(id1))
	}
}