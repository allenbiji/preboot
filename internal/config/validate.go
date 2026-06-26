package config

import (
	"fmt"
	"strings"

	"github.com/allenbiji/clone-sage/internal/model"
	"github.com/allenbiji/clone-sage/internal/registry"
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

// validateCheckTypes ensures the check driver exists in the registry.
func validateCheckTypes(check model.CheckConfig) error {
	if !registry.IsKnownType(check.Type) {
		return fmt.Errorf("unknown check type %q", check.Type)
	}
	return nil
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
