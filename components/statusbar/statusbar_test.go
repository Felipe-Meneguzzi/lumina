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

func TestUpdate_MetricsTickMsg_UpdatesCPUAndMem(t *testing.T) {
	m := newTestModel()
	m2, _ := m.Update(msgs.StatusBarResizeMsg{Width: 120})
	m = m2.(statusbar.Model)
	tick := msgs.MetricsTickMsg{
		CPU:      45.5,
		MemUsed:  2 * 1024 * 1024 * 1024,
		MemTotal: 8 * 1024 * 1024 * 1024,
		Tick:     time.Now(),
	}
	next, _ := m.Update(tick)
	nm := next.(statusbar.Model)

	view := nm.View()
	if !strings.Contains(view, "45.5") {
		t.Errorf("expected CPU 45.5 in view, got: %q", view)
	}
}

// TestFocusedPaneContext_DirtyGlyph verifies FR-006: when the focused pane
// reports a dirty git tree, the status bar renders the branch followed by `●`.
func TestFocusedPaneContext_DirtyGlyph(t *testing.T) {
	m := newTestModel()
	m2, _ := m.Update(msgs.StatusBarResizeMsg{Width: 200})
	m = m2.(statusbar.Model)
	next, _ := m.Update(msgs.FocusedPaneContextMsg{
		CWD:       "/home/user/x",
		GitBranch: "feature/foo",
		GitDirty:  true,
	})
	nm := next.(statusbar.Model)
	v := nm.View()
	if !strings.Contains(v, "feature/foo") {
		t.Errorf("expected branch in view, got: %q", v)
	}
	if !strings.Contains(v, "●") {
		t.Errorf("expected dirty glyph ● in view, got: %q", v)
	}
}

// TestFocusedPaneContext_CleanGlyph verifies the `✓` glyph path.
func TestFocusedPaneContext_CleanGlyph(t *testing.T) {
	m := newTestModel()
	m2, _ := m.Update(msgs.StatusBarResizeMsg{Width: 200})
	m = m2.(statusbar.Model)
	next, _ := m.Update(msgs.FocusedPaneContextMsg{
		CWD:       "/tmp",
		GitBranch: "main",
		GitDirty:  false,
	})
	nm := next.(statusbar.Model)
	v := nm.View()
	if !strings.Contains(v, "main") {
		t.Errorf("expected branch 'main' in view, got: %q", v)
	}
	if !strings.Contains(v, "✓") {
		t.Errorf("expected clean glyph ✓ in view, got: %q", v)
	}
}

// TestFocusedPaneContext_NoGitHidesBranch verifies FR-007: when the focused
// pane reports no branch (non-repo directory), the status bar omits the git
// segment entirely instead of reusing the previous value.
func TestFocusedPaneContext_NoGitHidesBranch(t *testing.T) {
	m := newTestModel()
	m2, _ := m.Update(msgs.StatusBarResizeMsg{Width: 200})
	m = m2.(statusbar.Model)
	next, _ := m.Update(msgs.FocusedPaneContextMsg{
		CWD:       "/home/user/x",
		GitBranch: "feature/foo",
		GitDirty:  true,
	})
	m = next.(statusbar.Model)
	next, _ = m.Update(msgs.FocusedPaneContextMsg{CWD: "/tmp"})
	nm := next.(statusbar.Model)
	v := nm.View()
	if strings.Contains(v, "feature/foo") {
		t.Errorf("branch should disappear after focus switches to non-git pane; got: %q", v)
	}
	if strings.Contains(v, "●") || strings.Contains(v, "✓") {
		t.Errorf("git glyph should disappear for non-git pane; got: %q", v)
	}
}

// TestClockTick_RendersHHMM verifies FR-005: the ClockTickMsg updates the
// displayed HH:MM segment.
func TestClockTick_RendersHHMM(t *testing.T) {
	m := newTestModel()
	m2, _ := m.Update(msgs.StatusBarResizeMsg{Width: 200})
	m = m2.(statusbar.Model)
	fixed := time.Date(2026, 4, 17, 14, 23, 0, 0, time.UTC)
	next, cmd := m.Update(msgs.ClockTickMsg{Now: fixed})
	nm := next.(statusbar.Model)
	v := nm.View()
	if !strings.Contains(v, "14:23") {
		t.Errorf("expected 14:23 in view, got: %q", v)
	}
	if cmd == nil {
		t.Error("expected a Cmd to re-arm the clock tick")
	}

	// Advance the clock — View should reflect the new time on the next tick.
	next2, _ := nm.Update(msgs.ClockTickMsg{Now: fixed.Add(time.Minute)})
	nm2 := next2.(statusbar.Model)
	if !strings.Contains(nm2.View(), "14:24") {
		t.Errorf("expected 14:24 after advancing clock, got: %q", nm2.View())
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

func TestModel_ImplementsTeaModel(t *testing.T) {
	var _ tea.Model = newTestModel()
}
