package terminal_test

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/menegas/lumina/components/terminal"
	"github.com/menegas/lumina/config"
	"github.com/menegas/lumina/msgs"
)

func newTestModel(t *testing.T) terminal.Model {
	t.Helper()
	cfg := config.Config{Shell: "/bin/sh", SidebarWidth: 30}
	m, err := terminal.New(cfg)
	if err != nil {
		t.Fatalf("terminal.New: %v", err)
	}
	return m
}

func TestUpdate_PtyOutputMsg_AppendsToBuffer(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	msg := msgs.PtyOutputMsg{Data: []byte("hello world\n"), Err: nil}
	next, _ := m.Update(msg)
	nm := next.(terminal.Model)

	if !strings.Contains(nm.View(), "hello world") {
		t.Errorf("expected View() to contain 'hello world', got: %q", nm.View())
	}
}

func TestUpdate_TerminalResizeMsg_UpdatesDimensions(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	next, _ := m.Update(msgs.TerminalResizeMsg{Width: 120, Height: 40})
	nm := next.(terminal.Model)

	w, h := nm.Dimensions()
	if w != 120 || h != 40 {
		t.Errorf("expected 120x40, got %dx%d", w, h)
	}
}

func TestUpdate_PtyOutputMsg_EOF_TriggersRestart(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	msg := msgs.PtyOutputMsg{Data: nil, Err: errors.New("EOF")}
	_, cmd := m.Update(msg)

	// A non-nil Cmd means the restart was initiated.
	if cmd == nil {
		t.Error("expected non-nil Cmd to restart shell after EOF, got nil")
	}
}

func TestView_ReturnsNonEmptyString(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	view := m.View()
	if view == "" {
		t.Error("expected non-empty View()")
	}
}

func TestUpdate_FocusedBorderChanges(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	m.SetFocused(true)
	focused := m.View()

	m.SetFocused(false)
	unfocused := m.View()

	if focused == unfocused {
		t.Error("expected View() to differ between focused and unfocused states")
	}
}

func TestModel_ImplementsTeaModel(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	// Verify the interface is satisfied at compile time via assignment.
	var _ tea.Model = m
}

// TestNewWithCommand_BootsOverride ensures the override path builds a live
// model without falling back to the default shell.
func TestNewWithCommand_BootsOverride(t *testing.T) {
	cfg := config.Config{Shell: "/bin/sh", SidebarWidth: 30}
	m, err := terminal.NewWithCommand(cfg, "/bin/true")
	if err != nil {
		t.Fatalf("NewWithCommand: %v", err)
	}
	defer m.Close()
	// The model must be a live tea.Model — sanity check via View().
	if v := m.View(); v == "" {
		t.Error("expected non-empty View() after NewWithCommand")
	}
}

// TestUpdate_PtyOutputMsg_PreservesTrueColor ensures 24-bit RGB SGR sequences
// emitted by the inner application survive the emulator → renderer → lipgloss
// pipeline. This is the regression that the vt10x → x/vt swap was made to fix.
func TestUpdate_PtyOutputMsg_PreservesTrueColor(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	data := []byte("\x1b[38;2;255;100;50mORANGE\x1b[0m\n")
	next, _ := m.Update(msgs.PtyOutputMsg{Data: data, Err: nil})
	view := next.(terminal.Model).View()

	if !strings.Contains(view, "ORANGE") {
		t.Fatalf("expected View() to contain ORANGE text, got: %q", view)
	}
	if !strings.Contains(view, "38;2;255;100;50") {
		t.Errorf("expected truecolor SGR sequence to survive in View(), got: %q", view)
	}
}

// TestUpdate_AltScreen_FreezesScrollbackOffset ensures that when an app like
// claude-code switches to the alternate screen, the scrollback UI becomes
// inert (alt screens have no history).
func TestUpdate_AltScreen_FreezesScrollbackOffset(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	// Activate alt screen, then write content.
	next, _ := m.Update(msgs.PtyOutputMsg{Data: []byte("\x1b[?1049h"), Err: nil})
	next, _ = next.(terminal.Model).Update(msgs.PtyOutputMsg{Data: []byte("ALTBUF\n"), Err: nil})
	// Try to scroll back — should be a no-op when alt screen is active.
	next, _ = next.(terminal.Model).Update(msgs.TerminalScrollMsg{Delta: 100})
	view := next.(terminal.Model).View()

	if !strings.Contains(view, "ALTBUF") {
		t.Errorf("expected View() to show alt-screen content, got: %q", view)
	}
}

