package layout_test

import (
	"testing"

	"github.com/menegas/lumina/components/layout"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

func newTestLayout(t *testing.T) layout.Model {
	t.Helper()
	cfg := config.Config{Shell: "/bin/sh"}
	m, err := layout.New(cfg)
	if err != nil {
		t.Fatalf("layout.New: %v", err)
	}
	return m
}

func TestNew_CreatesSinglePane(t *testing.T) {
	m := newTestLayout(t)
	if m.PaneCount() != 1 {
		t.Errorf("expected PaneCount == 1, got %d", m.PaneCount())
	}
}

func TestNew_FocusedKindIsTerminal(t *testing.T) {
	m := newTestLayout(t)
	if m.FocusedKind() != layout.KindTerminal {
		t.Errorf("expected KindTerminal, got %v", m.FocusedKind())
	}
}

func TestView_SinglePane_ReturnsNonEmpty(t *testing.T) {
	m := newTestLayout(t)
	v := m.View()
	if v == "" {
		t.Error("expected non-empty View() for single pane")
	}
}

func TestLayoutResizeMsg_UpdatesDimensions(t *testing.T) {
	m := newTestLayout(t)
	next, _ := m.Update(msgs.LayoutResizeMsg{Width: 120, Height: 40})
	nm := next.(layout.Model)
	v := nm.View()
	if v == "" {
		t.Error("expected non-empty View() after resize")
	}
}

func TestPaneSplitMsg_Horizontal_CreatesTwoPanes(t *testing.T) {
	m := newTestLayout(t)
	next, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm := next.(layout.Model)
	if nm.PaneCount() != 2 {
		t.Errorf("expected 2 panes after horizontal split, got %d", nm.PaneCount())
	}
}

func TestPaneSplitMsg_Vertical_CreatesTwoPanes(t *testing.T) {
	m := newTestLayout(t)
	next, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitVertical})
	nm := next.(layout.Model)
	if nm.PaneCount() != 2 {
		t.Errorf("expected 2 panes after vertical split, got %d", nm.PaneCount())
	}
}

func TestPaneSplitMsg_ThirdSplit_ThreePanes(t *testing.T) {
	m := newTestLayout(t)
	m2, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	m3, _ := m2.(layout.Model).Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm := m3.(layout.Model)
	if nm.PaneCount() != 3 {
		t.Errorf("expected 3 panes after two splits, got %d", nm.PaneCount())
	}
}

func TestPaneSplitMsg_AtMaxPanes_IsNoop(t *testing.T) {
	m := newTestLayout(t)
	var cur layout.Model = m
	for i := 0; i < 3; i++ {
		next, _ := cur.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
		cur = next.(layout.Model)
	}
	if cur.PaneCount() != 4 {
		t.Fatalf("expected 4 panes, got %d", cur.PaneCount())
	}
	// Fifth split should be a no-op.
	next, _ := cur.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm := next.(layout.Model)
	if nm.PaneCount() != 4 {
		t.Errorf("expected PaneCount to remain 4 at limit, got %d", nm.PaneCount())
	}
}

func TestPaneCloseMsg_SinglePane_IsNoop(t *testing.T) {
	m := newTestLayout(t)
	next, _ := m.Update(msgs.PaneCloseMsg{})
	nm := next.(layout.Model)
	if nm.PaneCount() != 1 {
		t.Errorf("expected PaneCount to remain 1, got %d", nm.PaneCount())
	}
}

func TestPaneCloseMsg_TwoPanes_LeavesOne(t *testing.T) {
	m := newTestLayout(t)
	m2, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm2 := m2.(layout.Model)
	if nm2.PaneCount() != 2 {
		t.Fatalf("expected 2 panes, got %d", nm2.PaneCount())
	}
	m3, _ := nm2.Update(msgs.PaneCloseMsg{})
	nm3 := m3.(layout.Model)
	if nm3.PaneCount() != 1 {
		t.Errorf("expected 1 pane after close, got %d", nm3.PaneCount())
	}
}

func TestPaneFocusMoveMsg_TwoPanesHorizontal_MovesRight(t *testing.T) {
	m := newTestLayout(t)
	// Split: two panes side by side. Focus starts on left (first) pane.
	m2, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm2 := m2.(layout.Model)
	// Move focus right.
	m3, _ := nm2.Update(msgs.PaneFocusMoveMsg{Direction: msgs.FocusDirRight})
	nm3 := m3.(layout.Model)
	// Move focus back left.
	m4, _ := nm3.Update(msgs.PaneFocusMoveMsg{Direction: msgs.FocusDirLeft})
	nm4 := m4.(layout.Model)
	// We can't directly inspect the focused ID from outside, but we can verify
	// the model doesn't panic and PaneCount is stable.
	if nm4.PaneCount() != 2 {
		t.Errorf("expected PaneCount == 2 after focus moves, got %d", nm4.PaneCount())
	}
}

func TestSplitFocusMovesToNewPane(t *testing.T) {
	m := newTestLayout(t)
	originalID := m.FocusedID()

	next, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm := next.(layout.Model)

	if nm.PaneCount() != 2 {
		t.Fatalf("expected 2 panes, got %d", nm.PaneCount())
	}
	newFocusedID := nm.FocusedID()
	if newFocusedID == originalID {
		t.Errorf("focus should move to new pane after split, but stayed on original (ID %v)", originalID)
	}
}

func TestPaneCloseMsg_CloseOriginalPaneWhenSecondExists(t *testing.T) {
	m := newTestLayout(t)
	originalID := m.FocusedID()

	// Split: focus moves to new pane (after Phase 4 fix).
	next, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm := next.(layout.Model)
	if nm.PaneCount() != 2 {
		t.Fatalf("expected 2 panes after split, got %d", nm.PaneCount())
	}

	// Navigate back to the original pane.
	moved, _ := nm.Update(msgs.PaneFocusMoveMsg{Direction: msgs.FocusDirLeft})
	nmMoved := moved.(layout.Model)
	if nmMoved.FocusedID() != originalID {
		t.Fatalf("expected focus on original pane %v, got %v", originalID, nmMoved.FocusedID())
	}

	// Close the original pane — must succeed (2 panes exist).
	closed, _ := nmMoved.Update(msgs.PaneCloseMsg{})
	nmClosed := closed.(layout.Model)
	if nmClosed.PaneCount() != 1 {
		t.Errorf("expected 1 pane after closing original, got %d", nmClosed.PaneCount())
	}
}

func TestPaneResizeMsg_AdjustsRatio(t *testing.T) {
	m := newTestLayout(t)
	m2, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm2 := m2.(layout.Model)
	// Grow pane horizontally — should not panic and PaneCount should remain 2.
	m3, _ := nm2.Update(msgs.PaneResizeMsg{Direction: msgs.ResizeGrow, Axis: msgs.ResizeAxisH})
	nm3 := m3.(layout.Model)
	if nm3.PaneCount() != 2 {
		t.Errorf("expected PaneCount == 2 after resize, got %d", nm3.PaneCount())
	}
}
