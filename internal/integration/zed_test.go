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
	refs []ExtensionRef
	err  error
}

func (f *fakeInstaller) Ensure(ref ExtensionRef) error {
	f.refs = append(f.refs, ref)
	return f.err
}

func TestSetZedString(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		key     string
		value   string
		want    string
		wantErr bool
	}{
		{name: "simple", in: `{"theme": "old"}`, key: "theme", value: "One Dark", want: `{"theme": "One Dark"}`},
		{name: "loose spacing", in: `{"theme"   :   "old"}`, key: "theme", value: "new", want: `{"theme"   :   "new"}`},
		{name: "preserves siblings and comments",
			in:    "{\n  // pick a theme\n  \"theme\": \"old\",\n  \"vim_mode\": true\n}",
			key:   "theme",
			value: "Catppuccin Mocha",
			want:  "{\n  // pick a theme\n  \"theme\": \"Catppuccin Mocha\",\n  \"vim_mode\": true\n}"},
		{name: "icon_theme", in: `{"icon_theme": "old"}`, key: "icon_theme", value: "new", want: `{"icon_theme": "new"}`},
		{name: "missing key appended", in: `{"vim_mode": true}`, key: "theme", value: "new",
			want: "{\"vim_mode\": true,\n  \"theme\": \"new\"\n}"},
		{name: "missing key appended after trailing comma", in: "{\n  \"vim_mode\": true,\n}", key: "theme", value: "new",
			want: "{\n  \"vim_mode\": true,\n  \"theme\": \"new\"\n}"},
		{name: "missing key appended to empty object", in: `{}`, key: "theme", value: "new",
			want: "{\n  \"theme\": \"new\"\n}"},
		{name: "value with quotes escaped", in: `{}`, key: "theme", value: `say "hi"`,
			want: "{\n  \"theme\": \"say \\\"hi\\\"\"\n}"},
		{name: "replaced value with dollar sign", in: `{"theme": "old"}`, key: "theme", value: "a$1b",
			want: `{"theme": "a$1b"}`},
		{name: "leading comment with brace ignored", in: "// {settings}\n{\n  \"vim_mode\": true\n}", key: "theme", value: "new",
			want: "// {settings}\n{\n  \"vim_mode\": true,\n  \"theme\": \"new\"\n}"},
		{name: "no object", in: `// just a comment`, key: "theme", value: "new", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := setZedString(tt.in, tt.key, tt.value)
			testutil.Equal(t, err != nil, tt.wantErr)
			if !tt.wantErr {
				testutil.Equal(t, got, tt.want)
			}
		})
	}
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
	z := Zed{SettingsPath: settings, Installer: installer}
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
	testutil.Diff(t, []ExtensionRef{{URL: "github.com/catppuccin/zed"}}, installer.refs)
}

func TestZed_Apply_installsExtensionsInOrder(t *testing.T) {
	settings := writeZedSettings(t, `{"theme": "old"}`)

	installer := &fakeInstaller{}
	z := Zed{SettingsPath: settings, Installer: installer}
	res := theme.Resolved{Zed: &theme.ZedSpec{
		Theme:      "X",
		Extensions: []string{"github.com/catppuccin/zed", "github.com/other/ext"},
	}}

	testutil.NoErr(t, z.Apply(res))
	testutil.Diff(t, []ExtensionRef{
		{URL: "github.com/catppuccin/zed"},
		{URL: "github.com/other/ext"},
	}, installer.refs)
}

func TestZed_Apply_noExtensionsSkipsInstaller(t *testing.T) {
	settings := writeZedSettings(t, `{"theme": "old"}`)

	installer := &fakeInstaller{}
	z := Zed{SettingsPath: settings, Installer: installer}

	testutil.NoErr(t, z.Apply(theme.Resolved{Zed: &theme.ZedSpec{Theme: "X"}}))
	testutil.Equal(t, len(installer.refs), 0)
}

func TestZed_Apply_nilInstaller(t *testing.T) {
	settings := writeZedSettings(t, `{"theme": "old"}`)

	z := Zed{SettingsPath: settings}
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

	z := Zed{SettingsPath: settings}
	res := theme.Resolved{Zed: &theme.ZedSpec{Theme: "X", IconTheme: "X Icons"}}

	testutil.NoErr(t, z.Apply(res))

	out, _ := os.ReadFile(settings)
	if !strings.Contains(string(out), `"icon_theme": "X Icons"`) {
		t.Errorf("icon_theme not inserted: %q", out)
	}
}

func TestZed_Apply_noOverrideFails(t *testing.T) {
	z := Zed{SettingsPath: "unused"}
	if err := z.Apply(theme.Resolved{}); err == nil {
		t.Error("expected error when theme has no zed override")
	}
}

func TestZed_Apply_installerErrorAborts(t *testing.T) {
	settings := writeZedSettings(t, `{"theme": "old"}`)

	installer := &fakeInstaller{err: os.ErrPermission}
	z := Zed{SettingsPath: settings, Installer: installer}
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
