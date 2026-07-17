package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
)

type Settings struct {
	Integrations []string          `toml:"integrations,omitempty" jsonschema:"description=Integrations to run on theme apply. Replaces the default list.,uniqueItems=true"`
	DefaultTheme string            `toml:"default-theme,omitempty" jsonschema:"pattern=^[^/]+/[^/]+$,description=Theme id in family/variant form."`
	ConfigDirs   map[string]string `toml:"config-dirs,omitempty" jsonschema:"description=Per-integration config dir overrides. Supports env vars ($VAR) and a leading ~."`
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
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}

	winConfigHome := ""
	if runtime.GOOS == "windows" {
		winConfigHome, _ = os.UserConfigDir()
	}

	return Settings{
		Integrations: []string{
			"ghostty",
			"zed",
			"system-appearance",
			"wallpaper",
			"yazi",
			"eza",
			"nvim",
			"helix",
		},
		ConfigDirs: defaultConfigDirs(home, winConfigHome),
	}
}

func defaultConfigDirs(home, winConfigHome string) map[string]string {
	dirs := map[string]string{
		"ghostty": filepath.Join(home, ".config", "ghostty"),
		"helix":   filepath.Join(home, ".config", "helix"),
		"zed":     filepath.Join(home, ".config", "zed"),
		"yazi":    filepath.Join(home, ".config", "yazi"),
	}
	if winConfigHome != "" {
		dirs["zed"] = filepath.Join(winConfigHome, "zed")
		dirs["yazi"] = filepath.Join(winConfigHome, "yazi", "config")
	}
	return dirs
}

func (s Settings) ConfigDirFor(integration string) string {
	path := strings.TrimSpace(s.ConfigDirs[integration])
	if path == "" {
		return ""
	}
	return expandPath(path)
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
