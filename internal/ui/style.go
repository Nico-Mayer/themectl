package ui

import (
	"github.com/Nico-Mayer/themectl/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

var (
	accent  = lipgloss.Color("5")
	success = lipgloss.Color("2")
	warning = lipgloss.Color("3")
	danger  = lipgloss.Color("1")
	info    = lipgloss.Color("4")
)

var (
	Accent  = lipgloss.NewStyle().Bold(true).Foreground(accent)
	Muted   = lipgloss.NewStyle().Faint(true)
	Success = lipgloss.NewStyle().Foreground(success)
	Warning = lipgloss.NewStyle().Foreground(warning)
	Danger  = lipgloss.NewStyle().Foreground(danger)

	light = lipgloss.NewStyle().Bold(true).Foreground(warning)
	dark  = lipgloss.NewStyle().Bold(true).Foreground(info)
)

func Appearance(a theme.Appearance) lipgloss.Style {
	if a == theme.Light {
		return light
	}
	return dark
}
