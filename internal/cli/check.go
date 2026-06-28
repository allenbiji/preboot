package cli

import (
	"github.com/allenbiji/preboot/internal/config"
	"github.com/allenbiji/preboot/internal/engine"
	"github.com/spf13/cobra"
)

func NewCheckCmd() *cobra.Command {
	var isQuickMode bool
	var cfgFile string

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Check the preboot.yml file for any errors",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadFrom(cfgFile)
			if err != nil {
				return err
			}

			return engine.Run(cfg, isQuickMode)
		},
	}

	checkCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Path to custom preboot.yml")
	checkCmd.Flags().BoolVarP(&isQuickMode, "quick", "q", false, "Run only fast, low-cost checks")

	return checkCmd
}
