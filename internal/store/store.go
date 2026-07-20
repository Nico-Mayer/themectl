package store

import (
	"cmp"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"math/rand/v2"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

var reservedNames = []string{
	"theme.toml",
}

type Store struct {
	fsys fs.FS
}

func NewStore(fsys fs.FS) *Store {
	return &Store{
		fsys: fsys,
	}
}

func (s *Store) Resolve(id string) (theme.Resolved, error) {
	famName, variantName, found := strings.Cut(id, "/")
	if !found {
		return theme.Resolved{}, fmt.Errorf("theme id %q: want \"family/variant\"", id)
	}

	tf, err := s.themeFile(famName)
	if err != nil {
		return theme.Resolved{}, err
	}

	vs, ok := tf.Variants[variantName]
	if !ok {
		return theme.Resolved{}, fmt.Errorf("theme %s: variant %q not declared in theme.toml", id, variantName)
	}

	return theme.Resolve(
		theme.Family{Name: famName, Defaults: tf.Defaults},
		theme.Variant{Name: variantName, Spec: vs},
	)
}

func (s *Store) IDs() ([]string, error) {
	families, err := s.allFamilies()
	if err != nil {
		return nil, err
	}

	var out []string
	for _, fam := range families {
		variants, err := s.listVariants(fam)
		if err != nil {
			return nil, err
		}
		for _, v := range variants {
			out = append(out, fam+"/"+v)
		}
	}
	return out, nil
}

func (s *Store) List(a theme.Appearance) ([]theme.Resolved, error) {
	all, err := s.resolveAll()
	if err != nil {
		return nil, err
	}

	var out []theme.Resolved
	for _, r := range all {
		if r.Appearance == a || a == theme.AnyAppearance {
			out = append(out, r)
		}
	}

	return out, nil
}

func (s *Store) PickRandom(a theme.Appearance) (theme.Resolved, error) {
	candidates, err := s.List(a)
	if err != nil {
		return theme.Resolved{}, err
	}

	if len(candidates) == 0 {
		return theme.Resolved{}, fmt.Errorf("no matching candidates found for appearance %v", a)
	}

	return candidates[rand.IntN(len(candidates))], nil
}

func (s *Store) listVariants(family string) ([]string, error) {
	tf, err := s.themeFile(family)
	if err != nil {
		return nil, err
	}

	return slices.Sorted(maps.Keys(tf.Variants)), nil
}

func (s *Store) resolveAll() ([]theme.Resolved, error) {
	families, err := s.allFamilies()
	if err != nil {
		return nil, err
	}

	var (
		mu  sync.Mutex
		out []theme.Resolved
		wg  sync.WaitGroup
	)

	for _, name := range families {
		wg.Go(func() {
			resolved := s.resolveFamily(name)
			mu.Lock()
			out = append(out, resolved...)
			mu.Unlock()
		})
	}
	wg.Wait()

	slices.SortFunc(out, func(a, b theme.Resolved) int {
		return cmp.Compare(a.ID(), b.ID())
	})
	return out, nil
}

func (s *Store) resolveFamily(name string) []theme.Resolved {
	tf, err := s.themeFile(name)
	if err != nil {
		slog.Debug("skipping unresolvable family", "family", name, "err", err)
		return nil
	}

	fam := theme.Family{Name: name, Defaults: tf.Defaults}
	var out []theme.Resolved

	for _, v := range slices.Sorted(maps.Keys(tf.Variants)) {
		res, err := theme.Resolve(fam, theme.Variant{Name: v, Spec: tf.Variants[v]})
		if err != nil {
			slog.Debug("skipping unresolvable theme", "theme", name+"/"+v, "err", err)
			continue
		}
		out = append(out, res)
	}
	return out
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
	slices.Sort(out)

	return out, nil
}

func (s *Store) themeFile(family string) (theme.ThemeFile, error) {
	var tf theme.ThemeFile
	if err := s.decode(path.Join(family, "theme.toml"), &tf); err != nil {
		return theme.ThemeFile{}, err
	}
	return tf, nil
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
