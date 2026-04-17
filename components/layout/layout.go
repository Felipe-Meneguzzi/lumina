// Package layout manages a binary split tree of panes.
// It implements tea.Model and is owned by app.Model, routing messages to
// the correct pane and handling split / close / focus-move operations.
package layout

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Felipe-Meneguzzi/lumina/components/terminal"
	"github.com/Felipe-Meneguzzi/lumina/config"
	"github.com/Felipe-Meneguzzi/lumina/msgs"
)

const (
	defaultMaxPanes = 4
	minPaneW        = 20
	minPaneH        = 5
	ratioStep       = 0.05
)

// ── Small interfaces to avoid type assertions on concrete types ──────────────

// paneIDer is implemented by terminal.Model (value receiver).
type paneIDer interface{ PaneID() int }

// resourceCloser is implemented by terminal.Model (value receiver).
type resourceCloser interface{ CloseResources() }

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
	maxPanes     int           // session ceiling; replaces former package const
	startCommand string        // applied only to initial panes built in New; never propagated to splits
	initialDir   msgs.SplitDir // set by WithInitialLayout; consumed by New then zeroed
	initialCount int           // set by WithInitialLayout; >1 triggers pre-split
}

// Option configures a Model at construction time.
type Option func(*Model)

// WithMaxPanes overrides the session's pane ceiling (default 4).
func WithMaxPanes(n int) Option {
	return func(m *Model) {
		if n > 0 {
			m.maxPanes = n
		}
	}
}

// WithStartCommand makes initial panes built by New run the given command
// instead of the default shell. Panes created later via split always use
// the default shell — this field is consumed only during initial tree build.
func WithStartCommand(cmd string) Option {
	return func(m *Model) { m.startCommand = cmd }
}

// WithInitialLayout requests that New pre-splits the root pane into `count`
// panes with the given direction. count <= 1 is treated as "single pane"
// (no-op). Only msgs.SplitHorizontal and msgs.SplitVertical are honored.
func WithInitialLayout(dir msgs.SplitDir, count int) Option {
	return func(m *Model) {
		if count > 1 {
			m.initialDir = dir
			m.initialCount = count
		}
	}
}

// New creates a layout. With no opts, produces a single terminal pane —
// preserving the historical single-pane boot.
func New(cfg config.Config, opts ...Option) (Model, error) {
	m := Model{
		nextID:   1,
		width:    80,
		height:   24,
		cfg:      cfg,
		maxPanes: defaultMaxPanes,
	}
	for _, opt := range opts {
		opt(&m)
	}

	root, lastID, err := buildInitialTree(&m)
	if err != nil {
		return m, fmt.Errorf("layout.New: %w", err)
	}
	m.root = root
	m.focused = lastID
	// Clear transient init fields so they cannot leak into runtime behaviour.
	m.initialDir = msgs.SplitHorizontal
	m.initialCount = 0
	return m, nil
}

// newTerminalLeaf creates a LeafNode backed by a new terminal.Model running the
// default shell.
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

