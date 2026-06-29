package config_test

import (
	"strings"
	"testing"

	_ "github.com/allenbiji/preboot/internal/checks"
	"github.com/allenbiji/preboot/internal/config"
	"github.com/allenbiji/preboot/internal/model"
)

func validCfg() *model.PrebootConfig {
	return &model.PrebootConfig{
		Version: 1,
		Checks: []model.CheckConfig{
			{Name: "x", Type: model.TypeFileExists, Severity: model.SeverityBlocker},
		},
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*model.PrebootConfig)
		wantErr string
	}{
		{
			name:   "valid config",
			mutate: func(c *model.PrebootConfig) {},
		},
		{
			name:    "version 0",
			mutate:  func(c *model.PrebootConfig) { c.Version = 0 },
			wantErr: "Unsupported config version: 0",
		},
		{
			name:    "version 2",
			mutate:  func(c *model.PrebootConfig) { c.Version = 2 },
			wantErr: "Unsupported config version: 2",
		},
		{
			name:    "blank name",
			mutate:  func(c *model.PrebootConfig) { c.Checks[0].Name = "" },
			wantErr: "name cannot be blank",
		},
		{
			name:    "whitespace name",
			mutate:  func(c *model.PrebootConfig) { c.Checks[0].Name = "   " },
			wantErr: "name cannot be blank",
		},
		{
			name:    "invalid severity",
			mutate:  func(c *model.PrebootConfig) { c.Checks[0].Severity = "critical" },
			wantErr: "Invalid severity",
		},
		{
			name:    "unknown check type",
			mutate:  func(c *model.PrebootConfig) { c.Checks[0].Type = "does_not_exist" },
			wantErr: "unknown check type",
		},
		// s40: blank type field (omitted or empty string) — treated as unknown type.
		{
			name:    "blank type",
			mutate:  func(c *model.PrebootConfig) { c.Checks[0].Type = "" },
			wantErr: "unknown check type",
		},
		// s56: blank severity — ValidateConfig rejects it directly; MergeDefaults provides
		// the default in the production Load() flow before ValidateConfig is called.
		{
			name:    "blank severity",
			mutate:  func(c *model.PrebootConfig) { c.Checks[0].Severity = "" },
			wantErr: "Invalid severity",
		},
		{
			name: "multiple errors accumulated",
			mutate: func(c *model.PrebootConfig) {
				c.Checks[0].Name = ""
				c.Checks[0].Severity = "critical"
			},
			wantErr: "name cannot be blank",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := validCfg()
			tt.mutate(c)
			err := config.ValidateConfig(c)
			if (err != nil) != (tt.wantErr != "") {
				t.Fatalf("wantErr=%q got=%v", tt.wantErr, err)
			}
			if err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_MultipleErrors(t *testing.T) {
	c := validCfg()
	c.Checks[0].Name = ""
	c.Checks[0].Severity = "critical"
	err := config.ValidateConfig(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "name cannot be blank") {
		t.Errorf("error %q missing 'name cannot be blank'", err.Error())
	}
	if !strings.Contains(err.Error(), "Invalid severity") {
		t.Errorf("error %q missing 'Invalid severity'", err.Error())
	}
}
