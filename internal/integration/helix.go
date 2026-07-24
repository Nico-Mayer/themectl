package integration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

type Helix struct {
	ConfigFile string
}

func (Helix) Name() string {
	return "helix"
}

func (Helix) Supports(t theme.Resolved) bool {
	return t.Helix != nil && t.Helix.Theme != ""
}

func (h Helix) Apply(t theme.Resolved) error {
	name := t.Helix.Theme

	data, err := os.ReadFile(h.ConfigFile)
	if err != nil {
		return fmt.Errorf("read helix config: %w", err)
	}

	updated, err := setTOMLString(string(data), "theme", name)
	if err != nil {
		return err
	}

	if err := os.WriteFile(h.ConfigFile, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write helix config: %w", err)
	}

	return nil
}

func (h Helix) Check() error {
	return checkConfigDir(h.Name(), filepath.Dir(h.ConfigFile))
}

func newHelix(cfg config.Config) Integration {
	return Helix{ConfigFile: cfg.Settings.Helix.ConfigFileOr(appConfigFile("helix", "config.toml"))}
}
