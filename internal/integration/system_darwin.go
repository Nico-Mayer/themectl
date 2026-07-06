//go:build darwin

package integration

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/nico-mayer/themectl-cli/internal/theme"
)

func setSystemAppearance(appearance theme.Appearance) error {
	var script string
	switch appearance {
	case theme.Dark:
		script = `tell application "System Events" to tell appearance preferences to set dark mode to true`
	case theme.Light:
		script = `tell application "System Events" to tell appearance preferences to set dark mode to false`
	default:
		return fmt.Errorf("unsupported appearance %q: expected %q or %q", appearance, theme.Dark, theme.Light)
	}

	output, err := exec.Command("osascript", "-e", script).CombinedOutput()
	if err != nil {
		return fmt.Errorf("osascript failed: %w (output: %s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}
