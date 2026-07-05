package checks

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

type CommandCheck struct {
	Command string
}

func (c *CommandCheck) Execute(_ context.Context) error {
	_, err := exec.LookPath(c.Command)
	if err != nil {
		return fmt.Errorf("command %q not found in $PATH: %w", c.Command, err)
	}
	return nil
}

func buildCommandExistsCheck(cfg model.CheckConfig) (registry.Check, error) {
	cmd, ok := cfg.Options["command"]
	if !ok || cmd == "" {
		return nil, fmt.Errorf("command_exists check requires a 'command' option")
	}
	if strings.ContainsAny(cmd, "/\\") {
		return nil, fmt.Errorf("command_exists command %q must be a bare name, not a path", cmd)
	}
	return &CommandCheck{Command: cmd}, nil
}

func init() {
	registry.Register(model.TypeCommandExists, buildCommandExistsCheck)
}
