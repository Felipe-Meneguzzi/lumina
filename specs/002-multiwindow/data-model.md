# Data Model: Multiwindow Layout

**Branch**: `002-multiwindow` | **Date**: 2026-04-16  
**Research**: [research.md](research.md)

---

## Entidades Principais

### PaneNode (interface)

Interface que unifica os dois tipos de nós da árvore de layout.

```
PaneNode
├── SplitNode  — nó interno: divide espaço entre dois filhos
└── LeafNode   — nó folha: contém o modelo de um painel real
```

**Responsabilidade**: abstração sobre o que pode ocupar espaço na área de conteúdo.  
**Implementações**: `SplitNode`, `LeafNode` (ambos no package `components/layout`).

---

### SplitNode

Representa uma divisão do espaço em dois subpainéis.

| Campo | Tipo | Descrição |
|---|---|---|
| `Direction` | `SplitDir` (enum) | `Horizontal` (lado a lado) ou `Vertical` (empilhado) |
| `Ratio` | `float64` | Fração do espaço que vai para `First` (0.0–1.0, padrão: 0.5) |
| `First` | `PaneNode` | Filho à esquerda (horizontal) ou cima (vertical) |
| `Second` | `PaneNode` | Filho à direita (horizontal) ou baixo (vertical) |

**Invariantes**:
- `0.1 ≤ Ratio ≤ 0.9` — garante tamanho mínimo para ambos os filhos.
- `First != nil && Second != nil` sempre (nó interno nunca tem filho nulo).

**Estado**: sem estado próprio além dos campos — o espaço é recalculado a cada `View()`.

---

### LeafNode

Representa um painel real exibindo um editor de arquivo ou um terminal PTY.

| Campo | Tipo | Descrição |
|---|---|---|
| `ID` | `PaneID` (`int`) | Identificador único, atribuído sequencialmente pelo `layout.Model` |
| `Kind` | `PaneKind` (enum) | `KindEditor` ou `KindTerminal` |
| `model` | `tea.Model` | O `editor.Model` ou `terminal.Model` interno (não exportado) |
| `width` | `int` | Largura atual em colunas (propagada via resize) |
| `height` | `int` | Altura atual em linhas (propagada via resize) |

**Invariantes**:
- `ID` é único em toda a árvore.
- `model` nunca é `nil` — é inicializado no momento do split ou da abertura do Lumina.

---

### layout.Model

O `tea.Model` raiz do sistema de multiwindow. Gerenciado por `components/layout`.

| Campo | Tipo | Descrição |
|---|---|---|
| `root` | `PaneNode` | Raiz da árvore de painéis |
| `focused` | `PaneID` | ID do `LeafNode` que possui foco ativo |
| `nextID` | `PaneID` | Contador para geração de IDs únicos |
| `width` | `int` | Largura total disponível (excluindo sidebar e statusbar) |
| `height` | `int` | Altura total disponível (excluindo statusbar) |
| `cfg` | `config.Config` | Configuração passada para novos modelos criados em splits |

**Operações expostas** (métodos que geram `tea.Cmd` ou retornam novo `layout.Model`):
- `Split(dir SplitDir) (layout.Model, tea.Cmd)` — divide o painel focado.
- `CloseFocused() (layout.Model, tea.Cmd)` — remove o painel focado.
- `MoveFocus(dir FocusDir) layout.Model` — move o foco para o vizinho.
- `Resize(delta int, axis Axis) layout.Model` — ajusta o `Ratio` do split pai do painel focado.
- `SetSize(w, h int) (layout.Model, tea.Cmd)` — propaga novo tamanho a toda a árvore.

---

### Tipos Auxiliares

```go
type PaneID int

type PaneKind int
const (
    KindTerminal PaneKind = iota
    KindEditor
)

type SplitDir int
const (
    SplitHorizontal SplitDir = iota  // lado a lado
    SplitVertical                     // empilhado
)

type FocusDir int
const (
    FocusLeft FocusDir = iota
    FocusRight
    FocusUp
    FocusDown
)

type Axis int
const (
    AxisHorizontal Axis = iota
    AxisVertical
)
```

---

## Transições de Estado do Layout

```
[1 painel]
   LeafNode(ID=1, Terminal)

↓  Alt+\ (split horizontal)

[2 painéis]
   SplitNode(H, 0.5)
   ├── LeafNode(ID=1, Terminal)   ← foco permanece aqui
   └── LeafNode(ID=2, Terminal)   ← novo painel (herda tipo do irmão)

↓  Alt+- (split vertical no painel ID=2)

[3 painéis]
   SplitNode(H, 0.5)
   ├── LeafNode(ID=1, Terminal)
   └── SplitNode(V, 0.5)
       ├── LeafNode(ID=2, Terminal)  ← foco
       └── LeafNode(ID=3, Terminal)

↓  Alt+Q (fechar painel focado ID=2)

[2 painéis novamente]
   SplitNode(H, 0.5)
   ├── LeafNode(ID=1, Terminal)
   └── LeafNode(ID=3, Terminal)   ← o irmão do fechado ocupa o espaço
```

---

## Regras de Validação

| Regra | Detalhe |
|---|---|
| Máximo 4 painéis | `layout.Model` conta `LeafNode`s; rejeita split se contagem = 4 |
| Mínimo 1 painel | `CloseFocused()` é no-op se a árvore tem apenas 1 `LeafNode` |
| Ratio mínimo | `Ratio` nunca vai abaixo de 0.1 ou acima de 0.9 |
| Sidebar mínima | `sidebarWidth` ≥ 16 colunas quando visível |
| Sidebar máxima | `sidebarWidth` ≤ `totalWidth / 3` |
| Conteúdo mínimo de painel | Cada painel recebe no mínimo 20 colunas e 5 linhas |

---

## Relações com Entidades Existentes

| Entidade existente | Relação |
|---|---|
| `editor.Model` | Armazenado dentro de `LeafNode.model` quando `Kind == KindEditor` |
| `terminal.Model` | Armazenado dentro de `LeafNode.model` quando `Kind == KindTerminal` |
| `sidebar.Model` | Permanece em `app.Model` — externo ao `layout.Model` |
| `statusbar.Model` | Permanece em `app.Model` — externo ao `layout.Model` |
| `config.Config` | Passado para `layout.New()` e propagado para novos modelos criados em splits |
