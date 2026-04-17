// Package layout manages a binary split tree of panes (terminals and editors).
// It implements tea.Model and is owned by app.Model, replacing the former
// direct fields term/ed. Inspired by the Hyprland window manager's tiling model.
package layout

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/menegas/lumina/components/editor"
	"github.com/menegas/lumina/components/terminal"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

const (
	maxPanes  = 4
	minPaneW  = 20
	minPaneH  = 5
	ratioStep = 0.05
)

// ── Small interfaces to avoid type assertions on concrete types ──────────────

// paneIDer is implemented by terminal.Model (value receiver).
type paneIDer interface{ PaneID() int }

// resourceCloser is implemented by terminal.Model (value receiver).
type resourceCloser interface{ CloseResources() }

// dirtyChecker is implemented by editor.Model (value receiver).
type dirtyChecker interface{ Dirty() bool }

// ── leafModelAdapter ─────────────────────────────────────────────────────────

// leafModelAdapter wraps a tea.Model so LeafNode.Model satisfies View().
// The inner model is updated after each Update call to preserve Bubble Tea's
// immutable-model contract.
type leafModelAdapter struct {
	inner tea.Model
}

func (a *leafModelAdapter) View() string { return a.inner.View() }

// ── Model ────────────────────────────────────────────────────────────────────

// Model is the Bubble Tea model for the layout manager.
type Model struct {
	root         PaneNode
	focused      PaneID
	nextID       PaneID
	width        int
	height       int
	cfg          config.Config
	pendingClose bool // waiting for confirm-close of a dirty editor pane
}

// New creates a layout with a single terminal pane.
func New(cfg config.Config) (Model, error) {
	m := Model{
		nextID: 1,
		width:  80,
		height: 24,
		cfg:    cfg,
	}
	leaf, err := newTerminalLeaf(1, cfg)
	if err != nil {
		return m, fmt.Errorf("layout.New: %w", err)
	}
	m.root = leaf
	m.focused = leaf.ID
	return m, nil
}

// newTerminalLeaf creates a LeafNode backed by a new terminal.Model.
// paneID is set before Init() to ensure the PTY read goroutine captures it.
func newTerminalLeaf(id PaneID, cfg config.Config) (*LeafNode, error) {
	t, err := terminal.New(cfg)
	if err != nil {
		return nil, err
	}
	t.SetPaneID(int(id))
	return &LeafNode{
		ID:    id,
		Kind:  KindTerminal,
		Model: &leafModelAdapter{inner: t},
	}, nil
}

// newEditorLeaf creates a LeafNode backed by a new editor.Model.
func newEditorLeaf(id PaneID, cfg config.Config) *LeafNode {
	e := editor.New(cfg)
	return &LeafNode{
		ID:    id,
		Kind:  KindEditor,
		Model: &leafModelAdapter{inner: e},
	}
}

// leafInner returns the inner tea.Model stored in the leaf's adapter.
func leafInner(leaf *LeafNode) tea.Model {
	if a, ok := leaf.Model.(*leafModelAdapter); ok {
		return a.inner
	}
	return nil
}

// setLeafInner replaces the inner model in the leaf's adapter.
func setLeafInner(leaf *LeafNode, m tea.Model) {
	if a, ok := leaf.Model.(*leafModelAdapter); ok {
		a.inner = m
	}
}

// ── tea.Model ────────────────────────────────────────────────────────────────

// Init starts all pane Cmds.
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, leaf := range allLeaves(m.root) {
		if inner := leafInner(leaf); inner != nil {
			cmds = append(cmds, inner.Init())
		}
	}
	return tea.Batch(cmds...)
}

// Update routes messages to the appropriate pane(s) and handles layout operations.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case msgs.LayoutResizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, m.propagateResize()

	case msgs.PaneSplitMsg:
		return m.handleSplit(msg.Direction)

	case msgs.PaneCloseMsg:
		return m.handleClose()

	case msgs.PaneFocusMoveMsg:
		return m.handleFocusMove(msg.Direction)

	case msgs.PaneResizeMsg:
		return m.handlePaneResize(msg)

	// Pane close confirmation flow.
	case msgs.CloseConfirmedMsg:
		if m.pendingClose {
			m.pendingClose = false
			return m.doCloseLeaf(m.focused)
		}
		return m.updateFocused(msg)

	case msgs.CloseAbortedMsg:
		m.pendingClose = false
		return m, nil

	case msgs.PtyOutputMsg:
		return m.routePtyOutput(msg)

	case msgs.PtyInputMsg:
		// Input always routes to the focused terminal.
		return m.updateFocused(msg)

	case msgs.OpenFileMsg:
		return m.handleOpenFile(msg)

	case msgs.PaneFocusMsg:
		m.applyFocus(m.focused, msg.Focused)
		return m, nil

	// Keys and other pane-specific messages delegate to the focused pane.
	default:
		return m.updateFocused(msg)
	}
}

