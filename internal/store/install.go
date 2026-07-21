package store

import (
	"errors"
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
		return "", errors.New("name not allowed")
	}

	if err := git.Installed(); err != nil {
		return "", err
	}

	if err := os.MkdirAll(themesDir, 0o755); err != nil {
		return "", fmt.Errorf("unable to create themes directory: %w", err)
	}

	temp, err := os.MkdirTemp(themesDir, ".install-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(temp)

	dst := filepath.Join(temp, "repo")
	if err := git.CloneShallow(url, dst); err != nil {
		return "", err
	}

	data, err := os.ReadFile(filepath.Join(dst, "theme.toml"))
	if err != nil {
		return "", fmt.Errorf("not a theme repo (not readable theme.toml): %w", err)
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
		return "", errors.New("theme repo has no resolvable variant")
	}

	target := filepath.Join(themesDir, name)
	if _, err := os.Stat(target); err == nil {
		if !force {
			return "", fmt.Errorf("theme family %q already installed (use --force to replace)", name)
		}
		if err := os.RemoveAll(target); err != nil {
			return "", fmt.Errorf("unable to remove existing theme: %w", err)
		}
	}
	if err := os.Rename(dst, target); err != nil {
		return "", fmt.Errorf("unable to install theme: %w", err)
	}

	return name, nil
}

func Uninstall(themesDir, name string) error {
	target := filepath.Join(themesDir, name)

	if _, err := os.Stat(target); os.IsNotExist(err) {
		return fmt.Errorf("theme %q is not installed", name)
	}

	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("unable to remove theme: %w", err)
	}

	return nil
}

func deriveFamilyName(url string) string {
	base := path.Base(strings.TrimRight(url, "/"))
	base = strings.TrimSuffix(base, ".git")
	return strings.ToLower(base)
}
