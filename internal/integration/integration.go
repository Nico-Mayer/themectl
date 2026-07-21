package integration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Nico-Mayer/themectl/internal/theme"
)

type Integration interface {
	Name() string
	Check() error
	Apply(t theme.Resolved) error
}

func checkConfigDir(name, path string) error {
	if path == "" {
		return fmt.Errorf("no config dir configured for %s", name)
	}
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		return fmt.Errorf("%s config dir missing: %w", name, err)
	}
	return nil
}

func checkFileExists(desc, path string) error {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s not found", desc)
	}
	if err != nil {
		return fmt.Errorf("check %s: %w", desc, err)
	}
	return nil
}
