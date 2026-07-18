package theme

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

func (s *Store) Resolve(id string) (Resolved, error) {
	famName, variantName, found := strings.Cut(id, "/")
	if !found {
		return Resolved{}, fmt.Errorf("theme id %q: want \"family/variant\"", id)
	}

	tf, err := s.themeFile(famName)
	if err != nil {
		return Resolved{}, err
	}

	vs, ok := tf.Variants[variantName]
	if !ok {
		return Resolved{}, fmt.Errorf("theme %s: variant %q not declared in theme.toml", id, variantName)
	}

	return Resolve(
		Family{Name: famName, Defaults: tf.Defaults},
		Variant{Name: variantName, VariantSpec: vs},
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

func (s *Store) List(a Appearance) ([]Resolved, error) {
	all, err := s.resolveAll()
	if err != nil {
		return nil, err
	}

	var out []Resolved
	for _, r := range all {
		if r.Appearance == a || a == AnyAppearance {
			out = append(out, r)
		}
	}

	return out, nil
}

func (s *Store) PickRandom(a Appearance) (Resolved, error) {
	candidates, err := s.List(a)
	if err != nil {
		return Resolved{}, err
	}

	if len(candidates) == 0 {
		return Resolved{}, fmt.Errorf("no matching candidates found for appearance %v", a)
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

func (s *Store) resolveAll() ([]Resolved, error) {
	families, err := s.allFamilies()
	if err != nil {
		return nil, err
	}

	var (
		mu  sync.Mutex
		out []Resolved
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

	slices.SortFunc(out, func(a, b Resolved) int {
		return cmp.Compare(a.ID(), b.ID())
	})
	return out, nil
}

func (s *Store) resolveFamily(name string) []Resolved {
	tf, err := s.themeFile(name)
	if err != nil {
		slog.Debug("skipping unresolvable family", "family", name, "err", err)
		return nil
	}

	fam := Family{Name: name, Defaults: tf.Defaults}
	var out []Resolved

	for _, v := range slices.Sorted(maps.Keys(tf.Variants)) {
		res, err := Resolve(fam, Variant{Name: v, VariantSpec: tf.Variants[v]})
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

func (s *Store) themeFile(family string) (ThemeFile, error) {
	var tf ThemeFile
	if err := s.decode(path.Join(family, "theme.toml"), &tf); err != nil {
		return ThemeFile{}, err
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
