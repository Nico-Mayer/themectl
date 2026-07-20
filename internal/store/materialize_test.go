package store

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/Nico-Mayer/themectl/internal/testutil"
)

func TestStore_Materialize(t *testing.T) {
	fsys := fstest.MapFS{
		"catppuccin/theme.toml":     {Data: []byte("[defaults]\nappearance='dark'\n")},
		"catppuccin/zed.json":       {Data: []byte(`{"from":"family"}`)},
		"catppuccin/mocha/nvim.lua": {Data: []byte("-- mocha")},
	}
	dest := filepath.Join(t.TempDir(), "current")
	testutil.NoErr(t, os.MkdirAll(dest, 0o755))
	testutil.NoErr(t, os.WriteFile(filepath.Join(dest, "stale.txt"), []byte("stale"), 0o644))

	testutil.NoErr(t, NewStore(fsys).Materialize("catppuccin/mocha", dest))

	zed, _ := os.ReadFile(filepath.Join(dest, "zed.json"))
	testutil.Equal(t, string(zed), `{"from":"family"}`)

	nvim, _ := os.ReadFile(filepath.Join(dest, "nvim.lua"))
	testutil.Equal(t, string(nvim), "-- mocha")

	if _, err := os.Stat(filepath.Join(dest, "stale.txt")); err == nil {
		t.Error("stale file survived; dest must be rebuilt from scratch")
	}
}

func TestStore_Assets(t *testing.T) {
	fsys := fstest.MapFS{
		"catppuccin/theme.toml":            {Data: []byte("[defaults]\nappearance='dark'\n[variants.mocha]\n[variants.latte]\n")},
		"catppuccin/zed.json":              {Data: []byte(`{"from":"family"}`)},
		"catppuccin/nvim.lua":              {Data: []byte("-- family")},
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

func TestStore_Assets_variantWithoutDirectory(t *testing.T) {
	fsys := fstest.MapFS{
		"catppuccin/theme.toml": {Data: []byte("[defaults]\nappearance='dark'\n[variants.latte]\n")},
		"catppuccin/nvim.lua":   {Data: []byte("-- family")},
	}
	s := NewStore(fsys)

	got, err := s.Assets("catppuccin", "latte")
	testutil.NoErr(t, err)
	testutil.Diff(t, map[string]string{
		"nvim.lua": "catppuccin/nvim.lua",
	}, got)
}
