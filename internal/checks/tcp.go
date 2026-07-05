package checks

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

type TcpReachableCheck struct {
	Address string
	Timeout time.Duration
}

func (t *TcpReachableCheck) Execute(ctx context.Context) error {
	dialer := &net.Dialer{Timeout: t.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", t.Address)
	if err != nil {
		return fmt.Errorf("tcp address %q is not reachable: %w", t.Address, err)
	}
	defer conn.Close()
	return nil
}

func validateTCPAddress(address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("tcp address %q must be in host:port format: %w", address, err)
	}
	if host == "" {
		return fmt.Errorf("tcp address %q has no host", address)
	}
	if port == "" {
		return fmt.Errorf("tcp address %q has no port", address)
	}
	return nil
}

func buildTcpReachableCheck(cfg model.CheckConfig) (registry.Check, error) {
	address, ok := cfg.Options["address"]
	if !ok || address == "" {
		return nil, fmt.Errorf("tcp_reachable check requires an 'address' option")
	}
	if err := validateTCPAddress(address); err != nil {
		return nil, err
	}

	timeout := 5 * time.Second
	if ms, ok := cfg.Options["timeout_ms"]; ok {
		if v, err := strconv.Atoi(ms); err == nil {
			timeout = time.Duration(v) * time.Millisecond
		}
	}

	return &TcpReachableCheck{Address: address, Timeout: timeout}, nil
}

func init() {
	registry.Register(model.TypeTcpReachable, buildTcpReachableCheck)
}
