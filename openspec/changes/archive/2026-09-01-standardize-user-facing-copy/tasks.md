## 1. Copy Standard and Safety Net

- [x] 1.1 Use `specs/user-facing-copy/spec.md` as the project-specific contract and the UX-writing skill as general guidance; verify no duplicate repository copy guide remains
- [x] 1.2 Inventory human-readable strings in `internal/cli`, `internal/ui`, and surfaced lower-level errors against command discovery, input, progress, result, status, and recovery journeys; verify repository searches classify every discovered shipped string or explicitly exclude it as internal-only
- [x] 1.3 Add focused characterization tests for current command syntax, aliases, flags, exit behavior, and `current`/`doctor` JSON keys and value meanings; verify tests pass before prose changes

## 2. Command Discovery and Interactive Input

- [x] 2.1 Standardize app, command, argument, and flag help text for sentence case, concise action wording, canonical terminology, and spelling; verify the audited help surfaces follow the spec and UX-writing guidance without locking routine prose into exact-string tests
- [x] 2.2 Make shared selection UI honor its caller-provided title and give theme and theme-family flows distinct titles; verify focused tests or captured interactive output identify the selected object correctly
- [x] 2.3 Revise update and other consequential confirmations to name the affected object and disclose overwrite or deletion consequences before continuing; verify tests assert object identity and consequence wording

## 3. Progress, Results, and Status

- [x] 3.1 Standardize spinner, success, skip, warning, and verbose diagnostic messages using the documented patterns while retaining useful structured attributes; verify representative command tests distinguish completed, skipped, and failed outcomes with their reasons
- [x] 3.2 Revise human-readable `current` and `doctor` labels and guidance for consistent terminology and exact recovery commands; verify rendering tests cover unset, missing, unhealthy, unsupported, and unknown states
- [x] 3.3 Ensure every colored or symbolic status includes an equivalent text label; verify no tested `doctor` state depends on color or symbols for meaning

## 4. Errors and Recovery

- [x] 4.1 Audit errors reaching the CLI boundary and add lowercase `%w` operation context where the failing action would otherwise be unclear, without logging and returning the same error; verify `errors.Is` behavior and representative wrapped causes remain intact
- [x] 4.2 Revise known invalid-input and recoverable failures to state the requirement or exact recovery command without blame; verify table-driven tests cover missing arguments, incompatible appearance flags, absent current themes, missing wallpaper candidates, and install/uninstall failures with known recovery paths
- [x] 4.3 Review normal and verbose failure output to confirm normal text stays actionable while verbose diagnostics retain technical causes; verify manual failing-command runs show one handling path without duplicate error reports

## 5. Documentation and Final Validation

- [x] 5.1 Update README examples or descriptions that quote changed terminology or output; verify every documented command and message matches current CLI behavior
- [x] 5.2 Run `go fmt ./...`, `mise run check`, and targeted JSON compatibility tests; verify all commands pass without changing JSON fixtures or schemas
- [x] 5.3 Manually exercise help, omitted-argument selection, update confirmation, successful mutation, skipped update, `doctor`, and representative recovery errors in color and no-color terminal modes; verify observed text satisfies the UX-writing guidance and all spec scenarios
