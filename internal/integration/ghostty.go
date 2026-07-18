package integration

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"

	"github.com/Nico-Mayer/themectl/internal/theme"
)

type Ghostty struct {
	ConfigPath string
}

var ghosttyThemeLine = regexp.MustCompile(`(?m)^(\s*theme\s*=\s*).*$`)

func setGhosttyTheme(config, themeName string) (string, error) {
	if !ghosttyThemeLine.MatchString(config) {
		return "", fmt.Errorf("no `theme =` setting found in ghostty config")
	}
	return ghosttyThemeLine.ReplaceAllString(config, `${1}"`+themeName+`"`), nil
}

func (Ghostty) Name() string {
	return "ghostty"
}

func (g Ghostty) Apply(t theme.Resolved) error {
	if t.Ghostty == nil || t.Ghostty.Theme == "" {
		return fmt.Errorf("theme %s has no ghostty override", t.ID())
	}
	name := t.Ghostty.Theme

	data, err := os.ReadFile(g.ConfigPath)
	if err != nil {
		return fmt.Errorf("read ghostty config: %w", err)
	}

	updated, err := setGhosttyTheme(string(data), name)
	if err != nil {
		return err
	}

	if err := os.WriteFile(g.ConfigPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write ghostty config: %w", err)
	}

	if err := reloadGhostty(); err != nil {
		slog.Warn("ghostty config reload failed", "err", err)
	}

	return nil
}

func (g Ghostty) Check() error {
	return checkConfigDir(g.Name(), g.ConfigPath)
}

func reloadGhostty() error {
	err := exec.Command("pkill", "-USR2", "ghostty").Run()
	if err == nil {
		return nil
	}

	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return nil
	}
	return fmt.Errorf("signal ghostty reload: %w", err)
}
