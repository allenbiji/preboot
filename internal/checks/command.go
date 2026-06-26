package checks

import (
	"fmt"
	"os/exec"

	"github.com/allenbiji/clone-sage/internal/model"
	"github.com/allenbiji/clone-sage/internal/registry"
)

type CommandCheck struct {
	Command string
}

//execute method for the command_exists check
func (c *CommandCheck) Execute() error {
	_, err := exec.LookPath(c.Command)
	if err != nil {
		return fmt.Errorf("The command %s was not found in your $PATH", c.Command)
	}
	
	return nil
}

//builds command exists factory
func buildCommandExistsCheck(cfg model.CheckConfig) (registry.Check, error){
	cmd, ok := cfg.Options["Command"]
	if !ok || cmd == "" {
		return nil, fmt.Errorf("command_exists check requires a 'Command' option")	
	}

	return &CommandCheck{
		Command: cmd,
	}, nil
}

//register check to registry
func init() {
	registry.Register(model.TypeCommandExists, buildCommandExistsCheck)
}