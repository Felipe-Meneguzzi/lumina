package statusbar

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Felipe-Meneguzzi/lumina/config"
	"github.com/Felipe-Meneguzzi/lumina/msgs"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// clockInterval is the cadence of the HH:MM ticker. A 30-second interval keeps
// SC-006 (max 60s drift) comfortably met without waking the event loop more
// than necessary.
const clockInterval = 30 * time.Second

var (
	barStyle    = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("252"))
	notifyStyle = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("214")).Bold(true)
	errStyle    = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("196")).Bold(true)
	dirtyStyle  = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("214")).Bold(true)
	cleanStyle  = lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("42")).Bold(true)
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
	now       time.Time
	cwd       string // derived from FocusedPaneContextMsg
	gitBranch string // derived from FocusedPaneContextMsg (empty = no git)
	gitDirty  bool
	width     int
	notify    *notification
}

// New creates a new statusbar Model.
func New(cfg config.Config) Model {
	return Model{
		interval: time.Duration(cfg.MetricsInterval) * time.Millisecond,
		width:    80,
		now:      time.Now(),
	}
}

// Width returns the current render width.
func (m Model) Width() int { return m.width }

// ClockInterval returns the cadence of the statusbar clock tick (exported for tests).
func ClockInterval() time.Duration { return clockInterval }

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

		return msgs.MetricsTickMsg{
			CPU:      cpuPct,
			MemUsed:  memUsed,
			MemTotal: memTotal,
			Tick:     t,
		}
	})
}

// tickClock re-emits a ClockTickMsg every clockInterval seconds. The first
// emission is immediate so the HH:MM value shows up on the very first frame
// after Init rather than after the first 30s interval.
func tickClock() tea.Cmd {
	return tea.Tick(clockInterval, func(t time.Time) tea.Msg {
		return msgs.ClockTickMsg{Now: t}
	})
}

// Init starts the metrics + clock tickers. The clock fires immediately to
// populate the HH:MM field before the first 30s elapses.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tickMetrics(m.interval),
		func() tea.Msg { return msgs.ClockTickMsg{Now: time.Now()} },
	)
}

// Update handles messages for the status bar.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case msgs.MetricsTickMsg:
		m.cpu = msg.CPU
		m.memUsed = msg.MemUsed
		m.memTotal = msg.MemTotal
		// Clear expired notifications.
		if m.notify != nil && time.Now().After(m.notify.expires) {
			m.notify = nil
		}
		return m, tickMetrics(m.interval)

	case msgs.ClockTickMsg:
		m.now = msg.Now
		return m, tickClock()

	case msgs.FocusedPaneContextMsg:
		m.cwd = msg.CWD
		m.gitBranch = msg.GitBranch
		m.gitDirty = msg.GitDirty
		return m, nil

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

	// Clock segment goes first, then metrics, then focused-pane context.
	clock := m.now.Format("15:04")
	var parts []string
	parts = append(parts, " "+clock)
	parts = append(parts, fmt.Sprintf("CPU: %.1f%%", m.cpu))
	parts = append(parts, fmt.Sprintf("MEM: %.1f/%.1fGB", memGB, totalGB))
	if m.cwd != "" {
		parts = append(parts, m.cwd)
	}

	line := strings.Join(parts, "  ")

	// Git segment rendered with dedicated style for the glyph.
	if m.gitBranch != "" {
		glyph := "✓"
		style := cleanStyle
		if m.gitDirty {
			glyph = "●"
			style = dirtyStyle
		}
		gitSeg := fmt.Sprintf("  %s %s", m.gitBranch, glyph)
		line += style.Render(gitSeg)
	}

	runes := []rune(line)
	if len(runes) > m.width && m.width > 3 {
		line = string(runes[:m.width-3]) + "..."
	}

	return barStyle.Width(m.width).Render(line)
}
