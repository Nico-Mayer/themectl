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
		Usage: "Update theme families installed from Git",
		Action: func(ctx context.Context, c *cli.Command) error {
			var aborted bool

			confirm := func(name string) (bool, error) {
				if aborted || !isatty.IsTerminal(os.Stderr.Fd()) {
					return false, nil
				}
				message := updateConfirmation(name)
				ok, err := ui.Confirm(message)
				if errors.Is(err, huh.ErrUserAborted) {
					aborted = true
					return false, nil
				}
				if err != nil {
					return false, err
				}
				return ok, nil
			}

			results, err := store.Update(a.cfg.ThemesDir(), confirm)
			if err != nil {
				return fmt.Errorf("update theme families: %w", err)
			}

			for _, result := range results {
				logUpdateResult(result)
			}

			return nil
		},
	}
}

func updateConfirmation(name string) string {
	return fmt.Sprintf("Theme family %q has local changes. Updating may discard them. Continue?", name)
}

func logUpdateResult(result store.UpdateResult) {
	switch result.Status {
	case store.UpdateUpdated:
		slog.Info("theme family updated", "family", result.Name)
	case store.UpdateDeclined:
		if result.Err != nil {
			slog.Warn("theme family skipped", "family", result.Name, "reason", "confirmation failed", "err", result.Err)
			return
		}
		slog.Info("theme family skipped", "family", result.Name, "reason", "local changes kept")
	case store.UpdateSkipped:
		slog.Info("theme family skipped", "family", result.Name, "reason", "not installed from Git")
	case store.UpdateFailed:
		slog.Warn("theme family update failed", "family", result.Name, "err", result.Err)
	}
}
