package checks_test

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/allenbiji/clone-sage/internal/checks"
	"github.com/allenbiji/clone-sage/internal/model"
	"github.com/allenbiji/clone-sage/internal/registry"
)

func TestBuildTcpReachableCheck(t *testing.T) {
	tests := []struct {
		name        string
		opts        map[string]string
		wantErr     string
		wantTimeout time.Duration
	}{
		{
			name:    "no address",
			opts:    nil,
			wantErr: "requires an 'address' option",
		},
		{
			name:        "default timeout",
			opts:        map[string]string{"address": "127.0.0.1:1"},
			wantTimeout: 5 * time.Second,
		},
		{
			name:        "custom timeout_ms",
			opts:        map[string]string{"address": "127.0.0.1:1", "timeout_ms": "1000"},
			wantTimeout: 1 * time.Second,
		},
		{
			name:        "invalid timeout_ms falls back to default",
			opts:        map[string]string{"address": "127.0.0.1:1", "timeout_ms": "abc"},
			wantTimeout: 5 * time.Second,
		},
		{
			name:    "no port",
			opts:    map[string]string{"address": "localhost"},
			wantErr: "host:port format",
		},
		{
			name:    "bare word",
			opts:    map[string]string{"address": "notanaddress"},
			wantErr: "host:port format",
		},
		{
			name:    "empty host",
			opts:    map[string]string{"address": ":8080"},
			wantErr: "has no host",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			check, err := registry.Build(cfg(model.TypeTcpReachable, tt.opts))
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tc, ok := check.(*checks.TcpReachableCheck)
			if !ok {
				t.Fatalf("expected *checks.TcpReachableCheck, got %T", check)
			}
			if tc.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %v, want %v", tc.Timeout, tt.wantTimeout)
			}
		})
	}
}

func TestTcpReachableCheck_Execute(t *testing.T) {
	t.Run("port open", func(t *testing.T) {
		t.Parallel()
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()
		check := &checks.TcpReachableCheck{Address: l.Addr().String(), Timeout: 5 * time.Second}
		if err := check.Execute(); err != nil {
			t.Errorf("expected nil for open port, got: %v", err)
		}
	})

	t.Run("port closed", func(t *testing.T) {
		t.Parallel()
		check := &checks.TcpReachableCheck{Address: "127.0.0.1:1", Timeout: 1 * time.Second}
		err := check.Execute()
		if err == nil {
			t.Fatal("expected error for closed port, got nil")
		}
		if !strings.Contains(err.Error(), "not reachable") {
			t.Errorf("error %q does not contain 'not reachable'", err.Error())
		}
	})
}
