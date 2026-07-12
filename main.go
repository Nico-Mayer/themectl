package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/nico-mayer/themectl-cli/internal/cli"
	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/engine"
	"github.com/nico-mayer/themectl-cli/internal/integration"
	"github.com/nico-mayer/themectl-cli/internal/theme"
)

func main() {
	root := defaultRoot()
	cfg, err := config.Load(root)
	if err != nil {
		log.Fatal(err)
	}

	store := theme.NewStore(os.DirFS(cfg.ThemesDir()))
	engine := engine.New(integration.Enabled(cfg))

	app := cli.New(cfg, store, engine)
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func defaultRoot() string {
	userHome, _ := os.UserHomeDir()
	return filepath.Join(userHome, ".config", "themectl")
}
