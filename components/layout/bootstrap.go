package layout

// buildInitialTree constructs the initial PaneNode tree for a freshly
// constructed Model. It honors m.initialCount (>1 triggers pre-split) and
// m.startCommand (non-empty runs the command in each initial pane).
//
// Returns the root of the tree and the PaneID of the first pane (used as
// the initial focus).
func buildInitialTree(m *Model) (PaneNode, PaneID, error) {
	firstLeaf, err := m.newInitialLeaf(1)
	if err != nil {
		return nil, 0, err
	}
	m.nextID = 1

	if m.initialCount <= 1 {
		return firstLeaf, firstLeaf.ID, nil
	}

	root := PaneNode(firstLeaf)
	for i := 1; i < m.initialCount; i++ {
		m.nextID++
		leaf, err := m.newInitialLeaf(m.nextID)
		if err != nil {
			return nil, 0, err
		}
		// Split from the most recently added leaf: keeps every new pane the
		// same fraction of the layout and avoids ballooning one side.
		root = splitLeaf(root, m.nextID-1, m.initialDir, leaf)
	}
	return root, m.nextID, nil
}

// newInitialLeaf selects the right constructor based on whether a start
// command was configured.
func (m *Model) newInitialLeaf(id PaneID) (*LeafNode, error) {
	if m.startCommand == "" {
		return newTerminalLeaf(id, m.cfg)
	}
	return newTerminalLeafWithCommand(id, m.cfg, m.startCommand)
}