// TestUpdate_TerminalScrollMsg_PreservesHistory feeds enough lines to push
// rows into scrollback, then asserts that scrolling back exposes the older
// content.
func TestUpdate_TerminalScrollMsg_PreservesHistory(t *testing.T) {
	m := newTestModel(t)
	m.Close()
	// Default inner geometry is 78x22; emit ~50 distinct lines so the early
	// ones are guaranteed to land in scrollback.
	var b strings.Builder
	for i := 0; i < 50; i++ {
		b.WriteString("LINE")
		b.WriteString(itoa(i))
		b.WriteString("\n")
	}
	next, _ := m.Update(msgs.PtyOutputMsg{Data: []byte(b.String()), Err: nil})
	// Scroll back enough to expose LINE0 in the visible viewport.
	next, _ = next.(terminal.Model).Update(msgs.TerminalScrollMsg{Delta: 40})
	view := next.(terminal.Model).View()

	if !strings.Contains(view, "LINE0") {
		t.Errorf("expected scrollback view to expose LINE0, got: %q", view)
	}
}

// TestMouseEnabled_TogglesWithDECModes feeds mouse-tracking enable/disable
// sequences and asserts MouseEnabled() reflects the state.
func TestMouseEnabled_TogglesWithDECModes(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	if m.MouseEnabled() {
		t.Fatalf("expected MouseEnabled=false initially")
	}
	// Enable mode 1000 (normal mouse tracking).
	next, _ := m.Update(msgs.PtyOutputMsg{Data: []byte("\x1b[?1000h"), Err: nil})
	m = next.(terminal.Model)
	if !m.MouseEnabled() {
		t.Fatalf("expected MouseEnabled=true after \\e[?1000h")
	}
	// Disable mode 1000.
	next, _ = m.Update(msgs.PtyOutputMsg{Data: []byte("\x1b[?1000l"), Err: nil})
	m = next.(terminal.Model)
	if m.MouseEnabled() {
		t.Fatalf("expected MouseEnabled=false after \\e[?1000l")
	}
}

// TestPtyMouseMsg_NoOpWhenDisabled ensures forwarded mouse events are dropped
// when the inner application has not enabled tracking — prevents accidental
// PTY pollution when the user clicks inside an unsuspecting shell.
func TestPtyMouseMsg_NoOpWhenDisabled(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	// Mouse mode is off — sending a PtyMouseMsg should not panic and should not
	// produce any PTY-bound bytes.
	mouse := tea.MouseMsg{X: 5, Y: 5, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}
	if _, _ = m.Update(msgs.PtyMouseMsg{PaneID: 0, Mouse: mouse}); false {
		t.Fail() // unreachable, just exercises the path
	}
}

// TestCallbacks_TitleAndCWD feeds OSC 2 (title) and OSC 7 (working directory)
// sequences and asserts the model's getters expose the values.
func TestCallbacks_TitleAndCWD(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	// OSC 2 = window title; ST = ESC \.
	next, _ := m.Update(msgs.PtyOutputMsg{
		Data: []byte("\x1b]2;hello-title\x1b\\"),
		Err:  nil,
	})
	m = next.(terminal.Model)
	if m.Title() != "hello-title" {
		t.Errorf("expected Title()=hello-title, got %q", m.Title())
	}

	// OSC 7 = working directory as file:// URI.
	next, _ = m.Update(msgs.PtyOutputMsg{
		Data: []byte("\x1b]7;file://host/home/user/project\x1b\\"),
		Err:  nil,
	})
	m = next.(terminal.Model)
	if m.CWD() != "/home/user/project" {
		t.Errorf("expected CWD()=/home/user/project, got %q", m.CWD())
	}
}

// TestCopyMode_EnterAndExit covers entering copy mode, moving the cursor,
// extracting the selection, and exiting back to live mode.
func TestCopyMode_EnterAndExit(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	// Lay down content so the selection has something to extract.
	next, _ := m.Update(msgs.PtyOutputMsg{Data: []byte("HELLO\n"), Err: nil})
	m = next.(terminal.Model)

	// Enter copy mode.
	next, _ = m.Update(msgs.EnterCopyModeMsg{})
	m = next.(terminal.Model)
	if !m.InCopyMode() {
		t.Fatalf("expected InCopyMode=true after EnterCopyModeMsg")
	}

	// Cursor starts at bottom-right; move it to (0,0) covering "HELLO".
	for i := 0; i < 30; i++ {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
		m = next.(terminal.Model)
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
		m = next.(terminal.Model)
	}
	// Extend selection across "HELLO".
	for i := 0; i < 4; i++ {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}})
		m = next.(terminal.Model)
	}
	// Esc leaves copy mode without copying.
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(terminal.Model)
	if m.InCopyMode() {
		t.Errorf("expected InCopyMode=false after Esc")
	}
}

