package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/menegas/lumina/components/layout"
	"github.com/menegas/lumina/components/sidebar"
	"github.com/menegas/lumina/components/statusbar"
	"github.com/menegas/lumina/components/terminal"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

// shellEscape wraps a path in single quotes and escapes any embedded single quotes,
// making it safe to pass as an argument to any POSIX shell.
func shellEscape(path string) string {
	return "'" + strings.ReplaceAll(path, "'", `'\''`) + "'"
}

const (
	statusBarHeight = 1
	sidebarMinWidth = 80 // hide sidebar below this total width
	sidebarMinSize  = 16 // minimum sidebar width in columns
	sidebarMaxRatio = 3  // sidebar max = totalWidth / sidebarMaxRatio
	sidebarStep     = 2  // columns per resize keypress
	scrollPageStep  = 10 // rows per PgUp/PgDown in terminal scrollback
	scrollWheelStep = 3  // rows per mouse wheel tick in terminal scrollback

	// gitQueryTimeout caps the background `git` invocation triggered by OSC 7
	// CWD announcements. Beyond this, the pane reports an empty branch so the
	// status bar never stalls waiting for a slow filesystem.
	gitQueryTimeout = 200 * time.Millisecond
)

// focusOwner identifies which top-level region has keyboard focus.
type focusOwner int

const (
	focusContent focusOwner = iota // layout manager (one or more panes)
	focusSidebar
)

// Option configures the app Model at construction time.
type Option func(*Model)

// WithNoSidebar hides the sidebar for all panes on startup.
func WithNoSidebar() Option {
	return func(m *Model) {
		m.sidebarVisible = false
		m.sidebarWidth = 0
	}
}

// paneContext caches the per-pane CWD and git state observed via OSC 7 +
// background `git` queries. FocusedPaneContextMsg consolidates the entry for
// the currently-focused pane and forwards it to the status bar.
type paneContext struct {
	cwd       string
	gitBranch string
	gitDirty  bool
}

// Model is the root Bubble Tea model that composes all panes.
type Model struct {
	cfg          config.Config
	keymap       KeyMap
	keys         config.Keybindings
	layout       layout.Model
	sbar         statusbar.Model
	side         sidebar.Model
	help         help.Model
	focus        focusOwner
	width        int
	height       int
	sidebarWidth int
	showHelp     bool
	shell        string // active shell path, shown on startup notification
	shellWarning string // non-empty if configured shell was rejected (e.g. .exe on WSL)

	// Sidebar per-pane visibility state
	sidebarVisible   bool                   // current sidebar visible state
	sidebarPrevWidth int                    // width before hiding (to restore)
	paneShowSidebar  map[layout.PaneID]bool // per-pane visibility; absent = visible

	// Statusbar visibility
	sbarVisible bool

	// Mouse drag for sidebar resize
	sidebarDragging   bool
	sidebarDragStartX int

	// Per-pane CWD + git state (feature 006 / US2).
	paneCtx map[layout.PaneID]paneContext
}

// New initialises the application.
func New(cfg config.Config, layoutOpts []layout.Option, appOpts ...Option) (Model, error) {
	cwd, _ := os.Getwd()

	lay, err := layout.New(cfg, layoutOpts...)
	if err != nil {
		return Model{}, fmt.Errorf("creating layout: %w", err)
	}

	m := Model{
		cfg:             cfg,
		keymap:          NewKeyMap(cfg.Keys),
		keys:            cfg.Keys,
		layout:          lay,
		sbar:            statusbar.New(cfg),
		side:            sidebar.New(cwd, cfg),
		help:            help.New(),
		focus:           focusContent,
		width:           80,
		height:          24,
		shell:           cfg.Shell,
		shellWarning:    cfg.ShellWarning,
		sidebarVisible:  true,
		sbarVisible:     true,
		paneShowSidebar: map[layout.PaneID]bool{},
		paneCtx:         map[layout.PaneID]paneContext{},
	}

	for _, opt := range appOpts {
		opt(&m)
	}

	// Mark every pane produced by layout.New with the current sidebar visibility.
	for _, id := range lay.AllPaneIDs() {
		m.paneShowSidebar[id] = m.sidebarVisible
	}
	m.layout = m.layout.SetContentFocused(true)
	return m, nil
}

