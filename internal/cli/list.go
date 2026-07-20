package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/Nico-Mayer/themectl/internal/theme"
	"github.com/Nico-Mayer/themectl/internal/ui"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v3"
)

const listColGap = 4

func (a app) listCmd() *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List all available themes",
		Flags: append(
			appearanceFlags(),
			jsonFlag(),
		),
		Action: func(ctx context.Context, c *cli.Command) error {
			appearance, err := appearanceFromFlags(c)
			if err != nil {
				return err
			}

			all, err := a.store.List(appearance)
			if err != nil {
				return err
			}

			if c.Bool("json") {
				return printThemesJSON(all)
			}

			if !isatty.IsTerminal(os.Stdout.Fd()) {
				for _, t := range all {
					fmt.Println(t.ID())
				}
				return nil
			}

			curr, _ := store.ReadCurrent(a.cfg.CurrentFile())
			fmt.Println(renderThemeList(all, strings.TrimSpace(curr)))

			return nil
		},
	}
}

func printThemesJSON(themes []theme.Resolved) error {
	type themeJSON struct {
		ID         string           `json:"id"`
		Family     string           `json:"family"`
		Variant    string           `json:"variant"`
		Appearance theme.Appearance `json:"appearance"`
	}

	out := make([]themeJSON, 0, len(themes))
	for _, t := range themes {
		out = append(out, themeJSON{t.ID(), t.Family, t.Variant, t.Appearance})
	}

	return json.NewEncoder(os.Stdout).Encode(out)
}

func renderThemeList(themes []theme.Resolved, current string) string {
	width := len("Theme")
	for _, t := range themes {
		width = max(width, len(t.ID()))
	}
	width += listColGap

	lines := []string{ui.Muted.Render(fmt.Sprintf("  %-*s%s", width, "Theme", "Appearance"))}
	for _, t := range themes {
		id := fmt.Sprintf("  %-*s", width, t.ID())
		if t.ID() == current {
			id = ui.Accent.Render(fmt.Sprintf("● %-*s", width, t.ID()))
		}

		lines = append(lines, id+ui.Appearance(t.Appearance).Render(string(t.Appearance)))
	}

	return strings.Join(lines, "\n")
}
