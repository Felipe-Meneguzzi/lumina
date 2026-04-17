package statusbar

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	barStyle    = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("252"))
	notifyStyle = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("214")).Bold(true)
	errStyle    = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("196")).Bold(true)
)

type notification struct {
	text    string
	level   msgs.NotifyLevel
	expires time.Time
}

// Model is the Bubble Tea model for the status bar.
type Model struct {
	interval  time.Duration
	cpu       float64
	memUsed   uint64
	memTotal  uint64
	cwd       string
	gitBranch string
	width     int
	notify    *notification
}

// New creates a new statusbar Model.
func New(cfg config.Config) Model {
	return Model{
		interval: time.Duration(cfg.MetricsInterval) * time.Millisecond,
		width:    80,
	}
}

// Width returns the current render width.
func (m Model) Width() int { return m.width }

func tickMetrics(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		var cpuPct float64
		if pcts, err := cpu.Percent(0, false); err == nil && len(pcts) > 0 {
			cpuPct = pcts[0]
		}

		var memUsed, memTotal uint64
		if vm, err := mem.VirtualMemory(); err == nil {
			memUsed = vm.Used
			memTotal = vm.Total
		}

		branch := gitBranch()

		return msgs.MetricsTickMsg{
			CPU:       cpuPct,
			MemUsed:   memUsed,
			MemTotal:  memTotal,
			GitBranch: branch,
			Tick:      t,
		}
	})
}

func gitBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// Init starts the metrics tick.
func (m Model) Init() tea.Cmd {
	return tickMetrics(m.interval)
}

// Update handles messages for the status bar.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case msgs.MetricsTickMsg:
		m.cpu = msg.CPU
		m.memUsed = msg.MemUsed
		m.memTotal = msg.MemTotal
		if msg.GitBranch != "" {
			m.gitBranch = msg.GitBranch
		}
		if msg.CWD != "" {
			m.cwd = msg.CWD
		}
		// Clear expired notifications.
		if m.notify != nil && time.Now().After(m.notify.expires) {
			m.notify = nil
		}
		return m, tickMetrics(m.interval)

	case msgs.StatusBarNotifyMsg:
		m.notify = &notification{
			text:    msg.Text,
			level:   msg.Level,
			expires: time.Now().Add(msg.Duration),
		}
		return m, nil

	case msgs.StatusBarResizeMsg:
		m.width = msg.Width
		return m, nil
	}

	return m, nil
}

// View renders the one-line status bar.
func (m Model) View() string {
	if m.notify != nil && time.Now().Before(m.notify.expires) {
		style := notifyStyle
		if m.notify.level == msgs.NotifyError {
			style = errStyle
		}
		return style.Width(m.width).Render(" " + m.notify.text)
	}

	memGB := float64(m.memUsed) / (1024 * 1024 * 1024)
	totalGB := float64(m.memTotal) / (1024 * 1024 * 1024)

	var parts []string
	parts = append(parts, fmt.Sprintf(" CPU: %.1f%%", m.cpu))
	parts = append(parts, fmt.Sprintf("MEM: %.1f/%.1fGB", memGB, totalGB))
	if m.gitBranch != "" {
		parts = append(parts, fmt.Sprintf("[%s]", m.gitBranch))
	}
	if m.cwd != "" {
		parts = append(parts, m.cwd)
	}

	line := strings.Join(parts, "  ") + " "
	runes := []rune(line)
	if len(runes) > m.width && m.width > 3 {
		line = string(runes[:m.width-3]) + "..."
	}

	return barStyle.Width(m.width).Render(line)
}
