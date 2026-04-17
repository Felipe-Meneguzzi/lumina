# Implementation Plan: Multiwindow Layout

**Branch**: `002-multiwindow` | **Date**: 2026-04-16 | **Spec**: [spec.md](spec.md)  
**Input**: Feature specification de `specs/002-multiwindow/spec.md`  
**UX Reference**: Hyprland (binary split tree, Alt+H/J/K/L navigation, Alt+Q close)

## Summary

Adicionar suporte a múltiplos painéis simultâneos ao Lumina, usando um **binary split tree** inspirado no Hyprland. Cada painel pode conter independentemente um editor de arquivo ou um terminal PTY. Máximo de 4 painéis. Toda interação via teclado com atalhos `Alt+` (Hyprland-inspired). A árvore de painéis é gerenciada por um novo package `components/layout`, substituindo os campos `term` e `ed` diretos do `app.Model`.

## Technical Context

**Language/Version**: Go 1.26.1  
**Primary Dependencies**: Bubble Tea v1.3.10, Lip Gloss v1.1.0, Bubbles v1.0.0, creack/pty v1.1.24  
**Storage**: Nenhum novo — os painéis usam os modelos existentes (`editor`, `terminal`)  
**Testing**: `go test ./...` — unit tests por componente, integration tests em `tests/integration/`  
**Target Platform**: Linux e macOS (PTY — sem suporte a Windows)  
**Project Type**: Desktop TUI application  
**Performance Goals**: ≥30 FPS render loop; `Update()` ≤16ms; PTY resize ≤50ms  
**Constraints**: Máximo 4 painéis; tamanho mínimo por painel: 20×5; sem suporte a mouse na v1  
**Scale/Scope**: Feature única em aplicação single-binary; nenhuma mudança de escala de usuários

## Constitution Check

### I. Code Quality ✅

- `components/layout` terá responsabilidade única: gerenciar a árvore de painéis.
- Complexidade do `layout.Update` será mantida ≤ 10 por função — split e resize em funções separadas.
- Cross-component communication via `msgs.PaneSplitMsg`, `msgs.PaneCloseMsg`, etc. — sem imports circulares.
- Sem estado global mutável.

### II. Testing Standards ✅

- `layout.Model` terá unit tests isolados (sem `app.Model`).
- Cada novo `tea.Msg` em `msgs.go` terá integration test em `tests/integration/`.
- TDD para bugfixes: reproduzir com teste antes de corrigir.
- `go test ./...` é pré-requisito para merge.

### III. User Experience Consistency ✅

- Todos os novos keybindings definidos em `app/keymap.go` via `config.Keybindings` — nenhum hardcode nos componentes.
- Bordas de painel via Lip Gloss exclusivamente — sem ANSI direto.
- Foco ativo visível por borda de cor de acento.
- Teclas Alt+ seguem convenção terminal-native (Alt não está reservado pelo WM em emuladores de terminal).

### IV. Performance Requirements ✅

- `layout.View()` percorre a árvore recursivamente e monta string via `strings.Builder` — sem alocações desnecessárias.
- Cada `terminal.Model` interno tem seu próprio `tea.Cmd` para leitura de PTY — não bloqueiam o loop principal.
- `pty.Setsize` é chamado dentro de `tea.Cmd` ao receber `TerminalResizeMsg` para cada painel terminal.

**Resultado**: Nenhuma violação. Nenhuma entrada necessária na tabela de Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/002-multiwindow/
├── plan.md              ← este arquivo
├── spec.md              ← especificação funcional
├── research.md          ← pesquisa: Hyprland UX, arquitetura binary tree
├── data-model.md        ← entidades: PaneNode, SplitNode, LeafNode, layout.Model
├── quickstart.md        ← guia de uso e desenvolvimento
├── contracts/
│   ├── msgs.md          ← contrato: novos tea.Msg em msgs/msgs.go
│   ├── keybindings.md   ← contrato: novos key.Binding em app/keymap.go
│   └── layout-model.md  ← contrato: API pública de components/layout
└── tasks.md             ← gerado por /speckit.tasks (próximo passo)
```

### Source Code (repository root)

```text
components/
├── layout/
│   ├── layout.go        # Model raiz: Init/Update/View, PaneCount, FocusedKind
│   ├── tree.go          # PaneNode interface, SplitNode, LeafNode, operações de split/close
│   ├── focus.go         # Algoritmo de navegação direcional na árvore
│   ├── render.go        # Renderização recursiva da árvore com Lip Gloss
│   ├── layout_test.go   # Unit tests do layout.Model
│   ├── tree_test.go     # Unit tests das operações de árvore
│   └── focus_test.go    # Unit tests da navegação
├── terminal/            # Sem mudanças estruturais; PtyOutputMsg ganha campo PaneID
├── editor/              # Sem mudanças estruturais
├── sidebar/             # Sem mudanças (permanece em app.Model)
└── statusbar/           # Sem mudanças

app/
├── app.go               # Substituir term+ed por layout.Model; atualizar handleResize/handleKey/View
└── keymap.go            # Adicionar 13 novos key.Binding

msgs/
└── msgs.go              # Adicionar 5 novos Msg types; modificar PtyOutputMsg/PtyInputMsg (PaneID)

config/
└── config.go            # Adicionar campos de keybinding para os novos atalhos

tests/
└── integration/
    └── multiwindow_test.go  # Integration tests para cada novo tea.Msg
