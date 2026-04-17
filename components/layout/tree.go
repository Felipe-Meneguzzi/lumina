package layout

import "github.com/menegas/lumina/msgs"

// PaneID uniquely identifies a LeafNode in the pane tree.
type PaneID int

// PaneKind identifies the type of content in a leaf pane.
type PaneKind int

const (
	KindTerminal PaneKind = iota
	KindEditor
)

// PaneNode is the interface implemented by both SplitNode and LeafNode.
type PaneNode interface {
	isPaneNode()
}

// SplitNode is an internal tree node that divides its space between two children.
type SplitNode struct {
	Direction msgs.SplitDir
	Ratio     float64 // fraction of space given to First (clamped to [0.1, 0.9])
	First     PaneNode
	Second    PaneNode
}

func (s *SplitNode) isPaneNode() {}

// LeafNode is a leaf that holds a real pane model (terminal or editor).
type LeafNode struct {
	ID   PaneID
	Kind PaneKind
	// model is the tea.Model for this pane (terminal.Model or editor.Model).
	// Stored as a tea.Model so layout has no import cycle with concrete types.
	Model interface{ View() string }
}

func (l *LeafNode) isPaneNode() {}

// countLeaves returns the number of LeafNodes in the subtree.
func countLeaves(n PaneNode) int {
	switch v := n.(type) {
	case *LeafNode:
		return 1
	case *SplitNode:
		return countLeaves(v.First) + countLeaves(v.Second)
	}
	return 0
}

// findLeaf returns the LeafNode with the given ID, or nil if not found.
func findLeaf(n PaneNode, id PaneID) *LeafNode {
	switch v := n.(type) {
	case *LeafNode:
		if v.ID == id {
			return v
		}
	case *SplitNode:
		if l := findLeaf(v.First, id); l != nil {
			return l
		}
		return findLeaf(v.Second, id)
	}
	return nil
}

// firstLeaf returns the first (leftmost/topmost) LeafNode in the subtree.
func firstLeaf(n PaneNode) *LeafNode {
	switch v := n.(type) {
	case *LeafNode:
		return v
	case *SplitNode:
		return firstLeaf(v.First)
	}
	return nil
}

// lastLeaf returns the last (rightmost/bottommost) LeafNode in the subtree.
func lastLeaf(n PaneNode) *LeafNode {
	switch v := n.(type) {
	case *LeafNode:
		return v
	case *SplitNode:
		return lastLeaf(v.Second)
	}
	return nil
}

// allLeaves returns all LeafNodes in the subtree in left-to-right, top-to-bottom order.
func allLeaves(n PaneNode) []*LeafNode {
	switch v := n.(type) {
	case *LeafNode:
		return []*LeafNode{v}
	case *SplitNode:
		return append(allLeaves(v.First), allLeaves(v.Second)...)
	}
	return nil
}

// clampRatio clamps ratio to [0.1, 0.9].
func clampRatio(r float64) float64 {
	if r < 0.1 {
		return 0.1
	}
	if r > 0.9 {
		return 0.9
	}
	return r
}

// splitLeaf replaces the leaf with the given targetID with a SplitNode containing
// the original leaf and a new sibling leaf. Returns the (possibly new) root.
func splitLeaf(root PaneNode, targetID PaneID, dir msgs.SplitDir, newLeaf *LeafNode) PaneNode {
	switch v := root.(type) {
	case *LeafNode:
		if v.ID == targetID {
			return &SplitNode{
				Direction: dir,
				Ratio:     0.5,
				First:     v,
				Second:    newLeaf,
			}
		}
		return root
	case *SplitNode:
		v.First = splitLeaf(v.First, targetID, dir, newLeaf)
		v.Second = splitLeaf(v.Second, targetID, dir, newLeaf)
		return v
	}
	return root
}

// closeResult holds the result of a closeLeaf operation.
type closeResult struct {
	root    PaneNode
	removed *LeafNode // the leaf that was removed (nil if not found or single pane)
}

// closeLeaf removes the leaf with the given targetID from the tree.
// If the tree contains only one leaf, returns unchanged root and nil removed.
// The parent SplitNode is replaced by the surviving sibling.
func closeLeaf(root PaneNode, targetID PaneID) closeResult {
	if countLeaves(root) <= 1 {
		return closeResult{root: root}
	}
	newRoot, removed := closeLeafInner(root, targetID)
	if newRoot == nil {
		newRoot = root
	}
	return closeResult{root: newRoot, removed: removed}
}

// closeLeafInner is the recursive implementation.
// Returns (newNode, removedLeaf). newNode may be nil if the caller should replace
// the parent reference with the sibling.
func closeLeafInner(n PaneNode, targetID PaneID) (PaneNode, *LeafNode) {
	switch v := n.(type) {
	case *LeafNode:
		if v.ID == targetID {
			return nil, v // signal: replace me with sibling
		}
		return n, nil
	case *SplitNode:
		newFirst, removed := closeLeafInner(v.First, targetID)
		if removed != nil {
			if newFirst == nil {
				// First child was removed — replace split with Second child.
				return v.Second, removed
			}
			v.First = newFirst
			return v, removed
		}
		newSecond, removed := closeLeafInner(v.Second, targetID)
		if removed != nil {
			if newSecond == nil {
				// Second child was removed — replace split with First child.
				return v.First, removed
			}
			v.Second = newSecond
			return v, removed
		}
	}
	return n, nil
}

// adjustRatio adjusts the Ratio of the SplitNode that is the parent of the leaf
// with the given targetID and whose direction matches the requested axis.
// delta is positive to grow First child (and shrink Second), negative to shrink First.
func adjustRatio(root PaneNode, targetID PaneID, delta float64, axis msgs.ResizeAxis) PaneNode {
	adjustRatioInner(root, targetID, delta, axis)
	return root
}

// adjustRatioInner walks the tree to find the right SplitNode and mutates its Ratio.
// Returns true if the target was found in this subtree.
func adjustRatioInner(n PaneNode, targetID PaneID, delta float64, axis msgs.ResizeAxis) bool {
	switch v := n.(type) {
	case *LeafNode:
		return v.ID == targetID
	case *SplitNode:
		firstHas := adjustRatioInner(v.First, targetID, delta, axis)
		secondHas := adjustRatioInner(v.Second, targetID, delta, axis)

		if firstHas || secondHas {
			// Check if this split's direction matches the requested axis.
			splitMatchesAxis :=
				(v.Direction == msgs.SplitHorizontal && axis == msgs.ResizeAxisH) ||
					(v.Direction == msgs.SplitVertical && axis == msgs.ResizeAxisV)

			if splitMatchesAxis {
				// Grow First (positive delta) or shrink First (negative delta).
				if firstHas {
					v.Ratio = clampRatio(v.Ratio + delta)
				} else {
					v.Ratio = clampRatio(v.Ratio - delta)
				}
			}
			return true
		}
	}
	return false
}
