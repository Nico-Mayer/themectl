package theme

import (
	"fmt"
	"reflect"
	"slices"
)

type Family struct {
	Name     string
	Defaults Spec
}

type Variant struct {
	Name string
	VariantSpec
}

type Resolved struct {
	Family           string
	Variant          string
	Appearance       Appearance
	WallpaperSources []string
	Ghostty          *GhosttySpec
	Helix            *HelixSpec
	Zed              *ZedSpec
}

func (r *Resolved) ID() string {
	return fmt.Sprintf("%s/%s", r.Family, r.Variant)
}

func (r *Resolved) Themes() map[string]string {
	out := make(map[string]string)
	if r.Ghostty != nil && r.Ghostty.Theme != "" {
		out["ghostty"] = r.Ghostty.Theme
	}
	if r.Helix != nil && r.Helix.Theme != "" {
		out["helix"] = r.Helix.Theme
	}
	if r.Zed != nil && r.Zed.Theme != "" {
		out["zed"] = r.Zed.Theme
	}
	return out
}

func Resolve(fam Family, variant Variant) (Resolved, error) {
	spec := merge(fam.Defaults, variant.Spec)
	if spec.Appearance == nil {
		return Resolved{}, fmt.Errorf("theme %s/%s: appearance set by neither variant nor family", fam.Name, variant.Name)
	}

	id := fam.Name + "/" + variant.Name
	return Resolved{
		Family:           fam.Name,
		Variant:          variant.Name,
		Appearance:       *spec.Appearance,
		WallpaperSources: append(slices.Clone(variant.WallpaperSources), id),
		Ghostty:          spec.Ghostty,
		Helix:            spec.Helix,
		Zed:              spec.Zed,
	}, nil
}

func merge(base, over Spec) Spec {
	out := over
	if out.Appearance == nil {
		out.Appearance = base.Appearance
	}
	ov := reflect.ValueOf(&out).Elem()
	bv := reflect.ValueOf(base)
	for i := range ov.NumField() {
		f := ov.Field(i)
		if f.Kind() != reflect.Pointer || f.Type().Elem().Kind() != reflect.Struct {
			continue // only the *XxxSpec section fields
		}
		b := bv.Field(i)
		switch {
		case b.IsNil():
			// no default section: keep the variant's as-is
		case f.IsNil():
			f.Set(b) // no variant section: inherit the default wholesale
		default:
			f.Set(mergeSection(b, f))
		}
	}
	return out
}

// mergeSection returns a fresh *T where zero fields of over are filled in
// from base. Inputs are never mutated.
func mergeSection(base, over reflect.Value) reflect.Value {
	out := reflect.New(over.Type().Elem())
	out.Elem().Set(over.Elem())
	for i := range out.Elem().NumField() {
		f := out.Elem().Field(i)
		if f.IsZero() {
			f.Set(base.Elem().Field(i))
		}
	}
	return out
}
