package checks

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

type PortFreeCheck struct {
	Port string
}

// Execute binds 127.0.0.1:<port> to verify it is free. It only tests the loopback
// interface — a service listening exclusively on a non-loopback interface (e.g.
// 0.0.0.0) may still be reported as free by this check.
func (p *PortFreeCheck) Execute(_ context.Context) error {
	l, err := net.Listen("tcp", "127.0.0.1:"+p.Port)
	if err != nil {
		// Wrap underlying error so callers can distinguish "in use" from "permission denied".
		return fmt.Errorf("port %s is not free: %w", p.Port, err)
	}
	defer l.Close()
	return nil
}

func validatePort(portStr string) error {
	n, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("port %q is not a valid number: %w", portStr, err)
	}
	if n < 1 || n > 65535 {
		return fmt.Errorf("port %d is out of range (must be 1–65535)", n)
	}
	return nil
}

func buildPortFreeCheck(cfg model.CheckConfig) (registry.Check, error) {
	port, ok := cfg.Options["port"]
	if !ok || port == "" {
		return nil, fmt.Errorf("port_free check requires a 'port' option")
	}
	if err := validatePort(port); err != nil {
		return nil, err
	}
	return &PortFreeCheck{Port: port}, nil
}

func init() {
	registry.Register(model.TypePortFree, buildPortFreeCheck)
}
