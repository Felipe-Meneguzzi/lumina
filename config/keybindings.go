package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Keybindings defines all user-configurable key sequences for Lumina.
// Each action maps to a list of keys — any of them triggers the action.
// Key strings follow Bubble Tea notation: "ctrl+s", "alt+1", "f1", "?", etc.
type Keybindings struct {
	FocusSidebar     []string `json:"focus_sidebar"`
	FocusTerminal    []string `json:"focus_terminal"`
	FocusEditor      []string `json:"focus_editor"`
	OpenTerminalHere []string `json:"open_terminal_here"`
	Save             []string `json:"save"`
	Quit             []string `json:"quit"`
	Help             []string `json:"help"`

	// Multiwindow: pane splits
	SplitHorizontal []string `json:"split_horizontal"`
	SplitVertical   []string `json:"split_vertical"`
	ClosePane       []string `json:"close_pane"`

	// Multiwindow: pane focus navigation (Hyprland-inspired)
	FocusPaneLeft  []string `json:"focus_pane_left"`
	FocusPaneRight []string `json:"focus_pane_right"`
	FocusPaneUp    []string `json:"focus_pane_up"`
	FocusPaneDown  []string `json:"focus_pane_down"`

	// Multiwindow: pane resize
	GrowPaneH   []string `json:"grow_pane_h"`
	ShrinkPaneH []string `json:"shrink_pane_h"`
	GrowPaneV   []string `json:"grow_pane_v"`
	ShrinkPaneV []string `json:"shrink_pane_v"`

	// Multiwindow: pane resize boundary-absolute (arrow keys — boundary always moves in key direction)
	BoundaryRight []string `json:"boundary_right"`
	BoundaryLeft  []string `json:"boundary_left"`
	BoundaryDown  []string `json:"boundary_down"`
	BoundaryUp    []string `json:"boundary_up"`

	// Multiwindow: sidebar resize
	GrowSidebar   []string `json:"grow_sidebar"`
	ShrinkSidebar []string `json:"shrink_sidebar"`

	// Sidebar and statusbar visibility toggles
	ToggleSidebar   []string `json:"toggle_sidebar"`
	ToggleStatusBar []string `json:"toggle_statusbar"`

	// Terminal copy mode (tmux-style selection + OSC52 clipboard)
	EnterCopyMode []string `json:"enter_copy_mode"`
}

func defaultKeybindings() Keybindings {
	return Keybindings{
		FocusSidebar:     []string{"alt+1", "f1", "ctrl+1"},
		FocusTerminal:    []string{"alt+2", "f2", "ctrl+2"},
		FocusEditor:      []string{"alt+3", "f3", "ctrl+3"},
		OpenTerminalHere: []string{"ctrl+t"},
		Save:             []string{"ctrl+s"},
		Quit:             []string{"ctrl+c"},
		Help:             []string{"?"},

		// alt+| (alt+shift+\) is safe — Windows Terminal does not capture it.
		// alt+_ (alt+shift+-) conflicts with Windows Terminal "split pane down" on WSL —
		// replaced with alt+v which is unambiguous and passes through Windows Terminal.
		SplitHorizontal: []string{"alt+b"},
		SplitVertical:   []string{"alt+v"},
		ClosePane:       []string{"alt+q"},

		FocusPaneLeft:  []string{"alt+h", "alt+left"},
		FocusPaneRight: []string{"alt+l", "alt+right"},
		FocusPaneUp:    []string{"alt+k", "alt+up"},
		FocusPaneDown:  []string{"alt+j", "alt+down"},

		GrowPaneH:   []string{"alt+L"},
		ShrinkPaneH: []string{"alt+H"},
		GrowPaneV:   []string{"alt+J"},
		ShrinkPaneV: []string{"alt+K"},

		// Arrow keys move the split boundary in the key direction (boundary-absolute).
		// Requires unbinding alt+shift+arrow in Windows Terminal settings on WSL.
		BoundaryRight: []string{"alt+shift+right"},
		BoundaryLeft:  []string{"alt+shift+left"},
		BoundaryDown:  []string{"alt+shift+down"},
		BoundaryUp:    []string{"alt+shift+up"},

		GrowSidebar:   []string{"alt+}"},
		ShrinkSidebar: []string{"alt+{"},

		// alt+b conflicts with readline "backward-word" in some terminal emulators.
		// alt+e (Explorer, like VSCode) passes through Windows Terminal safely.
		ToggleSidebar:   []string{"alt+e"},
		ToggleStatusBar: []string{"alt+m"},

		// alt+y enters copy mode (tmux convention is prefix+[, but we have no prefix).
		EnterCopyMode: []string{"alt+y"},
	}
}

