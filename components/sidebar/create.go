package sidebar

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Felipe-Meneguzzi/lumina/msgs"
)

// createPrompt is the ephemeral submodel that captures the user's input when
// creating a file (alt+f) or directory (alt+d) from the sidebar. It lives in
// sidebar.Model.prompt and consumes every key while active.
type createPrompt struct {
	kind      string // "dir" | "file"
	parentDir string
	input     textinput.Model
	err       string
}

var (
	promptLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true)
	promptErrStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// newCreatePrompt builds a new submodel rooted at parentDir.
func newCreatePrompt(kind, parentDir string) *createPrompt {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 255
	if kind == "dir" {
		ti.Placeholder = "new-folder"
	} else {
		ti.Placeholder = "arquivo.txt"
	}
	return &createPrompt{
		kind:      kind,
		parentDir: parentDir,
		input:     ti,
	}
}

// Update consumes a key message. Returns (nil, cmd) when the prompt should be
// dismissed (either via ESC or a successful create); otherwise returns the
// updated prompt.
func (p *createPrompt) Update(msg tea.KeyMsg) (*createPrompt, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		return nil, nil
	case tea.KeyEnter:
		name := strings.TrimSpace(p.input.Value())
		if err := validateName(name); err != "" {
			p.err = err
			return p, nil
		}
		path := filepath.Join(p.parentDir, name)
		if _, err := os.Stat(path); err == nil {
			p.err = "already exists"
			return p, nil
		}
		if err := create(p.kind, path); err != nil {
			p.err = err.Error()
			return p, nil
		}
		kind := p.kind
		return nil, func() tea.Msg { return msgs.SidebarCreatedMsg{Kind: kind, Path: path} }
	}
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	// Typing after an error message clears the error so the user can retry
	// without the stale complaint lingering.
	if p.err != "" {
		p.err = ""
	}
	return p, cmd
}

// View renders the prompt line: label + input + optional error.
func (p *createPrompt) View() string {
	label := "New folder: "
	if p.kind == "file" {
		label = "New file: "
	}
	out := promptLabelStyle.Render(label) + p.input.View()
	if p.err != "" {
		out += "\n" + promptErrStyle.Render(p.err)
	}
	return out
}

// validateName returns an empty string when name is valid, or a human-readable
// reason why it is not. Linux filesystem invariants: empty names, path
// separators (`/`), and the NUL byte are rejected. Leading dots are allowed
// (the user may want to create `.gitignore`), but `.` and `..` themselves are
// not sensible names for a new entry and are rejected.
func validateName(name string) string {
	if name == "" {
		return "empty name"
	}
	if name == "." || name == ".." {
		return "invalid name"
	}
	if strings.ContainsRune(name, '/') || strings.ContainsRune(name, 0) {
		return "invalid characters"
	}
	return ""
}

// create performs the filesystem mutation. Directories are created with 0o755
// and files with 0o644 content-empty.
func create(kind, path string) error {
	if kind == "dir" {
		return os.Mkdir(path, 0o755)
	}
	return os.WriteFile(path, nil, 0o644)
}
