## Purpose

Applies a themectl theme to Windows Terminal, covering the terminal color
scheme, the window chrome that surrounds it, where the settings file is found,
and what happens when the active theme has no Windows Terminal support.

## ADDED Requirements

### Requirement: Windows-only availability

The `windows-terminal` integration SHALL be registered and enabled only on
Windows. On every other platform it SHALL be absent from the integration
registry and from the default integration list, so it is neither reported by
`themectl doctor` nor flagged as an unknown integration.

Theme files and both generated JSON schemas SHALL be identical on every
platform, so a theme authored on macOS or Linux can declare Windows Terminal
support without warnings.

#### Scenario: Doctor on a non-Windows host

- **WHEN** `themectl doctor` runs on macOS or Linux
- **THEN** no `windows-terminal` row appears in the integrations section
- **AND** no unknown-integration warning is emitted for it

#### Scenario: Doctor on Windows

- **WHEN** `themectl doctor` runs on Windows with default settings
- **THEN** a `windows-terminal` row appears and is reported as enabled

#### Scenario: Theme authored off Windows

- **WHEN** a theme file declaring a `windows-terminal` section is validated
  against `schemas/theme.schema.json` on any platform
- **THEN** validation succeeds

### Requirement: Theme declares a scheme and an optional chrome

A theme SHALL declare Windows Terminal support with a `windows-terminal` section
holding two independent assets, because Windows Terminal treats the terminal
colors and the window chrome as separate concepts:

- the color scheme, canonical asset name `windows-terminal.json`, declarable
  remotely with `scheme_url`
- the window chrome, canonical asset name `windows-terminal-theme.json`,
  declarable remotely with `theme_url`

Each asset follows the same bundled-asset and remote-asset rules as every other
asset-based integration: a URL, when set, is the source of truth and shadows a
bundled file of the same canonical name.

The color scheme SHALL be required and the chrome SHALL be optional. The
integration SHALL report that it supports the current theme only when the color
scheme asset is present in the materialized theme directory.

#### Scenario: Bundled assets

- **WHEN** a theme ships `windows-terminal.json` and
  `windows-terminal-theme.json` next to its `theme.toml`
- **THEN** the integration reports the theme as supported and both assets are
  used

#### Scenario: Remote assets

- **WHEN** a theme sets `scheme_url` and `theme_url`
- **THEN** each file is fetched and materialized under its canonical asset name
- **AND** a bundled file of the same canonical name is ignored with a warning

#### Scenario: Scheme without chrome

- **WHEN** a theme provides only the color scheme asset
- **THEN** the integration reports the theme as supported
- **AND** the color scheme is applied without any chrome being installed

#### Scenario: Chrome without scheme

- **WHEN** a theme provides only the chrome asset
- **THEN** the integration reports the theme as unsupported

#### Scenario: No Windows Terminal support

- **WHEN** a theme declares no `windows-terminal` section and ships neither
  asset
- **THEN** the integration reports the theme as unsupported

### Requirement: Settings file location

The integration SHALL default to the packaged Windows Terminal settings file at
`%LOCALAPPDATA%\Packages\Microsoft.WindowsTerminal_8wekyb3d8bbwe\LocalState\settings.json`.

A `config_file` value under a `[windows-terminal]` table in `themectl.toml`
SHALL override that path, expanding environment variables and a leading `~` the
same way every other file-editing integration does.

The integration SHALL report itself unhealthy when the directory holding the
resolved settings file does not exist, and SHALL be skipped for that run rather
than creating the file.

#### Scenario: Default path on a Store install

- **WHEN** no `config_file` is configured
- **THEN** the packaged settings path is used

#### Scenario: Overridden path

- **WHEN** `[windows-terminal] config_file` is set in `themectl.toml`
- **THEN** that path is used instead of the default

#### Scenario: Windows Terminal not installed

- **WHEN** the directory holding the resolved settings file does not exist
- **THEN** `themectl doctor` reports the integration as unhealthy with the
  missing directory as the detail
- **AND** applying a theme skips the integration without an error

### Requirement: Applying the color scheme

On apply the integration SHALL make the theme's color scheme available to
Windows Terminal under a fixed themectl-owned name, regardless of the name
carried by the theme's asset, and select it for every profile that does not
override it by setting the color scheme on the profile defaults.

