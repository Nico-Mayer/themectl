package cli

import (
	"strings"
	"testing"
)

func TestUninstallConfirmation_DisclosesConsequence(t *testing.T) {
	got := uninstallConfirmation("custom")
	for _, text := range []string{"Uninstall theme family \"custom\"", "deletes its local files", "cannot be undone"} {
		if !strings.Contains(got, text) {
			t.Errorf("confirmation %q missing %q", got, text)
		}
	}
}
