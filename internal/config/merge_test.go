package config_test

import (
	"testing"

	"github.com/allenbiji/preboot/internal/config"
	"github.com/allenbiji/preboot/internal/model"
)

func TestMergeDefaults(t *testing.T) {
	tests := []struct {
		name        string
		defaults    map[string]interface{}
		wantStrict  interface{}
		wantTimeout interface{}
	}{
		{
			name:        "nil defaults",
			defaults:    nil,
			wantStrict:  true,
			wantTimeout: 3000,
		},
		{
			name:        "existing strict false preserved",
			defaults:    map[string]interface{}{"strict": false},
			wantStrict:  false,
			wantTimeout: 3000,
		},
		{
			name:        "existing timeout preserved",
			defaults:    map[string]interface{}{"timeout_ms": 5000},
			wantStrict:  true,
			wantTimeout: 5000,
		},
		{
			name:        "both set — both preserved",
			defaults:    map[string]interface{}{"strict": true, "timeout_ms": 1000},
			wantStrict:  true,
			wantTimeout: 1000,
		},
		{
			name:        "empty map — both defaults added",
			defaults:    map[string]interface{}{},
			wantStrict:  true,
			wantTimeout: 3000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &model.PrebootConfig{Defaults: tt.defaults}
			config.MergeDefaults(c)

			if c.Defaults["strict"] != tt.wantStrict {
				t.Errorf("strict: got %v, want %v", c.Defaults["strict"], tt.wantStrict)
			}
			if c.Defaults["timeout_ms"] != tt.wantTimeout {
				t.Errorf("timeout_ms: got %v, want %v", c.Defaults["timeout_ms"], tt.wantTimeout)
			}
		})
	}
}
