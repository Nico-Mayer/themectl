package integration

import (
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
		"eza": func() Integration {
			return Eza{Cfg: cfg}
		},
		"yazi": func() Integration {
			return Yazi{Cfg: cfg}
		},
		"system-appearance": func() Integration {
			return SystemAppearance{}
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
