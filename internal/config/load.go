package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/spf13/viper"
)

func Load() (*model.PrebootConfig, error) {
	autoCfg, errAuto := readConfigFile("preboot-auto.yml")

	explicitCfg, errExplicit := readConfigFile("preboot.yml")

	if errors.Is(errAuto, fs.ErrNotExist) && errors.Is(errExplicit, fs.ErrNotExist) {
		return nil, fmt.Errorf("No config files found. Run 'preboot init' to generate config files")
	}

	if errAuto != nil && !errors.Is(errAuto, fs.ErrNotExist) {
		return nil, errAuto
	}
	if errExplicit != nil && !errors.Is(errExplicit, fs.ErrNotExist) {
		return nil, errExplicit
	}

	var finalCfg *model.PrebootConfig
	if explicitCfg == nil {
		fmt.Fprintln(os.Stderr, "Using auto-generated config (no preboot.yml found).")
		finalCfg = autoCfg
	} else if autoCfg == nil {
		fmt.Fprintln(os.Stderr, "Using explicit config (no preboot-auto.yml found).")
		finalCfg = explicitCfg
	} else {
		fmt.Fprintln(os.Stderr, "Merging preboot-auto.yml with preboot.yml...")
		finalCfg = mergeConfigs(autoCfg, explicitCfg)
	}

	MergeDefaults(finalCfg)

	if err := ValidateConfig(finalCfg); err != nil {
		return nil, err
	}

	return finalCfg, nil
}

// LoadFrom loads a single named config file when --config is specified.
// When customPath is empty it falls back to the standard two-file merge via Load.
func LoadFrom(customPath string) (*model.PrebootConfig, error) {
	if customPath != "" {
		cfg, err := readConfigFile(customPath)
		if err != nil {
			return nil, fmt.Errorf("loading %s: %w", customPath, err)
		}
		MergeDefaults(cfg)
		if err := ValidateConfig(cfg); err != nil {
			return nil, fmt.Errorf("validating %s: %w", customPath, err)
		}
		return cfg, nil
	}
	return Load()
}

// this function is used to unmarshal the given file into a prebootconfig struct
func readConfigFile(filename string) (*model.PrebootConfig, error) {
	v := viper.New()

	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}

	v.SetConfigFile(filename)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error reading config file: %w", err)
	}

	var cfg model.PrebootConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("There was an error while unmarshalling: %w", err)
	}

	return &cfg, nil
}

// mergeConfigs safely layers the explicit config over the auto config
func mergeConfigs(auto, explicit *model.PrebootConfig) *model.PrebootConfig {
	merged := &model.PrebootConfig{
		Version:  explicit.Version,
		Defaults: make(map[string]interface{}),
	}
	if merged.Version == 0 {
		merged.Version = auto.Version // Fallback if explicit didn't define it
	}

	for k, v := range auto.Defaults {
		merged.Defaults[k] = v
	}
	for k, v := range explicit.Defaults {
		merged.Defaults[k] = v
	}

	checkMap := make(map[string]model.CheckConfig)
	var order []string

	for _, c := range auto.Checks {
		checkMap[c.Name] = c
		order = append(order, c.Name)
	}

	for _, c := range explicit.Checks {
		if _, exists := checkMap[c.Name]; !exists {
			order = append(order, c.Name)
		}
		checkMap[c.Name] = c
	}

	for _, name := range order {
		merged.Checks = append(merged.Checks, checkMap[name])
	}

	return merged
}
