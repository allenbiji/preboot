package checks_test

import (
	"strings"
	"testing"

	"github.com/allenbiji/preboot/internal/checks"
	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

func TestBuildCommandExistsCheck(t *testing.T) {
	tests := []struct {
		name    string
		opts    map[string]string
		wantErr string
	}{
		{"no options", nil, "requires"},
		{"empty command", map[string]string{"command": ""}, "requires"},
		{"valid command", map[string]string{"command": "go"}, ""},
		{"path with slash", map[string]string{"command": "/usr/bin/go"}, "must be a bare name"},
		{"path with backslash", map[string]string{"command": "go\\env"}, "must be a bare name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := registry.Build(cfg(model.TypeCommandExists, tt.opts))
			if (err != nil) != (tt.wantErr != "") {
				t.Fatalf("wantErr=%q got=%v", tt.wantErr, err)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestCommandCheck_Execute(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr string
	}{
		{"go exists", "go", ""},
		{"missing command", "xyz-sage-impossible-cmd", "not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			check := &checks.CommandCheck{Command: tt.command}
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

// s13/s59: binary not in PATH because PATH is stripped — must fail gracefully, no panic.
func TestCommandCheck_Execute_StrippedPath(t *testing.T) {
	t.Setenv("PATH", "")
	check := &checks.CommandCheck{Command: "go"}
	err := check.Execute()
	if err == nil {
		t.Fatal("expected error when PATH is empty, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error %q does not contain 'not found'", err.Error())
	}
}
