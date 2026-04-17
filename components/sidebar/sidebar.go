package sidebar

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

// alreadyAtRootDuration controls how long the status bar keeps showing the
// "Já na raiz" notification when the user hits Backspace at the configured
// root (feature 006 / FR-009).
const alreadyAtRootDuration = 2 * time.Second

var (
	focusedBorder   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62"))
	unfocusedBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240"))
)

// entry represents a file or directory item in the sidebar list.
type entry struct {
	path  string
	name  string
	isDir bool
}

func (e entry) Title() string {
	if e.isDir {
		return "▸ " + e.name
	}
	return "  " + e.name
}

func (e entry) Description() string { return "" }
func (e entry) FilterValue() string { return e.name }

// Model is the Bubble Tea model for the sidebar file explorer.
type Model struct {
	list       list.Model
	root       string
	cwd        string
	showHidden bool
	focused    bool
	width      int
	height     int
	keys       config.Keybindings
	prompt     *createPrompt
}

// New creates a sidebar rooted at the given directory.
func New(root string, cfg config.Config) Model {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	l := list.New(nil, delegate, cfg.SidebarWidth-2, 20)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	m := Model{
		list:       l,
		root:       root,
		cwd:        root,
		showHidden: cfg.ShowHidden,
		width:      cfg.SidebarWidth,
		height:     24,
		keys:       cfg.Keys,
	}
	m.loadDir(root)
	return m
}

// Width returns the current pane width.
func (m Model) Width() int { return m.width }

// CWD returns the directory currently shown in the sidebar.
func (m Model) CWD() string { return m.cwd }

// Root returns the sidebar's configured root (the Lumina start directory).
func (m Model) Root() string { return m.root }

// SetFocused sets the focus state.
func (m *Model) SetFocused(f bool) { m.focused = f }

// PromptActive reports whether the inline create prompt is currently open.
// Used by tests and by the app layer to skip unrelated key routing while the
// user is typing a name.
func (m Model) PromptActive() bool { return m.prompt != nil }

func (m *Model) loadDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// Sort: dirs first, then files, both alphabetical.
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	items := make([]list.Item, 0, len(entries))
	for _, de := range entries {
		if !m.showHidden && strings.HasPrefix(de.Name(), ".") {
			continue
		}
		items = append(items, entry{
			path:  filepath.Join(dir, de.Name()),
			name:  de.Name(),
			isDir: de.IsDir(),
		})
	}
	m.list.SetItems(items)
}

func fileInfo(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

// Init loads the root directory (no async Cmd needed).
func (m Model) Init() tea.Cmd { return nil }

// Update handles messages for the sidebar.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case msgs.SidebarCreatedMsg:
		// Refresh the listing so the new entry is visible and, for dirs,
		// navigate into the newly created directory.
		if msg.Kind == "dir" {
			m.cwd = msg.Path
			m.loadDir(msg.Path)
		} else {
			m.loadDir(m.cwd)
		}
		return m, nil

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		// Inline create prompt consumes every key while active.
		if m.prompt != nil {
			next, cmd := m.prompt.Update(msg)
			if next == nil {
				m.prompt = nil
				// Refresh listing after successful create.
				m.loadDir(m.cwd)
			} else {
				m.prompt = next
			}
			return m, cmd
		}
		switch m.keys.Action(msg.String()) {
		case "sidebar_new_dir":
			m.prompt = newCreatePrompt("dir", m.cwd)
			return m, nil
		case "sidebar_new_file":
			m.prompt = newCreatePrompt("file", m.cwd)
			return m, nil
		case "sidebar_parent":
			return m.navigateParent()
		}
		switch msg.Type {
		case tea.KeyEnter:
			if selected, ok := m.list.SelectedItem().(entry); ok {
				info, err := fileInfo(selected.path)
				if err != nil {
					return m, nil
				}
				if info.IsDir() {
					m.cwd = selected.path
					m.loadDir(selected.path)
					return m, nil
				}
				path := selected.path
				return m, func() tea.Msg { return msgs.OpenInExternalEditorMsg{Path: path} }
			}
		}

	case msgs.SidebarResizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(max(1, m.width-2), max(1, m.height-2))
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// navigateParent moves the sidebar one directory up, stopping at the
// configured root. When already at the root, emits a transient status-bar
// notification without changing the sidebar state.
func (m Model) navigateParent() (tea.Model, tea.Cmd) {
	parent := filepath.Dir(m.cwd)
	atRoot := parent == m.cwd || m.cwd == m.root
	if atRoot {
		return m, func() tea.Msg {
			return msgs.StatusBarNotifyMsg{
				Text:     "Já na raiz",
				Level:    msgs.NotifyInfo,
				Duration: alreadyAtRootDuration,
			}
		}
	}
	m.cwd = parent
	m.loadDir(parent)
	return m, nil
}

// View renders the sidebar pane.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	style := unfocusedBorder
	if m.focused {
		style = focusedBorder
	}
	content := m.list.View()
	if m.prompt != nil {
		content = m.prompt.View() + "\n" + content
	}
	return style.Width(m.width - 2).Height(m.height - 2).Render(content)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
