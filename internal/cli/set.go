package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"charm.land/huh/v2"
	"github.com/Nico-Mayer/themectl/internal/integration"
	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/Nico-Mayer/themectl/internal/theme"
	"github.com/Nico-Mayer/themectl/internal/ui"
	"github.com/urfave/cli/v3"
)

func (a app) setCmd() *cli.Command {
	return &cli.Command{
		Name:      "set",
		Aliases:   []string{"use", "apply"},
		Usage:     "Set the active theme",
		ArgsUsage: "<theme>",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "theme",
				UsageText: "theme name (see 'themectl list')",
			},
		},
		Commands: []*cli.Command{a.setRandom()},
		Action: func(ctx context.Context, c *cli.Command) error {
			themeName, err := resolveThemeArg(c.StringArg("theme"), a.store)
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			if err != nil {
				return err
			}

			slog.Debug("resolving theme", "theme", themeName)
			res, err := a.store.Resolve(themeName)
			if err != nil {
				return err
			}
			return applyTheme(ctx, res, a)
		},
		ShellComplete: func(ctx context.Context, c *cli.Command) {
			if c.Args().Len() > 0 {
				return // theme already typed, don't re-suggest
			}
			all, err := a.store.IDs()
			if err != nil {
				return
			}
			for _, t := range all {
				fmt.Fprintln(c.Root().Writer, t)
			}
		},
	}
}

func (a app) setRandom() *cli.Command {
	return &cli.Command{
		Name:  "random",
		Usage: "sets a random theme",
		Flags: appearanceFlags(),
		Action: func(ctx context.Context, c *cli.Command) error {
			appearance, err := appearanceFromFlags(c)
			if err != nil {
				return err
			}

			resolved, err := a.store.PickRandom(appearance)
			if err != nil {
				return err
			}
			return applyTheme(ctx, resolved, a)
		},
	}
}

func applyTheme(ctx context.Context, resolvedTheme theme.Resolved, app app) error {
	slog.Debug("materializing theme", "theme", resolvedTheme.ID(), "dir", app.cfg.CurrentDir())
	err := ui.Spin("Applying theme", func() error {
		if err := app.store.Materialize(ctx, resolvedTheme.ID(), app.cfg.CurrentDir()); err != nil {
			return err
		}
		if err := store.WriteCurrent(app.cfg.CurrentFile(), resolvedTheme.ID()); err != nil {
			return err
		}

		return integration.ApplyAll(app.integrations, resolvedTheme)
	})

	if err != nil {
		return err
	}
	slog.Info("theme set", "theme", resolvedTheme.ID())
	return nil
}

func pickTheme(store *store.Store) (string, error) {
	all, err := store.IDs()
	if err != nil {
		return "", err
	}
	return ui.Select("Themes", all)
}

func resolveThemeArg(arg string, store *store.Store) (string, error) {
	if arg != "" {
		return arg, nil
	}
	return pickTheme(store)
}