// LoadKeybindings reads ~/.config/lumina/keybindings.json, falling back to defaults.
// Individual actions not specified in the file inherit their defaults.
func LoadKeybindings() (Keybindings, error) {
	kb := defaultKeybindings()

	home, err := os.UserHomeDir()
	if err != nil {
		return kb, nil //nolint:nilerr
	}

	path := filepath.Join(home, ".config", "lumina", "keybindings.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		_ = WriteDefaults(path) // best-effort: generate default file on first run
		return kb, nil
	}
	if err != nil {
		return kb, err
	}

	var partial Keybindings
	if err := json.Unmarshal(data, &partial); err != nil {
		return kb, err
	}
	if len(partial.FocusTerminal) > 0    { kb.FocusTerminal    = partial.FocusTerminal    }
	if len(partial.FocusSidebar) > 0     { kb.FocusSidebar     = partial.FocusSidebar     }
	if len(partial.FocusEditor) > 0      { kb.FocusEditor      = partial.FocusEditor      }
	if len(partial.OpenTerminalHere) > 0 { kb.OpenTerminalHere = partial.OpenTerminalHere }
	if len(partial.Save) > 0             { kb.Save             = partial.Save             }
	if len(partial.Quit) > 0             { kb.Quit             = partial.Quit             }
	if len(partial.Help) > 0             { kb.Help             = partial.Help             }
	if len(partial.SplitHorizontal) > 0  { kb.SplitHorizontal  = partial.SplitHorizontal  }
	if len(partial.SplitVertical) > 0    { kb.SplitVertical    = partial.SplitVertical    }
	if len(partial.ClosePane) > 0        { kb.ClosePane        = partial.ClosePane        }
	if len(partial.FocusPaneLeft) > 0    { kb.FocusPaneLeft    = partial.FocusPaneLeft    }
	if len(partial.FocusPaneRight) > 0   { kb.FocusPaneRight   = partial.FocusPaneRight   }
	if len(partial.FocusPaneUp) > 0      { kb.FocusPaneUp      = partial.FocusPaneUp      }
	if len(partial.FocusPaneDown) > 0    { kb.FocusPaneDown    = partial.FocusPaneDown    }
	if len(partial.GrowPaneH) > 0        { kb.GrowPaneH        = partial.GrowPaneH        }
	if len(partial.ShrinkPaneH) > 0      { kb.ShrinkPaneH      = partial.ShrinkPaneH      }
	if len(partial.GrowPaneV) > 0        { kb.GrowPaneV        = partial.GrowPaneV        }
	if len(partial.ShrinkPaneV) > 0      { kb.ShrinkPaneV      = partial.ShrinkPaneV      }
	if len(partial.BoundaryRight) > 0    { kb.BoundaryRight    = partial.BoundaryRight    }
	if len(partial.BoundaryLeft) > 0     { kb.BoundaryLeft     = partial.BoundaryLeft     }
	if len(partial.BoundaryDown) > 0     { kb.BoundaryDown     = partial.BoundaryDown     }
	if len(partial.BoundaryUp) > 0       { kb.BoundaryUp       = partial.BoundaryUp       }
	if len(partial.GrowSidebar) > 0      { kb.GrowSidebar      = partial.GrowSidebar      }
	if len(partial.ShrinkSidebar) > 0    { kb.ShrinkSidebar    = partial.ShrinkSidebar    }
	if len(partial.ToggleSidebar) > 0    { kb.ToggleSidebar    = partial.ToggleSidebar    }
	if len(partial.ToggleStatusBar) > 0  { kb.ToggleStatusBar  = partial.ToggleStatusBar  }
	if len(partial.EnterCopyMode) > 0    { kb.EnterCopyMode    = partial.EnterCopyMode    }

	return kb, nil
}