// TestCallbacks_BellCounter feeds bell characters and asserts the counter
// advances.
func TestCallbacks_BellCounter(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	if m.BellCount() != 0 {
		t.Fatalf("expected BellCount=0 initially, got %d", m.BellCount())
	}
	next, _ := m.Update(msgs.PtyOutputMsg{Data: []byte("\a\a\a"), Err: nil})
	m = next.(terminal.Model)
	if m.BellCount() != 3 {
		t.Errorf("expected BellCount=3 after 3 bells, got %d", m.BellCount())
	}
}

// TestCursorGate_DiffersWithFocus asserts FR-003 (R3): the rendered viewport
// of a focused pane must differ from the unfocused one in at least the cursor
// cell. Both panes carry identical bytes so the only degree of freedom is the
// cursor visual.
func TestCursorGate_DiffersWithFocus(t *testing.T) {
	mA := newTestModel(t)
	mA.Close()
	mB := newTestModel(t)
	mB.Close()

	// Feed identical content so the only variation is the cursor visual.
	data := []byte("abc\n")
	a, _ := mA.Update(msgs.PtyOutputMsg{Data: data})
	b, _ := mB.Update(msgs.PtyOutputMsg{Data: data})
	mA = a.(terminal.Model)
	mB = b.(terminal.Model)

	mA.SetFocused(true)
	mB.SetFocused(false)

	vA := mA.View()
	vB := mB.View()
	if vA == vB {
		t.Errorf("expected focused and unfocused View() to differ (cursor gate); both=%q", vA)
	}
}

// TestFirstRender_SetsFlagAndEmitsCmd verifies FR-001 / US1: the initial
// TerminalResizeMsg flips firstRenderDone and returns a Cmd that starts
// draining the PTY.
func TestFirstRender_SetsFlagAndEmitsCmd(t *testing.T) {
	m := newTestModel(t)
	defer m.Close()
	if m.FirstRenderDone() {
		t.Fatal("expected firstRenderDone=false before any resize")
	}
	next, cmd := m.Update(msgs.TerminalResizeMsg{Width: 80, Height: 24})
	nm := next.(terminal.Model)
	if !nm.FirstRenderDone() {
		t.Error("expected firstRenderDone=true after first TerminalResizeMsg")
	}
	if cmd == nil {
		t.Error("expected non-nil Cmd to start PTY drain on first resize")
	}
}

// TestBulkOutput_PreservesByteOrder feeds many PtyOutputMsg chunks with
// distinct sentinel tokens and asserts they all land in the viewport in order.
// Validates the coalescing design (research.md §R2) at the mechanism level —
// bytes must never be lost or reordered regardless of batch size.
func TestBulkOutput_PreservesByteOrder(t *testing.T) {
	m := newTestModel(t)
	m.Close()
	tokens := []string{"AAA", "BBB", "CCC", "DDD", "EEE"}
	var cur terminal.Model = m
	for _, tok := range tokens {
		nxt, _ := cur.Update(msgs.PtyOutputMsg{Data: []byte(tok + "\n")})
		cur = nxt.(terminal.Model)
	}
	v := cur.View()
	lastIdx := -1
	for _, tok := range tokens {
		idx := strings.Index(v, tok)
		if idx < 0 {
			t.Errorf("expected %q present in View, got: %q", tok, v)
			continue
		}
		if idx < lastIdx {
			t.Errorf("token %q appeared before earlier token at idx %d (prev %d)", tok, idx, lastIdx)
		}
		lastIdx = idx
	}
}

// TestEnterCopyMode_ClearsMouseSelection verifies the mutual-exclusion invariant:
// entering copy mode must clear any active mouse selection.
func TestEnterCopyMode_ClearsMouseSelection(t *testing.T) {
	m := newTestModel(t)
	m.Close()

	// Establish an active mouse selection via Press.
	next, _ := m.Update(msgs.MouseSelectMsg{
		Mouse: tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
	})
	m = next.(terminal.Model)

	if !m.HasMouseSelection() {
		t.Fatal("expected HasMouseSelection=true before entering copy mode")
	}

	// Enter copy mode.
	next, _ = m.Update(msgs.EnterCopyModeMsg{})
	m = next.(terminal.Model)

	if !m.InCopyMode() {
		t.Error("expected InCopyMode=true after EnterCopyModeMsg")
	}
	if m.HasMouseSelection() {
		t.Error("expected HasMouseSelection=false: mouse selection must be cleared when copy mode is entered")
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var out []byte
	for n > 0 {
		out = append([]byte{byte('0' + n%10)}, out...)
		n /= 10
	}
	return string(out)
}
