package cli_test

import (
	"testing"

	"github.com/allenbiji/preboot/internal/cli"
)

// exit matrix: unknown flag passed to a subcommand — RootCmd().Execute() returns a non-nil error.
// In the real binary Execute() calls os.Exit(2); in tests we call the Cobra method directly.
func TestRootCmd_UnknownFlag(t *testing.T) {
	cmd := cli.RootCmd()
	cmd.SetArgs([]string{"check", "--no-such-flag"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
}
