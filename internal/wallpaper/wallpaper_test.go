package wallpaper

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Nico-Mayer/themectl-cli/internal/testutil"
	"github.com/Nico-Mayer/themectl-cli/internal/theme"
)

func TestCollectSourceDirs(t *testing.T) {
	themesDir := t.TempDir()
	sharedDir := t.TempDir()
	mkDirAll(t, themesDir, "nord/dark/wallpaper", "catppuccin/mocha/wallpaper")
	mkDirAll(t, sharedDir, "gen_dark")

	res := func(sources ...string) theme.Resolved {
		return theme.Resolved{
			Family:           "catppuccin",
			Variant:          "mocha",
			WallpaperSources: append(sources, "catppuccin/mocha"),
		}
	}

	tests := []struct {
		name  string
		theme theme.Resolved
		want  []string
	}{
		{
			name:  "no sources returns the theme's own wallpaper dir",
			theme: res(),
			want:  []string{filepath.Join(themesDir, "catppuccin", "mocha", "wallpaper")},
		},
		{
			name:  "source resolving another theme",
			theme: res("nord/dark"),
			want: []string{
				filepath.Join(themesDir, "catppuccin", "mocha", "wallpaper"),
				filepath.Join(themesDir, "nord", "dark", "wallpaper"),
			},
		},
		{
			name:  "source resolving shared wallpapers",
			theme: res("gen_dark"),
			want: []string{
				filepath.Join(themesDir, "catppuccin", "mocha", "wallpaper"),
				filepath.Join(sharedDir, "gen_dark"),
			},
		},
		{
			name:  "unknown source contributes nothing",
			theme: res("does-not-exist"),
			want:  []string{filepath.Join(themesDir, "catppuccin", "mocha", "wallpaper")},
		},
		{
			name:  "duplicate source resolving to the same dir is collapsed",
			theme: res("catppuccin/mocha"),
			want:  []string{filepath.Join(themesDir, "catppuccin", "mocha", "wallpaper")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slices.Sort(tt.want)
			got := NewManager(themesDir, sharedDir).collectSourceDirs(tt.theme)
			testutil.Diff(t, tt.want, got)
		})
	}
}

func TestCollectCandidates(t *testing.T) {
	dir := t.TempDir()
	jpegUpper := touch(t, dir, "a.JPEG")
	png := touch(t, dir, "b.png")
	heic := touch(t, dir, "c.heic")
	jpg := touch(t, dir, "d.jpg")
	touch(t, dir, "note.txt")
	touch(t, dir, "nested/deep.png")

	got := collectCandidates([]string{dir, filepath.Join(dir, "missing")})

	testutil.Diff(t, []string{jpegUpper, png, heic, jpg}, got)
}

func TestPickWallpaper(t *testing.T) {
	t.Run("single candidate returned even if already current", func(t *testing.T) {
		got := pickWallpaper([]string{"/w/a.png"}, "/w/a.png")
		testutil.Equal(t, got, "/w/a.png")
	})

	t.Run("no candidates returns current", func(t *testing.T) {
		got := pickWallpaper(nil, "/w/a.png")
		testutil.Equal(t, got, "/w/a.png")
	})
}

func mkDirAll(t *testing.T, root string, rel ...string) {
	t.Helper()
	for _, p := range rel {
		testutil.NoErr(t, os.MkdirAll(filepath.Join(root, p), 0o755))
	}
}

func touch(t *testing.T, root, rel string) string {
	t.Helper()
	p := filepath.Join(root, rel)
	testutil.NoErr(t, os.MkdirAll(filepath.Dir(p), 0o755))
	testutil.NoErr(t, os.WriteFile(p, nil, 0o644))
	return p
}
