package theme

import (
	"reflect"
	"slices"
	"testing"
	"testing/fstest"
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
appearance = "dark"
wallpaper_sources = ["catppuccin/latte", "nature"]
[themes]
ghostty = "catppuccin-mocha"
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

	got, err := s.Resolve("catppuccin/latte")
	if err != nil {
		t.Fatal(err)
	}
	if got.Appearance != Light {
		t.Errorf("appearance = %q, want light", got.Appearance)
	}
	if got.Themes["ghostty"] != "catppuccin-latte" {
		t.Errorf("ghostty = %q", got.Themes["ghostty"])
	}
	if got.Themes["eza"] != "cat-eza" {
		t.Errorf("eza = %q, want inherited cat-eza", got.Themes["eza"])
	}
	if len(got.WallpaperSources) != 0 {
		t.Errorf("latte wallpaper sources = %v, want none", got.WallpaperSources)
	}

	mocha, err := s.Resolve("catppuccin/mocha")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(mocha.WallpaperSources, []string{"catppuccin/latte", "nature"}) {
		t.Errorf("mocha wallpaper sources = %v", mocha.WallpaperSources)
	}
}

func TestStore_ListVariants(t *testing.T) {
	s := NewStore(testFS())
	got, err := s.List("catppuccin")
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 {
		t.Fatalf("want 2 variants, got %v", got)
	}
}

func TestStore_AssetPath_variantShadowsFamily(t *testing.T) {
	s := NewStore(testFS())

	tests := []struct {
		name    string
		family  string
		variant string
		asset   string
		want    string
		wantOk  bool
	}{
		{name: "variant has asset", family: "catppuccin", variant: "mocha", asset: "nvim.lua", want: "catppuccin/mocha/nvim.lua", wantOk: true},
		{name: "variant inharits from family", family: "catppuccin", variant: "latte", asset: "nvim.lua", want: "catppuccin/nvim.lua", wantOk: true},
		{name: "neither has asset", family: "catppuccin", variant: "latte", asset: "eza.yml", want: "", wantOk: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, ok := s.AssetPath(tt.family, tt.variant, tt.asset)
			if ok != tt.wantOk || p != tt.want {
				t.Errorf("got: %v want: %v (got ok=%v, want ok;%v)", p, tt.want, ok, tt.wantOk)
			}
		})
	}
}

func TestStore_Assets_overlay(t *testing.T) {
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
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"zed.json": "catppuccin/zed.json",
		"nvim.lua": "catppuccin/mocha/nvim.lua",
		"eza.yml":  "catppuccin/mocha/eza.yml",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("assets = %v\nwant %v", got, want)
	}

	bad := []string{"wallpaper", "a.png", "variant.toml", "family.toml"}

	for _, b := range bad {
		_, ok := got[b]
		if ok {
			t.Errorf("%q must not be an asset", b)
		}
	}
}

func TestStore_Resolve_badID(t *testing.T) {
	s := NewStore(testFS())
	if _, err := s.Resolve("nofamilyslash"); err == nil {
		t.Error("expected error for id without a slash")
	}
}

func TestStore_PickRandom(t *testing.T) {
	s := NewStore(testFS())

	t.Run("pick the only dark theme", func(t *testing.T) {
		got, err := s.PickRandom(new(Dark))
		if err != nil {
			t.Fatal(err)
		}
		if got.ID() != "catppuccin/mocha" {
			t.Errorf("dark pick = %q, want catppuccin/mocha", got.ID())
		}
	})

	t.Run("pick the only light theme", func(t *testing.T) {
		got, err := s.PickRandom(new(Light))
		if err != nil {
			t.Fatal(err)
		}
		if got.ID() != "catppuccin/latte" {
			t.Errorf("light pick = %q, want catppuccin/latte", got.ID())
		}
	})

	t.Run("no appearance picks any known theme", func(t *testing.T) {
		all, err := s.ListAll()
		if err != nil {
			t.Fatal(err)
		}
		for range 20 {
			got, err := s.PickRandom(nil)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Contains(all, got.ID()) {
				t.Fatalf("picked %q, not in %v", got.ID(), all)
			}
		}
	})
}
