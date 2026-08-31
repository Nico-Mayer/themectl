# themectl

Manage and apply themes across your tools with one command. Define a theme once
as a `family/variant` (e.g. `catppuccin/mocha`) and `themectl` propagates it to
every configured integration editor, terminal, shell tooling, wallpaper and
system appearance in a single, concurrent pass.

## Usage

```bash
themectl list                         # list available themes (ls)
themectl set <theme-id>               # set and apply current theme (use, apply)
themectl set random                   # set random theme (--light or --dark to filter)
themectl current                      # show current theme
themectl wallpaper                    # select wallpaper from current theme
themectl wallpaper set --random       # select random wallpaper from current theme
themectl install <git-url>            # install theme family (--name, --force)
themectl uninstall <theme-family>     # uninstall theme family
themectl update                       # update theme families installed from Git
themectl refresh                      # reapply current theme to all integrations
themectl doctor                       # show theme, settings, and integration status
themectl cache clear                  # delete cached remote assets
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

Asset-based integrations (`nvim`, `eza`, `yazi`, `rio`, etc.) can reference their
asset by URL instead of bundling the file — handy for linking an existing port
of a theme without duplicating it:

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

### Windows Terminal

Windows Terminal keeps the terminal colors and the window chrome as two separate
things, so the integration takes two assets. `scheme_url` is the color scheme,
`theme_url` colors the tabs and title bar. Upstream ports usually ship both, so
you can point at them directly:

```toml
[defaults.windows-terminal]
scheme_url = "https://raw.githubusercontent.com/catppuccin/windows-terminal/main/mocha.json"
theme_url  = "https://raw.githubusercontent.com/catppuccin/windows-terminal/main/mochaTheme.json"

[variants.latte.windows-terminal]
scheme_url = "https://raw.githubusercontent.com/catppuccin/windows-terminal/main/latte.json"
theme_url  = "https://raw.githubusercontent.com/catppuccin/windows-terminal/main/latteTheme.json"
```

Bundled assets work too, as `windows-terminal.json` and
`windows-terminal-theme.json`. Whatever the assets call themselves, themectl
installs both as `themectl`, so you never have to repeat a name. Light or dark
window chrome comes from the theme asset itself, not from `appearance`.

`theme_url` is optional. With only a scheme, the terminal colors change and the
chrome falls back to Windows Terminal's own default.

The integration runs on Windows only. It is registered and enabled there by
default, and is invisible everywhere else - a theme file can declare a
`windows-terminal` section on any OS without warnings.

themectl writes the color scheme as a fragment under
`%LOCALAPPDATA%\Microsoft\Windows Terminal\Fragments\themectl\` and only
touches `profiles.defaults.colorScheme`, the `themes` array and `theme` in
`settings.json`. Comments and formatting are preserved, and a profile that sets
its own `colorScheme` keeps it.

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

[rio]
config_file = "~/.config/rio/config.toml" # theme symlink lands in themes/ next to it

# defaults to the Store install; set this for unpackaged, Preview or portable
[windows-terminal]
config_file = "$LOCALAPPDATA/Microsoft/Windows Terminal/settings.json"

# symlink integrations: choose where the theme asset is linked
[nvim]
target = "~/.dotfiles/nvim/plugin/99_theme.lua"
```

## Roadmap

### Features

- `create` command: TUI form that scaffolds a new theme folder in `themesDir()`
- Add a settings option to include or exclude an integration by operating system

### Missing integrations

- [ ] Other terminal emulators _(low)_ — Ghostty, Rio, and Windows Terminal are covered
- [ ] Chromium: verify feasibility; setting policies may need elevated privileges on macOS (Helium and other Chromium forks)

### Quick wins

### Maybe

- Expose a color palette per theme so the Raycast extension can display it in the theme picker
- Add a `sha256` field next to `url` to pin remote assets against upstream tampering
- Add Philips Hue integration
- Allow users to mark themes as favorites
