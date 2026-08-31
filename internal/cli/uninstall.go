package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"charm.land/huh/v2"
	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/Nico-Mayer/themectl/internal/ui"
	"github.com/urfave/cli/v3"
)

func (a app) uninstallCmd() *cli.Command {
	return &cli.Command{
		Name:      "uninstall",
		Usage:     "Uninstall a theme family",
		ArgsUsage: "<name>",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "family", UsageText: "Theme family to uninstall"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			familyArg := c.StringArg("family")
			themeFamily, err := resolveThemeFamilyArg(familyArg, a.store)
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("select theme family: %w", err)
			}

			if familyArg == "" {
				confirmed, err := ui.Confirm(uninstallConfirmation(themeFamily))
				if errors.Is(err, huh.ErrUserAborted) {
					return nil
				}
				if err != nil {
					return fmt.Errorf("confirm uninstall of theme family %q: %w", themeFamily, err)
				}
				if !confirmed {
					slog.Info("theme family kept", "family", themeFamily, "reason", "uninstall canceled")
					return nil
				}
			}

			if err := store.Uninstall(a.cfg.ThemesDir(), themeFamily); err != nil {
				return fmt.Errorf("uninstall theme family %q: %w", themeFamily, err)
			}
			slog.Info("theme family uninstalled", "family", themeFamily)
			return nil
		},
	}
}

func uninstallConfirmation(name string) string {
	return fmt.Sprintf("Uninstall theme family %q? This deletes its local files and cannot be undone.", name)
}

func resolveThemeFamilyArg(arg string, store *store.Store) (string, error) {
	if arg != "" {
		return arg, nil
	}
	return pickThemeFamily(store)
}

func pickThemeFamily(store *store.Store) (string, error) {
	all, err := store.IDs()
	if err != nil {
		return "", fmt.Errorf("list theme families: %w", err)
	}

	var families []string
	seen := map[string]bool{}
	for _, t := range all {
		family, _, ok := strings.Cut(t, "/")
		if !ok || seen[family] {
			continue
		}
		seen[family] = true
		families = append(families, family)
	}
	return ui.Select("Select a theme family", families)
}
