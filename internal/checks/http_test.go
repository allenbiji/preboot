package checks_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/allenbiji/preboot/internal/checks"
	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

// s30: self-signed TLS cert — standard http.Client rejects it, check fails.
func TestHttpReachableCheck_SelfSignedCert(t *testing.T) {
	t.Parallel()
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer s.Close()
	check := &checks.HttpReachableCheck{Address: s.URL, Timeout: 5 * time.Second}
	err := check.Execute()
	if err == nil {
		t.Fatal("expected TLS error for self-signed cert, got nil")
	}
	if !strings.Contains(err.Error(), "not reachable") {
		t.Errorf("error %q does not contain 'not reachable'", err.Error())
	}
}

func TestBuildHttpReachableCheck(t *testing.T) {
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
			opts:        map[string]string{"address": "http://x"},
			wantTimeout: 5 * time.Second,
		},
		{
			name:        "custom timeout_ms",
			opts:        map[string]string{"address": "http://x", "timeout_ms": "2000"},
			wantTimeout: 2 * time.Second,
		},
		{
			name:        "invalid timeout_ms falls back to default",
			opts:        map[string]string{"address": "http://x", "timeout_ms": "bad"},
			wantTimeout: 5 * time.Second,
		},
		{
			name:    "no scheme",
			opts:    map[string]string{"address": "localhost:8080"},
			wantErr: "must use http or https",
		},
		{
			name:    "ftp scheme",
			opts:    map[string]string{"address": "ftp://host.example.com"},
			wantErr: "must use http or https",
		},
		{
			name:    "file scheme",
			opts:    map[string]string{"address": "file:///etc/passwd"},
			wantErr: "must use http or https",
		},
		{
			name:    "no host",
			opts:    map[string]string{"address": "http://"},
			wantErr: "has no host",
		},
		{
			name:        "valid https",
			opts:        map[string]string{"address": "https://x"},
			wantTimeout: 5 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			check, err := registry.Build(cfg(model.TypeHttpReachable, tt.opts))
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
			hc, ok := check.(*checks.HttpReachableCheck)
			if !ok {
				t.Fatalf("expected *checks.HttpReachableCheck, got %T", check)
			}
			if hc.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %v, want %v", hc.Timeout, tt.wantTimeout)
			}
		})
	}
}

func TestHttpReachableCheck_Execute(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		addr    string // overrides server URL when set
		timeout time.Duration
		wantErr string
	}{
		{
			name:    "200 OK",
			handler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) },
			timeout: 5 * time.Second,
		},
		{
			name:    "301 redirect",
			handler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusMovedPermanently) },
			timeout: 5 * time.Second,
			wantErr: "301",
		},
		{
			name:    "404 not found",
			handler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNotFound) },
			timeout: 5 * time.Second,
			wantErr: "404",
		},
		{
			name:    "500 internal error",
			handler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) },
			timeout: 5 * time.Second,
			wantErr: "500",
		},
		{
			name:    "unreachable address",
			addr:    "http://127.0.0.1:1",
			timeout: 1 * time.Second,
			wantErr: "not reachable",
		},
		{
			name: "timeout exceeded",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(200 * time.Millisecond)
			},
			timeout: 10 * time.Millisecond,
			wantErr: "not reachable",
		},
		// s27: other 2xx codes besides 200 must pass
		{
			name:    "201 Created",
			handler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusCreated) },
			timeout: 5 * time.Second,
		},
		{
			name:    "204 No Content",
			handler: func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) },
			timeout: 5 * time.Second,
		},
		// s61: DNS resolution failure — .invalid TLD never resolves (RFC 2606)
		{
			name:    "DNS resolution failure",
			addr:    "http://preboot-no-such-host.invalid:8080/",
			timeout: 5 * time.Second,
			wantErr: "not reachable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			addr := tt.addr
			if addr == "" {
				s := httptest.NewServer(tt.handler)
				defer s.Close()
				addr = s.URL
			}
			check := &checks.HttpReachableCheck{Address: addr, Timeout: tt.timeout}
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
