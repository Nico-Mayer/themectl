package integration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/BurntSushi/toml"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

type Zed struct {
	SettingsPath string
	CurrentDir   string
	Installer    ExtensionInstaller
}

type ExtensionInstaller interface {
	Ensure(ref ExtensionRef) error
}

type ExtensionRef struct {
	URL string
}

type zedSidecar struct {
	ExtensionURL string `toml:"extension_url"`
}

var zedThemeLine = regexp.MustCompile(`("theme"\s*:\s*)"[^"]*"`)

func (Zed) Name() string {
	return "zed"
}

func (z Zed) Apply(t theme.Resolved) error {
	if t.Zed == nil || t.Zed.Theme == "" {
		return fmt.Errorf("theme %s has no zed override", t.ID())
	}
	themeName := t.Zed.Theme

	if err := z.ensureExtension(); err != nil {
		return err
	}

	data, err := os.ReadFile(z.SettingsPath)
	if err != nil {
		return fmt.Errorf("read zed settings: %w", err)
	}

	updated, err := setZedTheme(string(data), themeName)
	if err != nil {
		return err
	}

	if err := os.WriteFile(z.SettingsPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write zed settings: %w", err)
	}

	return nil
}

func (z Zed) Check() error {
	return checkConfigDir(z.Name(), z.SettingsPath)
}

func (z Zed) ensureExtension() error {
	if z.Installer == nil {
		return nil
	}

	data, err := os.ReadFile(filepath.Join(z.CurrentDir, "zed.toml"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read zed sidecar: %w", err)
	}
	sidecar, err := parseZedSidecar(data)
	if err != nil {
		return err
	}
	if sidecar.ExtensionURL == "" {
		return nil
	}

	return z.Installer.Ensure(ExtensionRef{URL: sidecar.ExtensionURL})
}

func setZedTheme(config string, themeName string) (string, error) {
	if !zedThemeLine.MatchString(config) {
		return "", errors.New("no `\"theme\"` key found in zed config")
	}
	return zedThemeLine.ReplaceAllString(config, `${1}"`+themeName+`"`), nil
}

func parseZedSidecar(data []byte) (zedSidecar, error) {
	var s zedSidecar
	if err := toml.Unmarshal(data, &s); err != nil {
		return zedSidecar{}, fmt.Errorf("parse zed sidecar: %w", err)
	}
	return s, nil
}
