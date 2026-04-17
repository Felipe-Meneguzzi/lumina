// Mouse selection mode (normal mode) — click-and-drag text selection
// without entering the keyboard-driven copy mode.
package terminal

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	uv "github.com/charmbracelet/ultraviolet"
)

// mouseSelection holds the active mouse-driven text selection in viewport-local
// coordinates (0,0 = top-left of the pane's content area). Non-nil when a drag
// is in progress or the selection is pending explicit confirmation.
// start is always the drag anchor; end is the current cursor position.
type mouseSelection struct {
	start   pos
	end     pos
	pending bool // true when mouse_auto_copy=false and waiting for 'y'
}

// selectionMode normalises the config string and defaults to "linear".
func selectionMode(s string) string {
	if s == "block" {
		return "block"
	}
	return "linear"
}

// HasMouseSelection reports whether a mouse selection is currently active.
func (m Model) HasMouseSelection() bool { return m.mouseSelection != nil }

// HasPendingSelection reports whether a selection exists and is waiting for
// explicit 'y' confirmation before being copied to the clipboard.
func (m Model) HasPendingSelection() bool {
	return m.mouseSelection != nil && m.mouseSelection.pending
}

func clampPos(p pos, cols, rows int) pos {
	if cols <= 0 {
		cols = 1
	}
	if rows <= 0 {
		rows = 1
	}
	if p.x < 0 {
		p.x = 0
	}
	if p.x >= cols {
		p.x = cols - 1
	}
	if p.y < 0 {
		p.y = 0
	}
	if p.y >= rows {
		p.y = rows - 1
	}
	return p
}

func (m *Model) startMouseSelection(x, y int) {
	cols, rows := m.vt.Width(), m.vt.Height()
	p := clampPos(pos{x: x, y: y}, cols, rows)
	m.mouseSelection = &mouseSelection{start: p, end: p}
}

func (m *Model) updateMouseSelection(x, y int) {
	if m.mouseSelection == nil {
		return
	}
	cols, rows := m.vt.Width(), m.vt.Height()
	m.mouseSelection.end = clampPos(pos{x: x, y: y}, cols, rows)
}

func (m *Model) finalizeMouseSelection(x, y int, autoCopy bool) tea.Cmd {
	if m.mouseSelection == nil {
		return nil
	}
	cols, rows := m.vt.Width(), m.vt.Height()
	m.mouseSelection.end = clampPos(pos{x: x, y: y}, cols, rows)
	if autoCopy {
		text := m.extractMouseSelection()
		m.mouseSelection = nil
		return copyToClipboard(text)
	}
	m.mouseSelection.pending = true
	return nil
}

func (m *Model) confirmMouseSelection() tea.Cmd {
	if m.mouseSelection == nil {
		return nil
	}
	text := m.extractMouseSelection()
	m.mouseSelection = nil
	return copyToClipboard(text)
}

func (m *Model) cancelMouseSelection() {
	m.mouseSelection = nil
}

// linearBounds normalises anchor+cursor into (firstLine, firstX, lastLine, lastX)
// so that firstLine <= lastLine and, on the same line, firstX <= lastX.
// The anchor is preserved as the drag origin; cursor is the current position.
func linearBounds(anchor, cursor pos) (firstLine, firstX, lastLine, lastX int) {
	if cursor.y > anchor.y || (cursor.y == anchor.y && cursor.x >= anchor.x) {
		return anchor.y, anchor.x, cursor.y, cursor.x
	}
	return cursor.y, cursor.x, anchor.y, anchor.x
}

// isInLinear reports whether cell (x, y) falls inside a linear (notepad-style)
// selection. The selection covers:
//   - first line: columns firstX … end-of-line
//   - middle lines: entire line
//   - last line: columns 0 … lastX
//   - single-line: firstX … lastX
func isInLinear(x, y, firstLine, firstX, lastLine, lastX int) bool {
	if y < firstLine || y > lastLine {
		return false
	}
	if firstLine == lastLine {
		return x >= firstX && x <= lastX
	}
	if y == firstLine {
		return x >= firstX
	}
	if y == lastLine {
		return x <= lastX
	}
	return true
}

func (m Model) extractMouseSelection() string {
	if m.mouseSelection == nil {
		return ""
	}
	if m.mouseSelMode == "block" {
		return m.extractBlockSelection()
	}
	return m.extractLinearSelection()
}

