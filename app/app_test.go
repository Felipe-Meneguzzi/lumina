package app_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/menegas/lumina/app"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

func newTestApp(t *testing.T) app.Model {
	t.Helper()
	kb, err := config.LoadKeybindings()
	if err != nil {
		t.Fatalf("LoadKeybindings: %v", err)
	}
	cfg := config.Config{Shell: "/bin/sh", Keys: kb}
	m, appErr := app.New(cfg)
	if appErr != nil {
		t.Fatalf("app.New: %v", appErr)
	}
	// Simulate a window size so sidebarWidth is initialised.
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return next.(app.Model)
}

func TestToggleSidebar_HideAndShow(t *testing.T) {
	m := newTestApp(t)

	// Sidebar should start visible (wide terminal).
	if m.SidebarWidth() == 0 {
		t.Fatal("expected sidebar visible after initial resize, got width 0")
	}

	// Toggle off.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e"), Alt: true})
	nm := next.(app.Model)
	if nm.SidebarWidth() != 0 {
		t.Errorf("after toggle off: expected sidebarWidth 0, got %d", nm.SidebarWidth())
	}

	// Toggle on.
	next2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e"), Alt: true})
	nm2 := next2.(app.Model)
	if nm2.SidebarWidth() == 0 {
		t.Error("after toggle on: expected sidebar visible (width > 0)")
	}
}

func TestToggleStatusBar_HideAndShow(t *testing.T) {
	m := newTestApp(t)

	// Status bar should start visible.
	if !m.SbarVisible() {
		t.Fatal("expected statusbar visible by default")
	}

	// Toggle off.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m"), Alt: true})
	nm := next.(app.Model)
	if nm.SbarVisible() {
		t.Error("after toggle off: expected statusbar hidden")
	}

	// Toggle on.
	next2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m"), Alt: true})
	nm2 := next2.(app.Model)
	if !nm2.SbarVisible() {
		t.Error("after toggle on: expected statusbar visible")
	}
}

func TestToggleSidebar_StatePerPane(t *testing.T) {
	m := newTestApp(t)

	// Split: creates pane 2, focus moves to pane 2 (new behaviour).
	next, _ := m.Update(msgs.PaneSplitMsg{Direction: msgs.SplitHorizontal})
	nm := next.(app.Model)

	// Toggle sidebar off on pane 2 (currently focused).
	nextOff, _ := nm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e"), Alt: true})
	nmOff := nextOff.(app.Model)
	if nmOff.SidebarWidth() != 0 {
		t.Errorf("pane 2: expected sidebar hidden, got width %d", nmOff.SidebarWidth())
	}

	// Move focus back to pane 1.
	nextPane1, _ := nmOff.Update(msgs.PaneFocusMoveMsg{Direction: msgs.FocusDirLeft})
	nmPane1 := nextPane1.(app.Model)

	// Pane 1 should have its sidebar visible (default).
	if nmPane1.SidebarWidth() == 0 {
		t.Error("pane 1: expected sidebar visible after switching from pane 2")
	}
}
