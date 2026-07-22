package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Nico-Mayer/themectl/internal/testutil"
)

func TestPutGet(t *testing.T) {
	c := New(filepath.Join(t.TempDir(), "sub"))

	key := "https://github.com/x/y"
	testutil.NoErr(t, c.Put(key, "abc123"))

	got, ok := c.Get(key)
	testutil.Equal(t, got, "abc123")
	testutil.Equal(t, ok, true)
}

func TestFresh(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	key := "k"

	testutil.Equal(t, c.Fresh(key, time.Hour), false)
	testutil.NoErr(t, c.Put(key, "v"))
	testutil.Equal(t, c.Fresh(key, time.Hour), true)
	backdate(t, dir, 2*time.Hour)
	testutil.Equal(t, c.Fresh(key, time.Hour), false)
}

func TestTouch(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	testutil.NoErr(t, c.Put("k", "v"))
	backdate(t, dir, 2*time.Hour)

	testutil.NoErr(t, c.Touch("k"))
	testutil.Equal(t, c.Fresh("k", time.Hour), true)
	got, _ := c.Get("k")
	testutil.Equal(t, got, "v")
}

func TestClear(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	testutil.NoErr(t, c.Put("k", "v"))
	testutil.NoErr(t, c.Clear())

	got, ok := c.Get("k")
	testutil.Equal(t, got, "")
	testutil.Equal(t, ok, false)
}

func backdate(t *testing.T, dir string, age time.Duration) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	testutil.NoErr(t, err)
	testutil.Equal(t, len(files), 1)

	old := time.Now().Add(-age)
	testutil.NoErr(t, os.Chtimes(files[0], old, old))
}
