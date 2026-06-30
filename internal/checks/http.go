package checks

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

type HttpReachableCheck struct {
	Address string
	Timeout time.Duration
}

func (h *HttpReachableCheck) Execute(ctx context.Context) error {
	client := &http.Client{
		Timeout: h.Timeout,
		// Do not follow redirects — a 3xx to an unexpected host would silently
		// succeed otherwise. The status-code check below surfaces it correctly.
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.Address, nil)
	if err != nil {
		return fmt.Errorf("http address %q is not reachable: %w", h.Address, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http address %q is not reachable: %w", h.Address, err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http address %q returned unhealthy status code: %d", h.Address, resp.StatusCode)
	}
	return nil
}

func validateHTTPAddress(address string) error {
	u, err := url.Parse(address)
	if err != nil {
		return fmt.Errorf("invalid address %q: %w", address, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("address %q must use http or https scheme, got %q", address, u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("address %q has no host", address)
	}
	return nil
}

func buildHttpReachableCheck(cfg model.CheckConfig) (registry.Check, error) {
	address, ok := cfg.Options["address"]
	if !ok || address == "" {
		return nil, fmt.Errorf("http_reachable check requires an 'address' option")
	}
	if err := validateHTTPAddress(address); err != nil {
		return nil, err
	}

	timeout := 5 * time.Second
	if ms, ok := cfg.Options["timeout_ms"]; ok {
		if v, err := strconv.Atoi(ms); err == nil {
			timeout = time.Duration(v) * time.Millisecond
		}
	}

	return &HttpReachableCheck{Address: address, Timeout: timeout}, nil
}

func init() {
	registry.Register(model.TypeHttpReachable, buildHttpReachableCheck)
}
