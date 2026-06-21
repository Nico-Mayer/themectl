package integrations

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/model"
)

type Ghostty struct{}

func init() {
	Register(Ghostty{})
}

func (Ghostty) Name() string {
	return "ghostty"
}

func (i Ghostty) Apply(themeInfo model.ThemeInfo) error {
	logger := integrationLogger(i)

	ghosttyThemeOverride, ok := themeInfo.Overrides[i.Name()]
	if !ok {
		return fmt.Errorf("no ghostty theme override provided in theme %s", themeInfo.Name)
	}

	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ghosttyConfigDir := cfg.ConfigDirFor(i.Name())
	ghosttyConfigPath := filepath.Join(ghosttyConfigDir, "config.ghostty")

	logger.Debug("updating theme", "ghostty_theme", ghosttyThemeOverride)

	data, err := os.ReadFile(ghosttyConfigPath)
	if err != nil {
		return fmt.Errorf("read Ghostty config %q: %w", ghosttyConfigPath, err)
	}
	content := string(data)

	re := regexp.MustCompile(`(theme\s*=\s*")([^"]*)(")`)
	if !re.MatchString(content) {
		return fmt.Errorf("Ghostty config %q does not contain a theme setting", ghosttyConfigPath)
	}

	updatedSettings := re.ReplaceAllString(content, `${1}`+ghosttyThemeOverride+`${3}`)

	err = os.WriteFile(ghosttyConfigPath, []byte(updatedSettings), 0644)
	if err != nil {
		return fmt.Errorf("write Ghostty config %q: %w", ghosttyConfigPath, err)
	}

	logger.Info("applied", "theme", ghosttyThemeOverride)

	cmd := exec.Command("pkill", "-USR2", "-f", "ghostty")
	if err := cmd.Run(); err != nil {
		logger.Warn("reload signal failed", "err", err)
	}

	return nil
}