// Init starts all component Init Cmds.
func (m Model) Init() tea.Cmd {
	shell := m.shell
	shellWarning := m.shellWarning
	startupNotify := func() tea.Msg {
		if shellWarning != "" {
			return msgs.StatusBarNotifyMsg{
				Text:     shellWarning,
				Level:    msgs.NotifyWarning,
				Duration: 6 * time.Second,
			}
		}
		return msgs.StatusBarNotifyMsg{
			Text:     "shell: " + shell,
			Level:    msgs.NotifyInfo,
			Duration: 3 * time.Second,
		}
	}
	return tea.Batch(
		m.layout.Init(),
		m.sbar.Init(),
		m.side.Init(),
		startupNotify,
	)
}

// Update is the central message router.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		return m.handleResize(msg)

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	// ── Layout-level messages forwarded to layout.Model ───────────────────────
	case msgs.PaneFocusMoveMsg:
		return m.focusPaneThen(msg.Direction)

	case msgs.PaneSplitMsg:
		return m.handleSplit(msg)

	case msgs.PaneCloseMsg,
		msgs.PaneResizeMsg, msgs.LayoutResizeMsg,
		msgs.PtyOutputMsg, msgs.PtyInputMsg:
		return m.updateLayout(msg)

	case msgs.PaneAutoCloseMsg:
		closedID := layout.PaneID(msg.PaneID)
		next, cmd := m.layout.Update(msg)
		m.layout = next.(layout.Model)
		delete(m.paneShowSidebar, closedID)
		delete(m.paneCtx, closedID)
		// After an auto-close the layout re-assigns focus internally; refresh
		// the status bar for the new focused pane.
		return m, tea.Batch(cmd, m.focusedContextCmd())

	case msgs.PaneCWDChangeMsg:
		return m.handlePaneCWDChange(msg)

	case msgs.PaneGitStateMsg:
		return m.handlePaneGitState(msg)

	case msgs.OpenInExternalEditorMsg:
		return m.openInExternalEditor(msg.Path)

	case msgs.SidebarCreatedMsg:
		return m.handleSidebarCreated(msg)

	// ── Status bar ────────────────────────────────────────────────────────────
	case msgs.MetricsTickMsg, msgs.ClockTickMsg,
		msgs.StatusBarNotifyMsg, msgs.StatusBarResizeMsg,
		msgs.FocusedPaneContextMsg:
		next, cmd := m.sbar.Update(msg)
		m.sbar = next.(statusbar.Model)
		return m, cmd

	// ── Sidebar ───────────────────────────────────────────────────────────────
	case msgs.SidebarResizeMsg:
		next, cmd := m.side.Update(msg)
		m.side = next.(sidebar.Model)
		return m, cmd
	}

	return m, nil
}

func (m Model) updateLayout(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.layout.Update(msg)
	m.layout = next.(layout.Model)
	return m, cmd
}

func (m Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	// Auto-hide sidebar on narrow terminals; restore when wide enough.
	if msg.Width < sidebarMinWidth {
		if m.sidebarWidth > 0 {
			m.sidebarPrevWidth = m.sidebarWidth
		}
		m.sidebarWidth = 0
	} else if m.sidebarVisible && m.sidebarWidth == 0 {
		w := m.sidebarPrevWidth
		if w < sidebarMinSize {
			w = 30
		}
		m.sidebarWidth = w
	}
	contentWidth := msg.Width - m.sidebarWidth
	contentHeight := msg.Height - m.statusBarHeight()

	var cmds []tea.Cmd

	next, cmd := m.layout.Update(msgs.LayoutResizeMsg{Width: contentWidth, Height: contentHeight})
	m.layout = next.(layout.Model)
	cmds = append(cmds, cmd)

	next2, cmd2 := m.side.Update(msgs.SidebarResizeMsg{Width: m.sidebarWidth, Height: contentHeight})
	m.side = next2.(sidebar.Model)
	cmds = append(cmds, cmd2)

	next3, cmd3 := m.sbar.Update(msgs.StatusBarResizeMsg{Width: msg.Width})
	m.sbar = next3.(statusbar.Model)
	cmds = append(cmds, cmd3)

	return m, tea.Batch(cmds...)
}

