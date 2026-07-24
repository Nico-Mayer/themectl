package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/Nico-Mayer/themectl/internal/testutil"
)

type fakeFetcher struct {
	responses map[string][]byte
	errs      map[string]error
	requested []string
}

func (f *fakeFetcher) Fetch(_ context.Context, url string) ([]byte, error) {
	f.requested = append(f.requested, url)
	if err, ok := f.errs[url]; ok {
		return nil, err
	}
	return f.responses[url], nil
}

func TestStore_Materialize(t *testing.T) {
	fsys := fstest.MapFS{
		"catppuccin/theme.toml": {Data: []byte(`
			[defaults]
			appearance = 'dark'

			[variants.mocha]
			`)},
		"catppuccin/eza.yml":        {Data: []byte(`-- local`)},
		"catppuccin/mocha/nvim.lua": {Data: []byte("-- mocha")},
	}
	fake := &fakeFetcher{}
	dest := filepath.Join(t.TempDir(), "current")
	testutil.NoErr(t, os.MkdirAll(dest, 0o755))
	testutil.NoErr(t, os.WriteFile(filepath.Join(dest, "stale.txt"), []byte("stale"), 0o644))

	testutil.NoErr(t, NewStore(fsys, fake).Materialize(context.Background(), "catppuccin/mocha", dest))

	eza, _ := os.ReadFile(filepath.Join(dest, "eza.yml"))
	testutil.Equal(t, string(eza), `-- local`)

	nvim, _ := os.ReadFile(filepath.Join(dest, "nvim.lua"))
	testutil.Equal(t, string(nvim), "-- mocha")
	testutil.Equal(t, len(fake.requested), 0)

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
	s := NewStore(fsys, nil)

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
	s := NewStore(fsys, nil)

	got, err := s.Assets("catppuccin", "latte")
	testutil.NoErr(t, err)
	testutil.Diff(t, map[string]string{
		"nvim.lua": "catppuccin/nvim.lua",
	}, got)
}

func TestMaterialize_RemoteAssetsLandsOnDisk(t *testing.T) {
	fsys := fstest.MapFS{
		"catppuccin/theme.toml": {Data: []byte(`
	[defaults]
	appearance = 'dark'

	[defaults.nvim]
	url = 'https://example.com/mocha.lua'

	[variants.mocha]
	`)},
	}

	fake := &fakeFetcher{
		responses: map[string][]byte{
			"https://example.com/mocha.lua": []byte("-- remote"),
		},
	}

	dest := filepath.Join(t.TempDir(), "current")
	testutil.NoErr(t, NewStore(fsys, fake).Materialize(context.Background(), "catppuccin/mocha", dest))
	got, err := os.ReadFile(filepath.Join(dest, "nvim.lua"))
	testutil.NoErr(t, err)
	testutil.Equal(t, string(got), "-- remote")
	testutil.Diff(t, []string{"https://example.com/mocha.lua"}, fake.requested)
}

func TestMaterialize_RemoteBeatsLocal(t *testing.T) {
	fsys := fstest.MapFS{
		"catppuccin/theme.toml": {Data: []byte(`
	[defaults]
	appearance = 'dark'

	[defaults.nvim]
	url = 'https://example.com/mocha.lua'

	[variants.mocha]
	`)},
		"catppuccin/mocha/nvim.lua": {Data: []byte("-- local")},
	}

	fake := &fakeFetcher{
		responses: map[string][]byte{
			"https://example.com/mocha.lua": []byte("-- remote"),
		},
	}
	dest := filepath.Join(t.TempDir(), "current")

	testutil.NoErr(t, NewStore(fsys, fake).Materialize(context.Background(), "catppuccin/mocha", dest))
	got, err := os.ReadFile(filepath.Join(dest, "nvim.lua"))
	testutil.NoErr(t, err)
	testutil.Equal(t, string(got), "-- remote")
}

func TestMaterialize_RemoteVariantOverFamily(t *testing.T) {
	fsys := fstest.MapFS{
		"catppuccin/theme.toml": {Data: []byte(`
	[defaults]
	appearance = 'dark'

	[defaults.nvim]
	url = 'https://example.com/catppuccin.lua'

	[variants.mocha.nvim]
	url = "https://example.com/mocha.lua"
	`)},
		"catppuccin/mocha/nvim.lua": {Data: []byte("-- local")},
	}

	fake := &fakeFetcher{
		responses: map[string][]byte{
			"https://example.com/catppuccin.lua": []byte("-- family"),
			"https://example.com/mocha.lua":      []byte("-- variant"),
		},
	}
	dest := filepath.Join(t.TempDir(), "current")

	testutil.NoErr(t, NewStore(fsys, fake).Materialize(context.Background(), "catppuccin/mocha", dest))
	got, err := os.ReadFile(filepath.Join(dest, "nvim.lua"))
	testutil.NoErr(t, err)
	testutil.Equal(t, string(got), "-- variant")
	testutil.Diff(t, []string{"https://example.com/mocha.lua"}, fake.requested)
}

func TestMaterialize_FailedFetchSkips(t *testing.T) {
	fsys := fstest.MapFS{
		"catppuccin/theme.toml": {Data: []byte(`
	[defaults]
	appearance = 'dark'

	[defaults.nvim]
	url = 'https://example.com/catppuccin.lua'

	[variants.mocha.nvim]
	url = "https://example.com/mocha.lua"
	[variants.mocha.eza]
	url = "https://example.com/eza.yml"	`)},
		"catppuccin/mocha/nvim.lua": {Data: []byte("-- local")},
	}

	fake := &fakeFetcher{
		responses: map[string][]byte{
			"https://example.com/eza.yml": []byte("-- eza remote"),
		},
		errs: map[string]error{
			"https://example.com/mocha.lua": errors.New("faild to fetch"),
		},
	}
	dest := filepath.Join(t.TempDir(), "current")

	testutil.NoErr(t, NewStore(fsys, fake).Materialize(context.Background(), "catppuccin/mocha", dest))
	got, err := os.ReadFile(filepath.Join(dest, "eza.yml"))
	testutil.NoErr(t, err)
	testutil.Equal(t, string(got), "-- eza remote")
}
