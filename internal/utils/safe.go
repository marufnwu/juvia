package utils

import (
	"errors"
	"path/filepath"
	"strings"
)

// ErrPathTraversal is returned when a path traversal attack is detected
var ErrPathTraversal = errors.New("path traversal attempt detected")

// SafePath validates that a user-provided path does not escape the base directory.
func SafePath(base, userInput string) (string, error) {
	resolved := filepath.Clean(filepath.Join(base, userInput))
	if !strings.HasPrefix(resolved, filepath.Clean(base)) {
		return "", ErrPathTraversal
	}
	return resolved, nil
}
