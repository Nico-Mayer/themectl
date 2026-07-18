package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

func TestSetHelixTheme(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		theme   string
		want    string
		wantErr bool
	}{
		{name: "simple", in: `theme = "old"`, theme: "new", want: `theme = "new"`},
		{name: "loose spacing", in: `theme   =   "old"`, theme: "new", want: `theme   =   "new"`},
		{name: "preserves siblings",
			in:    "theme = \"old\"\n[editor]\nline-number = \"relative\"\n",
			theme: "catppuccin_mocha",
			want:  "theme = \"catppuccin_mocha\"\n[editor]\nline-number = \"relative\"\n"},
		{name: "unquoted value not matched", in: `theme = old`, theme: "new", wantErr: true},
		{name: "missing theme key", in: "[editor]\nline-number = \"relative\"\n", theme: "new", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := setHelixTheme(tt.in, tt.theme)
			testutil.Equal(t, err != nil, tt.wantErr)
			if !tt.wantErr {
				testutil.Equal(t, got, tt.want)
			}
		})
	}
}

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

func TestHelix_Apply_noOverrideFails(t *testing.T) {
	h := Helix{ConfigPath: "unused"}
	if err := h.Apply(theme.Resolved{}); err == nil {
		t.Error("expected error when theme has no helix override")
	}
}
