package checks_test

import "github.com/allenbiji/preboot/internal/model"

// cfg builds a minimal CheckConfig for registry.Build factory tests.
func cfg(typ model.CheckType, opts map[string]string) model.CheckConfig {
	return model.CheckConfig{Name: "t", Type: typ, Severity: model.SeverityBlocker, Options: opts}
}
