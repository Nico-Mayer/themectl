package cli

import (
	"context"
	"fmt"

	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/theme"
	urFaveCli "github.com/urfave/cli/v3"
)

func currentCmd(cfg config.Config) *urFaveCli.Command {
	return &urFaveCli.Command{
		Name:  "current",
		Usage: "get the current active theme",
		Action: func(ctx context.Context, c *urFaveCli.Command) error {
			curr, err := theme.ReadCurrent(cfg.CurrentFile())
			if err != nil {
				return err
			}

			fmt.Println(curr)

			return nil
		},
	}
}
