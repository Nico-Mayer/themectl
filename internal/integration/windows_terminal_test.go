package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/testutil"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

// The real files from catppuccin/windows-terminal, so the rename and the
// fragment are exercised against what a theme actually ships.
const catppuccinMochaScheme = `{
  "name": "Catppuccin Mocha",
  "cursorColor": "#F5E0DC",
  "selectionBackground": "#585B70",
  "background": "#1E1E2E",
  "foreground": "#CDD6F4",
  "black": "#45475A",
  "red": "#F38BA8",
  "brightWhite": "#A6ADC8"
}`

const catppuccinMochaChrome = `{
  "name": "Catppuccin Mocha",
  "tab": {
    "background": "#1E1E2EFF",
    "showCloseButton": "always",
    "unfocusedBackground": null
  },
  "tabRow": {
    "background": "#181825FF",
    "unfocusedBackground": "#11111BFF"
  },
  "window": {
    "applicationTheme": "dark"
  }
}`

// A settings file shaped like a real one: comments, a user scheme, a user
// chrome, and a profile with its own colorScheme.
const wtUserSettings = `{
  // themectl must not disturb any of this
  "$schema": "https://aka.ms/terminal-profiles-schema",
  "defaultProfile": "{guid-a}",
  "profiles": {
    "defaults": {
      "font": { "face": "Cascadia Code" }
    },
    "list": [
      {
        "guid": "{guid-a}",
        "colorScheme": "MyOwnScheme"
      }
    ]
  },
  "schemes": [
    { "name": "MyOwnScheme", "background": "#000000" }
  ],
  "themes": [
    { "name": "MyOwnChrome", "window": { "applicationTheme": "light" } }
  ]
}`

func newTestWT(t *testing.T, settings, scheme, chrome string) WindowsTerminal {
	t.Helper()
	dir := t.TempDir()
	cur := filepath.Join(dir, "current")
	testutil.NoErr(t, os.MkdirAll(cur, 0o755))

	cfgFile := filepath.Join(dir, "settings.json")
	testutil.NoErr(t, os.WriteFile(cfgFile, []byte(settings), 0o644))

	w := WindowsTerminal{
		ConfigFile:   cfgFile,
		FragmentFile: filepath.Join(dir, "Fragments", "themectl", "scheme.json"),
		SchemeFile:   filepath.Join(cur, theme.WindowsTerminalAssetName),
		ThemeFile:    filepath.Join(cur, theme.WindowsTerminalThemeAssetName),
	}
	if scheme != "" {
		testutil.NoErr(t, os.WriteFile(w.SchemeFile, []byte(scheme), 0o644))
	}
	if chrome != "" {
		testutil.NoErr(t, os.WriteFile(w.ThemeFile, []byte(chrome), 0o644))
	}
	return w
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	testutil.NoErr(t, err)
	return string(b)
}

func TestWindowsTerminal_paths(t *testing.T) {
	t.Run("defaults to the store install", func(t *testing.T) {
		t.Setenv("LOCALAPPDATA", "/local")
		cfg := config.Config{Root: t.TempDir()}
		w := newWindowsTerminal(cfg).(WindowsTerminal)

		testutil.Equal(t, w.ConfigFile, filepath.Join(
			"/local", "Packages", "Microsoft.WindowsTerminal_8wekyb3d8bbwe", "LocalState", "settings.json"))
		testutil.Equal(t, w.FragmentFile, filepath.Join(
			"/local", "Microsoft", "Windows Terminal", "Fragments", "themectl", "scheme.json"))
	})

	t.Run("config_file overrides the settings path only", func(t *testing.T) {
		t.Setenv("LOCALAPPDATA", "/local")
		cfg := config.Config{Root: t.TempDir()}
		cfg.Settings.WindowsTerminal.ConfigFile = "/portable/settings.json"
		w := newWindowsTerminal(cfg).(WindowsTerminal)

		testutil.Equal(t, w.ConfigFile, "/portable/settings.json")
		// fragments live at the same user-level path for every install type
		testutil.Equal(t, w.FragmentFile, filepath.Join(
			"/local", "Microsoft", "Windows Terminal", "Fragments", "themectl", "scheme.json"))
	})

	t.Run("missing LOCALAPPDATA falls back rather than building a rooted path", func(t *testing.T) {
		t.Setenv("LOCALAPPDATA", "")
		got := windowsTerminalSettingsFile()
		if strings.HasPrefix(got, string(filepath.Separator)+"Packages") {
			t.Errorf("built a path with an empty base: %q", got)
		}
		if !strings.HasSuffix(got, filepath.Join("LocalState", "settings.json")) {
			t.Errorf("unexpected settings path: %q", got)
		}
	})
}

