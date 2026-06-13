package terminal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var activeSessions = make(map[string]*sessionInfo)

type sessionInfo struct {
	ID        string
	UserID    int64
	StartTime time.Time
	StopTime  time.Time
}

func HandleTerminalStart(ctx context.Context, params json.RawMessage) (interface{}, error) {
	sessionID := generateSessionID()

	activeSessions[sessionID] = &sessionInfo{
		ID:        sessionID,
		StartTime: time.Now(),
	}

	recordingsDir := "/var/log/juvia/terminal"
	os.MkdirAll(recordingsDir, 0755)

	filename := sessionID + ".cast"
	recordingPath := filepath.Join(recordingsDir, filename)

	f, err := os.Create(recordingPath)
	if err == nil {
		f.Close()
		os.Chmod(recordingPath, 0644)
	}

	return map[string]interface{}{
		"session_id": sessionID,
		"recording":  filename,
	}, nil
}

func HandleTerminalStop(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	var recordingFilename string
	if sess, ok := activeSessions[req.SessionID]; ok {
		sess.StopTime = time.Now()
		delete(activeSessions, req.SessionID)
	}

	recordingsDir := "/var/log/juvia/terminal"
	filename := req.SessionID + ".cast"
	recordingPath := filepath.Join(recordingsDir, filename)

	if _, err := os.Stat(recordingPath); err == nil {
		recordingFilename = filename
	}

	return map[string]interface{}{
		"stopped":   true,
		"recording": recordingFilename,
	}, nil
}

func startRecording(sessionID string, userID int64, outputPath string) error {
	recordingsDir := "/var/log/juvia/terminal"
	os.MkdirAll(recordingsDir, 0755)

	filename := fmt.Sprintf("%s_%s_%s.cast",
		time.Now().Format("2006-01-02_15-04-05"),
		"user",
		sessionID)
	recordingPath := filepath.Join(recordingsDir, filename)

	cmd := exec.Command("script", "-q", "-c", "/bin/bash", "-e", recordingPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SESSION_ID=%s", sessionID),
		fmt.Sprintf("USER_ID=%d", userID))

	return cmd.Start()
}

func stopRecording(sessionID string) (string, error) {
	if sess, ok := activeSessions[sessionID]; ok {
		sess.StopTime = time.Now()
	}

	recordingsDir := "/var/log/juvia/terminal"
	entries, err := os.ReadDir(recordingsDir)
	if err != nil {
		return "", nil
	}

	var latestFile string
	var latestModTime time.Time
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".cast" {
			continue
		}
		info, _ := entry.Info()
		if info.ModTime().After(latestModTime) {
			latestFile = entry.Name()
			latestModTime = info.ModTime()
		}
	}

	return latestFile, nil
}

func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}