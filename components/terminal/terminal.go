package terminal

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

var (
	// ThickBorder (focused) vs RoundedBorder (unfocused) differ even without color support.
	focusedBorder   = lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("62"))
	unfocusedBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))
)

// Model is the Bubble Tea model for the terminal pane.
type Model struct {
	shell          string
	shellOverride  string // when non-empty, runs via `sh -c <override>` instead of the default shell
	forceTheme     bool
	mouseAutoCopy  bool   // copy to clipboard automatically on mouse release (config.mouse_auto_copy)
	mouseSelMode   string // "linear" (notepad-style) or "block" (rectangular); default "linear"
	ptyFile        *os.File
	cmd            *exec.Cmd
	vt             *vt.Emulator    // virtual terminal — handles all escape sequences
	reservedKeys   map[string]bool // keys not to forward to PTY (global shortcuts)
	width          int
	height         int
	focused        bool
	closed         bool            // true when running without a live PTY (used in tests)
	paneID         int             // identifies this terminal in multi-pane output routing
	scrollOffset   int             // how many rows above the live view the user is currently looking at
	state          *sharedState    // mutable state populated by emulator callbacks (mouse modes, title, cwd, bell)
	copy           *copyState      // non-nil when the terminal is in tmux-style copy mode
	mouseSelection *mouseSelection // non-nil when a mouse drag selection is active or pending
}

// New creates a new terminal Model and starts the shell process.
func New(cfg config.Config) (Model, error) {
	return newModel(cfg, "")
}

// NewWithCommand is like New but runs the given command instead of the
// default shell. Empty command falls back to New's behaviour.
func NewWithCommand(cfg config.Config, command string) (Model, error) {
	return newModel(cfg, command)
}

func newModel(cfg config.Config, override string) (Model, error) {
	cols, rows := 78, 22 // default inner dimensions (width-2, height-2 for border)
	m := Model{
		shell:         cfg.Shell,
		shellOverride: override,
		forceTheme:    cfg.ForceShellTheme,
		mouseAutoCopy: cfg.MouseAutoCopy,
		mouseSelMode:  selectionMode(cfg.SelectionMode),
		width:         80,
		height:        24,
		vt:            vt.NewEmulator(cols, rows),
		state:         &sharedState{},
	}
	m.vt.SetScrollbackSize(scrollbackMax)
	installCallbacks(m.vt, m.state)

	if err := m.startShell(); err != nil {
		return m, fmt.Errorf("starting shell: %w", err)
	}
	return m, nil
}

// Close shuts down the PTY without restarting — used in tests to skip PTY creation.
// The emulator is intentionally left open so tests can still feed PtyOutputMsg
// and assert View(); the InputPipe goroutine exits on its own once the PTY closes.
func (m *Model) Close() {
	m.closed = true
	if m.ptyFile != nil {
		_ = m.ptyFile.Close()
		m.ptyFile = nil
	}
	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
	}
}

// CloseResources releases the PTY file and kills the shell process.
// Uses a value receiver so it can be called after a type assertion from tea.Model.
// The underlying OS resources are freed even though the value copy's fields are not nulled.
func (m Model) CloseResources() {
	if m.ptyFile != nil {
		_ = m.ptyFile.Close()
	}
	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
	}
	if m.vt != nil {
		_ = m.vt.Close()
	}
}

// SetFocused sets the focus state.
func (m *Model) SetFocused(f bool) { m.focused = f }

// SetPaneID sets the pane identifier used to tag PtyOutputMsg.
// Must be called before Init() to ensure the read goroutine captures the correct ID.
func (m *Model) SetPaneID(id int) { m.paneID = id }

// PaneID returns the terminal's pane identifier.
func (m Model) PaneID() int { return m.paneID }

// SetReservedKeys sets the keys that must not be forwarded to the PTY.
// Called once at startup with the keys loaded from keybindings.json.
func (m *Model) SetReservedKeys(keys map[string]bool) { m.reservedKeys = keys }

// Dimensions returns the current width and height of the terminal pane.
func (m Model) Dimensions() (int, int) { return m.width, m.height }