```

**Structure Decision**: Package `components/layout` novo, responsabilidade única de layout management. `app.go` se torna apenas roteador — não gerencia geometria diretamente. Sidebar e statusbar permanecem em `app.Model`.

## Implementation Phases

### Fase 1 — Core Layout Engine

**Objetivo**: `components/layout` funcional com 1 painel, mas estrutura da árvore pronta.

1. Criar `components/layout/tree.go`: tipos `PaneNode`, `SplitNode`, `LeafNode`, `PaneID`, `PaneKind`, `SplitDir`, `FocusDir`.
2. Criar `components/layout/layout.go`: `Model` struct, `New()`, `Init()`, `Update()` básico (só `LayoutResizeMsg`), `View()` para um único painel, `PaneCount()`, `FocusedKind()`.
3. Migrar `app.go`: substituir `term`/`ed` por `layout layout.Model`; adaptar `handleResize`, `View()`.
4. Garantir que o comportamento existente (1 terminal) permanece idêntico.
5. **Gate**: `go test ./...` zero falhas; comportamento visual unchanged com 1 painel.

### Fase 2 — Split e Close

**Objetivo**: Usuário consegue criar 2, 3, 4 painéis e fechar.

1. Adicionar msgs em `msgs/msgs.go`: `PaneSplitMsg`, `PaneCloseMsg`, `LayoutResizeMsg`.
2. Implementar `tree.go`: `splitLeaf(id PaneID, dir SplitDir)` — encontra a folha e substitui por `SplitNode` com dois filhos.
3. Implementar `tree.go`: `closeLeaf(id PaneID)` — remove a folha, substitui o pai pelo irmão.
4. Implementar `layout.Update` para `PaneSplitMsg` e `PaneCloseMsg`.
5. Adicionar keybindings em `app/keymap.go` e `config.go`: `SplitHorizontal`, `SplitVertical`, `ClosePane`.
6. Adicionar lógica em `app.handleKey` para emitir `PaneSplitMsg` e `PaneCloseMsg`.
7. **Gate**: Unit tests de split e close; integration test de fluxo completo.

### Fase 3 — Múltiplos Terminais PTY

**Objetivo**: Cada painel terminal tem seu próprio processo PTY funcional.

1. Modificar `msgs.PtyOutputMsg` e `msgs.PtyInputMsg` para carregar `PaneID` (breaking change → `DECISIONS.md`).
2. Atualizar `terminal.Model`: cada instância emite `PtyOutputMsg{PaneID: id, ...}`.
3. Implementar roteamento no `layout.Update`: `PtyOutputMsg` e `PtyInputMsg` roteados pelo `PaneID` para a folha correta.
4. Garantir que `TerminalResizeMsg` com dimensões corretas é enviado para cada folha terminal ao fazer split ou redimensionar janela.
5. **Gate**: Dois terminais side-by-side com processos independentes; resize propaga `pty.Setsize` em ambos.

### Fase 4 — Navegação de Foco

**Objetivo**: `Alt+H/J/K/L` move foco entre painéis (inspirado em Hyprland).

1. Adicionar `msgs.PaneFocusMoveMsg` em `msgs.go`.
2. Implementar `focus.go`: algoritmo de navegação direcional na árvore binária (encontrar vizinho mais próximo na direção dada).
3. Implementar `layout.Update` para `PaneFocusMoveMsg`.
4. Adicionar keybindings: `FocusPaneLeft/Right/Up/Down` em `keymap.go` e `config.go`.
5. Garantir que `Alt+` keys são adicionados a `GlobalKeys()` para não serem encaminhados ao PTY.
6. Renderizar borda de acento no painel focado vs. borda neutra nos demais (em `render.go`).
7. **Gate**: Unit tests de algoritmo de foco em todas as direções; visual: borda muda ao navegar.

### Fase 5 — Resize de Painéis e Sidebar

**Objetivo**: `Alt+Shift+H/L/J/K` ajusta o ratio do split; `Alt+Shift+[/]` ajusta sidebar.

1. Adicionar `msgs.PaneResizeMsg` em `msgs.go`.
2. Implementar `layout.Update` para `PaneResizeMsg`: encontrar o `SplitNode` pai do painel focado, ajustar `Ratio` em ±0.05, clampar em [0.1, 0.9].
3. Propagar novas dimensões calculadas para os `LeafNode`s afetados.
4. Adicionar keybindings de resize de painel e de sidebar em `keymap.go` e `config.go`.
5. Implementar lógica de resize de sidebar em `app.handleKey` (modifica `m.sidebarWidth` e emite `LayoutResizeMsg`).
6. **Gate**: Unit tests de resize com clamp; resize de sidebar não ultrapassa min/max.

## Complexity Tracking

Nenhuma violação da Constituição identificada. Tabela omitida.

## Artefatos Gerados

| Artefato | Caminho |
|---|---|
| Spec | `specs/002-multiwindow/spec.md` |
| Research | `specs/002-multiwindow/research.md` |
| Data Model | `specs/002-multiwindow/data-model.md` |
| Contract: Messages | `specs/002-multiwindow/contracts/msgs.md` |
| Contract: Keybindings | `specs/002-multiwindow/contracts/keybindings.md` |
| Contract: Layout API | `specs/002-multiwindow/contracts/layout-model.md` |
| Quickstart | `specs/002-multiwindow/quickstart.md` |
| Tasks | `specs/002-multiwindow/tasks.md` (próximo: `/speckit-tasks`) |
