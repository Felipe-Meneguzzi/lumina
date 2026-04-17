package integration_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/menegas/lumina/components/layout"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

func newMouseLayout(t *testing.T, autoCopy bool) layout.Model {
	t.Helper()
	cfg := config.Config{Shell: "/bin/sh", MouseAutoCopy: autoCopy}
	m, err := layout.New(cfg)
	if err != nil {
		t.Fatalf("layout.New: %v", err)
	}
	return m
}

// TestMouseSelectMsg_PressMotionRelease verifies the full drag flow:
// Press → Motion → Release generates a clipboard Cmd when auto-copy is on.
func TestMouseSelectMsg_PressMotionRelease(t *testing.T) {
	m := newMouseLayout(t, true) // auto-copy=true

	paneID := int(m.FocusedID())

	next, _ := m.Update(msgs.MouseSelectMsg{
		PaneID: paneID,
		Mouse:  tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
	})
	m = next.(layout.Model)

	if !m.FocusedHasMouseSelection() {
		t.Fatal("expected FocusedHasMouseSelection=true after Press")
	}

	next, _ = m.Update(msgs.MouseSelectMsg{
		PaneID: paneID,
		Mouse:  tea.MouseMsg{X: 5, Y: 0, Action: tea.MouseActionMotion},
	})
	m = next.(layout.Model)

	next, cmd := m.Update(msgs.MouseSelectMsg{
		PaneID: paneID,
		Mouse:  tea.MouseMsg{X: 5, Y: 0, Action: tea.MouseActionRelease},
	})
	m = next.(layout.Model)

	if m.FocusedHasMouseSelection() {
		t.Error("expected FocusedHasMouseSelection=false after auto-copy Release")
	}
	if cmd == nil {
		t.Error("expected non-nil Cmd (clipboard) after Release with auto-copy")
	}
}

// TestMouseSelectConfirmMsg_CopiesAndClears verifies MouseSelectConfirmMsg
// triggers a clipboard cmd and clears the pending selection.
func TestMouseSelectConfirmMsg_CopiesAndClears(t *testing.T) {
	m := newMouseLayout(t, false) // auto-copy=false

	paneID := int(m.FocusedID())

	// Feed content so there is something to select.
	next, _ := m.Update(msgs.PtyOutputMsg{PaneID: paneID, Data: []byte("HELLO\n"), Err: nil})
	m = next.(layout.Model)

	// Build a pending selection over "HELLO".
	next, _ = m.Update(msgs.MouseSelectMsg{
		PaneID: paneID,
		Mouse:  tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
	})
	m = next.(layout.Model)
	next, _ = m.Update(msgs.MouseSelectMsg{
		PaneID: paneID,
		Mouse:  tea.MouseMsg{X: 4, Y: 0, Action: tea.MouseActionRelease},
	})
	m = next.(layout.Model)

	if !m.FocusedHasPendingSelection() {
		t.Fatal("expected FocusedHasPendingSelection=true before confirm")
	}

	next, cmd := m.Update(msgs.MouseSelectConfirmMsg{PaneID: paneID})
	m = next.(layout.Model)

	if m.FocusedHasMouseSelection() {
		t.Error("expected FocusedHasMouseSelection=false after confirm")
	}
	if cmd == nil {
		t.Error("expected non-nil Cmd after confirm")
	}
	// Execute the cmd to get the clipboard notification.
	result := cmd()
	notify, ok := result.(msgs.StatusBarNotifyMsg)
	if !ok {
		t.Fatalf("expected StatusBarNotifyMsg, got %T", result)
	}
	if !strings.Contains(notify.Text, "copiado") {
		t.Errorf("expected 'copiado' in notification, got %q", notify.Text)
	}
}

// TestMouseSelectCancelMsg_ClearsWithoutCopy verifies MouseSelectCancelMsg
// discards the selection without issuing a clipboard command.
func TestMouseSelectCancelMsg_ClearsWithoutCopy(t *testing.T) {
	m := newMouseLayout(t, false) // auto-copy=false

	paneID := int(m.FocusedID())

	next, _ := m.Update(msgs.MouseSelectMsg{
		PaneID: paneID,
		Mouse:  tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
	})
	m = next.(layout.Model)
	next, _ = m.Update(msgs.MouseSelectMsg{
		PaneID: paneID,
		Mouse:  tea.MouseMsg{X: 5, Y: 0, Action: tea.MouseActionRelease},
	})
	m = next.(layout.Model)

	if !m.FocusedHasPendingSelection() {
		t.Fatal("expected FocusedHasPendingSelection=true before cancel")
	}

	next, cmd := m.Update(msgs.MouseSelectCancelMsg{PaneID: paneID})
	m = next.(layout.Model)

	if m.FocusedHasMouseSelection() {
		t.Error("expected FocusedHasMouseSelection=false after cancel")
	}
	if cmd != nil {
		t.Error("expected nil Cmd after cancel (no clipboard change)")
	}
}

// TestMouseSelectMsg_ShiftBypassPTYTracking verifies that a MouseSelectMsg
// (already routed by app.handleMouse) updates the terminal selection even when
// the terminal's inner application has mouse tracking enabled.
func TestMouseSelectMsg_ShiftBypassPTYTracking(t *testing.T) {
	m := newMouseLayout(t, true)
	paneID := int(m.FocusedID())

	// Enable inner-app mouse tracking by feeding the DEC mode sequence.
	next, _ := m.Update(msgs.PtyOutputMsg{PaneID: paneID, Data: []byte("\x1b[?1000h"), Err: nil})
	m = next.(layout.Model)

	if !m.FocusedMouseEnabled() {
		t.Fatal("expected FocusedMouseEnabled=true after sending \\e[?1000h")
	}

	// MouseSelectMsg (Shift bypass already resolved by app.handleMouse)
	// must still activate the selection even with mouse tracking on.
	next, _ = m.Update(msgs.MouseSelectMsg{
		PaneID: paneID,
		Mouse:  tea.MouseMsg{X: 2, Y: 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, Shift: true},
	})
	m = next.(layout.Model)

	if !m.FocusedHasMouseSelection() {
		t.Error("expected FocusedHasMouseSelection=true: MouseSelectMsg must activate selection regardless of PTY tracking")
	}
}
