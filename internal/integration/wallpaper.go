package integration

import (
	"github.com/Nico-Mayer/themectl-cli/internal/theme"
	"github.com/Nico-Mayer/themectl-cli/internal/wallpaper"
)

type Wallpaper struct {
	SharedWallpapersDir string
	ThemesDir           string
}

func (Wallpaper) Name() string {
	return "wallpaper"
}

func (w Wallpaper) Apply(t theme.Resolved) error {
	manager := wallpaper.NewManager(w.ThemesDir, w.SharedWallpapersDir)
	return manager.ApplyRandom(t)
}
