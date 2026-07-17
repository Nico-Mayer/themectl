package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/theme"
	"github.com/nico-mayer/themectl-cli/internal/wallpaper"
	"github.com/urfave/cli/v3"
)

func wallpaperCmd(cfg config.Config, store *theme.Store) *cli.Command {
	return &cli.Command{
		Name:    "wallpaper",
		Aliases: []string{"wall"},
		Usage:   "Show or set the desktop wallpaper",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "random",
				Aliases: []string{"r"},
				Usage:   "Set a random wallpaper from the current theme",
			},
		},
		Commands: []*cli.Command{
			listWallpapersCmd(cfg, store),
			setWallpaperCmd(cfg),
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			manager := wallpaper.NewManager(cfg.ThemesDir(), cfg.SharedWallpapersDir())

			if c.Bool("random") {
				return applyRandomWallpaper(cfg, store, manager)
			}

			current, err := manager.Current()
			if err != nil {
				return fmt.Errorf("get current wallpaper: %w", err)
			}

			fmt.Println(current)
			return nil
		},
	}
}

func applyRandomWallpaper(cfg config.Config, store *theme.Store, manager wallpaper.Manager) error {
	current, err := theme.ReadCurrent(cfg.CurrentFile())
	if err != nil {
		return fmt.Errorf("read current theme: %w", err)
	}

	slog.Debug("resolving theme", "theme", current)
	resolved, err := store.Resolve(current)
	if err != nil {
		return fmt.Errorf("resolve theme %q: %w", current, err)
	}

	if err := manager.ApplyRandom(resolved); err != nil {
		return fmt.Errorf("apply random wallpaper: %w", err)
	}

	slog.Info("wallpaper set", "theme", resolved.ID())
	return nil
}

func listWallpapersCmd(cfg config.Config, store *theme.Store) *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List wallpaper candidates for the current theme",
		Action: func(ctx context.Context, c *cli.Command) error {
			current, err := theme.ReadCurrent(cfg.CurrentFile())
			if err != nil {
				return fmt.Errorf("read current theme: %w", err)
			}

			slog.Debug("resolving theme", "theme", current)
			resolved, err := store.Resolve(current)
			if err != nil {
				return fmt.Errorf("resolve theme %q: %w", current, err)
			}

			manager := wallpaper.NewManager(cfg.ThemesDir(), cfg.SharedWallpapersDir())
			for _, candidate := range manager.ListCandidates(resolved) {
				fmt.Println(candidate)
			}

			return nil
		},
	}
}

func setWallpaperCmd(cfg config.Config) *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Set the wallpaper from a file",
		ArgsUsage: "<filepath>",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "path",
				UsageText: "filepath to the wallpaper image",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			path := c.StringArg("path")
			if path == "" {
				return fmt.Errorf("no wallpaper path provided")
			}

			manager := wallpaper.NewManager(cfg.ThemesDir(), cfg.SharedWallpapersDir())
			if err := manager.Set(path); err != nil {
				return fmt.Errorf("set wallpaper %q: %w", path, err)
			}

			slog.Info("wallpaper set", "file", path)
			return nil
		},
	}
}
