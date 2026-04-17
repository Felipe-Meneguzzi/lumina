package terminal

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/vt"
)

// sharedState carries mutable runtime state that Bubble Tea callbacks need to
// update from inside vt.Write. Held by pointer so the callback closures and
// successive Model copies share the same memory (Bubble Tea copies Models by
// value, but they all reference this single instance).
type sharedState struct {
	// Mouse-tracking DEC modes set by the inner application.
	mouseNormal   bool // mode 1000 (X10/normal)
	mouseBtnEvent bool // mode 1002 (button event tracking)
	mouseAnyEvent bool // mode 1003 (any event tracking)

	// Window title (OSC 0/2) — last value reported by the inner app.
	title string
	// Working directory (OSC 7) — last value reported, decoded from file:// URI.
	cwd string
	// Number of bell characters received since startup.
	bellCount int
}

func (s *sharedState) MouseEnabled() bool {
	return s != nil && (s.mouseNormal || s.mouseBtnEvent || s.mouseAnyEvent)
}

// MouseEnabled reports whether the inner application has put the terminal
// into a mouse tracking mode. Used by the app layer to decide whether to
// forward tea.MouseMsg to the PTY or handle it locally.
func (m Model) MouseEnabled() bool { return m.state.MouseEnabled() }

// Title returns the last window title set by the inner application via
// OSC 0/2. Empty string if the app never set one.
func (m Model) Title() string {
	if m.state == nil {
		return ""
	}
	return m.state.title
}

// CWD returns the last working directory reported by the inner application
// via OSC 7. Empty string if the app never set one.
func (m Model) CWD() string {
	if m.state == nil {
		return ""
	}
	return m.state.cwd
}

// BellCount returns how many bell characters the inner application has
// emitted since startup. Useful for visual indicators of activity.
func (m Model) BellCount() int {
	if m.state == nil {
		return 0
	}
	return m.state.bellCount
}

// installCallbacks wires DEC mode + OSC callbacks on the emulator so the
// model can observe changes to mouse tracking, title, working directory and
// bell. Called once per emulator (initial create + after EOF restart).
func installCallbacks(e *vt.Emulator, state *sharedState) {
	e.SetCallbacks(vt.Callbacks{
		EnableMode: func(mode ansi.Mode) {
			if dm, ok := mode.(ansi.DECMode); ok {
				switch dm {
				case ansi.ModeMouseNormal:
					state.mouseNormal = true
				case ansi.ModeMouseButtonEvent:
					state.mouseBtnEvent = true
				case ansi.ModeMouseAnyEvent:
					state.mouseAnyEvent = true
				}
			}
		},
		DisableMode: func(mode ansi.Mode) {
			if dm, ok := mode.(ansi.DECMode); ok {
				switch dm {
				case ansi.ModeMouseNormal:
					state.mouseNormal = false
				case ansi.ModeMouseButtonEvent:
					state.mouseBtnEvent = false
				case ansi.ModeMouseAnyEvent:
					state.mouseAnyEvent = false
				}
			}
		},
		Title:            func(s string) { state.title = s },
		WorkingDirectory: func(s string) { state.cwd = decodeCWD(s) },
		Bell:             func() { state.bellCount++ },
	})
}

// decodeCWD turns an OSC 7 payload (typically a file:// URI) into a plain
// path. Falls back to the raw string when no URI prefix is present.
func decodeCWD(raw string) string {
	const prefix = "file://"
	if !strings.HasPrefix(raw, prefix) {
		return raw
	}
	rest := raw[len(prefix):]
	// Skip the host portion (everything up to the first '/').
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		return rest[i:]
	}
	return rest
}

// teaMouseToVT converts a Bubble Tea mouse message (already translated to
// pane-local coordinates) into the corresponding ultraviolet event and
// dispatches it to the emulator. Bytes flow back via the InputPipe goroutine.
func teaMouseToVT(e *vt.Emulator, msg tea.MouseMsg) {
	m := uv.Mouse{
		X:      msg.X,
		Y:      msg.Y,
		Button: teaButtonToANSI(msg.Button),
		Mod:    teaModsToUV(msg.Alt, msg.Ctrl, msg.Shift),
	}
	switch {
	case isWheelButton(msg.Button):
		e.SendMouse(uv.MouseWheelEvent(m))
	case msg.Action == tea.MouseActionRelease:
		e.SendMouse(uv.MouseReleaseEvent(m))
	case msg.Action == tea.MouseActionMotion:
		e.SendMouse(uv.MouseMotionEvent(m))
	default:
		e.SendMouse(uv.MouseClickEvent(m))
	}
}

func isWheelButton(b tea.MouseButton) bool {
	switch b {
	case tea.MouseButtonWheelUp, tea.MouseButtonWheelDown,
		tea.MouseButtonWheelLeft, tea.MouseButtonWheelRight:
		return true
	}
	return false
}

func teaButtonToANSI(b tea.MouseButton) ansi.MouseButton {
	switch b {
	case tea.MouseButtonLeft:
		return ansi.MouseLeft
	case tea.MouseButtonMiddle:
		return ansi.MouseMiddle
	case tea.MouseButtonRight:
		return ansi.MouseRight
	case tea.MouseButtonWheelUp:
		return ansi.MouseWheelUp
	case tea.MouseButtonWheelDown:
		return ansi.MouseWheelDown
	case tea.MouseButtonWheelLeft:
		return ansi.MouseWheelLeft
	case tea.MouseButtonWheelRight:
		return ansi.MouseWheelRight
	case tea.MouseButtonBackward:
		return ansi.MouseBackward
	case tea.MouseButtonForward:
		return ansi.MouseForward
	}
	return ansi.MouseNone
}

func teaModsToUV(alt, ctrl, shift bool) uv.KeyMod {
	var m uv.KeyMod
	if alt {
		m |= uv.ModAlt
	}
	if ctrl {
		m |= uv.ModCtrl
	}
	if shift {
		m |= uv.ModShift
	}
	return m
}
