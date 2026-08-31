package store

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
)

func TestUpdate(t *testing.T) {
	upstream := gitFixture(t, validThemeToml)
	themesDir := t.TempDir()

	clean := filepath.Join(themesDir, "clean")
	gitClone(t, upstream, clean)

	dirty := filepath.Join(themesDir, "dirty")
	gitClone(t, upstream, dirty)
	testutil.NoErr(t, os.WriteFile(filepath.Join(dirty, "theme.toml"),
		[]byte(validThemeToml+"\n# local tweak\n"), 0o644))

	plain := filepath.Join(themesDir, "plain")
	testutil.NoErr(t, os.MkdirAll(plain, 0o755))

	advanceUpstream(t, upstream)

	asked := map[string]bool{}
	confirm := func(name string) (bool, error) {
		asked[name] = true
		return false, nil
	}

	results, err := Update(themesDir, confirm)
	testutil.NoErr(t, err)
	testutil.Equal(t, len(results), 3)

	if _, err := os.Stat(filepath.Join(clean, "extra.txt")); err != nil {
		t.Errorf("clean repo not updated: %v", err)
	}
	if !asked["dirty"] {
		t.Error("expected confirm to be asked for dirty repo")
	}
	if _, err := os.Stat(filepath.Join(dirty, "extra.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Error("declined dirty repo should not have been pulled")
	}
	if asked["plain"] {
		t.Error("non-repo should not be prompted")
	}
}

func TestUpdate_confirmProceeds(t *testing.T) {
	upstream := gitFixture(t, validThemeToml)
	themesDir := t.TempDir()

	dirty := filepath.Join(themesDir, "dirty")
	gitClone(t, upstream, dirty)
	testutil.NoErr(t, os.WriteFile(filepath.Join(dirty, "theme.toml"),
		[]byte(validThemeToml+"\n# local tweak\n"), 0o644))

	advanceUpstream(t, upstream)

	_, err := Update(themesDir, func(string) (bool, error) { return true, nil })
	testutil.NoErr(t, err)

	if _, err := os.Stat(filepath.Join(dirty, "extra.txt")); err != nil {
		t.Errorf("approved dirty repo not updated: %v", err)
	}
}

func TestUpdate_confirmError(t *testing.T) {
	upstream := gitFixture(t, validThemeToml)
	themesDir := t.TempDir()
	dirty := filepath.Join(themesDir, "dirty")
	gitClone(t, upstream, dirty)
	testutil.NoErr(t, os.WriteFile(filepath.Join(dirty, "theme.toml"),
		[]byte(validThemeToml+"\n# local tweak\n"), 0o644))

	confirmErr := errors.New("terminal unavailable")
	results, err := Update(themesDir, func(string) (bool, error) { return false, confirmErr })
	testutil.NoErr(t, err)
	if len(results) != 1 || results[0].Status != UpdateDeclined {
		t.Fatalf("results = %v, want one declined update", results)
	}
	if !errors.Is(results[0].Err, confirmErr) {
		t.Fatalf("result error = %v, want wrapped confirmation error", results[0].Err)
	}
}

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v (%s)", args, err, out)
	}
}

func gitClone(t *testing.T, src, dst string) {
	t.Helper()
	cmd := exec.Command("git", "clone", src, dst)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v (%s)", err, out)
	}
}

func advanceUpstream(t *testing.T, upstream string) {
	t.Helper()
	testutil.NoErr(t, os.WriteFile(filepath.Join(upstream, "extra.txt"), []byte("v2"), 0o644))
	gitCmd(t, upstream, "add", ".")
	gitCmd(t, upstream, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-m", "v2")
}
