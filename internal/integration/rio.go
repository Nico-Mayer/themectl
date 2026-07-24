package integration

import (
	"os"
	"path/filepath"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

type Rio struct {
	ConfigPath string
	SourceFile string
}

func (r Rio) Name() string { return "rio" }

func (r Rio) Apply(t theme.Resolved) error {
	before, err := os.ReadFile(r.ConfigPath)
	if err != nil {
		return err
	}

	updated, err := setTOMLString(string(before), "theme", "themectl")
	if err != nil {
		return err
	}

	err = os.WriteFile(r.ConfigPath, []byte(updated), 0o644)
	if err != nil {
		return err
	}

	configDir := filepath.Dir(r.ConfigPath)
	err = symlink(r.SourceFile, filepath.Join(configDir, "themes", "themectl.toml"))
	if err != nil {
		return err
	}

	return nil
}

func (r Rio) Check() error {
	return checkConfigDir(r.Name(), r.ConfigPath)
}

func (r Rio) Supports(t theme.Resolved) bool {
	_, err := os.Stat(r.SourceFile)
	return err == nil
}

func newRio(cfg config.Config) Integration {
	return Rio{
		ConfigPath: cfg.Settings.Rio.Path(appConfigFile("rio", "config.toml")),
		SourceFile: filepath.Join(cfg.CurrentDir(), theme.RioAssetName),
	}
}