// Action returns the action name for a given key string, or "" if not bound.
func (kb Keybindings) Action(key string) string {
	for _, k := range kb.FocusTerminal    { if k == key { return "focus_terminal"      } }
	for _, k := range kb.FocusSidebar     { if k == key { return "focus_sidebar"       } }
	for _, k := range kb.FocusEditor      { if k == key { return "focus_editor"        } }
	for _, k := range kb.OpenTerminalHere { if k == key { return "open_terminal_here"  } }
	for _, k := range kb.Save             { if k == key { return "save"                } }
	for _, k := range kb.Quit             { if k == key { return "quit"                } }
	for _, k := range kb.Help             { if k == key { return "help"                } }
	for _, k := range kb.SplitHorizontal  { if k == key { return "split_horizontal"    } }
	for _, k := range kb.SplitVertical    { if k == key { return "split_vertical"      } }
	for _, k := range kb.ClosePane        { if k == key { return "close_pane"          } }
	for _, k := range kb.FocusPaneLeft    { if k == key { return "focus_pane_left"     } }
	for _, k := range kb.FocusPaneRight   { if k == key { return "focus_pane_right"    } }
	for _, k := range kb.FocusPaneUp      { if k == key { return "focus_pane_up"       } }
	for _, k := range kb.FocusPaneDown    { if k == key { return "focus_pane_down"     } }
	for _, k := range kb.GrowPaneH        { if k == key { return "grow_pane_h"         } }
	for _, k := range kb.ShrinkPaneH      { if k == key { return "shrink_pane_h"       } }
	for _, k := range kb.GrowPaneV        { if k == key { return "grow_pane_v"         } }
	for _, k := range kb.ShrinkPaneV      { if k == key { return "shrink_pane_v"       } }
	for _, k := range kb.BoundaryRight    { if k == key { return "boundary_right"      } }
	for _, k := range kb.BoundaryLeft     { if k == key { return "boundary_left"       } }
	for _, k := range kb.BoundaryDown     { if k == key { return "boundary_down"       } }
	for _, k := range kb.BoundaryUp       { if k == key { return "boundary_up"         } }
	for _, k := range kb.GrowSidebar      { if k == key { return "grow_sidebar"        } }
	for _, k := range kb.ShrinkSidebar    { if k == key { return "shrink_sidebar"      } }
	for _, k := range kb.ToggleSidebar    { if k == key { return "toggle_sidebar"      } }
	for _, k := range kb.ToggleStatusBar  { if k == key { return "toggle_statusbar"    } }
	for _, k := range kb.EnterCopyMode    { if k == key { return "enter_copy_mode"     } }
	return ""
}

// GlobalKeys returns all keys that must not be forwarded to the PTY.
func (kb Keybindings) GlobalKeys() map[string]bool {
	reserved := make(map[string]bool)
	for _, k := range kb.FocusTerminal    { reserved[k] = true }
	for _, k := range kb.FocusSidebar     { reserved[k] = true }
	for _, k := range kb.FocusEditor      { reserved[k] = true }
	for _, k := range kb.OpenTerminalHere { reserved[k] = true }
	for _, k := range kb.SplitHorizontal  { reserved[k] = true }
	for _, k := range kb.SplitVertical    { reserved[k] = true }
	for _, k := range kb.ClosePane        { reserved[k] = true }
	for _, k := range kb.FocusPaneLeft    { reserved[k] = true }
	for _, k := range kb.FocusPaneRight   { reserved[k] = true }
	for _, k := range kb.FocusPaneUp      { reserved[k] = true }
	for _, k := range kb.FocusPaneDown    { reserved[k] = true }
	for _, k := range kb.GrowPaneH        { reserved[k] = true }
	for _, k := range kb.ShrinkPaneH      { reserved[k] = true }
	for _, k := range kb.GrowPaneV        { reserved[k] = true }
	for _, k := range kb.ShrinkPaneV      { reserved[k] = true }
	for _, k := range kb.BoundaryRight    { reserved[k] = true }
	for _, k := range kb.BoundaryLeft     { reserved[k] = true }
	for _, k := range kb.BoundaryDown     { reserved[k] = true }
	for _, k := range kb.BoundaryUp       { reserved[k] = true }
	for _, k := range kb.GrowSidebar      { reserved[k] = true }
	for _, k := range kb.ShrinkSidebar    { reserved[k] = true }
	for _, k := range kb.ToggleSidebar    { reserved[k] = true }
	for _, k := range kb.ToggleStatusBar  { reserved[k] = true }
	for _, k := range kb.EnterCopyMode    { reserved[k] = true }
	return reserved
}

// WriteDefaults writes the default keybindings.json to the given path.
func WriteDefaults(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(defaultKeybindings(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
