# Lumina

> **"We have Hyprland at home"** — the Hyprland at home.

## Quick install

```bash
curl -fsSL https://raw.githubusercontent.com/Felipe-Meneguzzi/lumina/main/install.sh | bash
```

Detects OS/architecture (Linux/macOS, amd64/arm64), downloads the latest release binary,
and installs it to `~/.local/bin`. Then just run `lumina` — on Windows use **WSL 2**.

### Update

If Lumina is already installed, the easiest way is to use the binary itself:

```bash
lumina --update
```

Fetches the latest GitHub release, compares it with the installed version, and replaces
the binary in-place if a newer version is available. Does nothing if already up to date.

Alternatively, the same installer line also updates:

```bash
curl -fsSL https://raw.githubusercontent.com/Felipe-Meneguzzi/lumina/main/install.sh | bash
```

To check the currently running version:

```bash
lumina --version
```

Advanced options, pinning a version, building from source: see [Installation](#installation) below.

---

Lumina is a TUI (Terminal User Interface) workspace inspired by [Hyprland](https://hyprland.org/),
written in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

**The primary target is WSL**: users who live in `wsl` inside Windows Terminal and want the
ergonomics of a tiling window manager (tile panes, chord shortcuts, keyboard focus,
fluid resizing) without depending on X11/Wayland, without leaving the terminal, and
without losing integration with the distro's native shell. It also runs on any Linux/macOS
with a modern terminal — but keybinding choices, shell detection, and default behaviour
are tuned for the WSL + Windows Terminal case.

Inside a single Lumina instance you get:

- multiple real terminals (PTY), in recursive side-by-side / stacked tiles
- a native built-in text editor, with save and unsaved-changes protection
- a per-pane file explorer, resizable by keyboard or mouse
- a system monitor (CPU, memory, git branch, CWD) in the footer
- tmux-style copy mode with copy to host clipboard via OSC 52
- mouse passthrough to apps inside the terminal (vim, htop, lazygit…)

> **Built on [Speckkit](https://github.com/github/spec-kit)** — Lumina uses Speckkit as a
> spec-driven development base. Features are designed via structured specs (in `specs/`)
> that drive architecture decisions, contracts, and implementation tasks before any code
> is written.

---

## Why "Hyprland for WSL"?

WSL delivers an excellent CLI Linux experience, but loses the entire graphical window
manager layer. People who live in the terminal typically glue together tmux + vim +
lazygit + htop in a mosaic of Windows Terminal windows, which works but has friction:

- each terminal is a decoupled session — no native tiling, no consistent copy-mode
  across them, no unified metrics;
- resizing a pane requires the mouse or Windows Terminal's own command sequence;
- graphical WM shortcuts (Hyprland `SUPER+arrow`, `SUPER+v`, etc.) don't exist.

Lumina mimics the experience of a tiling compositor inside a single terminal emulator:
`alt+b` / `alt+v` split, `alt+hjkl` move focus, `alt+HJKL` resize the focused pane,
`alt+shift+←→↑↓` move the border between panes. The binary split tree follows the mental
model of anyone already using Hyprland, i3, or sway.

---

## Features

### Windows and layout
- **Binary split tree** (Hyprland-inspired): recursive horizontal and vertical splits,
  up to 4 simultaneous panes
- **Spatial keyboard focus** (`alt+hjkl` or `alt+arrows`) — the neighbouring pane in the
  arrow direction receives focus, respecting real geometry
- **Focus-relative resize** (`alt+HJKL`) and **absolute border resize** (`alt+shift+arrows`)
- **Per-pane sidebar** — each pane can have its own file explorer, visible/hidden and with
  an independent width

### Terminal
- **Real PTY** using the user's `$SHELL` (zsh / bash / fish) via `creack/pty`
- **Full VT emulator** (`charmbracelet/x/vt`) with 24-bit colour, styles, and DEC modes
- **2000-line scrollback**, navigable with `PgUp`/`PgDown` or `Alt+Wheel`
- **tmux-style copy mode** (`alt+y`): Vim-like cursor (hjkl + `v` + `y`), rectangular
  selection with visual highlight, copy to host clipboard via **OSC 52** — works even
  over SSH/WSL because the sequence passes through the host terminal
- **Mouse passthrough**: when the app inside the terminal enables mouse tracking (DEC modes
  1000/1002/1003, used by vim, htop, tmux, lazygit), Lumina forwards events with
  coordinates translated to the pane interior
- **OSC 7 / OSC 0/2**: captures CWD and title reported by the shell; reused by
  `Open terminal here` and the status bar
- **Auto-restart** of the shell on exit (without tearing down the Lumina session)
- **Optional forced theme** (`force_shell_theme`): injects an oh-my-zsh-inspired prompt
  to normalise shells that lack their own configuration

### Editor
- Opens files in the **configured external editor** (`nano` by default, configurable via
  the `editor` field in `config.toml`)
- The sidebar launches the external editor when opening a file; `ctrl+s` works inside the
  external editor normally

### Mouse
- **Click-to-focus** on any pane (sidebar, editor, terminal)
- **Drag** on the sidebar border to resize
- **Mouse text selection** (drag) in the terminal — dragging selects text with visual
  highlight; on button release the text is automatically copied to the host clipboard via
  OSC 52 if `mouse_auto_copy = true` (default). With `mouse_auto_copy = false` a
  confirmation appears in the status bar.
- `selection_mode` controls the selection style: `"linear"` (default, notepad-style) or
  `"block"` (rectangular, vim visual-block style)
- **Alt+wheel** for terminal scrollback (hotkey preserved even when the app requests mouse
  tracking, serving as an escape hatch)
- **Wheel without Alt** passes directly to the app inside the terminal when it is in mouse
  mode

### Status bar
- CPU (%), used/total memory, git branch, CWD of the focused pane
- Title (OSC 0/2) reported by the internal app of the focused terminal
- Temporary notifications (save, copy, warnings)
- Hideable with `alt+m`

---

## Installation

**Requirements**: Linux or macOS. On Windows, use **WSL 2** with Ubuntu / Debian / Fedora —
native Windows PTY is not supported.

### Option 1 — one-liner (recommended)

Downloads the latest release binary and installs it to `~/.local/bin` (or
`/usr/local/bin`, if available):

```bash
curl -fsSL https://raw.githubusercontent.com/Felipe-Meneguzzi/lumina/main/install.sh | bash
```

Optional environment variables:

| Variable | Default | Purpose |
|---|---|---|
| `LUMINA_VERSION` | `latest` | Tag to install (e.g. `v0.3.1`) |
| `INSTALL_DIR`    | `~/.local/bin` | Destination directory |
| `LUMINA_REPO`    | `Felipe-Meneguzzi/lumina` | Fork override |

Example pinning version and directory:

```bash
LUMINA_VERSION=v0.3.1 INSTALL_DIR=/usr/local/bin \
  curl -fsSL https://raw.githubusercontent.com/Felipe-Meneguzzi/lumina/main/install.sh | bash
```

The script detects OS (`linux` / `darwin`) and architecture (`amd64` / `arm64`), validates
the SHA256 checksum (if the release publishes `checksums.txt`), and warns if the
installation directory is not in `PATH`.

> **Publishing releases**: the installer expects assets named
> `lumina-<os>-<arch>` (e.g. `lumina-linux-amd64`) attached to the GitHub release.
> Optionally a `checksums.txt` with lines in the format `sha256  lumina-linux-amd64`.

### Option 2 — build from source

Requires Go 1.26+.

```bash
git clone https://github.com/Felipe-Meneguzzi/lumina.git
cd lumina
go build -o lumina .
./lumina
```

Open a specific file:

```bash
lumina path/to/file.txt
```

### CLI flags

| Flag | Description |
|------|-------------|
| `--update` | Checks for a new GitHub release and installs it if available. |
| `--version`, `-v` | Prints the installed version and exits. |
| `--help`, `-h` | Shows the full help and exits. |

Session flags (ephemeral — do not modify `config.toml`):

| Flag | Format | Default | Description |
|------|--------|---------|-------------|
| `-mp` | `-mp N` | 4 | Maximum number of panes allowed in the session. |
| `-sp` | `-sp h<N>` / `-sp v<N>` | 1 pane | Creates `N` initial panes laid out horizontally (`h`) or vertically (`v`). |
| `-sc` | `-sc "<command>"` | default shell | Runs `<command>` in panes created by `-sp` (initial panes only — later manual splits open the default shell). |

Examples:

```bash
lumina                                  # default boot: 1 pane, ceiling 4
lumina -mp 10                           # ceiling 10, 1 initial pane
lumina -sp h3                           # 3 side-by-side panes
lumina -sp v2 -sc claude                # 2 stacked panes running claude
lumina -mp 10 -sp h3 -sc claude         # full combination
lumina notes.md -sp h2                  # file + custom layout
```

Validation rules:

- `-mp < 1`, `-sp` outside the `h<N>`/`v<N>` format, or `-sc ""` abort startup with a
  message to stderr (exit code 2).
- If an explicit `-mp` is lower than the `N` of `-sp`, startup is aborted.
- If `-mp` is omitted and `-sp hN` / `-sp vN` exceeds the default (4), the effective
  ceiling is automatically raised to `N`.

See `lumina --help` for the full help message.

### Tip for WSL + Windows Terminal

Some default Windows Terminal shortcuts (e.g. `alt+shift+arrow` to move a pane) are
captured before reaching Lumina. Unbind them in *Settings → Actions* in Windows Terminal
to allow passthrough.

---

## Configuration

On first run, Lumina creates two files in `~/.config/lumina/`:

| File | Purpose |
|---|---|
| `config.toml` | General settings (shell, theme, metrics, sidebar) |
| `keybindings.json` | Key mapping for each action |

### config.toml

```toml
shell             = "/bin/zsh"   # Shell executable for PTYs. Default: $SHELL.
metrics_interval  = 1000         # Status bar refresh rate in ms.
show_hidden       = true         # Show dotfiles in the sidebar.
sidebar_width     = 30           # Sidebar width in columns.
theme             = "default"    # UI theme.
force_shell_theme = true         # Inject Lumina's custom prompt into the shell.
mouse_auto_copy   = true         # Auto-copy to clipboard on mouse selection release.
selection_mode    = "linear"     # Selection style: "linear" (default) or "block" (rectangular).
editor            = "nano"       # External editor used by the sidebar ("nano"|"vim"|"nvim"|absolute path).
```

In **WSL**, if `shell` points to a Windows executable (`.exe`), Lumina automatically
rejects it and falls back to the first available POSIX shell, with a warning in the
status bar.

### keybindings.json

Each action maps to a list of keys — any of them triggers the action. The notation
follows Bubble Tea's: `"ctrl+s"`, `"alt+h"`, `"f1"`, `"?"`.

```json
{
  "toggle_sidebar":   ["alt+e"],
  "split_horizontal": ["alt+b", "alt+|"],
  "enter_copy_mode":  ["alt+y"],
  "sidebar_new_dir":  ["alt+d"],
  "sidebar_new_file": ["alt+f"],
  "sidebar_parent":   ["backspace"]
}
```

Only include the actions you want to override; the rest inherit their defaults.

---

## Keybindings

### Focus

| Action | Default key | Description |
|---|---|---|
| Focus sidebar | `alt+1` / `f1` / `ctrl+1` | Move focus to the file explorer |
| Focus terminal | `alt+2` / `f2` / `ctrl+2` | Move focus to the terminal |
| Focus editor | `alt+3` / `f3` / `ctrl+3` | Move focus to the editor |
| Open terminal here | `ctrl+t` | New terminal at the active pane's CWD |

### Pane management

| Action | Default key | Description |
|---|---|---|
| Split horizontal | `alt+b` | Split the active pane side by side |
| Split vertical | `alt+v` | Split the active pane stacked |
| Close pane | `alt+q` | Close the pane; the sibling expands |

### Pane navigation

| Action | Default key | Description |
|---|---|---|
| Focus left | `alt+h` / `alt+←` | Move focus to the pane on the left |
| Focus right | `alt+l` / `alt+→` | Move focus to the pane on the right |
| Focus up | `alt+k` / `alt+↑` | Move focus to the pane above |
| Focus down | `alt+j` / `alt+↓` | Move focus to the pane below |

### Resize — relative to the focused pane

Moves the border adjacent to the active pane.

| Action | Default key | Description |
|---|---|---|
| Grow right | `alt+L` | Expands the pane by pushing the right border |
| Shrink left | `alt+H` | Narrows the pane by pulling the right border |
| Grow down | `alt+J` | Expands vertically by pushing the bottom border |
| Shrink up | `alt+K` | Shrinks vertically by pulling the bottom border |

### Resize — absolute border

Arrows move the nearest border in the key direction, regardless of focus.

| Action | Default key | Description |
|---|---|---|
| Border → | `alt+shift+→` | Push the vertical divider to the right |
| Border ← | `alt+shift+←` | Push the vertical divider to the left |
| Border ↓ | `alt+shift+↓` | Push the horizontal divider down |
| Border ↑ | `alt+shift+↑` | Push the horizontal divider up |

> **WSL**: `alt+shift+arrow` may be captured by Windows Terminal ("move pane").
> Unbind it in Windows Terminal settings to allow passthrough.

### Sidebar

| Action | Default key | Description |
|---|---|---|
| Grow sidebar | `alt+}` | +1 column |
| Shrink sidebar | `alt+{` | −1 column |
| Toggle sidebar | `alt+e` | Show/hide the active pane's sidebar |
| Navigate to parent | `backspace` | Go up to the parent directory (sidebar must be focused) |
| New directory | `alt+d` | Create a new directory at the current location |
| New file | `alt+f` | Create a new file at the current location |

### Copy mode (terminal)

Enters a tmux-style mode to select text with the keyboard and copy to the host clipboard
via OSC 52.

| Action | Default key | Description |
|---|---|---|
| Enter copy mode | `alt+y` | Start selection at the bottom-right of the pane |
| Move cursor | `h` `j` `k` `l` or arrows | Vim-like movement |
| Extend selection | `H` `J` `K` `L` / `shift+arrows` | Anchor fixed, cursor moves |
| Toggle anchor | `v` | Reset anchor to cursor position |
| Start / end of line | `0` / `$` | `home` / `end` |
| Top / bottom | `g` / `G` | — |
| Copy and exit | `y` / `enter` | Sends OSC 52; shows confirmation in status bar |
| Cancel | `esc` / `q` / `ctrl+c` | Exit without copying |

While in copy mode the pane border turns **yellow** and all keyboard input is consumed —
the shell receives nothing until the mode ends. The viewport freezes: new shell output is
preserved in scrollback so you don't lose the selected content.

### Terminal scrollback

| Action | Default key | Description |
|---|---|---|
| Scroll up | `PgUp` | 10 lines |
| Scroll down | `PgDown` | 10 lines |
| Scroll up 3 lines | `alt+wheel up` | Scroll into history |
| Scroll down 3 lines | `alt+wheel down` | Scroll toward live output |
| Return to live | Any key | Exit scroll mode |

In apps using alt-screen (vim, less, htop) scrollback is disabled — this is the correct
behaviour because alt-screen never feeds the history buffer.

### File and application

| Action | Default key | Description |
|---|---|---|
| Save file | `ctrl+s` | Save the file open in the active editor |
| Quit | `ctrl+c` | Exit Lumina (prompts for confirmation if there are unsaved changes) |
| Help | `?` | Open the keyboard shortcuts overlay |
| Toggle status bar | `alt+m` | Show/hide the metrics bar |

---

## Architecture

Lumina follows the Elm architecture (Model / Update / View) via Bubble Tea. Each pane is
an independent `tea.Model`, composed by the root `app.Model` through delegation and typed
messages — **no circular imports, no mutable global state**.

```
lumina/
├── main.go
├── app/
│   ├── app.go             # Root model — routes messages between components
│   └── keymap.go          # Single source of truth for all key bindings
├── components/
│   ├── layout/            # Binary split tree (Hyprland-inspired)
│   │   ├── layout.go      # Model, Update, View for the pane manager
│   │   ├── tree.go        # Recursive insert, remove, and walk
│   │   ├── focus.go       # Spatial neighbour search by direction
│   │   ├── bounds.go      # Rectangle calculation for each pane
│   │   └── render.go      # Final string composition with borders
│   ├── terminal/
│   │   ├── terminal.go    # Main model + PTY lifecycle
│   │   ├── scrollback.go  # Composite render (scrollback + live)
│   │   ├── copymode.go    # Copy mode state + render + OSC 52
│   │   ├── mouseselect.go # Mouse text selection with highlight + OSC 52
│   │   ├── mouse.go       # Emulator callbacks (DEC modes, title, CWD, bell)
│   │   ├── keys.go        # tea.KeyMsg → PTY byte translation
│   │   └── theme.go       # Optional custom prompt injection
│   ├── sidebar/           # File explorer (bubbles/list + os.ReadDir) + file/dir creation
│   └── statusbar/         # Metrics (gopsutil ticker)
├── msgs/
│   └── msgs.go            # ALL custom tea.Msg types
├── config/
│   ├── config.go
│   └── keybindings.go
├── specs/                 # Spec-kit: specs and contracts for each feature
└── tests/integration/     # Cross-component message flow tests
```

### Library stack

| Layer | Lib | Use |
|---|---|---|
| TUI framework | [bubbletea](https://github.com/charmbracelet/bubbletea) | Model/Update/View runtime |
| Styling | [lipgloss](https://github.com/charmbracelet/lipgloss) | Borders, colours, layout |
| Widgets | [bubbles](https://github.com/charmbracelet/bubbles) | viewport, list, help |
| VT emulator | [charmbracelet/x/vt](https://github.com/charmbracelet/x) | Escape sequence parser, scrollback, DEC modes |
| Cell render | [charmbracelet/ultraviolet](https://github.com/charmbracelet/ultraviolet) | Styled glyph access |
| PTY | [creack/pty](https://github.com/creack/pty) | fork+exec with pseudo-terminal |
| Metrics | [gopsutil/v3](https://github.com/shirou/gopsutil) | CPU, memory |
| Clipboard | [go-osc52](https://github.com/aymanbagabas/go-osc52) | Copy via OSC 52 |
| Config | [BurntSushi/toml](https://github.com/BurntSushi/toml) | TOML parsing |

### Architecture rules

- All async I/O (PTY reads, ticker, file reads) **must** be a `tea.Cmd` — never block `Update()`.
- `Update()` **must** return in ≤16ms (frame budget for ≥30 FPS).
- `View()` **must** return a string with exactly `m.height` lines, each ≤ `m.width` columns.
- Keybindings **only** in `app/keymap.go` via `key.Binding`.
- Cross-component communication **only** via types in `msgs/msgs.go`.
- Styles **only** via Lip Gloss — raw ANSI is forbidden outside `components/terminal/`.
- PTY resize propagates via `pty.Setsize` whenever `tea.WindowSizeMsg` is received.

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

Each exported `tea.Model` has isolated unit tests (no dependency on other components),
fed with synthetic `tea.Msg` values. Integration tests in `tests/integration/` cover new
`tea.Msg` types added to `msgs/msgs.go`.

---
