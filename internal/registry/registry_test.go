package registry_test

import (
	"context"
	"strings"
	"testing"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

type stubCheck struct{}

func (s *stubCheck) Execute(_ context.Context) error { return nil }

func TestRegister_Build(t *testing.T) {
	const typ model.CheckType = "stub_for_registry_test"
	registry.Register(typ, func(cfg model.CheckConfig) (registry.Check, error) {
		return &stubCheck{}, nil
	})
	check, err := registry.Build(model.CheckConfig{Type: typ})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if check == nil {
		t.Fatal("expected non-nil check")
	}
}

func TestRegister_DuplicatePanics(t *testing.T) {
	const typ model.CheckType = "stub_duplicate_panic_test"
	registry.Register(typ, func(cfg model.CheckConfig) (registry.Check, error) {
		return &stubCheck{}, nil
	})
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic on duplicate registration, got none")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, string(typ)) {
			t.Errorf("panic message %q does not mention type %q", r, typ)
		}
	}()
	registry.Register(typ, func(cfg model.CheckConfig) (registry.Check, error) {
		return &stubCheck{}, nil
	})
}

func TestBuild_UnknownType(t *testing.T) {
	_, err := registry.Build(model.CheckConfig{Type: "completely_unknown_xyz"})
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}
	if !strings.Contains(err.Error(), "unknown check type") {
		t.Errorf("error %q does not contain 'unknown check type'", err.Error())
	}
}

func TestIsKnownType(t *testing.T) {
	const typ model.CheckType = "stub_isknown_test"
	if registry.IsKnownType(typ) {
		t.Error("unregistered type should return false")
	}
	registry.Register(typ, func(cfg model.CheckConfig) (registry.Check, error) {
		return &stubCheck{}, nil
	})
	if !registry.IsKnownType(typ) {
		t.Error("registered type should return true")
	}
}
