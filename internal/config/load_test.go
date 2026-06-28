package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/allenbiji/preboot/internal/checks" // register check types for ValidateConfig
)

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
}

const validYAML = `version: 1
checks:
  - name: go-installed
    type: command_exists
    severity: blocker
    options:
      command: go
`

func TestLoadFrom_PathNotFound(t *testing.T) {
	_, err := LoadFrom("/nonexistent/path/preboot.yml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadFrom_InvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	bad := filepath.Join(tmp, "bad.yml")
	if err := os.WriteFile(bad, []byte(":::not yaml:::\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadFrom(bad)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadFrom_ValidConfig(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "preboot.yml")
	if err := os.WriteFile(f, []byte(validYAML), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadFrom(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != 1 {
		t.Errorf("version = %d, want 1", cfg.Version)
	}
	if len(cfg.Checks) != 1 || cfg.Checks[0].Name != "go-installed" {
		t.Errorf("unexpected checks: %+v", cfg.Checks)
	}
}

func TestLoadFrom_EmptyPath_FallsBackToLoad(t *testing.T) {
	// No config files in a fresh temp dir → Load() returns "No config files found"
	chdir(t, t.TempDir())
	_, err := LoadFrom("")
	if err == nil {
		t.Fatal("expected error when no config files present, got nil")
	}
	if !strings.Contains(err.Error(), "No config files found") {
		t.Errorf("error %q does not contain 'No config files found'", err.Error())
	}
}

func TestLoad_NeitherFile(t *testing.T) {
	chdir(t, t.TempDir())
	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "No config files found") {
		t.Errorf("error %q does not contain 'No config files found'", err.Error())
	}
}

func TestLoad_OnlySageYml(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "preboot.yml"), []byte(validYAML), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Checks) != 1 {
		t.Errorf("expected 1 check, got %d", len(cfg.Checks))
	}
}

func TestLoad_OnlySageAutoYml(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "preboot-auto.yml"), []byte(validYAML), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Checks) != 1 {
		t.Errorf("expected 1 check, got %d", len(cfg.Checks))
	}
}

func TestLoad_BothFiles(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	autoYAML := `version: 1
checks:
  - name: auto-check
    type: command_exists
    severity: blocker
    options:
      command: go
`
	explicitYAML := `version: 1
checks:
  - name: explicit-check
    type: command_exists
    severity: blocker
    options:
      command: go
`
	if err := os.WriteFile(filepath.Join(dir, "preboot-auto.yml"), []byte(autoYAML), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "preboot.yml"), []byte(explicitYAML), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Checks) != 2 {
		t.Errorf("expected 2 merged checks, got %d", len(cfg.Checks))
	}
}

func TestLoad_AutoParseError_SageMissing(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "preboot-auto.yml"), []byte(":::bad yaml:::\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when preboot-auto.yml fails to parse, got nil")
	}
}

func TestLoad_SageParseError_AutoMissing(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "preboot.yml"), []byte(":::bad yaml:::\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when preboot.yml fails to parse, got nil")
	}
}

func TestLoad_InvalidVersionRejects(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	bad := `version: 2
checks:
  - name: go-installed
    type: command_exists
    severity: blocker
    options:
      command: go
`
	if err := os.WriteFile(filepath.Join(dir, "preboot.yml"), []byte(bad), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for version 2, got nil")
	}
	if !strings.Contains(err.Error(), "Unsupported config version") {
		t.Errorf("error %q does not contain 'Unsupported config version'", err.Error())
	}
}
