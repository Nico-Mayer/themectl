package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nico-mayer/themectl-cli/internal/theme"
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

func TestSetZedTheme_pure(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		themeName string
		want      string
		wantErr   bool
	}{
		{name: "simple", in: `{"theme": "old"}`, themeName: "One Dark", want: `{"theme": "One Dark"}`},
		{name: "loose spacing", in: `"theme"   :   "old"`, themeName: "new",
			want: `"theme"   :   "new"`},
		{name: "preserves siblings and comments",
			in:        "{\n  // pick a theme\n  \"theme\": \"old\",\n  \"vim_mode\": true\n}",
			themeName: "Catppuccin Mocha",
			want:      "{\n  // pick a theme\n  \"theme\": \"Catppuccin Mocha\",\n  \"vim_mode\": true\n}"},
		{name: "no theme key", in: `{"vim_mode": true}`, themeName: "new", wantErr: true}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := setZedTheme(tc.in, tc.themeName)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && tc.want != got {
				t.Fatalf("got: %q, want %q", got, tc.want)
			}
		})
	}
}

func TestZed_Apply(t *testing.T) {
	dir := t.TempDir()

	settings := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(settings, []byte(`{"theme": "old"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	current := filepath.Join(dir, "current")
	if err := os.MkdirAll(current, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(current, "zed.toml"),
		[]byte(`extension_url = "github.com/catppuccin/zed"`), 0o644); err != nil {
		t.Fatal(err)
	}

	installer := &fakeInstaller{}
	z := Zed{
		SettingsPath: settings,
		CurrentDir:   current,
		Installer:    installer,
	}

	res := theme.Resolved{
		Family:  "catppuccin",
		Variant: "mocha",
		Themes:  map[string]string{"zed": "Catppuccin Mocha"},
	}

	if err := z.Apply(res); err != nil {
		t.Fatal(err)
	}

	out, _ := os.ReadFile(settings)
	if !strings.Contains(string(out), `"theme": "Catppuccin Mocha"`) {
		t.Errorf("settings not rewritten: %q", out)
	}

	if installer.ncall != 1 || installer.got.URL != "github.com/catppuccin/zed" {
		t.Errorf("installer got %+v after %d calls", installer.got, installer.ncall)
	}
}

func TestZed_Apply_noSidecar_skipsInstaller(t *testing.T) {
	dir := t.TempDir()
	settings := filepath.Join(dir, "settings.json")
	os.WriteFile(settings, []byte(`{"theme": "old"}`), 0o644)
	current := filepath.Join(dir, "current")
	os.MkdirAll(current, 0o755)

	inst := &fakeInstaller{}
	z := Zed{SettingsPath: settings, CurrentDir: current, Installer: inst}
	res := theme.Resolved{Themes: map[string]string{"zed": "X"}}

	if err := z.Apply(res); err != nil {
		t.Fatal(err)
	}
	if inst.ncall != 0 {
		t.Errorf("installer called %d times without a sidecar", inst.ncall)
	}
}
