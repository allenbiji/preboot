package checks

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

type CommandCheck struct {
	Command string
}

// execute method for the command_exists check
func (c *CommandCheck) Execute() error {
	_, err := exec.LookPath(c.Command)
	if err != nil {
		return fmt.Errorf("The command %s was not found in your $PATH", c.Command)
	}

	return nil
}

// builds command exists factory
func buildCommandExistsCheck(cfg model.CheckConfig) (registry.Check, error) {
	cmd, ok := cfg.Options["command"]
	if !ok || cmd == "" {
		return nil, fmt.Errorf("command_exists check requires a 'Command' option")
	}
	if strings.ContainsAny(cmd, "/\\") {
		return nil, fmt.Errorf("command_exists command %q must be a bare name, not a path", cmd)
	}

	return &CommandCheck{
		Command: cmd,
	}, nil
}

// register check to registry
func init() {
	registry.Register(model.TypeCommandExists, buildCommandExistsCheck)
}
