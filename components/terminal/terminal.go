package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/creack/pty"
	vt10x "github.com/hinshun/vt10x"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

// Glyph attribute bit flags (matching vt10x internal constants).
const (
	attrReverse   int16 = 1 << 0
	attrUnderline int16 = 1 << 1
	attrBold      int16 = 1 << 2
	attrItalic    int16 = 1 << 4
)

var (
	// ThickBorder (focused) vs RoundedBorder (unfocused) differ even without color support.
	focusedBorder   = lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("62"))
	unfocusedBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))
)

// Model is the Bubble Tea model for the terminal pane.
type Model struct {
	shell        string
	forceTheme   bool
	ptyFile      *os.File
	cmd          *exec.Cmd
	vt           vt10x.Terminal  // VT100 screen buffer — handles all escape sequences
	reservedKeys map[string]bool // keys not to forward to PTY (global shortcuts)
	width        int
	height       int
	focused      bool
	closed       bool // true when running without a live PTY (used in tests)
	paneID       int  // identifies this terminal in multi-pane output routing
}

// New creates a new terminal Model and starts the shell process.
func New(cfg config.Config) (Model, error) {
	cols, rows := 78, 22 // default inner dimensions (width-2, height-2 for border)
	m := Model{
		shell:      cfg.Shell,
		forceTheme: cfg.ForceShellTheme,
		width:      80,
		height:     24,
		vt:         vt10x.New(vt10x.WithSize(cols, rows)),
	}

	if err := m.startShell(); err != nil {
		return m, fmt.Errorf("starting shell: %w", err)
	}
	return m, nil
}

// Close shuts down the PTY without restarting — used in tests to skip PTY creation.
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

	cmd := buildShellCommand(m.shell, m.forceTheme, &env)
	cmd.Env = env

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	cols, rows := m.vt.Size()
	if err := pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	}); err != nil {
		_ = ptmx.Close()
		return err
	}

	m.ptyFile = ptmx
	m.cmd = cmd
	return nil
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
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
			cols, rows := m.vt.Size()
			m.vt = vt10x.New(vt10x.WithSize(cols, rows))
			if err := m.startShell(); err != nil {
				return m, nil
			}
			return m, waitForOutput(m.ptyFile, m.paneID)
		}
		// Feed raw bytes to the VT100 emulator — it handles all escape sequences.
		_, _ = m.vt.Write(msg.Data)
		if m.ptyFile != nil {
			return m, waitForOutput(m.ptyFile, m.paneID)
		}
		return m, nil

	case msgs.PaneFocusMsg:
		m.focused = msg.Focused
		return m, nil

	case msgs.PtyInputMsg:
		if m.ptyFile != nil && len(msg.Data) > 0 {
			_, _ = m.ptyFile.Write(msg.Data)
		}
		return m, nil

	case msgs.TerminalResizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		innerCols := max(1, m.width-2)
		innerRows := max(1, m.height-2)
		m.vt.Resize(innerCols, innerRows)
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

// vtColor converts a vt10x Color to a lipgloss Color string.
// Returns "" for default/unset colors so lipgloss uses the terminal default.
func vtColor(c vt10x.Color) lipgloss.Color {
	if c >= vt10x.DefaultFG {
		return lipgloss.Color("")
	}
	return lipgloss.Color(fmt.Sprintf("%d", c))
}

// renderScreen converts the vt10x screen buffer to a styled string.
// Consecutive cells with the same style are batched for efficiency.
func renderScreen(vt vt10x.Terminal) string {
	cols, rows := vt.Size()
	var out strings.Builder

	for y := 0; y < rows; y++ {
		var (
			rowBuf  strings.Builder
			curFG   = vt10x.DefaultFG
			curBG   = vt10x.DefaultBG
			curMode int16
			pending strings.Builder
		)

		flush := func() {
			if pending.Len() == 0 {
				return
			}
			style := lipgloss.NewStyle()
			fg := vtColor(curFG)
			bg := vtColor(curBG)
			if fg != "" {
				style = style.Foreground(fg)
			}
			if bg != "" {
				style = style.Background(bg)
			}
			if curMode&attrBold != 0 {
				style = style.Bold(true)
			}
			if curMode&attrUnderline != 0 {
				style = style.Underline(true)
			}
			if curMode&attrItalic != 0 {
				style = style.Italic(true)
			}
			if curMode&attrReverse != 0 {
				style = style.Reverse(true)
			}
			rowBuf.WriteString(style.Render(pending.String()))
			pending.Reset()
		}

		for x := 0; x < cols; x++ {
			g := vt.Cell(x, y)
			if g.FG != curFG || g.BG != curBG || g.Mode != curMode {
				flush()
				curFG = g.FG
				curBG = g.BG
				curMode = g.Mode
			}
			ch := g.Char
			if ch == 0 {
				ch = ' '
			}
			pending.WriteRune(ch)
		}
		flush()

		if y < rows-1 {
			out.WriteString(rowBuf.String() + "\n")
		} else {
			out.WriteString(rowBuf.String())
		}
	}

	return out.String()
}

// View renders the terminal pane using the VT100 screen buffer.
func (m Model) View() string {
	style := unfocusedBorder
	if m.focused {
		style = focusedBorder
	}
	content := renderScreen(m.vt)
	return style.Width(m.width - 2).Height(m.height - 2).Render(content)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