func (m *Model) startShell() error {
	env := os.Environ()
	env = setEnv(env, "TERM", "xterm-256color")
	env = setEnv(env, "COLORTERM", "truecolor")

	var cmd *exec.Cmd
	if m.shellOverride != "" {
		cmd = exec.Command("sh", "-c", m.shellOverride)
	} else {
		cmd = buildShellCommand(m.shell, m.forceTheme, &env)
	}
	cmd.Env = env

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	cols, rows := m.vt.Width(), m.vt.Height()
	if err := pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	}); err != nil {
		_ = ptmx.Close()
		return err
	}

	m.ptyFile = ptmx
	m.cmd = cmd

	// Forward emulator-generated input (mouse SGR, paste, query responses)
	// to the PTY. The goroutine exits when the emulator is Closed.
	emu := m.vt
	go func() {
		_, _ = io.Copy(ptmx, emu)
	}()
	return nil
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if e[:min(len(e), len(prefix))] == prefix {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

// waitForOutput returns a Cmd that reads one chunk from the PTY.
// paneID is captured by value so the goroutine always uses the correct identifier.
func waitForOutput(f *os.File, paneID int) tea.Cmd {
	return func() tea.Msg {
		buf := make([]byte, 4096)
		n, err := f.Read(buf)
		if n > 0 {
			return msgs.PtyOutputMsg{PaneID: paneID, Data: buf[:n], Err: err}
		}
		return msgs.PtyOutputMsg{PaneID: paneID, Err: err}
	}
}

// Init starts reading from the PTY.
func (m Model) Init() tea.Cmd {
	if m.closed || m.ptyFile == nil {
		return nil
	}
	return waitForOutput(m.ptyFile, m.paneID)
}

// Update handles messages for the terminal pane.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case msgs.PtyOutputMsg:
		if msg.Err != nil {
			// Shell exited — auto-restart (FR-010).
			// The previous emulator's InputPipe goroutine already exited because
			// its PTY was closed, so we just discard it and start fresh.
			cols, rows := m.vt.Width(), m.vt.Height()
			m.vt = vt.NewEmulator(cols, rows)
			m.vt.SetScrollbackSize(scrollbackMax)
			*m.state = sharedState{}
			installCallbacks(m.vt, m.state)
			m.scrollOffset = 0
			if err := m.startShell(); err != nil {
				return m, nil
			}
			return m, waitForOutput(m.ptyFile, m.paneID)
		}
		// Feed raw bytes to the emulator. If the user is currently scrolled into
		// history — or in copy mode (where the viewport must stay frozen while
		// selecting) — preserve their viewing position by tracking how many lines
		// were pushed into scrollback by this write.
		prevSBLen := m.vt.ScrollbackLen()
		freezeView := m.scrollOffset > 0 || m.copy != nil
		_, _ = m.vt.Write(msg.Data)
		if freezeView {
			pushed := m.vt.ScrollbackLen() - prevSBLen
			if pushed > 0 {
				m.scrollOffset += pushed
			}
		}
		if sbLen := m.vt.ScrollbackLen(); m.scrollOffset > sbLen {
			m.scrollOffset = sbLen
		}
		if m.ptyFile != nil {
			return m, waitForOutput(m.ptyFile, m.paneID)
		}
		return m, nil

	case msgs.PaneFocusMsg:
		m.focused = msg.Focused
		if !msg.Focused {
			m.mouseSelection = nil
		}
		return m, nil

	case msgs.PtyInputMsg:
		if m.ptyFile != nil && len(msg.Data) > 0 {
			// Typing returns the user to the live view, mirroring tmux/screen behaviour.
			m.scrollOffset = 0
			_, _ = m.ptyFile.Write(msg.Data)
		}
		return m, nil

	case msgs.PtyMouseMsg:
		if m.MouseEnabled() && !m.InCopyMode() {
			teaMouseToVT(m.vt, msg.Mouse)
		}
		return m, nil

	case msgs.MouseSelectMsg:
		switch msg.Mouse.Action {
		case tea.MouseActionPress:
			m.startMouseSelection(msg.Mouse.X, msg.Mouse.Y)
		case tea.MouseActionMotion:
			m.updateMouseSelection(msg.Mouse.X, msg.Mouse.Y)
		case tea.MouseActionRelease:
			return m, m.finalizeMouseSelection(msg.Mouse.X, msg.Mouse.Y, m.mouseAutoCopy)
		}
		return m, nil

	case msgs.MouseSelectConfirmMsg:
		return m, m.confirmMouseSelection()

	case msgs.MouseSelectCancelMsg:
		m.cancelMouseSelection()
		return m, nil

	case msgs.EnterCopyModeMsg:
		m.enterCopyMode()
		return m, nil

	case tea.KeyMsg:
		// Copy mode swallows all keys until the user copies or aborts.
		if m.copy != nil {
			cmd := m.handleCopyKey(msg)
			return m, cmd
		}
		return m, nil

	case msgs.TerminalScrollMsg:
		if msg.Reset {
			m.scrollReset()
			return m, nil
		}
		m.scrollDelta(msg.Delta)
		return m, nil

	case msgs.TerminalResizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.mouseSelection = nil // coordinates become stale after resize
		innerCols := max(1, m.width-2)
		innerRows := max(1, m.height-2)
		m.vt.Resize(innerCols, innerRows)
		// Resize may shrink/reflow scrollback; clamp any stored offset so it
		// still points inside the valid range.
		if sbLen := m.vt.ScrollbackLen(); m.scrollOffset > sbLen {
			m.scrollOffset = sbLen
		}
		// Copy-mode cursor coordinates are viewport-local; clamp them to the
		// new viewport to avoid selecting past the edge after shrink.
		if m.copy != nil {
			if m.copy.cursor.x >= innerCols {
				m.copy.cursor.x = innerCols - 1
			}
			if m.copy.cursor.y >= innerRows {
				m.copy.cursor.y = innerRows - 1
			}
			if m.copy.anchor.x >= innerCols {
				m.copy.anchor.x = innerCols - 1
			}
			if m.copy.anchor.y >= innerRows {
				m.copy.anchor.y = innerRows - 1
			}
		}
		if m.ptyFile != nil {
			_ = pty.Setsize(m.ptyFile, &pty.Winsize{
				Rows: uint16(innerRows),
				Cols: uint16(innerCols),
			})
		}
		return m, nil
	}

	return m, nil
}

// View renders the terminal pane using the virtual terminal screen.
func (m Model) View() string {
	style := unfocusedBorder
	if m.focused {
		style = focusedBorder
	}
	if m.copy != nil {
		style = style.BorderForeground(lipgloss.Color("214")) // amber border in copy mode
	} else if m.mouseSelection != nil {
		style = style.BorderForeground(lipgloss.Color("75")) // blue border during mouse selection
	}
	var content string
	switch {
	case m.copy != nil:
		content = m.renderCopyMode()
	case m.mouseSelection != nil:
		content = m.renderWithMouseSelection()
	default:
		content = m.renderViewport()
	}
	return style.Width(m.width - 2).Height(m.height - 2).Render(content)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
