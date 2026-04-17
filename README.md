# Lumina

> **"We have Hyprland at home"** — the Hyprland at home.

Lumina is an open-source VSCode-inspired TUI (Terminal User Interface) editor built with Go and Bubble Tea. It runs entirely inside your terminal, combining a real interactive shell, a file explorer, a text editor, and a live system monitor — all in one keyboard-driven workspace.

The goal is simple: give developers a productive, lightweight editing environment without ever leaving the terminal, inspired by the tiling and multi-window philosophy of compositors like Hyprland, but running anywhere a terminal runs.

> **Built on [Speckkit](https://github.com/github/spec-kit)** — Lumina uses Speckkit as its spec-driven development foundation. Features are designed via structured specs (stored in `specs/`) that drive architecture decisions, contracts, and implementation tasks before any code is written.

---

## Features

- **Interactive terminal** — real PTY sessions using your system's default shell (`$SHELL`)
- **Multi-window layout** — split panes horizontally or vertically (up to 4 panels), inspired by tiling window managers
- **File explorer sidebar** — per-window sidebar with keyboard navigation; toggle it on/off with a single keybind
- **Text editor** — open, edit, and save files directly in Lumina, with cursor navigation and unsaved-changes protection
- **System resource monitor** — live CPU, memory, and git branch info in a global status bar; hideable when you need more space
- **Fully keyboard-driven** — every action has a keybinding; no mouse required
- **Auto-generated config** — `~/.config/lumina/config.toml` and `keybindings.json` are created on first run with sensible defaults

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

## Configuration

On first launch, Lumina creates two files under `~/.config/lumina/`:

| File | Purpose |
|---|---|
| `config.toml` | General settings (shell, theme, sidebar width, etc.) |
| `keybindings.json` | Key bindings for every action |

### config.toml

```toml
shell            = "/bin/zsh"   # Shell executable for PTY sessions. Defaults to $SHELL.
metrics_interval = 1000         # Status bar refresh rate in milliseconds.
show_hidden      = true         # Show hidden files (dotfiles) in the sidebar.
sidebar_width    = 30           # Sidebar width in terminal columns.
theme            = "default"    # UI colour theme.
force_shell_theme = true        # Inject Lumina's custom prompt into spawned shells.
```

On **WSL**, if `shell` points to a Windows executable (`.exe`), Lumina rejects it automatically and falls back to the first available POSIX shell, displaying a warning in the status bar.

### keybindings.json

Each action maps to a list of key strings — any of them triggers the action. Key notation follows Bubble Tea: `"ctrl+s"`, `"alt+h"`, `"f1"`, `"?"`, etc.

```json
{
  "toggle_sidebar": ["alt+e"],
  "split_horizontal": ["alt+b", "alt+|"]
}
```

Only include the actions you want to override; everything else inherits its default.

---

## Keybindings

### Focus

| Action | Default keys | Description |
|---|---|---|
| Focus sidebar | `alt+1` / `f1` / `ctrl+1` | Move keyboard focus to the file explorer sidebar |
| Focus terminal | `alt+2` / `f2` / `ctrl+2` | Move keyboard focus to the terminal pane |
| Focus editor | `alt+3` / `f3` / `ctrl+3` | Move keyboard focus to the text editor pane |
| Open terminal here | `ctrl+t` | Open a new terminal in the current working directory |

### Pane management

| Action | Default keys | Description |
|---|---|---|
| Split horizontal | `alt+b` | Split the active pane side-by-side (left / right) |
| Split vertical | `alt+v` | Split the active pane top-and-bottom (up / down) |
| Close pane | `alt+q` | Close the active pane; its sibling expands to fill the space |

### Pane navigation

| Action | Default keys | Description |
|---|---|---|
| Focus pane left | `alt+h` / `alt+←` | Move focus to the nearest pane to the left |
| Focus pane right | `alt+l` / `alt+→` | Move focus to the nearest pane to the right |
| Focus pane up | `alt+k` / `alt+↑` | Move focus to the nearest pane above |
| Focus pane down | `alt+j` / `alt+↓` | Move focus to the nearest pane below |

### Pane resize — relative

These resize relative to the **active pane** (the boundary adjacent to it moves).

| Action | Default keys | Description |
|---|---|---|
| Grow pane right | `alt+L` | Widen the active pane by moving its right boundary outward |
| Shrink pane left | `alt+H` | Narrow the active pane by moving its right boundary inward |
| Grow pane down | `alt+J` | Tighten the active pane by moving its bottom boundary downward |
| Shrink pane up | `alt+K` | Shrink the active pane by moving its bottom boundary upward |

### Pane resize — boundary-absolute

Arrow keys move the **split boundary** in the direction of the key, regardless of which pane is active.

| Action | Default keys | Description |
|---|---|---|
| Boundary right | `alt+shift+→` | Push the nearest vertical split boundary rightward |
| Boundary left | `alt+shift+←` | Push the nearest vertical split boundary leftward |
| Boundary down | `alt+shift+↓` | Push the nearest horizontal split boundary downward |
| Boundary up | `alt+shift+↑` | Push the nearest horizontal split boundary upward |

> **WSL note**: `alt+shift+arrow` may be captured by Windows Terminal ("move pane"). Unbind those shortcuts in Windows Terminal settings to pass them through.

### Sidebar resize

| Action | Default keys | Description |
|---|---|---|
| Grow sidebar | `alt+}` | Increase the sidebar width by one column |
| Shrink sidebar | `alt+{` | Decrease the sidebar width by one column |

### Toggles

| Action | Default keys | Description |
|---|---|---|
| Toggle sidebar | `alt+e` | Show or hide the file explorer sidebar for the active pane |
| Toggle status bar | `alt+m` | Show or hide the bottom system-metrics bar |

### File & app

| Action | Default keys | Description |
|---|---|---|
| Save file | `ctrl+s` | Save the file open in the active editor pane |
| Quit | `ctrl+c` | Exit Lumina (prompts if there are unsaved changes) |
| Help | `?` | Toggle the keybinding help overlay |

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
