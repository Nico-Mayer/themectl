package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Nico-Mayer/themectl-cli/internal/testutil"
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
		data := "default-theme = \"catppuccin\"\n\n[config-dirs]\nghostty = \"/custom\"\n"
		testutil.NoErr(t, os.WriteFile(path, []byte(data), 0o644))

		s, err := loadSettings(path)
		testutil.NoErr(t, err)
		testutil.Equal(t, s.DefaultTheme, "catppuccin")
		testutil.Equal(t, s.ConfigDirs["ghostty"], "/custom")
		if len(s.Integrations) == 0 {
			t.Error("default integrations lost during merge")
		}
	})

	t.Run("invalid toml errors", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "settings.toml")
		testutil.NoErr(t, os.WriteFile(path, []byte("= nope"), 0o644))
		if _, err := loadSettings(path); err == nil {
			t.Error("expected parse error")
		}
	})
}

func TestDefaultConfigDirs(t *testing.T) {
	unix := defaultConfigDirs("/home/u", "")
	testutil.Equal(t, unix["zed"], filepath.Join("/home/u", ".config", "zed"))

	win := defaultConfigDirs(`C:\Users\u`, `C:\Users\u\AppData\Roaming`)
	testutil.Equal(t, win["yazi"], filepath.Join(`C:\Users\u\AppData\Roaming`, "yazi", "config"))
	testutil.Equal(t, win["ghostty"], filepath.Join(`C:\Users\u`, ".config", "ghostty"))
}

func TestConfigDirFor(t *testing.T) {
	t.Setenv("HOME", "/home/u")
	t.Setenv("THEMECTL_TEST_DIR", "/from-env")

	tests := []struct {
		name string
		dirs map[string]string
		want string
	}{
		{name: "nil map", dirs: nil, want: ""},
		{name: "missing key", dirs: map[string]string{}, want: ""},
		{name: "blank value", dirs: map[string]string{"ghostty": "   "}, want: ""},
		{name: "plain path", dirs: map[string]string{"ghostty": "/etc/ghostty"}, want: "/etc/ghostty"},
		{name: "env var", dirs: map[string]string{"ghostty": "$THEMECTL_TEST_DIR"}, want: "/from-env"},
		{name: "bare tilde", dirs: map[string]string{"ghostty": "~"}, want: "/home/u"},
		{name: "tilde prefix", dirs: map[string]string{"ghostty": "~/x"}, want: "/home/u/x"},
		{name: "tilde mid-path untouched", dirs: map[string]string{"ghostty": "/a/~/b"}, want: "/a/~/b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Settings{ConfigDirs: tt.dirs}
			testutil.Equal(t, s.ConfigDirFor("ghostty"), tt.want)
		})
	}
}
