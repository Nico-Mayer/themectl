package integration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Nico-Mayer/themectl/internal/theme"
)

type Integration interface {
	Name() string
	Apply(t theme.Resolved) error
}

type HealthChecker interface {
	Check() error
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
