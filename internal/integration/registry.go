package integration

import (
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/Nico-Mayer/themectl/internal/config"
)

var available = map[string]func(cfg config.Config) Integration{
	"ghostty": func(cfg config.Config) Integration {
		return Ghostty{ConfigPath: cfg.Settings.Ghostty.Path(defaultConfigFile("ghostty", "config.ghostty"))}
	},
	"helix": func(cfg config.Config) Integration {
		return Helix{ConfigPath: cfg.Settings.Helix.Path(defaultConfigFile("helix", "config.toml"))}
	},
	"nvim": func(cfg config.Config) Integration {
		return SymlinkIntegration{
			IntegrationName: "nvim",
			SourceFile:      filepath.Join(cfg.CurrentDir(), "nvim.lua"),
			Target:          cfg.Settings.Nvim.Path(filepath.Join(homeConfig(), "nvim", "plugin", "99_theme.lua")),
		}
	},
	"eza": func(cfg config.Config) Integration {
		return SymlinkIntegration{
			IntegrationName: "eza",
			SourceFile:      filepath.Join(cfg.CurrentDir(), "eza.yml"),
			Target:          cfg.Settings.Eza.Path(filepath.Join(homeConfig(), "eza", "theme.yml")),
		}
	},
	"yazi": func(cfg config.Config) Integration {
		return SymlinkIntegration{
			IntegrationName: "yazi",
			SourceFile:      filepath.Join(cfg.CurrentDir(), "yazi-flavor.toml"),
			Target:          cfg.Settings.Yazi.Path(filepath.Join(homeConfig(), "yazi", "flavors", "themectl.yazi", "flavor.toml")),
		}
	},
	"system-appearance": func(cfg config.Config) Integration {
		return SystemAppearance{}
	},
	"wallpaper": func(cfg config.Config) Integration {
		return Wallpaper{
			ThemesDir:           cfg.ThemesDir(),
			SharedWallpapersDir: cfg.SharedWallpapersDir(),
		}
	},
	"zed": func(cfg config.Config) Integration {
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
		}
		return z
	},
}

func defaultConfigFile(app, file string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", app, file)
}

func defaultZedSettingsFile() string {
	if runtime.GOOS == "windows" {
		if dir, err := os.UserConfigDir(); err == nil {
			return filepath.Join(dir, "zed", "settings.json")
		}
	}
	return defaultConfigFile("zed", "settings.json")
}

func Names() []string {
	return slices.Sorted(maps.Keys(available))
}

func Enabled(cfg config.Config) []Integration {
	var out []Integration
	for _, name := range cfg.Settings.Integrations {
		i, ok := available[name]
		if ok {
			out = append(out, i(cfg))
		}
	}

	return out
}

func Unknown(cfg config.Config) []string {
	var out []string
	for _, name := range cfg.Settings.Integrations {
		if _, ok := available[name]; !ok {
			out = append(out, name)
		}
	}
	return out
}

func homeConfig() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}