// handleMouse handles mouse events: click-to-focus with pass-through to the
// clicked pane (feature 006 / FR-004a/b/c), sidebar drag-to-resize, and
// Alt+wheel scrollback for terminal panes.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	const borderTolerance = 1

	// Alt+wheel scrolls the focused terminal's scrollback history, regardless
	// of pointer position — Alt gates it so plain wheel events stay reserved
	// for shell apps that opt into mouse reporting.
	if msg.Alt && (msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown) {
		if m.focus == focusContent && m.layout.FocusedKind() == layout.KindTerminal {
			delta := scrollWheelStep
			if msg.Button == tea.MouseButtonWheelDown {
				delta = -scrollWheelStep
			}
			return m.updateLayout(msgs.TerminalScrollMsg{Delta: delta})
		}
	}

	// For motion/release events on a focused terminal pane, preserve the
	// existing selection/PTY passthrough path.
	if msg.Action != tea.MouseActionPress && !msg.Alt &&
		m.focus == focusContent && m.layout.FocusedKind() == layout.KindTerminal {
		if routed, m2, cmd, ok := m.routeMouseToFocusedPane(msg); ok {
			return m2, cmd
		} else {
			_ = routed // silence linter; routed unused when not handled
		}
	}

	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		// Click on the sidebar border starts a drag to resize.
		if m.sidebarWidth > 0 && abs(msg.X-m.sidebarWidth) <= borderTolerance {
			m.sidebarDragging = true
			m.sidebarDragStartX = msg.X
			return m, nil
		}
		// Click on the status bar area is ignored.
		if msg.Y >= m.height-m.statusBarHeight() {
			return m, nil
		}
		// Click on sidebar area focuses the sidebar.
		if m.sidebarWidth > 0 && msg.X < m.sidebarWidth {
			prev := m
			_ = prev
			mm := m.applyFocusOwner(focusSidebar)
			return mm, mm.focusedContextCmd()
		}
		// Click on layout content: hit-test, transfer focus, deliver click.
		gx := msg.X - m.sidebarWidth
		gy := msg.Y
		paneID, target, localX, localY, ok := m.layout.HitTest(gx, gy)
		if !ok {
			mm := m.applyFocusOwner(focusContent)
			return mm, mm.focusedContextCmd()
		}
		focusChanged := m.focus != focusContent || m.layout.FocusedID() != paneID
		m = m.applyFocusOwner(focusContent)
		if m.layout.FocusedID() != paneID {
			m.layout = m.layout.SetFocusedID(paneID)
		}
		inner := msg
		inner.X = localX
		inner.Y = localY
		// Decide delivery path: PTY passthrough when the inner application has
		// opted into mouse reporting and Shift is not held; Lumina-side
		// selection otherwise.
		var deliver tea.Msg
		if m.layout.FocusedMouseEnabled() && !msg.Shift {
			deliver = msgs.PtyMouseMsg{PaneID: int(paneID), Mouse: inner}
		} else {
			deliver = msgs.MouseSelectMsg{PaneID: int(paneID), Mouse: inner}
		}
		_ = target
		var cmds []tea.Cmd
		if focusChanged {
			cmds = append(cmds, m.focusedContextCmd())
		}
		next, cmd := m.layout.Update(deliver)
		m.layout = next.(layout.Model)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	case tea.MouseActionMotion:
		if m.sidebarDragging {
			return m.resizeSidebarTo(msg.X)
		}

	case tea.MouseActionRelease:
		m.sidebarDragging = false
	}

	return m, nil
}

