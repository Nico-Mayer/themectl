package integration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Nico-Mayer/themectl-cli/internal/config"
	"github.com/Nico-Mayer/themectl-cli/internal/theme"
)

type Eza struct {
	Cfg config.Config
}

func (Eza) Name() string {
	return "eza"
}

func (i Eza) Apply(t theme.Resolved) error {
	sourceFile := filepath.Join(i.Cfg.CurrentDir(), "eza.yml")
	userHomePath, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve user home dir: %w", err)
	}
	targetFile := filepath.Join(userHomePath, ".config", "eza", "theme.yml")

	return symlink(sourceFile, targetFile)
}
