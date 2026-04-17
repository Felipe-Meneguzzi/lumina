# Contracts: Mouse Text Selection in Normal Mode

**Feature**: 005-mouse-select-copy
**Date**: 2026-04-17

---

## 1. API Pública do `terminal.Model`

Novos métodos exportados adicionados a `terminal.Model`:

### `HasMouseSelection() bool`
Retorna `true` quando há uma seleção de mouse ativa (drag em andamento ou pendente
de confirmação). Usado por `layout.go` para expor o estado ao `app.go`.

```go
func (m Model) HasMouseSelection() bool { return m.mouseSelection != nil }
```

### `HasPendingSelection() bool`
Retorna `true` especificamente quando a seleção foi completada (Release ocorreu) mas
ainda aguarda confirmação com `y` (`mouse_auto_copy=false`). Usado por `app.handleKey`
para saber quando interceptar `y`.

```go
func (m Model) HasPendingSelection() bool {
    return m.mouseSelection != nil && m.mouseSelection.pending
}
```

---

## 2. API Pública do `layout.Model`

Novos métodos exportados adicionados a `layout.Model` (análogos a `FocusedInCopyMode()`):

### `FocusedHasMouseSelection() bool`
Delega para `HasMouseSelection()` do painel terminal focado. Retorna `false` se o
painel focado não for do tipo `KindTerminal`.

### `FocusedHasPendingSelection() bool`
Delega para `HasPendingSelection()` do painel terminal focado. Retorna `false` se o
painel focado não for do tipo `KindTerminal`.

---

## 3. Novos `tea.Msg` em `msgs/msgs.go`

### `MouseSelectMsg`

```go
// MouseSelectMsg roteia um evento de mouse para seleção Lumina no terminal focado,
// em vez de encaminhar ao PTY. Coordenadas X/Y em msg.Mouse são pane-local
// (0,0 = célula superior esquerda do conteúdo interno, border subtraído).
type MouseSelectMsg struct {
    PaneID int
    Mouse  tea.MouseMsg
}
```

**Quando emitido**: por `app.handleMouse` quando:
- Caso A: `!FocusedMouseEnabled()` (inner app sem mouse tracking) + press/motion/release sobre o painel terminal focado
- Caso B: `msg.Shift == true` + `FocusedMouseEnabled()` + press/motion/release sobre o painel terminal focado

**Quem consome**: `layout.Update()` → roteia para `terminal.Update()` pelo `PaneID`.

**Efeito em `terminal.Update()`**:
- `Action == Press` → inicia seleção (`startMouseSelection`)
- `Action == Motion` → atualiza coordenada final (`updateMouseSelection`)
- `Action == Release` → finaliza (`finalizeMouseSelection`):
  - `mouse_auto_copy=true` → copia texto + retorna `copyToClipboard(text)` cmd + limpa seleção
  - `mouse_auto_copy=false` → marca `pending=true`, mantém highlight

---

### `MouseSelectConfirmMsg`

```go
// MouseSelectConfirmMsg confirma uma seleção pendente, copiando o texto para o clipboard.
// Emitido por app.handleKey quando o usuário pressiona 'y' e há seleção pendente.
type MouseSelectConfirmMsg struct {
    PaneID int
}
```

**Pré-condição**: `FocusedHasPendingSelection() == true`

**Efeito em `terminal.Update()`**:
1. Extrai texto da seleção (`extractMouseSelection()`)
2. Limpa `m.mouseSelection = nil`
3. Retorna `copyToClipboard(text)` cmd (mesmo mecanismo do copy mode)

---

### `MouseSelectCancelMsg`

```go
// MouseSelectCancelMsg descarta uma seleção pendente sem alterar o clipboard.
// Emitido por app.handleKey quando o usuário pressiona 'esc' e há seleção pendente.
type MouseSelectCancelMsg struct {
    PaneID int
}
```

**Pré-condição**: `FocusedHasPendingSelection() == true`

**Efeito em `terminal.Update()`**: `m.mouseSelection = nil` (sem cmd).

---

## 4. Contrato do `config.toml`

```toml
# ~/.config/lumina/config.toml

# Comportamento de cópia após seleção via mouse:
#   true  = copia automaticamente ao soltar o botão (padrão)
#   false = mantém a seleção visível; pressione 'y' para copiar ou 'Esc' para cancelar
mouse_auto_copy = true
```

**Valores válidos**: `true` / `false` (TOML boolean). Qualquer outro valor causa erro
de parsing no `toml.DecodeFile`, que é propagado por `LoadConfig()` ao caller.

**Retrocompatibilidade**: configs existentes sem o campo recebem `true` via `defaults()`.

---

## 5. Contrato de Roteamento em `app.handleMouse`

```
tea.MouseMsg recebido
         │
         ▼
É Alt+wheel? ──yes──► scrollback (lógica existente, inalterada)
         │
         no
         ▼
focus == focusContent && KindTerminal?
         │no
         ▼
   sidebar drag / focus click (lógica existente)
         │
         yes
         ▼
Inner app tem mouse tracking ativo? (FocusedMouseEnabled())
   ├─ NÃO ─────────────────────────────────► [A] emite MouseSelectMsg (plain drag = seleção Lumina)
   └─ SIM
         │
         ▼
    msg.Shift == true?
   ├─ SIM ─────────────────────────────────► [A] emite MouseSelectMsg (Shift bypass)
   └─ NÃO ─────────────────────────────────► [B] emite PtyMouseMsg (forwarding existente)
```

**[A]** Antes de emitir `MouseSelectMsg`, calcular coordenadas pane-local:
```
localX = msg.X - sidebarWidth - borderTolerance - focusBoundsX
localY = msg.Y - focusBoundsY - borderTolerance
```
Clampar ao range `[0, paneWidth-2] × [0, paneHeight-2]`.

---

## 6. Contrato de Interceção de Teclas em `app.handleKey`

Inserido ANTES do bloco de forwarding PTY, após a verificação de copy mode:

```
FocusedHasPendingSelection() == true?
   ├─ NÃO ──► continua fluxo normal (forwarding PTY ou atalhos globais)
   └─ SIM
         │
         ▼
    msg.String() == "y"  ──► MouseSelectConfirmMsg{PaneID: FocusedID()}
    msg.String() == "esc" ─► MouseSelectCancelMsg{PaneID: FocusedID()}
    qualquer outra tecla ──► forwarding normal para PTY
```

**Nota**: `y` é consumido (não encaminhado ao PTY) apenas quando há seleção pendente.
Em todos os outros contextos, `y` vai ao PTY como de costume.
