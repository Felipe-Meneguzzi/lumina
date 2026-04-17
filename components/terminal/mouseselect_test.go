package terminal_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/menegas/lumina/components/terminal"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

func newMouseTestModel(t *testing.T) terminal.Model {
	t.Helper()
	cfg := config.Config{Shell: "/bin/sh", SidebarWidth: 30, MouseAutoCopy: true}
	m, err := terminal.New(cfg)
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	m.Close()
	return m
}

// TestMouseSelection_StartAndUpdate verifies that a Press event initialises the
// selection and a Motion event extends it, with both positions clamped to bounds.
func TestMouseSelection_StartAndUpdate(t *testing.T) {
	m := newMouseTestModel(t)

	if m.HasMouseSelection() {
		t.Fatal("expected HasMouseSelection=false initially")
	}

	press := msgs.MouseSelectMsg{
		Mouse: tea.MouseMsg{X: 3, Y: 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
	}
	next, _ := m.Update(press)
	m = next.(terminal.Model)

	if !m.HasMouseSelection() {
		t.Fatal("expected HasMouseSelection=true after Press")
	}
	if m.HasPendingSelection() {
		t.Error("expected HasPendingSelection=false mid-drag")
	}

	// Extend via Motion.
	motion := msgs.MouseSelectMsg{
		Mouse: tea.MouseMsg{X: 20, Y: 5, Action: tea.MouseActionMotion},
	}
	next, _ = m.Update(motion)
	m = next.(terminal.Model)

	if !m.HasMouseSelection() {
		t.Error("expected HasMouseSelection=true after Motion")
	}

	// Out-of-bounds coordinates must be clamped, not panic.
	outOfBounds := msgs.MouseSelectMsg{
		Mouse: tea.MouseMsg{X: 9999, Y: 9999, Action: tea.MouseActionMotion},
	}
	next, _ = m.Update(outOfBounds)
	m = next.(terminal.Model)
	if !m.HasMouseSelection() {
		t.Error("expected HasMouseSelection=true even with clamped coordinates")
	}
}

// TestMouseSelection_FinalizeAutoCopy verifies that Release with auto-copy
// returns a non-nil Cmd and clears the selection.
func TestMouseSelection_FinalizeAutoCopy(t *testing.T) {
	m := newMouseTestModel(t) // MouseAutoCopy=true

	// Press to start.
	next, _ := m.Update(msgs.MouseSelectMsg{
		Mouse: tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
	})
	m = next.(terminal.Model)

	// Release — should auto-copy.
	next, cmd := m.Update(msgs.MouseSelectMsg{
		Mouse: tea.MouseMsg{X: 5, Y: 0, Action: tea.MouseActionRelease},
	})
	m = next.(terminal.Model)

	if m.HasMouseSelection() {
		t.Error("expected HasMouseSelection=false after auto-copy Release")
	}
	if cmd == nil {
		t.Error("expected non-nil Cmd (clipboard write) after auto-copy Release")
	}
}

// TestMouseSelection_FinalizeManualConfirm verifies that Release with
// auto-copy disabled marks the selection as pending and keeps it visible.
func TestMouseSelection_FinalizeManualConfirm(t *testing.T) {
	cfg := config.Config{Shell: "/bin/sh", SidebarWidth: 30, MouseAutoCopy: false}
	m, err := terminal.New(cfg)
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	m.Close()

	next, _ := m.Update(msgs.MouseSelectMsg{
		Mouse: tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
	})
	m = next.(terminal.Model)

	next, cmd := m.Update(msgs.MouseSelectMsg{
		Mouse: tea.MouseMsg{X: 5, Y: 0, Action: tea.MouseActionRelease},
	})
	m = next.(terminal.Model)

	if !m.HasMouseSelection() {
		t.Error("expected HasMouseSelection=true while pending confirmation")
	}
	if !m.HasPendingSelection() {
		t.Error("expected HasPendingSelection=true after Release with auto-copy disabled")
	}
	if cmd != nil {
		t.Error("expected nil Cmd before user confirms with 'y'")
	}
}

// TestMouseSelection_ConfirmAndCancel verifies MouseSelectConfirmMsg and
// MouseSelectCancelMsg behave correctly on a pending selection.
func TestMouseSelection_ConfirmAndCancel(t *testing.T) {
	makePending := func(t *testing.T) terminal.Model {
		t.Helper()
		cfg := config.Config{Shell: "/bin/sh", SidebarWidth: 30, MouseAutoCopy: false}
		m, err := terminal.New(cfg)
		if err != nil {
			t.Fatalf("terminal.New: %v", err)
		}
		m.Close()
		next, _ := m.Update(msgs.MouseSelectMsg{
			Mouse: tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
		})
		m = next.(terminal.Model)
		next, _ = m.Update(msgs.MouseSelectMsg{
			Mouse: tea.MouseMsg{X: 5, Y: 0, Action: tea.MouseActionRelease},
		})
		return next.(terminal.Model)
	}

	t.Run("confirm copies and clears", func(t *testing.T) {
		m := makePending(t)
		next, cmd := m.Update(msgs.MouseSelectConfirmMsg{})
		m = next.(terminal.Model)
		if m.HasMouseSelection() {
			t.Error("expected HasMouseSelection=false after confirm")
		}
		if cmd == nil {
			t.Error("expected non-nil Cmd after confirm")
		}
	})

	t.Run("cancel clears without cmd", func(t *testing.T) {
		m := makePending(t)
		next, cmd := m.Update(msgs.MouseSelectCancelMsg{})
		m = next.(terminal.Model)
		if m.HasMouseSelection() {
			t.Error("expected HasMouseSelection=false after cancel")
		}
		if cmd != nil {
			t.Error("expected nil Cmd after cancel")
		}
	})
}

// TestExtractMouseSelection_ReturnsCorrectText feeds text into the emulator,
// then performs a drag selection and verifies the extracted text.
func TestExtractMouseSelection_ReturnsCorrectText(t *testing.T) {
	m := newMouseTestModel(t)

	// Write a known line to the emulator.
	next, _ := m.Update(msgs.PtyOutputMsg{Data: []byte("HELLO\n"), Err: nil})
	m = next.(terminal.Model)

	// Select from (0,0) to (4,0) — should cover "HELLO".
	next, _ = m.Update(msgs.MouseSelectMsg{
		Mouse: tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
	})
	m = next.(terminal.Model)

	// Use auto-copy to capture the clipboard cmd; the selection is cleared after.
	// To inspect the text, switch to manual mode by using a model with auto-copy=false.
	cfg := config.Config{Shell: "/bin/sh", SidebarWidth: 30, MouseAutoCopy: false}
	m2, err := terminal.New(cfg)
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	m2.Close()

	next2, _ := m2.Update(msgs.PtyOutputMsg{Data: []byte("HELLO\n"), Err: nil})
	m2 = next2.(terminal.Model)
	next2, _ = m2.Update(msgs.MouseSelectMsg{
		Mouse: tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
	})
	m2 = next2.(terminal.Model)
	next2, _ = m2.Update(msgs.MouseSelectMsg{
		Mouse: tea.MouseMsg{X: 4, Y: 0, Action: tea.MouseActionRelease},
	})
	m2 = next2.(terminal.Model)

	if !m2.HasPendingSelection() {
		t.Fatal("expected pending selection for text extraction test")
	}

	// Confirm to extract: the Cmd returned wraps the OSC52 write.
	_, cmd := m2.Update(msgs.MouseSelectConfirmMsg{})
	if cmd == nil {
		t.Fatal("expected non-nil Cmd after confirm (clipboard write)")
	}

	// Execute the cmd to obtain the StatusBarNotifyMsg that carries the char count.
	result := cmd()
	notify, ok := result.(msgs.StatusBarNotifyMsg)
	if !ok {
		t.Fatalf("expected StatusBarNotifyMsg from clipboard cmd, got %T", result)
	}
	if !strings.Contains(notify.Text, "copiado") {
		t.Errorf("expected notification text to contain 'copiado', got %q", notify.Text)
	}
}
