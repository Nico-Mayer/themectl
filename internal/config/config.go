package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const settingsFileName = "themectl.toml"

type Config struct {
	Root      string
	Settings  Settings
	CacheRoot string
}

func (c Config) ThemesDir() string           { return filepath.Join(c.Root, "themes") }
func (c Config) CurrentDir() string          { return filepath.Join(c.Root, "current") }
func (c Config) CurrentFile() string         { return filepath.Join(c.Root, ".current") }
func (c Config) SharedWallpapersDir() string { return filepath.Join(c.Root, "shared_wallpapers") }
func (c Config) SettingsFile() string        { return filepath.Join(c.Root, settingsFileName) }
func (c Config) CacheDir() string            { return filepath.Join(c.CacheRoot, "themectl") }

func Load(root string) (Config, error) {
	s, err := loadSettings(filepath.Join(root, settingsFileName))
	if err != nil {
		return Config{}, err
	}

	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return Config{}, fmt.Errorf("resolve cache dir: %w", err)
	}

	return Config{Root: root, Settings: s, CacheRoot: cacheRoot}, nil
}
