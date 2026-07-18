package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/theme"
	"github.com/urfave/cli/v3"
)

const listColGap = 4

var (
	activeStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	listHeaderStyle = lipgloss.NewStyle().Faint(true)
)

func listCmd(cfg config.Config, store *theme.Store) *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls"},
		Usage:   "List all available themes",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "light",
				Aliases: []string{"l"},
				Usage:   "only list light themes",
			},
			&cli.BoolFlag{
				Name:    "dark",
				Aliases: []string{"d"},
				Usage:   "only list dark themes",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "output themes in JSON format",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			light := c.Bool("light")
			dark := c.Bool("dark")

			if light && dark {
				return fmt.Errorf("cannot use --light and --dark together")
			}

			var all []theme.Resolved
			var err error
			switch {
			case light:
				all, err = store.ListAllByAppearance(theme.Light)
			case dark:
				all, err = store.ListAllByAppearance(theme.Dark)
			default:
				all, err = store.ListAllResolved()
			}
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

			curr, _ := theme.ReadCurrent(cfg.CurrentFile())
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

	lines := []string{listHeaderStyle.Render(fmt.Sprintf("  %-*s%s", width, "Theme", "Appearance"))}
	for _, t := range themes {
		appearanceStyle := darkStyle
		if t.Appearance == theme.Light {
			appearanceStyle = lightStyle
		}

		id := fmt.Sprintf("  %-*s", width, t.ID())
		if t.ID() == current {
			id = activeStyle.Render(fmt.Sprintf("● %-*s", width, t.ID()))
		}

		lines = append(lines, id+appearanceStyle.UnsetPadding().Render(string(t.Appearance)))
	}

	return strings.Join(lines, "\n")
}
