package checks

import (
	"context"
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

// Execute uses os.Stat (follows symlinks) to confirm the path exists and is a file.
// A symlink pointing to a file passes; a broken symlink or a path that resolves to
// a directory fails. File contents are never read.
func (f *FileCheck) Execute(_ context.Context) error {
	info, err := os.Stat(f.Path)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", f.Path)
	}
	if err != nil {
		return fmt.Errorf("error accessing file %s: %w", f.Path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("expected a file but found a directory: %s", f.Path)
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

func buildFileCheck(cfg model.CheckConfig) (registry.Check, error) {
	path, ok := cfg.Options["path"]
	if !ok || path == "" {
		return nil, fmt.Errorf("file_exists check requires a 'path' option")
	}
	if err := validateRelativePath(path, "path"); err != nil {
		return nil, err
	}
	return &FileCheck{Path: filepath.Clean(path)}, nil
}

func init() {
	registry.Register(model.TypeFileExists, buildFileCheck)
}
