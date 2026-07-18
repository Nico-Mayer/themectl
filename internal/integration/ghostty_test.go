package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

func TestSetGhosttyTheme(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		theme   string
		want    string
		wantErr bool
	}{
		{name: "unquoted value", in: "theme = old\nfont=mono\n", theme: "new", want: "theme = \"new\"\nfont=mono\n"},
		{name: "loose spacing", in: `theme     =    "old"`, theme: "new", want: `theme     =    "new"`},
		{name: "missing theme key", in: "font = mono\n", theme: "new", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := setGhosttyTheme(tt.in, tt.theme)
			testutil.Equal(t, err != nil, tt.wantErr)
			if !tt.wantErr {
				testutil.Equal(t, got, tt.want)
			}
		})
	}
}

func TestGhostty_Apply(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.ghostty")
	testutil.NoErr(t, os.WriteFile(cfgPath, []byte("theme = old"), 0o644))

	g := Ghostty{ConfigPath: cfgPath}
	res := theme.Resolved{
		Family:  "catppuccin",
		Variant: "mocha",
		Ghostty: &theme.GhosttySpec{Theme: "catppuccin-mocha"},
	}

	testutil.NoErr(t, g.Apply(res))

	out, _ := os.ReadFile(cfgPath)
	if !strings.Contains(string(out), `theme = "catppuccin-mocha"`) {
		t.Errorf("config not rewritten: %q", out)
	}
}
