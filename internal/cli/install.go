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
		Usage:     "Install a theme family from a Git repository",
		ArgsUsage: "<git-url>",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "url", UsageText: "Git repository URL"},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Usage: "Name for installed theme family (default: repository name)"},
			&cli.BoolFlag{Name: "force", Aliases: []string{"f"}, Usage: "Replace an installed theme family"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			url := c.StringArg("url")
			if url == "" {
				return fmt.Errorf("git repository URL is required")
			}
			var family string
			err := ui.Spin("Installing theme family", func() (err error) {
				family, err = store.Install(a.cfg.ThemesDir(), url, c.String("name"), c.Bool("force"))
				return err
			})
			if err != nil {
				return fmt.Errorf("install theme family: %w", err)
			}
			slog.Info("theme family installed", "family", family)
			return nil
		},
	}
}
