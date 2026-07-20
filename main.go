package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/Nico-Mayer/themectl/internal/cli"
	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/integration"
	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/charmbracelet/log"
)

func main() {
	root := defaultRoot()
	cfg, err := config.Load(root)
	if err != nil {
		log.Fatal(err)
	}

	store := store.NewStore(os.DirFS(cfg.ThemesDir()))

	app := cli.New(cfg, store, integration.Enabled(cfg))
	app.Version = version()
	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func version() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "unknown"
}

func defaultRoot() string {
	userHome, _ := os.UserHomeDir()
	return filepath.Join(userHome, ".config", "themectl")
}
