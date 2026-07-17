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

## Roadmap

### Features

- [ ] Install command to install themes from a GitHub URL
- [ ] Theme-spec JSON schema for TOML LSP completions
- [ ] `doctor` / `status` command — report current theme, settings, and which integrations are applied/available
- [ ] `--json` output flag for `list` (let others consume it for example read appearance)
- [ ] Restrict asset copying on materialize to active integrations only, so the current folder doesn't get polluted _(low)_

### Missing integrations

- [ ] VSCode
- [ ] Other terminal emulators _(low)_
- [ ] Chromium — verify feasibility, may need elevated privileges on macOS to set policies (Helium and other Chromium forks)

### Quick wins

- [ ] `reapply` / `refresh` — re-run integrations for the current theme without changing it
- [ ] `list` filters + active marker — `--light` / `--dark`, `*` on the active theme
- [ ] `current --details` — resolved appearance, per-tool theme names, wallpaper sources
- [ ] `wallpaper set <file>` + `wallpaper list [theme]` — direct control and discovery
