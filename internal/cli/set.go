package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Nico-Mayer/themectl-cli/internal/config"
	"github.com/Nico-Mayer/themectl-cli/internal/engine"
	"github.com/Nico-Mayer/themectl-cli/internal/theme"
	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

func setCmd(cfg config.Config, store *theme.Store, eng *engine.Engine) *cli.Command {
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
		Commands: []*cli.Command{setRandom(cfg, store, eng)},
		Action: func(ctx context.Context, c *cli.Command) error {
			themeName, err := resolveThemeArg(c.StringArg("theme"), store)
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			if err != nil {
				return err
			}

			slog.Debug("resolving theme", "theme", themeName)
			res, err := store.Resolve(themeName)
			if err != nil {
				return err
			}
			return applyTheme(res, cfg, store, eng)
		},
		ShellComplete: func(ctx context.Context, c *cli.Command) {
			if c.Args().Len() > 0 {
				return // theme already typed, don't re-suggest
			}
			all, err := store.IDs()
			if err != nil {
				return
			}
			for _, t := range all {
				fmt.Fprintln(c.Root().Writer, t)
			}
		},
	}
}

func resolveThemeArg(arg string, store *theme.Store) (string, error) {
	if arg != "" {
		return arg, nil
	}
	return pickTheme(store)
}

func setRandom(cfg config.Config, store *theme.Store, eng *engine.Engine) *cli.Command {
	return &cli.Command{
		Name:  "random",
		Usage: "sets a random theme",
		Flags: appearanceFlags(),
		Action: func(ctx context.Context, c *cli.Command) error {
			appearance, err := appearanceFromFlags(c)
			if err != nil {
				return err
			}

			resolved, err := store.PickRandom(appearance)
			if err != nil {
				return err
			}
			return applyTheme(resolved, cfg, store, eng)
		},
	}
}

func applyTheme(resolvedTheme theme.Resolved, cfg config.Config, store *theme.Store, eng *engine.Engine) error {
	slog.Debug("materializing theme", "theme", resolvedTheme.ID(), "dir", cfg.CurrentDir())
	if err := store.Materialize(resolvedTheme.ID(), cfg.CurrentDir()); err != nil {
		return err
	}
	if err := eng.Apply(resolvedTheme); err != nil {
		return err
	}
	if err := theme.WriteCurrent(cfg.CurrentFile(), resolvedTheme.ID()); err != nil {
		return err
	}
	slog.Info("theme set", "theme", resolvedTheme.ID())
	return nil
}

func pickTheme(store *theme.Store) (string, error) {
	all, err := store.IDs()
	if err != nil {
		return "", err
	}

	options := make([]huh.Option[string], len(all))
	for i, t := range all {
		options[i] = huh.NewOption(t, t)
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Themes").
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
