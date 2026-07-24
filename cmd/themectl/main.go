package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"charm.land/log/v2"
	"github.com/Nico-Mayer/themectl/internal/cache"
	"github.com/Nico-Mayer/themectl/internal/cli"
	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/fetch"
	"github.com/Nico-Mayer/themectl/internal/integration"
	"github.com/Nico-Mayer/themectl/internal/store"
)

func main() {
	root := defaultRoot()
	cfg, err := config.Load(root)
	if err != nil {
		log.Fatal(err)
	}

	assetTTL := time.Hour * 24 * 7
	webCache := cache.New(filepath.Join(cfg.CacheDir(), "web"))
	fetcher := fetch.NewFetcher(&http.Client{Timeout: time.Second * 15}, webCache, assetTTL)
	store := store.NewStore(os.DirFS(cfg.ThemesDir()), fetcher)

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
