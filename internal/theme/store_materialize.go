package theme

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

func (s *Store) Materialize(themeId, targetDir string) error {
	family, variant, exists := strings.Cut(themeId, "/")
	if !exists {
		return fmt.Errorf("theme id %q: want \"family/variant\"", themeId)
	}

	assets, err := s.Assets(family, variant)
	if err != nil {
		return err
	}

	err = os.RemoveAll(targetDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(targetDir, 0o755)
	if err != nil {
		return err
	}

	for name, p := range assets {
		data, err := fs.ReadFile(s.fsys, p)
		if err != nil {
			return fmt.Errorf("read asset %s: %w", p, err)
		}

		err = os.WriteFile(filepath.Join(targetDir, name), data, 0o644)
		if err != nil {
			return fmt.Errorf("write asset %s: %w", name, err)
		}
	}

	return nil
}

func (s *Store) Assets(family, variant string) (map[string]string, error) {
	familyAssets, err := s.assetsIn(family)
	if err != nil {
		return nil, err
	}

	variantAssets, err := s.assetsIn(path.Join(family, variant))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	maps.Copy(familyAssets, variantAssets)

	return familyAssets, nil
}

func (s *Store) assetsIn(dir string) (map[string]string, error) {
	entries, err := fs.ReadDir(s.fsys, dir)

	out := make(map[string]string)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() || slices.Contains(reservedNames, e.Name()) {
			continue
		}
		out[e.Name()] = path.Join(dir, e.Name())
	}

	return out, nil
}
