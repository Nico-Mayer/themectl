## Context

User-facing text is distributed across `internal/cli`, `internal/ui`, and errors returned by lower-level packages. Surfaces include urfave command metadata, huh selectors and confirmations, spinner titles, slog messages, rendered status tables, and terminal errors. Copy currently mixes capitalization and grammatical styles, contains at least one spelling error, sometimes omits recovery guidance, and uses a hard-coded theme selector title even when selecting a family.

JSON report fields and command syntax are already automation interfaces. Human-readable revisions must remain separate from those contracts. See `proposal.md` and `specs/user-facing-copy/spec.md` for scope and required behavior.

## Goals / Non-Goals

**Goals:**

- Apply one clear, direct, concise, and calm voice across human-readable CLI surfaces.
- Make known recovery paths and material confirmation consequences explicit.
- Preserve useful technical causes while adding user-level operation context.
- Keep project-specific copy requirements in the OpenSpec capability and use the UX-writing skill for general writing guidance, without duplicating either in repository documentation.
- Add focused tests for important copy contracts without snapshotting every incidental word.

**Non-Goals:**

- Changing command names, aliases, flags, arguments, exit semantics, or JSON schemas.
- Replacing urfave, huh, slog, or terminal styling libraries.
- Building localization infrastructure or a centralized message catalog.
- Rewriting internal-only test failures, code comments, or debug messages that users do not rely on for decisions or recovery.
- Removing technical detail from verbose diagnostics.

## Decisions

### Audit copy by user journey and surface

Build an inventory grouped into command discovery, input and selection, in-progress feedback, completion and skip results, status inspection, and failure recovery. Review both direct strings in CLI/UI packages and lower-level errors that reach `main` unchanged.

This journey-based inventory is preferred over editing files independently because one action often crosses several packages and needs consistent terminology from help through completion. A blind global string rewrite was rejected because identical words can serve different contexts and lower-level errors may also be used programmatically.

### Use surface-specific writing patterns

Apply these patterns consistently:

- Command and flag help: concise sentence fragments that start with a specific action where natural; sentence case; no terminal period for short entries.
- Selector titles: name the object being selected, such as `Select a theme` or `Select a theme family`; shared UI helpers must honor caller-provided titles.
- Confirmations: identify the affected object, state the material consequence, then ask whether to continue.
- Spinner text: present-progress action, such as `Applying theme`.
- Success text: past-tense result plus affected object in structured attributes where already supported.
- Skip and warning text: affected operation or object, reason, then recovery action when known.
- Errors: operation context plus preserved cause; add an exact recovery command only when it is reliable.
- Status text: concise state labels accompanied by words, so color and symbols are supplementary.

A single grammatical form for every surface was rejected because help, progress, confirmation, and completion text serve different user needs.

### Keep copy near behavior and avoid duplicate guidance

Strings remain near their command or rendering logic. The OpenSpec capability is the project-specific behavioral contract, while the UX-writing skill supplies general writing methods and patterns. Do not add a repository copy guide that repeats those sources. Extract a helper or formatter only when several call sites share behavior as well as wording.

A separate contributor guide and a centralized constants or message-catalog package were rejected: both add maintenance surfaces without improving runtime behavior at the current scale. Localization remains possible later but is not a current requirement.

### Add context at CLI boundaries without flattening causes

Lower-level packages may retain concise technical errors. Command actions should wrap errors when users otherwise cannot identify the failed operation. Existing causes remain available through Go error wrapping and continue to appear in terminal output. Known invalid-input and recovery cases receive direct user-level messages.

Rewriting every lower-level error into polished prose was rejected because those packages need precise diagnostics and some errors are composed or inspected by callers.

### Separate human copy from machine contracts

Do not alter JSON struct tags, serialized value meanings, command syntax, or exit behavior while revising human-readable output. Tests should compare JSON structures independently from rendered terminal text. Styled status output may change words and layout only where machine consumers are not promised compatibility.

Treating all stdout as stable machine output was rejected because it would prevent improving interactive and terminal-oriented text. JSON remains the explicit structured interface where available.

### Test behavior-significant copy, not every sentence

Add focused tests for:

- command names, aliases, arguments, and flags as compatibility contracts, without asserting every help sentence;
- selectors honoring context-specific titles;
- destructive or overwrite confirmations disclosing consequences;
- representative invalid-input errors and known recovery commands;
- skip reasons and text alternatives for styled status states;
- unchanged JSON keys and representative values.

Assert exact words only when they carry behavior: consequences, recovery commands, accessibility labels, or machine contracts. Do not assert routine help prose whose wording can change independently of behavior. Full snapshots were rejected because valid copy and library-formatting changes would create noise unrelated to regressions.

## Risks / Trade-offs

- [Risk] Human-readable scripts may depend on current prose despite JSON being the supported structured interface. -> Preserve syntax and exit behavior, avoid gratuitous layout changes, and call out revised human output in release notes if needed.
- [Risk] Broad copy edits can accidentally hide useful technical context. -> Preserve wrapped causes and review normal and verbose output separately.
- [Risk] Subjective wording review can expand scope indefinitely. -> Limit implementation to inventoried product surfaces and validate each against explicit spec scenarios and the UX-writing skill.
- [Risk] Exact-string tests can make future copy improvements costly. -> Assert behavior-significant terms, consequences, recovery commands, and machine contracts rather than routine prose.
- [Risk] Spec and writing guidance can drift when duplicated. -> Keep project requirements in the capability and general guidance in the UX-writing skill; do not create a third source.

## Migration Plan

1. Record current command metadata, prompts, status labels, result messages, and surfaced errors in an audit checklist.
2. Add focused characterization tests for machine-readable contracts and behavior-significant user guidance.
3. Revise help and selection text, then prompts, operation feedback, status output, and failure messages in small groups.
4. Add or update tests with each group and align README examples that quote affected text.
5. Remove temporary or duplicate copy guidance, then run formatting, unit tests, static checks, and representative manual CLI flows.

Rollback consists of reverting human-copy commits. No persisted data or configuration migration is required.
