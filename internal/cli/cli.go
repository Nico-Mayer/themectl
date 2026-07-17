package cli

import (
	"context"
	"log/slog"
	"os"

	"github.com/charmbracelet/log"
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
				Usage:   "Prints more logs to stderr",
				Aliases: []string{"v"},
			},
		},
		Commands: []*urfaveCli.Command{
			listCmd(store),
			setCmd(cfg, store, engine),
			currentCmd(cfg),
			wallpaperCmd(cfg, store),
		},
		Before: func(ctx context.Context, c *urfaveCli.Command) (context.Context, error) {
			level := slog.LevelInfo
			withTime := false
			if c.Bool("verbose") {
				level = slog.LevelDebug
				withTime = true
			}
			handler := log.NewWithOptions(os.Stderr, log.Options{
				Level:           log.Level(level),
				ReportTimestamp: withTime,
				TimeFormat:      "15:04:05",
			})
			slog.SetDefault(slog.New(handler))

			return nil, nil
		},
	}
}
