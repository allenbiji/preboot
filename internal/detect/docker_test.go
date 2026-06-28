package detect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/allenbiji/preboot/internal/model"
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

func TestDetectDockerCompose_NoFile(t *testing.T) {
	chdir(t, t.TempDir())
	got := detectDockerCompose()
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d checks", len(got))
	}
}

func TestDetectDockerCompose_YmlPresent(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	content := "services:\n  web:\n    ports:\n      - \"8080:80\"\n"
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	got := detectDockerCompose()
	var types []model.CheckType
	for _, c := range got {
		types = append(types, c.Type)
	}
	hasDockerInstalled := false
	hasPort := false
	for _, c := range got {
		if c.Type == model.TypeCommandExists && c.Options["command"] == "docker" {
			hasDockerInstalled = true
		}
		if c.Type == model.TypePortFree && c.Options["port"] == "8080" {
			hasPort = true
		}
	}
	_ = types
	if !hasDockerInstalled {
		t.Error("expected docker-installed check")
	}
	if !hasPort {
		t.Error("expected port-free check for port 8080")
	}
}

func TestDetectDockerCompose_YamlFallback(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	content := "services:\n  db:\n    ports:\n      - \"5432:5432\"\n"
	if err := os.WriteFile(filepath.Join(dir, "compose.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	got := detectDockerCompose()
	if len(got) == 0 {
		t.Fatal("expected checks from compose.yaml fallback, got none")
	}
	hasDockerInstalled := false
	for _, c := range got {
		if c.Type == model.TypeCommandExists && c.Options["command"] == "docker" {
			hasDockerInstalled = true
		}
	}
	if !hasDockerInstalled {
		t.Error("expected docker-installed check from compose.yaml")
	}
}

func TestDetectDockerCompose_EnvVarPort(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	content := "services:\n  web:\n    ports:\n      - \"${PORT}:80\"\n"
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	got := detectDockerCompose()
	for _, c := range got {
		if c.Type == model.TypePortFree {
			t.Errorf("env var port reference should be skipped, got port check: %+v", c)
		}
	}
}

func TestDetectDockerCompose_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(":::invalid yaml:::\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got := detectDockerCompose()
	if len(got) == 0 {
		t.Fatal("expected at least docker-installed check despite invalid YAML")
	}
	if got[0].Type != model.TypeCommandExists || got[0].Options["command"] != "docker" {
		t.Errorf("expected docker-installed check, got %+v", got[0])
	}
	if len(got) > 1 {
		t.Errorf("invalid YAML should produce only docker-installed check, got %d checks", len(got))
	}
}

func TestExtractHostPort(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"host:container", "8080:80", "8080"},
		{"ip:host:container", "127.0.0.1:8080:80", "8080"},
		{"bare port no colon", "80", ""},
		{"empty string", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractHostPort(tt.input)
			if got != tt.want {
				t.Errorf("extractHostPort(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
