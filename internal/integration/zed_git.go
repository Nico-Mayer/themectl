package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type extensionManifest struct {
	ID          string   `toml:"id"`
	Name        string   `toml:"name"`
	Version     string   `toml:"version"`
	SchemaVer   int      `toml:"schema_version"`
	Authors     []string `toml:"authors"`
	Description string   `toml:"description"`
	Repository  string   `toml:"repository"`
}

type gitInstaller struct {
	extensionsDir string
}

func (g gitInstaller) Ensure(ref ExtensionRef) error {
	if err := os.MkdirAll(g.extensionsDir, 0o755); err != nil {
		return fmt.Errorf("ensure extensions dir: %w", err)
	}
	tmp, err := os.MkdirTemp(g.extensionsDir, ".zed-ext-*")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	cmd := exec.Command("git", "clone", "--depth", "1", "https://"+ref.URL, tmp)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("clone %s: %w (%s)", ref.URL, err, strings.TrimSpace(string(out)))
	}

	manifest, err := parseManifest(filepath.Join(tmp, "extension.toml"))
	if err != nil {
		return err
	}

	target := filepath.Join(g.extensionsDir, manifest.ID)
	if installedVersion(target) == manifest.Version {
		return nil
	}
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("clear old extension %q: %w", target, err)
	}
	return os.Rename(tmp, target)
}

func parseManifest(path string) (extensionManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return extensionManifest{}, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest extensionManifest
	if err := toml.Unmarshal(data, &manifest); err != nil {
		return extensionManifest{}, fmt.Errorf("parse extension.toml: %w", err)
	}
	return manifest, nil
}

func installedVersion(pathe string) string {
	manifest, err := parseManifest(pathe)
	if err != nil {
		return ""
	}

	return manifest.Version
}
