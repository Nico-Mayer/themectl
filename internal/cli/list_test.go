package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/theme"
)

func TestPrintThemesJSON_Contract(t *testing.T) {
	themes := []theme.Resolved{{
		Family:     "catppuccin",
		Variant:    "mocha",
		Appearance: theme.Dark,
	}}
	out, err := captureStdout(t, func() error { return printThemesJSON(themes) })
	if err != nil {
		t.Fatal(err)
	}

	var got []map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"id":         "catppuccin/mocha",
		"family":     "catppuccin",
		"variant":    "mocha",
		"appearance": "dark",
	}
	if len(got) != 1 || len(got[0]) != len(want) {
		t.Fatalf("JSON = %v, want one item with keys %v", got, want)
	}
	for key, value := range want {
		if got[0][key] != value {
			t.Errorf("%s = %v, want %v", key, got[0][key], value)
		}
	}
}

func TestRenderThemeList_LabelsCurrentThemeWithText(t *testing.T) {
	themes := []theme.Resolved{
		{Family: "catppuccin", Variant: "mocha", Appearance: theme.Dark},
		{Family: "catppuccin", Variant: "latte", Appearance: theme.Light},
	}
	out := renderThemeList(themes, "catppuccin/mocha")
	for _, text := range []string{"Status", "catppuccin/mocha", "current", "catppuccin/latte"} {
		if !strings.Contains(out, text) {
			t.Errorf("output missing %q:\n%s", text, out)
		}
	}
	if strings.Count(out, "current") != 1 {
		t.Errorf("current label count = %d, want 1:\n%s", strings.Count(out, "current"), out)
	}
}
