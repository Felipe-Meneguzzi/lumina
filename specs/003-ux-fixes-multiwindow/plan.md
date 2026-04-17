# Implementation Plan: UX Fixes — Multi-Window Layout

**Branch**: `003-ux-fixes-multiwindow` | **Date**: 2026-04-16 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification from `specs/003-ux-fixes-multiwindow/spec.md`

---

## Summary

Cinco correções de UX para o layout multi-janela do Lumina: (1) sidebar toggling por pane via keybind + resize via mouse; (2) validação de shell com fallback automático; (3) resource monitor toggling global; (4) foco movido para o novo pane após split; (5) close bug causado por falha de PTY.

Root cause unificado de Issues 4+5: quando o shell falha ao iniciar no novo pane, o split é abortado silenciosamente — `PaneCount()` permanece 1, e `handleClose()` rejeita o fechamento. Fixar o shell resolve os dois.

---

## Technical Context

**Language/Version**: Go 1.26  
**Primary Dependencies**: Bubble Tea, Lip Gloss, Bubbles (Charm), creack/pty, gopsutil/v3  
**Storage**: N/A (sem persistência nova)  
**Testing**: `go test ./...` (unit + integration)  
**Target Platform**: Linux/macOS (WSL incluso)  
**Project Type**: TUI desktop app  
**Performance Goals**: `Update()` ≤16ms; `View()` ≥30 FPS  
**Constraints**: Nenhum ANSI escape code direto (exceto camada PTY); keybindings somente em `app/keymap.go`  
**Scale/Scope**: Single binary, max 4 panes simultâneos

---

## Constitution Check

### Princípio I — Code Quality ✓
- Nenhuma função nova excederá 40 linhas; `validateShell()` terá ~10 linhas
- Sem estado global mutável — novos campos em `app.Model` (value type via Bubble Tea)
- Cross-component exclusivamente via `tea.Msg` — nenhuma exceção

### Princípio II — Testing Standards ✓
- Cada novo comportamento testável em isolamento (`app.Model.Update()` com msgs sintéticos)
- TDD para bug fix do shell: primeiro o teste que reproduz o comportamento incorreto, depois o fix
- Integration test para novo msg de toggle (se adicionado a msgs.go — neste caso não há msgs novos)

### Princípio III — User Experience Consistency ✓
- Dois novos keybindings adicionados APENAS em `app/keymap.go` e `config/keybindings.go`
- Nenhum ANSI escape direto
- Estado de foco do pane: border highlight já existente continua funcionando

### Princípio IV — Performance Requirements ✓
- `validateShell()` chamado somente no startup (fora do loop de eventos)
- Mouse drag não bloqueia `Update()` — apenas modifica campos do model
- Toggle sidebar/sbar: O(1) — apenas ajuste de largura e re-render

**Gates**: Sem violações. Nenhuma justificativa necessária.

---

## Project Structure

### Documentation (this feature)

```text
specs/003-ux-fixes-multiwindow/
├── plan.md              # Este arquivo
├── spec.md              # Especificação funcional
├── research.md          # Análise de root causes e decisões
├── data-model.md        # Mudanças em structs e novos campos
├── quickstart.md        # Guia para o usuário
├── contracts/
│   └── keybindings.md   # Contrato dos novos keybindings
└── checklists/
    └── requirements.md  # Checklist de qualidade do spec
```

### Source Code — Arquivos Alterados

```text
config/
├── config.go        # + validateShell(), chamada em defaults() e LoadConfig()
└── keybindings.go   # + ToggleSidebar, ToggleStatusBar em struct + defaults + Action() + GlobalKeys() + LoadKeybindings()

app/
├── keymap.go        # + ToggleSidebar, ToggleStatusBar key.Binding
└── app.go           # + sidebarVisible, sidebarPrevWidth, paneShowSidebar, sbarVisible
                     # + sidebarDragging, sidebarDragStartX
                     # + handlers para toggle_sidebar, toggle_statusbar
                     # + mouse drag em handleMouse()
                     # + propagação de pane focus change para atualizar sidebar
                     # + FocusedID() chamada ao mudar foco

components/layout/
└── layout.go        # handleSplit: mover foco para novo pane
                     # + FocusedID() accessor

main.go              # + tea.WithMouseAllMotion() para suporte a drag
```

**Structure Decision**: Single project — sem novos packages, apenas extensão dos existentes.

---

## Implementation Steps

### Fix 1: Shell Validation (`config/config.go`)

**Problema**: `cfg.Shell` pode conter um executável inválido (ex: `powershell.exe` configurado pelo usuário). `terminal.New()` chama `exec.Command(m.shell)` sem validar, causando falha em `pty.Start()`. O erro aparece como status bar notification de 3s e o split não ocorre.

