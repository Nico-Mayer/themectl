## Why

`themectl` exposes inconsistent user-facing copy across command help, prompts, status output, progress messages, and errors. Auditing and standardizing this text will make the CLI easier to scan, understand, and recover from while establishing a clear baseline for future commands.

## What Changes

- Audit existing human-readable CLI and interactive text for clarity, consistency, actionability, and accessibility.
- Standardize product voice as clear, direct, concise, and calm for developer-tool users.
- Define consistent patterns for command and flag help, prompts, progress and success messages, warnings, status labels, and errors.
- Give recoverable failures a specific cause or context and a useful next action when one is available.
- Use consistent terminology, capitalization, and command references across user-facing output.
- Preserve command names, flags, arguments, exit behavior, structured JSON field names, and machine-readable output contracts.
- Keep verbose diagnostic logs technically precise; apply UX-writing rules when those logs communicate decisions or recovery information to users.

## Capabilities

### New Capabilities

- `user-facing-copy`: Defines quality, consistency, recovery, and compatibility requirements for human-readable CLI and interactive text.

### Modified Capabilities

None.

## Impact

- Human-readable strings throughout `internal/cli`, `internal/ui`, and user-visible errors propagated from supporting packages may change.
- Tests that assert rendered help, prompts, logs, status output, or errors will need corresponding updates or additions.
- README command examples may need alignment when they quote changed output or terminology.
- No new runtime dependency is expected.
- Machine-readable JSON, command syntax, and automation-facing behavior remain compatible.
