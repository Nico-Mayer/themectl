package cli

import (
	"context"
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
		Action: func(ctx context.Context, c *cli.Command) error {
			themeName := c.StringArg("theme")
			slog.Debug("resolving theme", "theme", themeName)
			res, err := store.Resolve(themeName)
			if err != nil {
				return err
			}
			slog.Debug("materializing theme", "theme", themeName, "dir", cfg.CurrentDir())
			if err := store.Materialize(themeName, cfg.CurrentDir()); err != nil {
				return err
			}
			if err := eng.Apply(res); err != nil {
				return err
			}
			if err := theme.WriteCurrent(cfg.CurrentFile(), res.ID()); err != nil {
				return err
			}
			slog.Info("theme set", "theme", res.ID())
			return nil
		},
	}
}
