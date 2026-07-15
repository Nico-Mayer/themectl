package integration

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type gitInstaller struct {
	extensionsDir string
}

func (g gitInstaller) Ensure(ref ExtensionRef) error {
	if err := os.MkdirAll(g.extensionsDir, 0o755); err != nil {
		return fmt.Errorf("ensure extensions dir: %w", err)
	}

	head, err := remoteHead(ref.URL)
	if err != nil {
		return err
	}
	marker := g.markerPath(ref.URL)
	if prev, _ := os.ReadFile(marker); string(prev) == head {
		return nil
	}

	tmp, err := os.MkdirTemp(g.extensionsDir, ".zed-ext-*")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	if err := sparseClone(ref.URL, tmp); err != nil {
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

	return os.WriteFile(marker, []byte(head), 0o644)
}

func remoteHead(url string) (string, error) {
	out, err := exec.Command("git", "ls-remote", "https://"+url, "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("ls-remote %s: %w", url, err)
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return "", fmt.Errorf("ls-remote %s: empty HEAD", url)
	}
	return fields[0], nil
}

func sparseClone(url, dst string) error {
	steps := [][]string{
		{"clone", "--depth", "1", "--filter=blob:none", "--no-checkout", "https://" + url, dst},
		{"-C", dst, "sparse-checkout", "set", "themes"},
		{"-C", dst, "checkout"},
	}
	for _, args := range steps {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			return fmt.Errorf("git %s: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
		}
	}
	return nil
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