func TestWindowsTerminal_Supports(t *testing.T) {
	tests := []struct {
		name   string
		scheme string
		chrome string
		want   bool
	}{
		{name: "scheme and chrome", scheme: catppuccinMochaScheme, chrome: catppuccinMochaChrome, want: true},
		{name: "scheme only", scheme: catppuccinMochaScheme, want: true},
		{name: "chrome only is not enough", chrome: catppuccinMochaChrome, want: false},
		{name: "neither", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newTestWT(t, wtUserSettings, tt.scheme, tt.chrome)
			testutil.Equal(t, w.Supports(theme.Resolved{}), tt.want)
		})
	}
}

func TestWindowsTerminal_Check(t *testing.T) {
	w := newTestWT(t, wtUserSettings, catppuccinMochaScheme, "")
	testutil.NoErr(t, w.Check())

	w.ConfigFile = filepath.Join(t.TempDir(), "nope", "settings.json")
	if err := w.Check(); err == nil {
		t.Error("expected an error when the settings dir is missing")
	}
}

func TestRenameJSONAsset(t *testing.T) {
	for _, in := range []string{catppuccinMochaScheme, catppuccinMochaChrome} {
		got, err := renameJSONAsset([]byte(in))
		testutil.NoErr(t, err)
		if !strings.Contains(string(got), `"name": "themectl"`) {
			t.Errorf("not renamed: %s", got)
		}
		if strings.Contains(string(got), "Catppuccin Mocha") {
			t.Errorf("original name survived: %s", got)
		}
	}

	// an asset with no name at all still gets one
	got, err := renameJSONAsset([]byte(`{"background": "#000000"}`))
	testutil.NoErr(t, err)
	if !strings.Contains(string(got), `"name": "themectl"`) {
		t.Errorf("name not added: %s", got)
	}
}

func TestWindowsTerminal_Apply_scheme(t *testing.T) {
	w := newTestWT(t, wtUserSettings, catppuccinMochaScheme, catppuccinMochaChrome)
	testutil.NoErr(t, w.Apply(theme.Resolved{}))

	out := readFile(t, w.ConfigFile)
	if !strings.Contains(out, `"colorScheme": "themectl"`) {
		t.Errorf("profile defaults not pointed at themectl: %s", out)
	}

	// the scheme itself lives in the fragment, never in the user's schemes array
	frag := readFile(t, w.FragmentFile)
	if !strings.Contains(frag, `"schemes"`) || !strings.Contains(frag, `"name": "themectl"`) {
		t.Errorf("fragment missing the renamed scheme: %s", frag)
	}
	if !strings.Contains(frag, "#1E1E2E") {
		t.Errorf("fragment lost the theme colors: %s", frag)
	}
}

func TestWindowsTerminal_Apply_preservesUserContent(t *testing.T) {
	w := newTestWT(t, wtUserSettings, catppuccinMochaScheme, catppuccinMochaChrome)
	testutil.NoErr(t, w.Apply(theme.Resolved{}))

	out := readFile(t, w.ConfigFile)
	for _, keep := range []string{
		"// themectl must not disturb any of this",
		`"$schema": "https://aka.ms/terminal-profiles-schema"`,
		`{ "name": "MyOwnScheme", "background": "#000000" }`,
		`{ "name": "MyOwnChrome", "window": { "applicationTheme": "light" } }`,
		`"face": "Cascadia Code"`,
		// a profile that sets its own scheme keeps it
		`"colorScheme": "MyOwnScheme"`,
	} {
		if !strings.Contains(out, keep) {
			t.Errorf("lost user content %q:\n%s", keep, out)
		}
	}
}

