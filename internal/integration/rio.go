package integration

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

type Rio struct {
	ConfigFile string
	SourceFile string
}

func (r Rio) Name() string { return "rio" }

func (r Rio) Apply(t theme.Resolved) error {
	before, err := os.ReadFile(r.ConfigFile)
	if err != nil {
		return err
	}

	updated, err := setTOMLString(string(before), "theme", "themectl")
	if err != nil {
		return err
	}

	err = os.WriteFile(r.ConfigFile, []byte(updated), 0o644)
	if err != nil {
		return err
	}

	configDir := filepath.Dir(r.ConfigFile)
	err = symlink(r.SourceFile, filepath.Join(configDir, "themes", "themectl.toml"))
	if err != nil {
		return err
	}

	return nil
}

func (r Rio) Check() error {
	return checkConfigDir(r.Name(), filepath.Dir(r.ConfigFile))
}

func (r Rio) Supports(t theme.Resolved) bool {
	_, err := os.Stat(r.SourceFile)
	if err != nil {
		slog.Debug("theme does not support rio, reseeting theme key in config", "theme", t.ID())
		setTOMLString(r.ConfigFile, "theme", "")
	}
	return err == nil
}

func newRio(cfg config.Config) Integration {
	return Rio{
		ConfigFile: cfg.Settings.Rio.ConfigFileOr(appConfigFile("rio", "config.toml")),
		SourceFile: filepath.Join(cfg.CurrentDir(), theme.RioAssetName),
	}
}
