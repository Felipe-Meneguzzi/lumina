# Research: Multiwindow Layout

**Branch**: `002-multiwindow` | **Date**: 2026-04-16  
**Spec**: [spec.md](spec.md)

---

## 1. Hyprland UX Model → Tradução para TUI

### 1.1 Como o Hyprland gerencia janelas

Hyprland é um compositor Wayland de tiling dinâmico. Suas decisões de UX relevantes para Lumina:

| Conceito Hyprland | Comportamento | Tradução para Lumina |
|---|---|---|
| **Binary split tree** | Cada split divide um nó em dois filhos; a árvore define o layout | `layout.PaneNode` — interface com `SplitNode` e `LeafNode` |
| **Foco direcional** | `Super+H/J/K/L` move o foco para o vizinho mais próximo na direção | `Alt+H/J/K/L` (ou arrows) move foco entre painéis |
| **Resize mode** | `Super+R` entra em modo resize, então arrows ajustam o split ratio | Modo resize implícito: `Alt+Shift+H/L` ajusta ratio do split pai |
| **Split direction** | Hyprland detecta automaticamente horizontal vs. vertical por proporção | Usuário escolhe explicitamente: `Alt+\` (vertical) ou `Alt+-` (horizontal) |
| **Close window** | `Super+Q` fecha a janela ativa; o espaço é redistribuído | `Alt+Q` fecha o painel ativo |
| **Master-stack** | Layout padrão: um master à esquerda, stack à direita | v1: binary dwindling apenas (sem master explícito) |
| **Borders** | A janela com foco tem borda colorida diferente | Lip Gloss border com cor de acento para painel ativo |
| **Window gaps** | Gaps configuráveis entre janelas | Nenhum gap em TUI — bordas LipGloss são o separador |

### 1.2 Por que binary split tree (e não grid fixo)

**Alternativa avaliada**: Grid fixo (2×1, 1×2, 2×2).
- Simples de implementar, mas inflexível: não permite 3 painéis em L, ou um painel grande + dois pequenos.
- Hyprland usa árvore binária justamente para evitar esse problema.

**Decisão**: Usar árvore binária de splits, como Hyprland/i3/Sway.
- Cada `SplitNode` divide seu espaço entre dois filhos (`First` e `Second`).
- Cada `LeafNode` é um painel real (editor ou terminal).
- Limite de 4 folhas = máximo 3 splits na árvore (árvore binária com N folhas tem N-1 nós internos).

**Rationale**: Flexível, incremental (cada split é uma ação), alinha com o mental model do Hyprland que o usuário espera.

---

## 2. Keybindings: Escolhas e Justificativas

### 2.1 Por que `Alt+` como modificador

- `Super` não existe em emuladores de terminal — é interpretado pelo WM antes de chegar à aplicação.
- `Ctrl+W` (prefixo vim/tmux) exigiria sequência dupla, aumentando a latência percebida.
- `Alt+` produz escape sequences detectáveis pela Bubble Tea via `tea.KeyMsg`.
- **Conflito potencial**: `Alt+H/J/K/L` pode conflitar com aplicações que rodam dentro do terminal PTY. Mitigação: quando o foco está em um painel terminal, os `Alt+` de navegação são capturados pelo app e **não** repassados ao PTY. Isso está alinhado com o comportamento do `SetReservedKeys` já existente.

### 2.2 Mapeamento final

| Ação | Keybinding | Paralelo Hyprland |
|---|---|---|
| Split vertical (side by side) | `Alt+\` | `Super+S` ou mouse drag |
| Split horizontal (stacked) | `Alt+-` | `Super+S` (rotacionado) |
| Fechar painel ativo | `Alt+Q` | `Super+Q` |
| Foco → esquerda | `Alt+H` ou `Alt+Left` | `Super+H` |
| Foco → direita | `Alt+L` ou `Alt+Right` | `Super+L` |
| Foco → cima | `Alt+K` ou `Alt+Up` | `Super+K` |
| Foco → baixo | `Alt+J` ou `Alt+Down` | `Super+J` |
| Expandir painel (→ direita/baixo) | `Alt+Shift+L` ou `Alt+Shift+Right` | `Super+R` then arrow |
| Recolher painel (← esquerda/cima) | `Alt+Shift+H` ou `Alt+Shift+Left` | `Super+R` then arrow |
| Redimensionar sidebar (+) | `Alt+Shift+]` | (sem paralelo direto) |
| Redimensionar sidebar (-) | `Alt+Shift+[` | (sem paralelo direto) |

Todos configuráveis via `config.Keybindings` — nunca hardcoded nos componentes.

---

## 3. Arquitetura: Layout Manager em Bubble Tea

### 3.1 Estrutura de dados escolhida

```
PaneNode (interface)
├── SplitNode  { Direction, Ratio float64, First PaneNode, Second PaneNode }
└── LeafNode   { ID PaneID, Kind PaneKind (Editor|Terminal), model tea.Model }
```

**Alternativas consideradas**:
- Lista flat de painéis com coordenadas absolutas: simples, mas não escala para splits aninhados.
- Quadtree: overkill para 4 painéis, complexidade desnecessária.
- **Árvore binária**: escolhida — mínima complexidade, máxima flexibilidade, idiomática do Hyprland/i3.

### 3.2 Onde vive o Layout Manager

**Opção A**: Nova package `components/layout/` com `layout.Model` que contém a árvore e expõe `Init/Update/View`.
**Opção B**: Lógica de layout embutida em `app/app.go`.

**Decisão**: Opção A — `components/layout/`.
- Mantém `app.go` coeso (roteador de mensagens, não gerente de geometria).
- Permite testar o layout manager em isolamento (alinhado com a Constituição).
- Package tem responsabilidade única: transformar mensagens de split/resize/focus em dimensionamento e renderização da árvore.

### 3.3 Integração com app.go

`app.Model` substituirá os campos `term terminal.Model` e `ed editor.Model` por um único campo `layout layout.Model`. A sidebar permanece em `app.Model` por ser sempre lateral e ter redimensionamento próprio.

```
app.Model {
    layout   layout.Model   // novo — contém a árvore de painéis
    side     sidebar.Model  // existente, separado do layout
    sbar     statusbar.Model
    keymap   KeyMap
    ...
}
```

### 3.4 Renderização

A árvore é percorrida recursivamente no `View()`:
- `SplitNode` horizontal: `lipgloss.JoinHorizontal` dos dois filhos com larguras `Ratio * parentWidth` e `(1-Ratio) * parentWidth`.
- `SplitNode` vertical: `lipgloss.JoinVertical` com alturas análogas.
- `LeafNode`: chama `model.View()` com as dimensões calculadas.

### 3.5 PTY resize em múltiplos terminais

Cada `LeafNode` do tipo `Terminal` tem seu próprio `terminal.Model` (e portanto seu próprio PTY). Ao receber `tea.WindowSizeMsg`, o `layout.Model` propaga `TerminalResizeMsg` individualmente para cada folha terminal com suas dimensões corretas.

---

## 4. Sidebar Resize: Mecanismo

A sidebar não entra na árvore de painéis — ela é gerenciada por `app.Model` diretamente.
- `app.sidebarWidth` passa de fixo (30) para variável.
- Mínimo: 16 colunas. Máximo: `width/3` (um terço da tela).
- Incremento: 2 colunas por keypress.
- Ao redimensionar sidebar, o espaço disponível para o `layout.Model` é recalculado e propagado via `LayoutResizeMsg`.

---

## 5. Decisões de Escopo (v1)

| Decisão | Escolha v1 | Futuro |
|---|---|---|
| Splits mistos (ex: 2 colunas + 1 coluna maior) | Suportado via árvore binária | Layouts predefinidos (ex: monocle, master-stack) |
| Abas dentro de painel | Fora de escopo | v2: cada LeafNode pode ter N tabs |
| Drag-to-resize com mouse | Fora de escopo | v3: mouse support |
| Mover painel para outra posição | Fora de escopo | v2: reorder nodes na árvore |
| Workspaces (Hyprland workspaces) | Fora de escopo | v3: multiple layouts |
| Painel fullscreen temporário | Fora de escopo | v2: toggle monocle |

---

## Resolução de Clarificações (Spec)

Nenhum marcador `[NEEDS CLARIFICATION]` permaneceu na spec. Todas as decisões foram tomadas com base nesta pesquisa e documentadas acima.
