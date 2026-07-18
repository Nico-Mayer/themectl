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
	got   ExtensionRef
	err   error
	ncall int
}

func (f *fakeInstaller) Ensure(ref ExtensionRef) error {
	f.ncall++
	f.got = ref
	return f.err
}

func TestSetZedTheme(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		theme   string
		want    string
		wantErr bool
	}{
		{name: "simple", in: `{"theme": "old"}`, theme: "One Dark", want: `{"theme": "One Dark"}`},
		{name: "loose spacing", in: `"theme"   :   "old"`, theme: "new", want: `"theme"   :   "new"`},
		{name: "preserves siblings and comments",
			in:    "{\n  // pick a theme\n  \"theme\": \"old\",\n  \"vim_mode\": true\n}",
			theme: "Catppuccin Mocha",
			want:  "{\n  // pick a theme\n  \"theme\": \"Catppuccin Mocha\",\n  \"vim_mode\": true\n}"},
		{name: "missing theme key", in: `{"vim_mode": true}`, theme: "new", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := setZedTheme(tt.in, tt.theme)
			testutil.Equal(t, err != nil, tt.wantErr)
			if !tt.wantErr {
				testutil.Equal(t, got, tt.want)
			}
		})
	}
}

func TestZed_Apply(t *testing.T) {
	dir := t.TempDir()
	settings := filepath.Join(dir, "settings.json")
	current := filepath.Join(dir, "current")
	testutil.NoErr(t, os.WriteFile(settings, []byte(`{"theme": "old"}`), 0o644))
	testutil.NoErr(t, os.MkdirAll(current, 0o755))
	testutil.NoErr(t, os.WriteFile(filepath.Join(current, "zed.toml"),
		[]byte(`extension_url = "github.com/catppuccin/zed"`), 0o644))

	installer := &fakeInstaller{}
	z := Zed{SettingsPath: settings, CurrentDir: current, Installer: installer}
	res := theme.Resolved{
		Family:  "catppuccin",
		Variant: "mocha",
		Zed:     &theme.ZedSpec{Theme: "Catppuccin Mocha"},
	}

	testutil.NoErr(t, z.Apply(res))

	out, _ := os.ReadFile(settings)
	if !strings.Contains(string(out), `"theme": "Catppuccin Mocha"`) {
		t.Errorf("settings not rewritten: %q", out)
	}
	testutil.Equal(t, installer.ncall, 1)
	testutil.Equal(t, installer.got.URL, "github.com/catppuccin/zed")
}

func TestZed_Apply_noSidecarSkipsInstaller(t *testing.T) {
	dir := t.TempDir()
	settings := filepath.Join(dir, "settings.json")
	current := filepath.Join(dir, "current")
	testutil.NoErr(t, os.WriteFile(settings, []byte(`{"theme": "old"}`), 0o644))
	testutil.NoErr(t, os.MkdirAll(current, 0o755))

	installer := &fakeInstaller{}
	z := Zed{SettingsPath: settings, CurrentDir: current, Installer: installer}

	testutil.NoErr(t, z.Apply(theme.Resolved{Zed: &theme.ZedSpec{Theme: "X"}}))
	testutil.Equal(t, installer.ncall, 0)
}
