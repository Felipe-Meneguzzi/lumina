package terminal

import (
	vt10x "github.com/hinshun/vt10x"
)

// scrollbackMax caps the number of history rows the terminal keeps.
const scrollbackMax = 2000

// snapshotRow returns a copy of the glyphs on a given terminal row.
func snapshotRow(vt vt10x.Terminal, y, cols int) []vt10x.Glyph {
	row := make([]vt10x.Glyph, cols)
	for x := 0; x < cols; x++ {
		row[x] = vt.Cell(x, y)
	}
	return row
}

// rowEmpty reports whether a glyph row contains only whitespace/null runes
// with no explicit coloring. Empty rows are dropped before they reach scrollback.
func rowEmpty(row []vt10x.Glyph) bool {
	for _, g := range row {
		if g.Char != 0 && g.Char != ' ' {
			return false
		}
	}
	return true
}

// rowsEqual compares two glyph rows for character + style equality.
func rowsEqual(a, b []vt10x.Glyph) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Char != b[i].Char || a[i].FG != b[i].FG ||
			a[i].BG != b[i].BG || a[i].Mode != b[i].Mode {
			return false
		}
	}
	return true
}

// writeWithScrollback writes data to the terminal while capturing rows that
// get scrolled off the top into the scrollback buffer. The input is split at
// newlines so each potential scroll point is evaluated independently.
//
// Returns the number of rows that were appended to scrollback, which the
// caller uses to preserve the user's scroll offset when they are in history.
func (m *Model) writeWithScrollback(data []byte) int {
	cols, rows := m.vt.Size()
	pushed := 0

	// Split at newlines so we can detect scrolls per-line.
	start := 0
	for i := 0; i <= len(data); i++ {
		atEnd := i == len(data)
		if !atEnd && data[i] != '\n' {
			continue
		}
		chunk := data[start : min(i+1, len(data))]
		if len(chunk) == 0 {
			start = i + 1
			continue
		}

		// Snapshot all rows prior to the write so we can identify any that
		// shifted off the top. Only rows 0..rows-1 can be lost to a scroll.
		preRows := make([][]vt10x.Glyph, rows)
		for y := 0; y < rows; y++ {
			preRows[y] = snapshotRow(m.vt, y, cols)
		}

		_, _ = m.vt.Write(chunk)

		// Determine the shift amount: how many rows of preRows were dropped.
		// Post-row 0 should match preRow[shift] if the screen scrolled by `shift`.
		post0 := snapshotRow(m.vt, 0, cols)
		shift := 0
		for k := 1; k < rows; k++ {
			if rowsEqual(post0, preRows[k]) {
				shift = k
				break
			}
		}

		for s := 0; s < shift; s++ {
			if rowEmpty(preRows[s]) {
				continue
			}
			m.scrollback = append(m.scrollback, preRows[s])
			pushed++
		}
		start = i + 1
	}

	// Trim scrollback to max, remembering how many we dropped so the caller
	// can clamp any stored offset if needed.
	if overflow := len(m.scrollback) - scrollbackMax; overflow > 0 {
		m.scrollback = m.scrollback[overflow:]
	}
	return pushed
}

// scrollDelta applies a positive (into history) or negative (toward live)
// change to the current scroll offset, clamping to [0, len(scrollback)].
func (m *Model) scrollDelta(delta int) {
	m.scrollOffset += delta
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	if m.scrollOffset > len(m.scrollback) {
		m.scrollOffset = len(m.scrollback)
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
