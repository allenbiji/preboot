package checks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/allenbiji/preboot/internal/model"
	"github.com/allenbiji/preboot/internal/registry"
)

type DirectoryCheck struct {
	Folder string
}

func (d *DirectoryCheck) Execute(_ context.Context) error {
	info, err := os.Stat(d.Folder)
	if os.IsNotExist(err) {
		return fmt.Errorf("folder does not exist: %s", d.Folder)
	}
	if err != nil {
		return fmt.Errorf("error accessing directory %s: %w", d.Folder, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("expected a directory but found a file: %s", d.Folder)
	}
	return nil
}

func buildDirectoryExistsCheck(cfg model.CheckConfig) (registry.Check, error) {
	folder, ok := cfg.Options["folder"]
	if !ok || folder == "" {
		return nil, fmt.Errorf("directory_exists check requires a 'folder' option")
	}
	if err := validateRelativePath(folder, "folder"); err != nil {
		return nil, err
	}
	return &DirectoryCheck{Folder: filepath.Clean(folder)}, nil
}

func init() {
	registry.Register(model.TypeDirectoryExists, buildDirectoryExistsCheck)
}
