package msgs

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// FocusTarget identifies which pane has keyboard focus.
type FocusTarget int

const (
	FocusTerminal FocusTarget = iota
	FocusSidebar
	FocusLayout // layout manager holds focus (content area with multiple panes)
)

// NotifyLevel controls the visual severity of a status bar notification.
type NotifyLevel int

const (
	NotifyInfo NotifyLevel = iota
	NotifyWarning
	NotifyError
)

// SplitDir defines the direction of a pane split.
type SplitDir int

const (
	SplitHorizontal SplitDir = iota // side by side
	SplitVertical                   // stacked
)

// FocusDir defines the direction of pane focus movement.
type FocusDir int

const (
	FocusDirLeft FocusDir = iota
	FocusDirRight
	FocusDirUp
	FocusDirDown
)

// ResizeDir defines whether a pane grows or shrinks.
type ResizeDir int

const (
	ResizeGrow ResizeDir = iota
	ResizeShrink
)

// ResizeAxis defines the axis of a resize operation.
type ResizeAxis int

const (
	ResizeAxisH ResizeAxis = iota // horizontal axis
	ResizeAxisV                   // vertical axis
)

// FocusChangeMsg is emitted by app when the user switches focus between panes.
type FocusChangeMsg struct {
	Target FocusTarget
}

// PaneFocusMsg is sent to a pane model to update its focused state.
type PaneFocusMsg struct {
	Focused bool
}

// PtyOutputMsg carries bytes read from the PTY process.
// PaneID identifies which terminal pane produced this output.
type PtyOutputMsg struct {
	PaneID int
	Data   []byte
	Err    error
}

// PtyInputMsg carries bytes to write to the PTY process.
// PaneID identifies which terminal pane receives this input.
type PtyInputMsg struct {
	PaneID int
	Data   []byte
}

// PtyMouseMsg routes a mouse event to a specific terminal pane that has
// requested mouse tracking. The Mouse field's X/Y are pane-local coordinates
// relative to the inside of the pane's border (0,0 = first content cell).
type PtyMouseMsg struct {
	PaneID int
	Mouse  tea.MouseMsg
}

// EnterCopyModeMsg requests that the focused terminal enter tmux-style copy
// mode (interactive selection + OSC52 clipboard).
type EnterCopyModeMsg struct{}

// PaneSplitMsg requests splitting the active pane. When Command is non-empty
// the new pane runs that command (under `sh -c`) instead of the default shell
// and, if Transient is true, closes the pane when the command exits.
type PaneSplitMsg struct {
	Direction SplitDir
	Command   string // optional: run this command instead of the default shell
	Transient bool   // when true, close the pane when the command exits (no shell restart)
}

// PaneCloseMsg requests closing the active pane.
type PaneCloseMsg struct{}

// PaneAutoCloseMsg requests closing a specific pane whose PTY process exited.
// Emitted by transient terminal panes when the command completes.
type PaneAutoCloseMsg struct {
	PaneID int
}

// PaneFocusMoveMsg requests moving focus to a neighbouring pane.
type PaneFocusMoveMsg struct {
	Direction FocusDir
}

// PaneResizeMsg requests an incremental resize of the active pane.
// When Boundary is false (default), the focused pane grows or shrinks (focus-relative).
// When Boundary is true, the split boundary moves in the direction indicated by Direction
// and Axis regardless of which pane is focused (boundary-absolute, for arrow keys).
type PaneResizeMsg struct {
	Direction ResizeDir
	Axis      ResizeAxis
	Boundary  bool
}

// LayoutResizeMsg propagates new content-area dimensions to the layout manager.
type LayoutResizeMsg struct {
	Width  int
	Height int
}

// TerminalResizeMsg propagates computed terminal pane dimensions.
type TerminalResizeMsg struct {
	Width  int
	Height int
}

// TerminalScrollMsg adjusts the focused terminal's scrollback view.
// Positive Delta scrolls up (into history); negative scrolls down toward live.
// Zero snaps back to the live view (exit scrollback mode).
type TerminalScrollMsg struct {
	Delta int
	Reset bool
}

// SidebarResizeMsg propagates computed sidebar pane dimensions.
type SidebarResizeMsg struct {
	Width  int
	Height int
}

