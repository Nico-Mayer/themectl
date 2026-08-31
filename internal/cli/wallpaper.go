package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"charm.land/huh/v2"
	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/Nico-Mayer/themectl/internal/theme"
	"github.com/Nico-Mayer/themectl/internal/wallpaper"
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
				return fmt.Errorf("select wallpaper: %w", err)
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
				UsageText: "Theme ID (default: current theme)",
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
		Usage:     "Set the wallpaper from a file or the current theme",
		ArgsUsage: "<filepath>",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "path",
				UsageText: "Path to wallpaper image",
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "random",
				Aliases: []string{"r"},
				Usage:   "Select a random wallpaper from the current theme",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			manager := wallpaper.NewManager(a.cfg.ThemesDir(), a.cfg.SharedWallpapersDir())

			if c.Bool("random") {
				return applyRandomWallpaper(a.cfg, a.store, manager)
			}

			path := c.StringArg("path")
			if path == "" {
				return fmt.Errorf("wallpaper path is required unless --random is used")
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
		if errors.Is(err, wallpaper.ErrNoCandidates) {
			slog.Warn(
				"wallpaper not changed",
				"theme", resolved.ID(),
				"reason", "no candidates; run `themectl wallpaper set <filepath>` to choose a file",
			)
			return nil
		}
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
		return "", fmt.Errorf("no wallpaper candidates found; run `themectl wallpaper set <filepath>` to choose a file")
	}

	options := make([]huh.Option[string], len(candidates))
	for i, candidate := range candidates {
		options[i] = huh.NewOption(filepath.Base(candidate), candidate)
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a wallpaper").
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
	usesCurrent := themeID == ""
	if usesCurrent {
		current, err := readCurrentTheme(cfg.CurrentFile())
		if err != nil {
			return theme.Resolved{}, err
		}
		themeID = current
	}

	slog.Debug("resolving theme", "theme", themeID)
	resolved, err := st.Resolve(themeID)
	if err != nil {
		if usesCurrent {
			return theme.Resolved{}, fmt.Errorf("resolve current theme %q: %w; run `themectl set` to select an installed theme", strings.TrimSpace(themeID), err)
		}
		return theme.Resolved{}, fmt.Errorf("resolve theme %q: %w; run `themectl list` to view available themes", themeID, err)
	}

	return resolved, nil
}
