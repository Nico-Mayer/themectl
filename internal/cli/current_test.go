package cli

import (
	"encoding/json"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/store"
)

func TestPrintCurrentJSON_Contract(t *testing.T) {
	st := store.NewStore(testThemeFS(), nil)
	out, err := captureStdout(t, func() error {
		return printCurrentJSON("catppuccin/mocha", st)
	})
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"id":         "catppuccin/mocha",
		"family":     "catppuccin",
		"variant":    "mocha",
		"appearance": "dark",
	}
	if len(got) != len(want) {
		t.Fatalf("JSON keys = %v, want %v", got, want)
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("%s = %v, want %v", key, got[key], value)
		}
	}
}
