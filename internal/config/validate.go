package config

import (
	"fmt"
	"strings"

	"github.com/allenbiji/clone-sage/internal/model"
)

// validateSeverity ensures the string cast by Viper matches our strict enums.
func validateSeverity(check model.CheckConfig) error {
	switch check.Severity {
	case model.SeverityInfo, model.SeverityBlocker, model.SeverityWarning:
		return nil
	default:
		return fmt.Errorf("Invalid severity '%s' (allowed: info, warning, blocker)", check.Severity)
	}
}

// validateType ensures the check driver exists.
// Note: We are using a static switch here for now. Later, this will be
// replaced by asking the internal/registry package!
func validateCheckTypes(check model.CheckConfig) error {
	switch check.Type {
	case model.TypeCommandExists, model.TypeDirectoryExists, model.TypeEnvExists, model.TypeFileExists, model.TypeHttpReachable, model.TypePortFree, model.TypeTcpReachable:
		return nil
	default:
		return fmt.Errorf("Invalid check type '%s' (allowed: env_exists, file_exists, directory_exists, http_reachable, port_free, command_exists, tcp_reachable)", check.Type)
	}
}

// ValidateConfig acts as the runtime firewall, ensuring the unmarshaled YAML data
// strictly conforms to our domain types and business rules.
func ValidateConfig(cfg *model.ClonesageConfig) error {
	var errs []string

	if cfg.Version != 1 {
		errs = append(errs, fmt.Sprintf("Unsupported config version: %d", cfg.Version))
	}

	for i, check := range cfg.Checks {
		if strings.TrimSpace(check.Name) == "" {
			errs = append(errs, fmt.Sprintf("Checks[%d]: name cannot be blank", i))
		}

		if err := validateSeverity(check); err != nil {
			errs = append(errs, fmt.Sprintf("Checks[%d] (%s): %v", i, check.Name, err))

		}

		if err := validateCheckTypes(check); err != nil {
			errs = append(errs, fmt.Sprintf("Checks[%d] (%s): %v", i, check.Name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("Configuration validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}
