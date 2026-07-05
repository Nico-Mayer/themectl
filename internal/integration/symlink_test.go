package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateSymlink(t *testing.T) {
	tmp := t.TempDir()

	source := filepath.Join(tmp, "source")
	target := filepath.Join(tmp, "foo", "bar", "link")

	_ = os.WriteFile(source, []byte("test"), 0644)

	err := symlink(source, target)
	if err != nil {
		t.Fatalf("got error but wanted none %v", err)
	}

	info, err := os.Lstat(target)
	if err != nil {
		t.Fatal("got error reading target info")
	}

	if info.Mode()&os.ModeSymlink != os.ModeSymlink {
		t.Error("target is not a symlink")
	}

	dest, err := os.Readlink(target)
	if err != nil {
		t.Fatal("failed to read target from symlink")
	}

	if dest != source {
		t.Errorf("target and source diverge, source: %q, target:%q", source, target)
	}
}

func TestOverwritesLinkWithDifferentTarget(t *testing.T) {
	tmp := t.TempDir()

	source := filepath.Join(tmp, "source")
	secondSource := filepath.Join(tmp, "source2")
	target := filepath.Join(tmp, "foo", "bar", "link")

	_ = os.WriteFile(source, []byte("test"), 0644)
	_ = os.WriteFile(secondSource, []byte("evil file"), 0644)

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal("failed to create target dir")
	}
	_ = os.Symlink(secondSource, target)

	err := symlink(source, target)
	if err != nil {
		t.Fatalf("got error but wanted none %v", err)
	}

	dest, err := os.Readlink(target)
	if err != nil {
		t.Fatal("failed to read target from symlink")
	}

	if dest != source {
		t.Errorf("want: %q, got: %q", source, dest)
	}
}

func TestErrorWhenOverwritingRealFile(t *testing.T) {
	tmp := t.TempDir()

	source := filepath.Join(tmp, "source")
	target := filepath.Join(tmp, "foo", "bar", "link")

	_ = os.WriteFile(source, []byte("test"), 0644)

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal("failed to create target dir")
	}
	_ = os.WriteFile(target, []byte("i am a real file on disc"), 0644)

	err := symlink(source, target)
	if err == nil {
		t.Fatalf("got no error but want one")
	}
}

func TestSourceNeedsToExist(t *testing.T) {
	tmp := t.TempDir()

	source := filepath.Join(tmp, "source")
	target := filepath.Join(tmp, "foo", "bar", "link")

	err := symlink(source, target)
	if err == nil {
		t.Errorf("symlinc created, but shoudent")
	}
}
