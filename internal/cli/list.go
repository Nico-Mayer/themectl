package cli

import (
	"context"
	"fmt"

	"github.com/nico-mayer/themectl-cli/internal/theme"
	"github.com/urfave/cli/v3"
)

func listCmd(store *theme.Store) *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List all available themes",
		Action: func(ctx context.Context, c *cli.Command) error {
			all, _ := store.ListAll()

			for _, t := range all {
				fmt.Println(t)
			}

			return nil
		},
	}
}
