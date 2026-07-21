package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/integration"
	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/Nico-Mayer/themectl/internal/ui"
	"github.com/urfave/cli/v3"
)

type doctorReport struct {
	ConfigFile        string              `json:"config_file"`
	ConfigFileExists  bool                `json:"config_file_exists"`
	Root              string              `json:"root"`
	Cache             string              `json:"cache"`
	CurrentTheme      string              `json:"current_theme"`
	CurrentThemeFound bool                `json:"current_theme_found"`
	InstalledThemes   int                 `json:"installed_themes"`
	Integrations      []integrationStatus `json:"integrations"`
	Unknown           []string            `json:"unknown_integrations,omitempty"`
}

type integrationStatus struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Healthy bool   `json:"healthy"`
	Detail  string `json:"detail,omitempty"`
}

func (a app) doctorCmd() *cli.Command {
	return &cli.Command{
		Name:    "doctor",
		Aliases: []string{"status"},
		Usage:   "report current theme, settings, and integration status",
		Flags:   []cli.Flag{jsonFlag()},
		Action: func(ctx context.Context, c *cli.Command) error {
			report := buildDoctorReport(a.cfg, a.store)

			if c.Bool("json") {
				return json.NewEncoder(os.Stdout).Encode(report)
			}

			return renderDoctorReport(report)
		},
	}
}

func buildDoctorReport(cfg config.Config, st *store.Store) doctorReport {
	r := doctorReport{
		ConfigFile:   cfg.SettingsFile(),
		Root:         cfg.Root,
		Cache:        cfg.CacheDir(),
		Integrations: integrationStatuses(cfg),
		Unknown:      integration.Unknown(cfg),
	}

	curr, err := store.ReadCurrent(cfg.CurrentFile())
	switch {
	case err == nil:
		r.CurrentTheme = strings.TrimSpace(curr)
	case !errors.Is(err, os.ErrNotExist):
		r.CurrentTheme = "unreadable: " + err.Error()
	}

	_, err = os.Stat(cfg.SettingsFile())
	r.ConfigFileExists = err == nil

	if ids, err := st.IDs(); err == nil {
		r.InstalledThemes = len(ids)
	}
	r.CurrentThemeFound = themeFound(st, r.CurrentTheme)

	return r
}

func integrationStatuses(cfg config.Config) []integrationStatus {
	enabled := make(map[string]integration.Integration)
	for _, i := range integration.Enabled(cfg) {
		enabled[i.Name()] = i
	}

	names := integration.Names()
	statuses := make([]integrationStatus, 0, len(names))
	for _, name := range names {
		s := integrationStatus{Name: name, Healthy: true}
		if i, ok := enabled[name]; ok {
			s.Enabled = true
			if err := i.Check(); err != nil {
				s.Healthy = false
				s.Detail = err.Error()
			}
		}
		statuses = append(statuses, s)
	}
	return statuses
}

func themeFound(store *store.Store, id string) bool {
	if id == "" {
		return false
	}
	_, err := store.Resolve(id)
	return err == nil
}

type kvRow struct {
	key   string
	value string
}

func renderDoctorReport(r doctorReport) error {
	sections := []string{
		renderSection("Settings", renderKVRows(settingsRows(r))),
		renderSection("Theme", renderKVRows(themeRows(r))),
		renderSection("Integrations", renderIntegrations(r)),
	}
	fmt.Println(strings.Join(sections, "\n\n"))
	return nil
}

func renderSection(title, body string) string {
	return ui.Accent.Render(title) + "\n" + body
}

func renderKVRows(rows []kvRow) string {
	width := 0
	for _, row := range rows {
		width = max(width, len(row.key))
	}
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		key := ui.Muted.Render(fmt.Sprintf("%-*s", width, row.key))
		lines = append(lines, fmt.Sprintf("  %s  %s", key, row.value))
	}
	return strings.Join(lines, "\n")
}

func settingsRows(r doctorReport) []kvRow {
	configFile := r.ConfigFile
	if !r.ConfigFileExists {
		configFile = ui.Muted.Render("(none - using built-in defaults)")
	}
	rows := []kvRow{
		{"Root", r.Root},
		{"Config", configFile},
		{"Cache", r.Cache},
	}
	if !r.ConfigFileExists {
		rows = append(rows, kvRow{"", ui.Muted.Render("create " + r.ConfigFile + " to customize")})
	}
	return rows
}

func themeRows(r doctorReport) []kvRow {
	currentTheme := r.CurrentTheme
	switch {
	case currentTheme == "":
		currentTheme = ui.Muted.Render("(none set - run `themectl set`)")
	case !r.CurrentThemeFound:
		currentTheme += "  " + ui.Danger.Render("(not found in themes dir)")
	}
	installed := fmt.Sprintf("%d", r.InstalledThemes)
	if r.InstalledThemes == 0 {
		installed = ui.Danger.Render("0 - add themes under " + r.Root + "/themes")
	}
	return []kvRow{
		{"Current", currentTheme},
		{"Installed", installed},
	}
}

func renderIntegrations(r doctorReport) string {
	if len(r.Integrations) == 0 && len(r.Unknown) == 0 {
		return ui.Muted.Render("  (none)")
	}

	width := 0
	for _, s := range r.Integrations {
		width = max(width, len(s.Name))
	}
	for _, name := range r.Unknown {
		width = max(width, len(name))
	}

	lines := make([]string, 0, len(r.Integrations)+len(r.Unknown))
	for _, s := range r.Integrations {
		switch {
		case !s.Enabled:
			lines = append(lines, integrationLine(ui.Muted, "○", s.Name, "available", width))
		case !s.Healthy:
			lines = append(lines, integrationLine(ui.Warning, "!", s.Name, s.Detail, width))
		default:
			lines = append(lines, integrationLine(ui.Success, "●", s.Name, "enabled", width))
		}
	}
	for _, name := range r.Unknown {
		lines = append(lines, integrationLine(ui.Danger, "✗", name, "unknown - enabled but not registered", width))
	}
	return strings.Join(lines, "\n")
}

func integrationLine(style lipgloss.Style, marker, name, status string, width int) string {
	return fmt.Sprintf("  %s %-*s  %s", style.Render(marker), width, name, style.Render(status))
}
