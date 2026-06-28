package detect_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/allenbiji/preboot/internal/detect"
)

func TestExtractEnvKeys(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "simple pairs",
			content: "DB_HOST=localhost\nDB_PORT=5432\n",
			want:    map[string]string{"DB_HOST": "localhost", "DB_PORT": "5432"},
		},
		{
			name:    "comment lines skipped",
			content: "# this is a comment\nKEY=val\n",
			want:    map[string]string{"KEY": "val"},
		},
		{
			name:    "blank lines skipped",
			content: "\n\nKEY=val\n",
			want:    map[string]string{"KEY": "val"},
		},
		{
			name:    "inline comment stripped",
			content: "KEY=val # trailing comment\n",
			want:    map[string]string{"KEY": "val"},
		},
		{
			name:    "empty value",
			content: "KEY=\n",
			want:    map[string]string{"KEY": ""},
		},
		{
			name:    "whitespace trimmed around equals",
			content: "KEY = val\n",
			want:    map[string]string{"KEY": "val"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, ".env")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			got, err := detect.ExtractEnvKeys(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d keys, want %d: got=%v", len(got), len(tt.want), got)
			}
			for k, wantVal := range tt.want {
				if gotVal, ok := got[k]; !ok {
					t.Errorf("key %q missing from result", k)
				} else if gotVal != wantVal {
					t.Errorf("key %q: got %q, want %q", k, gotVal, wantVal)
				}
			}
		})
	}

	t.Run("file missing", func(t *testing.T) {
		t.Parallel()
		_, err := detect.ExtractEnvKeys(filepath.Join(t.TempDir(), "no-such-file.env"))
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})
}
