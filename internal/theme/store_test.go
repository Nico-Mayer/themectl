package theme

import (
	"fmt"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/nico-mayer/themectl-cli/internal/testutil"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"catppuccin/family.toml": {Data: []byte(`
[defaults]
appearance = "dark"
[defaults.themes]
eza = "cat-eza"
`)},
		"catppuccin/mocha/variant.toml": {Data: []byte(`
wallpaper_sources = ["catppuccin/latte", "nature"]
[themes]
ghostty = "catppuccin-mocha"
`)},
		"catppuccin/frappe/variant.toml": {Data: []byte(`
wallpaper_sources = ["catppuccin/latte", "nature"]
[themes]
ghostty = "catppuccin-frappe"
`)},
		"catppuccin/latte/variant.toml": {Data: []byte(`
appearance = "light"
[themes]
ghostty = "catppuccin-latte"
`)},
		"catppuccin/mocha/nvim.lua": {Data: []byte("-- mocha")},
		"catppuccin/nvim.lua":       {Data: []byte("-- family default")},
	}
}

func TestStore_Resolve(t *testing.T) {
	s := NewStore(testFS())

	latte, err := s.Resolve("catppuccin/latte")
	testutil.NoErr(t, err)
	testutil.Equal(t, latte.Appearance, Light)
	testutil.Equal(t, latte.Themes["ghostty"], "catppuccin-latte")
	testutil.Equal(t, latte.Themes["eza"], "cat-eza")
	testutil.Diff(t, []string{"catppuccin/latte"}, latte.WallpaperSources)

	mocha, err := s.Resolve("catppuccin/mocha")
	testutil.NoErr(t, err)
	testutil.Diff(t, []string{"catppuccin/latte", "nature", "catppuccin/mocha"}, mocha.WallpaperSources)
}

func TestStore_Resolve_inheritsAppearanceFromFamily(t *testing.T) {
	s := NewStore(testFS())

	mocha, err := s.Resolve("catppuccin/mocha")
	testutil.NoErr(t, err)
	testutil.Equal(t, mocha.Appearance, Dark)

	latte, err := s.Resolve("catppuccin/latte")
	testutil.NoErr(t, err)
	testutil.Equal(t, latte.Appearance, Light)
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
		want       []Resolved
		appearance Appearance
	}{
		{"all dark themes", []Resolved{frappe, mocha}, Dark},
		{"all light themes", []Resolved{latte}, Light},
	}

	for _, tc := range tests {
		got, err := s.ListAllByAppearance(tc.appearance)
		testutil.NoErr(t, err)
		testutil.Diff(t, tc.want, got)
	}
}

func TestStore_AssetPath(t *testing.T) {
	s := NewStore(testFS())

	tests := []struct {
		name    string
		variant string
		asset   string
		want    string
		wantOk  bool
	}{
		{name: "variant has asset", variant: "mocha", asset: "nvim.lua", want: "catppuccin/mocha/nvim.lua", wantOk: true},
		{name: "variant inherits from family", variant: "latte", asset: "nvim.lua", want: "catppuccin/nvim.lua", wantOk: true},
		{name: "neither has asset", variant: "latte", asset: "eza.yml", want: "", wantOk: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := s.AssetPath("catppuccin", tt.variant, tt.asset)
			testutil.Equal(t, got, tt.want)
			testutil.Equal(t, ok, tt.wantOk)
		})
	}
}

func TestStore_Assets(t *testing.T) {
	fsys := fstest.MapFS{
		"catppuccin/family.toml":           {Data: []byte("[defaults]\nappearance='dark'\n")},
		"catppuccin/zed.json":              {Data: []byte(`{"from":"family"}`)},
		"catppuccin/nvim.lua":              {Data: []byte("-- family")},
		"catppuccin/mocha/variant.toml":    {Data: []byte("appearance='dark'\n")},
		"catppuccin/mocha/nvim.lua":        {Data: []byte("-- mocha")},
		"catppuccin/mocha/eza.yml":         {Data: []byte("mocha-only")},
		"catppuccin/mocha/wallpaper/a.png": {Data: []byte("img")},
	}
	s := NewStore(fsys)

	got, err := s.Assets("catppuccin", "mocha")
	testutil.NoErr(t, err)
	testutil.Diff(t, map[string]string{
		"zed.json": "catppuccin/zed.json",
		"nvim.lua": "catppuccin/mocha/nvim.lua",
		"eza.yml":  "catppuccin/mocha/eza.yml",
	}, got)
}

func TestStore_PickRandom(t *testing.T) {
	s := NewStore(testFS())

	t.Run("dark picks a dark theme", func(t *testing.T) {
		got, err := s.PickRandom(Dark)
		testutil.NoErr(t, err)
		testutil.Equal(t, got.Appearance, Dark)
	})

	t.Run("light picks the only light theme", func(t *testing.T) {
		got, err := s.PickRandom(Light)
		testutil.NoErr(t, err)
		testutil.Equal(t, got.ID(), "catppuccin/latte")
	})

	t.Run("no appearance picks any known theme", func(t *testing.T) {
		all, err := s.ListAll()
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

func benchFS(n int) fstest.MapFS {
	fsys := fstest.MapFS{}
	for i := range n {
		path := fmt.Sprintf("family%04d/family.toml", i)
		fsys[path] = &fstest.MapFile{Data: []byte("[defaults]\nappearance = \"dark\"\n")}
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
