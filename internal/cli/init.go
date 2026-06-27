package cli

import (
	"fmt"
	"os"

	"github.com/allenbiji/clone-sage/internal/detect"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewInitCmd() *cobra.Command {
	var force bool

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Run init to initalize the project and generate sage-auto.yml",
		Long:  "Run this command to scan your entire repository and from the inferences in your repo, a sage-auto.yml file will be generated and which can also be extended via a sage.yml file",
		RunE: func(cmd *cobra.Command, args []string) error {

			if _, err := os.Stat("sage-auto.yml"); err == nil && !force {
				return fmt.Errorf("sage-auto.yml already exists — use --force to overwrite")
			}

			// Run ScanRepo() to generate baseline
			cfgs := detect.ScanRepo()

			if len(cfgs.Checks) == 0 {
				fmt.Fprintln(os.Stderr, "No recognised frameworks found. Generating empty baseline.")
			} else {
				fmt.Fprintf(os.Stderr, "Detected %d requirements. Building configuration...\n", len(cfgs.Checks))
			}

			// Marshal the struct cleanly into YAML bytes
			configYaml, err := yaml.Marshal(cfgs)
			if err != nil {
				return err
			}

			// Write generated baseline into sage-auto.yml file with standard 0644 file permissions
			if err := os.WriteFile("sage-auto.yml", configYaml, 0644); err != nil {
				return err
			}

			fmt.Fprintln(os.Stderr, "Baseline has been generated successfully in sage-auto.yml!")
			fmt.Fprintln(os.Stderr, "Run sage check to verify your local environment")
			return nil
		},
	}

	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing sage-auto.yml")

	return initCmd
}
