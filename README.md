# Lumina

> **"We have Hyprland at home"** — the Hyprland at home:

Lumina is an open-source VSCode-inspired TUI (Terminal User Interface) editor built with Go and Bubble Tea. It runs entirely inside your terminal, combining a real interactive shell, a file explorer, a text editor, and a live system monitor — all in one keyboard-driven workspace.

The goal is simple: give developers a productive, lightweight editing environment without ever leaving the terminal, inspired by the tiling and multi-window philosophy of compositors like Hyprland, but running anywhere a terminal runs.

---

## Features

- **Interactive terminal** — real PTY sessions using your system's default shell (`$SHELL`)
- **Multi-window layout** — split panes horizontally or vertically (up to 4 panels), inspired by tiling window managers
- **File explorer sidebar** — per-window sidebar with keyboard navigation; toggle it on/off with a single keybind
- **Text editor** — open, edit, and save files directly in Lumina, with cursor navigation and unsaved-changes protection
- **System resource monitor** — live CPU, memory, and git branch info in a global status bar; hideable when you need more space
- **Fully keyboard-driven** — every action has a keybinding; no mouse required (mouse resize optional)
- **Configurable** — keybindings and shell preference via `~/.config/lumina/config.toml`

---

## Installation

**Requirements**: Go 1.26+, Linux or macOS (PTY not supported on Windows)

```bash
git clone https://github.com/your-org/lumina.git
cd lumina
go build -o lumina .
./lumina
```

Open a specific file:

```bash
./lumina path/to/file.txt
```

---

## Default Keybindings

| Action | Key |
|---|---|
| Switch focus between panes | `Ctrl+hjkl` |
| Split pane horizontally | `Ctrl+\` |
| Split pane vertically | `Ctrl+-` |
| Close active pane | `Ctrl+w` |
| Toggle sidebar | `Ctrl+b` |
| Toggle resource monitor | `Ctrl+m` |
| Save file | `Ctrl+s` |
| Quit | `Ctrl+q` |

All keybindings are defined in `app/keymap.go` and overridable via `config.toml`.

---

## Architecture

Lumina follows the Elm architecture (Model/Update/View) via Bubble Tea. Each panel is an independent `tea.Model` composed by the root `app.Model` through delegation and typed messages.

```
lumina/
├── main.go
├── app/            # Root model — routes messages between components
├── components/
│   ├── terminal/   # PTY wrapper (creack/pty)
│   ├── sidebar/    # File explorer
│   ├── editor/     # Text buffer
│   └── statusbar/  # Live system metrics
├── msgs/           # All cross-component tea.Msg types
└── config/         # Config struct + loader
```

---

## Development

```bash
# Run tests
go test ./...

# Lint
golangci-lint run

# Build
go build -o lumina .
```

---