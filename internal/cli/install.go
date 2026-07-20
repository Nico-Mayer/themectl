package cli

import (
	"context"

	"github.com/Nico-Mayer/themectl/internal/theme"
	"github.com/urfave/cli/v3"
)

func (a app) installCmd() *cli.Command {
	return &cli.Command{
		Name: "install",
		Arguments: []cli.Argument{
			&cli.StringArg{
				UsageText: "<url>",
				Name:      "url",
			},
			&cli.StringArg{
				UsageText: "<name>",
				Name:      "name",
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			url := c.StringArg("url")
			name := c.StringArg("name")
			force := c.Bool("force")

			_, err := theme.Install(a.cfg.ThemesDir(), url, name, force)

			return err
		},
	}
}
