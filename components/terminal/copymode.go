package terminal

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	osc52 "github.com/aymanbagabas/go-osc52/v2"
	"github.com/charmbracelet/lipgloss"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/menegas/lumina/msgs"
)

// copyState holds the cursor + selection anchor for tmux-style copy mode.
// Coordinates are viewport-local: (0,0) = top-left of the currently rendered
// terminal area (which may include scrollback rows when m.scrollOffset > 0).
type copyState struct {
	cursor pos
	anchor pos // selection start; equals cursor when no selection
}

type pos struct{ x, y int }

// selectionStyle highlights cells inside the selection rectangle.
var selectionStyle = lipgloss.NewStyle().Reverse(true)

// InCopyMode reports whether the terminal is currently in copy mode.
func (m Model) InCopyMode() bool { return m.copy != nil }

// enterCopyMode initialises copy mode at the bottom-right cell of the viewport.
func (m *Model) enterCopyMode() {
	if m.copy != nil {
		return
	}
	cols, rows := m.vt.Width(), m.vt.Height()
	if cols <= 0 || rows <= 0 {
		return
	}
	c := pos{x: cols - 1, y: rows - 1}
	m.copy = &copyState{cursor: c, anchor: c}
}

// exitCopyMode discards copy state and returns to normal terminal rendering.
func (m *Model) exitCopyMode() { m.copy = nil }

// handleCopyKey processes a key press while in copy mode. Returns a Cmd that
// may emit OSC52 clipboard bytes when the user confirms a copy.
func (m *Model) handleCopyKey(msg tea.KeyMsg) tea.Cmd {
	if m.copy == nil {
		return nil
	}
	cols, rows := m.vt.Width(), m.vt.Height()
	c := m.copy
	shift := false

	switch msg.String() {
	case "esc", "q", "ctrl+c":
		m.exitCopyMode()
		return nil

	case "y", "enter":
		text := m.extractSelection()
		m.exitCopyMode()
		return copyToClipboard(text)

	case "h", "left":
		c.cursor.x--
	case "l", "right":
		c.cursor.x++
	case "j", "down":
		c.cursor.y++
	case "k", "up":
		c.cursor.y--

	// Shift+motion extends the selection (anchor stays put).
	case "H", "shift+left":
		c.cursor.x--
		shift = true
	case "L", "shift+right":
		c.cursor.x++
		shift = true
	case "J", "shift+down":
		c.cursor.y++
		shift = true
	case "K", "shift+up":
		c.cursor.y--
		shift = true

	// Jump-to commands.
	case "0", "home":
		c.cursor.x = 0
	case "$", "end":
		c.cursor.x = cols - 1
	case "g":
		c.cursor.y = 0
	case "G":
		c.cursor.y = rows - 1

	// Toggle selection: collapse to point or restart selection here.
	case "v":
		c.anchor = c.cursor
		return nil
	}

	// Clamp cursor to viewport.
	if c.cursor.x < 0 {
		c.cursor.x = 0
	}
	if c.cursor.x >= cols {
		c.cursor.x = cols - 1
	}
	if c.cursor.y < 0 {
		c.cursor.y = 0
	}
	if c.cursor.y >= rows {
		c.cursor.y = rows - 1
	}
	if !shift {
		c.anchor = c.cursor
	}
	return nil
}

// extractSelection returns the plain text inside the current selection
// rectangle (block-mode, like tmux without rectangle-selection toggle).
func (m Model) extractSelection() string {
	if m.copy == nil {
		return ""
	}
	x0, y0 := m.copy.anchor.x, m.copy.anchor.y
	x1, y1 := m.copy.cursor.x, m.copy.cursor.y
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	cols, _ := m.vt.Width(), m.vt.Height()
	sbLen := m.vt.ScrollbackLen()
	// Translate viewport y to combined index (sbStart at top of viewport).
	sbStart := viewportStart(m.vt, m.scrollOffset, sbLen)
	var out strings.Builder
	for y := y0; y <= y1; y++ {
		idx := sbStart + y
		var line strings.Builder
		for x := 0; x < cols && x <= x1; x++ {
			if x < x0 {
				continue
			}
			var cell *uv.Cell
			if idx < sbLen {
				cell = m.vt.ScrollbackCellAt(x, idx)
			} else {
				cell = m.vt.CellAt(x, idx-sbLen)
			}
			if cell == nil || cell.Content == "" {
				line.WriteByte(' ')
			} else {
				line.WriteString(cell.Content)
			}
		}
		// Trim trailing spaces per line — common terminal copy convention.
		out.WriteString(strings.TrimRight(line.String(), " "))
		if y < y1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// renderCopyMode draws the viewport with a selection rectangle highlighted
// and an inverse-video cursor. Used in place of renderViewport when
// m.copy != nil.
func (m Model) renderCopyMode() string {
	cols, rows := m.vt.Width(), m.vt.Height()
	sbLen := m.vt.ScrollbackLen()
	sbStart := viewportStart(m.vt, m.scrollOffset, sbLen)
	x0, y0 := m.copy.anchor.x, m.copy.anchor.y
	x1, y1 := m.copy.cursor.x, m.copy.cursor.y
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}

	var out strings.Builder
	for y := 0; y < rows; y++ {
		idx := sbStart + y
		for x := 0; x < cols; x++ {
			var cell *uv.Cell
			if idx < sbLen {
				cell = m.vt.ScrollbackCellAt(x, idx)
			} else {
				cell = m.vt.CellAt(x, idx-sbLen)
			}
			content := " "
			if cell != nil && cell.Content != "" {
				content = cell.Content
			}
			styled := content
			if cell != nil && !cell.Style.IsZero() {
				styled = cell.Style.Styled(content)
			}
			isCursor := x == m.copy.cursor.x && y == m.copy.cursor.y
			inSel := y >= y0 && y <= y1 && x >= x0 && x <= x1
			if isCursor || inSel {
				out.WriteString(selectionStyle.Render(styled))
			} else {
				out.WriteString(styled)
			}
		}
		if y < rows-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// copyToClipboard returns a Cmd that writes an OSC 52 sequence to stderr so
// the host terminal copies the given text to the system clipboard. The Cmd
// also surfaces a status notification confirming the copy.
func copyToClipboard(text string) tea.Cmd {
	if text == "" {
		return func() tea.Msg {
			return msgs.StatusBarNotifyMsg{
				Text:     "copy: nada selecionado",
				Level:    msgs.NotifyWarning,
				Duration: 2_000_000_000,
			}
		}
	}
	return func() tea.Msg {
		// OSC 52 reaches the host terminal even on stderr; using stderr avoids
		// fighting the Bubble Tea renderer that owns stdout.
		_, _ = fmt.Fprint(os.Stderr, osc52.New(text).String())
		return msgs.StatusBarNotifyMsg{
			Text:     fmt.Sprintf("copiado: %d caractere(s)", len(text)),
			Level:    msgs.NotifyInfo,
			Duration: 2_000_000_000,
		}
	}
}
