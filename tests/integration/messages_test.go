package integration_test

import (
	"testing"
	"time"

	"github.com/Felipe-Meneguzzi/lumina/components/statusbar"
	"github.com/Felipe-Meneguzzi/lumina/config"
	"github.com/Felipe-Meneguzzi/lumina/msgs"
)

var testCfg = config.Config{
	Shell:           "/bin/sh",
	MetricsInterval: 100,
	ShowHidden:      false,
	SidebarWidth:    30,
	Editor:          "nano",
}

// TestMetricsTick_ReEnqueues verifies that MetricsTickMsg produces another Cmd.
func TestMetricsTick_ReEnqueues(t *testing.T) {
	sb := statusbar.New(testCfg)
	_, cmd := sb.Update(msgs.MetricsTickMsg{
		Tick: time.Now(),
	})
	if cmd == nil {
		t.Error("expected non-nil Cmd from MetricsTickMsg (re-enqueue next tick)")
	}
}

// TestWindowResize_StatusBar verifies that StatusBarResizeMsg propagates width.
func TestWindowResize_StatusBar(t *testing.T) {
	sb := statusbar.New(testCfg)
	next, _ := sb.Update(msgs.StatusBarResizeMsg{Width: 120})
	nm := next.(statusbar.Model)
	if nm.Width() != 120 {
		t.Errorf("expected width 120, got %d", nm.Width())
	}
}
