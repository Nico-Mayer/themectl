package store

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Nico-Mayer/themectl/internal/git"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

var familyNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

func Install(themesDir, url, name string, force bool) (string, error) {
	if name == "" {
		name = deriveFamilyName(url)
	}

	if !familyNamePattern.MatchString(name) {
		return "", fmt.Errorf("theme family name %q must start with a lowercase letter or number and contain only lowercase letters, numbers, dots, underscores, or hyphens", name)
	}

	if err := git.Installed(); err != nil {
		return "", err
	}

	if err := os.MkdirAll(themesDir, 0o755); err != nil {
		return "", fmt.Errorf("create themes directory: %w", err)
	}

	temp, err := os.MkdirTemp(themesDir, ".install-*")
	if err != nil {
		return "", fmt.Errorf("create temporary install directory: %w", err)
	}
	defer os.RemoveAll(temp)

	dst := filepath.Join(temp, "repo")
	if err := git.CloneShallow(url, dst); err != nil {
		return "", err
	}

	data, err := os.ReadFile(filepath.Join(dst, "theme.toml"))
	if err != nil {
		return "", fmt.Errorf("repository does not contain a readable theme.toml: %w", err)
	}

	var tf theme.ThemeFile
	if err := toml.Unmarshal(data, &tf); err != nil {
		return "", fmt.Errorf("parse theme.toml: %w", err)
	}

	fam := theme.Family{Name: name, Defaults: tf.Defaults}
	ok := false

	for v, spec := range tf.Variants {
		if _, err := theme.Resolve(fam, theme.Variant{Name: v, Spec: spec}); err == nil {
			ok = true
			break
		}
	}
	if !ok {
		return "", fmt.Errorf("theme repository has no resolvable variants")
	}

	target := filepath.Join(themesDir, name)
	if _, err := os.Stat(target); err == nil {
		if !force {
			return "", fmt.Errorf("theme family %q is already installed; rerun with --force to replace it", name)
		}
		if err := os.RemoveAll(target); err != nil {
			return "", fmt.Errorf("remove existing theme family: %w", err)
		}
	}
	if err := os.Rename(dst, target); err != nil {
		return "", fmt.Errorf("install theme family: %w", err)
	}

	return name, nil
}

func Uninstall(themesDir, name string) error {
	target := filepath.Join(themesDir, name)

	if _, err := os.Stat(target); os.IsNotExist(err) {
		return fmt.Errorf("theme family %q is not installed; run `themectl uninstall` to select an installed theme family", name)
	}

	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("remove theme family: %w", err)
	}

	return nil
}

func deriveFamilyName(url string) string {
	base := path.Base(strings.TrimRight(url, "/"))
	base = strings.TrimSuffix(base, ".git")
	return strings.ToLower(base)
}
