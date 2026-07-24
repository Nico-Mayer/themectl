package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

func rioFixture(t *testing.T) Rio {
	t.Helper()
	dir := t.TempDir()

	cfgPath := filepath.Join(dir, "rio", "config.toml")
	testutil.NoErr(t, os.MkdirAll(filepath.Dir(cfgPath), 0o755))
	testutil.NoErr(t, os.WriteFile(cfgPath, []byte("theme = \"old\"\n[window]\nopacity = 0.9\n"), 0o644))

	src := filepath.Join(dir, "rio.toml")
	testutil.NoErr(t, os.WriteFile(src, []byte("[colors]\nbackground = \"#000000\"\n"), 0o644))

	return Rio{ConfigFile: cfgPath, SourceFile: src}
}

func (r Rio) linkPath() string {
	return filepath.Join(filepath.Dir(r.ConfigFile), "themes", "themectl.toml")
}

func TestRio_Apply_linksAndActivates(t *testing.T) {
	r := rioFixture(t)

	testutil.NoErr(t, r.Apply(theme.Resolved{}))

	fi, err := os.Lstat(r.linkPath())
	testutil.NoErr(t, err)
	testutil.Equal(t, fi.Mode()&os.ModeSymlink != 0, true)

	dest, err := os.Readlink(r.linkPath())
	testutil.NoErr(t, err)
	testutil.Equal(t, dest, r.SourceFile)

	out, err := os.ReadFile(r.ConfigFile)
	testutil.NoErr(t, err)
	if !strings.Contains(string(out), `theme = "themectl"`) {
		t.Errorf("config not rewritten: %q", out)
	}
	if !strings.Contains(string(out), "[window]\nopacity = 0.9") {
		t.Errorf("sibling config clobbered: %q", out)
	}
}

func TestRio_Apply_missingThemeLineFailsWithoutSideEffects(t *testing.T) {
	r := rioFixture(t)
	before := []byte("[window]\nopacity = 0.9\n")
	testutil.NoErr(t, os.WriteFile(r.ConfigFile, before, 0o644))

	if err := r.Apply(theme.Resolved{}); err == nil {
		t.Fatal("expected error for config without theme line")
	}

	if _, err := os.Lstat(r.linkPath()); err == nil {
		t.Error("symlink created despite config error")
	}
	out, err := os.ReadFile(r.ConfigFile)
	testutil.NoErr(t, err)
	testutil.Equal(t, string(out), string(before))
}

func TestRio_Supports_requiresAsset(t *testing.T) {
	r := rioFixture(t)

	testutil.Equal(t, r.Supports(theme.Resolved{}), true)

	testutil.NoErr(t, os.Remove(r.SourceFile))
	testutil.Equal(t, r.Supports(theme.Resolved{}), false)
}

func TestRio_Apply_isIdempotent(t *testing.T) {
	r := rioFixture(t)

	testutil.NoErr(t, r.Apply(theme.Resolved{}))
	testutil.NoErr(t, r.Apply(theme.Resolved{}))

	dest, err := os.Readlink(r.linkPath())
	testutil.NoErr(t, err)
	testutil.Equal(t, dest, r.SourceFile)
}
