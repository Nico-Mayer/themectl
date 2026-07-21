package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/testutil"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	testutil.NoErr(t, os.WriteFile(filepath.Join(dir, "themectl.toml"), []byte(`integrations = ["ghostty"]`), 0o644))

	c, err := config.Load(dir)
	testutil.NoErr(t, err)
	testutil.Equal(t, len(c.Settings.Integrations), 1)
	testutil.Equal(t, c.Settings.Integrations[0], "ghostty")
}

func TestCacheDir(t *testing.T) {
	cacheRoot := t.TempDir()
	c := config.Config{CacheRoot: cacheRoot}
	testutil.Equal(t, c.CacheDir(), filepath.Join(cacheRoot, "themectl"))
}
