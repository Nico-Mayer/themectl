package store

import (
	"fmt"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/Nico-Mayer/themectl/internal/testutil"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"catppuccin/theme.toml": {Data: []byte(`
[defaults]
appearance = "dark"

[defaults.ghostty]
theme = "catppuccin-default"

[defaults.zed]
theme = "Catppuccin Mocha"
icon_theme = "Catppuccin Mocha"
extensions = ["github.com/catppuccin/zed"]

[variants.mocha]
wallpaper_sources = ["catppuccin/latte", "nature"]
[variants.mocha.ghostty]
theme = "catppuccin-mocha"

[variants.frappe]
wallpaper_sources = ["catppuccin/latte", "nature"]
[variants.frappe.ghostty]
theme = "catppuccin-frappe"

[variants.latte]
appearance = "light"
[variants.latte.ghostty]
theme = "catppuccin-latte"
[variants.latte.zed]
theme = "Catppuccin Latte"
`)},
		"catppuccin/mocha/nvim.lua": {Data: []byte("-- mocha")},
		"catppuccin/nvim.lua":       {Data: []byte("-- family default")},
	}
}

func TestStore_Resolve(t *testing.T) {
	s := NewStore(testFS())

	latte, err := s.Resolve("catppuccin/latte")
	testutil.NoErr(t, err)
	testutil.Equal(t, latte.Appearance, theme.Light)
	testutil.Equal(t, latte.Ghostty.Theme, "catppuccin-latte")
	testutil.Diff(t, []string{"catppuccin/latte"}, latte.WallpaperSources)

	mocha, err := s.Resolve("catppuccin/mocha")
	testutil.NoErr(t, err)
	testutil.Diff(t, []string{"catppuccin/latte", "nature", "catppuccin/mocha"}, mocha.WallpaperSources)
}

func TestStore_Resolve_inheritsAppearanceFromFamily(t *testing.T) {
	s := NewStore(testFS())

	mocha, err := s.Resolve("catppuccin/mocha")
	testutil.NoErr(t, err)
	testutil.Equal(t, mocha.Appearance, theme.Dark)

	latte, err := s.Resolve("catppuccin/latte")
	testutil.NoErr(t, err)
	testutil.Equal(t, latte.Appearance, theme.Light)
}

func TestStore_Resolve_badID(t *testing.T) {
	s := NewStore(testFS())
	if _, err := s.Resolve("nofamilyslash"); err == nil {
		t.Error("expected error for id without a slash")
	}
}

func TestStore_List(t *testing.T) {
	s := NewStore(testFS())
	got, err := s.listVariants("catppuccin")
	testutil.NoErr(t, err)
	testutil.Diff(t, []string{"frappe", "latte", "mocha"}, got)
}

func TestStore_ListAllByAppearance(t *testing.T) {
	s := NewStore(testFS())
	mocha, err := s.Resolve("catppuccin/mocha")
	testutil.NoErr(t, err)
	frappe, err := s.Resolve("catppuccin/frappe")
	testutil.NoErr(t, err)
	latte, err := s.Resolve("catppuccin/latte")
	testutil.NoErr(t, err)

	tests := []struct {
		name       string
		want       []theme.Resolved
		appearance theme.Appearance
	}{
		{"all dark themes", []theme.Resolved{frappe, mocha}, theme.Dark},
		{"all light themes", []theme.Resolved{latte}, theme.Light},
	}

	for _, tc := range tests {
		got, err := s.List(tc.appearance)
		testutil.NoErr(t, err)
		testutil.Diff(t, tc.want, got)
	}
}

func TestStore_PickRandom(t *testing.T) {
	s := NewStore(testFS())

	t.Run("dark picks a dark theme", func(t *testing.T) {
		got, err := s.PickRandom(theme.Dark)
		testutil.NoErr(t, err)
		testutil.Equal(t, got.Appearance, theme.Dark)
	})

	t.Run("light picks the only light theme", func(t *testing.T) {
		got, err := s.PickRandom(theme.Light)
		testutil.NoErr(t, err)
		testutil.Equal(t, got.ID(), "catppuccin/latte")
	})

	t.Run("no appearance picks any known theme", func(t *testing.T) {
		all, err := s.IDs()
		testutil.NoErr(t, err)
		for range 20 {
			got, err := s.PickRandom("")
			testutil.NoErr(t, err)
			if !slices.Contains(all, got.ID()) {
				t.Fatalf("picked %q, not in %v", got.ID(), all)
			}
		}
	})
}

func TestStore_Resolve_variantInheritsIntegrationFields(t *testing.T) {
	s := NewStore(testFS())

	latte, err := s.Resolve("catppuccin/latte")
	testutil.NoErr(t, err)
	testutil.Equal(t, latte.Zed.Theme, "Catppuccin Latte")
	testutil.Equal(t, latte.Zed.IconTheme, "Catppuccin Mocha")
	testutil.Diff(t, []string{"github.com/catppuccin/zed"}, latte.Zed.Extensions)
}

func TestStore_Resolve_variantWithoutDirectory(t *testing.T) {
	s := NewStore(testFS())
	got, err := s.Resolve("catppuccin/frappe")
	testutil.NoErr(t, err)
	testutil.Equal(t, got.ID(), "catppuccin/frappe")
}

func benchFS(n int) fstest.MapFS {
	fsys := fstest.MapFS{}
	for i := range n {
		path := fmt.Sprintf("family%04d/theme.toml", i)
		fsys[path] = &fstest.MapFile{Data: []byte("[defaults]\nappearance = \"dark\"\n[variants.a]\n")}
	}
	return fsys
}

func BenchmarkStore_allFamilies(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("families=%d", n), func(b *testing.B) {
			s := NewStore(benchFS(n))
			b.ReportAllocs()
			for b.Loop() {
				if _, err := s.allFamilies(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