**Mudanças em `config/config.go`**:
1. Adicionar import `"os/exec"`
2. Adicionar `validateShell(configured string) string`:
   - Testa `configured`, depois `os.Getenv("SHELL")`, depois `/bin/bash`, `/bin/zsh`, `/bin/sh`
   - Usa `exec.LookPath()` para verificar existência
   - Retorna o primeiro válido encontrado
3. Em `defaults()`: `cfg.Shell = validateShell(shell)` (substitui a atribuição direta)
4. Em `LoadConfig()`: após `toml.DecodeFile`, adicionar `cfg.Shell = validateShell(cfg.Shell)`

**Impacto**: Shell nunca será inválido na inicialização ou ao criar novo pane.

---

### Fix 2: Focus Move para Novo Pane (`components/layout/layout.go`)

**Problema**: `handleSplit` mantém foco no pane original. Usuário não percebe que o novo pane está à direita sem foco.

**Mudanças em `layout.go`**:
1. Em `handleSplit()`, substituir:
   ```go
   m.applyFocus(m.focused, true)
   ```
   Por:
   ```go
   m.focused = m.nextID
   m.applyFocus(m.nextID, true)
   ```
2. Adicionar `FocusedID() PaneID { return m.focused }` como accessor público.

**Impacto em `app.go`**: Ao receber a resposta do `updateLayout(PaneSplitMsg{})`, `m.layout.FocusedID()` retornará o novo pane. O `paneShowSidebar` será consultado para o novo pane.

---

### Fix 3: Sidebar Toggle + Estado por Pane (`config/keybindings.go`, `app/keymap.go`, `app/app.go`)

**Fase 3a — Keybinding** (`config/keybindings.go`):
1. Adicionar `ToggleSidebar []string` e `ToggleStatusBar []string` à struct
2. Defaults: `["alt+b"]` e `["alt+m"]`
3. Adicionar em `defaultKeybindings()`, `Action()`, `GlobalKeys()`, `LoadKeybindings()`

**Fase 3b — KeyMap** (`app/keymap.go`):
1. Adicionar `ToggleSidebar key.Binding` e `ToggleStatusBar key.Binding`
2. Inicializar em `NewKeyMap()`
3. Adicionar em `ShortHelp()` e `FullHelp()`

**Fase 3c — App Model** (`app/app.go`):
1. Adicionar campos a `Model`:
   ```go
   sidebarVisible   bool
   sidebarPrevWidth int
   paneShowSidebar  map[layout.PaneID]bool
   sbarVisible      bool
   ```
2. Em `New()`: inicializar `sidebarVisible = true`, `sbarVisible = true`, `paneShowSidebar = make(map[layout.PaneID]bool)`

**Fase 3d — Handlers** (`app/app.go`):

Handler `toggle_sidebar` em `handleKey()`:
```go
case "toggle_sidebar":
    return m.toggleSidebar(), nil
```

Método `toggleSidebar()`:
```go
func (m Model) toggleSidebar() Model {
    focusedID := m.layout.FocusedID()
    // inferir estado atual para este pane (default: visível)
    wasVisible := m.paneShowSidebar[focusedID]
    if _, exists := m.paneShowSidebar[focusedID]; !exists {
        wasVisible = true
    }
    nowVisible := !wasVisible
    m.paneShowSidebar[focusedID] = nowVisible
    return m.applySidebarState(nowVisible)
}

func (m Model) applySidebarState(visible bool) Model {
    if visible {
        w := m.sidebarPrevWidth
        if w < sidebarMinSize {
            w = 30 // default
        }
        m.sidebarWidth = w
        m.sidebarVisible = true
    } else {
        if m.sidebarWidth > 0 {
            m.sidebarPrevWidth = m.sidebarWidth
        }
        m.sidebarWidth = 0
        m.sidebarVisible = false
    }
    return m
}
```

**Propagação ao mudar foco entre panes**: Em `handleKey()`, nos cases `focus_pane_*`, após `updateLayout(PaneFocusMoveMsg{})`, chamar `m.applySidebarForFocusedPane()`:
```go
func (m Model) applySidebarForFocusedPane() Model {
    id := m.layout.FocusedID()
    visible, exists := m.paneShowSidebar[id]
    if !exists {
        visible = true // default
    }
    return m.applySidebarState(visible)
}
```

**Fase 3e — View e Resize**: Sem mudanças em `View()` — sidebar já se auto-oculta quando `sidebarWidth == 0`. `handleResize()` já propaga `sidebarWidth` corretamente.

