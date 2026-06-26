package checks

import (
	"fmt"
	"net"

	"github.com/allenbiji/clone-sage/internal/model"
	"github.com/allenbiji/clone-sage/internal/registry"
)

type TcpReachableCheck struct {
	Address string
}

//execute method for the tcp_reachable check
func (t *TcpReachableCheck) Execute() error {
	conn, err := net.DialTimeout("tcp", t.Address, 3000)
	if err != nil {
		return fmt.Errorf("The tcp address '%s' is  not reachable", t.Address)
	}

	defer conn.Close()

	return nil
}

//build a factory for the tcp_reachable check
func buildTcpReachableCheck(cfg model.CheckConfig) (registry.Check, error) {
	address, ok := cfg.Options["address"]
	if !ok || address == "" {
		return nil, fmt.Errorf("The tcp_reachable check requires a 'address' option")
	}

	return &TcpReachableCheck{
		Address: address,
	}, nil
}

//registers the check in the registry
func init() {
	registry.Register(model.TypeTcpReachable, buildTcpReachableCheck)
}