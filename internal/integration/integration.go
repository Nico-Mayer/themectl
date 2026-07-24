package integration

import (
	"errors"
	"fmt"
	"os"

	"github.com/Nico-Mayer/themectl/internal/theme"
)

type Integration interface {
	Name() string
	Check() error
	Supports(t theme.Resolved) bool
	Apply(t theme.Resolved) error
}

type Resetter interface {
	Reset() error
}

func checkConfigDir(name, dir string) error {
	if dir == "" {
		return fmt.Errorf("no config dir configured for %s", name)
	}
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("%s config dir missing: %w", name, err)
	}
	return nil
}

func checkDirExists(desc, dir string) error {
	_, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s not found", desc)
	}
	if err != nil {
		return fmt.Errorf("check %s: %w", desc, err)
	}
	return nil
}
