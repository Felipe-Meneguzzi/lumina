package editor_test

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/menegas/lumina/components/editor"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

func newEditor() editor.Model {
	return editor.New(config.Config{})
}

func TestUpdate_TypingChar_SetsDirty(t *testing.T) {
	m := newEditor()
	m.SetFocused(true)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	nm := next.(editor.Model)

	if !nm.Dirty() {
		t.Error("expected Dirty() == true after typing")
	}
}

func TestUpdate_CtrlS_ClearsDirty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := editor.Open(path, config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	m.SetFocused(true)

	// Make dirty.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'!'}})
	nm := next.(editor.Model)

	// Save.
	next2, _ := nm.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	nm2 := next2.(editor.Model)

	if nm2.Dirty() {
		t.Error("expected Dirty() == false after Ctrl+S")
	}
}

func TestUpdate_CtrlW_WithDirty_EmitsConfirmClose(t *testing.T) {
	m := newEditor()
	m.SetFocused(true)

	// Type something to make dirty.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	nm := next.(editor.Model)

	_, cmd := nm.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	if cmd == nil {
		t.Fatal("expected non-nil Cmd from Ctrl+W with dirty state")
	}
	result := cmd()
	if _, ok := result.(msgs.ConfirmCloseMsg); !ok {
		t.Errorf("expected ConfirmCloseMsg, got %T", result)
	}
}

func TestUpdate_OpenFileMsg_LoadsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(path, []byte("line1\nline2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m := newEditor()
	next, _ := m.Update(msgs.OpenFileMsg{Path: path})
	nm := next.(editor.Model)

	if nm.LineCount() != 2 {
		t.Errorf("expected 2 lines, got %d", nm.LineCount())
	}
}

func TestModel_ImplementsTeaModel(t *testing.T) {
	var _ tea.Model = newEditor()
}
