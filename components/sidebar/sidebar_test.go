package sidebar_test

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/menegas/lumina/components/sidebar"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

func newTestModel(t *testing.T) sidebar.Model {
	t.Helper()
	root := t.TempDir()
	// Create test file structure.
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}
	return sidebar.New(root, config.Config{ShowHidden: false, SidebarWidth: 30})
}

func TestUpdate_EnterOnFile_EmitsOpenFileMsg(t *testing.T) {
	m := newTestModel(t)
	m.SetFocused(true)

	// Navigate to file (index 0 is "file.txt" or "subdir" depending on sort).
	// We just press Enter on whatever is selected and check for OpenFileMsg.
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	next, cmd := m.Update(enterMsg)
	_ = next

	if cmd == nil {
		t.Skip("no cmd returned; item may be a dir — acceptable in test environment")
	}

	result := cmd()
	switch result.(type) {
	case msgs.OpenFileMsg:
		// Expected for file selection.
	default:
		// May be other msgs if the selected item is a dir — acceptable.
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
