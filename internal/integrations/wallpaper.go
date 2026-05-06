package integrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/fs"
	"github.com/nico-mayer/themectl-cli/internal/model"
	"github.com/nico-mayer/themectl-cli/internal/random"
	"github.com/reujab/wallpaper"
)

type Wallpaper struct{}

func init() {
	Register(Wallpaper{})
}

func (Wallpaper) Name() string {
	return "wallpaper"
}

func (i Wallpaper) Apply(themeInfo model.ThemeInfo) error {
	cfg, _ := config.Get()
	logger := integrationLogger(i)

	sourceDirs := collectWallpaperSourceDirs(cfg, themeInfo)
	candidates := collectWallpaperCandidates(sourceDirs)
	if len(candidates) == 0 {
		logger.Warn("no wallpaper files found skipping...", "theme", themeInfo.Name)
		return nil
	}

	current, _ := wallpaper.Get()
	selected := pickWallpaper(candidates, current)

	logger.Debug("setting wallpaper", "selected", selected, "candidates", len(candidates))
	if err := wallpaper.SetFromFile(selected); err != nil {
		return fmt.Errorf("set wallpaper from %q: %w", selected, err)
	}

	logger.Info("set", "wallpaper", selected)
	return nil
}

func collectWallpaperSourceDirs(cfg config.Config, themeInfo model.ThemeInfo) []string {
	dirs := []string{
		filepath.Join(cfg.Paths.ThemesDir, themeInfo.Name, "wallpaper"),
	}

	for _, source := range themeInfo.WallpaperSources {
		if themePath := filepath.Join(cfg.Paths.ThemesDir, source, "wallpaper"); fs.Exists(themePath) {
			dirs = append(dirs, themePath)
		}
		if sourcePath := filepath.Join(cfg.Paths.WallpaperSourcesDir, source); fs.Exists(sourcePath) {
			dirs = append(dirs, sourcePath)
		}
	}

	return dirs
}

func collectWallpaperCandidates(sourceDirs []string) []string {
	supported := map[string]struct{}{
		".png": {}, ".jpeg": {}, ".jpg": {}, ".heic": {},
	}

	var candidates []string
	for _, dir := range sourceDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if _, ok := supported[ext]; ok {
				candidates = append(candidates, filepath.Join(dir, entry.Name()))
			}
		}
	}
	return candidates
}

func pickWallpaper(candidates []string, current string) string {
	if len(candidates) == 1 {
		return candidates[0]
	}
	selected := random.Element(candidates)
	for selected == current {
		selected = random.Element(candidates)
	}
	return selected
}
