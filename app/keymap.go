package app

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/menegas/lumina/config"
)

// KeyMap centralises all keybindings for Lumina.
// Built from config.Keybindings — do not hardcode keys here.
type KeyMap struct {
	FocusTerminal    key.Binding
	FocusSidebar     key.Binding
	OpenTerminalHere key.Binding
	Save             key.Binding
	Quit             key.Binding
	Help             key.Binding

	// Multiwindow
	SplitHorizontal key.Binding
	SplitVertical   key.Binding
	ClosePane       key.Binding
	FocusPaneLeft   key.Binding
	FocusPaneRight  key.Binding
	FocusPaneUp     key.Binding
	FocusPaneDown   key.Binding
	GrowPaneH       key.Binding
	ShrinkPaneH     key.Binding
	GrowPaneV       key.Binding
	ShrinkPaneV     key.Binding
	BoundaryRight   key.Binding
	BoundaryLeft    key.Binding
	BoundaryDown    key.Binding
	BoundaryUp      key.Binding
	GrowSidebar     key.Binding
	ShrinkSidebar   key.Binding
	ToggleSidebar   key.Binding
	ToggleStatusBar key.Binding
	EnterCopyMode   key.Binding

	// Sidebar file-manager (feature 006)
	SidebarNewDir  key.Binding
	SidebarNewFile key.Binding
	SidebarParent  key.Binding
}

