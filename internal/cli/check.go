package cli

import (
	"fmt"

	"github.com/allenbiji/preboot/internal/config"
	"github.com/allenbiji/preboot/internal/engine"
	"github.com/spf13/cobra"
)

func NewCheckCmd() *cobra.Command {
	var isQuickMode bool
	var cfgFile string
	var format string

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Check the preboot.yml file for any errors",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch format {
			case "text", "json":
				// valid
			default:
				return fmt.Errorf("unknown format %q: must be \"text\" or \"json\"", format)
			}

			cfg, err := config.LoadFrom(cfgFile)
			if err != nil {
				return err
			}

			return engine.Run(cfg, engine.RunOptions{
				QuickMode: isQuickMode,
				Format:    format,
			})
		},
	}

	checkCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Path to custom preboot.yml")
	checkCmd.Flags().BoolVarP(&isQuickMode, "quick", "q", false, "Run only fast, low-cost checks")
	checkCmd.Flags().StringVarP(&format, "format", "f", "text", `Output format: "text" or "json"`)

	return checkCmd
}
