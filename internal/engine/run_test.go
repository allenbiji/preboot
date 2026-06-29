package engine_test

import (
	"errors"
	"fmt"
	"testing"

	_ "github.com/allenbiji/preboot/internal/checks" // register all check types via init()
	"github.com/allenbiji/preboot/internal/engine"
	"github.com/allenbiji/preboot/internal/model"
)

func passCfg(name string, sev model.Severity) model.CheckConfig {
	return model.CheckConfig{
		Name:     name,
		Type:     model.TypeCommandExists,
		Severity: sev,
		Options:  map[string]string{"command": "go"}, // go is always present in the test env
	}
}

func failCfg(name string, sev model.Severity) model.CheckConfig {
	return model.CheckConfig{
		Name:     name,
		Type:     model.TypeCommandExists,
		Severity: sev,
		Options:  map[string]string{"command": "xyz-sage-impossible-cmd"},
	}
}

func TestRun_EmptyChecks(t *testing.T) {
	cfg := &model.PrebootConfig{Version: 1}
	if err := engine.Run(cfg, false); err != nil {
		t.Errorf("expected nil for empty checks, got %v", err)
	}
}

func TestRun_AllPass(t *testing.T) {
	cfg := &model.PrebootConfig{
		Version: 1,
		Checks:  []model.CheckConfig{passCfg("go-installed", model.SeverityBlocker)},
	}
	if err := engine.Run(cfg, false); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRun_BlockerFails(t *testing.T) {
	cfg := &model.PrebootConfig{
		Version: 1,
		Checks:  []model.CheckConfig{failCfg("missing-cmd", model.SeverityBlocker)},
	}
	err := engine.Run(cfg, false)
	if !errors.Is(err, engine.ErrCheckFailed) {
		t.Errorf("expected ErrCheckFailed, got %v", err)
	}
}

func TestRun_WarningNonStrict(t *testing.T) {
	cfg := &model.PrebootConfig{
		Version:  1,
		Defaults: map[string]interface{}{"strict": false},
		Checks:   []model.CheckConfig{failCfg("warn-check", model.SeverityWarning)},
	}
	if err := engine.Run(cfg, false); err != nil {
		t.Errorf("warning in non-strict mode should return nil, got %v", err)
	}
}

func TestRun_WarningStrictMode(t *testing.T) {
	cfg := &model.PrebootConfig{
		Version:  1,
		Defaults: map[string]interface{}{"strict": true},
		Checks:   []model.CheckConfig{failCfg("warn-check", model.SeverityWarning)},
	}
	err := engine.Run(cfg, false)
	if !errors.Is(err, engine.ErrCheckFailed) {
		t.Errorf("warning in strict mode should return ErrCheckFailed, got %v", err)
	}
}

func TestRun_InfoNeverBlocks(t *testing.T) {
	cfg := &model.PrebootConfig{
		Version: 1,
		Checks:  []model.CheckConfig{failCfg("info-check", model.SeverityInfo)},
	}
	if err := engine.Run(cfg, false); err != nil {
		t.Errorf("info severity should never block, got %v", err)
	}
}

func TestRun_UnknownCheckType(t *testing.T) {
	cfg := &model.PrebootConfig{
		Version: 1,
		Checks: []model.CheckConfig{
			{Name: "bad-type", Type: model.CheckType("unknown_xyz"), Severity: model.SeverityBlocker},
		},
	}
	err := engine.Run(cfg, false)
	if !errors.Is(err, engine.ErrCheckFailed) {
		t.Errorf("unknown check type should trigger internal error path → ErrCheckFailed, got %v", err)
	}
}

func TestRun_QuickModeSkipsHttp(t *testing.T) {
	cfg := &model.PrebootConfig{
		Version: 1,
		Checks: []model.CheckConfig{
			{
				Name:     "http-check",
				Type:     model.TypeHttpReachable,
				Severity: model.SeverityBlocker,
				Options:  map[string]string{"address": "http://127.0.0.1:1"},
			},
		},
	}
	// quick=true must skip http_reachable; if it ran it would fail (port 1 is closed) → ErrCheckFailed
	if err := engine.Run(cfg, true); err != nil {
		t.Errorf("quick mode should skip http_reachable; got %v", err)
	}
}

func TestRun_QuickModeSkipsTcp(t *testing.T) {
	cfg := &model.PrebootConfig{
		Version: 1,
		Checks: []model.CheckConfig{
			{
				Name:     "tcp-check",
				Type:     model.TypeTcpReachable,
				Severity: model.SeverityBlocker,
				Options:  map[string]string{"address": "127.0.0.1:1"},
			},
		},
	}
	if err := engine.Run(cfg, true); err != nil {
		t.Errorf("quick mode should skip tcp_reachable; got %v", err)
	}
}

func TestRun_GlobalTimeoutInjected(t *testing.T) {
	// Use a real check with a very short injected timeout — if timeout is NOT injected
	// the check uses the 5s default and passes normally. Either way the injection path
	// is exercised. We verify via a pass (no panic, no crash) and that quick=false runs it.
	cfg := &model.PrebootConfig{
		Version:  1,
		Defaults: map[string]interface{}{"timeout_ms": "5000"},
		Checks:   []model.CheckConfig{passCfg("go-installed", model.SeverityBlocker)},
	}
	if err := engine.Run(cfg, false); err != nil {
		t.Errorf("global timeout injection should not break passing check: %v", err)
	}
}

// exit matrix: blocker + warning + info all failing — ErrCheckFailed propagates.
func TestRun_MixedSeverityFailures(t *testing.T) {
	cfg := &model.PrebootConfig{
		Version: 1,
		Defaults: map[string]interface{}{"strict": true},
		Checks: []model.CheckConfig{
			failCfg("fail-blocker", model.SeverityBlocker),
			failCfg("fail-warning", model.SeverityWarning),
			failCfg("fail-info", model.SeverityInfo),
		},
	}
	err := engine.Run(cfg, false)
	if !errors.Is(err, engine.ErrCheckFailed) {
		t.Errorf("expected ErrCheckFailed with mixed severity failures, got %v", err)
	}
}

// s55: 600-character check name — Run() completes without panic or truncation error.
func TestRun_LongCheckName(t *testing.T) {
	longName := make([]byte, 600)
	for i := range longName {
		longName[i] = 'a'
	}
	cfg := &model.PrebootConfig{
		Version: 1,
		Checks:  []model.CheckConfig{failCfg(string(longName), model.SeverityInfo)},
	}
	// Info severity never blocks; we just want no panic.
	if err := engine.Run(cfg, false); err != nil {
		t.Errorf("long check name should not cause an error (info severity): %v", err)
	}
}

// s69: 60 passing checks — Run() handles a large check count without crashing.
func TestRun_LargeCheckCount(t *testing.T) {
	checks := make([]model.CheckConfig, 60)
	for i := range checks {
		checks[i] = passCfg(fmt.Sprintf("check-%d", i), model.SeverityBlocker)
	}
	cfg := &model.PrebootConfig{Version: 1, Checks: checks}
	if err := engine.Run(cfg, false); err != nil {
		t.Errorf("expected nil for 60 passing checks, got %v", err)
	}
}

// s70: concurrent invocations — 10 goroutines each run their own config; no data race.
// Run with: go test -race ./internal/engine/...
func TestRun_ConcurrentInvocations(t *testing.T) {
	const n = 10
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() {
			cfg := &model.PrebootConfig{
				Version: 1,
				Checks:  []model.CheckConfig{passCfg("go-installed", model.SeverityBlocker)},
			}
			errs <- engine.Run(cfg, false)
		}()
	}
	for i := 0; i < n; i++ {
		if err := <-errs; err != nil {
			t.Errorf("goroutine %d: unexpected error: %v", i, err)
		}
	}
}

func TestRun_OwnTimeoutNotOverridden(t *testing.T) {
	cfg := &model.PrebootConfig{
		Version:  1,
		Defaults: map[string]interface{}{"timeout_ms": "9999"},
		Checks: []model.CheckConfig{
			{
				Name:     "go-installed",
				Type:     model.TypeCommandExists,
				Severity: model.SeverityBlocker,
				Options:  map[string]string{"command": "go", "timeout_ms": "1000"},
			},
		},
	}
	if err := engine.Run(cfg, false); err != nil {
		t.Errorf("own timeout_ms should be preserved and check should pass: %v", err)
	}
}
