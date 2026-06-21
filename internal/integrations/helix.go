package integrations

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/model"
)

type Helix struct{}

func init() {
	Register(Helix{})
}

func (Helix) Name() string {
	return "helix"
}

func (i Helix) Apply(themeInfo model.ThemeInfo) error {
	logger := integrationLogger(i)

	helixThemeOverride, ok := themeInfo.Overrides[i.Name()]
	if !ok {
		return fmt.Errorf("no helix theme override provided in theme %s", themeInfo.Name)
	}

	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	helixConfigDir := cfg.Settings.ConfigDirFor(i.Name())
	helixConfigPath := filepath.Join(helixConfigDir, "config.toml")

	logger.Debug("updating theme,", "helix_theme", helixThemeOverride)
	data, err := os.ReadFile(helixConfigPath)
	if err != nil {
		return fmt.Errorf("read helix config %q: %w", helixConfigPath, err)
	}
	content := string(data)

	re := regexp.MustCompile(`(theme\s*=\s*")([^"]*)(")`)
	if !re.MatchString(content) {
		return fmt.Errorf("Helix config %q does not contain a theme setting", helixConfigPath)
	}
	updatedSettings := re.ReplaceAllString(content, `${1}`+helixThemeOverride+`${3}`)

	err = os.WriteFile(helixConfigPath, []byte(updatedSettings), 0644)
	if err != nil {
		return fmt.Errorf("write helix config %q: %w", helixConfigPath, err)
	}

	logger.Info("applied", "theme", helixThemeOverride)

	return nil
}
