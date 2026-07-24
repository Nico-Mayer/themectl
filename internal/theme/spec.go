package theme

type SymlinkSpec struct {
	URL string `toml:"url,omitempty" jsonschema:"description=URL of the theme file to symlink. Falls back to the local asset when unset."`
}

type GhosttySpec struct {
	Theme string `toml:"theme,omitempty" jsonschema:"description=Ghostty theme name."`
}

type HelixSpec struct {
	Theme string `toml:"theme,omitempty" jsonschema:"description=Helix theme name."`
}

type ZedSpec struct {
	Theme      string   `toml:"theme,omitempty" jsonschema:"description=Zed theme name."`
	IconTheme  string   `toml:"icon_theme,omitempty" jsonschema:"description=Zed icon theme name. Defaults to Zed (Default) when unset."`
	Extensions []string `toml:"extensions,omitempty" jsonschema:"description=Zed extension repo URLs to install.,uniqueItems=true"`
}

type VSCodeSpec struct {
	Theme      string   `toml:"theme,omitempty" jsonschema:"description=VS Code color theme name."`
	IconTheme  string   `toml:"icon_theme,omitempty" jsonschema:"description=VS Code file icon theme name. Left untouched when unset."`
	Extensions []string `toml:"extensions,omitempty" jsonschema:"description=VS Code marketplace extension IDs (publisher.name) to install.,uniqueItems=true"`
}

type Spec struct {
	Appearance       *Appearance  `toml:"appearance,omitempty" jsonschema:"description=Appearance of this variant. Falls back to the family default; resolving fails if neither sets it."`
	WallpaperSources []string     `toml:"wallpaper_sources,omitempty" jsonschema:"description=Extra wallpaper sources: theme ids (family/variant) or shared wallpaper dir names. Falls back to the family default; the variant's own wallpaper dir is always included.,uniqueItems=true"`
	Ghostty          *GhosttySpec `toml:"ghostty,omitempty" jsonschema:"description=Ghostty integration settings."`
	Helix            *HelixSpec   `toml:"helix,omitempty" jsonschema:"description=Helix integration settings."`
	Zed              *ZedSpec     `toml:"zed,omitempty" jsonschema:"description=Zed integration settings."`
	VSCode           *VSCodeSpec  `toml:"vscode,omitempty" jsonschema:"description=VS Code integration settings."`
	Nvim             *SymlinkSpec `toml:"nvim,omitempty" jsonschema:"description=Nvim integration settings."`
	Yazi             *SymlinkSpec `toml:"yazi,omitempty" jsonschema:"description=Yazi integration settings."`
	Eza              *SymlinkSpec `toml:"eza,omitempty" jsonschema:"description=Eza integration settings."`
}

type ThemeFile struct {
	Defaults Spec            `toml:"defaults,omitempty" jsonschema:"description=Defaults inherited by every variant; a variant may override any of them."`
	Variants map[string]Spec `toml:"variants,omitempty" jsonschema:"description=Variants of this family, keyed by variant name. An empty table declares a variant that inherits all defaults."`
}
