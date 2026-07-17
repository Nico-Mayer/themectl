package wallpaper

import (
	"log/slog"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/nico-mayer/themectl-cli/internal/theme"
	rjWall "github.com/reujab/wallpaper"
)

const (
	wallSufix = "wallpaper"
)

var allowedFileTypes = []string{".jpeg", ".jpg", ".png", ".heic"}

type Manager struct {
	ThemsDir            string
	SharedWallpapersDir string
}

func NewManager(themesDir, sharedWallpapersDir string) Manager {
	return Manager{
		ThemsDir:            themesDir,
		SharedWallpapersDir: sharedWallpapersDir,
	}
}

func (m Manager) ApplyRandom(t theme.Resolved) error {
	sources := m.collectSourceDirs(t)
	candidates := collectCandidates(sources)

	if len(candidates) == 0 {
		return nil
	}

	current, err := rjWall.Get()
	if err != nil {
		slog.Debug("failed to get current wallpaper", "err", err)
		current = ""
	}

	selected := pickWallpaper(candidates, current)

	err = rjWall.SetFromFile(selected)
	if err != nil {
		return err
	}

	return nil
}

func (m Manager) Current() (string, error) {
	return rjWall.Get()
}

func (m Manager) collectSourceDirs(theme theme.Resolved) []string {
	var sources = []string{}
	wallSources := append(theme.WallpaperSources, theme.ID())

	for _, s := range wallSources {
		sourcesPath := filepath.Join(m.SharedWallpapersDir, s)
		if exists(sourcesPath) {
			sources = append(sources, sourcesPath)
		}

		themesPath := filepath.Join(m.ThemsDir, s, wallSufix)
		if exists(themesPath) {
			sources = append(sources, themesPath)
		}
	}
	slices.Sort(sources)

	return sources
}

func collectCandidates(sources []string) []string {
	var candidates []string
	for _, s := range sources {
		dir, err := os.ReadDir(s)
		if err != nil {
			slog.Debug("failed to read", "dir", s)
			continue
		}

		for _, entry := range dir {
			if entry.IsDir() {
				continue
			}

			fileExtension := strings.ToLower(filepath.Ext(entry.Name()))

			if slices.Contains(allowedFileTypes, fileExtension) {
				candidates = append(candidates, filepath.Join(s, entry.Name()))
			}
		}
	}

	return candidates
}

func exists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func pickWallpaper(candidates []string, current string) string {
	switch len(candidates) {
	case 0:
		return current
	case 1:
		return candidates[0]
	default:
		picked := candidates[rand.IntN(len(candidates))]
		for picked == current {
			picked = candidates[rand.IntN(len(candidates))]
		}
		return picked
	}
}
