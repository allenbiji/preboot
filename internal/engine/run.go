package engine

import (
	"errors"
	"fmt"
	"os"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

// ANSI Color Codes for terminal output
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
)

// ErrCheckFailed is returned by Run when one or more blocker-severity checks fail.
// Callers use errors.Is to distinguish this from unexpected internal errors.
var ErrCheckFailed = errors.New("one or more blocker checks failed")

// Run executes the diagnostics. It returns nil if the environment is healthy,
// or ErrCheckFailed if a blocker check failed.
func Run(cfg *model.PrebootConfig, quickMode bool) error {
	fmt.Fprintln(os.Stderr, colorize(Cyan, "Running Preboot Diagnostics..."))

	hasBlockerFailed := false
	passedCount := 0
	failedCount := 0

	// Loop through every check in the YAML
	for _, checkCfg := range cfg.Checks {

		// In quick mode, skip network checks that require a round-trip.
		if quickMode && (checkCfg.Type == model.TypeHttpReachable || checkCfg.Type == model.TypeTcpReachable) {
			continue
		}

		// Inject global timeout_ms into per-check options if the check doesn't set its own.
		// Options is a map (reference type), so copy before modifying to avoid mutating cfg.
		if _, hasOwn := checkCfg.Options["timeout_ms"]; !hasOwn {
			if globalMs, ok := cfg.Defaults["timeout_ms"]; ok {
				merged := make(map[string]string, len(checkCfg.Options)+1)
				for k, v := range checkCfg.Options {
					merged[k] = v
				}
				merged["timeout_ms"] = fmt.Sprintf("%v", globalMs)
				checkCfg.Options = merged
			}
		}

		// Ask the Registry to build the physical check
		check, err := registry.Build(checkCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", colorize(Red, fmt.Sprintf("❌ %s [%s]: Internal Error - %v", checkCfg.Name, checkCfg.Type, err)))
			hasBlockerFailed = true
			failedCount++
			continue
		}

		// Execute the check
		err = check.Execute()

		// Handle the Result
		if err == nil {
			fmt.Printf("%s\n", colorize(Green, "✅ "+checkCfg.Name))
			passedCount++
		} else {
			failedCount++

			// Evaluate Severity
			switch checkCfg.Severity {
			case model.SeverityInfo:
				fmt.Printf("%s\n", colorize(Cyan, "ℹ️  "+checkCfg.Name+" (Info)"))
				fmt.Printf("   Reason: %v\n", err)
			case model.SeverityWarning:
				fmt.Printf("%s\n", colorize(Yellow, "⚠️  "+checkCfg.Name+" (Warning)"))
				fmt.Printf("   Reason: %v\n", err)
				if strict, _ := cfg.Defaults["strict"].(bool); strict {
					hasBlockerFailed = true
				}
			case model.SeverityBlocker:
				fmt.Printf("%s\n", colorize(Red, "❌ "+checkCfg.Name+" (BLOCKER)"))
				fmt.Printf("   Reason: %v\n", err)
				if checkCfg.Message != "" {
					fmt.Printf("   Message: %s\n", checkCfg.Message)
				}
				if checkCfg.Fix != "" {
					fmt.Printf("   Fix: %s\n", checkCfg.Fix)
				}
				hasBlockerFailed = true
			}
		}
	}

	if passedCount == 0 && failedCount == 0 {
		fmt.Fprintln(os.Stderr, "warn: no checks were configured")
	}

	// Print the Summary
	fmt.Println("\n----------------------------------------")
	if hasBlockerFailed {
		fmt.Printf("%s\n", colorize(Red, fmt.Sprintf("❌ DIAGNOSTICS FAILED: %d passed, %d failed", passedCount, failedCount)))
		return ErrCheckFailed
	}

	fmt.Printf("%s\n", colorize(Green, fmt.Sprintf("✅ DIAGNOSTICS PASSED: %d passed, %d failed (non-blocking)", passedCount, failedCount)))
	return nil
}
