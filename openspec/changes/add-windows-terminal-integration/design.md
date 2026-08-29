## Context

See proposal.md - Why.

The existing integrations fall into three shapes:

| Shape | Integrations | Mechanism |
| --- | --- | --- |
| Name only | ghostty, helix, zed, vscode | theme gives a name, edit one config key |
| Asset symlink | nvim, eza, yazi | theme ships a file, symlink it into place |
| Hybrid | rio | symlink the asset into `themes/`, then point a config key at it |

Windows Terminal is the hybrid shape, but every mechanical assumption behind
`Rio` breaks:

- There is no per-theme directory. Color schemes live in a `schemes` array
  inside `settings.json`, or in a JSON fragment file.
- The key to set is `profiles.defaults.colorScheme`, two levels deep. The
  existing `setJSONCString` is a flat regex and would bind to the first
  `colorScheme` anywhere in the file, which in a real settings file is almost
  certainly a profile in `profiles.list`.
- `settings.json` has two hardcoded locations and no environment override.
- Symlinks on Windows require Developer Mode or elevation, so `symlink.go` is
  not dependable on the one OS this integration targets.

Confirmed against the Windows Terminal docs and source:

- Fragment extensions may contribute `schemes`, and live at
  `%LOCALAPPDATA%\Microsoft\Windows Terminal\Fragments\<app-name>\*.json`.
  Microsoft's own tooling uses this to bulk-import iTerm2 schemes. This path is
  the same for packaged and unpackaged installs.
- Fragments may contribute `profiles` and `schemes` only. They cannot set
  `profiles.defaults` and cannot contribute `themes`.
- The window chrome is a separate concept from the color scheme, living in a
  `themes` array with its own keys: `window.applicationTheme`, `tab.background`,
  `tabRow.background`, `tabRow.unfocusedBackground`.
- `settings.json` live reload is validated by the user in this session.

The upstream ecosystem already reflects that split. `catppuccin/windows-terminal`
ships two files per variant, for example `mocha.json` and `mochaTheme.json`:

```json
// mocha.json                       // mochaTheme.json
{ "name": "Catppuccin Mocha",       { "name": "Catppuccin Mocha",
  "background": "#1E1E2E",            "tab":    { "background": "#1E1E2EFF",
  "foreground": "#CDD6F4",                        "unfocusedBackground": null },
  "black": "#45475A",                 "tabRow": { "background": "#181825FF",
  ... }                                           "unfocusedBackground": "#11111BFF" },
                                      "window": { "applicationTheme": "dark" } }
```

## Goals / Non-Goals

**Goals:**

- Consume upstream theme ports as-is, so a theme author links two existing files
  rather than authoring anything new.
- Never rewrite or reformat parts of `settings.json` the user owns.
- One JSON editing mechanism across every integration that edits JSON.

**Non-Goals:**

- Per-profile theming. Only `profiles.defaults` is written.
- Supporting Windows Terminal's `colorScheme: { light, dark }` pair form.
- Probing across install locations.

## Decisions

### Two assets, mirroring the upstream split

The theme section carries `scheme_url` and `theme_url`, materializing as
`windows-terminal.json` and `windows-terminal-theme.json`.

Named explicitly rather than reusing the conventional `url` for the primary
asset: the two files are peers with genuinely different roles, and `scheme_url`
next to `theme_url` says which is which at a glance. `url` next to `theme_url`
would not. This is the first integration with more than one asset, so there is
no existing convention to break.

Alternative considered: derive the chrome from the color scheme, using Windows
Terminal's `terminalBackground` literal for `tab.background` and
`tabRow.background`. That needs only one asset and no color parsing, but it
cannot express what the upstream chrome files actually do: catppuccin's tab row
uses the mantle and its unfocused background uses the crust, neither of which is
the terminal background. The result would be visibly flatter than the port the
theme author already published.

Alternative considered: parse colors out of the scheme and synthesize a chrome.
Rejected: it guesses at a design decision the theme author already made.

### The chrome carries its own light or dark mode

`window.applicationTheme` is inside the chrome asset, so the integration does not
consult `t.Appearance` at all. Deriving it independently would risk contradicting
the asset, and there is no case where themectl knows better than the port.

This also removes the need for the light and dark chrome definitions an earlier
draft of this design seeded. There is exactly one themectl-owned chrome
definition, replaced on each apply.

### Fragment file for the scheme, `settings.json` for the chrome

The scheme goes to
`%LOCALAPPDATA%\Microsoft\Windows Terminal\Fragments\themectl\scheme.json`, with
its `name` rewritten to `themectl`. The chrome is upserted into the `themes`
array of `settings.json`, also renamed to `themectl`.

The asymmetry is forced: fragments may contribute `schemes` but not `themes`.

Renaming both to a fixed `themectl` is the same trick `Rio` uses with
`theme = "themectl"`. The theme author supplies files and never has to make a
name in `theme.toml` agree with a name inside an asset. It also means the two
settings keys, `profiles.defaults.colorScheme` and the top-level `theme`, are
constants rather than per-theme values.

For the scheme, owning a whole file rather than a slice of the user's `schemes`
array means a bug can never eat a user-defined scheme, and reset is a file
delete. The chrome does not get that safety, which is why the upsert matches
strictly on the themectl-owned name.

