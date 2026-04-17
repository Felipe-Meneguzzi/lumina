package layout

import "github.com/menegas/lumina/msgs"

// PaneBounds returns the position and size of the pane with the given ID in
// layout-local coordinates (0,0 = top-left of the layout's content area).
// Returns ok=false when the pane is not found. Bounds match the geometry
// distributeSize uses, so the rectangle includes the pane's own border.
func (m Model) PaneBounds(id PaneID) (x, y, w, h int, ok bool) {
	return findBounds(m.root, id, 0, 0, m.width, m.height)
}

// FocusedBounds is a convenience wrapper around PaneBounds for the focused pane.
func (m Model) FocusedBounds() (x, y, w, h int, ok bool) {
	return m.PaneBounds(m.focused)
}

// FocusedMouseEnabled reports whether the focused pane is a terminal that the
// inner application has put into a mouse-tracking mode (1000/1002/1003).
// Returns false for editor panes or terminals not requesting mouse input.
func (m Model) FocusedMouseEnabled() bool {
	leaf := findLeaf(m.root, m.focused)
	if leaf == nil || leaf.Kind != KindTerminal {
		return false
	}
	inner := leafInner(leaf)
	if me, ok := inner.(interface{ MouseEnabled() bool }); ok {
		return me.MouseEnabled()
	}
	return false
}

// FocusedTitle returns the OSC 0/2 title last set by the focused terminal's
// inner application. Empty for editor panes or terminals that haven't reported.
func (m Model) FocusedTitle() string {
	leaf := findLeaf(m.root, m.focused)
	if leaf == nil || leaf.Kind != KindTerminal {
		return ""
	}
	if t, ok := leafInner(leaf).(interface{ Title() string }); ok {
		return t.Title()
	}
	return ""
}

// FocusedInCopyMode reports whether the focused pane is a terminal currently
// in tmux-style copy mode. Used by the app layer to route all keystrokes to
// the terminal (including printable runes) instead of forwarding to the PTY.
func (m Model) FocusedInCopyMode() bool {
	leaf := findLeaf(m.root, m.focused)
	if leaf == nil || leaf.Kind != KindTerminal {
		return false
	}
	if c, ok := leafInner(leaf).(interface{ InCopyMode() bool }); ok {
		return c.InCopyMode()
	}
	return false
}

// FocusedCWD returns the OSC 7 working directory last reported by the focused
// terminal's inner application. Empty for editors or terminals that haven't
// reported.
func (m Model) FocusedCWD() string {
	leaf := findLeaf(m.root, m.focused)
	if leaf == nil || leaf.Kind != KindTerminal {
		return ""
	}
	if c, ok := leafInner(leaf).(interface{ CWD() string }); ok {
		return c.CWD()
	}
	return ""
}

func findBounds(n PaneNode, id PaneID, ox, oy, w, h int) (int, int, int, int, bool) {
	switch v := n.(type) {
	case *LeafNode:
		if v.ID == id {
			return ox, oy, w, h, true
		}
	case *SplitNode:
		ratio := clampRatio(v.Ratio)
		switch v.Direction {
		case msgs.SplitHorizontal:
			firstW := max(1, int(float64(w)*ratio))
			secondW := max(1, w-firstW)
			if x, y, ww, hh, ok := findBounds(v.First, id, ox, oy, firstW, h); ok {
				return x, y, ww, hh, true
			}
			if x, y, ww, hh, ok := findBounds(v.Second, id, ox+firstW, oy, secondW, h); ok {
				return x, y, ww, hh, true
			}
		case msgs.SplitVertical:
			firstH := max(1, int(float64(h)*ratio))
			secondH := max(1, h-firstH)
			if x, y, ww, hh, ok := findBounds(v.First, id, ox, oy, w, firstH); ok {
				return x, y, ww, hh, true
			}
			if x, y, ww, hh, ok := findBounds(v.Second, id, ox, oy+firstH, w, secondH); ok {
				return x, y, ww, hh, true
			}
		}
	}
	return 0, 0, 0, 0, false
}
