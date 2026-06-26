package checks

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/allenbiji/clone-sage/internal/model"
	"github.com/allenbiji/clone-sage/internal/registry"
)

type HttpReachableCheck struct {
	Address string
	Timeout time.Duration
}

func (h *HttpReachableCheck) Execute() error {
	client := &http.Client{
		Timeout: h.Timeout,
	}

	resp, err := client.Get(h.Address)
	if err != nil {
		return fmt.Errorf("http address %q is not reachable: %w", h.Address, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http address %q returned unhealthy status code: %d", h.Address, resp.StatusCode)
	}

	return nil
}

func buildHttpReachableCheck(cfg model.CheckConfig) (registry.Check, error) {
	address, ok := cfg.Options["address"]
	if !ok || address == "" {
		return nil, fmt.Errorf("http_reachable check requires an 'address' option")
	}

	timeout := 5 * time.Second
	if ms, ok := cfg.Options["timeout_ms"]; ok {
		if v, err := strconv.Atoi(ms); err == nil {
			timeout = time.Duration(v) * time.Millisecond
		}
	}

	return &HttpReachableCheck{
		Address: address,
		Timeout: timeout,
	}, nil
}

func init() {
	registry.Register(model.TypeHttpReachable, buildHttpReachableCheck)
}