Alternative considered: splice the scheme into the `schemes` array too, for
symmetry with the chrome. Rejected: it gives up the isolation for nothing.

### Write order: fragment first, then `settings.json`

Windows Terminal's file watcher covers `settings.json`. A reload re-runs the
full settings load, which reads fragments. Writing the fragment first means the
single `settings.json` write triggers a reload that picks up both.

Verified on Windows: applying a theme while the terminal is running updates the
colors with no restart, so the fragment written earlier in the same apply is
picked up by the reload that the settings.json write triggers. The write order
is load-bearing, not a precaution.

### Copy the assets, do not symlink

Symlinks on Windows need Developer Mode or elevation. The fragment is a small
JSON file rewritten on every apply anyway, so copying costs nothing and removes
a whole class of failure. This is the one place the integration deliberately
diverges from `SymlinkIntegration` and `Rio`.

### One JSON editing helper, `github.com/tailscale/hujson`

`setJSONCString` and `jsonc.go` are deleted. A single hujson-backed helper
replaces them and `zed` and `vscode` migrate onto it.

Windows Terminal's `settings.json` is JWCC: shipped with `//` comments and
tolerant of trailing commas. Requirements are to address a nested path, upsert
into an array by element name, and preserve comments and formatting.

- The regex helper cannot address nested paths at all.
- `encoding/json` round-tripping destroys comments and reflows the file.
- `tidwall/gjson` and `tidwall/sjson` do paths and array indices but have no
  comment support, so they would mangle the default file.
- `tailscale/hujson` parses JWCC into a mutable tree and re-serializes with
  formatting and comments intact.

Keeping both helpers was the earlier plan. Unifying is better: setting a
top-level key is the one-segment case of a path set, so there is nothing for the
regex version to do that the new helper does not do more correctly. Two
mechanisms for one job would drift.

The migration also fixes two live bugs in `zed` and `vscode`. The regex binds to
the first occurrence of a key anywhere in the file, including inside a nested
object, and when the existing value is not a string it fails to match and
appends a second copy of the key at the end. Zed's object form,
`"theme": { "mode": "system", "light": ..., "dark": ... }`, hits the second case
exactly.

### Build tags for the Windows-only gate

`internal/integration/windows_terminal_windows.go` registers the integration
through an `init()` that adds to the `available` map, and a Windows-tagged file
adds `"windows-terminal"` to `defaultSettings().Integrations`.

Both are needed. Registering without the default-list entry means Windows users
have to opt in by hand; adding to the default list without a Windows-tagged
registration means `themectl doctor` on macOS prints
`windows-terminal  unknown - enabled but not registered` in red.

`theme.Spec`, `theme.Resolved` and `config.Settings` are deliberately not
tagged. They feed `cmd/genschema`, and the committed schemas must not differ per
platform, or a theme authored on macOS could not declare Windows Terminal
support.

Alternative considered: the settings-level per-OS integration filter from the
README roadmap. That is a larger, more general feature and this change should
not block on it. It can subsume the build tags later.

### The registry keeps two name lists

Build-tagging the registry alone is not enough to keep the schemas portable.
`cmd/genschema` builds the `integrations` enum from the registry, so a tagged
registration would make the committed enum depend on the OS that generated it.

The registry therefore exposes two lists:

- `Names()` returns what is registered on this platform. `themectl doctor` uses
  it, so an integration that cannot run here never appears.
- `AllNames()` returns every known integration, adding the `platformOnly` names
  that this build did not register. `cmd/genschema` uses it, so the schemas are
  identical everywhere.

`Unknown()` also moves to `AllNames()`. A name that is known but not built for
this platform is inactive, not unknown, so a settings file shared between a
Windows and a macOS machine does not light up red on one of them.

Alternative considered: hardcode the enum in `cmd/genschema`. Smaller, but it
silently rots the next time an integration is added.

### Default settings path

`%LOCALAPPDATA%\Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState\settings.json`,
overridable with `[windows-terminal] config_file`.

The Store install is the common case on Windows 11. Unpackaged, Preview and
portable installs set `config_file`, and `Check()` already surfaces a missing
directory through `themectl doctor`, the same way `ghostty` and `rio` do. This
matches the codebase, which does no path probing anywhere.

## Risks / Trade-offs

- **The zed and vscode migration touches working integrations on every
  platform** → It is the widest-blast-radius part of this change, and unlike the
  Windows Terminal code it cannot hide behind a build tag. Port the existing
  `jsonc_test.go` cases onto the new helper before deleting the old one, so the
  migration is proven against the behavior it replaces, then add cases for the
  two bugs it fixes.
- **The chrome upsert writes into a user-owned array on every apply** → It
  matches strictly on the themectl-owned name, appends when absent, and hujson
  leaves the rest of the document byte-stable. Covered by a test asserting
  user-defined chrome definitions survive repeated applies.
- **Unpackaged and Preview users get an unhealthy integration until they set
  `config_file`** → `themectl doctor` names the missing directory, and the
  integration is skipped rather than failing the apply. Documented in the README.
- **New dependency** → hujson is small, has no transitive dependencies of note,
  and is maintained by Tailscale for exactly this use.
- **No Windows CI** → the platform-specific code is compiled behind build tags
  and cannot be exercised on the maintainer's macOS host. The hujson helper, the
  asset renaming and path resolution get cross-platform tests; the integration
  itself needs a manual pass on Windows.
