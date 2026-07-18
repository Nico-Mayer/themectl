package cli

import (
	"context"

	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/engine"
	"github.com/nico-mayer/themectl-cli/internal/theme"
	"github.com/urfave/cli/v3"
)

func refreshCmd(cfg config.Config, store *theme.Store, eng *engine.Engine) *cli.Command {
	return &cli.Command{
		Name:    "refresh",
		Aliases: []string{"reapply"},
		Usage:   "reapply all integrations for current theme",
		Action: func(ctx context.Context, c *cli.Command) error {
			curr, err := theme.ReadCurrent(cfg.CurrentFile())
			if err != nil {
				return err
			}

			res, err := store.Resolve(curr)
			if err != nil {
				return err
			}
			return applyTheme(res, cfg, store, eng)
		},
	}
}
