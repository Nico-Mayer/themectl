package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/Nico-Mayer/themectl/internal/ui"
	"github.com/urfave/cli/v3"
)

func (a app) installCmd() *cli.Command {
	return &cli.Command{
		Name:      "install",
		Usage:     "Install a theme family from a git repository",
		ArgsUsage: "<git-url>",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "url", UsageText: "git URL of the theme repo"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Usage: "family name to install as (default: repo name)"},
			&cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "replace an already installed family"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			url := c.StringArg("url")
			if url == "" {
				return fmt.Errorf("no git URL provided")
			}
			var family string
			err := ui.Spin("Installing theme", func() (err error) {
				family, err = store.Install(a.cfg.ThemesDir(), url, c.String("name"), c.Bool("force"))
				return err
			})
			if err != nil {
				return err
			}
			slog.Info("theme installed", "family", family)
			return nil
		},
	}
}
