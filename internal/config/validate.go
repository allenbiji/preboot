package config

import (
	"fmt"
	"strings"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

// validateSeverity ensures the string cast by Viper matches our strict enums.
func validateSeverity(check model.CheckConfig) error {
	switch check.Severity {
	case model.SeverityInfo, model.SeverityBlocker, model.SeverityWarning:
		return nil
	default:
		return fmt.Errorf("invalid severity %q (allowed: info, warning, blocker)", check.Severity)
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
func ValidateConfig(cfg *model.PrebootConfig) error {
	var errs []string

	if cfg.Version != 1 {
		errs = append(errs, fmt.Sprintf("unsupported config version: %d", cfg.Version))
	}

	if v, exists := cfg.Defaults["strict"]; exists {
		if _, ok := v.(bool); !ok {
			errs = append(errs, fmt.Sprintf("defaults.strict must be a boolean, got %T", v))
		}
	}
	if v, exists := cfg.Defaults["timeout_ms"]; exists {
		switch v.(type) {
		case int, int64, float64:
			// ok
		default:
			errs = append(errs, fmt.Sprintf("defaults.timeout_ms must be an integer, got %T", v))
		}
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
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}
