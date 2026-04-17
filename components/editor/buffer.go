package editor

// Buffer holds the text content as a slice of lines and tracks cursor position.
type Buffer struct {
	lines []string
	row   int
	col   int
}

// NewBuffer creates a Buffer pre-populated with the given lines.
func NewBuffer(lines []string) Buffer {
	if len(lines) == 0 {
		lines = []string{""}
	}
	return Buffer{lines: lines}
}

// Line returns the content of the given row.
func (b *Buffer) Line(row int) string {
	if row < 0 || row >= len(b.lines) {
		return ""
	}
	return b.lines[row]
}

// LineCount returns the number of lines.
func (b *Buffer) LineCount() int { return len(b.lines) }

// Cursor returns the current (row, col) position.
func (b *Buffer) Cursor() (int, int) { return b.row, b.col }

// InsertAt inserts character ch before position col on the given row.
func (b *Buffer) InsertAt(row, col int, ch rune) {
	if row < 0 || row >= len(b.lines) {
		return
	}
	runes := []rune(b.lines[row])
	col = clamp(col, 0, len(runes))
	runes = append(runes[:col], append([]rune{ch}, runes[col:]...)...)
	b.lines[row] = string(runes)
}

// DeleteAt removes the character at 0-indexed position col on the given row.
// col must be a valid index (0 <= col < len(runes)); out-of-range calls are no-ops.
func (b *Buffer) DeleteAt(row, col int) {
	if row < 0 || row >= len(b.lines) {
		return
	}
	runes := []rune(b.lines[row])
	if col < 0 || col >= len(runes) {
		return
	}
	b.lines[row] = string(append(runes[:col], runes[col+1:]...))
}

// SplitLine breaks line at row into two lines at position col (Enter key).
func (b *Buffer) SplitLine(row, col int) {
	if row < 0 || row >= len(b.lines) {
		return
	}
	runes := []rune(b.lines[row])
	col = clamp(col, 0, len(runes))
	before := string(runes[:col])
	after := string(runes[col:])
	b.lines[row] = before
	tail := make([]string, len(b.lines)-row-1)
	copy(tail, b.lines[row+1:])
	b.lines = append(b.lines[:row+1], append([]string{after}, tail...)...)
}

// JoinLines merges lines[row] and lines[row+1] into lines[row].
func (b *Buffer) JoinLines(row int) {
	if row < 0 || row+1 >= len(b.lines) {
		return
	}
	b.lines[row] = b.lines[row] + b.lines[row+1]
	b.lines = append(b.lines[:row+1], b.lines[row+2:]...)
}

// MoveCursor moves the cursor by (dr, dc) rows and columns, clamping to buffer bounds.
func (b *Buffer) MoveCursor(dr, dc int) {
	b.row = clamp(b.row+dr, 0, len(b.lines)-1)
	lineLen := len([]rune(b.lines[b.row]))
	b.col = clamp(b.col+dc, 0, lineLen)
}

// SetCursor sets the cursor position directly.
func (b *Buffer) SetCursor(row, col int) {
	b.row = clamp(row, 0, len(b.lines)-1)
	lineLen := len([]rune(b.lines[b.row]))
	b.col = clamp(col, 0, lineLen)
}

// Content returns the full buffer as a single string.
func (b *Buffer) Content() string {
	result := ""
	for i, l := range b.lines {
		if i > 0 {
			result += "\n"
		}
		result += l
	}
	return result
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
