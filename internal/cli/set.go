package cli

import (
	"context"

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
			res, err := store.Resolve(themeName)
			if err != nil {
				return err
			}
			if err := store.Materialize(themeName, cfg.CurrentDir()); err != nil {
				return err
			}
			if err := eng.Apply(res); err != nil {
				return err
			}
			return theme.WriteCurrent(cfg.CurrentFile(), res.ID())
		},
	}
}
