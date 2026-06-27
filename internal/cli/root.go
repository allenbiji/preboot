package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "sage",
		Short: "Clonesage handles local setup diagnostics",
		Long:  "An open-source CLI for diagnosing local development setup failures in repositories.",
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
	if err := RootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
