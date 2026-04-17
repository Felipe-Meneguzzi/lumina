# Tasks: Multiwindow Layout

**Input**: Design documents de `specs/002-multiwindow/`  
**Prerequisites**: plan.md ✅ spec.md ✅ research.md ✅ data-model.md ✅ contracts/ ✅ quickstart.md ✅

**UX Reference**: Hyprland (binary split tree, `Alt+H/J/K/L` navigation)

**Testes**: incluídos — a Constituição exige unit tests para todo `tea.Model` exportado e integration tests para todo novo `tea.Msg`.

## Format: `[ID] [P?] [Story?] Descrição com caminho de arquivo`

- **[P]**: Pode rodar em paralelo (arquivos diferentes, sem dependências incompletas)
- **[Story]**: User story do spec.md (US1, US2, US3, US4)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Adicionar novos tipos de mensagem, campos de configuração e keybindings — tudo que outras fases dependem.

- [X] T001 Adicionar `LayoutResizeMsg`, `PaneSplitMsg`, `PaneCloseMsg`, `PaneFocusMoveMsg`, `PaneResizeMsg` a `msgs/msgs.go`; modificar `PtyOutputMsg` e `PtyInputMsg` adicionando campo `PaneID int`
- [X] T002 [P] Adicionar campos de keybinding multiwindow (`SplitHorizontal`, `SplitVertical`, `ClosePane`, `FocusPaneLeft/Right/Up/Down`, `GrowPane*`, `ShrinkPane*`, `GrowSidebar`, `ShrinkSidebar`) a `config/config.go` com valores default conforme `contracts/keybindings.md`
- [X] T003 [P] Adicionar os 13 novos `key.Binding` ao `KeyMap` em `app/keymap.go`; atualizar `NewKeyMap()`, `ShortHelp()`, `FullHelp()`; adicionar os atalhos `Alt+*` ao retorno de `GlobalKeys()` para que não sejam repassados ao PTY
- [X] T004 [P] Registrar a breaking change de `PtyOutputMsg`/`PtyInputMsg` em `DECISIONS.md` com justificativa (roteamento multi-terminal) e bump de versão MINOR na Constituição em `.specify/memory/constitution.md`

**Checkpoint**: Todos os tipos e configurações definidos — as demais fases podem referenciar estas definições.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Criar o package `components/layout/` com um único painel funcional e migrar `app.go`. Nenhuma user story pode ser implementada sem este fundamento.

**⚠️ CRÍTICO**: Esta fase deve ser completada antes de qualquer user story.

- [X] T005 Criar `components/layout/tree.go` com: interface `PaneNode`, structs `SplitNode` e `LeafNode`, tipos `PaneID`, `PaneKind` (`KindTerminal`/`KindEditor`), `SplitDir` (`SplitHorizontal`/`SplitVertical`), `FocusDir`, `ResizeDir`, `Axis`; funções auxiliares `countLeaves(PaneNode) int` e `findLeaf(PaneNode, PaneID) *LeafNode`
- [X] T006 Criar `components/layout/render.go` com: função recursiva `renderNode(node PaneNode, w, h int) string` usando `lipgloss.JoinHorizontal`/`JoinVertical` para `SplitNode` e chamando `model.View()` para `LeafNode`; borda Lip Gloss neutra em painéis sem foco (borda de acento será adicionada na US2)
- [X] T007 Criar `components/layout/layout.go` com: struct `Model` (campos `root PaneNode`, `focused PaneID`, `nextID PaneID`, `width`, `height`, `cfg config.Config`); `New(cfg config.Config) (Model, error)` criando único `LeafNode` terminal; `Init() tea.Cmd`; `Update(msg tea.Msg) (tea.Model, tea.Cmd)` tratando apenas `LayoutResizeMsg` e `PtyOutputMsg` roteado ao terminal único; `View() string` delegando a `renderNode`; `FocusedKind() PaneKind`; `PaneCount() int`; `SetContentFocused(active bool) Model`
- [X] T008 [P] Criar `components/layout/layout_test.go` com unit tests: `TestNew_CreatesSingleTerminalPane`, `TestPaneCount_AfterNew_ReturnsOne`, `TestView_SinglePane_ReturnsNonEmpty`, `TestLayoutResizeMsg_PropagatesCorrectDimensions`
- [X] T009 Atualizar `terminal.Model` em `components/terminal/terminal.go` para incluir campo `paneID int` e emitir `msgs.PtyOutputMsg{PaneID: m.paneID, ...}` em vez de `msgs.PtyOutputMsg{...}`; adicionar `SetPaneID(id int)` ao `terminal.Model`
- [X] T010 Migrar `app/app.go`: remover campos `term terminal.Model` e `ed editor.Model`; adicionar `layout layout.Model`; atualizar `New()` para chamar `layout.New(cfg)`; atualizar `Init()` para delegar a `m.layout.Init()`; atualizar `handleResize()` para calcular `contentWidth`/`contentHeight` e emitir `LayoutResizeMsg`; atualizar `View()` para usar `m.layout.View()`; rotear `PtyOutputMsg`, `PtyInputMsg`, `TerminalResizeMsg`, `OpenFileMsg` para `m.layout.Update(msg)`; remover casos `EditorResizeMsg` e `SidebarResizeMsg` que iam para ed/term diretos
- [X] T011 Executar `go test ./...` e corrigir falhas de compilação causadas pela migration — garantir zero falhas antes de avançar

