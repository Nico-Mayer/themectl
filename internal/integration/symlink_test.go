package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
	"github.com/Nico-Mayer/themectl/internal/theme"
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

func TestSymlinkIntegration_supportsRequiresSourceFile(t *testing.T) {
	tmp := t.TempDir()
	s := SymlinkIntegration{
		IntegrationName: "nvim",
		SourceFile:      filepath.Join(tmp, "nvim.lua"),
	}

	if s.Supports(theme.Resolved{}) {
		t.Error("theme without asset file must not be supported")
	}

	testutil.NoErr(t, os.WriteFile(s.SourceFile, []byte("theme"), 0o644))
	if !s.Supports(theme.Resolved{}) {
		t.Error("theme with asset file must be supported")
	}
}

func TestSymlinkIntegration_checkRequiresBinary(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "nvim")
	testutil.NoErr(t, os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755))
	t.Setenv("PATH", filepath.Dir(bin))

	s := SymlinkIntegration{IntegrationName: "nvim", Binary: "nvim"}
	testutil.NoErr(t, s.Check())

	s.Binary = "definitely-not-installed"
	if err := s.Check(); err == nil {
		t.Error("expected error when binary is not on PATH")
	}
}
