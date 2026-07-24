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
	return err == nil
}

func (r Rio) Reset() error {
	slog.Debug("resetting rio theme", "config", r.ConfigFile, "source", r.SourceFile)
	data, err := os.ReadFile(r.ConfigFile)
	if err != nil {
		slog.Debug("failed to read rio config", "err", err)
		return err
	}
	updated, err := setTOMLString(string(data), "theme", "")
	if err != nil {
		slog.Debug("failed to set theme", "err", err)
		return err
	}
	return os.WriteFile(r.ConfigFile, []byte(updated), 0o644)
}

func newRio(cfg config.Config) Integration {
	return Rio{
		ConfigFile: cfg.Settings.Rio.ConfigFileOr(appConfigFile("rio", "config.toml")),
		SourceFile: filepath.Join(cfg.CurrentDir(), theme.RioAssetName),
	}
}
