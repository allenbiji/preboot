package checks_test

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/allenbiji/preboot/internal/checks"
	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

func TestBuildPortFreeCheck(t *testing.T) {
	tests := []struct {
		name    string
		opts    map[string]string
		wantErr string
	}{
		{"no options", nil, "requires a 'port' option"},
		{"empty port", map[string]string{"port": ""}, "requires a 'port' option"},
		{"valid port", map[string]string{"port": "9000"}, ""},
		{"non-numeric", map[string]string{"port": "abc"}, "not a valid number"},
		{"port zero", map[string]string{"port": "0"}, "out of range"},
		{"port 65536", map[string]string{"port": "65536"}, "out of range"},
		{"negative port", map[string]string{"port": "-1"}, "out of range"},
		{"port min valid", map[string]string{"port": "1"}, ""},
		{"port max valid", map[string]string{"port": "65535"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := registry.Build(cfg(model.TypePortFree, tt.opts))
			if (err != nil) != (tt.wantErr != "") {
				t.Fatalf("wantErr=%q got=%v", tt.wantErr, err)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestPortFreeCheck_Execute(t *testing.T) {
	t.Run("port in use", func(t *testing.T) {
		t.Parallel()
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()
		port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
		check := &checks.PortFreeCheck{Port: port}
		err = check.Execute(context.Background())
		if err == nil {
			t.Fatal("expected error for port in use, got nil")
		}
		if !strings.Contains(err.Error(), "not free") {
			t.Errorf("error %q does not contain 'not free'", err.Error())
		}
	})

	t.Run("port free", func(t *testing.T) {
		t.Parallel()
		// Bind on :0 to get an ephemeral port, then release it.
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			t.Fatal(err)
		}
		port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
		l.Close()

		check := &checks.PortFreeCheck{Port: port}
		if err := check.Execute(context.Background()); err != nil {
			t.Errorf("expected nil for free port, got: %v", err)
		}
	})

	// s36: privileged port (<1024) as non-root — net.Listen fails with a permission error,
	// which the check surfaces as "not free". Skip when running as root.
	t.Run("privileged port as non-root", func(t *testing.T) {
		t.Parallel()
		if os.Getuid() == 0 {
			t.Skip("running as root — privileged port test skipped")
		}
		check := &checks.PortFreeCheck{Port: "80"}
		err := check.Execute(context.Background())
		if err == nil {
			t.Fatal("expected error for privileged port as non-root, got nil")
		}
		if !strings.Contains(err.Error(), "not free") {
			t.Errorf("error %q does not contain 'not free'", err.Error())
		}
	})
}
