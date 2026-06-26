package checks

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/allenbiji/clone-sage/internal/model"
	"github.com/allenbiji/clone-sage/internal/registry"
)

type TcpReachableCheck struct {
	Address string
	Timeout time.Duration
}

func (t *TcpReachableCheck) Execute() error {
	conn, err := net.DialTimeout("tcp", t.Address, t.Timeout)
	if err != nil {
		return fmt.Errorf("tcp address %q is not reachable: %w", t.Address, err)
	}

	defer conn.Close()

	return nil
}

func buildTcpReachableCheck(cfg model.CheckConfig) (registry.Check, error) {
	address, ok := cfg.Options["address"]
	if !ok || address == "" {
		return nil, fmt.Errorf("tcp_reachable check requires an 'address' option")
	}

	timeout := 5 * time.Second
	if ms, ok := cfg.Options["timeout_ms"]; ok {
		if v, err := strconv.Atoi(ms); err == nil {
			timeout = time.Duration(v) * time.Millisecond
		}
	}

	return &TcpReachableCheck{
		Address: address,
		Timeout: timeout,
	}, nil
}

func init() {
	registry.Register(model.TypeTcpReachable, buildTcpReachableCheck)
}
