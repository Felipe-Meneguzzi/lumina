package app

import (
	"fmt"
	"os"
	"strings"

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
)

var confirmStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("214")).
	Padding(0, 2)

// focusOwner identifies which top-level region has keyboard focus.
type focusOwner int

const (
	focusContent focusOwner = iota // layout manager (one or more panes)
	focusSidebar
)

// Model is the root Bubble Tea model that composes all panes.
type Model struct {
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
	confirmClose bool // waiting for user to confirm discarding unsaved changes
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
}

// New initialises the application.
func New(cfg config.Config) (Model, error) {
	cwd, _ := os.Getwd()

	lay, err := layout.New(cfg)
	if err != nil {
		return Model{}, fmt.Errorf("creating layout: %w", err)
	}
	// Update reserved keys so the layout's terminal won't forward Alt+* shortcuts.
	// (layout.New creates one terminal; its SetReservedKeys will be set via the
	// terminal package's exported method once we expose it — for now the terminal
	// reads reservedKeys from its own cfg.)

	m := Model{
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
		paneShowSidebar: make(map[layout.PaneID]bool),
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
				Duration: 6 * 1000000000, // 6 seconds
			}
		}
		return msgs.StatusBarNotifyMsg{
			Text:     "shell: " + shell,
			Level:    msgs.NotifyInfo,
			Duration: 3 * 1000000000, // 3 seconds
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

	// ── Confirm-close dialog (emitted by layout when a dirty editor is closing) ─
	case msgs.ConfirmCloseMsg:
		m.confirmClose = true
		return m, nil

	// ── Layout-level messages forwarded to layout.Model ───────────────────────
	case msgs.PaneFocusMoveMsg:
		// Route to layout and then restore sidebar state for the new focused pane.
		return m.focusPaneThen(msg.Direction)

	case msgs.PaneSplitMsg, msgs.PaneCloseMsg,
		msgs.PaneResizeMsg, msgs.LayoutResizeMsg,
		msgs.PtyOutputMsg, msgs.PtyInputMsg,
		msgs.OpenFileMsg:
		return m.updateLayout(msg)

	// ── Status bar ────────────────────────────────────────────────────────────
	case msgs.MetricsTickMsg, msgs.StatusBarNotifyMsg, msgs.StatusBarResizeMsg:
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
		// Wide enough and sidebar should be visible — restore to previous or default.
		w := m.sidebarPrevWidth
		if w < sidebarMinSize {
			w = 30
		}
		m.sidebarWidth = w
	}
	contentWidth := msg.Width - m.sidebarWidth
	contentHeight := msg.Height - m.statusBarHeight()

	var cmds []tea.Cmd

	// Propagate to layout.
	next, cmd := m.layout.Update(msgs.LayoutResizeMsg{Width: contentWidth, Height: contentHeight})
	m.layout = next.(layout.Model)
	cmds = append(cmds, cmd)

	// Propagate to sidebar.
	next2, cmd2 := m.side.Update(msgs.SidebarResizeMsg{Width: m.sidebarWidth, Height: contentHeight})
	m.side = next2.(sidebar.Model)
	cmds = append(cmds, cmd2)

	// Propagate to status bar.
	next3, cmd3 := m.sbar.Update(msgs.StatusBarResizeMsg{Width: msg.Width})
	m.sbar = next3.(statusbar.Model)
	cmds = append(cmds, cmd3)

	return m, tea.Batch(cmds...)
}

// handleMouse handles mouse events: click-to-focus and sidebar drag-to-resize.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	const borderTolerance = 1

	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		// Check if click is on the sidebar border for drag initiation.
		if m.sidebarWidth > 0 && abs(msg.X-m.sidebarWidth) <= borderTolerance {
			m.sidebarDragging = true
			m.sidebarDragStartX = msg.X
			return m, nil
		}
		// Click outside status bar area — change focus.
		if msg.Y >= m.height-m.statusBarHeight() {
			return m, nil
		}
		if m.sidebarWidth > 0 && msg.X < m.sidebarWidth {
			return m.applyFocusOwner(focusSidebar), nil
		}
		return m.applyFocusOwner(focusContent), nil

	case tea.MouseActionMotion:
		if m.sidebarDragging {
			return m.resizeSidebarTo(msg.X)
		}

	case tea.MouseActionRelease:
		m.sidebarDragging = false
	}

	return m, nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// ── Confirm-close dialog intercepts everything ────────────────────────────
	if m.confirmClose {
		switch strings.ToLower(msg.String()) {
		case "s", "y", "enter":
			m.confirmClose = false
			return m.updateLayout(msgs.CloseConfirmedMsg{})
		default:
			m.confirmClose = false
			return m.updateLayout(msgs.CloseAbortedMsg{})
		}
	}

	// ── Global shortcuts via config.Keybindings ───────────────────────────────
	switch m.keys.Action(msg.String()) {
	case "focus_sidebar":
		return m.applyFocusOwner(focusSidebar), nil
	case "focus_terminal", "focus_editor":
		// Legacy bindings — just give focus back to the content area.
		return m.applyFocusOwner(focusContent), nil

	case "split_horizontal":
		return m.updateLayout(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	case "split_vertical":
		return m.updateLayout(msgs.PaneSplitMsg{Direction: msgs.SplitVertical})
	case "close_pane":
		closedID := m.layout.FocusedID()
		next, cmd := m.updateLayout(msgs.PaneCloseMsg{})
		nm := next.(Model)
		delete(nm.paneShowSidebar, closedID)
		return nm, cmd

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

	// ── Editor and other content input — delegate to layout ──────────────────
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
	return nm, tea.Batch(cmds...)
}

// reapplyResize recomputes layout dimensions using current width/height — used after
// toggles that change the effective content area (statusbar toggle).
func (m Model) reapplyResize() (tea.Model, tea.Cmd) {
	return m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
}

// statusBarHeight returns the height consumed by the status bar (0 when hidden).
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
		visible = true // default: sidebar visible
	}
	return m.applySidebarState(visible)
}

// toggleSidebar inverts sidebar visibility for the focused pane.
func (m Model) toggleSidebar() Model {
	id := m.layout.FocusedID()
	// Determine current state; absent entry = visible.
	wasVisible, exists := m.paneShowSidebar[id]
	if !exists {
		wasVisible = true
	}
	nowVisible := !wasVisible
	m.paneShowSidebar[id] = nowVisible
	return m.applySidebarState(nowVisible)
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

	if m.confirmClose {
		dialog := confirmStyle.Render("Descartar alterações? (s=sim / qualquer tecla=cancelar)")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			dialog, lipgloss.WithWhitespaceChars(" "),
		) + "\n" + screen
	}

	if m.showHelp {
		helpView := m.help.View(m.keymap)
		return screen + "\n" + helpView
	}

	return screen
}
