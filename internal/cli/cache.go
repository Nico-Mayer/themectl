package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Nico-Mayer/themectl/internal/cache"
	"github.com/urfave/cli/v3"
)

func (a app) cacheCmd() *cli.Command {
	return &cli.Command{
		Name:     "cache",
		Usage:    "Manage the themectl cache",
		Commands: []*cli.Command{a.clearCacheCmd()},
		Action: func(ctx context.Context, c *cli.Command) error {
			fmt.Println(a.cfg.CacheDir())
			return nil
		},
	}
}

func (a app) clearCacheCmd() *cli.Command {
	return &cli.Command{
		Name:  "clear",
		Usage: "Delete all cached files",
		Action: func(ctx context.Context, c *cli.Command) error {
			if err := cache.New(a.cfg.CacheDir()).Clear(); err != nil {
				return err
			}
			slog.Info("cache cleared")
			return nil
		},
	}
}
