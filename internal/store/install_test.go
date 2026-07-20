package store

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
)

const validThemeToml = `
[defaults]
appearance = "dark"

[variants.mocha]
`

func gitFixture(t *testing.T, themeToml string) string {
	t.Helper()
	dir := t.TempDir()
	testutil.NoErr(t, os.WriteFile(filepath.Join(dir, "theme.toml"), []byte(themeToml), 0o644))
	for _, args := range [][]string{
		{"init"}, {"add", "."},
		{"-c", "user.email=t@t", "-c", "user.name=t", "commit", "-m", "init"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}
	return dir
}

func TestInstall(t *testing.T) {
	repo := gitFixture(t, validThemeToml)
	themesDir := t.TempDir()

	name, err := Install(themesDir, repo, "catppuccin", false)
	testutil.NoErr(t, err)
	testutil.Equal(t, name, "catppuccin")

	if _, err := os.Stat(filepath.Join(themesDir, "catppuccin", "theme.toml")); err != nil {
		t.Errorf("installed theme.toml missing: %v", err)
	}
}

func TestInstall_derivesNameFromURL(t *testing.T) {
	repo := gitFixture(t, validThemeToml)
	themesDir := t.TempDir()

	// Clone from a path whose base name is the expected family name.
	link := filepath.Join(t.TempDir(), "Catppuccin.git")
	testutil.NoErr(t, os.Symlink(repo, link))

	name, err := Install(themesDir, link, "", false)
	testutil.NoErr(t, err)
	testutil.Equal(t, name, "catppuccin")
}

func TestInstall_rejectsBadName(t *testing.T) {
	themesDir := t.TempDir()
	for _, name := range []string{"Bad", "../escape", ".hidden", "has space"} {
		if _, err := Install(themesDir, "unused", name, false); err == nil {
			t.Errorf("name %q: expected error", name)
		}
	}
}

func TestInstall_existingWithoutForce(t *testing.T) {
	repo := gitFixture(t, validThemeToml)
	themesDir := t.TempDir()

	_, err := Install(themesDir, repo, "fam", false)
	testutil.NoErr(t, err)

	_, err = Install(themesDir, repo, "fam", false)
	if err == nil || !strings.Contains(err.Error(), "already installed") {
		t.Errorf("expected already-installed error, got %v", err)
	}
}

func TestInstall_forceReplaces(t *testing.T) {
	repo := gitFixture(t, validThemeToml)
	themesDir := t.TempDir()

	_, err := Install(themesDir, repo, "fam", false)
	testutil.NoErr(t, err)

	marker := filepath.Join(themesDir, "fam", "old-marker")
	testutil.NoErr(t, os.WriteFile(marker, []byte("old"), 0o644))

	_, err = Install(themesDir, repo, "fam", true)
	testutil.NoErr(t, err)

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Errorf("marker from previous install still present: %v", err)
	}
	if _, err := os.Stat(filepath.Join(themesDir, "fam", "theme.toml")); err != nil {
		t.Errorf("installed theme.toml missing: %v", err)
	}
}

func TestInstall_notAThemeRepo(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"-c", "user.email=t@t", "-c", "user.name=t", "commit", "--allow-empty", "-m", "init"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}

	if _, err := Install(t.TempDir(), dir, "fam", false); err == nil {
		t.Error("expected error for repo without theme.toml")
	}
}

func TestInstall_noResolvableVariant(t *testing.T) {
	repo := gitFixture(t, "[variants.mocha]\n") // no appearance anywhere

	_, err := Install(t.TempDir(), repo, "fam", false)
	if err == nil || !strings.Contains(err.Error(), "no resolvable variant") {
		t.Errorf("expected no-resolvable-variant error, got %v", err)
	}
}

func TestInstall_cloneFails(t *testing.T) {
	themesDir := t.TempDir()

	_, err := Install(themesDir, filepath.Join(t.TempDir(), "missing"), "fam", false)
	if err == nil || !strings.Contains(err.Error(), "git clone") {
		t.Errorf("expected clone error, got %v", err)
	}
}

func TestInstall_cleansUpTempDir(t *testing.T) {
	repo := gitFixture(t, validThemeToml)
	themesDir := t.TempDir()

	_, err := Install(themesDir, repo, "fam", false)
	testutil.NoErr(t, err)
	_, _ = Install(themesDir, repo, "fam", false) // failure path (already installed)

	leftovers, err := filepath.Glob(filepath.Join(themesDir, ".install-*"))
	testutil.NoErr(t, err)
	if len(leftovers) != 0 {
		t.Errorf("temp dirs not cleaned up: %v", leftovers)
	}
}

func TestDeriveFamilyName(t *testing.T) {
	tests := []struct {
		url, want string
	}{
		{"https://github.com/user/Catppuccin.git", "catppuccin"},
		{"https://github.com/user/gruvbox", "gruvbox"},
		{"https://github.com/user/gruvbox/", "gruvbox"},
		{"git@github.com:user/rose-pine.git", "rose-pine"},
	}
	for _, tc := range tests {
		testutil.Equal(t, deriveFamilyName(tc.url), tc.want)
	}
}

func TestUninstall(t *testing.T) {
	dir := t.TempDir()

	target := filepath.Join(dir, "test-theme")
	testutil.NoErr(t, os.MkdirAll(target, 0o755))
	testutil.NoErr(t, os.WriteFile(filepath.Join(target, "theme.toml"), []byte(validThemeToml), 0o644))

	testutil.NoErr(t, Uninstall(dir, "test-theme"))

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("theme dir still present after uninstall: %v", err)
	}
}

func TestUninstall_notInstalled(t *testing.T) {
	if err := Uninstall(t.TempDir(), "missing"); err == nil {
		t.Error("expected error for theme that is not installed")
	}
}
