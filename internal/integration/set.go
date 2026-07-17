package integration

import (
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/nico-mayer/themectl-cli/internal/config"
)

var available = map[string]func(cfg config.Config) Integration{
	"ghostty": func(cfg config.Config) Integration {
		return Ghostty{ConfigPath: filepath.Join(cfg.Settings.ConfigDirFor("ghostty"), "config.ghostty")}
	},
	"helix": func(cfg config.Config) Integration {
		return Helix{ConfigPath: filepath.Join(cfg.Settings.ConfigDirFor("helix"), "config.toml")}
	},
	"nvim": func(cfg config.Config) Integration {
		return Nvim{Cfg: cfg}
	},
	"eza": func(cfg config.Config) Integration {
		return Eza{Cfg: cfg}
	},
	"yazi": func(cfg config.Config) Integration {
		return Yazi{Cfg: cfg}
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
			SettingsPath: filepath.Join(cfg.Settings.ConfigDirFor("zed"), "settings.json"),
			CurrentDir:   cfg.CurrentDir(),
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
