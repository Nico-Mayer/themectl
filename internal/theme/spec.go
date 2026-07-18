package theme

type GhosttySpec struct {
	Theme string `toml:"theme,omitempty" jsonschema:"description=Ghostty theme name."`
}

type HelixSpec struct {
	Theme string `toml:"theme,omitempty" jsonschema:"description=Helix theme name."`
}

type ZedSpec struct {
	Theme      string   `toml:"theme,omitempty" jsonschema:"description=Zed theme name."`
	IconTheme  string   `toml:"icon_theme,omitempty" jsonschema:"description=Zed icon theme name."`
	Extensions []string `toml:"extensions,omitempty" jsonschema:"description=Zed extension repo URLs to install.,uniqueItems=true"`
}

type Spec struct {
	Appearance *Appearance  `toml:"appearance,omitempty" jsonschema:"description=Appearance of this variant. Falls back to the family default; resolving fails if neither sets it."`
	Ghostty    *GhosttySpec `toml:"ghostty,omitempty" jsonschema:"description=Ghostty integration settings."`
	Helix      *HelixSpec   `toml:"helix,omitempty" jsonschema:"description=Helix integration settings."`
	Zed        *ZedSpec     `toml:"zed,omitempty" jsonschema:"description=Zed integration settings."`
}

type VariantSpec struct {
	Spec
	WallpaperSources []string `toml:"wallpaper_sources,omitempty" jsonschema:"description=Extra wallpaper sources: theme ids (family/variant) or shared wallpaper dir names. The variant's own wallpaper dir is always included.,uniqueItems=true"`
}

type ThemeFile struct {
	Defaults Spec                   `toml:"defaults,omitempty" jsonschema:"description=Defaults inherited by every variant; a variant may override any of them."`
	Variants map[string]VariantSpec `toml:"variants,omitempty" jsonschema:"description=Variants of this family, keyed by variant name. An empty table declares a variant that inherits all defaults."`
}