// View renders the pane tree.
func (m Model) View() string {
	return renderNode(m.root, m.width, m.height)
}

// ── Public accessors ─────────────────────────────────────────────────────────

// PaneCount returns the number of leaf panes currently open.
func (m Model) PaneCount() int { return countLeaves(m.root) }

// FocusedKind returns the content kind of the currently focused pane.
func (m Model) FocusedKind() PaneKind {
	if leaf := findLeaf(m.root, m.focused); leaf != nil {
		return leaf.Kind
	}
	return KindTerminal
}

// FocusedID returns the PaneID of the currently focused pane.
func (m Model) FocusedID() PaneID { return m.focused }

// SetContentFocused marks whether the layout's content area has app-level focus.
func (m Model) SetContentFocused(active bool) Model {
	m.applyFocus(m.focused, active)
	return m
}

// ── Internal helpers ─────────────────────────────────────────────────────────

// applyFocus sends PaneFocusMsg{Focused: true} to the focused leaf and
// PaneFocusMsg{Focused: false} to all others.
// Mutates leaves in-place through pointer chain (leaves are *LeafNode).
func (m Model) applyFocus(id PaneID, contentActive bool) {
	for _, leaf := range allLeaves(m.root) {
		focused := contentActive && leaf.ID == id
		if inner := leafInner(leaf); inner != nil {
			next, _ := inner.Update(msgs.PaneFocusMsg{Focused: focused})
			setLeafInner(leaf, next)
		}
	}
}

// propagateResize computes each leaf's dimensions and sends resize messages.
func (m Model) propagateResize() tea.Cmd {
	var cmds []tea.Cmd
	distributeSize(m.root, m.width, m.height, func(leaf *LeafNode, w, h int) {
		inner := leafInner(leaf)
		if inner == nil {
			return
		}
		var resizeMsg tea.Msg
		switch leaf.Kind {
		case KindTerminal:
			resizeMsg = msgs.TerminalResizeMsg{Width: w, Height: h}
		case KindEditor:
			resizeMsg = msgs.EditorResizeMsg{Width: w, Height: h}
		}
		next, cmd := inner.Update(resizeMsg)
		setLeafInner(leaf, next)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	})
	return tea.Batch(cmds...)
}

// distributeSize walks the tree, computing dimensions for each leaf, and calls cb.
func distributeSize(n PaneNode, w, h int, cb func(*LeafNode, int, int)) {
	switch v := n.(type) {
	case *LeafNode:
		cb(v, w, h)
	case *SplitNode:
		ratio := clampRatio(v.Ratio)
		switch v.Direction {
		case msgs.SplitHorizontal:
			firstW := max(1, int(float64(w)*ratio))
			secondW := max(1, w-firstW)
			distributeSize(v.First, firstW, h, cb)
			distributeSize(v.Second, secondW, h, cb)
		case msgs.SplitVertical:
			firstH := max(1, int(float64(h)*ratio))
			secondH := max(1, h-firstH)
			distributeSize(v.First, w, firstH, cb)
			distributeSize(v.Second, w, secondH, cb)
		}
	}
}

// updateFocused sends a message to the focused leaf's inner model.
func (m Model) updateFocused(msg tea.Msg) (tea.Model, tea.Cmd) {
	leaf := findLeaf(m.root, m.focused)
	if leaf == nil {
		return m, nil
	}
	inner := leafInner(leaf)
	if inner == nil {
		return m, nil
	}
	next, cmd := inner.Update(msg)
	setLeafInner(leaf, next)
	return m, cmd
}

// routePtyOutput routes PtyOutputMsg to the terminal leaf with matching PaneID.
func (m Model) routePtyOutput(msg msgs.PtyOutputMsg) (tea.Model, tea.Cmd) {
	for _, leaf := range allLeaves(m.root) {
		if leaf.Kind != KindTerminal {
			continue
		}
		inner := leafInner(leaf)
		if inner == nil {
			continue
		}
		if p, ok := inner.(paneIDer); ok && p.PaneID() == msg.PaneID {
			next, cmd := inner.Update(msg)
			setLeafInner(leaf, next)
			return m, cmd
		}
	}
	return m, nil
}