// newTerminalLeafWithCommand is like newTerminalLeaf but runs the given command
// under `sh -c` instead of the default shell. When transient is true the pane
// emits PaneAutoCloseMsg on PTY EOF so the layout can remove it without the
// usual shell auto-restart.
func newTerminalLeafWithCommand(id PaneID, cfg config.Config, command string, transient bool) (*LeafNode, error) {
	t, err := terminal.NewWithCommand(cfg, command)
	if err != nil {
		return nil, err
	}
	t.SetPaneID(int(id))
	if transient {
		t.SetTransient(true)
	}
	return &LeafNode{
		ID:    id,
		Kind:  KindTerminal,
		Model: &leafModelAdapter{inner: t},
	}, nil
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
		return m.handleSplit(msg)

	case msgs.PaneCloseMsg:
		return m.handleClose()

	case msgs.PaneAutoCloseMsg:
		return m.doCloseLeaf(PaneID(msg.PaneID))

	case msgs.PaneFocusMoveMsg:
		return m.handleFocusMove(msg.Direction)

	case msgs.PaneResizeMsg:
		return m.handlePaneResize(msg)

	case msgs.PtyOutputMsg:
		return m.routePtyOutput(msg)

	case msgs.PtyInputMsg:
		return m.updateFocused(msg)

	case msgs.PtyMouseMsg:
		return m.routePtyMouse(msg)

	case msgs.MouseSelectMsg, msgs.MouseSelectConfirmMsg, msgs.MouseSelectCancelMsg:
		return m.updateFocused(msg)

	case msgs.EnterCopyModeMsg:
		return m.updateFocused(msg)

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

// AllPaneIDs returns the IDs of every leaf pane in left-to-right / top-to-bottom order.
func (m Model) AllPaneIDs() []PaneID {
	leaves := allLeaves(m.root)
	ids := make([]PaneID, 0, len(leaves))
	for _, leaf := range leaves {
		ids = append(ids, leaf.ID)
	}
	return ids
}

// SetContentFocused marks whether the layout's content area has app-level focus.
func (m Model) SetContentFocused(active bool) Model {
	m.applyFocus(m.focused, active)
	return m
}

// SetFocusedID updates the focused pane. Used by app.go after click-to-focus
// hit-tests resolve which pane should receive input.
func (m Model) SetFocusedID(id PaneID) Model {
	if findLeaf(m.root, id) == nil {
		return m
	}
	m.focused = id
	m.applyFocus(id, true)
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
		next, cmd := inner.Update(msgs.TerminalResizeMsg{Width: w, Height: h})
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

// routePtyMouse routes PtyMouseMsg to the terminal leaf with matching PaneID.
func (m Model) routePtyMouse(msg msgs.PtyMouseMsg) (tea.Model, tea.Cmd) {
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

// handleSplit splits the focused pane in the given direction, optionally
// running a specific command in the new pane (for external-editor spawns).
func (m Model) handleSplit(msg msgs.PaneSplitMsg) (tea.Model, tea.Cmd) {
	dir := msg.Direction
	if m.PaneCount() >= m.maxPanes {
		return m, notifyStatus(fmt.Sprintf("Máximo de %d painéis atingido", m.maxPanes), msgs.NotifyWarning)
	}
	if dir == msgs.SplitHorizontal && m.width/2 < minPaneW {
		return m, notifyStatus("Tela muito pequena para dividir horizontalmente", msgs.NotifyWarning)
	}
	if dir == msgs.SplitVertical && m.height/2 < minPaneH {
		return m, notifyStatus("Tela muito pequena para dividir verticalmente", msgs.NotifyWarning)
	}

	m.nextID++
	var newLeaf *LeafNode
	var err error
	if msg.Command != "" {
		newLeaf, err = newTerminalLeafWithCommand(m.nextID, m.cfg, msg.Command, msg.Transient)
	} else {
		newLeaf, err = newTerminalLeaf(m.nextID, m.cfg)
	}
	if err != nil {
		return m, notifyStatus("Erro ao criar terminal: "+err.Error(), msgs.NotifyError)
	}

	m.root = splitLeaf(m.root, m.focused, dir, newLeaf)

	initCmd := leafInner(newLeaf).Init()
	resizeCmd := m.propagateResize()
	m.focused = m.nextID // focus moves to the new pane
	m.applyFocus(m.nextID, true)

	return m, tea.Batch(initCmd, resizeCmd)
}

// handleClose closes the focused pane.
func (m Model) handleClose() (tea.Model, tea.Cmd) {
	if m.PaneCount() <= 1 {
		return m, notifyStatus("Não é possível fechar o único painel", msgs.NotifyWarning)
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
	if msg.Boundary {
		m.root = adjustRatioAbsolute(m.root, m.focused, delta, msg.Axis)
	} else {
		m.root = adjustRatio(m.root, m.focused, delta, msg.Axis)
	}
	return m, m.propagateResize()
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
