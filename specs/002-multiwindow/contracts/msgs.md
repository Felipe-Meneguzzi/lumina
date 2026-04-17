# Contract: msgs/msgs.go — Multiwindow Messages

**Branch**: `002-multiwindow` | **Date**: 2026-04-16

Todos os novos `tea.Msg` para a feature de multiwindow devem ser adicionados a `msgs/msgs.go`. Este contrato define os tipos, seus campos, e quem os emite/consome.

---

## Mensagens Novas

### PaneSplitMsg

Solicita a divisão do painel com foco ativo.

```go
// PaneSplitMsg solicita a divisão do painel ativo.
// Emitido por: app.handleKey quando o usuário pressiona o keybinding de split.
// Consumido por: layout.Model.Update()
type PaneSplitMsg struct {
    Direction SplitDir // SplitHorizontal | SplitVertical
}
```

**Fluxo**: `app.handleKey → PaneSplitMsg → layout.Update → layout.Model (nova árvore)`

---

### PaneCloseMsg

Solicita o fechamento do painel com foco ativo.

```go
// PaneCloseMsg solicita o fechamento do painel ativo.
// Emitido por: app.handleKey quando o usuário pressiona o keybinding de fechar painel.
// Consumido por: layout.Model.Update()
type PaneCloseMsg struct{}
```

**Comportamento esperado**:
- Se apenas 1 painel existir: ignorado (sem efeito, sem erro).
- Se o painel for um terminal: o processo PTY associado deve ser encerrado antes de remover o nó.

---

### PaneFocusMoveMsg

Solicita a mudança de foco para o painel vizinho na direção indicada.

```go
// PaneFocusMoveMsg solicita mover o foco para um painel vizinho.
// Emitido por: app.handleKey nos keybindings de navegação.
// Consumido por: layout.Model.Update()
type PaneFocusMoveMsg struct {
    Direction FocusDir // FocusLeft | FocusRight | FocusUp | FocusDown
}
```

**Comportamento esperado**: se não houver vizinho na direção solicitada, o foco não muda (sem wrap-around).

---

### PaneResizeMsg

Solicita o ajuste do ratio de divisão do painel pai do painel ativo.

```go
// PaneResizeMsg solicita ajuste incremental do tamanho do painel ativo.
// Emitido por: app.handleKey nos keybindings de resize.
// Consumido por: layout.Model.Update()
type PaneResizeMsg struct {
    Direction ResizeDir // ResizeGrow | ResizeShrink
    Axis      Axis      // AxisHorizontal | AxisVertical
}
```

**Comportamento esperado**:
- Ajusta o `Ratio` do `SplitNode` pai do painel ativo em incrementos de `0.05`.
- Respeitando os limites `[0.1, 0.9]`.

---

### LayoutResizeMsg

Propaga novo tamanho total da área de conteúdo (excluindo sidebar e statusbar) para o `layout.Model`.

```go
// LayoutResizeMsg propaga o tamanho da área de conteúdo para o layout manager.
// Emitido por: app.handleResize sempre que tea.WindowSizeMsg ou sidebar resize ocorrer.
// Consumido por: layout.Model.Update() — repropaga individualmente para cada LeafNode.
type LayoutResizeMsg struct {
    Width  int
    Height int
}
```

---

### SidebarResizeMsg (modificação)

A `SidebarResizeMsg` existente em `msgs.go` já está correta para a sidebar. Nenhuma mudança necessária.

---

## Tipos de Suporte (definir em msgs.go ou em components/layout)

Os tipos `SplitDir`, `FocusDir`, `Axis`, `ResizeDir`, e `PaneID` devem ser definidos em `components/layout/layout.go` (não em `msgs.go`), pois são internos ao layout manager. As mensagens em `msgs.go` referenciam esses tipos via import de `components/layout`.

**Alternativa considerada**: definir os tipos em `msgs.go` para evitar import de `components/layout` em `msgs.go`. Rejeitada porque criaria dependência circular potencial: `app` importa `msgs` importa `layout`.

**Decisão**: tipos de suporte ficam em `components/layout`; `msgs.go` importa `components/layout` apenas para os tipos enum. Se isso criar ciclo circular, os tipos são promovidos para `msgs.go` como tipos autônomos.

---

## Mensagens Existentes: Impacto do Multiwindow

| Mensagem existente | Mudança necessária |
|---|---|
| `TerminalResizeMsg` | Continua igual — o `layout.Model` envia uma por `LeafNode` terminal |
| `EditorResizeMsg` | Continua igual — o `layout.Model` envia uma por `LeafNode` editor |
| `PtyOutputMsg`, `PtyInputMsg` | Devem carregar um `PaneID` para roteamento correto com múltiplos terminais |
| `FocusChangeMsg` | Substituída internamente pelo `PaneFocusMoveMsg`; `FocusTarget` de app permanece para sidebar vs. layout |

### Mudança em PtyOutputMsg e PtyInputMsg

```go
// PtyOutputMsg — adicionado campo PaneID para roteamento com múltiplos terminais.
type PtyOutputMsg struct {
    PaneID int   // NOVO: identifica qual terminal emitiu o output
    Data   []byte
    Err    error
}

// PtyInputMsg — adicionado campo PaneID.
type PtyInputMsg struct {
    PaneID int   // NOVO: identifica qual terminal recebe o input
    Data   []byte
}
```

Esta é uma **breaking change** em `msgs.go` — requer bump de MINOR na versão da Constituição e entrada em `DECISIONS.md` conforme § Development Workflow.
