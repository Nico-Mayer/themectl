package theme

import (
	"fmt"
	"maps"
)

type Spec struct {
	Appearance *Appearance       `toml:"appearance,omitempty" jsonschema:"description=Appearance of this variant. Falls back to the family default; resolving fails if neither sets it."`
	Themes     map[string]string `toml:"themes,omitempty" jsonschema:"description=Per-integration theme names keyed by integration. Overrides the family defaults."`
}

type FamilyFile struct {
	Defaults Spec `toml:"defaults,omitempty" jsonschema:"description=Defaults inherited by every variant; a variant may override any of them."`
}

type VariantFile struct {
	Spec
	WallpaperSources []string `toml:"wallpaper_sources,omitempty" jsonschema:"description=Extra wallpaper sources: theme ids (family/variant) or shared wallpaper dir names. The variant's own wallpaper dir is always included.,uniqueItems=ture"`
}

type Family struct {
	Name     string
	Defaults Spec
}

type Variant struct {
	Name             string
	Spec             Spec
	WallpaperSources []string
}

type Resolved struct {
	Family           string
	Variant          string
	Appearance       Appearance
	WallpaperSources []string
	Themes           map[string]string
}

func (r *Resolved) ID() string {
	return fmt.Sprintf("%s/%s", r.Family, r.Variant)
}

func Resolve(fam Family, variant Variant) (Resolved, error) {
	spec := merge(fam.Defaults, variant.Spec)
	if spec.Appearance == nil {
		return Resolved{}, fmt.Errorf("theme %s/%s: appearance set by neither variant nor family", fam.Name, variant.Name)
	}

	return Resolved{
		Family:           fam.Name,
		Variant:          variant.Name,
		Appearance:       *spec.Appearance,
		WallpaperSources: variant.WallpaperSources,
		Themes:           spec.Themes,
	}, nil
}

func merge(base, over Spec) Spec {
	out := Spec{
		Appearance: base.Appearance,
		Themes:     maps.Clone(base.Themes),
	}

	if over.Appearance != nil {
		out.Appearance = over.Appearance
	}
	if over.Themes != nil {
		if out.Themes == nil {
			out.Themes = make(map[string]string, len(over.Themes))
		}
		maps.Copy(out.Themes, over.Themes) // per-key override
	}

	return out
}
