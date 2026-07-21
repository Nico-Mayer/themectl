package integration

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Nico-Mayer/themectl/internal/cache"
	"github.com/Nico-Mayer/themectl/internal/git"
)

const installCheckTTL = 30 * 24 * time.Hour

type gitInstaller struct {
	extensionsDir string
	cache         cache.Cache
}

func (g gitInstaller) Ensure(ref ExtensionRef) error {
	if err := os.MkdirAll(g.extensionsDir, 0o755); err != nil {
		return fmt.Errorf("ensure extensions dir: %w", err)
	}

	if g.cache.Fresh(ref.URL, installCheckTTL) {
		slog.Debug("zed extension recently checked, skipping", "url", ref.URL)
		return nil
	}

	slog.Debug("checking zed extension for updates", "url", ref.URL)
	head, err := git.RemoteHead(ref.URL)
	if err != nil {
		return err
	}

	if prev, ok := g.cache.Get(ref.URL); ok && prev == head {
		slog.Debug("zed extension up to date", "url", ref.URL)
		return g.cache.Touch(ref.URL)
	}

	tmp, err := os.MkdirTemp(g.extensionsDir, ".zed-ext-*")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	if err := git.SparseClone(ref.URL, tmp, "themes", "icon_themes", "icons"); err != nil {
		return err
	}

	id, err := extensionID(filepath.Join(tmp, "extension.toml"))
	if err != nil {
		return err
	}

	target := filepath.Join(g.extensionsDir, id)
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("clear old extension %q: %w", target, err)
	}
	if err := os.Rename(tmp, target); err != nil {
		return fmt.Errorf("install extension %q: %w", id, err)
	}
	slog.Info("zed extension installed", "extension", id, "url", ref.URL)

	return g.cache.Put(ref.URL, head)
}

func extensionID(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read extension.toml: %w", err)
	}
	var m struct {
		ID string `toml:"id"`
	}
	if err := toml.Unmarshal(data, &m); err != nil {
		return "", fmt.Errorf("parse extension.toml: %w", err)
	}
	return m.ID, nil
}
