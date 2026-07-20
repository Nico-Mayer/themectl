package integration

import (
	"path/filepath"
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
	testutil.Equal(t, ints[0].(Ghostty).ConfigPath, "/custom/config.ghostty")
	testutil.Equal(t, ints[1].(Helix).ConfigPath, "/custom/config.toml")
	testutil.Equal(t, ints[2].(Zed).SettingsPath, "/custom/settings.json")
}

func TestEnabled_defaultPathsWhenUnset(t *testing.T) {
	cfg := config.Config{
		Settings: config.Settings{Integrations: []string{"ghostty"}},
	}

	ints := Enabled(cfg)
	testutil.Equal(t, len(ints), 1)
	got := ints[0].(Ghostty).ConfigPath
	want := filepath.Join(".config", "ghostty", "config.ghostty")
	if !strings.HasSuffix(got, want) {
		t.Errorf("default ghostty path = %q, want suffix %q", got, want)
	}
}
