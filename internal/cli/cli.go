package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/engine"
	"github.com/nico-mayer/themectl-cli/internal/theme"
	urfaveCli "github.com/urfave/cli/v3"
)

func New(cfg config.Config, store *theme.Store, engine *engine.Engine) *urfaveCli.Command {
	return &urfaveCli.Command{
		Name:  "themectl",
		Usage: "Manage and apply themes across your tools",
		Flags: []urfaveCli.Flag{
			&urfaveCli.BoolFlag{
				Name:    "verbose",
				Usage:   "Prints more logs to stder",
				Aliases: []string{"v"},
			},
		},
		Commands: []*urfaveCli.Command{
			listCmd(store),
			setCmd(cfg, store, engine),
		},
		Before: func(ctx context.Context, c *urfaveCli.Command) (context.Context, error) {
			level := slog.LevelInfo
			if c.Bool("verbose") {
				level = slog.LevelDebug
			}
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))
			return nil, nil
		},
	}
}
