package config

import "path/filepath"

type Config struct {
	Root     string
	Settings Settings
}

func (c Config) ThemesDir() string           { return filepath.Join(c.Root, "themes") }
func (c Config) CurrentDir() string          { return filepath.Join(c.Root, "current") }
func (c Config) CurrentFile() string         { return filepath.Join(c.Root, ".current") }
func (c Config) SharedWallpapersDir() string { return filepath.Join(c.Root, "shared_wallpapers") }

func Load(root string) (Config, error) {
	s, err := loadSettings(filepath.Join(root, "themectl.toml"))
	if err != nil {
		return Config{}, err
	}

	return Config{Root: root, Settings: s}, nil
}
