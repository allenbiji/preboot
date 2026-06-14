package config

import (
	"fmt"

	"github.com/allenbiji/clone-sage/internal/model"
	"github.com/spf13/viper"
)

func Load() (*model.ClonesageConfig, error) {
	v := viper.New()

	v.SetConfigFile("sage-auto.yaml")

	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("Error reading config file: %w", err)
		}
	}

	v.SetConfigFile("sage.yaml")
	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("Error merging config files: %w", err)
	}

	var cfg model.ClonesageConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("There was an error in unmarshaling the yaml config: %w", err)
	}

	if err := ValidateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("There was an error in validating the yaml configs: %w", err)
	}

	return &cfg, nil
}
