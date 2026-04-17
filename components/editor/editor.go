package editor

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

var (
	focusedBorder   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))
	unfocusedBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))
	cursorStyle     = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("0"))
	lineNumStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// Model is the Bubble Tea model for the text editor pane.
type Model struct {
	buf      Buffer
	path     string
	dirty    bool
	viewport viewport.Model
	focused  bool
	width    int
	height   int
	open     bool
}

// New returns an empty, closed editor.
func New(cfg config.Config) Model {
	_ = cfg
	m := Model{
		buf:    NewBuffer([]string{""}),
		width:  80,
		height: 24,
	}
	m.viewport = viewport.New(max(1, m.width-2), max(1, m.height-2))
	return m
}

// Open loads a file into a new editor Model.
func Open(path string, cfg config.Config) (Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return New(cfg), fmt.Errorf("opening %s: %w", path, err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	m := New(cfg)
	m.buf = NewBuffer(lines)
	m.path = path
	m.open = true
	m.refreshViewport()
	return m, nil
}

// SetFocused sets focus state.
func (m *Model) SetFocused(f bool) { m.focused = f }

// Dirty reports whether the buffer has unsaved changes.
func (m Model) Dirty() bool { return m.dirty }

// LineCount returns the number of lines in the buffer.
func (m Model) LineCount() int { return m.buf.LineCount() }

func (m *Model) refreshViewport() {
	m.viewport.SetContent(m.renderContent())
}

func (m Model) renderContent() string {
	curRow, curCol := m.buf.Cursor()
	var sb strings.Builder
	for i := 0; i < m.buf.LineCount(); i++ {
		line := m.buf.Line(i)
		lineNum := lineNumStyle.Render(fmt.Sprintf("%4d  ", i+1))
		if i == curRow {
			runes := []rune(line)
			col := clamp(curCol, 0, len(runes))
			var before, cursor, after string
			before = string(runes[:col])
			if col < len(runes) {
				cursor = cursorStyle.Render(string(runes[col : col+1]))
				after = string(runes[col+1:])
			} else {
				cursor = cursorStyle.Render(" ")
			}
			sb.WriteString(lineNum + before + cursor + after)
		} else {
			sb.WriteString(lineNum + line)
		}
		if i < m.buf.LineCount()-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// Init is a no-op for the editor (no background Cmds needed at start).
func (m Model) Init() tea.Cmd { return nil }

// Update handles messages for the editor pane.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case msgs.OpenFileMsg:
		loaded, err := Open(msg.Path, config.Config{})
		if err != nil {
			return m, func() tea.Msg {
				return msgs.StatusBarNotifyMsg{
					Text:     "Erro ao abrir: " + err.Error(),
					Level:    msgs.NotifyError,
					Duration: 4 * time.Second,
				}
			}
		}
		loaded.focused = m.focused
		loaded.width = m.width
		loaded.height = m.height
		loaded.viewport = viewport.New(max(1, m.width-2), max(1, m.height-2))
		loaded.refreshViewport()
		return loaded, nil

	case msgs.EditorResizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport = viewport.New(max(1, m.width-2), max(1, m.height-2))
		m.refreshViewport()
		return m, nil

	case msgs.CloseConfirmedMsg:
		m.open = false
		m.dirty = false
		m.buf = NewBuffer([]string{""})
		m.path = ""
		return m, nil

	case msgs.CloseAbortedMsg:
		return m, nil

	case msgs.PaneFocusMsg:
		m.focused = msg.Focused
		return m, nil

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		row, col := m.buf.Cursor()
		for _, r := range msg.Runes {
			m.buf.InsertAt(row, col, r)
			col++
		}
		m.buf.SetCursor(row, col)
		m.dirty = true
		m.refreshViewport()

	case tea.KeyBackspace:
		row, col := m.buf.Cursor()
		if col > 0 {
			m.buf.DeleteAt(row, col-1) // delete char to the left of cursor
			m.buf.SetCursor(row, col-1)
		} else if row > 0 {
			prevLen := len([]rune(m.buf.Line(row - 1)))
			m.buf.JoinLines(row - 1)
			m.buf.SetCursor(row-1, prevLen)
		}
		m.dirty = true
		m.refreshViewport()

	case tea.KeyEnter:
		row, col := m.buf.Cursor()
		m.buf.SplitLine(row, col)
		m.buf.SetCursor(row+1, 0)
		m.dirty = true
		m.refreshViewport()

	case tea.KeyUp:
		m.buf.MoveCursor(-1, 0)
		m.refreshViewport()

	case tea.KeyDown:
		m.buf.MoveCursor(1, 0)
		m.refreshViewport()

	case tea.KeyLeft:
		m.buf.MoveCursor(0, -1)
		m.refreshViewport()

	case tea.KeyRight:
		m.buf.MoveCursor(0, 1)
		m.refreshViewport()

	case tea.KeyCtrlS:
		if m.path != "" {
			if err := os.WriteFile(m.path, []byte(m.buf.Content()+"\n"), 0644); err == nil {
				m.dirty = false
				return m, func() tea.Msg {
					return msgs.StatusBarNotifyMsg{
						Text:     "Salvo",
						Level:    msgs.NotifyInfo,
						Duration: 2 * time.Second,
					}
				}
			}
		}

	case tea.KeyCtrlW:
		if m.dirty {
			return m, func() tea.Msg { return msgs.ConfirmCloseMsg{} }
		}
		m.open = false

	case tea.KeyPgDown:
		m.viewport.HalfViewDown()

	case tea.KeyPgUp:
		m.viewport.HalfViewUp()
	}

	return m, nil
}

// View renders the editor pane.
func (m Model) View() string {
	if !m.open {
		style := unfocusedBorder
		if m.focused {
			style = focusedBorder
		}
		placeholder := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("  No file open")
		return style.Width(m.width - 2).Height(m.height - 2).Render(placeholder)
	}

	title := m.path
	if m.dirty {
		title += " [*]"
	}

	style := unfocusedBorder
	if m.focused {
		style = focusedBorder
	}

	header := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(title) + "\n"
	content := m.viewport.View()
	return style.Width(m.width - 2).Height(m.height - 2).Render(header + content)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
