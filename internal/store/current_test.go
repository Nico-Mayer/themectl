package store

import (
	"path/filepath"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
)

func TestCurrent_roundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".current")

	if _, err := ReadCurrent(path); err == nil {
		t.Error("expected error when no current theme is set")
	}

	testutil.NoErr(t, WriteCurrent(path, "catppuccin/mocha"))

	got, err := ReadCurrent(path)
	testutil.NoErr(t, err)
	testutil.Equal(t, got, "catppuccin/mocha")
}