A profile that sets its own color scheme SHALL keep it.

#### Scenario: Scheme applied

- **WHEN** a theme with Windows Terminal support is applied
- **THEN** the theme's colors are available to Windows Terminal under the
  themectl-owned scheme name
- **AND** the profile defaults select that scheme

#### Scenario: Asset name is irrelevant

- **WHEN** the theme's scheme asset declares its own name, such as
  `Catppuccin Mocha`
- **THEN** it is still installed under the themectl-owned name, so the theme
  author never has to match a name in two places

#### Scenario: Profile-level override survives

- **WHEN** a profile sets its own color scheme and a theme is applied
- **THEN** that profile keeps its own scheme

#### Scenario: Scheme reaches a running terminal

- **WHEN** a theme is applied while Windows Terminal is running
- **THEN** the color scheme is in place before the settings file is written, so
  the terminal's own live reload picks up both together

### Requirement: Applying the window chrome

When the theme provides a chrome asset, the integration SHALL install it under
the same fixed themectl-owned name and select it, so the tab row, tab
backgrounds and window mode match the terminal colors.

The light or dark window mode SHALL come from the chrome asset itself. The
integration SHALL NOT derive it from the theme's `appearance`, because the
chrome asset already carries it and duplicating that decision could contradict
the asset.

Installing the chrome SHALL be idempotent: repeated applies SHALL leave exactly
one themectl-owned chrome definition, and SHALL leave user-defined chrome
definitions untouched.

When the theme provides no chrome asset, the integration SHALL remove its own
chrome definition and clear the chrome selection rather than leaving the
previous theme's chrome in place.

#### Scenario: Chrome applied

- **WHEN** a theme providing a chrome asset is applied
- **THEN** the chrome is installed under the themectl-owned name and selected
- **AND** the tab and tab row backgrounds come from that asset

#### Scenario: Window mode comes from the asset

- **WHEN** a chrome asset declares a dark window mode
- **THEN** Windows Terminal uses dark window chrome, whatever the theme's
  `appearance` says

#### Scenario: Repeated applies

- **WHEN** several themes providing chrome are applied in sequence
- **THEN** exactly one themectl-owned chrome definition exists
- **AND** user-defined chrome definitions are unchanged

#### Scenario: Switching to a theme without chrome

- **WHEN** a theme providing only a color scheme is applied after one that
  provided chrome
- **THEN** the themectl-owned chrome definition is removed and the chrome
  selection is cleared
- **AND** the new color scheme is still applied

### Requirement: User content in the settings file is preserved

Applying a theme SHALL NOT remove, reorder, or rewrite color schemes, profiles,
chrome definitions, or any other settings the user owns, and SHALL preserve the
comments, trailing commas and formatting of the settings file. Windows Terminal
ships its settings file with comments, so an edit that strips them is a
regression the user sees.

#### Scenario: Comments and formatting survive

- **WHEN** the settings file contains comments, trailing commas and irregular
  indentation, and a theme is applied
- **THEN** every part of the file the integration does not own is byte-identical
  afterwards

#### Scenario: User definitions survive

- **WHEN** the settings file contains user-defined color schemes, profiles and
  chrome definitions, and a theme is applied
- **THEN** all of them are unchanged

### Requirement: Reset when the theme has no Windows Terminal support

When the active theme does not support the integration, the integration SHALL
reset: it SHALL clear the themectl color scheme and chrome selections and remove
the color scheme and chrome definition it owns, leaving Windows Terminal on its
own defaults.

Reset SHALL NOT fail a theme apply. A reset that cannot complete SHALL be
reported as a warning.

#### Scenario: Switching to a theme without Windows Terminal support

- **WHEN** a theme with no Windows Terminal support is applied after one that
  had it
- **THEN** the profile defaults no longer select the themectl scheme, and the
  chrome selection is cleared
- **AND** the themectl-owned color scheme and chrome definition are removed

#### Scenario: Reset fails

- **WHEN** reset cannot complete, for example because the settings file is
  unreadable
- **THEN** a warning is logged
- **AND** the theme apply as a whole still succeeds
