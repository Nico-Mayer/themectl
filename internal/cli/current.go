package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/theme"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v3"
)

var (
	keyStyle    = lipgloss.NewStyle().Faint(true).Padding(0, 1)
	valueStyle  = lipgloss.NewStyle().Padding(0, 1)
	accentStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")).Padding(0, 1)
	lightStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3")).Padding(0, 1)
	darkStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")).Padding(0, 1)
)

func currentCmd(cfg config.Config, store *theme.Store) *cli.Command {
	return &cli.Command{
		Name:  "current",
		Usage: "get the current active theme",
		Flags: []cli.Flag{
			jsonFlag(),
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			curr, err := theme.ReadCurrent(cfg.CurrentFile())
			if err != nil {
				return err
			}

			if c.Bool("json") {
				return printCurrentJSON(curr, store)
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				fmt.Println(strings.TrimSpace(curr))
				return nil
			}

			resolved, err := store.Resolve(curr)
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

	appearanceStyle := darkStyle
	if r.Appearance == theme.Light {
		appearanceStyle = lightStyle
	}

	return table.New().
		Border(lipgloss.RoundedBorder()).
		BorderColumn(false).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == themesRow:
				return accentStyle
			case row == 1 && col == 1:
				return appearanceStyle
			case col == 0:
				return keyStyle
			default:
				return valueStyle
			}
		}).
		Render()
}

func printCurrentJSON(current string, store *theme.Store) error {
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
