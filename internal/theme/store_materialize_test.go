package theme

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
