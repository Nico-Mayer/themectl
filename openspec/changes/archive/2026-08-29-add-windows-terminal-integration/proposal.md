## Why

themectl already applies themes to Ghostty and Rio, but has no Windows-native
terminal integration. Windows Terminal is the default terminal on Windows 11,
and on that OS it is the only terminal most users have. Without it, a Windows
user running `themectl set` sees their editor, wallpaper and system appearance
change while their terminal stays on the old colors.

## What Changes

- New `windows-terminal` integration, registered and enabled on Windows only via
  build tags. It never appears in `themectl doctor` on other platforms.
- Themes declare Windows Terminal support with a `[<...>.windows-terminal]`
  section carrying two optional URLs: `scheme_url` for the terminal color scheme
  and `theme_url` for the window chrome. Windows Terminal treats these as
  separate concepts and upstream theme ports ship them as separate files, so
  themectl mirrors that split. The bundled asset names are
  `windows-terminal.json` and `windows-terminal-theme.json`.
- On apply the integration writes the color scheme as a JSON fragment, points
  `profiles.defaults.colorScheme` at it, and installs the chrome definition into
  the `themes` array of `settings.json`. The chrome carries its own light or
  dark mode, so the integration does not map `appearance` itself.
- The chrome asset is optional. A theme may ship a color scheme alone.
- When the current theme declares no Windows Terminal scheme, the integration
  resets: it clears the scheme and chrome selection and removes what it owns,
  mirroring the existing `Rio.Reset` behavior.
- `[windows-terminal] config_file` in `themectl.toml` overrides the settings
  file path, for unpackaged, Preview and portable installs.
- **BREAKING (internal only)** `setJSONCString` is removed. All JSON config
  editing moves to a single `github.com/tailscale/hujson` backed helper, and the
  `zed` and `vscode` integrations migrate onto it. The regex helper cannot
  address nested paths, binds to the first match anywhere in a file, and appends
  a duplicate key when the existing value is not a string. No user-facing
  setting or theme field changes.

## Capabilities

### New Capabilities

- `integrations/windows-terminal`: applying a theme's color scheme and window
  chrome to Windows Terminal, path resolution for its settings file, and reset
  behavior when a theme does not support it.

### Modified Capabilities

None. No existing capability has a spec under `openspec/specs/` yet. The `zed`
and `vscode` migration is a refactor plus a bug fix: the observable contract for
those integrations, setting a named key in a JSON settings file, is unchanged.

## Impact

- New: `internal/integration/windows_terminal_windows.go`, and a hujson-backed
  JSONC helper replacing `internal/integration/jsonc.go`.
- Modified: `internal/integration/zed.go` and `internal/integration/vscode.go`
  move onto the new helper; `internal/integration/registry.go` and
  `internal/config/settings.go` gain a Windows-tagged registration and
  default-list entry; `internal/theme/spec.go` and `internal/theme/resolve.go`
  gain a `WindowsTerminalSpec` with two URLs and two asset names.
- Removed: `internal/integration/jsonc.go` and `jsonc_test.go`.
- Both generated schemas (`schemas/theme.schema.json`,
  `schemas/settings.schema.json`) gain the new sections. The spec and settings
  structs are deliberately not build-tagged, so theme files and schema output
  stay identical on every platform.
- New direct dependency: `github.com/tailscale/hujson`.
- `README.md` integration and settings documentation. The "Other terminal
  emulators" roadmap entry stays: it covers emulators beyond this one.
- Out of scope: probing for the settings file across packaged and unpackaged
  install locations, and per-profile theming.
