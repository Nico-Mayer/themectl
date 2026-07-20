package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/Nico-Mayer/themectl/internal/theme"
	"github.com/Nico-Mayer/themectl/internal/wallpaper"
	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

func (a app) wallpaperCmd() *cli.Command {
	return &cli.Command{
		Name:    "wallpaper",
		Aliases: []string{"wall"},
		Usage:   "Select and set the desktop wallpaper",
		Commands: []*cli.Command{
			a.listWallpapersCmd(),
			a.setWallpaperCmd(),
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			resolved, err := resolveThemeOrCurrent(a.cfg, a.store, "")
			if err != nil {
				return err
			}

			manager := wallpaper.NewManager(a.cfg.ThemesDir(), a.cfg.SharedWallpapersDir())
			selected, err := pickWallpaper(manager.ListCandidates(resolved))
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			if err != nil {
				return err
			}

			return setWallpaper(manager, selected)
		},
	}
}

func (a app) listWallpapersCmd() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Aliases:   []string{"ls"},
		Usage:     "List wallpaper candidates for a theme",
		ArgsUsage: "<theme>",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "theme",
				UsageText: "theme ID (defaults to the current theme)",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			resolved, err := resolveThemeOrCurrent(a.cfg, a.store, c.StringArg("theme"))
			if err != nil {
				return err
			}

			manager := wallpaper.NewManager(a.cfg.ThemesDir(), a.cfg.SharedWallpapersDir())
			for _, candidate := range manager.ListCandidates(resolved) {
				fmt.Println(candidate)
			}

			return nil
		},
	}
}

func (a app) setWallpaperCmd() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Usage:     "Set the wallpaper from a file or at random",
		ArgsUsage: "<filepath>",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "path",
				UsageText: "filepath to the wallpaper image",
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "random",
				Aliases: []string{"r"},
				Usage:   "Set a random wallpaper from the current theme",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			manager := wallpaper.NewManager(a.cfg.ThemesDir(), a.cfg.SharedWallpapersDir())

			if c.Bool("random") {
				return applyRandomWallpaper(a.cfg, a.store, manager)
			}

			path := c.StringArg("path")
			if path == "" {
				return fmt.Errorf("no wallpaper path provided")
			}
			return setWallpaper(manager, path)
		},
	}
}

func applyRandomWallpaper(cfg config.Config, store *store.Store, manager wallpaper.Manager) error {
	resolved, err := resolveThemeOrCurrent(cfg, store, "")
	if err != nil {
		return err
	}

	if err := manager.ApplyRandom(resolved); err != nil {
		return fmt.Errorf("apply random wallpaper: %w", err)
	}

	slog.Info("wallpaper set", "theme", resolved.ID())
	return nil
}

func setWallpaper(manager wallpaper.Manager, path string) error {
	if err := manager.Set(path); err != nil {
		return fmt.Errorf("set wallpaper %q: %w", path, err)
	}

	slog.Info("wallpaper set", "file", path)
	return nil
}

func pickWallpaper(candidates []string) (string, error) {
	if len(candidates) == 0 {
		return "", fmt.Errorf("no wallpaper candidates found")
	}

	options := make([]huh.Option[string], len(candidates))
	for i, candidate := range candidates {
		options[i] = huh.NewOption(filepath.Base(candidate), candidate)
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Wallpapers").
				Options(options...).
				Filtering(true).
				// Height(10).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	return selected, nil
}

func resolveThemeOrCurrent(cfg config.Config, st *store.Store, themeID string) (theme.Resolved, error) {
	if themeID == "" {
		current, err := store.ReadCurrent(cfg.CurrentFile())
		if err != nil {
			return theme.Resolved{}, fmt.Errorf("read current theme: %w", err)
		}
		themeID = current
	}

	slog.Debug("resolving theme", "theme", themeID)
	resolved, err := st.Resolve(themeID)
	if err != nil {
		return theme.Resolved{}, fmt.Errorf("resolve theme %q: %w", themeID, err)
	}

	return resolved, nil
}
