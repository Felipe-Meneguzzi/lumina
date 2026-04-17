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
	FocusEditor
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

// PaneSplitMsg requests splitting the active pane.
type PaneSplitMsg struct {
	Direction SplitDir
}

// PaneCloseMsg requests closing the active pane.
type PaneCloseMsg struct{}

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

// EditorResizeMsg propagates computed editor pane dimensions.
type EditorResizeMsg struct {
	Width  int
	Height int
}

// StatusBarResizeMsg propagates the full window width to the status bar.
type StatusBarResizeMsg struct {
	Width int
}

// MetricsTickMsg carries a snapshot of system metrics collected in the background.
type MetricsTickMsg struct {
	CPU       float64
	MemUsed   uint64
	MemTotal  uint64
	CWD       string
	GitBranch string
	Tick      time.Time
}

// OpenFileMsg requests that the editor open the given file path.
type OpenFileMsg struct {
	Path string
}

// ConfirmCloseMsg is emitted by the editor when it has unsaved changes and close is requested.
type ConfirmCloseMsg struct{}

// CloseConfirmedMsg is emitted by app after the user confirms discarding unsaved changes.
type CloseConfirmedMsg struct{}

// CloseAbortedMsg is emitted by app when the user cancels the close confirmation.
type CloseAbortedMsg struct{}

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
