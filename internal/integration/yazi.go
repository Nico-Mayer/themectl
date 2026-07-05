package integration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/theme"
)

type Yazi struct {
	Cfg config.Config
}

func (Yazi) Name() string {
	return "eza"
}

func (i Yazi) Apply(t theme.Resolved) error {
	sourceFile := filepath.Join(i.Cfg.CurrentDir(), "yazi-flavor.toml")
	userHomePath, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve user home dir: %w", err)
	}

	targetFile := filepath.Join(userHomePath, ".config", "yazi", "flavors", "themectl.yazi", "flavor.toml")

	return symlink(sourceFile, targetFile)
}
