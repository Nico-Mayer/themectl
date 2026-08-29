## 1. Unified JSONC editing helper

- [x] 1.1 Add `github.com/tailscale/hujson` to go.mod and verify `go mod tidy` and `go build ./...` succeed
- [x] 1.2 Add `internal/integration/hujson.go` with a nested-path string setter that creates missing intermediate objects; verify a unit test sets `profiles.defaults.colorScheme` in a fixture that also has `colorScheme` inside `profiles.list` and only the defaults entry changes
- [x] 1.3 Add a path deleter to the same helper for clearing a key during reset; verify a unit test that deleting an absent path is a no-op and deleting a present one leaves the surrounding object intact
- [x] 1.4 Add an array upsert-by-name helper that appends an object when no element carries the given `name` and replaces it otherwise, plus a matching remove-by-name; verify unit tests that two upserts yield exactly one matching element and that removal leaves other elements untouched
- [x] 1.5 Verify a unit test that round-tripping a settings fixture with `//` comments, trailing commas and irregular indentation through every helper leaves each untouched byte identical
- [x] 1.6 Port the existing `jsonc_test.go` cases onto the new helper so the replacement is proven against the behavior it supersedes, and verify they pass
- [x] 1.7 Migrate `zed.go` and `vscode.go` off `setJSONCString`, then delete `jsonc.go` and `jsonc_test.go`; verify `rg setJSONCString` returns nothing and the existing zed and vscode tests pass
- [x] 1.8 Add regression tests for the two bugs the migration fixes: a nested duplicate of the target key elsewhere in the file is not written to, and an object-valued `theme` in a zed settings file is replaced rather than having a second `theme` key appended

## 2. Theme spec and asset plumbing

- [x] 2.1 Add `theme.WindowsTerminalSpec` with `scheme_url` and `theme_url`, wire it into `theme.Spec` and `theme.Resolved` and through `Resolve`; verify the existing `internal/theme` tests pass and a new case covers per-field defaults-to-variant inheritance across both URLs
- [x] 2.2 Add `WindowsTerminalAssetName = "windows-terminal.json"` and `WindowsTerminalThemeAssetName = "windows-terminal-theme.json"` with their entries in `RemoteAssets()`; verify a unit test that a spec setting only `scheme_url` yields one remote asset and setting both yields two
- [x] 2.3 Add `WindowsTerminal FileSettings` to `config.Settings`; verify a settings test that `[windows-terminal] config_file` parses and expands `~` and env vars
- [x] 2.4 Split the registry into platform-gated `Names()` and portable `AllNames()`, point `cmd/genschema` and `Unknown()` at the latter, then regenerate both schemas and verify `schemas/theme.schema.json` and `schemas/settings.schema.json` contain the new sections and the diff contains nothing else

## 3. Windows Terminal integration

- [x] 3.1 Add `internal/integration/windows_terminal_windows.go` with the `WindowsTerminal` type implementing `Integration` and `Resetter`; verify `GOOS=windows go build ./...` succeeds
- [x] 3.2 Implement path resolution: default to the packaged `LocalState` settings path, honor `config_file`, and derive the fragment path from `%LOCALAPPDATA%\Microsoft\Windows Terminal\Fragments\themectl`; verify a unit test covering the default, the override and a missing `LOCALAPPDATA`
- [x] 3.3 Implement `Supports` as an existence check on the materialized `windows-terminal.json` alone, and `Check` as a `checkConfigDir` on the settings file's directory; verify unit tests for scheme present, scheme absent with chrome present, and a missing settings directory
- [x] 3.4 Implement the asset rename that sets an asset's `name` to `themectl`, shared by both assets, and the fragment wrapper `{"schemes": [ ... ]}`; verify a cross-platform unit test against the real `mocha.json` and `mochaTheme.json` from `catppuccin/windows-terminal` as fixtures
- [x] 3.5 Implement the scheme half of `Apply`: write the fragment first, then set `profiles.defaults.colorScheme`; verify a unit test on a settings fixture asserting the key is set and that user schemes, profiles and comments are unchanged
- [x] 3.6 Implement the chrome half of `Apply`: upsert the renamed chrome asset into `themes[]` and set the top-level `theme` key, and when no chrome asset exists remove the themectl entry and clear the key; verify unit tests for chrome present, chrome absent, and a repeated apply leaving exactly one themectl entry with user-defined entries intact
- [x] 3.7 Implement `Reset`: delete the fragment, clear `profiles.defaults.colorScheme` and the top-level `theme` key, and remove the themectl chrome entry, tolerating an already-absent fragment; verify a unit test that reset after apply returns the settings fixture to its pre-apply content and removes the file

## 4. Registration and Windows-only gating

- [x] 4.1 Register `windows-terminal` in the `available` map from a Windows-tagged file; verify `integration.Names()` includes it under `GOOS=windows` and excludes it under `GOOS=darwin`
- [x] 4.2 Add `windows-terminal` to `defaultSettings().Integrations` from a Windows-tagged file; verify `integration.Unknown()` returns nothing for default settings on both `GOOS=windows` and `GOOS=darwin`
- [x] 4.3 Verify `go vet ./...` and the full test suite pass on darwin, and that `GOOS=windows go vet ./...` passes

## 5. Documentation

- [x] 5.1 Document the `windows-terminal` theme section with `scheme_url` and `theme_url`, using the catppuccin URLs as the worked example, and the `[windows-terminal] config_file` setting; verify the README examples match the regenerated schemas and note that unpackaged, Preview and portable installs must set `config_file`
- [x] 5.2 Leave the "Other terminal emulators" roadmap entry in place; verify it still reads correctly now that one emulator is covered, and reword only if it implies none are supported

## 6. Manual verification on Windows

- [x] 6.1 Apply a theme using the catppuccin mocha scheme and chrome URLs on a Store install with the terminal running, and verify the terminal colors change without a restart
- [x] 6.2 Confirmed: the fragment written in the same apply is picked up by the `settings.json` live reload, no restart needed. Recorded in design.md; no README change required
- [x] 6.3 Verify the tab and tab row backgrounds match the chrome asset, including the unfocused tab row color, and that switching between catppuccin latte and mocha flips the window mode
- [x] 6.4 Apply a theme providing only a scheme and verify the chrome entry is removed and the window mode falls back to Windows Terminal's own default
- [x] 6.5 Apply a theme without Windows Terminal support and verify the scheme selection, chrome entry and fragment are all removed
- [x] 6.6 Verify `themectl doctor` on a machine without Windows Terminal reports the integration as unhealthy with the missing directory, and that `themectl set` still succeeds
