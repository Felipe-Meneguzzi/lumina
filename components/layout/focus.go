package layout

import "github.com/menegas/lumina/msgs"

// findNeighbour returns the PaneID of the nearest leaf in the given direction
// relative to the leaf with currentID, or (0, false) if no neighbour exists.
//
// Algorithm (Hyprland-style):
//  1. Walk up the tree until we find a SplitNode whose direction matches the
//     requested axis AND the current leaf is on the "away" side.
//  2. Descend into the other child and return its nearest leaf.
//
// "away" side means:
//   - FocusDirLeft / FocusDirUp  → current leaf must be in Second (right/bottom)
//   - FocusDirRight / FocusDirDown → current leaf must be in First (left/top)
func findNeighbour(root PaneNode, currentID PaneID, dir msgs.FocusDir) (PaneID, bool) {
	// Collect the path from root to the target leaf.
	path := findPath(root, currentID)
	if len(path) == 0 {
		return 0, false
	}

	// Walk path backwards looking for a qualifying ancestor split.
	for i := len(path) - 1; i >= 0; i-- {
		split, ok := path[i].node.(*SplitNode)
		if !ok {
			continue
		}

		splitDir := split.Direction
		child := path[i].child // which child of this split contains our leaf

		switch dir {
		case msgs.FocusDirLeft:
			if splitDir == msgs.SplitHorizontal && child == childSecond {
				// Our leaf is in the right half — neighbour is in left half.
				return leafNearest(split.First, dir), true
			}
		case msgs.FocusDirRight:
			if splitDir == msgs.SplitHorizontal && child == childFirst {
				// Our leaf is in the left half — neighbour is in right half.
				return leafNearest(split.Second, dir), true
			}
		case msgs.FocusDirUp:
			if splitDir == msgs.SplitVertical && child == childSecond {
				// Our leaf is in the bottom half — neighbour is in top half.
				return leafNearest(split.First, dir), true
			}
		case msgs.FocusDirDown:
			if splitDir == msgs.SplitVertical && child == childFirst {
				// Our leaf is in the top half — neighbour is in bottom half.
				return leafNearest(split.Second, dir), true
			}
		}
	}
	return 0, false
}

// childSide records which child of a SplitNode is on the path to the target.
type childSide int

const (
	childFirst  childSide = iota
	childSecond childSide = iota
)

// pathEntry is one step along the root→leaf path.
type pathEntry struct {
	node  PaneNode
	child childSide // which child of node leads toward the target
}

// findPath returns the sequence of nodes from root down to (but not including)
// the target leaf, along with which child side was taken at each SplitNode.
func findPath(root PaneNode, targetID PaneID) []pathEntry {
	var result []pathEntry
	if walkPath(root, targetID, &result) {
		return result
	}
	return nil
}

// walkPath fills result with the path entries and returns true if target found.
func walkPath(n PaneNode, targetID PaneID, result *[]pathEntry) bool {
	switch v := n.(type) {
	case *LeafNode:
		return v.ID == targetID
	case *SplitNode:
		*result = append(*result, pathEntry{node: v, child: childFirst})
		if walkPath(v.First, targetID, result) {
			return true
		}
		(*result)[len(*result)-1].child = childSecond
		if walkPath(v.Second, targetID, result) {
			return true
		}
		*result = (*result)[:len(*result)-1]
	}
	return false
}

// leafNearest returns the PaneID of the leaf closest to the given direction
// inside the subtree n.
// For left/up: take the rightmost/bottommost leaf (closest to the boundary).
// For right/down: take the leftmost/topmost leaf.
func leafNearest(n PaneNode, dir msgs.FocusDir) PaneID {
	switch dir {
	case msgs.FocusDirLeft, msgs.FocusDirUp:
		// Moving toward this subtree from the right/bottom → pick the far-side leaf.
		l := lastLeaf(n)
		if l != nil {
			return l.ID
		}
	default:
		l := firstLeaf(n)
		if l != nil {
			return l.ID
		}
	}
	return 0
}
