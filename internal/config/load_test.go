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

// s41: empty checks list is valid — Load() succeeds with 0 checks.
func TestLoad_EmptyChecksList(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	yaml := "version: 1\nchecks: []\n"
	if err := os.WriteFile(filepath.Join(dir, "preboot.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error for empty checks list: %v", err)
	}
	if len(cfg.Checks) != 0 {
		t.Errorf("expected 0 checks, got %d", len(cfg.Checks))
	}
}

// s45: same check name in both files — preboot.yml definition wins completely.
func TestLoad_BothFiles_OverlappingName(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	autoYAML := `version: 1
checks:
  - name: shared
    type: command_exists
    severity: blocker
    options:
      command: git
`
	explicitYAML := `version: 1
checks:
  - name: shared
    type: file_exists
    severity: info
    options:
      path: go.mod
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
	if len(cfg.Checks) != 1 {
		t.Errorf("expected exactly 1 check after override, got %d", len(cfg.Checks))
	}
	if cfg.Checks[0].Type != "file_exists" {
		t.Errorf("expected preboot.yml type file_exists to win, got %q", cfg.Checks[0].Type)
	}
}

// s46: preboot.yml overrides severity from blocker → info for a same-named check.
func TestLoad_BothFiles_SeverityOverride(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	autoYAML := `version: 1
checks:
  - name: go-check
    type: command_exists
    severity: blocker
    options:
      command: go
`
	explicitYAML := `version: 1
checks:
  - name: go-check
    type: command_exists
    severity: info
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
	if len(cfg.Checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(cfg.Checks))
	}
	if cfg.Checks[0].Severity != "info" {
		t.Errorf("expected severity info from preboot.yml override, got %q", cfg.Checks[0].Severity)
	}
}

// s48: --config ./custom.yml uses only that file, ignoring preboot-auto.yml and preboot.yml.
func TestLoadFrom_CustomPathIgnoresStandardFiles(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	standardYAML := `version: 1
checks:
  - name: standard-check
    type: command_exists
    severity: blocker
    options:
      command: go
`
	customYAML := `version: 1
checks:
  - name: custom-only-check
    type: command_exists
    severity: blocker
    options:
      command: go
`
	if err := os.WriteFile(filepath.Join(dir, "preboot-auto.yml"), []byte(standardYAML), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "preboot.yml"), []byte(standardYAML), 0644); err != nil {
		t.Fatal(err)
	}
	customPath := filepath.Join(dir, "custom.yml")
	if err := os.WriteFile(customPath, []byte(customYAML), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadFrom(customPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Checks) != 1 {
		t.Fatalf("expected 1 check from custom file, got %d", len(cfg.Checks))
	}
	if cfg.Checks[0].Name != "custom-only-check" {
		t.Errorf("expected custom-only-check, got %q", cfg.Checks[0].Name)
	}
}

// s52: duplicate check names in same file — both entries are kept; validation does not reject.
func TestLoad_DuplicateCheckNamesInSameFile(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	yaml := `version: 1
checks:
  - name: shared
    type: command_exists
    severity: blocker
    options:
      command: go
  - name: shared
    type: file_exists
    severity: info
    options:
      path: go.mod
`
	if err := os.WriteFile(filepath.Join(dir, "preboot.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error — duplicate names within one file are not rejected: %v", err)
	}
	if len(cfg.Checks) != 2 {
		t.Errorf("expected both entries kept, got %d checks", len(cfg.Checks))
	}
}

// s53: unicode characters in name and message fields — load and round-trip correctly.
func TestLoad_UnicodeFields(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	yaml := "version: 1\nchecks:\n  - name: \"ไป-ติดตั้ง-go\"\n    type: command_exists\n    severity: blocker\n    options:\n      command: go\n    message: \"กรุณาติดตั้ง Go ก่อนใช้งาน\"\n"
	if err := os.WriteFile(filepath.Join(dir, "preboot.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error with unicode fields: %v", err)
	}
	if len(cfg.Checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(cfg.Checks))
	}
	if cfg.Checks[0].Name != "ไป-ติดตั้ง-go" {
		t.Errorf("unicode name not preserved: %q", cfg.Checks[0].Name)
	}
	if cfg.Checks[0].Message != "กรุณาติดตั้ง Go ก่อนใช้งาน" {
		t.Errorf("unicode message not preserved: %q", cfg.Checks[0].Message)
	}
}

// s54: fix field containing shell metacharacters — stored as-is, never executed by preboot.
func TestLoad_FixFieldMetacharacters(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	fixValue := `brew install go && export PATH=$HOME/go/bin:$PATH | tee ~/.zshrc`
	yaml := "version: 1\nchecks:\n  - name: go-check\n    type: command_exists\n    severity: blocker\n    options:\n      command: go\n    fix: \"" + fixValue + "\"\n"
	if err := os.WriteFile(filepath.Join(dir, "preboot.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Checks[0].Fix != fixValue {
		t.Errorf("fix field not preserved verbatim\ngot:  %q\nwant: %q", cfg.Checks[0].Fix, fixValue)
	}
}

// s56: severity omitted from YAML — MergeDefaults defaults it to blocker before validation.
func TestLoad_BlankSeverityDefaults(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	yaml := `version: 1
checks:
  - name: go-check
    type: command_exists
    options:
      command: go
`
	if err := os.WriteFile(filepath.Join(dir, "preboot.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("blank severity should default to blocker via MergeDefaults, got error: %v", err)
	}
	if cfg.Checks[0].Severity != "blocker" {
		t.Errorf("expected severity blocker after default, got %q", cfg.Checks[0].Severity)
	}
}

// s57: check with every optional field present — all fields survive the round-trip through LoadFrom.
func TestLoad_AllOptionalFields(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	yaml := `version: 1
checks:
  - name: full-check
    type: command_exists
    severity: info
    options:
      command: go
    message: "Go must be installed to build this project"
    fix: "brew install go"
`
	if err := os.WriteFile(filepath.Join(dir, "preboot.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := cfg.Checks[0]
	if c.Name != "full-check" {
		t.Errorf("Name = %q, want %q", c.Name, "full-check")
	}
	if c.Type != "command_exists" {
		t.Errorf("Type = %q, want %q", c.Type, "command_exists")
	}
	if c.Severity != "info" {
		t.Errorf("Severity = %q, want %q", c.Severity, "info")
	}
	if c.Options["command"] != "go" {
		t.Errorf("Options[command] = %q, want %q", c.Options["command"], "go")
	}
	if c.Message != "Go must be installed to build this project" {
		t.Errorf("Message = %q", c.Message)
	}
	if c.Fix != "brew install go" {
		t.Errorf("Fix = %q", c.Fix)
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
