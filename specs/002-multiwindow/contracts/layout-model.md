# Contract: components/layout — Layout Manager API

**Branch**: `002-multiwindow` | **Date**: 2026-04-16

Define a interface pública do package `components/layout` que será importado por `app/app.go`.

---

## Package: `components/layout`

### Tipos Exportados

```go
// PaneID identifica unicamente um LeafNode na árvore.
type PaneID int

// PaneKind distingue o tipo de conteúdo de um LeafNode.
type PaneKind int
const (
    KindTerminal PaneKind = iota
    KindEditor
)

// SplitDir define a direção de um split.
type SplitDir int
const (
    SplitHorizontal SplitDir = iota // painéis lado a lado
    SplitVertical                    // painéis empilhados
)

// FocusDir define a direção de movimento de foco.
type FocusDir int
const (
    FocusLeft FocusDir = iota
    FocusRight
    FocusUp
    FocusDown
)

// ResizeDir define se o painel cresce ou encolhe.
type ResizeDir int
const (
    ResizeGrow   ResizeDir = iota
    ResizeShrink
)

// Axis define o eixo de resize.
type Axis int
const (
    AxisHorizontal Axis = iota
    AxisVertical
)
```

### Model

```go
// Model é o tea.Model do layout manager.
// Implementa tea.Model: Init, Update, View.
type Model struct { /* unexported fields */ }

// New cria um layout com um único painel terminal.
// cfg é propagado para novos terminais/editores criados em splits.
func New(cfg config.Config) (Model, error)

// Init inicia os Cmds de todos os painéis existentes.
func (m Model) Init() tea.Cmd

// Update roteia mensagens para o painel correto e gerencia a árvore.
// Mensagens tratadas: PaneSplitMsg, PaneCloseMsg, PaneFocusMoveMsg,
// PaneResizeMsg, LayoutResizeMsg, PtyOutputMsg, PtyInputMsg,
// TerminalResizeMsg, EditorResizeMsg, OpenFileMsg, tea.KeyMsg.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd)

// View renderiza a árvore de painéis para uma string de exatamente
// m.height linhas e m.width colunas (contrato View da Constituição).
func (m Model) View() string

// FocusedKind retorna o tipo de conteúdo do painel ativo (para app.go
// saber se deve rotear teclas para o PTY ou para o editor).
func (m Model) FocusedKind() PaneKind

// PaneCount retorna o número de LeafNodes na árvore (máx 4).
func (m Model) PaneCount() int

// SetSidebarFocus define se algum painel da área de conteúdo tem foco
// (false quando a sidebar ou statusbar está com foco).
func (m Model) SetContentFocused(active bool) Model
```

---

## Contrato de Rendering

- `View()` retorna uma string com **exatamente** `m.height` linhas e cada linha ≤ `m.width` colunas.
- O painel com foco ativo recebe borda com cor de acento (definida em Lip Gloss via estilo compartilhado).
- Painéis sem foco recebem borda com cor neutra.
- Lip Gloss é o único meio de estilização — sem ANSI escape codes diretos.

---

## Garantias de Performance

- `Update()` retorna em ≤ 16ms (sem I/O bloqueante).
- Todo I/O de PTY (leitura de output, resize) ocorre via `tea.Cmd` assíncrono.
- `View()` não aloca novas strings desnecessariamente — usa `strings.Builder` interno.

---

## Integração com app.go

```go
// Substituição no app.Model:
//   Remover: term terminal.Model, ed editor.Model
//   Adicionar: layout layout.Model

// Em app.handleResize:
m.layout = m.layout.SetSize(contentWidth, contentHeight).(layout.Model)

// Em app.handleKey (roteamento de split, close, navegação):
next, cmd := m.layout.Update(msg)
m.layout = next.(layout.Model)

// Em app.View:
content := m.layout.View()
```
