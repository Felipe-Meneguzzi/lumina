package sidebar

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

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
	}
	m.loadDir(root)
	return m
}

// Width returns the current pane width.
func (m Model) Width() int { return m.width }

// CWD returns the directory currently shown in the sidebar.
func (m Model) CWD() string { return m.cwd }

// SetFocused sets the focus state.
func (m *Model) SetFocused(f bool) { m.focused = f }

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

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}
		switch msg.Type {
		case tea.KeyEnter, tea.KeyRight:
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
				return m, func() tea.Msg {
					return msgs.OpenFileMsg{Path: selected.path}
				}
			}

		case tea.KeyLeft:
			// Navigate up to parent directory.
			parent := filepath.Dir(m.cwd)
			if parent != m.cwd {
				m.cwd = parent
				m.loadDir(parent)
			}
			return m, nil
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

// View renders the sidebar pane.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	style := unfocusedBorder
	if m.focused {
		style = focusedBorder
	}
	return style.Width(m.width - 2).Height(m.height - 2).Render(m.list.View())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
