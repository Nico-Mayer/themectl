package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Settings struct {
	Integrations []string          `json:"integrations"`
	DefaultTheme string            `json:"default-theme,omitempty"`
	ConfigDirs   map[string]string `json:"configdirs,omitempty"`
}

func DefaultSettings() Settings {
	userHome, err := os.UserHomeDir()
	if err != nil {
		userHome = os.Getenv("HOME")
	}

	return Settings{
		Integrations: []string{
			"ghostty",
			"zed",
			"system-theme",
			"wallpaper",
			"yazi",
			"eza",
			"nvim",
			"helix",
		},
		ConfigDirs: map[string]string{
			"ghostty": filepath.Join(userHome, ".config", "ghostty"),
			"zed":     zedConfigDir(userHome),
			"helix":   filepath.Join(userHome, ".config", "helix"),
			"yazi":    yaziConfigDir(userHome),
		},
	}
}

func LoadSettings(path string) (Settings, error) {
	defaults := DefaultSettings()

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaults, nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("read settings %w", err)
	}

	// This merges because unmarshal is overwriting in existing struct
	if err := json.Unmarshal(data, &defaults); err != nil {
		return Settings{}, fmt.Errorf("parse settings: %w", err)
	}

	return defaults, nil
}

func (s Settings) ConfigDirFor(integration string) string {
	if s.ConfigDirs == nil {
		return ""
	}

	path, ok := s.ConfigDirs[integration]
	if !ok {
		return ""
	}

	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	path = os.ExpandEnv(path)

	if path == "~" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home
		}
		return path
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}

	return path
}

func zedConfigDir(userHome string) string {
	platform := runtime.GOOS

	configHome, err := os.UserConfigDir()
	if err != nil {
		log.Fatal("User Config dir not set")
	}

	if platform == "windows" {
		return filepath.Join(configHome, "zed")
	}

	return filepath.Join(userHome, ".config", "zed")
}

func yaziConfigDir(userHome string) string {
	platform := runtime.GOOS

	if platform == "windows" {
		configHome, err := os.UserConfigDir()
		if err != nil {
			return ""
		}
		return filepath.Join(configHome, "yazi", "config")
	}

	return filepath.Join(userHome, ".config", "yazi")
}