// NewKeyMap builds a KeyMap from the user's Keybindings config.
func NewKeyMap(kb config.Keybindings) KeyMap {
	return KeyMap{
		FocusTerminal: key.NewBinding(
			key.WithKeys(kb.FocusTerminal...),
			key.WithHelp(join(kb.FocusTerminal), "focus terminal"),
		),
		FocusSidebar: key.NewBinding(
			key.WithKeys(kb.FocusSidebar...),
			key.WithHelp(join(kb.FocusSidebar), "focus sidebar"),
		),
		OpenTerminalHere: key.NewBinding(
			key.WithKeys(kb.OpenTerminalHere...),
			key.WithHelp(join(kb.OpenTerminalHere), "open terminal here"),
		),
		Save: key.NewBinding(
			key.WithKeys(kb.Save...),
			key.WithHelp(join(kb.Save), "save file"),
		),
		Quit: key.NewBinding(
			key.WithKeys(kb.Quit...),
			key.WithHelp(join(kb.Quit), "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys(kb.Help...),
			key.WithHelp(join(kb.Help), "show help"),
		),
		SplitHorizontal: key.NewBinding(
			key.WithKeys(kb.SplitHorizontal...),
			key.WithHelp(join(kb.SplitHorizontal), "split horizontal"),
		),
		SplitVertical: key.NewBinding(
			key.WithKeys(kb.SplitVertical...),
			key.WithHelp(join(kb.SplitVertical), "split vertical"),
		),
		ClosePane: key.NewBinding(
			key.WithKeys(kb.ClosePane...),
			key.WithHelp(join(kb.ClosePane), "close pane"),
		),
		FocusPaneLeft: key.NewBinding(
			key.WithKeys(kb.FocusPaneLeft...),
			key.WithHelp(join(kb.FocusPaneLeft), "focus pane left"),
		),
		FocusPaneRight: key.NewBinding(
			key.WithKeys(kb.FocusPaneRight...),
			key.WithHelp(join(kb.FocusPaneRight), "focus pane right"),
		),
		FocusPaneUp: key.NewBinding(
			key.WithKeys(kb.FocusPaneUp...),
			key.WithHelp(join(kb.FocusPaneUp), "focus pane up"),
		),
		FocusPaneDown: key.NewBinding(
			key.WithKeys(kb.FocusPaneDown...),
			key.WithHelp(join(kb.FocusPaneDown), "focus pane down"),
		),
		GrowPaneH: key.NewBinding(
			key.WithKeys(kb.GrowPaneH...),
			key.WithHelp(join(kb.GrowPaneH), "grow pane →"),
		),
		ShrinkPaneH: key.NewBinding(
			key.WithKeys(kb.ShrinkPaneH...),
			key.WithHelp(join(kb.ShrinkPaneH), "shrink pane ←"),
		),
		GrowPaneV: key.NewBinding(
			key.WithKeys(kb.GrowPaneV...),
			key.WithHelp(join(kb.GrowPaneV), "grow pane ↓"),
		),
		ShrinkPaneV: key.NewBinding(
			key.WithKeys(kb.ShrinkPaneV...),
			key.WithHelp(join(kb.ShrinkPaneV), "shrink pane ↑"),
		),
		BoundaryRight: key.NewBinding(
			key.WithKeys(kb.BoundaryRight...),
			key.WithHelp(join(kb.BoundaryRight), "boundary →"),
		),
		BoundaryLeft: key.NewBinding(
			key.WithKeys(kb.BoundaryLeft...),
			key.WithHelp(join(kb.BoundaryLeft), "boundary ←"),
		),
		BoundaryDown: key.NewBinding(
			key.WithKeys(kb.BoundaryDown...),
			key.WithHelp(join(kb.BoundaryDown), "boundary ↓"),
		),
		BoundaryUp: key.NewBinding(
			key.WithKeys(kb.BoundaryUp...),
			key.WithHelp(join(kb.BoundaryUp), "boundary ↑"),
		),
		GrowSidebar: key.NewBinding(
			key.WithKeys(kb.GrowSidebar...),
			key.WithHelp(join(kb.GrowSidebar), "grow sidebar"),
		),
		ShrinkSidebar: key.NewBinding(
			key.WithKeys(kb.ShrinkSidebar...),
			key.WithHelp(join(kb.ShrinkSidebar), "shrink sidebar"),
		),
		ToggleSidebar: key.NewBinding(
			key.WithKeys(kb.ToggleSidebar...),
			key.WithHelp(join(kb.ToggleSidebar), "toggle sidebar"),
		),
		ToggleStatusBar: key.NewBinding(
			key.WithKeys(kb.ToggleStatusBar...),
			key.WithHelp(join(kb.ToggleStatusBar), "toggle monitor"),
		),
		EnterCopyMode: key.NewBinding(
			key.WithKeys(kb.EnterCopyMode...),
			key.WithHelp(join(kb.EnterCopyMode), "copy mode"),
		),
		SidebarNewDir: key.NewBinding(
			key.WithKeys(kb.SidebarNewDir...),
			key.WithHelp(join(kb.SidebarNewDir), "new dir (sidebar)"),
		),
		SidebarNewFile: key.NewBinding(
			key.WithKeys(kb.SidebarNewFile...),
			key.WithHelp(join(kb.SidebarNewFile), "new file (sidebar)"),
		),
		SidebarParent: key.NewBinding(
			key.WithKeys(kb.SidebarParent...),
			key.WithHelp(join(kb.SidebarParent), "parent dir (sidebar)"),
		),
	}
}

func join(keys []string) string {
	if len(keys) == 0 {
		return ""
	}
	result := keys[0]
	for _, k := range keys[1:] {
		result += " / " + k
	}
	return result
}

// ShortHelp returns bindings shown in the compact help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.SplitHorizontal, k.SplitVertical, k.ClosePane,
		k.FocusPaneLeft, k.FocusPaneRight,
		k.FocusSidebar, k.ToggleSidebar, k.ToggleStatusBar, k.Quit,
	}
}

// FullHelp returns all bindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.FocusTerminal, k.FocusSidebar, k.OpenTerminalHere},
		{k.SplitHorizontal, k.SplitVertical, k.ClosePane},
		{k.FocusPaneLeft, k.FocusPaneRight, k.FocusPaneUp, k.FocusPaneDown},
		{k.GrowPaneH, k.ShrinkPaneH, k.GrowSidebar, k.ShrinkSidebar},
		{k.ToggleSidebar, k.ToggleStatusBar, k.EnterCopyMode},
		{k.SidebarNewDir, k.SidebarNewFile, k.SidebarParent},
		{k.Save, k.Quit, k.Help},
	}
}