**Checkpoint**: Foundation pronta — `lumina` compila, abre com um terminal funcionando igual ao anterior. Tasks de user stories podem começar.

---

## Phase 3: User Story 1 — Dividir o espaço em múltiplos painéis (Priority: P1) 🎯 MVP

**Goal**: Usuário consegue dividir o painel ativo em dois com `Alt+\` (horizontal) ou `Alt+-` (vertical), até o máximo de 4 painéis.

**Independent Test**: Abrir o Lumina, pressionar `Alt+\`, verificar que dois painéis independentes aparecem side-by-side — cada um com seu próprio terminal funcional.

- [X] T012 [P] [US1] Escrever unit tests em `components/layout/tree_test.go`: `TestSplitLeaf_FromSingle_ReturnsTwoLeaves`, `TestSplitLeaf_ThirdSplit_ReturnsThreeLeaves`, `TestSplitLeaf_AtFourPanes_IsNoop`, `TestSplitLeaf_HorizontalVsVertical_CorrectDirection`
- [X] T013 [P] [US1] Escrever integration test em `tests/integration/multiwindow_test.go`: `TestPaneSplitMsg_CreatesNewPane` — envia `PaneSplitMsg` para `layout.Update`, verifica `PaneCount() == 2`
- [X] T014 [US1] Implementar `splitLeaf(root PaneNode, targetID PaneID, dir SplitDir, newLeaf *LeafNode) PaneNode` em `components/layout/tree.go` — encontra a folha pelo ID, substitui por `SplitNode{Dir, 0.5, original, newLeaf}` (depende de T005)
- [X] T015 [US1] Implementar `layout.Update(PaneSplitMsg)` em `components/layout/layout.go`: verificar `PaneCount() < 4`; criar novo `LeafNode` com `terminal.New(m.cfg)`; chamar `splitLeaf`; atribuir `paneID` ao novo terminal via `SetPaneID`; iniciar o novo terminal com `tea.Cmd`; propagar `TerminalResizeMsg` com dimensões corretas para ambos os filhos do novo split (depende de T014)
- [X] T016 [US1] Atualizar `app.handleKey` em `app/app.go`: adicionar casos `"split_horizontal"` → emitir `PaneSplitMsg{SplitHorizontal}` e `"split_vertical"` → emitir `PaneSplitMsg{SplitVertical}`; delegar `PaneSplitMsg` para `m.layout.Update`
- [X] T017 [US1] Implementar roteamento de `PtyOutputMsg` por `PaneID` em `components/layout/layout.go`: percorrer a árvore para encontrar `LeafNode` com `paneID == msg.PaneID`; rotear `PtyInputMsg` da mesma forma (garantir que input do teclado vai ao terminal correto quando focado)

**Checkpoint**: US1 completamente funcional — dois ou mais terminais funcionam em paralelo, split cria processos PTY independentes.

---

## Phase 4: User Story 4 — Fechar um painel (Priority: P2)

**Goal**: Usuário fecha o painel ativo com `Alt+Q`; os demais painéis redistribuem o espaço. Único painel restante não pode ser fechado.

**Independent Test**: Abrir dois painéis (US1 completo), pressionar `Alt+Q`, verificar que o painel fechado desaparece e o remanescente ocupa todo o espaço disponível.

- [X] T018 [P] [US4] Escrever unit tests em `components/layout/tree_test.go`: `TestCloseLeaf_FromTwo_ReturnsSingleLeaf`, `TestCloseLeaf_SinglePane_IsNoop`, `TestCloseLeaf_MiddleOfThree_SiblingExpands`, `TestCloseLeaf_TerminalPTY_ProcessTerminated`
- [X] T019 [P] [US4] Escrever integration test em `tests/integration/multiwindow_test.go`: `TestPaneCloseMsg_RemovesFocusedPane` — split + close, verifica `PaneCount() == 1`
- [X] T020 [US4] Implementar `closeLeaf(root PaneNode, targetID PaneID) (PaneNode, bool)` em `components/layout/tree.go` — encontra o pai do `LeafNode`, substitui o pai pelo irmão do fechado; retorna `(root, false)` se único painel; encerra o processo PTY do terminal fechado via `tea.Cmd`
- [X] T021 [US4] Implementar `layout.Update(PaneCloseMsg)` em `components/layout/layout.go`: chamar `closeLeaf`; se bem-sucedido, mover `m.focused` para o irmão sobrevivente; propagar `TerminalResizeMsg`/`EditorResizeMsg` com novas dimensões para todos os painéis restantes (depende de T020)
- [X] T022 [US4] Atualizar `app.handleKey` em `app/app.go`: adicionar caso `"close_pane"` → emitir `PaneCloseMsg{}`; verificar se o painel ativo tem mudanças não salvas (editor) e emitir `ConfirmCloseMsg` em vez de `PaneCloseMsg` direto

**Checkpoint**: US4 funcional — fechar painéis funciona corretamente, PTY encerrado, espaço redistribuído.

---

## Phase 5: User Story 2 — Navegar entre painéis (Priority: P2)

**Goal**: Usuário move o foco entre painéis com `Alt+H/J/K/L` ou `Alt+Arrows`; painel ativo tem borda de destaque.

**Independent Test**: Abrir dois painéis side-by-side (US1), pressionar `Alt+L`, verificar que a borda do painel direito fica destacada e o esquerdo fica neutra.

- [X] - [ ] T023 [P] [US2] Escrever unit tests em `components/layout/focus_test.go`: `TestMoveFocus_TwoPanesHorizontal_LeftRight`, `TestMoveFocus_NoNeighbor_FocusUnchanged`, `TestMoveFocus_ThreePanes_AllDirections`, `TestMoveFocus_VerticalSplit_UpDown`
- [X] - [ ] T024 [P] [US2] Escrever integration test em `tests/integration/multiwindow_test.go`: `TestPaneFocusMoveMsg_ChangesFocusedPane`
- [X] - [ ] T025 [US2] Criar `components/layout/focus.go` com função `findNeighbor(root PaneNode, currentID PaneID, dir FocusDir) (PaneID, bool)` — algoritmo de navegação na árvore binária: sobe até encontrar um `SplitNode` onde o nó atual está no lado oposto ao da direção, desce pelo outro filho, pega a folha mais próxima
- [X] - [ ] T026 [US2] Implementar `layout.Update(PaneFocusMoveMsg)` em `components/layout/layout.go`: chamar `findNeighbor`; se encontrado, atualizar `m.focused`; chamar `SetContentFocused` nos `LeafNode`s afetados (depende de T025)
- [X] - [ ] T027 [US2] Atualizar `components/layout/render.go`: aplicar borda com cor de acento Lip Gloss no `LeafNode` cujo ID == `m.focused`; borda neutra nos demais — implementar via estilo `focusedBorder` e `inactiveBorder` em variáveis de pacote usando `lipgloss.NewStyle()`
- [X] - [ ] T028 [US2] Atualizar `app.handleKey` em `app/app.go`: adicionar casos `"focus_pane_left/right/up/down"` → emitir `PaneFocusMoveMsg{Direction: layout.FocusLeft/Right/Up/Down}`; garantir que esses casos só emitem o msg quando o foco de app não está na sidebar (`m.focus != msgs.FocusSidebar`)

**Checkpoint**: US2 funcional — navegação direcional entre painéis funciona, borda de acento indica painel ativo.

---

## Phase 6: User Story 3 — Redimensionar painéis e sidebar (Priority: P3)

**Goal**: Usuário ajusta o tamanho dos painéis com `Alt+Shift+H/L/J/K` e da sidebar com `Alt+Shift+[/]`. Tamanhos mínimo e máximo respeitados.

**Independent Test**: Abrir dois painéis side-by-side (US1), pressionar `Alt+Shift+L` três vezes, verificar que o painel esquerdo ficou maior (e o direito menor).

- [X] - [ ] T029 [P] [US3] Escrever unit tests em `components/layout/tree_test.go`: `TestAdjustRatio_GrowsFirst`, `TestAdjustRatio_RespectsMinRatio_0_1`, `TestAdjustRatio_RespectsMaxRatio_0_9`, `TestAdjustRatio_VerticalSplit_CorrectAxis`
- [X] [P] [US3] Escrever integration test em `tests/integration/multiwindow_test.go`: `TestPaneResizeMsg_ChangesParentSplitRatio`
- [X] [US3] Implementar `adjustRatio(root PaneNode, targetID PaneID, delta float64, axis Axis) PaneNode` em `components/layout/tree.go` — encontra o `SplitNode` pai do painel focado cuja direção coincide com `axis`; ajusta `Ratio` em `delta` (±0.05); clamp em [0.1, 0.9]
- [X] [US3] Implementar `layout.Update(PaneResizeMsg)` em `components/layout/layout.go`: mapear `ResizeDir`+`Axis` para sinal de `delta`; chamar `adjustRatio`; recalcular e propagar dimensões para todos os `LeafNode`s afetados via `TerminalResizeMsg`/`EditorResizeMsg` (depende de T031)
- [X] [US3] Implementar resize de sidebar em `app/app.go`: adicionar casos `"grow_sidebar"` e `"shrink_sidebar"` em `handleKey`; ajustar `m.sidebarWidth` em ±2 colunas respeitando min=16 e max=`m.width/3`; emitir `LayoutResizeMsg` com nova `contentWidth` para o `layout.Model`
- [X] [US3] Atualizar `app.handleKey` em `app/app.go`: adicionar casos `"grow_pane_h"`, `"shrink_pane_h"`, `"grow_pane_v"`, `"shrink_pane_v"` → emitir `PaneResizeMsg` com `Direction` e `Axis` corretos

**Checkpoint**: US3 funcional — resize de painéis e sidebar funciona com limites respeitados.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Estabilidade, experiência de janela mínima e documentação final.

- [X] [P] Adicionar aviso no status bar (via `StatusBarNotifyMsg`) quando usuário tenta split com 4 painéis já abertos — emitir em `layout.Update(PaneSplitMsg)` quando `PaneCount() == 4`
- [X] [P] Adicionar aviso no status bar quando usuário tenta fechar único painel restante — emitir em `layout.Update(PaneCloseMsg)` quando `PaneCount() == 1`
- [X] Testar comportamento com janela pequena (80×24): garantir que split com menos de 40 colunas disponíveis por painel é impedido com aviso; adicionar validação em `layout.Update(PaneSplitMsg)` verificando dimensão mínima por painel (20 colunas, 5 linhas)
- [X] [P] Executar `go test ./...` completo e garantir zero falhas; executar `golangci-lint run` e corrigir warnings
- [X] [P] Atualizar `CLAUDE.md` na seção `## Spec Artifacts` para apontar para `specs/002-multiwindow/` como spec ativa da feature multiwindow
- [ ] T040 Validar o quickstart.md manualmente: executar cada cenário descrito em `specs/002-multiwindow/quickstart.md` no Lumina compilado e confirmar que todos funcionam conforme descrito

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup)           → Nenhuma — iniciar imediatamente
Phase 2 (Foundational)    → Depende de Phase 1 completa — BLOQUEIA todas as user stories
Phase 3 (US1 Split)       → Depende de Phase 2 completa ← MVP
Phase 4 (US4 Close)       → Depende de Phase 3 (precisa de multi-pane para testar fechar)
Phase 5 (US2 Navigate)    → Depende de Phase 3 (precisa de multi-pane para testar foco)
Phase 6 (US3 Resize)      → Depende de Phase 3 (precisa de multi-pane para testar resize)
Phase 7 (Polish)          → Depende de todas as fases de user stories
```

### User Story Dependencies

- **US1 (P1)**: Pode iniciar após Phase 2 — nenhuma dependência de outras user stories
- **US4 (P2)**: Depende de US1 completo (fechar um painel requer ter múltiplos)
- **US2 (P2)**: Pode iniciar após Phase 2 em paralelo com US4 — precisa de multi-pane (Phase 3) para teste real
- **US3 (P3)**: Pode iniciar após Phase 2 em paralelo com US4 e US2 — precisa de multi-pane (Phase 3) para teste real

### Within Each Phase

- Tests [T012, T013] e [T018, T019] escritos antes da implementação correspondente
- Structs de `tree.go` antes de `layout.go` (T005 antes de T007)
- `tree.go` antes de `layout.go` Update handlers
- `config.go` antes de `keymap.go` (T002 antes de T003)

---

## Parallel Examples

### Phase 1 (podem rodar em paralelo)

```
T001 msgs.go          ←→  T002 config.go     ←→  T003 keymap.go     ←→  T004 DECISIONS.md
```

### Phase 3 — US1 (tests em paralelo entre si, antes da implementação)

```
T012 tree_test.go  ←→  T013 integration_test.go
        ↓ (ambos falham — expected)
