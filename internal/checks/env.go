package checks

import (
	"fmt"
	"sync"

	"github.com/allenbiji/preboot/internal/detect"
	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

type EnvCheck struct {
	Key    string
	EnvMap map[string]string
}

func (e *EnvCheck) Execute() error {
	val, exists := e.EnvMap[e.Key]
	if !exists {
		return fmt.Errorf("key %q not found in .env", e.Key)
	}
	if val == "" {
		return fmt.Errorf("key %q is in .env but has no value", e.Key)
	}
	return nil
}

// cachedEnvMap holds the parsed .env contents for the lifetime of the process.
// cachedEnvMutex guards both the nil check and the write so concurrent callers are safe.
var (
	cachedEnvMap   map[string]string
	cachedEnvMutex sync.Mutex
)

func buildEnvExistsCheck(cfg model.CheckConfig) (registry.Check, error) {
	key, ok := cfg.Options["key"]
	if !ok || key == "" {
		return nil, fmt.Errorf("env_exists check requires a 'key' option")
	}

	cachedEnvMutex.Lock()
	if cachedEnvMap == nil {
		m, err := detect.ExtractEnvKeys(".env")
		if err != nil {
			cachedEnvMutex.Unlock()
			return nil, fmt.Errorf("could not read .env: %w", err)
		}
		cachedEnvMap = m
	}
	cachedEnvMutex.Unlock()

	return &EnvCheck{Key: key, EnvMap: cachedEnvMap}, nil
}

func init() {
	registry.Register(model.TypeEnvExists, buildEnvExistsCheck)
}
