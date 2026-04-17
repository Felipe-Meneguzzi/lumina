package layout

import (
	"testing"

	"github.com/Felipe-Meneguzzi/lumina/msgs"
)

// buildThreePaneTree builds Split(H,r1){ First: Split(H,r2){ First: leaf(1), Second: leaf(2) }, Second: leaf(3) }
// representing [A | B | C] where leaf(2)=B is the middle pane.
func buildThreePaneTree(r1, r2 float64) (*SplitNode, *SplitNode) {
	inner := &SplitNode{
		Direction: msgs.SplitHorizontal,
		Ratio:     r2,
		First:     &LeafNode{ID: 1},
		Second:    &LeafNode{ID: 2},
	}
	outer := &SplitNode{
		Direction: msgs.SplitHorizontal,
		Ratio:     r1,
		First:     inner,
		Second:    &LeafNode{ID: 3},
	}
	return outer, inner
}

// TestBoundaryRight_MiddlePane_OnlyMovesRightBoundary verifies that pressing
// BoundaryRight on the middle pane (B) moves only the boundary between B and C,
// not the boundary between A and B.
func TestBoundaryRight_MiddlePane_OnlyMovesRightBoundary(t *testing.T) {
	outer, inner := buildThreePaneTree(0.5, 0.5)
	const delta = 0.1
	// B is leaf(2), focused. delta > 0 = BoundaryRight.
	adjustRatioAbsolute(outer, 2, delta, msgs.ResizeAxisH)

	if inner.Ratio != 0.5 {
		t.Errorf("inner ratio (A|B boundary) should be unchanged: got %.2f, want 0.50", inner.Ratio)
	}
	if outer.Ratio != clampRatio(0.5+delta) {
		t.Errorf("outer ratio (B|C boundary) should increase: got %.2f, want %.2f", outer.Ratio, clampRatio(0.5+delta))
	}
}

// TestBoundaryLeft_MiddlePane_OnlyMovesLeftBoundary verifies that pressing
// BoundaryLeft on the middle pane (B) moves only the boundary between A and B,
// not the boundary between B and C.
func TestBoundaryLeft_MiddlePane_OnlyMovesLeftBoundary(t *testing.T) {
	outer, inner := buildThreePaneTree(0.5, 0.5)
	const delta = -0.1
	// B is leaf(2), focused. delta < 0 = BoundaryLeft.
	adjustRatioAbsolute(outer, 2, delta, msgs.ResizeAxisH)

	if inner.Ratio != clampRatio(0.5+delta) {
		t.Errorf("inner ratio (A|B boundary) should decrease: got %.2f, want %.2f", inner.Ratio, clampRatio(0.5+delta))
	}
	if outer.Ratio != 0.5 {
		t.Errorf("outer ratio (B|C boundary) should be unchanged: got %.2f, want 0.50", outer.Ratio)
	}
}

// TestBoundaryRight_LeftPane_MovesItsOnlyBoundary verifies that the leftmost
// pane (A) can still grow its right boundary.
func TestBoundaryRight_LeftPane_MovesItsOnlyBoundary(t *testing.T) {
	outer, inner := buildThreePaneTree(0.5, 0.5)
	// A is leaf(1). BoundaryRight should grow A (inner ratio increases).
	adjustRatioAbsolute(outer, 1, 0.1, msgs.ResizeAxisH)

	if inner.Ratio != clampRatio(0.5+0.1) {
		t.Errorf("inner ratio should increase: got %.2f, want %.2f", inner.Ratio, clampRatio(0.6))
	}
	if outer.Ratio != 0.5 {
		t.Errorf("outer ratio should be unchanged: got %.2f, want 0.50", outer.Ratio)
	}
}

// TestBoundaryLeft_RightPane_MovesItsOnlyBoundary verifies that the rightmost
// pane (C) can grow by moving the boundary between B and C leftward.
func TestBoundaryLeft_RightPane_MovesItsOnlyBoundary(t *testing.T) {
	outer, inner := buildThreePaneTree(0.5, 0.5)
	// C is leaf(3). BoundaryLeft should grow C (outer ratio decreases).
	adjustRatioAbsolute(outer, 3, -0.1, msgs.ResizeAxisH)

	if outer.Ratio != clampRatio(0.5-0.1) {
		t.Errorf("outer ratio should decrease: got %.2f, want %.2f", outer.Ratio, clampRatio(0.4))
	}
	if inner.Ratio != 0.5 {
		t.Errorf("inner ratio should be unchanged: got %.2f, want 0.50", inner.Ratio)
	}
}

// TestBoundaryRight_RightPane_FallsBackToLeftBoundary verifies that when the
// rightmost pane has no right boundary, pressing BoundaryRight still has an
// effect: the boundary on its left moves right (pane shrinks from the left).
func TestBoundaryRight_RightPane_FallsBackToLeftBoundary(t *testing.T) {
	outer, inner := buildThreePaneTree(0.5, 0.5)
	// C is leaf(3), rightmost. BoundaryRight with no right boundary should fall back
	// to moving the outer boundary (B|C) right → outer ratio increases, C shrinks.
	adjustRatioAbsolute(outer, 3, 0.1, msgs.ResizeAxisH)

	if outer.Ratio != clampRatio(0.5+0.1) {
		t.Errorf("outer ratio should increase (fallback): got %.2f, want %.2f", outer.Ratio, clampRatio(0.6))
	}
	if inner.Ratio != 0.5 {
		t.Errorf("inner ratio should be unchanged: got %.2f, want 0.50", inner.Ratio)
	}
}

// TestBoundaryLeft_LeftPane_FallsBackToRightBoundary verifies that when the
// leftmost pane has no left boundary, pressing BoundaryLeft still has an effect:
// the boundary on its right moves left (pane shrinks from the right).
func TestBoundaryLeft_LeftPane_FallsBackToRightBoundary(t *testing.T) {
	outer, inner := buildThreePaneTree(0.5, 0.5)
	// A is leaf(1), leftmost. BoundaryLeft with no left boundary should fall back
	// to moving the inner boundary (A|B) left → inner ratio decreases, A shrinks.
	adjustRatioAbsolute(outer, 1, -0.1, msgs.ResizeAxisH)

	if inner.Ratio != clampRatio(0.5-0.1) {
		t.Errorf("inner ratio should decrease (fallback): got %.2f, want %.2f", inner.Ratio, clampRatio(0.4))
	}
	if outer.Ratio != 0.5 {
		t.Errorf("outer ratio should be unchanged: got %.2f, want 0.50", outer.Ratio)
	}
}
