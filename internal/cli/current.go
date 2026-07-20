package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/Nico-Mayer/themectl/internal/theme"
	"github.com/Nico-Mayer/themectl/internal/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v3"
)

func (a app) currentCmd() *cli.Command {
	return &cli.Command{
		Name:  "current",
		Usage: "get the current active theme",
		Flags: []cli.Flag{
			jsonFlag(),
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			curr, err := store.ReadCurrent(a.cfg.CurrentFile())
			if err != nil {
				return err
			}

			if c.Bool("json") {
				return printCurrentJSON(curr, a.store)
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				fmt.Println(strings.TrimSpace(curr))
				return nil
			}

			resolved, err := a.store.Resolve(curr)
			if err != nil {
				return err
			}

			fmt.Println(renderThemeDetails(resolved))

			return nil
		},
	}
}

func renderThemeDetails(r theme.Resolved) string {
	rows := [][]string{
		{"Theme", r.ID()},
		{"Appearance", string(r.Appearance)},
		{"Family", r.Family},
		{"Variant", r.Variant},
		{"Wallpapers", strings.Join(r.WallpaperSources, "\n")},
	}

	themes := r.Themes()
	themesRow := -1
	if len(themes) > 0 {
		rows = append(rows, []string{})
		themesRow = len(rows)
		rows = append(rows, []string{"Themes:", ""})
		for _, k := range slices.Sorted(maps.Keys(themes)) {
			rows = append(rows, []string{k, themes[k]})
		}
	}

	cell := lipgloss.NewStyle().Padding(0, 1)
	appearanceStyle := ui.Appearance(r.Appearance).Padding(0, 1)

	return table.New().
		Border(lipgloss.RoundedBorder()).
		BorderColumn(false).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == themesRow:
				return ui.Accent.Padding(0, 1)
			case row == 1 && col == 1:
				return appearanceStyle
			case col == 0:
				return ui.Muted.Padding(0, 1)
			default:
				return cell
			}
		}).
		Render()
}

func printCurrentJSON(current string, store *store.Store) error {
	type themeJSON struct {
		ID         string           `json:"id"`
		Family     string           `json:"family"`
		Variant    string           `json:"variant"`
		Appearance theme.Appearance `json:"appearance"`
	}

	resolved, err := store.Resolve(current)
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(themeJSON{resolved.ID(), resolved.Family, resolved.Variant, resolved.Appearance})
}
