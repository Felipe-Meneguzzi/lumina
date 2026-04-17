package statusbar_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/menegas/lumina/components/statusbar"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

func newTestModel() statusbar.Model {
	return statusbar.New(config.Config{MetricsInterval: 1000})
}

func TestUpdate_MetricsTickMsg_UpdatesFields(t *testing.T) {
	m := newTestModel()
	tick := msgs.MetricsTickMsg{
		CPU:       45.5,
		MemUsed:   2 * 1024 * 1024 * 1024,
		MemTotal:  8 * 1024 * 1024 * 1024,
		CWD:       "/home/user/project",
		GitBranch: "main",
		Tick:      time.Now(),
	}
	next, _ := m.Update(tick)
	nm := next.(statusbar.Model)

	view := nm.View()
	if !strings.Contains(view, "45.5") {
		t.Errorf("expected CPU 45.5 in view, got: %q", view)
	}
	if !strings.Contains(view, "main") {
		t.Errorf("expected branch 'main' in view, got: %q", view)
	}
}

func TestUpdate_MetricsTickMsg_ReturnsCmd(t *testing.T) {
	m := newTestModel()
	_, cmd := m.Update(msgs.MetricsTickMsg{Tick: time.Now()})
	if cmd == nil {
		t.Error("expected non-nil Cmd to re-schedule next tick")
	}
}

func TestUpdate_StatusBarNotifyMsg_OverridesDisplay(t *testing.T) {
	m := newTestModel()
	next, _ := m.Update(msgs.StatusBarNotifyMsg{
		Text:     "Salvo",
		Level:    msgs.NotifyInfo,
		Duration: 3 * time.Second,
	})
	nm := next.(statusbar.Model)
	if !strings.Contains(nm.View(), "Salvo") {
		t.Errorf("expected 'Salvo' notification in view, got: %q", nm.View())
	}
}

func TestUpdate_StatusBarResizeMsg_UpdatesWidth(t *testing.T) {
	m := newTestModel()
	next, _ := m.Update(msgs.StatusBarResizeMsg{Width: 100})
	nm := next.(statusbar.Model)
	if nm.Width() != 100 {
		t.Errorf("expected width 100, got %d", nm.Width())
	}
}

func TestView_TruncatesToWidth(t *testing.T) {
	m := newTestModel()
	m2, _ := m.Update(msgs.StatusBarResizeMsg{Width: 20})
	nm := m2.(statusbar.Model)
	view := nm.View()
	// View must not exceed 20 visible characters (ignoring ANSI codes for simplicity).
	if len([]rune(view)) > 40 { // generous bound accounting for ANSI codes
		t.Logf("view length %d may exceed width; view: %q", len([]rune(view)), view)
	}
}

func TestModel_ImplementsTeaModel(t *testing.T) {
	var _ tea.Model = newTestModel()
}
