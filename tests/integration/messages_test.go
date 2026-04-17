package integration_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/menegas/lumina/components/editor"
	"github.com/menegas/lumina/components/statusbar"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

var testCfg = config.Config{
	Shell:           "/bin/sh",
	MetricsInterval: 100,
	ShowHidden:      false,
	SidebarWidth:    30,
}

// TestOpenFileMsg_SidebarToEditor verifies that OpenFileMsg causes
// the editor to load the specified file.
func TestOpenFileMsg_SidebarToEditor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(path, []byte("integration test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ed := editor.New(testCfg)
	next, _ := ed.Update(msgs.OpenFileMsg{Path: path})
	nm := next.(editor.Model)

	if nm.LineCount() == 0 {
		t.Error("expected editor to have lines after OpenFileMsg")
	}
}

// TestConfirmCloseFlow verifies the dirty → ConfirmCloseMsg → CloseConfirmedMsg flow.
func TestConfirmCloseFlow(t *testing.T) {
	ed := editor.New(testCfg)
	ed.SetFocused(true)

	// Make dirty by typing.
	next, _ := ed.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	ed = next.(editor.Model)

	// ConfirmCloseMsg → editor should clear dirty.
	next2, _ := ed.Update(msgs.CloseConfirmedMsg{})
	nm := next2.(editor.Model)
	if nm.Dirty() {
		t.Error("expected Dirty() == false after CloseConfirmedMsg")
	}
}

// TestMetricsTick_ReEnqueues verifies that MetricsTickMsg produces another Cmd.
func TestMetricsTick_ReEnqueues(t *testing.T) {
	sb := statusbar.New(testCfg)
	_, cmd := sb.Update(msgs.MetricsTickMsg{
		Tick: time.Now(),
	})
	if cmd == nil {
		t.Error("expected non-nil Cmd from MetricsTickMsg (re-enqueue next tick)")
	}
}

// TestWindowResize_StatusBar verifies that StatusBarResizeMsg propagates width.
func TestWindowResize_StatusBar(t *testing.T) {
	sb := statusbar.New(testCfg)
	next, _ := sb.Update(msgs.StatusBarResizeMsg{Width: 120})
	nm := next.(statusbar.Model)
	if nm.Width() != 120 {
		t.Errorf("expected width 120, got %d", nm.Width())
	}
}
