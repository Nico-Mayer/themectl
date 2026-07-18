# themectl

Manage and apply themes across your tools with one command. Define a theme once
as a `family/variant` (e.g. `catppuccin/mocha`) and `themectl` propagates it to
every configured integration editor, terminal, shell tooling, wallpaper and
system appearance in a single, concurrent pass.

## Usage

```sh
themectl list                 # list all themes (ls)
themectl set catppuccin/mocha # apply a theme (use, apply)
themectl set random           # random theme  (--light / --dark to filter)
themectl current              # print active theme
themectl wallpaper            # print current wallpaper
themectl wallpaper --random   # reshuffle wallpaper for current theme
themectl -v <cmd>             # verbose logs to stderr
```

## Schemas

JSON Schemas for the TOML files live in [`schemas/`](schemas/): `family.schema.json`,
`variant.schema.json` and `settings.schema.json`. TOML LSPs (tombi, taplo /
Even Better TOML) pick them up via a directive on the first line for
completions and validation:

```toml
#:schema https://raw.githubusercontent.com/Nico-Mayer/themectl-cli/main/schemas/variant.schema.json
appearance = "dark"
```

## Roadmap

### Features

- Expose a color palette per theme so the raycast extension can display it in the theme picker
- `create` command: TUI form that scaffolds a new theme folder in themesDir()
- `install` command: install themes from a GitHub URL
- Allow theme specs to reference assets by URL instead of bundling them (link existing ports, no duplication). Needs network + caching for offline use.
- Add a option in settings to make a integration exclusicve for one operating system or exlude for one

### Missing integrations

- [ ] VSCode
- [ ] Other terminal emulators _(low)_
- [ ] Chromium verify feasibility, may need elevated privileges on macOS to set policies (Helium and other Chromium forks)

### Quick wins