**Limpeza ao fechar pane**: Em `updateLayout(PaneCloseMsg{})` retornar um model com limpeza do map para o pane fechado. Detectar via `layout.FocusedID()` antes e depois do close — se mudou, o pane foi fechado; remover entrada do `paneShowSidebar`.

---

### Fix 4: Statusbar Toggle (`app/app.go`)

1. Adicionar `sbarVisible bool` (inicializado como `true` em `New()`)
2. Em `handleKey()`:
   ```go
   case "toggle_statusbar":
       m.sbarVisible = !m.sbarVisible
       return m.reapplyResize(), nil
   ```
3. Método `reapplyResize()` — recomputar dimensions sem evento externo:
   ```go
   func (m Model) reapplyResize() (tea.Model, tea.Cmd) {
       return m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})
   }
   ```
4. Em `handleResize()`:
   ```go
   effectiveStatusH := 0
   if m.sbarVisible {
       effectiveStatusH = statusBarHeight
   }
   contentHeight := msg.Height - effectiveStatusH
   ```
5. Em `View()`:
   ```go
   if m.sbarVisible {
       screen = lipgloss.JoinVertical(lipgloss.Left, content, sbarView)
   } else {
       screen = content
   }
   ```

---

### Fix 5: Mouse Drag para Sidebar (`app/app.go`, `main.go`)

**`main.go`**: Adicionar `tea.WithMouseAllMotion()` ao criar o programa:
```go
p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseAllMotion())
```

**`app/app.go`** — novos campos:
```go
sidebarDragging   bool
sidebarDragStartX int
```

**`handleMouse()` atualizado**:
```go
func (m Model) handleMouse(msg tea.MouseMsg) (Model, tea.Cmd) {
    const borderTolerance = 1

    switch msg.Action {
    case tea.MouseActionPress:
        if msg.Button == tea.MouseButtonLeft {
            // Iniciar drag se clicou na borda sidebar/content
            if m.sidebarWidth > 0 && abs(msg.X - m.sidebarWidth) <= borderTolerance {
                m.sidebarDragging = true
                m.sidebarDragStartX = msg.X
                return m, nil
            }
            // Comportamento existente: mudar foco por clique
            if msg.Y >= m.height-statusBarHeight {
                return m, nil
            }
            if m.sidebarWidth > 0 && msg.X < m.sidebarWidth {
                return m.applyFocusOwner(focusSidebar), nil
            }
            return m.applyFocusOwner(focusContent), nil
        }

    case tea.MouseActionMotion:
        if m.sidebarDragging {
            newW := msg.X
            // aplicar limites (reusar lógica de resizeSidebar)
            maxW := m.width / sidebarMaxRatio
            if newW < sidebarMinSize { newW = sidebarMinSize }
            if newW > maxW           { newW = maxW           }
            if newW != m.sidebarWidth {
                m.sidebarWidth = newW
                next, cmd := m.resizeSidebarTo(newW)
                return next.(Model), cmd
            }
        }

    case tea.MouseActionRelease:
        m.sidebarDragging = false
    }

    return m, nil
}
```

Refatorar `resizeSidebar(delta int)` para extrair `resizeSidebarTo(newW int)` que aceita largura absoluta (elimina duplicação entre resize-por-delta e drag).

---

## Complexity Tracking

> Nenhuma violação constitucional — tabela não aplicável.

---

## Test Plan

### Unit Tests (por arquivo modificado)

| Arquivo | Teste |
|---------|-------|
| `config/config_test.go` (novo) | `validateShell` com shell válido, inválido, vazio, env SHELL |
| `app/app_test.go` | Toggle sidebar: visível→oculta→visível; estado por pane; reapplyResize após toggle |
| `app/app_test.go` | Toggle statusbar: contentHeight ajustado; View() sem sbar |
| `app/app_test.go` | Mouse drag: drag inicia na borda; motion atualiza sidebarWidth; release encerra drag |
| `components/layout/layout_test.go` | handleSplit: foco vai para novo pane (não original) |

### Integration Tests (`tests/integration/`)

| Cenário | Verificação |
|---------|-------------|
| Split + close imediato | PaneCount == 2 após split; close funciona |
| Toggle sidebar em pane 1; switch para pane 2 | Sidebar estado independente por pane |
| Shell inválido em config.toml | Lumina inicia com fallback; sem panic |

### TDD para Bug Fix de Shell (Princípio II)

1. Escrever teste em `config/config_test.go` que cria config com shell `"invalid-shell-xyz"` e verifica que `validateShell` retorna um shell válido
2. Verificar que o teste falha (TDD vermelho)
3. Implementar `validateShell`
4. Verificar que o teste passa (TDD verde)
