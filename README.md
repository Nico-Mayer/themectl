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

- [ ] Add some sort of pallet witch could be diplayed when picking a theme in raycast extension
- [ ] Install command to install themes from a GitHub URL
- [ ] Create theme cmd which opens tui form and generates a folder to work on in themesDir()
- [ ] `doctor` / `status` command report current theme, settings, and which integrations are applied/available
- [ ] Add some sort of release strategie (GoReleaser?)
- [ ] Restrict asset copying on materialize to active integrations only, so the current folder doesn't get polluted _(low)_
- [ ] Maybe instead of providing assets like eza and yazi themes there should also be a way to provide a url whitch generates the asset from that url, the drawback is that this needs a internet connection to work, so we i may need cashing. this would bring the benefit of just linking to a source for a port whitch already exists so no asset duplication needed in a theme spec.

### Missing integrations

- [ ] VSCode
- [ ] Other terminal emulators _(low)_
- [ ] Chromium verify feasibility, may need elevated privileges on macOS to set policies (Helium and other Chromium forks)

### Quick wins
