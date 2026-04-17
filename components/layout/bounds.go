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

// FocusedHasMouseSelection reports whether the focused terminal pane has an
// active mouse selection (drag in progress or pending confirmation).
func (m Model) FocusedHasMouseSelection() bool {
	leaf := findLeaf(m.root, m.focused)
	if leaf == nil || leaf.Kind != KindTerminal {
		return false
	}
	if s, ok := leafInner(leaf).(interface{ HasMouseSelection() bool }); ok {
		return s.HasMouseSelection()
	}
	return false
}

// FocusedHasPendingSelection reports whether the focused terminal pane has a
// mouse selection that is waiting for explicit 'y' confirmation before being
// copied (mouse_auto_copy=false path).
func (m Model) FocusedHasPendingSelection() bool {
	leaf := findLeaf(m.root, m.focused)
	if leaf == nil || leaf.Kind != KindTerminal {
		return false
	}
	if s, ok := leafInner(leaf).(interface{ HasPendingSelection() bool }); ok {
		return s.HasPendingSelection()
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

// HitTest returns the PaneID under the given layout-local coordinates
// (0,0 = top-left of the layout content area). localX/localY are translated
// into the pane's content-cell space — subtracting the pane's origin plus
// one cell for the border (so (0,0) is the first cell inside the border).
// Coordinates falling on the pane border itself still map to the containing
// pane with clamped local coordinates.
func (m Model) HitTest(x, y int) (paneID PaneID, target msgs.FocusTarget, localX, localY int, ok bool) {
	px, py, pw, ph, leaf, found := findHit(m.root, 0, 0, m.width, m.height, x, y)
	if !found || leaf == nil {
		return 0, msgs.FocusLayout, 0, 0, false
	}
	localX = x - px - 1
	localY = y - py - 1
	if localX < 0 {
		localX = 0
	}
	if localY < 0 {
		localY = 0
	}
	if pw > 2 && localX > pw-3 {
		localX = pw - 3
	}
	if ph > 2 && localY > ph-3 {
		localY = ph - 3
	}
	return leaf.ID, msgs.FocusTerminal, localX, localY, true
}

// findHit descends the tree looking for the leaf whose bounds contain (x, y).
// Returns the leaf's bounds rectangle and the leaf itself.
func findHit(n PaneNode, ox, oy, w, h, x, y int) (int, int, int, int, *LeafNode, bool) {
	if x < ox || y < oy || x >= ox+w || y >= oy+h {
		return 0, 0, 0, 0, nil, false
	}
	switch v := n.(type) {
	case *LeafNode:
		return ox, oy, w, h, v, true
	case *SplitNode:
		ratio := clampRatio(v.Ratio)
		switch v.Direction {
		case msgs.SplitHorizontal:
			firstW := max(1, int(float64(w)*ratio))
			secondW := max(1, w-firstW)
			if px, py, pw, ph, leaf, ok := findHit(v.First, ox, oy, firstW, h, x, y); ok {
				return px, py, pw, ph, leaf, true
			}
			return findHit(v.Second, ox+firstW, oy, secondW, h, x, y)
		case msgs.SplitVertical:
			firstH := max(1, int(float64(h)*ratio))
			secondH := max(1, h-firstH)
			if px, py, pw, ph, leaf, ok := findHit(v.First, ox, oy, w, firstH, x, y); ok {
				return px, py, pw, ph, leaf, true
			}
			return findHit(v.Second, ox, oy+firstH, w, secondH, x, y)
		}
	}
	return 0, 0, 0, 0, nil, false
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
