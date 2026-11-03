# themectl

Go CLI that applies a `family/variant` theme to many tools (integrations) in one pass.

## Commands

Run tasks with `mise run <task>`.

- `check` - tests and `go vet`, run before finishing
- `schema:gen` - regenerate `schemas/` from Go types
- `schema:check` - fail on schema drift (CI runs this)
- `fmt` - `gofmt -w .`

## Layout

- `cmd/themectl` - entry point
- `cmd/genschema` - JSON schema generator
- `internal/cli` - commands (urfave/cli v3)
- `internal/integration` - one file per integration, registered in `registry.go`
- `internal/theme` - `theme.toml` parsing and variant resolution
- `internal/config` - `themectl.toml` settings
- `openspec/` - specs and change proposals

## Conventions

- Platform code uses `_windows.go` / `_darwin.go` / `_other.go` files, not runtime checks.
- Changed theme or settings structs: run `schema:gen` and commit `schemas/`.
- User-facing text follows `openspec/specs/user-facing-copy/spec.md`.
- Changed user-facing behavior: update `README.md`.
