package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/allenbiji/preboot/internal/engine"
	"github.com/allenbiji/preboot/internal/version"
	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "preboot",
		Short:         "Preboot handles local setup diagnostics",
		Long:          "An open-source CLI for diagnosing local development setup failures in repositories.",
		Version:       version.Version(),
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			printBanner()
		},
	}

	rootCmd.AddCommand(NewCheckCmd())
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewValidateCmd())

	return rootCmd
}

func Execute() {
	err := RootCmd().Execute()
	if err == nil {
		return
	}
	if errors.Is(err, engine.ErrCheckFailed) {
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
