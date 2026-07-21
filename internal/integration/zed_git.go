package integration

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Nico-Mayer/themectl/internal/git"
)

const installCheckTTL = 30 * 24 * time.Hour

type gitInstaller struct {
	extensionsDir string
}

func (g gitInstaller) Ensure(ref ExtensionRef) error {
	if err := os.MkdirAll(g.extensionsDir, 0o755); err != nil {
		return fmt.Errorf("ensure extensions dir: %w", err)
	}

	marker := g.markerPath(ref.URL)
	if info, err := os.Stat(marker); err == nil && time.Since(info.ModTime()) < installCheckTTL {
		slog.Debug("zed extension recently checked, skipping", "url", ref.URL)
		return nil
	}

	slog.Debug("checking zed extension for updates", "url", ref.URL)
	head, err := git.RemoteHead(ref.URL)
	if err != nil {
		return err
	}
	if prev, _ := os.ReadFile(marker); string(prev) == head {
		slog.Debug("zed extension up to date", "url", ref.URL)
		_ = os.Chtimes(marker, time.Now(), time.Now())
		return nil
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

	return os.WriteFile(marker, []byte(head), 0o644)
}

func (g gitInstaller) markerPath(url string) string {
	sum := sha256.Sum256([]byte(url))
	return filepath.Join(g.extensionsDir, ".head-"+hex.EncodeToString(sum[:8]))
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
