package sidebar

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/menegas/lumina/msgs"
)

func TestValidateName(t *testing.T) {
	cases := []struct {
		name string
		want bool // true == valid (empty error string)
	}{
		{"foo", true},
		{".gitignore", true},
		{"", false},
		{".", false},
		{"..", false},
		{"a/b", false},
		{"x\x00y", false},
	}
	for _, c := range cases {
		got := validateName(c.name) == ""
		if got != c.want {
			t.Errorf("validateName(%q) ok=%v, want %v", c.name, got, c.want)
		}
	}
}

func TestCreatePrompt_ValidDir_EmitsSidebarCreatedMsg(t *testing.T) {
	dir := t.TempDir()
	p := newCreatePrompt("dir", dir)
	// Type "sub".
	for _, r := range "sub" {
		p, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		if p == nil {
			t.Fatal("prompt dismissed while typing")
		}
	}
	next, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if next != nil {
		t.Fatal("expected prompt to dismiss after successful create")
	}
	if cmd == nil {
		t.Fatal("expected a Cmd emitting SidebarCreatedMsg")
	}
	msg := cmd()
	got, ok := msg.(msgs.SidebarCreatedMsg)
	if !ok {
		t.Fatalf("expected SidebarCreatedMsg, got %T", msg)
	}
	if got.Kind != "dir" || got.Path != filepath.Join(dir, "sub") {
		t.Errorf("unexpected payload: %+v", got)
	}
	if info, err := os.Stat(got.Path); err != nil || !info.IsDir() {
		t.Errorf("expected directory at %s, stat err=%v", got.Path, err)
	}
}

func TestCreatePrompt_ValidFile_EmitsSidebarCreatedMsg(t *testing.T) {
	dir := t.TempDir()
	p := newCreatePrompt("file", dir)
	for _, r := range "x.txt" {
		p, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	next, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if next != nil || cmd == nil {
		t.Fatalf("expected (nil, cmd); got next=%v, cmd=%v", next, cmd)
	}
	got, ok := cmd().(msgs.SidebarCreatedMsg)
	if !ok {
		t.Fatalf("expected SidebarCreatedMsg, got %T", cmd())
	}
	if got.Kind != "file" || filepath.Base(got.Path) != "x.txt" {
		t.Errorf("unexpected payload: %+v", got)
	}
	if _, err := os.Stat(got.Path); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}

func TestCreatePrompt_Escape_CancelsWithoutSideEffect(t *testing.T) {
	dir := t.TempDir()
	p := newCreatePrompt("dir", dir)
	next, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if next != nil {
		t.Error("expected nil prompt after Esc")
	}
	if cmd != nil {
		t.Error("expected no Cmd after Esc")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("expected no filesystem changes after Esc, got %d entries", len(entries))
	}
}

func TestCreatePrompt_DuplicateName_SetsError(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "dup"), 0o755); err != nil {
		t.Fatal(err)
	}
	p := newCreatePrompt("dir", dir)
	for _, r := range "dup" {
		p, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	next, cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if next == nil {
		t.Fatal("expected prompt to stay open on duplicate name")
	}
	if cmd != nil {
		t.Error("expected no Cmd on duplicate error")
	}
	if next.err == "" {
		t.Error("expected error message set on prompt")
	}
}

func TestCreatePrompt_InvalidName_SetsError(t *testing.T) {
	dir := t.TempDir()
	p := newCreatePrompt("file", dir)
	for _, r := range "a/b" {
		p, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	next, _ := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if next == nil || next.err == "" {
		t.Error("expected prompt to stay open with error for slash-containing name")
	}
}
