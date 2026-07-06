package integration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/theme"
)

type Nvim struct {
	Cfg config.Config
}

func (Nvim) Name() string {
	return "nvim"
}

func (i Nvim) Apply(t theme.Resolved) error {
	sourceFile := filepath.Join(i.Cfg.CurrentDir(), "nvim.lua")
	userHomePath, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve user home dir: %w", err)
	}
	targetFile := filepath.Join(userHomePath, ".config", "nvim", "plugin", "99_theme.lua")

	return symlink(sourceFile, targetFile)
}
