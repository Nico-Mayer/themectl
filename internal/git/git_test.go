package git

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
)

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	testutil.NoErr(t, os.MkdirAll(filepath.Join(dir, "themes"), 0o755))
	testutil.NoErr(t, os.MkdirAll(filepath.Join(dir, "other"), 0o755))
	testutil.NoErr(t, os.WriteFile(filepath.Join(dir, "themes", "a.json"), []byte("{}"), 0o644))
	testutil.NoErr(t, os.WriteFile(filepath.Join(dir, "other", "b.json"), []byte("{}"), 0o644))

	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@test.invalid"},
		{"config", "user.name", "test"},
		{"add", "."},
		{"commit", "-m", "init"},
	} {
		_, err := Run(dir, args...)
		testutil.NoErr(t, err)
	}
	return dir
}

func TestInstalled(t *testing.T) {
	testutil.NoErr(t, Installed())
}

func TestRun(t *testing.T) {
	dir := initRepo(t)

	out, err := Run(dir, "rev-parse", "HEAD")
	testutil.NoErr(t, err)
	testutil.Equal(t, len(out), 40) // a full commit hash, trimmed

	_, err = Run(dir, "not a git command")
	testutil.Equal(t, err != nil, true)
}

func TestRemoteHead(t *testing.T) {
	dir := initRepo(t)
	want, err := Run(dir, "rev-parse", "HEAD")
	testutil.NoErr(t, err)

	got, err := RemoteHead(dir)
	testutil.NoErr(t, err)
	testutil.Equal(t, got, want)
}

func TestCloneShallow(t *testing.T) {
	src := initRepo(t)
	dst := filepath.Join(t.TempDir(), "clone")

	testutil.NoErr(t, CloneShallow(src, dst))

	_, err := os.Stat(filepath.Join(dst, "themes", "a.json"))
	testutil.NoErr(t, err)
}

func TestSparseClone(t *testing.T) {
	src := initRepo(t)
	dst := filepath.Join(t.TempDir(), "clone")

	testutil.NoErr(t, SparseClone(src, dst, "themes"))

	_, err := os.Stat(filepath.Join(dst, "themes", "a.json"))
	testutil.NoErr(t, err)

	_, err = os.Stat(filepath.Join(dst, "other", "b.json"))
	testutil.Equal(t, errors.Is(err, os.ErrNotExist), true)
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct{ in, want string }{
		{"github.com/catppuccin/zed", "https://github.com/catppuccin/zed"},
		{"https://github.com/catppuccin/zed", "https://github.com/catppuccin/zed"},
		{"ssh://git@host/x", "ssh://git@host/x"},
	}
	for _, tt := range tests {
		testutil.Equal(t, NormalizeURL(tt.in), tt.want)
	}
}
