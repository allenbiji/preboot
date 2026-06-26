package checks

import (
	"fmt"
	"net/http"

	"github.com/allenbiji/clone-sage/internal/model"
	"github.com/allenbiji/clone-sage/internal/registry"
)

type HttpReachableCheck struct {
	Address string
}

//execute method for the http_reachable check
func (h *HttpReachableCheck) Execute() error {
	client := &http.Client{
		Timeout: 2000, //hardcoded for now
	}

	resp, err := client.Get(h.Address)
	if err != nil {
		return fmt.Errorf("the http address '%s' is not reachable: %w", h.Address, err)
	}
	
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("address '%s' returned unhealthy status code: %d", h.Address, resp.StatusCode)
	}

	return nil
}

//build a factory for the http_reachable check
func buildHttpReachableCheck(cfg model.CheckConfig) (registry.Check, error) {
	address, ok := cfg.Options["address"]
	if !ok || address == "" {
		return nil, fmt.Errorf("The http_reachable check requires a 'address' option")
	}

	return &HttpReachableCheck{
		Address: address,
	}, nil
}

//registers the check in the registry
func init() {
	registry.Register(model.TypeHttpReachable, buildHttpReachableCheck)
}