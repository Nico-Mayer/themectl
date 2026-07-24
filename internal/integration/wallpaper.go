package integration

import (
	"sync"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/theme"
	"github.com/Nico-Mayer/themectl/internal/wallpaper"
)

type Wallpaper struct {
	manager wallpaper.Manager

	once       sync.Once
	candidates []string
}

func (*Wallpaper) Name() string {
	return "wallpaper"
}

func (w *Wallpaper) Check() error {
	return checkDirExists("themes dir", w.manager.ThemesDir)
}

func (w *Wallpaper) Supports(t theme.Resolved) bool {
	return len(w.list(t)) > 0
}

func (w *Wallpaper) Apply(t theme.Resolved) error {
	return w.manager.SetRandomFrom(w.list(t))
}

func (w *Wallpaper) list(t theme.Resolved) []string {
	w.once.Do(func() {
		w.candidates = w.manager.ListCandidates(t)
	})
	return w.candidates
}

func newWallpaper(cfg config.Config) Integration {
	return &Wallpaper{
		manager: wallpaper.NewManager(cfg.ThemesDir(), cfg.SharedWallpapersDir()),
	}
}