func TestWindowsTerminal_Apply_chrome(t *testing.T) {
	w := newTestWT(t, wtUserSettings, catppuccinMochaScheme, catppuccinMochaChrome)
	testutil.NoErr(t, w.Apply(theme.Resolved{}))

	out := readFile(t, w.ConfigFile)
	if !strings.Contains(out, `"theme": "themectl"`) {
		t.Errorf("chrome not selected: %s", out)
	}
	// the window mode comes from the asset, not from theme.Resolved.Appearance
	if !strings.Contains(out, `"applicationTheme": "dark"`) {
		t.Errorf("chrome asset not installed: %s", out)
	}
	if !strings.Contains(out, "#181825FF") {
		t.Errorf("tab row color lost: %s", out)
	}
}

// The chrome upsert is on the hot path, so switching themes repeatedly must not
// pile up entries.
func TestWindowsTerminal_Apply_repeatedLeavesOneChrome(t *testing.T) {
	w := newTestWT(t, wtUserSettings, catppuccinMochaScheme, catppuccinMochaChrome)
	testutil.NoErr(t, w.Apply(theme.Resolved{}))
	first := readFile(t, w.ConfigFile)

	testutil.NoErr(t, os.WriteFile(w.ThemeFile,
		[]byte(`{"name": "Catppuccin Latte", "window": {"applicationTheme": "light"}}`), 0o644))
	testutil.NoErr(t, w.Apply(theme.Resolved{}))
	second := readFile(t, w.ConfigFile)

	testutil.Equal(t, strings.Count(first, `"name": "themectl"`), 1)
	testutil.Equal(t, strings.Count(second, `"name": "themectl"`), 1)
	if !strings.Contains(second, `"applicationTheme": "light"`) {
		t.Errorf("chrome not replaced on the second apply: %s", second)
	}
	if strings.Contains(second, "#181825FF") {
		t.Errorf("previous chrome survived: %s", second)
	}
	if !strings.Contains(second, `{ "name": "MyOwnChrome"`) {
		t.Errorf("user chrome lost across applies: %s", second)
	}
}

func TestWindowsTerminal_Apply_schemeOnlyClearsChrome(t *testing.T) {
	w := newTestWT(t, wtUserSettings, catppuccinMochaScheme, catppuccinMochaChrome)
	testutil.NoErr(t, w.Apply(theme.Resolved{}))

	// next theme ships a scheme but no chrome
	testutil.NoErr(t, os.Remove(w.ThemeFile))
	testutil.NoErr(t, w.Apply(theme.Resolved{}))

	out := readFile(t, w.ConfigFile)
	if strings.Contains(out, `"name": "themectl"`) {
		t.Errorf("stale chrome left behind: %s", out)
	}
	if strings.Contains(out, `"theme": "themectl"`) {
		t.Errorf("chrome still selected: %s", out)
	}
	// the scheme still applies
	if !strings.Contains(out, `"colorScheme": "themectl"`) {
		t.Errorf("scheme not applied: %s", out)
	}
	if !strings.Contains(out, `{ "name": "MyOwnChrome"`) {
		t.Errorf("user chrome lost: %s", out)
	}
}

func TestWindowsTerminal_Reset(t *testing.T) {
	w := newTestWT(t, wtUserSettings, catppuccinMochaScheme, catppuccinMochaChrome)
	testutil.NoErr(t, w.Apply(theme.Resolved{}))
	testutil.NoErr(t, w.Reset())

	testutil.Equal(t, readFile(t, w.ConfigFile), wtUserSettings)

	if _, err := os.Stat(w.FragmentFile); !os.IsNotExist(err) {
		t.Errorf("fragment not removed: %v", err)
	}
}

func TestWindowsTerminal_Reset_toleratesMissingFragment(t *testing.T) {
	w := newTestWT(t, wtUserSettings, catppuccinMochaScheme, catppuccinMochaChrome)
	testutil.NoErr(t, w.Reset())
	testutil.NoErr(t, w.Reset())
	testutil.Equal(t, readFile(t, w.ConfigFile), wtUserSettings)
}
