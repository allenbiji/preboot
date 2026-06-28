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

func TestBuildDirectoryExistsCheck(t *testing.T) {
	tests := []struct {
		name    string
		opts    map[string]string
		wantErr string
	}{
		{"no options", nil, "requires a 'folder' option"},
		{"empty folder", map[string]string{"folder": ""}, "requires a 'folder' option"},
		{"valid folder", map[string]string{"folder": "x"}, ""},
		{"absolute path", map[string]string{"folder": "/tmp"}, "must be a relative path"},
		{"tilde path", map[string]string{"folder": "~/projects"}, "must not be a home-directory"},
		{"parent traversal", map[string]string{"folder": "../outside"}, "must not traverse parent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := registry.Build(cfg(model.TypeDirectoryExists, tt.opts))
			if (err != nil) != (tt.wantErr != "") {
				t.Fatalf("wantErr=%q got=%v", tt.wantErr, err)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestDirectoryCheck_Execute(t *testing.T) {
	dir := t.TempDir()
	fileInDir := filepath.Join(dir, "afile.txt")
	if err := os.WriteFile(fileInDir, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		folder  string
		wantErr string
	}{
		{"directory exists", dir, ""},
		{"directory missing", filepath.Join(dir, "no-such-dir"), "does not exist"},
		{"path is a file", fileInDir, "file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			check := &checks.DirectoryCheck{Folder: tt.folder}
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
