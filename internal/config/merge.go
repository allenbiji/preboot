package config

import "github.com/allenbiji/preboot/internal/model"

// this function is to ensure that the system always has defaults to fallback to
func MergeDefaults(c *model.PrebootConfig) {
	if c.Defaults == nil {
		c.Defaults = make(map[string]interface{})
	}

	if _, exists := c.Defaults["strict"]; !exists {
		c.Defaults["strict"] = true
	}

	if _, exists := c.Defaults["timeout_ms"]; !exists {
		c.Defaults["timeout_ms"] = 3000
	}

	for i := range c.Checks {
		if c.Checks[i].Severity == "" {
			c.Checks[i].Severity = model.SeverityBlocker
		}
	}
}
