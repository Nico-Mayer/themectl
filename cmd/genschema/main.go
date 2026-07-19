//go:generate go run .
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/integration"
	"github.com/Nico-Mayer/themectl/internal/theme"
	"github.com/invopop/jsonschema"
)

const idBase = "https://raw.githubusercontent.com/Nico-Mayer/themectl-cli/main/schemas/"

type target struct {
	file        string
	title       string
	description string
	value       any
	post        func(*jsonschema.Schema)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "genschema:", err)
		os.Exit(1)
	}
}

func moduleRoot() (string, error) {
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		return "", fmt.Errorf("locate go.mod: %w", err)
	}
	gomod := strings.TrimSpace(string(out))
	if gomod == "" || gomod == os.DevNull {
		return "", fmt.Errorf("not inside a Go module")
	}
	return filepath.Dir(gomod), nil
}

func run() error {
	root, err := moduleRoot()
	if err != nil {
		return err
	}

	r := &jsonschema.Reflector{
		FieldNameTag:   "toml",
		DoNotReference: true, // inline nested types instead of $defs/$ref
	}

	targets := []target{
		{
			file:        "settings.schema.json",
			title:       "themectl.toml",
			description: "themectl settings. Values set here override the built-in defaults.",
			value:       &config.Settings{},
			post:        injectIntegrationNames,
		},
		{
			file:        "theme.schema.json",
			title:       "themectl theme.toml",
			description: "A theme family: defaults inherited by every variant, and the variants that override them.",
			value:       &theme.ThemeFile{},
		},
	}

	for _, t := range targets {
		s := r.Reflect(t.value)
		s.ID = jsonschema.ID(idBase + t.file)
		s.Title = t.title
		s.Description = t.description
		if t.post != nil {
			t.post(s)
		}

		data, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}
		data = append(data, '\n')

		path := filepath.Join(root, "schemas", t.file)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return err
		}
		fmt.Println("wrote", path)
	}

	return nil
}

func injectIntegrationNames(s *jsonschema.Schema) {
	names := integration.Names()
	enum := make([]any, len(names))
	for i, n := range names {
		enum[i] = n
	}

	prop, ok := s.Properties.Get("integrations")
	if !ok {
		panic("settings schema: integrations property missing")
	}
	prop.Items.Enum = enum
}
