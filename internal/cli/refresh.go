package cli

import (
	"context"

	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/urfave/cli/v3"
)

func (a app) refreshCmd() *cli.Command {
	return &cli.Command{
		Name:    "refresh",
		Aliases: []string{"reapply"},
		Usage:   "reapply all integrations for current theme",
		Action: func(ctx context.Context, c *cli.Command) error {
			curr, err := store.ReadCurrent(a.cfg.CurrentFile())
			if err != nil {
				return err
			}

			res, err := a.store.Resolve(curr)
			if err != nil {
				return err
			}
			return applyTheme(res, a)
		},
	}
}
