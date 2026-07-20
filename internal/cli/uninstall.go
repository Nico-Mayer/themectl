package cli

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/Nico-Mayer/themectl/internal/ui"
	"github.com/charmbracelet/huh"
	"github.com/urfave/cli/v3"
)

func (a app) uninstallCmd() *cli.Command {
	return &cli.Command{
		Name:      "uninstall",
		Usage:     "Uninstall a theme family",
		ArgsUsage: "<name>",
		Arguments: []cli.Argument{
			&cli.StringArg{Name: "family", UsageText: "name of the theme family to uninstall"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			themeFamily, err := resolveThemeFamilyArg(c.StringArg("family"), a.store)
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			if err != nil {
				return err
			}

			if err := store.Uninstall(a.cfg.ThemesDir(), themeFamily); err != nil {
				return err
			}
			slog.Info("theme uninstalled", "family", themeFamily)
			return nil
		},
	}
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
		return "", err
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
	return ui.Select("Themes", families)
}
