package checks

import (
	"fmt"
	"net"
	"strconv"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

type PortFreeCheck struct {
	Port string
}

// execute method for the port free check
func (p *PortFreeCheck) Execute() error {
	listen, err := net.Listen("tcp", "127.0.0.1:"+p.Port)
	if err != nil {
		return fmt.Errorf("port %s is not free", p.Port)
	}

	defer listen.Close()

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

// build a factory for the port free check
func buildPortFreeCheck(cfg model.CheckConfig) (registry.Check, error) {
	port, ok := cfg.Options["port"]
	if !ok || port == "" {
		return nil, fmt.Errorf("The port_free check requires a 'port' option")
	}
	if err := validatePort(port); err != nil {
		return nil, err
	}
	return &PortFreeCheck{
		Port: port,
	}, nil
}

// registers the check in the registry
func init() {
	registry.Register(model.TypePortFree, buildPortFreeCheck)
}
