<!--
SYNC IMPACT REPORT
==================
Version change: (unversioned template) → 1.0.0
Initial ratification — all placeholders replaced.

Principles defined:
  I.  Code Quality
  II. Testing Standards
  III. User Experience Consistency
  IV. Performance Requirements

Added sections:
  - Core Principles (4 principles)
  - Technical Standards
  - Development Workflow
  - Governance

Templates reviewed:
  ✅ .specify/templates/plan-template.md — Constitution Check section already generic; no edits needed
  ✅ .specify/templates/spec-template.md — Success Criteria / Measurable Outcomes align with principles; no edits needed
  ✅ .specify/templates/tasks-template.md — Phase structure aligns with testing and quality gates; no edits needed

Deferred TODOs: none
-->

# Lumina Constitution

## Core Principles

### I. Code Quality

Every Go package in Lumina MUST have a single, well-defined responsibility. Files,
functions, and types MUST follow standard Go naming conventions (`gofmt`-compliant,
exported names PascalCase, unexported camelCase). Cyclomatic complexity per function
MUST NOT exceed 10; functions exceeding 40 lines of non-comment code MUST be refactored.
No global mutable state outside of the `config` package. Cross-component communication
MUST use explicit `tea.Msg` types defined in `msgs/msgs.go` — no implicit coupling or
shared channels between component packages.

**Rationale**: Lumina's Bubble Tea architecture (Elm Model/Update/View) scales only
when each component (`terminal`, `sidebar`, `statusbar`, `editor`) stays independently
comprehensible. God objects and implicit coupling collapse the message-routing model
in `app/app.go` into unmaintainable complexity.

### II. Testing Standards

Every exported function and `tea.Model` implementation MUST have at least one unit test.
Unit tests for a component MUST NOT depend on other components — test each `tea.Model`
in isolation by constructing it directly and feeding synthetic `tea.Msg` values.
Integration tests for cross-component message flows are REQUIRED for every new `tea.Msg`
type added to `msgs/msgs.go`. Tests MUST be written before implementation (TDD) for
bug fixes: reproduce the bug with a failing test first, then fix. `go test ./...` MUST
pass with zero failures before any pull request is merged.

**Rationale**: The Elm-style unidirectional flow makes unit testing straightforward —
`Update(msg)` is a pure function of `(Model, Msg) → (Model, Cmd)`. Integration tests
catch message-routing regressions in `app.go` that unit tests cannot see. Failing CI
on test breakage is the primary quality gate.

### III. User Experience Consistency

All keyboard shortcuts MUST be defined in `app/keymap.go` and MUST NOT be hardcoded
inside component files. Visual styles (colors, borders, padding) MUST use Lip Gloss
style definitions — raw ANSI escape codes are FORBIDDEN outside of the `pty` integration
layer. Focus state MUST be visible: the active panel MUST render a distinct border or
highlight using the shared style system. Key bindings MUST follow terminal-native
conventions (e.g., `Ctrl+C` exits, `Ctrl+Z` suspends, arrow keys navigate) unless a
documented rationale overrides them. Every user-facing error MUST be surfaced in the
status bar — never silently discarded.

**Rationale**: Users switch between Lumina panes and external terminals constantly.
Inconsistent shortcuts or style breaks trust and productivity. Centralizing keymaps in
one file prevents shadow-binding conflicts across components. Lip Gloss ensures styles
compose correctly across different terminal color profiles.

### IV. Performance Requirements

The TUI render loop MUST maintain ≥30 FPS under normal operation (no pane actively
rendering high-frequency output). Status bar metrics (CPU, memory, disk, network via
`gopsutil`) MUST be sampled on a background ticker with a minimum interval of 1 second
— never blocking the main Bubble Tea event loop. PTY resize events (`pty.Setsize`) MUST
be processed within 50ms of the terminal window resize signal. No single `Update`
function call MUST block for longer than 16ms; expensive I/O MUST be delegated to a
`tea.Cmd` goroutine. Memory allocations inside the render path (`View()` methods) MUST
be minimized — reuse buffers where possible via `strings.Builder`.

**Rationale**: A terminal editor that lags betrays its core promise of "produtividade e
leveza". The Bubble Tea event loop is single-threaded; any blocking `Update` stalls the
entire application. Delegating I/O to `tea.Cmd` keeps the loop free, which is the
idiomatic Bubble Tea contract.

## Technical Standards

- **Language**: Go (latest stable minor release — update within 30 days of a new minor release)
- **Formatter**: `gofmt` is non-negotiable; `golangci-lint` with default ruleset MUST pass on CI
- **Dependencies**: New external dependencies MUST be justified in `DECISIONS.md` before
  being added. Prefer standard library; prefer Charm ecosystem packages for TUI concerns
- **PTY support**: Linux and macOS only — `creack/pty` MUST NOT be replaced without a
  DECISIONS.md entry documenting the migration rationale
- **Build**: `go build ./...` MUST produce zero warnings. Binary MUST be a single static
  executable with no runtime dependency on external shared libraries

## Development Workflow

- All new features MUST originate from a spec (`/speckit.specify`) before implementation
- Pull requests MUST reference the spec and include passing tests
- Commits MUST be atomic: one logical change per commit, with a descriptive message
  in the imperative mood (`Add sidebar focus highlight`, not `added stuff`)
- Breaking changes to `msgs/msgs.go` message contracts MUST bump the constitution
  version (MINOR) and be documented in `DECISIONS.md`
- Code review MUST verify compliance with all four Core Principles before merge

## Governance

This constitution supersedes all other development practices and informal agreements.
Amendments require: (1) a written proposal describing the change and its motivation,
(2) documentation of any migration plan for existing code, and (3) an update to the
version line below.

**Versioning policy**:
- MAJOR — backward-incompatible removal or redefinition of a Core Principle
- MINOR — new principle, new mandatory section, or materially expanded guidance
- PATCH — clarifications, wording refinements, typo fixes

All pull requests and code reviews MUST verify compliance with the Core Principles.
Complexity violations MUST be explicitly justified in the Complexity Tracking table
of the feature's `plan.md` before the PR is approved.

**Version**: 1.1.0 | **Ratified**: 2026-04-16 | **Last Amended**: 2026-04-16
