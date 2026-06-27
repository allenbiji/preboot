package engine

import (
	"fmt"

	"github.com/allenbiji/clone-sage/internal/model"
	"github.com/allenbiji/clone-sage/internal/registry"
)

// ANSI Color Codes for terminal output
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
)

// Run executes the diagnostics. It returns true if the environment is healthy
// (no blockers failed), and false if a blocker failed.
func Run(cfg *model.ClonesageConfig, quickMode bool) bool {
	fmt.Println(Cyan + "Running CloneSage Diagnostics...\n" + Reset)

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
			fmt.Printf("❌ %s [%s]: Internal Error - %v\n", checkCfg.Name, checkCfg.Type, err)
			hasBlockerFailed = true
			failedCount++
			continue
		}

		// Execute the check
		err = check.Execute()

		// Handle the Result
		if err == nil {
			fmt.Printf("%s✅ %s%s\n", Green, checkCfg.Name, Reset)
			passedCount++
		} else {
			failedCount++

			// Evaluate Severity
			switch checkCfg.Severity {
			case model.SeverityInfo:
				fmt.Printf("%sℹ️  %s (Info)%s\n", Cyan, checkCfg.Name, Reset)
				fmt.Printf("   Reason: %v\n", err)
			case model.SeverityWarning:
				fmt.Printf("%s⚠️  %s (Warning)%s\n", Yellow, checkCfg.Name, Reset)
				fmt.Printf("   Reason: %v\n", err)
				if strict, _ := cfg.Defaults["strict"].(bool); strict {
					hasBlockerFailed = true
				}
			case model.SeverityBlocker:
				fmt.Printf("%s❌ %s (BLOCKER)%s\n", Red, checkCfg.Name, Reset)
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

	// Print the Summary
	fmt.Println("\n----------------------------------------")
	if hasBlockerFailed {
		fmt.Printf("%s❌ DIAGNOSTICS FAILED: %d passed, %d failed%s\n", Red, passedCount, failedCount, Reset)
		return false
	}

	fmt.Printf("%s✅ DIAGNOSTICS PASSED: %d passed, %d failed (non-blocking)%s\n", Green, passedCount, failedCount, Reset)
	return true
}
