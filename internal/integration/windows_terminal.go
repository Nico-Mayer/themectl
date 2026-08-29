package integration

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

// themectlName is what the scheme and the chrome are installed as, whatever the
// theme's assets call themselves. Fixed, so the two settings keys never change.
const themectlName = "themectl"

type WindowsTerminal struct {
	ConfigFile   string
	FragmentFile string
	SchemeFile   string
	ThemeFile    string
}

func (WindowsTerminal) Name() string { return "windows-terminal" }

// Supports keys off the color scheme alone: the chrome is optional.
func (w WindowsTerminal) Supports(theme.Resolved) bool {
	_, err := os.Stat(w.SchemeFile)
	return err == nil
}

func (w WindowsTerminal) Check() error {
	return checkConfigDir(w.Name(), filepath.Dir(w.ConfigFile))
}

func (w WindowsTerminal) Apply(theme.Resolved) error {
	// the fragment goes first: writing settings.json is what triggers Windows
	// Terminal's live reload, and that reload re-reads fragments
	if err := w.writeFragment(); err != nil {
		return err
	}

	data, err := os.ReadFile(w.ConfigFile)
	if err != nil {
		return fmt.Errorf("read windows terminal settings: %w", err)
	}

	updated, err := setJSONPath(string(data), []string{"profiles", "defaults", "colorScheme"}, themectlName)
	if err != nil {
		return err
	}

	updated, err = w.applyChrome(updated)
	if err != nil {
		return err
	}

	return w.writeConfig(updated)
}

func (w WindowsTerminal) Reset() error {
	slog.Debug("resetting windows terminal theme", "config", w.ConfigFile)

	if err := os.Remove(w.FragmentFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove windows terminal fragment: %w", err)
	}

	data, err := os.ReadFile(w.ConfigFile)
	if err != nil {
		return fmt.Errorf("read windows terminal settings: %w", err)
	}

	updated, err := deleteJSONPath(string(data), []string{"profiles", "defaults", "colorScheme"})
	if err != nil {
		return err
	}

	updated, err = w.clearChrome(updated)
	if err != nil {
		return err
	}

	return w.writeConfig(updated)
}

// applyChrome installs the theme's chrome, or clears themectl's when the theme
// ships none, so the previous theme's title bar does not linger.
func (w WindowsTerminal) applyChrome(config string) (string, error) {
	chrome, err := os.ReadFile(w.ThemeFile)
	if os.IsNotExist(err) {
		return w.clearChrome(config)
	}
	if err != nil {
		return "", fmt.Errorf("read windows terminal theme asset: %w", err)
	}

	renamed, err := renameJSONAsset(chrome)
	if err != nil {
		return "", fmt.Errorf("windows terminal theme asset: %w", err)
	}

	updated, err := upsertJSONArrayByName(config, []string{"themes"}, themectlName, renamed)
	if err != nil {
		return "", err
	}
	return setJSONPath(updated, []string{"theme"}, themectlName)
}

func (w WindowsTerminal) clearChrome(config string) (string, error) {
	updated, err := removeJSONArrayByName(config, []string{"themes"}, themectlName)
	if err != nil {
		return "", err
	}
	return deleteJSONPath(updated, []string{"theme"})
}

func (w WindowsTerminal) writeFragment() error {
	scheme, err := os.ReadFile(w.SchemeFile)
	if err != nil {
		return fmt.Errorf("read windows terminal scheme asset: %w", err)
	}

	renamed, err := renameJSONAsset(scheme)
	if err != nil {
		return fmt.Errorf("windows terminal scheme asset: %w", err)
	}

	fragment, err := json.MarshalIndent(map[string]any{
		"schemes": []json.RawMessage{renamed},
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("build windows terminal fragment: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(w.FragmentFile), 0o755); err != nil {
		return fmt.Errorf("create windows terminal fragment dir: %w", err)
	}
	if err := os.WriteFile(w.FragmentFile, fragment, 0o644); err != nil {
		return fmt.Errorf("write windows terminal fragment: %w", err)
	}
	return nil
}

func (w WindowsTerminal) writeConfig(content string) error {
	if err := os.WriteFile(w.ConfigFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write windows terminal settings: %w", err)
	}
	return nil
}

// Windows Terminal hardcodes its settings path with no environment override,
// so themectl defaults to the Store install; unpackaged, Preview and portable
// installs point config_file at their own copy.
func windowsTerminalSettingsFile() string {
	return filepath.Join(localAppData(),
		"Packages", "Microsoft.WindowsTerminal_8wekyb3d8bbwe", "LocalState", "settings.json")
}

// Fragments sit at the same user-level path for packaged and unpackaged
// installs, so this one needs no override.
func windowsTerminalFragmentFile() string {
	return filepath.Join(localAppData(),
		"Microsoft", "Windows Terminal", "Fragments", "themectl", "scheme.json")
}

func localAppData() string {
	if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
		return dir
	}
	return localConfigDir()
}

// renameJSONAsset rewrites an asset's "name" to themectl, so a theme author
// never has to make a name in theme.toml agree with one inside an asset.
func renameJSONAsset(asset []byte) ([]byte, error) {
	out, err := setJSONPath(string(asset), []string{"name"}, themectlName)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

func newWindowsTerminal(cfg config.Config) Integration {
	return WindowsTerminal{
		ConfigFile:   cfg.Settings.WindowsTerminal.ConfigFileOr(windowsTerminalSettingsFile()),
		FragmentFile: windowsTerminalFragmentFile(),
		SchemeFile:   filepath.Join(cfg.CurrentDir(), theme.WindowsTerminalAssetName),
		ThemeFile:    filepath.Join(cfg.CurrentDir(), theme.WindowsTerminalThemeAssetName),
	}
}
