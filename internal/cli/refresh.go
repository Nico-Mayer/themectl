package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
)

func (a app) refreshCmd() *cli.Command {
	return &cli.Command{
		Name:    "refresh",
		Aliases: []string{"reapply"},
		Usage:   "Reapply the current theme to all integrations",
		Action: func(ctx context.Context, c *cli.Command) error {
			curr, err := readCurrentTheme(a.cfg.CurrentFile())
			if err != nil {
				return err
			}

			res, err := a.store.Resolve(curr)
			if err != nil {
				return fmt.Errorf("resolve current theme %q: %w; run `themectl set` to select an installed theme", strings.TrimSpace(curr), err)
			}
			return applyTheme(ctx, res, a)
		},
	}
}
