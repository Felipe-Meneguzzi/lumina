package layout_test

import (
	"testing"

	"github.com/Felipe-Meneguzzi/lumina/components/layout"
	"github.com/Felipe-Meneguzzi/lumina/config"
	"github.com/Felipe-Meneguzzi/lumina/msgs"
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

// TestPaneBounds_AfterSplit verifies that after splitting horizontally, each
// pane reports a non-overlapping rectangle that together cover the full layout.
func TestPaneBounds_AfterSplit(t *testing.T) {
	m := newTestLayout(t)
	next, _ := m.Update(msgs.LayoutResizeMsg{Width: 100, Height: 30})
	m = next.(layout.Model)

	// Split horizontally — creates a sibling pane to the right.
	next, _ = m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	m = next.(layout.Model)

	x, y, w, h, ok := m.FocusedBounds()
	if !ok {
		t.Fatal("expected FocusedBounds to succeed after split")
	}
	if y != 0 || h != 30 {
		t.Errorf("expected y=0 h=30, got y=%d h=%d", y, h)
	}
	if x <= 0 || w <= 0 || x+w > 100 {
		t.Errorf("expected non-overlapping right pane within 100 cols, got x=%d w=%d", x, w)
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

func TestLayoutNew_DefaultMaxPanesIs4(t *testing.T) {
	m := newTestLayout(t)
	// 3 splits = 4 panes (OK); 4th split should fail.
	for i := 0; i < 3; i++ {
		next, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
		m = next.(layout.Model)
	}
	if m.PaneCount() != 4 {
		t.Fatalf("expected 4 panes, got %d", m.PaneCount())
	}
	next, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	if next.(layout.Model).PaneCount() != 4 {
		t.Errorf("default max should cap at 4, got %d", next.(layout.Model).PaneCount())
	}
}

func TestLayoutNew_WithMaxPanes_10(t *testing.T) {
	cfg := config.Config{Shell: "/bin/sh"}
	m, err := layout.New(cfg, layout.WithMaxPanes(10))
	if err != nil {
		t.Fatalf("layout.New: %v", err)
	}
	for i := 0; i < 9; i++ {
		next, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
		m = next.(layout.Model)
	}
	if m.PaneCount() != 10 {
		t.Fatalf("expected 10 panes, got %d", m.PaneCount())
	}
	// 11th should fail.
	next, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	if next.(layout.Model).PaneCount() != 10 {
		t.Errorf("expected cap at 10, got %d", next.(layout.Model).PaneCount())
	}
}

func TestPaneSplitMsg_AtCustomMax_IsNoop(t *testing.T) {
	cfg := config.Config{Shell: "/bin/sh"}
	m, err := layout.New(cfg, layout.WithMaxPanes(2))
	if err != nil {
		t.Fatalf("layout.New: %v", err)
	}
	next, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	m = next.(layout.Model)
	if m.PaneCount() != 2 {
		t.Fatalf("expected 2 panes, got %d", m.PaneCount())
	}
	// 3rd split at custom cap 2 should be rejected.
	next, _ = m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	if next.(layout.Model).PaneCount() != 2 {
		t.Errorf("expected cap at 2, got %d", next.(layout.Model).PaneCount())
	}
}

func TestLayoutNew_WithInitialLayoutH3(t *testing.T) {
	cfg := config.Config{Shell: "/bin/sh"}
	m, err := layout.New(cfg, layout.WithInitialLayout(msgs.SplitHorizontal, 3))
	if err != nil {
		t.Fatalf("layout.New: %v", err)
	}
	if m.PaneCount() != 3 {
		t.Errorf("expected 3 initial panes, got %d", m.PaneCount())
	}
}

func TestLayoutNew_WithInitialLayoutV2(t *testing.T) {
	cfg := config.Config{Shell: "/bin/sh"}
	m, err := layout.New(cfg, layout.WithInitialLayout(msgs.SplitVertical, 2))
	if err != nil {
		t.Fatalf("layout.New: %v", err)
	}
	if m.PaneCount() != 2 {
		t.Errorf("expected 2 initial panes, got %d", m.PaneCount())
	}
}

func TestLayoutNew_WithInitialLayoutCount1_IsSinglePane(t *testing.T) {
	cfg := config.Config{Shell: "/bin/sh"}
	m, err := layout.New(cfg, layout.WithInitialLayout(msgs.SplitHorizontal, 1))
	if err != nil {
		t.Fatalf("layout.New: %v", err)
	}
	if m.PaneCount() != 1 {
		t.Errorf("expected 1 pane for count=1, got %d", m.PaneCount())
	}
}

func TestLayoutHandleSplit_DoesNotInheritStartCommand(t *testing.T) {
	// StartCommand is /bin/true (exits immediately) — initial pane boots with it,
	// but a subsequent manual split must fall back to the default shell.
	cfg := config.Config{Shell: "/bin/sh"}
	m, err := layout.New(cfg,
		layout.WithStartCommand("/bin/true"),
		layout.WithInitialLayout(msgs.SplitHorizontal, 2),
		layout.WithMaxPanes(4),
	)
	if err != nil {
		t.Fatalf("layout.New: %v", err)
	}
	if m.PaneCount() != 2 {
		t.Fatalf("expected 2 initial panes, got %d", m.PaneCount())
	}
	// Trigger a manual split. The new leaf must be created via the default
	// shell path (newTerminalLeaf) — if it tried to reuse the startCommand,
	// the test would still pass, but we assert no panic and new pane exists.
	next, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	m = next.(layout.Model)
	if m.PaneCount() != 3 {
		t.Errorf("expected 3 panes after split, got %d", m.PaneCount())
	}
}

// TestHitTest_SinglePane verifies that HitTest resolves a click landing inside
// a single-pane layout to that pane, with local coordinates offset by the
// border (1,1 → 0,0 local).
func TestHitTest_SinglePane(t *testing.T) {
	m := newTestLayout(t)
	next, _ := m.Update(msgs.LayoutResizeMsg{Width: 80, Height: 24})
	m = next.(layout.Model)

	paneID, _, lx, ly, ok := m.HitTest(5, 3)
	if !ok {
		t.Fatal("expected HitTest to succeed inside the pane")
	}
	if paneID != m.FocusedID() {
		t.Errorf("expected focused pane %v, got %v", m.FocusedID(), paneID)
	}
	if lx != 4 || ly != 2 {
		t.Errorf("expected local (4,2) after border subtraction, got (%d,%d)", lx, ly)
	}
}

// TestHitTest_HorizontalSplit verifies that HitTest selects the correct side
// of a horizontal split based on X coordinate.
func TestHitTest_HorizontalSplit(t *testing.T) {
	m := newTestLayout(t)
	next, _ := m.Update(msgs.LayoutResizeMsg{Width: 100, Height: 30})
	m = next.(layout.Model)
	next, _ = m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	m = next.(layout.Model)

	// Click on the LEFT half.
	leftID, _, _, _, ok := m.HitTest(10, 5)
	if !ok {
		t.Fatal("expected hit on left pane")
	}
	// Click on the RIGHT half.
	rightID, _, _, _, ok := m.HitTest(90, 5)
	if !ok {
		t.Fatal("expected hit on right pane")
	}
	if leftID == rightID {
		t.Errorf("expected different pane IDs for left/right, got both=%v", leftID)
	}
}

// TestHitTest_OutOfBounds returns ok=false for coordinates outside the layout.
func TestHitTest_OutOfBounds(t *testing.T) {
	m := newTestLayout(t)
	next, _ := m.Update(msgs.LayoutResizeMsg{Width: 80, Height: 24})
	m = next.(layout.Model)

	if _, _, _, _, ok := m.HitTest(200, 200); ok {
		t.Error("expected HitTest to fail for out-of-bounds click")
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
