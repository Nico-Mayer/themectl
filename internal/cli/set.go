package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/engine"
	"github.com/nico-mayer/themectl-cli/internal/theme"
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
			themeName := c.StringArg("theme")
			slog.Debug("resolving theme", "theme", themeName)
			res, err := store.Resolve(themeName)
			if err != nil {
				return err
			}
			return applyTheme(res, cfg, store, eng)
		},
	}
}

func setRandom(cfg config.Config, store *theme.Store, eng *engine.Engine) *cli.Command {
	return &cli.Command{
		Name:  "random",
		Usage: "sets a random theme",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "light",
				Aliases: []string{"l"},
				Usage:   "only use light themes",
			},
			&cli.BoolFlag{
				Name:    "dark",
				Aliases: []string{"d"},
				Usage:   "only use dark themes",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			light := c.Bool("light")
			dark := c.Bool("dark")

			if light && dark {
				return fmt.Errorf("cannot use --light and --dark together")
			}

			var appearance *theme.Appearance
			if light {
				appearance = new(theme.Light)
			} else if dark {
				appearance = new(theme.Dark)
			} else {
				appearance = nil
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
