package checks

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/allenbiji/preboot/internal/model"
)

func TestEnvCheck_Execute(t *testing.T) {
	tests := []struct {
		name    string
		envMap  map[string]string
		key     string
		wantErr string
	}{
		{
			name:   "key with value",
			envMap: map[string]string{"DB": "localhost"},
			key:    "DB",
		},
		{
			name:    "key missing from map",
			envMap:  map[string]string{},
			key:     "DB",
			wantErr: "not found in .env",
		},
		{
			name:    "key with empty value",
			envMap:  map[string]string{"DB": ""},
			key:     "DB",
			wantErr: "has no value",
		},
		// s18a: values with '#' are truncated at the '#' by the .env parser — document the behavior.
		// The EnvCheck itself sees the already-parsed (truncated) value, so the check passes
		// as long as the portion before '#' is non-empty.
		{
			name:   "value truncated at hash is non-empty",
			envMap: map[string]string{"KEY": "abc"},
			key:    "KEY",
		},
		// s18b: shell metacharacters other than '#' are preserved verbatim — no crash.
		{
			name:   "value with shell metacharacters preserved",
			envMap: map[string]string{"GOFLAGS": "-mod=vendor -v"},
			key:    "GOFLAGS",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			check := &EnvCheck{Key: tt.key, EnvMap: tt.envMap}
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

func TestBuildEnvExistsCheck(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("no key option", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.Chdir(tmp); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chdir(origDir); cachedEnvMap = nil })

		_, err := buildEnvExistsCheck(model.CheckConfig{Options: map[string]string{}})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "requires a 'key' option") {
			t.Errorf("error %q does not contain 'requires a key option'", err.Error())
		}
	})

	t.Run("env file missing", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.Chdir(tmp); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chdir(origDir); cachedEnvMap = nil })

		_, err := buildEnvExistsCheck(model.CheckConfig{Options: map[string]string{"key": "DB"}})
		if err == nil {
			t.Fatal("expected error for missing .env, got nil")
		}
		if !strings.Contains(err.Error(), "could not read .env") {
			t.Errorf("error %q does not contain 'could not read .env'", err.Error())
		}
	})

	t.Run("valid", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.Chdir(tmp); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chdir(origDir); cachedEnvMap = nil })

		if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte("DB=localhost\n"), 0644); err != nil {
			t.Fatal(err)
		}
		check, err := buildEnvExistsCheck(model.CheckConfig{Options: map[string]string{"key": "DB"}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if check == nil {
			t.Fatal("expected non-nil check")
		}
		if err := check.Execute(context.Background()); err != nil {
			t.Errorf("Execute() error: %v", err)
		}
	})

	t.Run("concurrent cache access", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.Chdir(tmp); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chdir(origDir); cachedEnvMap = nil })

		if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte("DB=localhost\n"), 0644); err != nil {
			t.Fatal(err)
		}

		const goroutines = 20
		var wg sync.WaitGroup
		wg.Add(goroutines)
		for range goroutines {
			go func() {
				defer wg.Done()
				_, _ = buildEnvExistsCheck(model.CheckConfig{Options: map[string]string{"key": "DB"}})
			}()
		}
		wg.Wait()
	})

	t.Run("cache reuse", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.Chdir(tmp); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chdir(origDir); cachedEnvMap = nil })

		if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte("DB=localhost\nAPI=secret\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// First call — populates cache.
		_, err := buildEnvExistsCheck(model.CheckConfig{Options: map[string]string{"key": "DB"}})
		if err != nil {
			t.Fatalf("first call error: %v", err)
		}
		if cachedEnvMap == nil {
			t.Fatal("cachedEnvMap should be populated after first call")
		}

		// Second call — must reuse the cache (different key, same map).
		check2, err := buildEnvExistsCheck(model.CheckConfig{Options: map[string]string{"key": "API"}})
		if err != nil {
			t.Fatalf("second call error: %v", err)
		}
		if err := check2.Execute(context.Background()); err != nil {
			t.Errorf("second check Execute() error: %v", err)
		}
	})
}