func (m Model) extractLinearSelection() string {
	fl, fx, ll, lx := linearBounds(m.mouseSelection.start, m.mouseSelection.end)
	cols := m.vt.Width()
	sbLen := m.vt.ScrollbackLen()
	sbStart := viewportStart(m.vt, m.scrollOffset, sbLen)
	var out strings.Builder
	for y := fl; y <= ll; y++ {
		idx := sbStart + y
		var xFrom, xTo int
		switch {
		case fl == ll:
			xFrom, xTo = fx, lx
		case y == fl:
			xFrom, xTo = fx, cols-1
		case y == ll:
			xFrom, xTo = 0, lx
		default:
			xFrom, xTo = 0, cols-1
		}
		var line strings.Builder
		for x := xFrom; x <= xTo; x++ {
			line.WriteString(m.cellContent(x, idx, sbLen))
		}
		out.WriteString(strings.TrimRight(line.String(), " "))
		if y < ll {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

func (m Model) extractBlockSelection() string {
	x0, y0 := m.mouseSelection.start.x, m.mouseSelection.start.y
	x1, y1 := m.mouseSelection.end.x, m.mouseSelection.end.y
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	sbLen := m.vt.ScrollbackLen()
	sbStart := viewportStart(m.vt, m.scrollOffset, sbLen)
	var out strings.Builder
	for y := y0; y <= y1; y++ {
		idx := sbStart + y
		var line strings.Builder
		for x := x0; x <= x1; x++ {
			line.WriteString(m.cellContent(x, idx, sbLen))
		}
		out.WriteString(strings.TrimRight(line.String(), " "))
		if y < y1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// cellContent returns the display string for cell (x, idx) where idx is an
// absolute scrollback index. Returns a space for nil or empty cells.
func (m Model) cellContent(x, idx, sbLen int) string {
	var cell *uv.Cell
	if idx < sbLen {
		cell = m.vt.ScrollbackCellAt(x, idx)
	} else {
		cell = m.vt.CellAt(x, idx-sbLen)
	}
	if cell == nil || cell.Content == "" {
		return " "
	}
	return cell.Content
}

func (m Model) renderWithMouseSelection() string {
	if m.mouseSelection == nil {
		return m.renderViewport()
	}
	if m.mouseSelMode == "block" {
		return m.renderBlockSelection()
	}
	return m.renderLinearSelection()
}

func (m Model) renderLinearSelection() string {
	fl, fx, ll, lx := linearBounds(m.mouseSelection.start, m.mouseSelection.end)
	cols, rows := m.vt.Width(), m.vt.Height()
	sbLen := m.vt.ScrollbackLen()
	sbStart := viewportStart(m.vt, m.scrollOffset, sbLen)
	var out strings.Builder
	for y := 0; y < rows; y++ {
		idx := sbStart + y
		for x := 0; x < cols; x++ {
			cell := m.rawCell(x, idx, sbLen)
			styled := m.styledCell(cell)
			if isInLinear(x, y, fl, fx, ll, lx) {
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

func (m Model) renderBlockSelection() string {
	x0, y0 := m.mouseSelection.start.x, m.mouseSelection.start.y
	x1, y1 := m.mouseSelection.end.x, m.mouseSelection.end.y
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	cols, rows := m.vt.Width(), m.vt.Height()
	sbLen := m.vt.ScrollbackLen()
	sbStart := viewportStart(m.vt, m.scrollOffset, sbLen)
	var out strings.Builder
	for y := 0; y < rows; y++ {
		idx := sbStart + y
		for x := 0; x < cols; x++ {
			cell := m.rawCell(x, idx, sbLen)
			styled := m.styledCell(cell)
			if y >= y0 && y <= y1 && x >= x0 && x <= x1 {
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

// rawCell returns the *uv.Cell at absolute scrollback index idx, column x.
func (m Model) rawCell(x, idx, sbLen int) *uv.Cell {
	if idx < sbLen {
		return m.vt.ScrollbackCellAt(x, idx)
	}
	return m.vt.CellAt(x, idx-sbLen)
}

// styledCell returns the display string for a cell, applying its stored style.
func (m Model) styledCell(cell *uv.Cell) string {
	content := " "
	if cell != nil && cell.Content != "" {
		content = cell.Content
	}
	if cell != nil && !cell.Style.IsZero() {
		return cell.Style.Styled(content)
	}
	return content
}
