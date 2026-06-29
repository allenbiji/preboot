package detect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/allenbiji/preboot/internal/model"
)

// ── ExtractEnvKeys edge cases ─────────────────────────────────────────────────

func TestExtractEnvKeys_NoEqualsSign(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, ".env")
	if err := os.WriteFile(f, []byte("NOVALUE\nKEY=val\n"), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := ExtractEnvKeys(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := m["NOVALUE"]; ok {
		t.Error("line without = should not appear in map")
	}
	if m["KEY"] != "val" {
		t.Errorf("KEY = %q, want %q", m["KEY"], "val")
	}
}

func TestExtractEnvKeys_MultipleEquals(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, ".env")
	if err := os.WriteFile(f, []byte("KEY=val=ue\n"), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := ExtractEnvKeys(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["KEY"] != "val=ue" {
		t.Errorf("KEY = %q, want %q", m["KEY"], "val=ue")
	}
}

// ── detectGo ──────────────────────────────────────────────────────────────────

func TestDetectGo_GoModPresent(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example\ngo 1.21\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got := detectGo()
	if len(got) == 0 {
		t.Fatal("expected go-installed check, got none")
	}
	if got[0].Type != model.TypeCommandExists || got[0].Options["command"] != "go" {
		t.Errorf("expected go command check, got %+v", got[0])
	}
}

func TestDetectGo_GoModAbsent(t *testing.T) {
	chdir(t, t.TempDir())
	got := detectGo()
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d checks", len(got))
	}
}

// ── detectEnv ─────────────────────────────────────────────────────────────────

func TestDetectEnv_NoFile(t *testing.T) {
	chdir(t, t.TempDir())
	got := detectEnv()
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestDetectEnv_ExamplePriority(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, ".env.example"), []byte("DB=\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.template"), []byte("OTHER=\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got := detectEnv()
	// .env.example takes precedence: should contain env-file-exists + DB check, not OTHER
	hasDB := false
	hasOther := false
	for _, c := range got {
		if c.Options["key"] == "DB" {
			hasDB = true
		}
		if c.Options["key"] == "OTHER" {
			hasOther = true
		}
	}
	if !hasDB {
		t.Error("expected DB check from .env.example")
	}
	if hasOther {
		t.Error(".env.template should not be used when .env.example is present")
	}
}

func TestDetectEnv_TemplateFallback(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, ".env.template"), []byte("SECRET=\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got := detectEnv()
	if len(got) == 0 {
		t.Fatal("expected checks from .env.template, got none")
	}
	hasSECRET := false
	for _, c := range got {
		if c.Options["key"] == "SECRET" {
			hasSECRET = true
		}
	}
	if !hasSECRET {
		t.Error("expected SECRET key check from .env.template")
	}
}

func TestGenerateEnvChecks_MissingFile(t *testing.T) {
	chdir(t, t.TempDir())
	// file does not exist — should return only the file-exists check, no panic
	got := generateEnvChecks(".env.example")
	if len(got) != 1 {
		t.Fatalf("expected 1 check (file-exists only), got %d", len(got))
	}
	if got[0].Type != model.TypeFileExists {
		t.Errorf("expected file_exists check, got %s", got[0].Type)
	}
}

// ── ScanRepo ──────────────────────────────────────────────────────────────────

func TestScanRepo_EmptyDir(t *testing.T) {
	chdir(t, t.TempDir())
	cfg := ScanRepo()
	if cfg.Version != 1 {
		t.Errorf("version = %d, want 1", cfg.Version)
	}
	if len(cfg.Checks) != 0 {
		t.Errorf("expected 0 checks in empty dir, got %d", len(cfg.Checks))
	}
}

func TestScanRepo_GoProject(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := ScanRepo()
	hasGo := false
	for _, c := range cfg.Checks {
		if c.Name == "go-installed" {
			hasGo = true
		}
	}
	if !hasGo {
		t.Error("expected go-installed check for go.mod project")
	}
}

func TestScanRepo_WithMakefile(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\tgo build\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := ScanRepo()
	hasMake := false
	for _, c := range cfg.Checks {
		if c.Name == "make-installed" {
			hasMake = true
		}
	}
	if !hasMake {
		t.Error("expected make-installed check for Makefile project")
	}
}

// s05: all recognised artifacts present — ScanRepo emits go-installed, make-installed,
// docker-installed, a port-free check, and an env key check; no duplicate Name values.
func TestScanRepo_AllArtifacts(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	files := map[string]string{
		"go.mod":            "module example\ngo 1.21\n",
		"Makefile":          "build:\n\tgo build\n",
		"docker-compose.yml": "services:\n  app:\n    ports:\n      - \"8080:8080\"\n",
		".env.example":      "SECRET_KEY=\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg := ScanRepo()

	want := map[string]bool{
		"go-installed":     false,
		"make-installed":   false,
		"docker-installed": false,
		"port-free-8080":   false,
	}
	names := make(map[string]int)
	for _, c := range cfg.Checks {
		names[c.Name]++
		if _, ok := want[c.Name]; ok {
			want[c.Name] = true
		}
	}

	for name, found := range want {
		if !found {
			t.Errorf("expected check %q in ScanRepo output", name)
		}
	}
	for name, count := range names {
		if count > 1 {
			t.Errorf("duplicate check name %q (appears %d times)", name, count)
		}
	}

	// Verify at least one env key check exists (key=SECRET_KEY).
	hasEnvKey := false
	for _, c := range cfg.Checks {
		if c.Options["key"] == "SECRET_KEY" {
			hasEnvKey = true
		}
	}
	if !hasEnvKey {
		t.Error("expected env key check for SECRET_KEY from .env.example")
	}
}

// s10: go.work present without go.mod — detectGo() emits go-installed check.
func TestDetectGo_GoWorkPresent(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "go.work"), []byte("go 1.21\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got := detectGo()
	if len(got) == 0 {
		t.Fatal("expected go-installed check when go.work is present, got none")
	}
	if got[0].Type != model.TypeCommandExists || got[0].Options["command"] != "go" {
		t.Errorf("expected go command check, got %+v", got[0])
	}
}
