package checks_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/allenbiji/preboot/internal/checks"
	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

func TestBuildFileCheck(t *testing.T) {
	tests := []struct {
		name    string
		opts    map[string]string
		wantErr string
	}{
		{"no options", nil, "requires a 'path' option"},
		{"empty path", map[string]string{"path": ""}, "requires a 'path' option"},
		{"valid path", map[string]string{"path": "x"}, ""},
		{"absolute path", map[string]string{"path": "/etc/passwd"}, "must be a relative path"},
		{"tilde path", map[string]string{"path": "~/config"}, "must not be a home-directory"},
		{"parent traversal", map[string]string{"path": "../../etc/hosts"}, "must not traverse parent"},
		{"nested valid path", map[string]string{"path": "sub/dir/file.txt"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := registry.Build(cfg(model.TypeFileExists, tt.opts))
			if (err != nil) != (tt.wantErr != "") {
				t.Fatalf("wantErr=%q got=%v", tt.wantErr, err)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestFileCheck_Execute(t *testing.T) {
	dir := t.TempDir()
	existingFile := filepath.Join(dir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	// s21: symlink pointing to an existing file (os.Stat follows symlinks)
	symlinkPath := filepath.Join(dir, "link.txt")
	if err := os.Symlink(existingFile, symlinkPath); err != nil {
		t.Fatal(err)
	}

	// s22: broken symlink whose target does not exist
	brokenLink := filepath.Join(dir, "broken.txt")
	if err := os.Symlink(filepath.Join(dir, "nonexistent.txt"), brokenLink); err != nil {
		t.Fatal(err)
	}

	// s65: file inside a directory whose name contains spaces
	spaceDir := filepath.Join(dir, "my project")
	if err := os.MkdirAll(spaceDir, 0755); err != nil {
		t.Fatal(err)
	}
	spaceFile := filepath.Join(spaceDir, "config.txt")
	if err := os.WriteFile(spaceFile, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	// s66: file inside a directory whose name contains unicode characters
	unicodeDir := filepath.Join(dir, "тест")
	if err := os.MkdirAll(unicodeDir, 0755); err != nil {
		t.Fatal(err)
	}
	unicodeFile := filepath.Join(unicodeDir, "file.txt")
	if err := os.WriteFile(unicodeFile, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{"file exists", existingFile, ""},
		{"file missing", filepath.Join(dir, "missing.txt"), "does not exist"},
		{"path is directory", dir, "directory"},
		{"symlink to existing file", symlinkPath, ""},
		{"broken symlink", brokenLink, "does not exist"},
		{"path with spaces", spaceFile, ""},
		{"path with unicode", unicodeFile, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			check := &checks.FileCheck{Path: tt.path}
			err := check.Execute(context.Background())
			if (err != nil) != (tt.wantErr != "") {
				t.Fatalf("wantErr=%q got=%v", tt.wantErr, err)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}