T014 tree.go (splitLeaf)
T015 layout.go (Update PaneSplitMsg)
T016 app.go (handleKey)
T017 layout.go (roteamento PtyOutputMsg)
```

### Phase 5 — US2 (tests em paralelo com render)

```
T023 focus_test.go  ←→  T024 integration_test.go  ←→  T027 render.go (bordas)
        ↓
T025 focus.go (findNeighbor)
T026 layout.go (Update PaneFocusMoveMsg)
T028 app.go (handleKey)
```

---

## Implementation Strategy

### MVP First (User Story 1 — Split)

1. Completar Phase 1: Setup (msgs + config + keymap)
2. Completar Phase 2: Foundational (layout engine + migração app.go)
3. Completar Phase 3: US1 (split + multi-PTY)
4. **PARAR e VALIDAR**: `go test ./...` passa, Lumina abre dois terminais funcionais
5. Demo: dois terminais side-by-side — cada um com processo shell independente

### Incremental Delivery

1. Phase 1 + 2 → Foundation (Lumina funciona igual ao anterior, mas sobre nova arquitetura)
2. Phase 3 (US1) → Split funcional → Demo MVP
3. Phase 4 (US4) → Fechar painéis → maior estabilidade de uso
4. Phase 5 (US2) → Navegação Hyprland-style → experiência completa
5. Phase 6 (US3) → Resize → polimento final
6. Phase 7 → Polish + testes completos → pronto para merge

---

## Notes

- **[P]** = arquivos diferentes, sem dependências incompletas — podem ser delegados a agentes paralelos
- **[US*]** = rastreabilidade para user story do spec
- Testes devem ser escritos antes da implementação e confirmados como FALHANDO antes de implementar
- Fazer commit após cada checkpoint de fase
- Verificar `go test ./...` em cada checkpoint antes de avançar
- A breaking change em `PtyOutputMsg`/`PtyInputMsg` (T001) requer que **todos** os arquivos que referenciam esses tipos sejam atualizados em T009 e T010
