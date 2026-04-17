# Data Model: Mouse Text Selection in Normal Mode

**Feature**: 005-mouse-select-copy
**Date**: 2026-04-17

---

## Entidades

### 1. `mouseSelection` (package `terminal`)

Estado de uma seleção de texto via mouse em andamento ou pendente de confirmação.
Vive como ponteiro `*mouseSelection` em `terminal.Model` — `nil` significa "sem seleção ativa".

| Campo   | Tipo  | Descrição |
|---------|-------|-----------|
| `start` | `pos` | Coordenada de início do drag (pane-local: col/row a partir de 0,0 = canto superior esquerdo do conteúdo interno) |
| `end`   | `pos` | Coordenada atual do cursor durante drag; finalizada no Release |
| `pending` | `bool` | `true` quando `mouse_auto_copy=false` e o Release ocorreu — seleção visível, aguardando `y` |

**Estado `pos`** (tipo já existente em `copymode.go`):
```go
type pos struct{ x, y int }
```

### Ciclo de vida da `mouseSelection`

```
nil
 │ Press (em painel terminal, sem PTY passthrough)
 ▼
{start=P, end=P, pending=false}   ← drag em andamento
 │ Motion events
 ▼
{start=P, end=Q, pending=false}   ← seleção se estendendo
 │ Release
 ├─ mouse_auto_copy=true ──────────────────► copyToClipboard → nil (seleção descartada)
 └─ mouse_auto_copy=false ────────────────► pending=true
                                               │
                              ┌────────────────┤
                              │ 'y' (confirm)  │ 'esc'/click (cancel)
                              ▼                ▼
                    copyToClipboard → nil     nil
```

**Eventos que descartam incondicionalmente** (→ `nil`):
- `PaneFocusMsg{Focused: false}` (foco movido para outro painel)
- `tea.WindowSizeMsg` (resize invalida coordenadas)

---

### 2. `Config.MouseAutoCopy` (package `config`)

Campo adicionado à struct `Config` existente.

| Campo | Tipo | TOML key | Default |
|-------|------|----------|---------|
| `MouseAutoCopy` | `bool` | `mouse_auto_copy` | `true` |

Localização em disco: `~/.config/lumina/config.toml`

```toml
# Comportamento de cópia após seleção via mouse.
# true  = copia automaticamente ao soltar o botão do mouse (padrão)
# false = mantém a seleção visível; pressione 'y' para copiar ou 'Esc' para cancelar
mouse_auto_copy = true
```

---

### 3. Novos `tea.Msg` (package `msgs`)

Três novos tipos de mensagem cross-component adicionados a `msgs/msgs.go`:

#### `MouseSelectMsg`

Roteia um evento de mouse para o terminal como operação de seleção Lumina (não PTY passthrough). Emitido por `app.handleMouse`.

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `PaneID` | `int` | ID do painel terminal de destino |
| `Mouse` | `tea.MouseMsg` | Evento de mouse com coordenadas pane-local (border subtraído) |

#### `MouseSelectConfirmMsg`

Confirma uma seleção pendente (copia texto para clipboard). Emitido por `app.handleKey` quando o usuário pressiona `y` e há seleção pendente.

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `PaneID` | `int` | ID do painel terminal com seleção pendente |

#### `MouseSelectCancelMsg`

Descarta uma seleção pendente sem alterar o clipboard. Emitido por `app.handleKey` quando o usuário pressiona `esc` e há seleção pendente.

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `PaneID` | `int` | ID do painel terminal com seleção pendente |

---

## Relacionamentos

```
app.Model
  └─ handleMouse() ──────────────► MouseSelectMsg ──────► layout → terminal.Update()
  └─ handleKey()   ──────────────► MouseSelectConfirmMsg ► layout → terminal.Update()
                                   MouseSelectCancelMsg  ► layout → terminal.Update()

terminal.Model
  └─ mouseSelection *mouseSelection  (nil = inativo)
  └─ copy           *copyState       (nil = inativo; copy mode existente, inalterado)

config.Config
  └─ MouseAutoCopy bool  (lido por app.Model no handleMouse e repassado via Mouse.Action)
```

**Invariante**: `m.copy` e `m.mouseSelection` NUNCA são ambos não-nil simultaneamente.
Entrar em copy mode descarta qualquer seleção de mouse ativa, e vice-versa.

---

## Validações

| Regra | Local de enforcement |
|-------|---------------------|
| Coordenadas de `mouseSelection` sempre clampadas a `[0, cols-1]` × `[0, rows-1]` | `startMouseSelection`, `updateMouseSelection` |
| `mouseSelection` descartada ao receber `PaneFocusMsg{Focused: false}` | `terminal.Update()` |
| `mouseSelection` descartada ao receber `tea.WindowSizeMsg` | `terminal.Update()` |
| `y` interceptado APENAS quando `m.layout.FocusedHasPendingSelection()` | `app.handleKey()` |
| `mouseSelection` e `copy` mutuamente exclusivos | `enterCopyMode()` limpa `m.mouseSelection`; `startMouseSelection()` verifica `m.copy == nil` |
