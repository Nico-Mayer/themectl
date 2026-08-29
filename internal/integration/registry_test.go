package integration

import (
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/testutil"
)

func TestEnabled_unknownNamesIgnored(t *testing.T) {
	cfg := config.Config{
		Settings: config.Settings{Integrations: []string{""}},
	}

	testutil.Equal(t, len(Enabled(cfg)), 0)
}

func TestEnabled_settingsOverridePaths(t *testing.T) {
	cfg := config.Config{
		Settings: config.Settings{
			Integrations: []string{"ghostty", "helix", "zed"},
			Ghostty:      config.FileSettings{ConfigFile: "/custom/config.ghostty"},
			Helix:        config.FileSettings{ConfigFile: "/custom/config.toml"},
			Zed:          config.FileSettings{ConfigFile: "/custom/settings.json"},
		},
	}

	ints := Enabled(cfg)
	testutil.Equal(t, len(ints), 3)
	testutil.Equal(t, ints[0].(Ghostty).ConfigFile, "/custom/config.ghostty")
	testutil.Equal(t, ints[1].(Helix).ConfigFile, "/custom/config.toml")
	testutil.Equal(t, ints[2].(Zed).ConfigFile, "/custom/settings.json")
}

func TestEnabled_defaultPathsWhenUnset(t *testing.T) {
	cfg := config.Config{
		Settings: config.Settings{Integrations: []string{"ghostty"}},
	}

	ints := Enabled(cfg)
	testutil.Equal(t, len(ints), 1)
	got := ints[0].(Ghostty).ConfigFile
	want := filepath.Join(".config", "ghostty", "config.ghostty")
	if !strings.HasSuffix(got, want) {
		t.Errorf("default ghostty path = %q, want suffix %q", got, want)
	}
}

// The registry is platform-gated but the schema must not be, so the two name
// lists have to disagree in exactly one direction.
func TestNamesAreGatedButAllNamesAreNot(t *testing.T) {
	onWindows := runtime.GOOS == "windows"

	testutil.Equal(t, slices.Contains(Names(), "windows-terminal"), onWindows)
	testutil.Equal(t, slices.Contains(AllNames(), "windows-terminal"), true)

	// AllNames must stay sorted and free of duplicates on every platform
	testutil.Equal(t, slices.IsSorted(AllNames()), true)
	testutil.Equal(t, len(slices.Compact(slices.Clone(AllNames()))), len(AllNames()))

	// every registered integration is also a known one
	for _, n := range Names() {
		testutil.Equal(t, slices.Contains(AllNames(), n), true)
	}
}

func TestUnknown_ignoresIntegrationsNotBuiltHere(t *testing.T) {
	var cfg config.Config
	cfg.Settings = config.Settings{Integrations: defaultSettingsIntegrations(t)}
	testutil.Equal(t, len(Unknown(cfg)), 0)

	// a settings file shared between a Windows and a macOS machine must not
	// light up red on either
	cfg.Settings.Integrations = []string{"ghostty", "windows-terminal"}
	testutil.Equal(t, len(Unknown(cfg)), 0)

	cfg.Settings.Integrations = []string{"ghostty", "not-a-thing"}
	testutil.Diff(t, []string{"not-a-thing"}, Unknown(cfg))
}

// defaultSettingsIntegrations mirrors what a user gets with no settings file.
func defaultSettingsIntegrations(t *testing.T) []string {
	t.Helper()
	cfg, err := config.Load(t.TempDir())
	testutil.NoErr(t, err)
	return cfg.Settings.Integrations
}
