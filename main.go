package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Nico-Mayer/themectl-cli/internal/cli"
	"github.com/Nico-Mayer/themectl-cli/internal/config"
	"github.com/Nico-Mayer/themectl-cli/internal/engine"
	"github.com/Nico-Mayer/themectl-cli/internal/integration"
	"github.com/Nico-Mayer/themectl-cli/internal/theme"
	"github.com/charmbracelet/log"
)

// version is set at build time by goreleaser via ldflags.
var version = "dev"

func main() {
	root := defaultRoot()
	cfg, err := config.Load(root)
	if err != nil {
		log.Fatal(err)
	}

	store := theme.NewStore(os.DirFS(cfg.ThemesDir()))
	engine := engine.New(integration.Enabled(cfg))

	app := cli.New(cfg, store, engine)
	app.Version = version
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func defaultRoot() string {
	userHome, _ := os.UserHomeDir()
	return filepath.Join(userHome, ".config", "themectl")
}
