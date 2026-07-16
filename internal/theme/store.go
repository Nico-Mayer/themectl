package theme

import (
	"fmt"
	"io/fs"
	"maps"
	"math/rand/v2"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

var reservedNames = []string{
	"family.toml",
	"variant.toml",
}

type Store struct {
	fsys fs.FS
}

func NewStore(fsys fs.FS) *Store {
	return &Store{
		fsys: fsys,
	}
}

func (s *Store) Resolve(id string) (Resolved, error) {
	famName, variantName, found := strings.Cut(id, "/")
	if !found {
		return Resolved{}, fmt.Errorf("theme id %q: want \"family/variant\"", id)
	}

	family, err := s.family(famName)
	if err != nil {
		return Resolved{}, err
	}

	variant, err := s.variant(famName, variantName)
	if err != nil {
		return Resolved{}, err
	}

	return Resolve(family, variant)
}

func (s *Store) List(family string) ([]string, error) {
	entries, err := fs.ReadDir(s.fsys, family)
	if err != nil {
		return []string{}, fmt.Errorf("read family %q: %w", family, err)
	}

	var out []string
	for _, e := range entries {
		if e.IsDir() && !(e.Name() == "wallpaper") {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

func (s *Store) ListAll() ([]string, error) {
	families, err := s.allFamilies()
	if err != nil {
		return nil, err
	}

	var out []string
	for _, fam := range families {
		variants, err := s.List(fam)
		if err != nil {
			return nil, err
		}
		for _, v := range variants {
			out = append(out, fam+"/"+v)
		}
	}
	return out, nil
}

func (s *Store) resolveAll() ([]Resolved, error) {
	all, err := s.ListAll()
	if err != nil {
		return nil, err
	}

	ch := make(chan Resolved, len(all))
	var wg sync.WaitGroup
	for _, themeID := range all {
		wg.Go(func() {
			if res, err := s.Resolve(themeID); err == nil {
				ch <- res
			}
		})
	}
	wg.Wait()
	close(ch)

	var out []Resolved
	for r := range ch {
		out = append(out, r)
	}

	return out, nil
}

func (s *Store) PickRandom(a *Appearance) (Resolved, error) {
	all, err := s.resolveAll()
	if err != nil {
		return Resolved{}, err
	}

	var candidates []Resolved
	for _, res := range all {
		if a == nil || res.Appearance == *a {
			candidates = append(candidates, res)
		}
	}

	if len(candidates) == 0 {
		return Resolved{}, fmt.Errorf("no matching candidates found for appearance %v", a)
	}

	return candidates[rand.IntN(len(candidates))], nil
}

func (s *Store) allFamilies() ([]string, error) {
	entries, err := fs.ReadDir(s.fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read themes root: %w", err)
	}

	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)

	return out, nil
}

func (s *Store) AssetPath(family, variant, asset string) (string, bool) {
	paths := []string{
		filepath.Join(family, variant, asset),
		filepath.Join(family, asset),
	}

	for _, p := range paths {
		_, err := fs.Stat(s.fsys, p)
		if err == nil {
			return p, true
		}
	}
	return "", false
}

func (s *Store) Assets(family, variant string) (map[string]string, error) {
	familyAssets, err := s.assetsIn(family)
	if err != nil {
		return nil, err
	}

	variantAssets, err := s.assetsIn(filepath.Join(family, variant))
	if err != nil {
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
		out[e.Name()] = filepath.Join(dir, e.Name())
	}

	return out, nil
}

func (s *Store) decode(path string, v any) error {
	data, err := fs.ReadFile(s.fsys, path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err = toml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	return nil
}

func (s *Store) family(name string) (Family, error) {
	var wrap struct {
		Defaults Spec `toml:"defaults"`
	}

	err := s.decode(filepath.Join(name, "family.toml"), &wrap)
	if err != nil {
		return Family{}, err
	}

	return Family{
		Name:     name,
		Defaults: wrap.Defaults,
	}, nil
}

func (s *Store) variant(family, name string) (Variant, error) {
	var v struct {
		Spec
		WallpaperSources []string `toml:"wallpaper_sources"`
	}

	err := s.decode(filepath.Join(family, name, "variant.toml"), &v)
	if err != nil {
		return Variant{}, err
	}

	return Variant{
		Name:             name,
		WallpaperSources: v.WallpaperSources,
		Spec:             v.Spec,
	}, nil
}
