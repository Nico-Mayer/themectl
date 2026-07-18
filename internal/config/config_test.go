package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Nico-Mayer/themectl-cli/internal/testutil"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	testutil.NoErr(t, os.WriteFile(filepath.Join(dir, "themectl.toml"), []byte(`default-theme = "catppuccin-mocha"`), 0o644))

	c, err := Load(dir)
	testutil.NoErr(t, err)
	testutil.Equal(t, c.Settings.DefaultTheme, "catppuccin-mocha")
}
