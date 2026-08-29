package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
)

func TestLoadSettings(t *testing.T) {
	t.Run("missing file returns defaults", func(t *testing.T) {
		s, err := loadSettings(filepath.Join(t.TempDir(), "nope.toml"))
		testutil.NoErr(t, err)
		if len(s.Integrations) == 0 {
			t.Error("expected default integrations")
		}
	})

	t.Run("file values override defaults, rest kept", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "settings.toml")
		data := "[ghostty]\nconfig_file = \"/custom/config.ghostty\"\n"
		testutil.NoErr(t, os.WriteFile(path, []byte(data), 0o644))

		s, err := loadSettings(path)
		testutil.NoErr(t, err)
		testutil.Equal(t, s.Ghostty.ConfigFile, "/custom/config.ghostty")
		if len(s.Integrations) == 0 {
			t.Error("default integrations lost during merge")
		}
	})

	t.Run("windows-terminal config_file expands env vars and ~", func(t *testing.T) {
		t.Setenv("HOME", "/home/u")
		t.Setenv("THEMECTL_WT_DIR", "/from-env")
		path := filepath.Join(t.TempDir(), "settings.toml")
		data := "[windows-terminal]\nconfig_file = \"$THEMECTL_WT_DIR/settings.json\"\n"
		testutil.NoErr(t, os.WriteFile(path, []byte(data), 0o644))

		s, err := loadSettings(path)
		testutil.NoErr(t, err)
		testutil.Equal(t, s.WindowsTerminal.ConfigFile, "$THEMECTL_WT_DIR/settings.json")
		testutil.Equal(t, s.WindowsTerminal.ConfigFileOr("/fallback"), "/from-env/settings.json")

		s.WindowsTerminal.ConfigFile = "~/wt/settings.json"
		testutil.Equal(t, s.WindowsTerminal.ConfigFileOr("/fallback"), "/home/u/wt/settings.json")
		s.WindowsTerminal.ConfigFile = ""
		testutil.Equal(t, s.WindowsTerminal.ConfigFileOr("/fallback"), "/fallback")
	})

	t.Run("invalid toml errors", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "settings.toml")
		testutil.NoErr(t, os.WriteFile(path, []byte("= nope"), 0o644))
		if _, err := loadSettings(path); err == nil {
			t.Error("expected parse error")
		}
	})
}

func TestFileSettings_configFileOr(t *testing.T) {
	t.Setenv("HOME", "/home/u")
	t.Setenv("THEMECTL_TEST_DIR", "/from-env")

	const fallback = "/default/config"

	tests := []struct {
		name string
		file string
		want string
	}{
		{name: "unset uses fallback", file: "", want: fallback},
		{name: "blank uses fallback", file: "   ", want: fallback},
		{name: "plain path", file: "/etc/ghostty/config", want: "/etc/ghostty/config"},
		{name: "env var", file: "$THEMECTL_TEST_DIR/config", want: "/from-env/config"},
		{name: "bare tilde", file: "~", want: "/home/u"},
		{name: "tilde prefix", file: "~/x", want: "/home/u/x"},
		{name: "tilde mid-path untouched", file: "/a/~/b", want: "/a/~/b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := FileSettings{ConfigFile: tt.file}
			testutil.Equal(t, f.ConfigFileOr(fallback), tt.want)
		})
	}
}
