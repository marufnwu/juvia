package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type UpdateInfo struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion string `json:"latest_version"`
	DownloadURL   string `json:"download_url"`
	Changelog    string `json:"changelog"`
	SigURL       string `json:"sig_url"`
}

var currentVersion = "0.1.0"

func CheckForUpdates() (*UpdateInfo, error) {
	updateURL := "https://updates.juvia.io/latest.json"

	resp, err := http.Get(updateURL)
	if err != nil {
		return &UpdateInfo{
			CurrentVersion: currentVersion,
			LatestVersion: currentVersion,
		}, nil
	}
	defer resp.Body.Close()

	var update UpdateInfo
	if resp.StatusCode != 200 {
		return &UpdateInfo{
			CurrentVersion: currentVersion,
			LatestVersion: currentVersion,
		}, nil
	}

	return &update, nil
}

func DownloadUpdate(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	tmpDir := os.TempDir()
	destPath := filepath.Join(tmpDir, "juvia-update")

	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return destPath, nil
}

func VerifyGPGSignature(binaryPath, sigURL string) error {
	sigData, err := downloadSig(sigURL)
	if err != nil {
		return fmt.Errorf("download sig: %w", err)
	}

	sigPath := binaryPath + ".sig"
	os.WriteFile(sigPath, sigData, 0644)
	defer os.Remove(sigPath)

	cmd := exec.Command("gpg", "--verify", sigPath, binaryPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gpg verify failed: %w: %s", err, string(output))
	}

	return nil
}

func ApplyUpdate(binaryPath string) error {
	currentBinary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get current binary: %w", err)
	}

	backupPath := currentBinary + ".backup"

	data, err := os.ReadFile(currentBinary)
	if err != nil {
		return fmt.Errorf("read current binary: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0755); err != nil {
		return fmt.Errorf("write backup: %w", err)
	}

	if err := os.Rename(binaryPath, currentBinary); err != nil {
		os.Rename(backupPath, currentBinary)
		return fmt.Errorf("apply update: %w", err)
	}

	os.Remove(backupPath)

	cmd := exec.Command("systemctl", "restart", "juvia")
	cmd.Run()

	return nil
}

func Rollback() error {
	currentBinary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get current binary: %w", err)
	}

	backupPath := currentBinary + ".backup"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found")
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("read backup: %w", err)
	}

	if err := os.WriteFile(currentBinary, data, 0755); err != nil {
		return fmt.Errorf("write binary: %w", err)
	}

	os.Remove(backupPath)

	cmd := exec.Command("systemctl", "restart", "juvia")
	cmd.Run()

	return nil
}

func ComputeChecksum(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func GetCurrentVersion() string {
	return currentVersion
}

func downloadSig(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func getGOOS() string {
	return runtime.GOOS
}

func getGOARCH() string {
	return runtime.GOARCH
}