# Contract: app/keymap.go — Multiwindow Keybindings

**Branch**: `002-multiwindow` | **Date**: 2026-04-16

Define os novos `key.Binding` a adicionar em `app/keymap.go` e as entradas correspondentes em `config.Keybindings`.

---

## Novos Key Bindings

### Adições ao `KeyMap` struct

```go
// Splits
SplitHorizontal key.Binding  // dividir lado a lado
SplitVertical   key.Binding  // dividir empilhado

// Fechar painel
ClosePane key.Binding

// Navegação entre painéis (inspirado em Hyprland Super+H/J/K/L)
FocusPaneLeft  key.Binding
FocusPaneRight key.Binding
FocusPaneUp    key.Binding
FocusPaneDown  key.Binding

// Resize de painel
GrowPaneHorizontal   key.Binding  // expandir painel ativo → direita
ShrinkPaneHorizontal key.Binding  // recolher painel ativo ← esquerda
GrowPaneVertical     key.Binding  // expandir painel ativo ↓ baixo
ShrinkPaneVertical   key.Binding  // recolher painel ativo ↑ cima

// Sidebar resize
GrowSidebar   key.Binding
ShrinkSidebar key.Binding
```

### Defaults (config.Keybindings)

| Campo em `config.Keybindings` | Default keys | Nota |
|---|---|---|
| `SplitHorizontal` | `["alt+\\"]` | Alt+backslash = side by side |
| `SplitVertical` | `["alt+-"]` | Alt+hyphen = stacked |
| `ClosePane` | `["alt+q"]` | Hyprland: Super+Q |
| `FocusPaneLeft` | `["alt+h", "alt+left"]` | Hyprland: Super+H |
| `FocusPaneRight` | `["alt+l", "alt+right"]` | Hyprland: Super+L |
| `FocusPaneUp` | `["alt+k", "alt+up"]` | Hyprland: Super+K |
| `FocusPaneDown` | `["alt+j", "alt+down"]` | Hyprland: Super+J |
| `GrowPaneHorizontal` | `["alt+shift+l", "alt+shift+right"]` | Hyprland resize mode |
| `ShrinkPaneHorizontal` | `["alt+shift+h", "alt+shift+left"]` | |
| `GrowPaneVertical` | `["alt+shift+j", "alt+shift+down"]` | |
| `ShrinkPaneVertical` | `["alt+shift+k", "alt+shift+up"]` | |
| `GrowSidebar` | `["alt+shift+]"]` | |
| `ShrinkSidebar` | `["alt+shift+["]` | |

### Ações em `config.Keybindings.Action()`

Adicionar ao switch em `Action()` (ou equivalente):
- `"split_horizontal"`
- `"split_vertical"`
- `"close_pane"`
- `"focus_pane_left"`, `"focus_pane_right"`, `"focus_pane_up"`, `"focus_pane_down"`
- `"grow_pane_h"`, `"shrink_pane_h"`, `"grow_pane_v"`, `"shrink_pane_v"`
- `"grow_sidebar"`, `"shrink_sidebar"`

---

## Integração com Reserved Keys do Terminal

As novas teclas de navegação e resize de painel devem ser adicionadas ao conjunto retornado por `cfg.Keys.GlobalKeys()` para que o `terminal.Model` não as encaminhe ao PTY quando o foco está no terminal.

---

## Adições ao ShortHelp / FullHelp

```go
// ShortHelp — adicionar os bindings mais usados
SplitHorizontal, SplitVertical, ClosePane, FocusPaneLeft, FocusPaneRight

// FullHelp — grupo adicional
{SplitHorizontal, SplitVertical, ClosePane},
{FocusPaneLeft, FocusPaneRight, FocusPaneUp, FocusPaneDown},
{GrowPaneHorizontal, ShrinkPaneHorizontal, GrowSidebar, ShrinkSidebar},
```