// handleSplit splits the focused pane in the given direction.
func (m Model) handleSplit(dir msgs.SplitDir) (tea.Model, tea.Cmd) {
	if m.PaneCount() >= maxPanes {
		return m, notifyStatus("Máximo de 4 painéis atingido", msgs.NotifyWarning)
	}
	if dir == msgs.SplitHorizontal && m.width/2 < minPaneW {
		return m, notifyStatus("Tela muito pequena para dividir horizontalmente", msgs.NotifyWarning)
	}
	if dir == msgs.SplitVertical && m.height/2 < minPaneH {
		return m, notifyStatus("Tela muito pequena para dividir verticalmente", msgs.NotifyWarning)
	}

	m.nextID++
	newLeaf, err := newTerminalLeaf(m.nextID, m.cfg)
	if err != nil {
		return m, notifyStatus("Erro ao criar terminal: "+err.Error(), msgs.NotifyError)
	}

	m.root = splitLeaf(m.root, m.focused, dir, newLeaf)

	initCmd := leafInner(newLeaf).Init()
	resizeCmd := m.propagateResize()
	m.focused = m.nextID          // focus moves to the new pane
	m.applyFocus(m.nextID, true)

	return m, tea.Batch(initCmd, resizeCmd)
}

// handleClose closes the focused pane, requesting confirmation if the editor is dirty.
func (m Model) handleClose() (tea.Model, tea.Cmd) {
	if m.PaneCount() <= 1 {
		return m, notifyStatus("Não é possível fechar o único painel", msgs.NotifyWarning)
	}
	leaf := findLeaf(m.root, m.focused)
	if leaf == nil {
		return m, nil
	}
	if leaf.Kind == KindEditor {
		if inner := leafInner(leaf); inner != nil {
			if dc, ok := inner.(dirtyChecker); ok && dc.Dirty() {
				m.pendingClose = true
				return m, func() tea.Msg { return msgs.ConfirmCloseMsg{} }
			}
		}
	}
	return m.doCloseLeaf(m.focused)
}

// doCloseLeaf removes the leaf with id from the tree and updates focus.
func (m Model) doCloseLeaf(id PaneID) (tea.Model, tea.Cmd) {
	result := closeLeaf(m.root, id)
	if result.removed == nil {
		return m, nil
	}
	// Release OS resources if the removed pane was a terminal.
	if result.removed.Kind == KindTerminal {
		if inner := leafInner(result.removed); inner != nil {
			if rc, ok := inner.(resourceCloser); ok {
				rc.CloseResources()
			}
		}
	}
	m.root = result.root
	// Move focus to the first remaining leaf.
	if remaining := allLeaves(m.root); len(remaining) > 0 {
		m.focused = remaining[0].ID
	}
	m.applyFocus(m.focused, true)
	return m, m.propagateResize()
}

// handleFocusMove moves focus in the given direction.
func (m Model) handleFocusMove(dir msgs.FocusDir) (tea.Model, tea.Cmd) {
	newID, ok := findNeighbour(m.root, m.focused, dir)
	if !ok {
		return m, nil
	}
	m.focused = newID
	m.applyFocus(newID, true)
	return m, nil
}

// handlePaneResize adjusts the split ratio of the focused pane's parent.
func (m Model) handlePaneResize(msg msgs.PaneResizeMsg) (tea.Model, tea.Cmd) {
	delta := ratioStep
	if msg.Direction == msgs.ResizeShrink {
		delta = -ratioStep
	}
	m.root = adjustRatio(m.root, m.focused, delta, msg.Axis)
	return m, m.propagateResize()
}

// handleOpenFile opens the file in the focused pane, converting a terminal to
// an editor if necessary.
func (m Model) handleOpenFile(msg msgs.OpenFileMsg) (tea.Model, tea.Cmd) {
	leaf := findLeaf(m.root, m.focused)
	if leaf == nil {
		return m, nil
	}
	if leaf.Kind == KindTerminal {
		if inner := leafInner(leaf); inner != nil {
			if rc, ok := inner.(resourceCloser); ok {
				rc.CloseResources()
			}
		}
		leaf.Kind = KindEditor
		setLeafInner(leaf, editor.New(m.cfg))
	}
	inner := leafInner(leaf)
	if inner == nil {
		return m, nil
	}
	next, cmd := inner.Update(msg)
	setLeafInner(leaf, next)
	m.applyFocus(m.focused, true)
	return m, cmd
}

// notifyStatus returns a Cmd that emits a StatusBarNotifyMsg.
func notifyStatus(text string, level msgs.NotifyLevel) tea.Cmd {
	return func() tea.Msg {
		return msgs.StatusBarNotifyMsg{
			Text:     text,
			Level:    level,
			Duration: 3 * time.Second,
		}
	}
}
