# Contract — `msgs/msgs.go` additions and removals

**Feature**: 006-ux-polish-pack
**Scope**: define the exact shape (Go source-level) and semantics of each new message type introduced by this feature, plus the ones being retired.

> This is the authoritative contract. Implementation (Phase 3) must match these signatures byte-for-byte; any deviation requires amending this document and re-validating acceptance scenarios in `spec.md`.

---

## 1. New messages

### 1.1 `ClickFocusMsg`

```go
// ClickFocusMsg is emitted by the app layer after hit-testing a mouse press
// event against the current layout.Tree. It carries the identity of the pane
// the click landed in plus the (x, y) coordinates translated into that pane's
// local content-cell space (0,0 = first cell inside the pane's border).
//
// Emission: app.Update on MouseMsg{Action: MouseActionPress, Button: MouseButtonLeft}.
// Consumers: layout (to switch focus), terminal / sidebar (as pass-through).
type ClickFocusMsg struct {
    PaneID int
    Target FocusTarget
    LocalX int
    LocalY int
}
```

- **Invariant**: `Target ∈ {FocusTerminal, FocusSidebar, FocusLayout}` — `FocusEditor` is removed (see §3).
- **Latency budget**: emission + layout focus switch + re-render under 16ms (one frame).

### 1.2 `SidebarCreateMsg`

```go
// SidebarCreateMsg confirms the user has typed a name in the sidebar create
// prompt and pressed Enter. The sidebar component handles the actual filesystem
// mutation in its own Update; this message exists primarily for test feed-in
// and for cross-component observability (e.g. integration tests).
//
// Emission: components/sidebar when the user confirms the create prompt.
// Consumers: sidebar (itself, for state-machine clarity); tests.
type SidebarCreateMsg struct {
    Kind      string // "dir" or "file"
    Name      string // raw user input, already trimmed
    ParentDir string // absolute path where the creation will happen
}
```

- **Invariant**: `Kind` is exactly `"dir"` or `"file"` — no other values.
- **Validation** (in receiver): reject empty `Name`, names containing `/` or `\0`, names already existing in `ParentDir`.

### 1.3 `SidebarCreatedMsg`

```go
// SidebarCreatedMsg announces a successful filesystem creation via the sidebar.
// If Kind == "file", the app layer reacts by opening the file in the external
// editor (emitting OpenInExternalEditorMsg). If Kind == "dir", the sidebar
// itself enters the new directory.
//
// Emission: components/sidebar on successful os.Mkdir / os.WriteFile.
// Consumers: sidebar (navigation), app (route to editor).
type SidebarCreatedMsg struct {
    Kind string // "dir" or "file"
    Path string // absolute path of the newly created entry
}
```

### 1.4 `OpenInExternalEditorMsg`

```go
// OpenInExternalEditorMsg requests that the app spawn a terminal pane running
// the configured external editor (cfg.Editor) with the given file path as its
// single argument. Replaces the prior OpenFileMsg + editor.Model flow.
//
// Emission: sidebar (Enter on file), SidebarCreatedMsg handler (after file create).
// Consumer: app (PTY spawn).
type OpenInExternalEditorMsg struct {
    Path string
}
```

- **Fallback behaviour**: if `exec.LookPath(cfg.Editor)` fails, the app emits a
  `StatusBarNotifyMsg{Level: NotifyError, Text: "editor '<cfg.Editor>' não encontrado no PATH"}`
  and does **not** create a pane.

### 1.5 `ClockTickMsg`

```go
// ClockTickMsg is emitted by a 30s tea.Tick scheduled in statusbar.Init.
// The statusbar updates its displayed HH:MM and re-arms the tick.
//
// Emission: statusbar (self-scheduled).
// Consumer: statusbar only.
type ClockTickMsg struct {
    Now time.Time
}
```

- **Cadence**: every 30 seconds. First emission is immediate (at `Init`).

### 1.6 `PaneCWDChangeMsg`

```go
// PaneCWDChangeMsg is emitted by a terminal pane when its stream-side OSC 7
// parser detects a current-working-directory announcement from the child shell.
// Format expected: "\x1b]7;file://<host>/<abs-path>\x07" (zsh/bash precmd hook).
//
// Emission: components/terminal on OSC 7 decode success.
// Consumers: layout (aggregation), statusbar (when focused pane).
type PaneCWDChangeMsg struct {
    PaneID int
    CWD    string // absolute path, percent-decoded
}
```

### 1.7 `PaneGitStateMsg`

```go
// PaneGitStateMsg carries the result of a background git query for a pane
// (triggered by PaneCWDChangeMsg). Branch is empty when the CWD is not a git
// repository; Dirty is false in that case as well.
//
// Emission: terminal (tea.Cmd that execs `git -C <cwd> …`).
// Consumer: layout (consolidation), statusbar (when focused pane).
type PaneGitStateMsg struct {
    PaneID int
    Branch string
    Dirty  bool
}
```

- **Timeout**: the underlying `exec.Command` must run with a `context.WithTimeout(ctx, 200*time.Millisecond)` — beyond that, the pane reports empty branch.

### 1.8 `FocusedPaneContextMsg`

```go
// FocusedPaneContextMsg reaches the statusbar whenever the focused pane's
// identity or context changes. It is the ONLY channel through which the
// statusbar learns about branch/CWD — the global MetricsTickMsg stops
// carrying these fields.
//
// Emission: layout, after any of: FocusChangeMsg, PaneCWDChangeMsg (for the
// focused pane), PaneGitStateMsg (for the focused pane).
// Consumer: statusbar.
type FocusedPaneContextMsg struct {
    PaneID    int
    CWD       string // empty if unknown
    GitBranch string // empty if not a git repo
    GitDirty  bool
}
```

---

## 2. Altered messages

### 2.1 `MetricsTickMsg` (trimmed)

The `CWD` and `GitBranch` fields are **removed** from `MetricsTickMsg`. The ticker continues to emit CPU/memory/disk metrics only. All git/CWD awareness moves to the per-pane flow described above.

```go
// Before (to be removed):
// type MetricsTickMsg struct {
//     CPU       float64
//     MemUsed   uint64
//     MemTotal  uint64
//     CWD       string   // <-- REMOVE
//     GitBranch string   // <-- REMOVE
//     Tick      time.Time
// }

// After:
type MetricsTickMsg struct {
    CPU      float64
    MemUsed  uint64
    MemTotal uint64
    Tick     time.Time
}
```

---

## 3. Removed messages and enum values

- `EditorResizeMsg` — the embedded editor is gone.
- `ConfirmCloseMsg` — no longer needed (no unsaved-changes dialog in external editor flow).
- `CloseConfirmedMsg` — idem.
- `CloseAbortedMsg` — idem.
- `FocusTarget.FocusEditor` enum value — remove; renumber remaining constants if necessary, but prefer keeping explicit `iota` ordering stable by dropping only the last-added value.

Any `switch` statement in the codebase that handled the removed cases MUST be updated — detection is via `go build ./...` (Go's exhaustiveness is not enforced by the compiler, but test coverage must catch any surviving references).

---

## 4. Backwards-compatibility note

The messages removed here are internal to the Lumina binary; there are no external consumers. Per the constitution (§Development Workflow), a MINOR version bump of the constitution is **not** required, because:

1. `msgs/msgs.go` is not a published module surface.
2. The removals are paired with a single atomic PR (feature 006) that updates every call site.
3. No downstream project depends on these message names.

The removal is documented in `DECISIONS.md` alongside the rationale from `research.md` §R9.
