package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Settings struct {
	Integrations []string        `toml:"integrations,omitempty" jsonschema:"description=Integrations to run on theme apply. Replaces the default list.,uniqueItems=true"`
	Ghostty      FileSettings    `toml:"ghostty,omitempty" jsonschema:"description=Ghostty integration settings."`
	Helix        FileSettings    `toml:"helix,omitempty" jsonschema:"description=Helix integration settings."`
	VSCode       FileSettings    `toml:"vscode,omitempty" jsonschema:"description=VS Code integration settings."`
	Zed          FileSettings    `toml:"zed,omitempty" jsonschema:"description=Zed integration settings."`
	Rio          FileSettings    `toml:"rio,omitempty" jsonschema:"description=Rio integration settings."`
	Nvim         SymlinkSettings `toml:"nvim,omitempty" jsonschema:"description=Neovim integration settings."`
	Eza          SymlinkSettings `toml:"eza,omitempty" jsonschema:"description=Eza integration settings."`
	Yazi         SymlinkSettings `toml:"yazi,omitempty" jsonschema:"description=Yazi integration settings."`
}

type FileSettings struct {
	ConfigFile string `toml:"config_file,omitempty" jsonschema:"description=Path to the file themectl edits. Supports env vars ($VAR) and a leading ~."`
}

type SymlinkSettings struct {
	Target    string `toml:"target,omitempty" jsonschema:"description=Where the symlink is created. Supports env vars ($VAR) and a leading ~. When moved outside the app's default config dir also set config_dir so the health check probes the right place."`
	ConfigDir string `toml:"config_dir,omitempty" jsonschema:"description=App config dir probed by the health check; the integration is skipped when it is missing. Defaults to the app's standard config dir. Supports env vars ($VAR) and a leading ~."`
}

func (f FileSettings) Path(fallback string) string {
	p := strings.TrimSpace(f.ConfigFile)
	if p == "" {
		return fallback
	}
	return expandPath(p)
}

func (s SymlinkSettings) Path(fallback string) string {
	p := strings.TrimSpace(s.Target)
	if p == "" {
		return fallback
	}
	return expandPath(p)
}

func (s SymlinkSettings) Dir(fallback string) string {
	p := strings.TrimSpace(s.ConfigDir)
	if p == "" {
		return fallback
	}
	return expandPath(p)
}

func loadSettings(path string) (Settings, error) {
	s := defaultSettings()

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("read settings: %w", err)
	}

	if err := toml.Unmarshal(data, &s); err != nil {
		return Settings{}, fmt.Errorf("parse settings: %w", err)
	}
	return s, nil
}

func defaultSettings() Settings {
	return Settings{
		Integrations: []string{
			"ghostty",
			"zed",
			"vscode",
			"system-appearance",
			"wallpaper",
			"yazi",
			"eza",
			"nvim",
			"helix",
			"rio",
		},
	}
}

func expandPath(path string) string {
	path = os.ExpandEnv(path)
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}
