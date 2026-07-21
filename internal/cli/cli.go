package cli

import (
	"context"
	"log/slog"
	"os"

	"charm.land/log/v2"
	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/integration"
	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/Nico-Mayer/themectl/internal/ui"
	"github.com/charmbracelet/colorprofile"
	urfaveCli "github.com/urfave/cli/v3"
)

type app struct {
	cfg          config.Config
	store        *store.Store
	integrations []integration.Integration
}

func New(cfg config.Config, store *store.Store, integrations []integration.Integration) *urfaveCli.Command {

	app := app{
		cfg:          cfg,
		store:        store,
		integrations: integrations,
	}

	return &urfaveCli.Command{
		Name:                  "themectl",
		Usage:                 "Manage and apply themes across your tools",
		EnableShellCompletion: true,
		Flags: []urfaveCli.Flag{
			&urfaveCli.BoolFlag{
				Name:    "verbose",
				Usage:   "Prints more logs to stderr",
				Aliases: []string{"v"},
			},
		},
		Commands: []*urfaveCli.Command{
			app.listCmd(),
			app.setCmd(),
			app.currentCmd(),
			app.wallpaperCmd(),
			app.refreshCmd(),
			app.doctorCmd(),
			app.installCmd(),
			app.uninstallCmd(),
			app.updateCmd(),
		},
		Before: func(ctx context.Context, c *urfaveCli.Command) (context.Context, error) {
			level := slog.LevelInfo
			withTime := false
			if c.Bool("verbose") {
				level = slog.LevelDebug
				withTime = true
			}
			handler := log.NewWithOptions(ui.Console, log.Options{
				Level:           log.Level(level),
				ReportTimestamp: withTime,
				TimeFormat:      "15:04:05",
			})
			handler.SetColorProfile(colorprofile.Detect(os.Stderr, os.Environ()))
			slog.SetDefault(slog.New(handler))

			return nil, nil
		},
	}
}
