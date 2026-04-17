package terminal_test

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/menegas/lumina/components/terminal"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

func newTestModel(t *testing.T) terminal.Model {
	t.Helper()
	cfg := config.Config{Shell: "/bin/sh", SidebarWidth: 30}
	m, err := terminal.New(cfg)
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	return m
}

func TestUpdate_PtyOutputMsg_AppendsToBuffer(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	msg := msgs.PtyOutputMsg{Data: []byte("hello world\n"), Err: nil}
	next, _ := m.Update(msg)
	nm := next.(terminal.Model)

	if !strings.Contains(nm.View(), "hello world") {
		t.Errorf("expected View() to contain 'hello world', got: %q", nm.View())
	}
}

func TestUpdate_TerminalResizeMsg_UpdatesDimensions(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	next, _ := m.Update(msgs.TerminalResizeMsg{Width: 120, Height: 40})
	nm := next.(terminal.Model)

	w, h := nm.Dimensions()
	if w != 120 || h != 40 {
		t.Errorf("expected 120x40, got %dx%d", w, h)
	}
}

func TestUpdate_PtyOutputMsg_EOF_TriggersRestart(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	msg := msgs.PtyOutputMsg{Data: nil, Err: errors.New("EOF")}
	_, cmd := m.Update(msg)

	// A non-nil Cmd means the restart was initiated.
	if cmd == nil {
		t.Error("expected non-nil Cmd to restart shell after EOF, got nil")
	}
}

func TestView_ReturnsNonEmptyString(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	view := m.View()
	if view == "" {
		t.Error("expected non-empty View()")
	}
}

func TestUpdate_FocusedBorderChanges(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	m.SetFocused(true)
	focused := m.View()

	m.SetFocused(false)
	unfocused := m.View()

	if focused == unfocused {
		t.Error("expected View() to differ between focused and unfocused states")
	}
}

func TestModel_ImplementsTeaModel(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	// Verify the interface is satisfied at compile time via assignment.
	var _ tea.Model = m
}
