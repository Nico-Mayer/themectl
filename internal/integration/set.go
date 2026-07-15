package integration

import (
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
			zedDir := cfg.Settings.ConfigDirFor("zed")
			// TODO: improve this by moving to cfg
			usercfgDir, _ := os.UserConfigDir()
			return Zed{
				SettingsPath: filepath.Join(zedDir, "settings.json"),
				CurrentDir:   cfg.CurrentDir(),
				Installer: gitInstaller{
					extensionsDir: filepath.Join(usercfgDir, "Zed", "extensions", "installed"),
				},
			}
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
