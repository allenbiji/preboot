package checks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

type FileCheck struct {
	Path string
}

// execute method for file check
func (f *FileCheck) Execute() error {
	info, err := os.Stat(f.Path)
	if os.IsNotExist(err) {
		return fmt.Errorf("File does not exist: %s", f.Path)
	}

	if err != nil {
		return fmt.Errorf("Error accessing file %s: %w", f.Path, err)
	}

	if info.IsDir() {
		return fmt.Errorf("Expected a file, returned a directory at %s", f.Path)
	}

	return nil
}

func validateRelativePath(path, field string) error {
	if filepath.IsAbs(path) {
		return fmt.Errorf("%s %q must be a relative path, not an absolute path", field, path)
	}
	if strings.HasPrefix(path, "~") {
		return fmt.Errorf("%s %q must not be a home-directory path", field, path)
	}
	for _, part := range strings.Split(filepath.Clean(path), string(filepath.Separator)) {
		if part == ".." {
			return fmt.Errorf("%s %q must not traverse parent directories", field, path)
		}
	}
	return nil
}

// creates file check factory
func buildFileCheck(cfg model.CheckConfig) (registry.Check, error) {
	path, ok := cfg.Options["path"]
	if !ok || path == "" {
		return nil, fmt.Errorf("file_exists check requires a 'path' option")
	}
	if err := validateRelativePath(path, "path"); err != nil {
		return nil, err
	}
	return &FileCheck{
		Path: filepath.Clean(path),
	}, nil
}

// Registers file exists check in registry
func init() {
	registry.Register(model.TypeFileExists, buildFileCheck)
}
