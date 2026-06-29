package checks

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

type DirectoryCheck struct {
	Folder string
}

// execute method for the check
func (d *DirectoryCheck) Execute() error {
	info, err := os.Stat(d.Folder)
	if os.IsNotExist(err) {
		return fmt.Errorf("The folder %s does not exist", d.Folder)
	}
	if err != nil {
		return fmt.Errorf("error accessing directory %s: %w", d.Folder, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("Expected a directory %s, but detected a file", d.Folder)
	}

	return nil
}

// build factory for the check
func buildDirectoryExistsCheck(cfg model.CheckConfig) (registry.Check, error) {
	folder, ok := cfg.Options["folder"]
	if !ok || folder == "" {
		return nil, fmt.Errorf("directory_exists check requires a 'folder' option")
	}
	if err := validateRelativePath(folder, "folder"); err != nil {
		return nil, err
	}
	return &DirectoryCheck{
		Folder: filepath.Clean(folder),
	}, nil
}

// register check in registry
func init() {
	registry.Register(model.TypeDirectoryExists, buildDirectoryExistsCheck)
}
