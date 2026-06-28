package checks_test

import (
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

	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{"file exists", existingFile, ""},
		{"file missing", filepath.Join(dir, "missing.txt"), "does not exist"},
		{"path is directory", dir, "directory"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			check := &checks.FileCheck{Path: tt.path}
			err := check.Execute()
			if (err != nil) != (tt.wantErr != "") {
				t.Fatalf("wantErr=%q got=%v", tt.wantErr, err)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}
