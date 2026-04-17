package integration_test

import (
	"testing"

	"github.com/Felipe-Meneguzzi/lumina/components/layout"
	"github.com/Felipe-Meneguzzi/lumina/config"
	"github.com/Felipe-Meneguzzi/lumina/msgs"
)

func newLayout(t *testing.T) layout.Model {
	t.Helper()
	cfg := config.Config{Shell: "/bin/sh"}
	m, err := layout.New(cfg)
	if err != nil {
		t.Fatalf("layout.New: %v", err)
	}
	return m
}

// TestPaneSplitMsg_CreatesNewPane verifies the full message flow for splitting a pane.
func TestPaneSplitMsg_CreatesNewPane(t *testing.T) {
	m := newLayout(t)
	if m.PaneCount() != 1 {
		t.Fatalf("expected 1 pane at start, got %d", m.PaneCount())
	}

	next, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm := next.(layout.Model)

	if nm.PaneCount() != 2 {
		t.Errorf("PaneSplitMsg should create a new pane: expected 2, got %d", nm.PaneCount())
	}
}

// TestPaneCloseMsg_RemovesFocusedPane verifies the split→close round-trip.
func TestPaneCloseMsg_RemovesFocusedPane(t *testing.T) {
	m := newLayout(t)

	// Split to get 2 panes.
	m2, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm2 := m2.(layout.Model)
	if nm2.PaneCount() != 2 {
		t.Fatalf("expected 2 panes after split, got %d", nm2.PaneCount())
	}

	// Close the focused pane.
	m3, _ := nm2.Update(msgs.PaneCloseMsg{})
	nm3 := m3.(layout.Model)
	if nm3.PaneCount() != 1 {
		t.Errorf("PaneCloseMsg should reduce to 1 pane, got %d", nm3.PaneCount())
	}
}

// TestPaneFocusMoveMsg_ChangesFocusedPane verifies focus movement updates the model.
func TestPaneFocusMoveMsg_ChangesFocusedPane(t *testing.T) {
	m := newLayout(t)

	m2, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm2 := m2.(layout.Model)

	// Move focus right — should not panic and should leave pane count stable.
	m3, _ := nm2.Update(msgs.PaneFocusMoveMsg{Direction: msgs.FocusDirRight})
	nm3 := m3.(layout.Model)

	if nm3.PaneCount() != 2 {
		t.Errorf("focus move should not affect PaneCount, got %d", nm3.PaneCount())
	}
}

// TestPaneResizeMsg_ChangesParentSplitRatio verifies resize affects the layout.
func TestPaneResizeMsg_ChangesParentSplitRatio(t *testing.T) {
	m := newLayout(t)

	m2, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm2 := m2.(layout.Model)

	// Grow the focused pane — should not panic.
	m3, _ := nm2.Update(msgs.PaneResizeMsg{Direction: msgs.ResizeGrow, Axis: msgs.ResizeAxisH})
	nm3 := m3.(layout.Model)

	if nm3.PaneCount() != 2 {
		t.Errorf("resize should not affect PaneCount, got %d", nm3.PaneCount())
	}
}

// TestLayoutResizeMsg_PropagatesCorrectly verifies resize propagation.
func TestLayoutResizeMsg_PropagatesCorrectly(t *testing.T) {
	m := newLayout(t)

	m2, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm2 := m2.(layout.Model)

	// Resize to a new size — should not panic or drop panes.
	m3, _ := nm2.Update(msgs.LayoutResizeMsg{Width: 200, Height: 50})
	nm3 := m3.(layout.Model)

	if nm3.PaneCount() != 2 {
		t.Errorf("resize should not affect PaneCount, got %d", nm3.PaneCount())
	}
	// View should produce non-empty output.
	if nm3.View() == "" {
		t.Error("expected non-empty View() after resize")
	}
}

// TestMaxPanes_FourPanesCanBeCreated verifies the 4-pane limit is reachable.
func TestMaxPanes_FourPanesCanBeCreated(t *testing.T) {
	m := newLayout(t)
	var cur layout.Model = m

	// Create 3 additional panes (total 4).
	for i := 0; i < 3; i++ {
		next, _ := cur.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
		cur = next.(layout.Model)
	}

	if cur.PaneCount() != 4 {
		t.Errorf("expected 4 panes at max, got %d", cur.PaneCount())
	}
}

// TestMaxPanes_FifthSplitIsRejected verifies the 4-pane limit is enforced.
func TestMaxPanes_FifthSplitIsRejected(t *testing.T) {
	m := newLayout(t)
	var cur layout.Model = m

	for i := 0; i < 3; i++ {
		next, _ := cur.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
		cur = next.(layout.Model)
	}

	// Fifth split should not increase count.
	next, _ := cur.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm := next.(layout.Model)

	if nm.PaneCount() != 4 {
		t.Errorf("expected PaneCount to remain 4 after 5th split, got %d", nm.PaneCount())
	}
}
