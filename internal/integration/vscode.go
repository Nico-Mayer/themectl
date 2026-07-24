package integration

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Nico-Mayer/themectl/internal/cache"
	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

type VSCode struct {
	SettingsPath string
	Installer    ExtensionInstaller
}

func (v VSCode) Name() string { return "vscode" }

func (VSCode) Supports(t theme.Resolved) bool {
	return t.VSCode != nil && t.VSCode.Theme != ""
}

func (v VSCode) Apply(t theme.Resolved) error {
	spec := t.VSCode

	if v.Installer != nil {
		for _, url := range spec.Extensions {
			if err := v.Installer.Ensure(url); err != nil {
				return err
			}
		}
	}

	data, err := os.ReadFile(v.SettingsPath)
	if err != nil {
		return fmt.Errorf("read vscode settings: %w", err)
	}

	updated, err := setJSONCString(string(data), "workbench.colorTheme", spec.Theme)
	if err != nil {
		return err
	}

	if spec.IconTheme != "" {
		updated, err = setJSONCString(updated, "workbench.iconTheme", spec.IconTheme)
		if err != nil {
			return err
		}
	}

	if err := os.WriteFile(v.SettingsPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write vscode settings: %w", err)
	}

	return nil
}

func (v VSCode) Check() error {
	return checkConfigDir(v.Name(), v.SettingsPath)
}

func newVSCode(cfg config.Config) Integration {
	v := VSCode{
		SettingsPath: cfg.Settings.VSCode.Path(appConfigFile("vscode", "settings.json")),
	}
	if _, err := exec.LookPath("code"); err != nil {
		slog.Debug("vscode extension install disabled, code CLI not found", "err", err)
		return v
	}

	v.Installer = codeInstaller{
		cache: cache.New(filepath.Join(cfg.CacheDir(), "vscode")),
		install: func(id string) error {
			out, err := exec.Command("code", "--install-extension", id).CombinedOutput()
			if err != nil {
				return fmt.Errorf("code --install-extension %s: %w (%s)", id, err, strings.TrimSpace(string(out)))
			}
			return nil
		},
	}
	return v
}