// routeMouseToFocusedPane delivers motion/release events inside the focused
// terminal's bounds. Returns handled=true when the event was consumed.
func (m Model) routeMouseToFocusedPane(msg tea.MouseMsg) (tea.Model, tea.Model, tea.Cmd, bool) {
	px, py, pw, ph, ok := m.layout.FocusedBounds()
	if !ok {
		return m, m, nil, false
	}
	gx := msg.X - m.sidebarWidth
	gy := msg.Y
	localX := gx - px - 1
	localY := gy - py - 1
	if localX < 0 || localY < 0 || localX >= pw-2 || localY >= ph-2 {
		return m, m, nil, false
	}
	inner := msg
	inner.X = localX
	inner.Y = localY
	paneID := int(m.layout.FocusedID())
	if !m.layout.FocusedMouseEnabled() || msg.Shift {
		next, cmd := m.updateLayout(msgs.MouseSelectMsg{PaneID: paneID, Mouse: inner})
		return next, next, cmd, true
	}
	next, cmd := m.updateLayout(msgs.PtyMouseMsg{PaneID: paneID, Mouse: inner})
	return next, next, cmd, true
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// ── Global shortcuts via config.Keybindings ───────────────────────────────
	switch m.keys.Action(msg.String()) {
	case "focus_sidebar":
		mm := m.applyFocusOwner(focusSidebar)
		return mm, mm.focusedContextCmd()
	case "focus_terminal":
		mm := m.applyFocusOwner(focusContent)
		return mm, mm.focusedContextCmd()

	case "split_horizontal":
		return m.handleSplit(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	case "split_vertical":
		return m.handleSplit(msgs.PaneSplitMsg{Direction: msgs.SplitVertical})
	case "close_pane":
		closedID := m.layout.FocusedID()
		next, cmd := m.updateLayout(msgs.PaneCloseMsg{})
		nm := next.(Model)
		delete(nm.paneShowSidebar, closedID)
		delete(nm.paneCtx, closedID)
		return nm, tea.Batch(cmd, nm.focusedContextCmd())

	case "focus_pane_left":
		return m.focusPaneThen(msgs.FocusDirLeft)
	case "focus_pane_right":
		return m.focusPaneThen(msgs.FocusDirRight)
	case "focus_pane_up":
		return m.focusPaneThen(msgs.FocusDirUp)
	case "focus_pane_down":
		return m.focusPaneThen(msgs.FocusDirDown)

	case "grow_pane_h":
		return m.updateLayout(msgs.PaneResizeMsg{Direction: msgs.ResizeGrow, Axis: msgs.ResizeAxisH})
	case "shrink_pane_h":
		return m.updateLayout(msgs.PaneResizeMsg{Direction: msgs.ResizeShrink, Axis: msgs.ResizeAxisH})
	case "grow_pane_v":
		return m.updateLayout(msgs.PaneResizeMsg{Direction: msgs.ResizeGrow, Axis: msgs.ResizeAxisV})
	case "shrink_pane_v":
		return m.updateLayout(msgs.PaneResizeMsg{Direction: msgs.ResizeShrink, Axis: msgs.ResizeAxisV})

	case "boundary_right":
		return m.updateLayout(msgs.PaneResizeMsg{Direction: msgs.ResizeGrow, Axis: msgs.ResizeAxisH, Boundary: true})
	case "boundary_left":
		return m.updateLayout(msgs.PaneResizeMsg{Direction: msgs.ResizeShrink, Axis: msgs.ResizeAxisH, Boundary: true})
	case "boundary_down":
		return m.updateLayout(msgs.PaneResizeMsg{Direction: msgs.ResizeGrow, Axis: msgs.ResizeAxisV, Boundary: true})
	case "boundary_up":
		return m.updateLayout(msgs.PaneResizeMsg{Direction: msgs.ResizeShrink, Axis: msgs.ResizeAxisV, Boundary: true})

	case "grow_sidebar":
		return m.resizeSidebar(sidebarStep)
	case "shrink_sidebar":
		return m.resizeSidebar(-sidebarStep)
	case "toggle_sidebar":
		m = m.toggleSidebar()
		return m.reapplyResize()
	case "toggle_statusbar":
		m.sbarVisible = !m.sbarVisible
		return m.reapplyResize()

	case "open_terminal_here":
		if m.focus == focusSidebar {
			return m.openTerminalInSidebarDir()
		}

	case "enter_copy_mode":
		if m.focus == focusContent && m.layout.FocusedKind() == layout.KindTerminal {
			return m.updateLayout(msgs.EnterCopyModeMsg{})
		}

	case "quit":
		if m.focus != focusContent || m.layout.FocusedKind() != layout.KindTerminal {
			return m, tea.Quit
		}
	case "help":
		if m.focus != focusContent || m.layout.FocusedKind() != layout.KindTerminal {
			m.showHelp = !m.showHelp
			return m, nil
		}
	}

	// ── Terminal raw mode — forward input to PTY when content is focused ──────
	if m.focus == focusContent && m.layout.FocusedKind() == layout.KindTerminal {
		if m.layout.FocusedInCopyMode() {
			return m.updateLayout(msg)
		}
		if m.layout.FocusedHasPendingSelection() {
			paneID := int(m.layout.FocusedID())
			switch msg.String() {
			case "y":
				return m.updateLayout(msgs.MouseSelectConfirmMsg{PaneID: paneID})
			case "esc":
				return m.updateLayout(msgs.MouseSelectCancelMsg{PaneID: paneID})
			}
		}
		switch msg.Type {
		case tea.KeyPgUp:
			return m.updateLayout(msgs.TerminalScrollMsg{Delta: scrollPageStep})
		case tea.KeyPgDown:
			return m.updateLayout(msgs.TerminalScrollMsg{Delta: -scrollPageStep})
		}
		data := terminal.KeyToBytes(msg, m.keys.GlobalKeys())
		if len(data) > 0 {
			return m.updateLayout(msgs.PtyInputMsg{Data: data})
		}
		return m, nil
	}

	// ── Sidebar input ─────────────────────────────────────────────────────────
	if m.focus == focusSidebar {
		next, cmd := m.side.Update(msg)
		m.side = next.(sidebar.Model)
		return m, cmd
	}

	// ── Layout catch-all ──────────────────────────────────────────────────────
	if m.focus == focusContent {
		return m.updateLayout(msg)
	}

	return m, nil
}

// openTerminalInSidebarDir sends "cd <dir>\r" to the focused terminal PTY.
func (m Model) openTerminalInSidebarDir() (tea.Model, tea.Cmd) {
	dir := m.side.CWD()
	if dir == "" {
		return m, nil
	}
	cdCmd := fmt.Sprintf("cd %s\r", shellEscape(dir))
	m = m.applyFocusOwner(focusContent)
	return m.updateLayout(msgs.PtyInputMsg{Data: []byte(cdCmd)})
}

// applyFocusOwner updates the focus state between sidebar and content area.
func (m Model) applyFocusOwner(owner focusOwner) Model {
	m.focus = owner
	m.side.SetFocused(owner == focusSidebar)
	m.layout = m.layout.SetContentFocused(owner == focusContent)
	return m
}

// resizeSidebar adjusts the sidebar width by delta columns.
func (m Model) resizeSidebar(delta int) (tea.Model, tea.Cmd) {
	return m.resizeSidebarTo(m.sidebarWidth + delta)
}

// resizeSidebarTo sets the sidebar to an absolute width and propagates sizes.
func (m Model) resizeSidebarTo(newW int) (tea.Model, tea.Cmd) {
	maxW := m.width / sidebarMaxRatio
	if newW < sidebarMinSize {
		newW = sidebarMinSize
	}
	if newW > maxW {
		newW = maxW
	}
	if newW == m.sidebarWidth {
		return m, nil
	}
	m.sidebarWidth = newW
	contentWidth := m.width - m.sidebarWidth
	contentHeight := m.height - m.statusBarHeight()

	var cmds []tea.Cmd
	next, cmd := m.layout.Update(msgs.LayoutResizeMsg{Width: contentWidth, Height: contentHeight})
	m.layout = next.(layout.Model)
	cmds = append(cmds, cmd)

	next2, cmd2 := m.side.Update(msgs.SidebarResizeMsg{Width: m.sidebarWidth, Height: contentHeight})
	m.side = next2.(sidebar.Model)
	cmds = append(cmds, cmd2)

	return m, tea.Batch(cmds...)
}

// handleSplit creates a new pane in the given direction, optionally running a
// specific command (used by openInExternalEditor to spawn editor panes).
func (m Model) handleSplit(msg msgs.PaneSplitMsg) (tea.Model, tea.Cmd) {
	next, cmd := m.updateLayout(msg)
	nm := next.(Model)
	newID := nm.layout.FocusedID()
	nm.paneShowSidebar[newID] = nm.sidebarVisible
	return nm, tea.Batch(cmd, nm.focusedContextCmd())
}

// focusPaneThen moves pane focus then restores the focused pane's sidebar state.
func (m Model) focusPaneThen(dir msgs.FocusDir) (tea.Model, tea.Cmd) {
	next, cmd := m.updateLayout(msgs.PaneFocusMoveMsg{Direction: dir})
	nm := next.(Model)
	nm = nm.applySidebarForFocusedPane()
	contentW := nm.width - nm.sidebarWidth
	contentH := nm.height - nm.statusBarHeight()
	var cmds []tea.Cmd
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	ln, lc := nm.layout.Update(msgs.LayoutResizeMsg{Width: contentW, Height: contentH})
	nm.layout = ln.(layout.Model)
	cmds = append(cmds, lc)
	sn, sc := nm.side.Update(msgs.SidebarResizeMsg{Width: nm.sidebarWidth, Height: contentH})
	nm.side = sn.(sidebar.Model)
	cmds = append(cmds, sc)
	cmds = append(cmds, nm.focusedContextCmd())
	return nm, tea.Batch(cmds...)
}

// reapplyResize recomputes layout dimensions using current width/height.
func (m Model) reapplyResize() (tea.Model, tea.Cmd) {
	return m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
}

// statusBarHeight returns the height consumed by the status bar.
func (m Model) statusBarHeight() int {
	if m.sbarVisible {
		return statusBarHeight
	}
	return 0
}

// SidebarWidth returns the current sidebar width (exported for tests).
func (m Model) SidebarWidth() int { return m.sidebarWidth }

// SbarVisible returns whether the status bar is currently visible (exported for tests).
func (m Model) SbarVisible() bool { return m.sbarVisible }

// applySidebarState hides or shows the sidebar, preserving width for restoration.
func (m Model) applySidebarState(visible bool) Model {
	if visible {
		w := m.sidebarPrevWidth
		if w < sidebarMinSize {
			w = 30
		}
		m.sidebarWidth = w
		m.sidebarVisible = true
	} else {
		if m.sidebarWidth > 0 {
			m.sidebarPrevWidth = m.sidebarWidth
		}
		m.sidebarWidth = 0
		m.sidebarVisible = false
	}
	return m
}

// applySidebarForFocusedPane restores the sidebar state for the currently focused pane.
func (m Model) applySidebarForFocusedPane() Model {
	id := m.layout.FocusedID()
	visible, exists := m.paneShowSidebar[id]
	if !exists {
		visible = false
	}
	return m.applySidebarState(visible)
}

// toggleSidebar inverts sidebar visibility for the focused pane.
func (m Model) toggleSidebar() Model {
	id := m.layout.FocusedID()
	wasVisible, exists := m.paneShowSidebar[id]
	if !exists {
		wasVisible = false
	}
	nowVisible := !wasVisible
	m.paneShowSidebar[id] = nowVisible
	return m.applySidebarState(nowVisible)
}

// ── External editor + pane context (feature 006) ─────────────────────────────

// openInExternalEditor spawns a new terminal pane running the configured
// editor against the given file. If the configured binary is not in PATH, no
// pane is created and an error notification lands on the status bar
// (FR-018 / remediation I1 — no silent fallback to nano).
func (m Model) openInExternalEditor(path string) (tea.Model, tea.Cmd) {
	editor := strings.TrimSpace(m.cfg.Editor)
	if editor == "" {
		editor = "nano"
	}
	if _, err := exec.LookPath(editor); err != nil {
		return m, func() tea.Msg {
			return msgs.StatusBarNotifyMsg{
				Text:     fmt.Sprintf("editor '%s' não encontrado no PATH", editor),
				Level:    msgs.NotifyError,
				Duration: 5 * time.Second,
			}
		}
	}
	command := fmt.Sprintf("%s %s", shellEscape(editor), shellEscape(path))
	return m.handleSplit(msgs.PaneSplitMsg{
		Direction: msgs.SplitHorizontal,
		Command:   command,
		Transient: true,
	})
}

// handleSidebarCreated reacts to a filesystem creation announced by the
// sidebar. Files trigger the external editor flow; directories are handled
// directly by the sidebar navigation (sidebar.Update consumes the msg too).
func (m Model) handleSidebarCreated(msg msgs.SidebarCreatedMsg) (tea.Model, tea.Cmd) {
	// Forward to sidebar so it refreshes the listing / navigates into the dir.
	next, cmd := m.side.Update(msg)
	m.side = next.(sidebar.Model)
	if msg.Kind != "file" {
		return m, cmd
	}
	// For files, also spawn the editor.
	m2, cmd2 := m.openInExternalEditor(msg.Path)
	return m2, tea.Batch(cmd, cmd2)
}

// handlePaneCWDChange records the new CWD for the originating pane and kicks
// off a background git query. If the updated pane is the focused one, the
// status bar is refreshed in the same update cycle.
func (m Model) handlePaneCWDChange(msg msgs.PaneCWDChangeMsg) (tea.Model, tea.Cmd) {
	id := layout.PaneID(msg.PaneID)
	ctx := m.paneCtx[id]
	ctx.cwd = msg.CWD
	ctx.gitBranch = ""
	ctx.gitDirty = false
	m.paneCtx[id] = ctx

	gitCmd := queryGitState(int(id), msg.CWD)
	if layout.PaneID(msg.PaneID) == m.layout.FocusedID() && m.focus == focusContent {
		return m, tea.Batch(gitCmd, m.focusedContextCmd())
	}
	return m, gitCmd
}

// handlePaneGitState records the git query result; if the pane is currently
// focused, the status bar is refreshed.
func (m Model) handlePaneGitState(msg msgs.PaneGitStateMsg) (tea.Model, tea.Cmd) {
	id := layout.PaneID(msg.PaneID)
	ctx := m.paneCtx[id]
	ctx.gitBranch = msg.Branch
	ctx.gitDirty = msg.Dirty
	m.paneCtx[id] = ctx
	if id == m.layout.FocusedID() && m.focus == focusContent {
		return m, m.focusedContextCmd()
	}
	return m, nil
}

// focusedContextCmd returns a Cmd emitting a FocusedPaneContextMsg describing
// the currently-focused pane. Sidebar focus reports the sidebar CWD (no git).
func (m Model) focusedContextCmd() tea.Cmd {
	if m.focus == focusSidebar {
		cwd := m.side.CWD()
		return func() tea.Msg {
			return msgs.FocusedPaneContextMsg{CWD: cwd}
		}
	}
	id := m.layout.FocusedID()
	ctx := m.paneCtx[id]
	return func() tea.Msg {
		return msgs.FocusedPaneContextMsg{
			PaneID:    int(id),
			CWD:       ctx.cwd,
			GitBranch: ctx.gitBranch,
			GitDirty:  ctx.gitDirty,
		}
	}
}

// queryGitState runs `git -C <cwd>` commands under a timeout and emits a
// PaneGitStateMsg when done. Empty cwd or a non-repo directory yields an
// empty branch; the query never blocks the event loop.
func queryGitState(paneID int, cwd string) tea.Cmd {
	if cwd == "" {
		return func() tea.Msg { return msgs.PaneGitStateMsg{PaneID: paneID} }
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), gitQueryTimeout)
		defer cancel()
		branch, err := exec.CommandContext(ctx, "git", "-C", cwd, "symbolic-ref", "--short", "HEAD").Output()
		if err != nil {
			return msgs.PaneGitStateMsg{PaneID: paneID}
		}
		branchName := strings.TrimSpace(string(branch))
		dirty := false
		statusOut, err := exec.CommandContext(ctx, "git", "-C", cwd, "status", "--porcelain").Output()
		if err == nil && len(strings.TrimSpace(string(statusOut))) > 0 {
			dirty = true
		}
		return msgs.PaneGitStateMsg{PaneID: paneID, Branch: branchName, Dirty: dirty}
	}
}

// View renders the full TUI.
func (m Model) View() string {
	layoutView := m.layout.View()
	sideView := m.side.View()

	var content string
	if sideView != "" {
		content = lipgloss.JoinHorizontal(lipgloss.Top, sideView, layoutView)
	} else {
		content = layoutView
	}

	var screen string
	if m.sbarVisible {
		screen = lipgloss.JoinVertical(lipgloss.Left, content, m.sbar.View())
	} else {
		screen = content
	}

	if m.showHelp {
		helpView := m.help.View(m.keymap)
		return screen + "\n" + helpView
	}

	return screen
}
