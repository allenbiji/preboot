package config

import (
	"fmt"
	"os"

	"github.com/allenbiji/clone-sage/internal/model"
	"github.com/spf13/viper"
)

func Load() (*model.ClonesageConfig, error) {
	autoCfg, errAuto := readConfigFile("sage-auto.yml")

	explicitCfg, errExplicit := readConfigFile("sage.yml")

	if os.IsNotExist(errAuto) && os.IsNotExist(errExplicit) {
		return nil, fmt.Errorf("No config files found. Run 'sage init' to generate config files")
	}

	var finalCfg *model.ClonesageConfig
	if explicitCfg == nil {
		fmt.Println("Using auto-generated config (no sage.yml found).")
		finalCfg = autoCfg
	} else if autoCfg == nil {
		fmt.Println("Using explicit config (no sage-auto.yml found).")
		finalCfg = explicitCfg
	} else {
		fmt.Println("Merging sage-auto.yml with sage.yml...")
		finalCfg = mergeConfigs(autoCfg, explicitCfg)
	}

	MergeDefaults(finalCfg)

	if err := ValidateConfig(finalCfg); err != nil {
		return nil, fmt.Errorf("There was an error in validating the yaml configs: %w", err)
	}

	return finalCfg, nil
}

// this function id used to unmarshal the given file into a clonesageconfig struct
func readConfigFile(filename string) (*model.ClonesageConfig, error) {
	v := viper.New()

	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}

	v.SetConfigFile(filename)

	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("Error reading config file: %w", err)
		}
	}

	var cfg model.ClonesageConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("There was an error while unmarshalling: %w", err)
	}

	return &cfg, nil
}

// mergeConfigs safely layers the explicit config over the auto config
func mergeConfigs(auto, explicit *model.ClonesageConfig) *model.ClonesageConfig {
	merged := &model.ClonesageConfig{
		Version: explicit.Version,
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
