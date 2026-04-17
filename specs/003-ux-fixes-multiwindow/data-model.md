# Data Model: UX Fixes — Multi-Window Layout

**Feature**: 003-ux-fixes-multiwindow  
**Date**: 2026-04-16

---

## Mudanças em Structs Existentes

### `app.Model` (`app/app.go`)

Campos novos:

```go
type Model struct {
    // ... campos existentes ...

    // Sidebar state
    sidebarVisible  bool             // true = sidebar visível; default true
    sidebarPrevWidth int             // largura antes de ocultar (para restaurar)
    paneShowSidebar map[PaneID]bool  // estado de visibilidade por pane; nil = todos visíveis

    // Statusbar toggle
    sbarVisible bool // true = resource monitor visível; default true

    // Mouse drag para sidebar
    sidebarDragging  bool
    sidebarDragStartX int
}
```

> **Nota**: `PaneID` é `layout.PaneID` (tipo `int`). `app.Model` já importa `layout`, então sem import circular.

### `config.Keybindings` (`config/keybindings.go`)

Campos novos:

```go
type Keybindings struct {
    // ... campos existentes ...
    ToggleSidebar   []string `json:"toggle_sidebar"`   // default: ["alt+b"]
    ToggleStatusBar []string `json:"toggle_statusbar"` // default: ["alt+m"]
}
```

### `app.KeyMap` (`app/keymap.go`)

Campos novos:

```go
type KeyMap struct {
    // ... campos existentes ...
    ToggleSidebar   key.Binding
    ToggleStatusBar key.Binding
}
```

---

## Nenhuma Mudança em `msgs/msgs.go`

Todas as operações novas são tratadas internamente em `app.Model` (sidebar toggle, statusbar toggle, mouse drag). Não há comunicação cross-component nova que exija novos `tea.Msg`.

A única exceção é o mouse drag: o `tea.MouseMsg` já existe como tipo nativo do Bubble Tea e não requer novo msg customizado.

---

## Estado Transitório: `paneShowSidebar`

```
paneShowSidebar: {
    PaneID(1): true,   // pane 1: sidebar visível
    PaneID(2): false,  // pane 2: sidebar oculta
    PaneID(3): true,   // pane 3: sidebar visível
}
```

**Ciclo de vida:**
- Criação do pane: entrada não existe → tratar como `true` (sidebar visível por default)
- Toggle: inverter entrada para o pane focado; atualizar `sidebarWidth` e `sidebarVisible` conforme
- Mudança de foco (focus_pane_*): ler entrada do novo pane focado e reconfigurar sidebar
- Fechamento do pane: remover entrada do map (evitar memory leak)

**Como descobrir o PaneID focado em `app.Model`:**  
`app.Model` tem `m.layout` que expõe `m.layout.FocusedID()` — esse método precisa ser adicionado em `layout.Model`.

---

## Novo Método: `layout.Model.FocusedID()`

```go
// FocusedID returns the PaneID of the currently focused pane.
func (m Model) FocusedID() PaneID { return m.focused }
```

Método simples de 1 linha, sem side effects.

---

## Mudança de Comportamento: `handleSplit` em `layout.go`

```go
// ANTES:
m.applyFocus(m.focused, true) // keep focus on the original pane

// DEPOIS:
oldID := m.focused
m.focused = m.nextID
m.applyFocus(m.nextID, true) // focus moves to new pane
_ = oldID // oldID preserved para possível uso futuro
```

---

## Mudança em `config/config.go`: Validação de Shell

```go
// validateShell verifica se o executável existe e retorna o primeiro shell válido.
func validateShell(configured string) string {
    candidates := []string{configured, os.Getenv("SHELL"), "/bin/bash", "/bin/zsh", "/bin/sh"}
    for _, s := range candidates {
        if s == "" {
            continue
        }
        if _, err := exec.LookPath(s); err == nil {
            return s
        }
    }
    return "/bin/sh" // fallback final — sempre existe em sistemas POSIX
}
```

Chamada em `defaults()` ao resolver `cfg.Shell`:
```go
cfg.Shell = validateShell(cfg.Shell)
```

E também em `LoadConfig()` após decodificar o TOML, para validar shell configurado pelo usuário:
```go
cfg.Shell = validateShell(cfg.Shell)
```
