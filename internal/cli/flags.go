package cli

import (
	"fmt"

	"github.com/Nico-Mayer/themectl/internal/theme"
	"github.com/urfave/cli/v3"
)

func jsonFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:  "json",
		Usage: "Print output as JSON",
	}
}

func appearanceFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "light",
			Aliases: []string{"l"},
			Usage:   "Include only light themes",
		},
		&cli.BoolFlag{
			Name:    "dark",
			Aliases: []string{"d"},
			Usage:   "Include only dark themes",
		},
	}
}

func appearanceFromFlags(c *cli.Command) (theme.Appearance, error) {
	light, dark := c.Bool("light"), c.Bool("dark")
	switch {
	case light && dark:
		return "", fmt.Errorf("--light and --dark cannot be used together")
	case light:
		return theme.Light, nil
	case dark:
		return theme.Dark, nil
	default:
		return theme.AnyAppearance, nil
	}
}
