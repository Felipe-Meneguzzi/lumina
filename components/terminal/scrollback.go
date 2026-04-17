package terminal

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

// scrollbackMax caps the number of history rows the emulator keeps.
const scrollbackMax = 2000

// renderViewport returns the terminal content as a styled string.
//
// When the user is viewing the live screen (or the alternate screen is
// active — alt-screen apps don't write to scrollback), it delegates to
// the emulator's own renderer. When scrolled into history on the main
// screen, it composes scrollback rows + live rows manually so the user
// can see prior output.
func (m Model) renderViewport() string {
	if m.scrollOffset <= 0 || m.vt.IsAltScreen() {
		if m.focused && m.state.cursorVisible() {
			pos := m.vt.CursorPosition()
			cols, rows := m.vt.Width(), m.vt.Height()
			sbLen := m.vt.ScrollbackLen()
			var out strings.Builder
			for y := 0; y < rows; y++ {
				cursorX := -1
				if y == pos.Y {
					cursorX = pos.X
				}
				writeRow(&out, m.vt, sbLen+y, sbLen, cols, cursorX)
				if y < rows-1 {
					out.WriteByte('\n')
				}
			}
			return out.String()
		}
		return m.vt.Render()
	}
	cols, rows := m.vt.Width(), m.vt.Height()
	sbLen := m.vt.ScrollbackLen()
	offset := m.scrollOffset
	if offset > sbLen {
		offset = sbLen
	}
	// sbStart is the index of the oldest visible scrollback line.
	// Combined indices: [sbStart..sbLen) = scrollback, [sbLen..sbLen+rows) = live rows.
	sbStart := sbLen - offset

	cursorViewRow := -1
	cursorX := -1
	if m.focused && m.state.cursorVisible() {
		pos := m.vt.CursorPosition()
		combined := sbLen + pos.Y
		if combined >= sbStart && combined < sbStart+rows {
			cursorViewRow = combined - sbStart
			cursorX = pos.X
		}
	}

	var out strings.Builder
	for i := 0; i < rows; i++ {
		idx := sbStart + i
		cx := -1
		if i == cursorViewRow {
			cx = cursorX
		}
		writeRow(&out, m.vt, idx, sbLen, cols, cx)
		if i < rows-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// writeRow appends a single row's worth of styled cells to out. When idx is
// below sbLen the row comes from scrollback; otherwise from the live screen.
// cursorX is the column to render as a block cursor (-1 = no cursor).
func writeRow(out *strings.Builder, e vtReader, idx, sbLen, cols, cursorX int) {
	for x := 0; x < cols; x++ {
		var cell *uv.Cell
		if idx < sbLen {
			cell = e.ScrollbackCellAt(x, idx)
		} else {
			cell = e.CellAt(x, idx-sbLen)
		}
		content := " "
		if cell != nil && cell.Content != "" {
			content = cell.Content
		}
		if x == cursorX {
			cs := uv.Style{}
			if cell != nil {
				cs = cell.Style
			}
			cs.Attrs |= uv.AttrReverse
			out.WriteString(cs.Styled(content))
		} else if cell != nil && !cell.Style.IsZero() {
			out.WriteString(cell.Style.Styled(content))
		} else {
			out.WriteString(content)
		}
	}
}

// vtReader is the subset of *vt.Emulator that renderViewport needs. Defined
// as a local interface so writeRow can be unit-tested without a full PTY.
type vtReader interface {
	ScrollbackCellAt(x, y int) *uv.Cell
	CellAt(x, y int) *uv.Cell
}

// viewportStart returns the combined-index of the topmost visible row.
// When the user is live (offset 0) or the alt-screen is active, the viewport
// starts at the first live row (combined index = sbLen). Otherwise it climbs
// back into scrollback by offset, clamped to [0, sbLen].
func viewportStart(e altScreenReader, scrollOffset, sbLen int) int {
	if scrollOffset <= 0 || e.IsAltScreen() {
		return sbLen
	}
	offset := scrollOffset
	if offset > sbLen {
		offset = sbLen
	}
	return sbLen - offset
}

// altScreenReader is the subset of *vt.Emulator that viewportStart needs.
type altScreenReader interface {
	IsAltScreen() bool
}

// scrollDelta applies a positive (into history) or negative (toward live)
// change to the current scroll offset, clamping to [0, ScrollbackLen()].
func (m *Model) scrollDelta(delta int) {
	// Alt-screen apps (vim, claude, htop) own the full viewport; scrollback
	// doesn't apply and attempting to scroll would corrupt the offset used
	// when returning to the main screen.
	if m.vt.IsAltScreen() {
		return
	}
	m.scrollOffset += delta
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	if sbLen := m.vt.ScrollbackLen(); m.scrollOffset > sbLen {
		m.scrollOffset = sbLen
	}
}

// scrollReset snaps the view back to the live terminal contents.
func (m *Model) scrollReset() {
	m.scrollOffset = 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
