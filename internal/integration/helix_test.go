package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

func TestHelix_Apply(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.toml")
	testutil.NoErr(t, os.WriteFile(cfgPath, []byte(`theme = "old"`), 0o644))

	h := Helix{ConfigPath: cfgPath}
	res := theme.Resolved{
		Family:  "catppuccin",
		Variant: "mocha",
		Helix:   &theme.HelixSpec{Theme: "catppuccin_mocha"},
	}

	testutil.NoErr(t, h.Apply(res))

	out, _ := os.ReadFile(cfgPath)
	if !strings.Contains(string(out), `theme = "catppuccin_mocha"`) {
		t.Errorf("config not rewritten: %q", out)
	}
}

func TestHelix_Supports_requiresOverride(t *testing.T) {
	h := Helix{ConfigPath: "unused"}
	if h.Supports(theme.Resolved{}) {
		t.Error("theme without helix override must not be supported")
	}
	if !h.Supports(theme.Resolved{Helix: &theme.HelixSpec{Theme: "X"}}) {
		t.Error("theme with helix override must be supported")
	}
}
