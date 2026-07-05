package integration

import (
	"fmt"
	"os"
	"regexp"

	"github.com/nico-mayer/themectl-cli/internal/theme"
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

func (i Helix) Apply(t theme.Resolved) error {
	name, ok := t.Themes[i.Name()]
	if !ok {
		return fmt.Errorf("theme %s has no helix overwride", t.ID())
	}

	data, err := os.ReadFile(i.ConfigPath)
	if err != nil {
		return fmt.Errorf("read helix config: %w", err)
	}

	updated, err := setHelixTheme(string(data), name)
	if err != nil {
		return err
	}

	if err := os.WriteFile(i.ConfigPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write helix config: %w", err)
	}

	return nil
}
