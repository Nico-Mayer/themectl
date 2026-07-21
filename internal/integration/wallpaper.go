package integration

import (
	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/theme"
	"github.com/Nico-Mayer/themectl/internal/wallpaper"
)

type Wallpaper struct {
	SharedWallpapersDir string
	ThemesDir           string
}

func (Wallpaper) Name() string {
	return "wallpaper"
}

func (w Wallpaper) Check() error {
	return checkFileExists("themes dir", w.ThemesDir)
}

func (w Wallpaper) Apply(t theme.Resolved) error {
	manager := wallpaper.NewManager(w.ThemesDir, w.SharedWallpapersDir)
	return manager.ApplyRandom(t)
}

func newWallpaper(cfg config.Config) Integration {
	return Wallpaper{
		ThemesDir:           cfg.ThemesDir(),
		SharedWallpapersDir: cfg.SharedWallpapersDir(),
	}
}
