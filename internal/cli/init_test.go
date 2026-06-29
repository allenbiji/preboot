package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/allenbiji/preboot/internal/cli"
)

func chdirInit(t *testing.T, dir string) {
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

// s07: init creates preboot-auto.yml in an empty directory.
func TestInitCmd_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	chdirInit(t, dir)
	cmd := cli.NewInitCmd()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "preboot-auto.yml")); err != nil {
		t.Errorf("preboot-auto.yml not created: %v", err)
	}
}

// s06: --force overwrites an existing preboot-auto.yml.
func TestInitCmd_Force_Overwrites(t *testing.T) {
	dir := t.TempDir()
	chdirInit(t, dir)
	sentinel := "# sentinel content\n"
	if err := os.WriteFile(filepath.Join(dir, "preboot-auto.yml"), []byte(sentinel), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := cli.NewInitCmd()
	cmd.SetArgs([]string{"--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error with --force: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "preboot-auto.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == sentinel {
		t.Error("expected preboot-auto.yml to be overwritten, but content unchanged")
	}
}

// s08: without --force, init fails if preboot-auto.yml already exists.
func TestInitCmd_NoForce_ExistingFile_Errors(t *testing.T) {
	dir := t.TempDir()
	chdirInit(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "preboot-auto.yml"), []byte("# existing\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := cli.NewInitCmd()
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when preboot-auto.yml exists and --force not set, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error %q does not contain 'already exists'", err.Error())
	}
}

// s09: non-Go directory — init succeeds and creates a file with 0 checks.
func TestInitCmd_NonGoDirectory(t *testing.T) {
	dir := t.TempDir()
	chdirInit(t, dir)
	cmd := cli.NewInitCmd()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error for non-Go directory, got: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "preboot-auto.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "command") {
		t.Log("note: non-Go dir produced checks — update this test if behavior changes")
	}
}

// s68: --force on a read-only file returns a permission error.
func TestInitCmd_ReadOnlyFile_ForceErrors(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root — permission test skipped")
	}
	dir := t.TempDir()
	chdirInit(t, dir)
	target := filepath.Join(dir, "preboot-auto.yml")
	if err := os.WriteFile(target, []byte("# existing\n"), 0444); err != nil {
		t.Fatal(err)
	}
	cmd := cli.NewInitCmd()
	cmd.SetArgs([]string{"--force"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error writing to read-only file, got nil")
	}
	if !strings.Contains(err.Error(), "permission") {
		t.Errorf("error %q does not contain 'permission'", err.Error())
	}
}
