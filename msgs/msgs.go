package msgs

import "time"

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
	SplitVertical                    // stacked
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
	ResizeGrow   ResizeDir = iota
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
type PaneResizeMsg struct {
	Direction ResizeDir
	Axis      ResizeAxis
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
