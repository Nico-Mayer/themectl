package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"charm.land/huh/v2"
	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/Nico-Mayer/themectl/internal/ui"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v3"
)

func (a app) updateCmd() *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Update alll themes installed via git",
		Action: func(ctx context.Context, c *cli.Command) error {
			var aborted bool

			confirm := func(name string) bool {
				if aborted || !isatty.IsTerminal(os.Stderr.Fd()) {
					return false
				}
				message := fmt.Sprintf("%q has local changes. Update anyway?", name)
				ok, err := ui.Confirm(message)
				if errors.Is(err, huh.ErrUserAborted) {
					aborted = true
					return false
				}
				return err == nil && ok
			}

			results, err := store.Update(a.cfg.ThemesDir(), confirm)
			if err != nil {
				return err
			}

			for _, r := range results {
				switch r.Status {
				case store.UpdateUpdated:
					slog.Info("theme updated", "family", r.Name)
				case store.UpdateDeclined:
					slog.Info("theme skipped, local changes kept", "family", r.Name)
				case store.UpdateSkipped:
					slog.Debug("not a git repo, skipped", "family", r.Name)
				case store.UpdateFailed:
					slog.Warn("theme update failed", "family", r.Name, "err", r.Err)
				}
			}

			return nil
		},
	}
}
