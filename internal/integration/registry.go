package integration

import (
	"maps"
	"path/filepath"
	"slices"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

var available = map[string]func(cfg config.Config) Integration{
	"ghostty": newGhostty,
	"helix":   newHelix,
	"nvim": func(cfg config.Config) Integration {
		return SymlinkIntegration{
			IntegrationName: "nvim",
			SourceFile:      filepath.Join(cfg.CurrentDir(), theme.NvimAssetName),
			Target:          cfg.Settings.Nvim.Path(appConfigFile("nvim", "plugins", "99_theme.lua")),
			AppConfigDir:    cfg.Settings.Nvim.Dir(appConfigDir("nvim")),
		}
	},
	"eza": func(cfg config.Config) Integration {
		return SymlinkIntegration{
			IntegrationName: "eza",
			SourceFile:      filepath.Join(cfg.CurrentDir(), theme.EzaAssetName),
			Target:          cfg.Settings.Eza.Path(appConfigFile("eza", "theme.yml")),
			AppConfigDir:    cfg.Settings.Eza.Dir(appConfigDir("eza")),
		}
	},
	"yazi": func(cfg config.Config) Integration {
		return SymlinkIntegration{
			IntegrationName: "yazi",
			SourceFile:      filepath.Join(cfg.CurrentDir(), theme.YaziAssetName),
			Target:          cfg.Settings.Yazi.Path(appConfigFile("yazi", "flavors", "themectl.yazi", "flavor.toml")),
			AppConfigDir:    cfg.Settings.Yazi.Dir(appConfigDir("yazi")),
		}
	},
	"system-appearance": newSystemAppearance,
	"wallpaper":         newWallpaper,
	"zed":               newZed,
	"vscode":            newVSCode,
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
