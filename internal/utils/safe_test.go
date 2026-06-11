package utils

import (
	"path/filepath"
	"testing"
)

func TestSafePath_Valid(t *testing.T) {
	base := "/home/user"
	cases := []struct {
		input string
		want  string
	}{
		{"file.txt", filepath.Join("/home/user", "file.txt")},
		{"subdir/file.txt", filepath.Join("/home/user", "subdir/file.txt")},
		{"./file.txt", filepath.Join("/home/user", "file.txt")},
	}

	for _, c := range cases {
		got, err := SafePath(base, c.input)
		if err != nil {
			t.Fatalf("SafePath(%q, %q) unexpected error: %v", base, c.input, err)
		}
		if got != c.want {
			t.Errorf("SafePath(%q, %q) = %q; want %q", base, c.input, got, c.want)
		}
	}
}

func TestSafePath_Traversal(t *testing.T) {
	base := "/home/user"
	cases := []string{
		"../etc/passwd",
		"../../etc/passwd",
		"subdir/../../../etc/passwd",
	}

	for _, input := range cases {
		_, err := SafePath(base, input)
		if err != ErrPathTraversal {
			t.Errorf("SafePath(%q, %q) expected ErrPathTraversal, got: %v", base, input, err)
		}
	}
}
