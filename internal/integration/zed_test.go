package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

type fakeInstaller struct {
	refs []string
	err  error
}

func (f *fakeInstaller) Ensure(ref string) error {
	f.refs = append(f.refs, ref)
	return f.err
}

func writeZedSettings(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "settings.json")
	testutil.NoErr(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestZed_Apply(t *testing.T) {
	settings := writeZedSettings(t, `{"theme": "old", "icon_theme": "old-icons"}`)

	installer := &fakeInstaller{}
	z := Zed{ConfigFile: settings, Installer: installer}
	res := theme.Resolved{
		Family:  "catppuccin",
		Variant: "mocha",
		Zed: &theme.ZedSpec{
			Theme:      "Catppuccin Mocha",
			IconTheme:  "Catppuccin Mocha",
			Extensions: []string{"github.com/catppuccin/zed"},
		},
	}

	testutil.NoErr(t, z.Apply(res))

	out, _ := os.ReadFile(settings)
	if !strings.Contains(string(out), `"theme": "Catppuccin Mocha"`) {
		t.Errorf("theme not rewritten: %q", out)
	}
	if !strings.Contains(string(out), `"icon_theme": "Catppuccin Mocha"`) {
		t.Errorf("icon_theme not rewritten: %q", out)
	}
	testutil.Diff(t, []string{"https://github.com/catppuccin/zed"}, installer.refs)
}

func TestZed_Apply_installsExtensionsInOrder(t *testing.T) {
	settings := writeZedSettings(t, `{"theme": "old"}`)

	installer := &fakeInstaller{}
	z := Zed{ConfigFile: settings, Installer: installer}
	res := theme.Resolved{Zed: &theme.ZedSpec{
		Theme:      "X",
		Extensions: []string{"github.com/catppuccin/zed", "https://github.com/other/ext"},
	}}

	testutil.NoErr(t, z.Apply(res))
	testutil.Diff(t, []string{
		"https://github.com/catppuccin/zed",
		"https://github.com/other/ext",
	}, installer.refs)
}

func TestZed_Apply_noExtensionsSkipsInstaller(t *testing.T) {
	settings := writeZedSettings(t, `{"theme": "old"}`)

	installer := &fakeInstaller{}
	z := Zed{ConfigFile: settings, Installer: installer}

	testutil.NoErr(t, z.Apply(theme.Resolved{Zed: &theme.ZedSpec{Theme: "X"}}))
	testutil.Equal(t, len(installer.refs), 0)
}

func TestZed_Apply_nilInstaller(t *testing.T) {
	settings := writeZedSettings(t, `{"theme": "old"}`)

	z := Zed{ConfigFile: settings}
	res := theme.Resolved{Zed: &theme.ZedSpec{
		Theme:      "X",
		Extensions: []string{"github.com/catppuccin/zed"},
	}}

	testutil.NoErr(t, z.Apply(res))

	out, _ := os.ReadFile(settings)
	if !strings.Contains(string(out), `"theme": "X"`) {
		t.Errorf("theme not rewritten: %q", out)
	}
}

func TestZed_Apply_addsMissingIconThemeKey(t *testing.T) {
	settings := writeZedSettings(t, `{"theme": "old"}`)

	z := Zed{ConfigFile: settings}
	res := theme.Resolved{Zed: &theme.ZedSpec{Theme: "X", IconTheme: "X Icons"}}

	testutil.NoErr(t, z.Apply(res))

	out, _ := os.ReadFile(settings)
	if !strings.Contains(string(out), `"icon_theme": "X Icons"`) {
		t.Errorf("icon_theme not inserted: %q", out)
	}
}

func TestZed_Supports_requiresOverride(t *testing.T) {
	z := Zed{ConfigFile: "unused"}
	if z.Supports(theme.Resolved{}) {
		t.Error("theme without zed override must not be supported")
	}
	if !z.Supports(theme.Resolved{Zed: &theme.ZedSpec{Theme: "X"}}) {
		t.Error("theme with zed override must be supported")
	}
}

func TestZed_Apply_installerErrorAborts(t *testing.T) {
	settings := writeZedSettings(t, `{"theme": "old"}`)

	installer := &fakeInstaller{err: os.ErrPermission}
	z := Zed{ConfigFile: settings, Installer: installer}
	res := theme.Resolved{Zed: &theme.ZedSpec{
		Theme:      "X",
		Extensions: []string{"github.com/catppuccin/zed"},
	}}

	if err := z.Apply(res); err == nil {
		t.Fatal("expected installer error to propagate")
	}

	out, _ := os.ReadFile(settings)
	testutil.Equal(t, string(out), `{"theme": "old"}`)
}
