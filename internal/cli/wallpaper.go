package cli

import (
	"context"
	"fmt"

	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/theme"
	"github.com/nico-mayer/themectl-cli/internal/wallpaper"
	urFaveCli "github.com/urfave/cli/v3"
)

func wallpaperCmd(cfg config.Config, store *theme.Store) *urFaveCli.Command {
	return &urFaveCli.Command{
		Name:    "wallpaper",
		Aliases: []string{"wall"},
		Usage:   "set wallpaper",
		Flags: []urFaveCli.Flag{
			&urFaveCli.BoolFlag{
				Name:    "random",
				Aliases: []string{"r"},
			},
		},
		Action: func(ctx context.Context, c *urFaveCli.Command) error {
			randomFlag := c.Bool("random")
			manager := wallpaper.NewManager(cfg.ThemesDir(), cfg.SharedWallpapersDir())
			if randomFlag {
				curr, err := theme.ReadCurrent(cfg.CurrentFile())
				if err != nil {
					return err
				}

				res, err := store.Resolve(curr)

				return manager.ApplyRandom(res)
			}

			curr, err := manager.Current()
			if err != nil {
				return err
			}

			fmt.Println(curr)

			return nil
		},
	}
}
