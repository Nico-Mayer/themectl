package cli

import (
	"fmt"

	"github.com/Nico-Mayer/themectl-cli/internal/theme"
	"github.com/urfave/cli/v3"
)

func jsonFlag() *cli.BoolFlag {
	return &cli.BoolFlag{
		Name:  "json",
		Usage: "output in JSON format",
	}
}

func appearanceFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "light",
			Aliases: []string{"l"},
			Usage:   "only include light themes",
		},
		&cli.BoolFlag{
			Name:    "dark",
			Aliases: []string{"d"},
			Usage:   "only include dark themes",
		},
	}
}

func appearanceFromFlags(c *cli.Command) (theme.Appearance, error) {
	light, dark := c.Bool("light"), c.Bool("dark")
	switch {
	case light && dark:
		return "", fmt.Errorf("cannot use --light and --dark together")
	case light:
		return theme.Light, nil
	case dark:
		return theme.Dark, nil
	default:
		return theme.AnyAppearance, nil
	}
}
