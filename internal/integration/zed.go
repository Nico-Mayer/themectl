package integration

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Nico-Mayer/themectl/internal/cache"
	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/git"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

type Zed struct {
	SettingsPath string
	Installer    ExtensionInstaller
}

type ExtensionInstaller interface {
	Ensure(string) error
}

func (Zed) Name() string {
	return "zed"
}

func (z Zed) Apply(t theme.Resolved) error {
	spec := t.Zed
	if spec == nil || spec.Theme == "" {
		return fmt.Errorf("no zed override in theme %s: %w", t.ID(), ErrUnsupported)
	}

	if z.Installer != nil {
		for _, url := range spec.Extensions {
			url = git.NormalizeURL(url)
			if err := z.Installer.Ensure(url); err != nil {
				return err
			}
		}
	}

	data, err := os.ReadFile(z.SettingsPath)
	if err != nil {
		return fmt.Errorf("read zed settings: %w", err)
	}

	updated, err := setJSONCString(string(data), "theme", spec.Theme)
	if err != nil {
		return err
	}

	if spec.IconTheme == "" {
		spec.IconTheme = "Zed (Default)"
	}

	updated, err = setJSONCString(updated, "icon_theme", spec.IconTheme)
	if err != nil {
		return err
	}

	if err := os.WriteFile(z.SettingsPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write zed settings: %w", err)
	}

	return nil
}

func (z Zed) Check() error {
	return checkConfigDir(z.Name(), z.SettingsPath)
}

func newZed(cfg config.Config) Integration {
	z := Zed{
		SettingsPath: cfg.Settings.Zed.Path(defaultZedSettingsFile()),
	}

	usrConfigDir, err := os.UserConfigDir()
	if err != nil {
		slog.Warn("zed extension install disabled, user config dir not found", "err", err)
		return z
	}
	z.Installer = gitInstaller{
		extensionsDir: filepath.Join(usrConfigDir, "Zed", "extensions", "installed"),
		cache:         cache.New(filepath.Join(cfg.CacheDir(), "zed")),
	}
	return z
}

func defaultZedSettingsFile() string {
	if runtime.GOOS == "windows" {
		if dir, err := os.UserConfigDir(); err == nil {
			return filepath.Join(dir, "zed", "settings.json")
		}
	}
	return defaultConfigFile("zed", "settings.json")
}