// StatusBarResizeMsg propagates the full window width to the status bar.
type StatusBarResizeMsg struct {
	Width int
}

// MetricsTickMsg carries a snapshot of system metrics collected in the background.
// CWD and git branch are now carried per-pane via PaneCWDChangeMsg/PaneGitStateMsg
// and consolidated into FocusedPaneContextMsg.
type MetricsTickMsg struct {
	CPU      float64
	MemUsed  uint64
	MemTotal uint64
	Tick     time.Time
}

// StatusBarNotifyMsg requests a temporary notification in the status bar.
type StatusBarNotifyMsg struct {
	Text     string
	Level    NotifyLevel
	Duration time.Duration
}

// MouseSelectMsg routes a mouse event to a terminal pane for Lumina-side text
// selection, bypassing PTY passthrough. X/Y in Mouse are pane-local coordinates
// (0,0 = top-left content cell, border already subtracted).
type MouseSelectMsg struct {
	PaneID int
	Mouse  tea.MouseMsg
}

// MouseSelectConfirmMsg confirms a pending mouse selection, copying the selected
// text to the clipboard. Emitted by app.handleKey when the user presses 'y' and
// the focused terminal has a pending selection (mouse_auto_copy=false).
type MouseSelectConfirmMsg struct {
	PaneID int
}

// MouseSelectCancelMsg discards a pending mouse selection without altering the
// clipboard. Emitted by app.handleKey when the user presses 'esc' and the focused
// terminal has a pending selection (mouse_auto_copy=false).
type MouseSelectCancelMsg struct {
	PaneID int
}

// ── Feature 006 (UX Polish Pack) ─────────────────────────────────────────────

// ClickFocusMsg is emitted by the app layer after hit-testing a mouse press
// event against the current layout.Tree. It carries the identity of the pane
// the click landed in plus the (x, y) coordinates translated into that pane's
// local content-cell space (0,0 = first cell inside the pane's border).
type ClickFocusMsg struct {
	PaneID int
	Target FocusTarget
	LocalX int
	LocalY int
}

// SidebarCreateMsg confirms the user has typed a name in the sidebar create
// prompt and pressed Enter. The sidebar component handles the actual filesystem
// mutation in its own Update; this message exists primarily for test feed-in
// and for cross-component observability.
type SidebarCreateMsg struct {
	Kind      string // "dir" or "file"
	Name      string // raw user input, already trimmed
	ParentDir string // absolute path where the creation will happen
}

// SidebarCreatedMsg announces a successful filesystem creation via the sidebar.
// If Kind == "file", the app layer reacts by opening the file in the external
// editor (emitting OpenInExternalEditorMsg). If Kind == "dir", the sidebar
// itself enters the new directory.
type SidebarCreatedMsg struct {
	Kind string // "dir" or "file"
	Path string // absolute path of the newly created entry
}

// OpenInExternalEditorMsg requests that the app spawn a terminal pane running
// the configured external editor (cfg.Editor) with the given file path as its
// single argument.
type OpenInExternalEditorMsg struct {
	Path string
}

// ClockTickMsg is emitted by a 30s tea.Tick scheduled in statusbar.Init.
// The statusbar updates its displayed HH:MM and re-arms the tick.
type ClockTickMsg struct {
	Now time.Time
}

// PaneCWDChangeMsg is emitted by a terminal pane when its OSC 7 callback
// detects a current-working-directory announcement from the child shell.
type PaneCWDChangeMsg struct {
	PaneID int
	CWD    string // absolute path, percent-decoded
}

// PaneGitStateMsg carries the result of a background git query for a pane
// (triggered by PaneCWDChangeMsg). Branch is empty when the CWD is not a git
// repository; Dirty is false in that case as well.
type PaneGitStateMsg struct {
	PaneID int
	Branch string
	Dirty  bool
}

// FocusedPaneContextMsg reaches the statusbar whenever the focused pane's
// identity or context changes. It is the ONLY channel through which the
// statusbar learns about branch/CWD.
type FocusedPaneContextMsg struct {
	PaneID    int
	CWD       string // empty if unknown
	GitBranch string // empty if not a git repo
	GitDirty  bool
}
