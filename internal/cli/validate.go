package cli

import (
	"fmt"

	"github.com/allenbiji/clone-sage/internal/config"
	"github.com/spf13/cobra"
)

// this is the command used to validate the user-defined checks
func NewValidateCmd() *cobra.Command {
	var cfgFile string

	validCmd := &cobra.Command{
		Use:   "validate",
		Short: "Used to validate the checks in sage.yml",
		Long:  "This command can be to used to verify if the user-defined checks in sage.yml are valid or not",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := config.LoadFrom(cfgFile); err != nil {
				return err
			}

			fmt.Println("✅ Configuration is valid!")
			return nil
		},
	}

	validCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Path to custom sage.yml")

	return validCmd
}
