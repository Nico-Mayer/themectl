# User-facing Copy Specification

## Purpose

Defines consistent, accessible, and actionable human-readable text across `themectl` command help, interactive flows, status output, and failure recovery.

## Requirements

### Requirement: Consistent product voice
Human-readable product text SHALL use a clear, direct, concise, and calm voice, with consistent terminology and sentence case for equivalent interface elements.

#### Scenario: Equivalent concepts appear across commands
- **WHEN** multiple commands refer to the same theme, family, appearance, integration, or action
- **THEN** they use the same user-facing term and capitalization for that concept

#### Scenario: User scans command help
- **WHEN** `themectl` renders command, argument, or flag help
- **THEN** each description states its purpose in plain language without spelling errors, redundant wording, or unexplained internal terminology

### Requirement: Actionable failure messages
A human-readable error or warning SHALL identify what failed or what needs attention and SHALL provide a concrete recovery action when the program knows one.

#### Scenario: Failure has a known recovery command
- **WHEN** an operation fails and a `themectl` command can help the user recover
- **THEN** the message names the failed operation and includes the relevant command using its exact syntax

#### Scenario: Failure has no known automatic recovery
- **WHEN** an operation fails and no reliable recovery action is known
- **THEN** the message identifies the operation and preserves useful cause or context without inventing guidance

#### Scenario: User input does not meet a requirement
- **WHEN** the CLI rejects an argument or incompatible flag combination
- **THEN** the message states the requirement or conflict without blaming the user

### Requirement: Clear interactive decisions
Interactive prompts SHALL state the decision being requested and SHALL disclose material consequences before a destructive or overwrite action is confirmed.

#### Scenario: Local changes may be overwritten
- **WHEN** an update would replace a theme family containing local changes
- **THEN** the confirmation names that family and clearly states that continuing may discard its local changes

#### Scenario: User selects an item
- **WHEN** an omitted argument causes `themectl` to open a selector
- **THEN** the selector title identifies the kind of item being selected

### Requirement: Understandable operation feedback
Progress, success, skip, warning, and status text SHALL name the relevant action or state and SHALL include the affected object when needed to distinguish results.

#### Scenario: Operation completes successfully
- **WHEN** a user-triggered mutation completes
- **THEN** human-readable feedback confirms the completed action and identifies the affected theme, family, file, or resource when useful

#### Scenario: Operation is skipped
- **WHEN** `themectl` intentionally skips work
- **THEN** human-readable feedback states what was skipped and why

#### Scenario: Status uses visual symbols
- **WHEN** status output uses color or symbols to distinguish states
- **THEN** adjacent text also communicates each state without requiring color or symbol interpretation

### Requirement: Machine-readable compatibility
Copy changes SHALL NOT alter established machine-readable interfaces unless a separate capability explicitly specifies that change.

#### Scenario: JSON output accompanies revised human text
- **WHEN** copy is revised for a command that supports JSON output
- **THEN** its JSON field names, value meanings, and serialization remain unchanged

#### Scenario: Automation invokes existing syntax
- **WHEN** a script invokes an existing command, alias, argument, or flag
- **THEN** copy standardization does not change that syntax or its exit behavior

### Requirement: Future copy follows the same standard
New or changed human-readable interface text SHALL be reviewed against the same voice, terminology, actionability, accessibility, and compatibility requirements.

#### Scenario: New command text is introduced
- **WHEN** a future change adds command help, prompts, status text, progress feedback, warnings, or errors
- **THEN** its acceptance checks cover applicable requirements from this capability
