package integration

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/nico-mayer/themectl-cli/internal/config"
)

func Enabled(cfg config.Config) []Integration {
	available := map[string]func() Integration{
		"ghostty": func() Integration {
			return Ghostty{ConfigPath: filepath.Join(cfg.Settings.ConfigDirFor("ghostty"), "config.ghostty")}
		},
		"helix": func() Integration {
			return Helix{ConfigPath: filepath.Join(cfg.Settings.ConfigDirFor("helix"), "config.toml")}
		},
		"nvim": func() Integration {
			return Nvim{Cfg: cfg}
		},
		"eza": func() Integration {
			return Eza{Cfg: cfg}
		},
		"yazi": func() Integration {
			return Yazi{Cfg: cfg}
		},
		"system-appearance": func() Integration {
			return SystemAppearance{}
		},
		"wallpaper": func() Integration {
			return Wallpaper{
				ThemesDir:           cfg.ThemesDir(),
				SharedWallpapersDir: cfg.SharedWallpapersDir(),
			}
		},
		"zed": func() Integration {
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

	var out []Integration
	for _, name := range cfg.Settings.Integrations {
		i, ok := available[name]
		if ok {
			out = append(out, i())
		}
	}

	return out
}
