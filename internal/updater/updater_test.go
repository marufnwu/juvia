package updater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetCurrentVersion(t *testing.T) {
	version := GetCurrentVersion()
	if version == "" {
		t.Error("expected non-empty version")
	}
}

func TestComputeChecksum(t *testing.T) {
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "juvia-test-"+t.Name())
	defer os.Remove(testFile)

	content := []byte("test content for checksum")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	checksum, err := ComputeChecksum(testFile)
	if err != nil {
		t.Fatalf("ComputeChecksum failed: %v", err)
	}

	if checksum == "" {
		t.Error("expected non-empty checksum")
	}

	checksum2, _ := ComputeChecksum(testFile)
	if checksum != checksum2 {
		t.Error("same content should produce same checksum")
	}
}

func TestComputeChecksum_FileNotFound(t *testing.T) {
	_, err := ComputeChecksum("/nonexistent/path/to/file")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestGetGOOS(t *testing.T) {
	os := getGOOS()
	if os == "" {
		t.Error("expected non-empty GOOS")
	}
}

func TestGetGOARCH(t *testing.T) {
	arch := getGOARCH()
	if arch == "" {
		t.Error("expected non-empty GOARCH")
	}
}