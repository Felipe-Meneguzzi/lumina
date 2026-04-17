# Research: UX Fixes — Multi-Window Layout

**Feature**: 003-ux-fixes-multiwindow  
**Date**: 2026-04-16  
**Status**: Complete

---

## 1. Diagnóstico: Bug do Shell "PowerShell" (Issues 2 e 5)

### O que o código faz hoje

`config/config.go` já faz a detecção correta:
```go
shell := os.Getenv("SHELL")
if shell == "" {
    shell = "/bin/sh"
}
```

**Root Cause identificada via leitura do código:**

O bug do "PowerShell" e o bug do "única janela aberta" têm a mesma raiz. Em `components/layout/layout.go`, `handleSplit()` chama `newTerminalLeaf()`. Se o shell falhar ao iniciar (PTY error), o erro é capturado e retorna sem criar o novo pane:

```go
newLeaf, err := newTerminalLeaf(m.nextID, m.cfg)
if err != nil {
    return m, notifyStatus("Erro ao criar terminal: "+err.Error(), msgs.NotifyError)
    // ← retorna sem splitLeaf → PaneCount permanece 1
}
```

Consequência: o usuário vê a mensagem de erro por 3 segundos, mas depois parece que nada aconteceu. `PaneCount() == 1`, então quando tenta fechar recebe "única janela aberta". O split simplesmente **não aconteceu**.

O "PowerShell" que o usuário vê pode ser: a sessão PTY do terminal inicial funcionando, mas com um shell configurado incorretamente via `~/.config/lumina/config.toml` (ex: `shell = "pwsh"` ou `shell = "powershell.exe"`). A mensagem de status de 3 segundos passa despercebida.

### Decisão

- Adicionar `validateShell()` em `config/config.go` que verifica se o executável existe com `exec.LookPath`
- Fallback chain: `$SHELL` → `/bin/bash` → `/bin/zsh` → `/bin/sh`
- Emitir `StatusBarNotifyMsg` no startup com o shell sendo usado (info)
- Se shell inválido for detectado na inicialização, substituir automaticamente pelo fallback e notificar

**Alternativas consideradas**: Manter como está (não resolve o bug para usuários com config errada).

---

## 2. Sidebar Toggle via Keybind (Issue 1)

### O que o código faz hoje

`app/app.go` tem `sidebarWidth int` que controla a largura da sidebar. Já existem `grow_sidebar` (`alt+shift+]`) e `shrink_sidebar` (`alt+shift+[`). Não existe keybind de toggle.

`sidebar.View()` retorna `""` quando `m.width == 0`. `app.View()` verifica se `sideView != ""` antes de renderizar. Portanto, **sidebar se auto-oculta quando `sidebarWidth == 0`**.

### Decisão: Toggle global + estado persistido por pane

A sidebar é globalmente única (renderizada à esquerda de todos os panes). "Por janela" significa: o estado de visibilidade é memorizado por pane — ao mudar de foco, a sidebar reflete o estado do pane em foco.

**Implementação mínima viável:**
- `sidebarVisible bool` em `app.Model` (estado atual)
- `sidebarPrevWidth int` em `app.Model` (para restaurar largura ao re-exibir)
- Keybind `toggle_sidebar` (ex: `alt+b`) — funciona mesmo com terminal em foco
- Ao ocultar: `sidebarWidth = 0`, ao exibir: `sidebarWidth = sidebarPrevWidth || 30`

**Requisito "por janela" (FR-004):**  
Armazenar `paneShowSidebar map[PaneID]bool` em `app.Model`. Quando a troca de foco ocorre (via `applyFocusOwner` ou focus_pane_*), ler o estado do novo pane focado e atualizar `sidebarWidth` conforme.

**Alternativas consideradas**: Sidebar por pane com rendering separado (muito complexo — exigiria layout redesign completo); descartado para esta feature.

---

## 3. Mouse Drag para Redimensionar Sidebar (Issue 1 - secundário)

### Bubble Tea Mouse API

Bubble Tea expõe `tea.MouseMsg` com campos `Action` e `Button`. Para drag, precisa detectar:
- `MouseActionPress` + `MouseButtonLeft` na borda sidebar/content (x == sidebarWidth)
- `MouseActionMotion` enquanto botão pressionado (requer `tea.WithMouseAllMotion()` no programa)
- `MouseActionRelease` + `MouseButtonLeft` para finalizar

`main.go` precisa adicionar `tea.WithMouseAllMotion()` à opção do programa.

Em `app.Model`, adicionar:
- `sidebarDragging bool`
- `sidebarDragStartX int`

Detectar clique na borda (tolerância de ±1 coluna ao redor de `sidebarWidth`) e arrastar.

**Decisão**: Implementar, mas como P2 (prioritário é keybind).

---

## 4. Focus Move para Novo Pane após Split (Issue 4)

### O que o código faz hoje

Em `layout.go` `handleSplit()`:
```go
m.applyFocus(m.focused, true) // keep focus on the original pane
```

Depois do split, foco fica no pane original. O novo pane não tem foco.

**Root Cause do "atalhos não funcionam":** O usuário vê o novo pane (à direita/abaixo) mas o foco ainda está no pane original. Qualquer tecla digitada vai ao PTY do pane original, não ao novo. O usuário pensa que está interagindo com o novo pane mas não está.

### Decisão

Mudar `handleSplit` para transferir o foco para o novo pane após a criação:
```go
m.focused = m.nextID
m.applyFocus(m.nextID, true)
```

**Alternativas consideradas**: Manter foco no original com indicação visual mais clara (complexo de implementar, menos intuitivo); descartado.

---

## 5. Statusbar (Resource Monitor) Toggle (Issue 3)

### O que o código faz hoje

`sbar statusbar.Model` em `app.Model` — já é singleton global ✓. Não há toggle.

Em `app.handleResize()`: `contentHeight := msg.Height - statusBarHeight`. Se ocultar, precisa restituir essa coluna ao content.

### Decisão

- `sbarVisible bool` em `app.Model` (default: true)
- Keybind `toggle_statusbar` (ex: `alt+m`)
- Em `handleResize()`: calcular `effectiveStatusHeight` = `statusBarHeight` se `sbarVisible` else `0`
- Em `View()`: incluir sbar no join vertical apenas se `sbarVisible`

---

## 6. Close Bug — Verificação Final

Com a correção do shell (Issue 2), o split passará a funcionar → `PaneCount() == 2` após split → close funcionará corretamente.

Verificação extra: `handleClose()` em `layout.go` já usa `PaneCount() <= 1` corretamente. Nenhuma mudança adicional necessária nesta função.

---

## Resumo de Decisões

| Decisão | Escolha | Rationale |
|---------|---------|-----------|
| Shell detection | Validar + fallback chain | Previne falha silenciosa de PTY |
| Sidebar "por janela" | Estado por PaneID em map | Balanceia requisito com arquitetura atual |
| Sidebar toggle | Keybind `alt+b` | Consistente com outros alt+key bindings |
| Mouse drag sidebar | `WithMouseAllMotion` + state em app.Model | API nativa do Bubble Tea |
| Focus após split | Mover para novo pane | UX mais intuitiva |
| Statusbar toggle | Keybind `alt+m` | Memônico para "monitor" |
