package wallpaper

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/nico-mayer/themectl-cli/internal/theme"
)

func TestCollectingSourceDirs(t *testing.T) {
	themesDir := t.TempDir()
	sharedWallpapersDir := t.TempDir()

	mkDirAll(t, themesDir, "nord/dark/wallpaper", "catppuccin/mocha/wallpaper")
	mkDirAll(t, sharedWallpapersDir, "gen_dark")

	res := func(sources ...string) theme.Resolved {
		return theme.Resolved{
			Family:           "catppuccin",
			Variant:          "mocha",
			WallpaperSources: sources,
		}
	}

	tests := []struct {
		name  string
		theme theme.Resolved
		want  []string
	}{
		{
			name:  "no sources: should return themes own wallpaper path",
			theme: res(),
			want:  []string{filepath.Join(themesDir, "catppuccin", "mocha", "wallpaper")},
		},
		{
			name:  "source resolving other theme",
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
				filepath.Join(sharedWallpapersDir, "gen_dark"),
			},
		},
		{
			name:  "unknown source contributes nothing",
			theme: res("does-not-exist"),
			want:  []string{filepath.Join(themesDir, "catppuccin/mocha/wallpaper")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			slices.Sort(tc.want)
			manager := NewManager(themesDir, sharedWallpapersDir)
			got := manager.collectSourceDirs(tc.theme)

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got: %v\n want: %v", got, tc.want)
			}
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

	got := collectCandidates([]string{
		dir,
		filepath.Join(dir, "missing"),
	})

	want := []string{jpegUpper, png, heic, jpg}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("\n got: %v \n want: %v", got, want)
	}
}

func mkDirAll(t *testing.T, root string, rel ...string) {
	t.Helper()
	for _, p := range rel {
		if err := os.MkdirAll(filepath.Join(root, p), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func touch(t *testing.T, root, rel string) string {
	t.Helper()
	p := filepath.Join(root, rel)

	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(p, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestPickWallpaper(t *testing.T) {
	t.Run("single candidate should return even if set already", func(t *testing.T) {
		candidates := []string{"/w/a.png"}
		got := pickWallpaper(candidates, "/w/a.png")
		want := candidates[0]
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("no candidate should return current", func(t *testing.T) {
		candidates := []string{}
		current := "/w/a.png"
		got := pickWallpaper(candidates, current)
		want := current
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
