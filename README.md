# themectl

Manage and apply themes across your tools with one command. Define a theme once
as a `family/variant` (e.g. `catppuccin/mocha`) and `themectl` propagates it to
every configured integration editor, terminal, shell tooling, wallpaper and
system appearance in a single, concurrent pass.

## Usage

```bash
themectl list                        # list all themes (ls)
themectl set <theme-name>            # apply a theme (use, apply)
themectl set random                  # random theme  (--light / --dark to filter)
themectl current                     # print active theme
themectl wallpaper                   # pick from all available wallpapers for this theme
themectl wallpaper --random          # reshuffle wallpaper for current theme
themectl install <git-url>           # install a theme from a git repo (--name, --force)
themectl uninstall <theme-family>    # uninstalls the selected theme family from
themectl update                      # updates all themes installed via git
```

## Configuration

Everything lives in `~/.config/themectl`. Each theme family is a folder under
`themes/` described by a single `theme.toml`: family-wide `[defaults]` plus one
`[variants.<name>]` table per variant, where a variant overrides individual
fields and inherits the rest. Assets (wallpapers, `nvim.lua`, `eza.yml`, …) sit
next to the spec or in an optional per-variant folder. Global settings go in
`themectl.toml` at the root; the `#:schema` directive on the first line gives
completion and validation in schema-aware TOML editors.

```toml
# themes/catppuccin/theme.toml
#:schema https://raw.githubusercontent.com/Nico-Mayer/themectl/main/schemas/theme.schema.json
[defaults]
appearance = "dark"

[defaults.zed]
theme = "Catppuccin Mocha"
icon_theme = "Catppuccin Mocha"
extensions = ["https://github.com/catppuccin/zed"]

[variants.mocha]
# empty table declares the variant; inherits all defaults

[variants.latte]
appearance = "light"

[variants.latte.zed]
theme = "Catppuccin Latte" # icon_theme and extensions inherited
```

### Remote assets

Symlink integrations (`nvim`, `eza`, `yazi`, etc.) can reference their asset by URL
instead of bundling the file — handy for linking an existing port of a theme
without duplicating it:

```toml
[defaults.nvim]
url = "https://raw.githubusercontent.com/catppuccin/nvim/main/lua/catppuccin/init.lua"

[variants.latte.nvim]
url = "https://raw.githubusercontent.com/catppuccin/nvim/main/lua/catppuccin/latte.lua"
```

When `url` is set it is the source of truth: a bundled file with the same
canonical name is ignored (with a warning). When it is unset, the bundled
file is used as before. Only `https://` URLs are accepted.

Downloads are cached for a week, so applying a theme works offline once the
asset has been fetched. If a fetch fails with nothing cached, that asset is
skipped with a warning and the integration skips itself for that run.
`themectl cache clear` forces a refetch.

```toml
# themectl.toml
#:schema https://raw.githubusercontent.com/Nico-Mayer/themectl/main/schemas/settings.schema.json

# integrations to run on apply; omit to run the default set
integrations = ["ghostty", "zed", "nvim", "wallpaper", "system-appearance"]

# file-editing integrations: point themectl at the file it should edit
[ghostty]
config_file = "~/.config/ghostty/config.ghostty"

[zed]
config_file = "$XDG_CONFIG_HOME/zed/settings.json"

# symlink integrations: choose where the theme asset is linked
[nvim]
target = "~/.dotfiles/nvim/plugin/99_theme.lua"
```

## Roadmap

### Features

- `create` command: TUI form that scaffolds a new theme folder in themesDir()
- Add a option in settings to make a integration exclusicve for one operating system or exlude for one

### Missing integrations

- [ ] Other terminal emulators _(low)_
- [ ] Chromium verify feasibility, may need elevated privileges on macOS to set policies (Helium and other Chromium forks)

### Quick wins

### Maybe

- Expose a color palette per theme so the raycast extension can display it in the theme picker
- `sha256` field next to `url` to pin remote assets against upstream tampering
- Philips Hue integration?
- Declare some themes as favorites ?
