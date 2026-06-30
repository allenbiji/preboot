package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

// ANSI color codes used by renderText.
const (
	Reset    = "\033[0m"
	Red      = "\033[31m"
	Green    = "\033[32m"
	Yellow   = "\033[33m"
	Cyan     = "\033[36m"
	CyanBold = "\033[1;36m"
)

// ErrCheckFailed is returned by Run when one or more blocker-severity checks fail.
// Callers use errors.Is to distinguish this from unexpected internal errors.
var ErrCheckFailed = errors.New("one or more blocker checks failed")

// RunOptions controls Run behaviour. Zero value runs in text mode writing to
// os.Stdout/os.Stderr with no quick-mode filtering.
type RunOptions struct {
	// QuickMode skips network checks (http_reachable, tcp_reachable).
	QuickMode bool
	// Format is "text" (default) or "json". Any other value is treated as "text".
	Format string
	// Stdout receives the report (check results + summary in text mode; JSON object in json mode).
	// Defaults to os.Stdout when nil.
	Stdout io.Writer
	// Stderr receives progress lines ("Running Preboot Diagnostics...") and warnings.
	// Defaults to os.Stderr when nil.
	Stderr io.Writer
	// Ctx is the parent context. When nil, Run creates one wired to SIGINT/SIGTERM so
	// Ctrl-C cancels in-flight HTTP/TCP checks cleanly.
	Ctx context.Context
}

// Run executes all configured checks and renders a report.
// It returns nil when the environment is healthy, or ErrCheckFailed when any
// blocker (or warning in strict mode) fails.
func Run(cfg *model.PrebootConfig, opts RunOptions) error {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	ctx := opts.Ctx
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
	}

	if opts.Format != "json" {
		fmt.Fprintln(opts.Stderr, colorize(Cyan, "Running Preboot Diagnostics..."))
	}

	report := RunReport{}

	for _, checkCfg := range cfg.Checks {
		if opts.QuickMode && (checkCfg.Type == model.TypeHttpReachable || checkCfg.Type == model.TypeTcpReachable) {
			report.Checks = append(report.Checks, CheckResult{
				Name:     checkCfg.Name,
				Type:     string(checkCfg.Type),
				Severity: string(checkCfg.Severity),
				Status:   "skipped",
			})
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

		check, buildErr := registry.Build(checkCfg)
		if buildErr != nil {
			report.Failed++
			report.BlockerFailed = true
			report.Checks = append(report.Checks, CheckResult{
				Name:     checkCfg.Name,
				Type:     string(checkCfg.Type),
				Severity: string(checkCfg.Severity),
				Status:   "fail",
				Reason:   fmt.Sprintf("internal error: %v", buildErr),
			})
			continue
		}

		result := CheckResult{
			Name:     checkCfg.Name,
			Type:     string(checkCfg.Type),
			Severity: string(checkCfg.Severity),
			Message:  checkCfg.Message,
			Fix:      checkCfg.Fix,
		}

		execErr := check.Execute(ctx)
		if execErr == nil {
			result.Status = "pass"
			report.Passed++
		} else {
			result.Status = "fail"
			result.Reason = execErr.Error()
			report.Failed++
			switch checkCfg.Severity {
			case model.SeverityWarning:
				if strict, _ := cfg.Defaults["strict"].(bool); strict {
					report.BlockerFailed = true
				}
			case model.SeverityBlocker:
				report.BlockerFailed = true
			}
		}

		report.Checks = append(report.Checks, result)
	}

	if opts.Format == "json" {
		renderJSON(opts.Stdout, report)
	} else {
		renderText(opts.Stdout, opts.Stderr, report)
	}

	if report.BlockerFailed {
		return ErrCheckFailed
	}
	return nil
}

// renderText writes the human-readable report to stdout and any warnings to stderr.
func renderText(stdout, stderr io.Writer, report RunReport) {
	if report.Passed == 0 && report.Failed == 0 {
		fmt.Fprintln(stderr, "warn: no checks were configured")
	}

	fmt.Fprintln(stdout)
	for _, r := range report.Checks {
		switch r.Status {
		case "skipped":
			continue
		case "pass":
			fmt.Fprintf(stdout, "%s\n", colorize(Green, "✅ "+r.Name))
		case "fail":
			switch model.Severity(r.Severity) {
			case model.SeverityInfo:
				fmt.Fprintf(stdout, "%s\n", colorize(Cyan, "ℹ️  "+r.Name+" (Info)"))
				fmt.Fprintf(stdout, "   Reason: %s\n", r.Reason)
			case model.SeverityWarning:
				fmt.Fprintf(stdout, "%s\n", colorize(Yellow, "⚠️  "+r.Name+" (Warning)"))
				fmt.Fprintf(stdout, "   Reason: %s\n", r.Reason)
			case model.SeverityBlocker:
				fmt.Fprintf(stdout, "%s\n", colorize(Red, "❌ "+r.Name+" (BLOCKER)"))
				fmt.Fprintf(stdout, "   Reason: %s\n", r.Reason)
				if r.Message != "" {
					fmt.Fprintf(stdout, "   Message: %s\n", r.Message)
				}
				if r.Fix != "" {
					fmt.Fprintf(stdout, "   Fix: %s\n", r.Fix)
				}
			}
		}
	}

	fmt.Fprintln(stdout, "----------------------------------------")
	if report.BlockerFailed {
		fmt.Fprintf(stdout, "%s\n", colorize(Red, fmt.Sprintf("❌ DIAGNOSTICS FAILED: %d passed, %d failed", report.Passed, report.Failed)))
	} else {
		fmt.Fprintf(stdout, "%s\n", colorize(Green, fmt.Sprintf("✅ DIAGNOSTICS PASSED: %d passed, %d failed (non-blocking)", report.Passed, report.Failed)))
	}
	fmt.Fprintln(stdout)
}

// renderJSON writes a single JSON object to stdout.
func renderJSON(stdout io.Writer, report RunReport) {
	data, _ := json.MarshalIndent(report, "", "  ")
	fmt.Fprintf(stdout, "%s\n", data)
}
