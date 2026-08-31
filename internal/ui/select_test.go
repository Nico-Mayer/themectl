package ui

import (
	"strings"
	"testing"
)

func TestNewSelect_UsesProvidedTitle(t *testing.T) {
	var selected string
	view := newSelect("Select a theme family", []string{"catppuccin"}, &selected).View()
	if !strings.Contains(view, "Select a theme family") {
		t.Fatalf("select view does not contain provided title:\n%s", view)
	}
	if strings.Contains(view, "Pick a Theme") {
		t.Fatalf("select view contains old hard-coded title:\n%s", view)
	}
}
