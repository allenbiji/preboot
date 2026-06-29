package registry

import (
	"fmt"

	"github.com/allenbiji/preboot/internal/model"
)

// Check is the strict contract that every single diagnostic test must follow.
type Check interface {
	Execute() error
}

// Factory is a function that takes the user's config and returns an executable Check.
type Factory func(cfg model.CheckConfig) (Check, error)

// backend is the private, centralized map of all known check types.
var backend = make(map[model.CheckType]Factory)

// Register is called by individual checks during application startup.
func Register(checkType model.CheckType, factory Factory) {
	if _, exists := backend[checkType]; exists {
		panic(fmt.Sprintf("Check type %s is already registered", checkType))
	}

	backend[checkType] = factory
}

// Build looks up the check type in the registry and constructs it.
func Build(cfg model.CheckConfig) (Check, error) {
	factory, exists := backend[cfg.Type]
	if !exists {
		return nil, fmt.Errorf("unknown check type: %s", cfg.Type)
	}
	return factory(cfg)
}

func IsKnownType(checkType model.CheckType) bool {
	_, exists := backend[checkType]
	return exists
}
