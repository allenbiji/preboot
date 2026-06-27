package cli

import (
	"fmt"

	"github.com/allenbiji/clone-sage/internal/config"
	"github.com/spf13/cobra"
)

// this is the command used to validate the user-defined checks
func NewValidateCmd() *cobra.Command {
	validCmd := &cobra.Command{
		Use:   "validate",
		Short: "Used to validate the checks in sage.yml",
		Long:  "This command can be to used to verify if the user-defined checks in sage.yml are valid or not",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := config.Load(); err != nil {
				return err
			}

			fmt.Println("✅ Configuration is valid!")
			return nil
		},
	}

	return validCmd
}
