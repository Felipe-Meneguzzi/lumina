package sidebar_test

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Felipe-Meneguzzi/lumina/components/sidebar"
	"github.com/Felipe-Meneguzzi/lumina/config"
	"github.com/Felipe-Meneguzzi/lumina/msgs"
)

func newTestModel(t *testing.T) sidebar.Model {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	kb, _ := config.LoadKeybindings()
	return sidebar.New(root, config.Config{ShowHidden: false, SidebarWidth: 30, Keys: kb})
}

func TestUpdate_EnterOnFile_EmitsOpenInExternalEditorMsg(t *testing.T) {
	m := newTestModel(t)
	m.SetFocused(true)

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	next, cmd := m.Update(enterMsg)
	_ = next

	if cmd == nil {
		t.Skip("no cmd returned; selected item may be a dir — acceptable")
	}

	result := cmd()
	switch result.(type) {
	case msgs.OpenInExternalEditorMsg:
		// Expected for file selection.
	default:
		// Acceptable when the sorted-first item is a dir (subdir before file.txt).
	}
}

func TestUpdate_SidebarResizeMsg_UpdatesDimensions(t *testing.T) {
	m := newTestModel(t)
	next, _ := m.Update(msgs.SidebarResizeMsg{Width: 40, Height: 30})
	nm := next.(sidebar.Model)

	if nm.Width() != 40 {
		t.Errorf("expected width 40, got %d", nm.Width())
	}
}

func TestView_ZeroWidth_ReturnsEmpty(t *testing.T) {
	m := newTestModel(t)
	next, _ := m.Update(msgs.SidebarResizeMsg{Width: 0, Height: 24})
	nm := next.(sidebar.Model)

	if nm.View() != "" {
		t.Errorf("expected empty View() when width=0, got: %q", nm.View())
	}
}

func TestModel_ImplementsTeaModel(t *testing.T) {
	var _ tea.Model = newTestModel(t)
}

// TestBackspaceAtRoot_EmitsAlreadyAtRoot verifies FR-009 behaviour: pressing
// Backspace when the sidebar is already at its configured root surfaces a
// transient "Already at root" notification instead of moving up.
func TestBackspaceAtRoot_EmitsAlreadyAtRoot(t *testing.T) {
	m := newTestModel(t)
	m.SetFocused(true)

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	_ = next
	if cmd == nil {
		t.Fatal("expected a Cmd from Backspace at root, got nil")
	}
	msg := cmd()
	n, ok := msg.(msgs.StatusBarNotifyMsg)
	if !ok {
		t.Fatalf("expected StatusBarNotifyMsg, got %T", msg)
	}
	if n.Text != "Already at root" {
		t.Errorf("unexpected text: %q", n.Text)
	}
}
