package integration

import (
	"fmt"
	"os"
	"regexp"

	"github.com/Nico-Mayer/themectl/internal/theme"
)

type Helix struct {
	ConfigPath string
}

var helixThemeLine = regexp.MustCompile(`(theme\s*=\s*)"[^"]*"`)

func setHelixTheme(config, themeName string) (string, error) {
	if !helixThemeLine.MatchString(config) {
		return "", fmt.Errorf("no `theme =` setting found in helix config")
	}
	return helixThemeLine.ReplaceAllString(config, `${1}"`+themeName+`"`), nil
}

func (Helix) Name() string {
	return "helix"
}

func (h Helix) Apply(t theme.Resolved) error {
	if t.Helix == nil || t.Helix.Theme == "" {
		return fmt.Errorf("theme %s has no helix override", t.ID())
	}
	name := t.Helix.Theme

	data, err := os.ReadFile(h.ConfigPath)
	if err != nil {
		return fmt.Errorf("read helix config: %w", err)
	}

	updated, err := setHelixTheme(string(data), name)
	if err != nil {
		return err
	}

	if err := os.WriteFile(h.ConfigPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write helix config: %w", err)
	}

	return nil
}

func (h Helix) Check() error {
	return checkConfigDir(h.Name(), h.ConfigPath)
}
