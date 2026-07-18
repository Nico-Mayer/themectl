package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Nico-Mayer/themectl-cli/internal/testutil"
)

func TestSymlink_createsLink(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "source")
	target := filepath.Join(tmp, "foo", "bar", "link")
	testutil.NoErr(t, os.WriteFile(source, []byte("test"), 0o644))

	testutil.NoErr(t, symlink(source, target))

	info, err := os.Lstat(target)
	testutil.NoErr(t, err)
	testutil.Equal(t, info.Mode()&os.ModeSymlink, os.ModeSymlink)

	dest, err := os.Readlink(target)
	testutil.NoErr(t, err)
	testutil.Equal(t, dest, source)
}

func TestSymlink_overwritesStaleLink(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "source")
	stale := filepath.Join(tmp, "stale")
	target := filepath.Join(tmp, "foo", "bar", "link")
	testutil.NoErr(t, os.WriteFile(source, []byte("test"), 0o644))
	testutil.NoErr(t, os.WriteFile(stale, []byte("stale"), 0o644))
	testutil.NoErr(t, os.MkdirAll(filepath.Dir(target), 0o755))
	testutil.NoErr(t, os.Symlink(stale, target))

	testutil.NoErr(t, symlink(source, target))

	dest, err := os.Readlink(target)
	testutil.NoErr(t, err)
	testutil.Equal(t, dest, source)
}

func TestSymlink_refusesToOverwriteRealFile(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "source")
	target := filepath.Join(tmp, "foo", "bar", "link")
	testutil.NoErr(t, os.WriteFile(source, []byte("test"), 0o644))
	testutil.NoErr(t, os.MkdirAll(filepath.Dir(target), 0o755))
	testutil.NoErr(t, os.WriteFile(target, []byte("real file"), 0o644))

	if err := symlink(source, target); err == nil {
		t.Fatal("expected error when target is a real file")
	}
}

func TestSymlink_missingSourceFails(t *testing.T) {
	tmp := t.TempDir()

	err := symlink(filepath.Join(tmp, "source"), filepath.Join(tmp, "link"))
	if err == nil {
		t.Error("expected error when source does not exist")
	}
}
