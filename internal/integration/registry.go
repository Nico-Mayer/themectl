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
			Target:          cfg.Settings.Nvim.TargetOr(appConfigFile("nvim", "plugins", "99_theme.lua")),
			Binary:          "nvim",
		}
	},
	"eza": func(cfg config.Config) Integration {
		return SymlinkIntegration{
			IntegrationName: "eza",
			SourceFile:      filepath.Join(cfg.CurrentDir(), theme.EzaAssetName),
			Target:          cfg.Settings.Eza.TargetOr(appConfigFile("eza", "theme.yml")),
			Binary:          "eza",
		}
	},
	"yazi": func(cfg config.Config) Integration {
		return SymlinkIntegration{
			IntegrationName: "yazi",
			SourceFile:      filepath.Join(cfg.CurrentDir(), theme.YaziAssetName),
			Target:          cfg.Settings.Yazi.TargetOr(appConfigFile("yazi", "flavors", "themectl.yazi", "flavor.toml")),
			Binary:          "yazi",
		}
	},
	"system-appearance": newSystemAppearance,
	"wallpaper":         newWallpaper,
	"zed":               newZed,
	"vscode":            newVSCode,
	"rio":               newRio,
}

// platformOnly lists integrations that are registered only on some platforms.
// They stay in AllNames everywhere so theme files, settings files and the
// generated schemas do not differ per OS.
var platformOnly = []string{"windows-terminal"}

// Names returns the integrations registered on this platform. Doctor reports
// these, so an integration that cannot run here never shows up.
func Names() []string {
	return slices.Sorted(maps.Keys(available))
}

// AllNames returns every integration themectl knows about, including ones not
// built for this platform. Schema generation uses this so the committed schemas
// are identical everywhere.
func AllNames() []string {
	out := slices.Collect(maps.Keys(available))
	for _, name := range platformOnly {
		if _, ok := available[name]; !ok {
			out = append(out, name)
		}
	}
	slices.Sort(out)
	return out
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

// Unknown returns configured names themectl does not recognize. A name that is
// known but not built for this platform is not unknown, just inactive, so a
// shared settings file does not light up red on the wrong OS.
func Unknown(cfg config.Config) []string {
	known := AllNames()
	var out []string
	for _, name := range cfg.Settings.Integrations {
		if !slices.Contains(known, name) {
			out = append(out, name)
		}
	}
	return out
}
