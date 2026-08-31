package cli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDoctorReport_JSONContract(t *testing.T) {
	supported := false
	report := doctorReport{
		ConfigFile:        "/tmp/themectl.toml",
		ConfigFileExists:  true,
		Root:              "/tmp/themectl",
		Cache:             "/tmp/cache",
		CurrentTheme:      "catppuccin/mocha",
		CurrentThemeFound: true,
		InstalledThemes:   1,
		Integrations: []integrationStatus{{
			Name:      "ghostty",
			Enabled:   true,
			Healthy:   true,
			Detail:    "ready",
			Supported: &supported,
		}},
		Unknown: []string{"custom"},
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	wantKeys := []string{
		"cache",
		"config_file",
		"config_file_exists",
		"current_theme",
		"current_theme_found",
		"installed_themes",
		"integrations",
		"root",
		"unknown_integrations",
	}
	for _, key := range wantKeys {
		if _, ok := got[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
	if len(got) != len(wantKeys) {
		t.Errorf("JSON has %d keys, want %d: %s", len(got), len(wantKeys), data)
	}

	integrations := got["integrations"].([]any)
	integration := integrations[0].(map[string]any)
	for _, key := range []string{"name", "enabled", "healthy", "detail", "supported_by_current_theme"} {
		if _, ok := integration[key]; !ok {
			t.Errorf("integration missing JSON key %q", key)
		}
	}
}

func TestThemeRows_IncludeRecoveryAndMissingStates(t *testing.T) {
	tests := []struct {
		name   string
		report doctorReport
		want   []string
	}{
		{
			name:   "unset",
			report: doctorReport{InstalledThemes: 1},
			want:   []string{"Not set", "`themectl set`"},
		},
		{
			name:   "missing",
			report: doctorReport{CurrentTheme: "custom/dark", InstalledThemes: 1},
			want:   []string{"custom/dark", "Not found in themes directory"},
		},
		{
			name:   "no installed themes",
			report: doctorReport{},
			want:   []string{"`themectl install <git-url>`"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := themeRows(tt.report)
			var out strings.Builder
			for _, row := range rows {
				out.WriteString(row.value)
			}
			for _, text := range tt.want {
				if !strings.Contains(out.String(), text) {
					t.Errorf("rows %q missing %q", out.String(), text)
				}
			}
		})
	}
}

func TestRenderDoctorReport_StatesHaveTextLabels(t *testing.T) {
	unsupported := false
	report := doctorReport{
		Root:            "/tmp/themectl",
		Cache:           "/tmp/cache",
		InstalledThemes: 1,
		Integrations: []integrationStatus{
			{Name: "available", Healthy: true},
			{Name: "unhealthy", Enabled: true, Detail: "config missing"},
			{Name: "unsupported", Enabled: true, Healthy: true, Supported: &unsupported},
		},
		Unknown: []string{"unknown"},
	}

	out := renderIntegrations(report)
	for _, text := range []string{"available", "unhealthy: config missing", "enabled; unused by current theme", "unknown; enabled but not registered"} {
		if !strings.Contains(out, text) {
			t.Errorf("output missing text state %q:\n%s", text, out)
		}
	}
}
