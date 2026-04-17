---

description: "Task list for feature 005-mouse-select-copy"
---

# Tasks: Mouse Text Selection in Normal Mode

**Input**: Design documents from `specs/005-mouse-select-copy/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/mouse-selection.md, quickstart.md

**Tests**: Incluídos. A Constitution de Lumina (Princípio II) exige unit tests para cada função exportada e integration tests para cada novo `tea.Msg`. Os test tasks abaixo refletem essa exigência.

**Organization**: Agrupado por user story (US1 = seleção normal, US2 = Shift+drag bypass, US3 = copy mode inalterado). MVP = US1.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Paralelizável (sem dependência bloqueante de outra task incompleta)
- **[Story]**: US1 / US2 / US3 — mapeia para a user story da spec
- Paths relativos à raiz do repo `lumina/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Criar os stubs dos arquivos novos para que `go test ./...` não quebre antes da implementação.

- [X] T001 Criar arquivo stub `components/terminal/mouseselect.go` com `package terminal` e comentário `// Package terminal — mouse selection mode (normal mode).`
- [X] T002 [P] Criar arquivo stub `components/terminal/mouseselect_test.go` com `package terminal` + import de `testing` (nenhum teste ainda)
- [X] T003 [P] Criar arquivo stub `tests/integration/mouse_select_test.go` com `package integration_test` + imports de `testing` e `github.com/menegas/lumina/msgs`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Estruturas compartilhadas que TODAS as user stories consomem — sem elas, nada compila.

**⚠️ CRITICAL**: Concluir integralmente antes de começar qualquer user story.

- [X] T004 Adicionar campo `MouseAutoCopy bool \`toml:"mouse_auto_copy"\`` a `Config` em `config/config.go`; adicionar `MouseAutoCopy: true` em `defaults()` (retrocompatível: configs sem o campo recebem `true` automaticamente)
- [X] T005 [P] Adicionar os três novos tipos de mensagem a `msgs/msgs.go`: `MouseSelectMsg{PaneID int; Mouse tea.MouseMsg}`, `MouseSelectConfirmMsg{PaneID int}`, `MouseSelectCancelMsg{PaneID int}` — com comentários de doc conforme `contracts/mouse-selection.md`
- [X] T006 [P] Adicionar métodos `FocusedHasMouseSelection() bool` e `FocusedHasPendingSelection() bool` em `components/layout/layout.go` — delegam para os métodos do `terminal.Model` focado (análogos a `FocusedInCopyMode()`)
- [X] T007 Rodar `go test ./...` — zero falhas (smoke check da fundação)

**Checkpoint**: Código compila; comportamento atual preservado.

---

## Phase 3: User Story 1 — Select and Copy Text with Mouse in Normal Mode (Priority: P1) 🎯 MVP

**Goal**: Click-and-drag no painel terminal (quando a aplicação interna não usa mouse tracking) seleciona texto visualmente; ao soltar, texto copiado automaticamente para o clipboard se `mouse_auto_copy=true`; se `false`, seleção persiste aguardando `y`.

**Independent Test**: Executar `./lumina`, focar o painel terminal com shell simples (sem mouse tracking da aplicação interna), clicar e arrastar sobre texto, verificar highlight e conteúdo do clipboard. Testar ambos os modos (`mouse_auto_copy=true` e `false` em `~/.config/lumina/config.toml`).

### Tests for User Story 1

> **Escrever estes testes ANTES da implementação e garantir que FALHAM inicialmente.**

- [X] T008 [P] [US1] Escrever `TestMouseSelection_StartAndUpdate` em `components/terminal/mouseselect_test.go`: verifica que `startMouseSelection` + `updateMouseSelection` preenchem `start` e `end` corretamente; clamp em bounds
- [X] T009 [P] [US1] Escrever `TestMouseSelection_FinalizeAutoCopy` em `components/terminal/mouseselect_test.go`: `finalizeMouseSelection(autoCopy=true)` retorna cmd não-nil e limpa `m.mouseSelection`
- [X] T010 [P] [US1] Escrever `TestMouseSelection_FinalizeManualConfirm` em `components/terminal/mouseselect_test.go`: `finalizeMouseSelection(autoCopy=false)` define `pending=true` e NÃO limpa `m.mouseSelection`
- [X] T011 [P] [US1] Escrever `TestMouseSelection_ConfirmAndCancel` em `components/terminal/mouseselect_test.go`: `confirmMouseSelection` retorna cmd + limpa; `cancelMouseSelection` limpa sem cmd
- [X] T012 [P] [US1] Escrever `TestExtractMouseSelection_ReturnsCorrectText` em `components/terminal/mouseselect_test.go`: feed de output para o emulador + seleção definida manualmente → texto extraído corresponde ao conteúdo esperado
- [X] T013 [P] [US1] Escrever integration test `TestMouseSelectMsg_PressMotionRelease` em `tests/integration/mouse_select_test.go`: sequência `MouseSelectMsg{Press} + MouseSelectMsg{Motion} + MouseSelectMsg{Release}` no terminal model gera cmd de cópia (autoCopy=true)
- [X] T014 [P] [US1] Escrever integration tests `TestMouseSelectConfirmMsg_CopiesAndClears` e `TestMouseSelectCancelMsg_ClearsWithoutCopy` em `tests/integration/mouse_select_test.go`

### Implementation for User Story 1

- [X] T015 [US1] Implementar em `components/terminal/mouseselect.go`: struct `mouseSelection{start, end pos; pending bool}`; funções `startMouseSelection(x, y int)`, `updateMouseSelection(x, y int)` (com clamp), `finalizeMouseSelection(x, y int, autoCopy bool) tea.Cmd`, `confirmMouseSelection() tea.Cmd`, `cancelMouseSelection()`; métodos `HasMouseSelection() bool` e `HasPendingSelection() bool` em `terminal.Model`
- [X] T016 [US1] Implementar em `components/terminal/mouseselect.go`: `extractMouseSelection() string` (análoga a `extractSelection()` em `copymode.go`, usando `m.mouseSelection.start/end`) e `renderWithMouseSelection() string` (reutiliza `selectionStyle` existente; extrair lógica comum com `renderCopyMode` para helper privado se >40 linhas)
- [X] T017 [US1] Adicionar campo `mouseSelection *mouseSelection` a `terminal.Model` em `components/terminal/terminal.go`; no `Update()`, tratar: `msgs.MouseSelectMsg` (Press → `startMouseSelection`, Motion → `updateMouseSelection`, Release → `finalizeMouseSelection(autoCopy: cfg.MouseAutoCopy)`), `msgs.MouseSelectConfirmMsg` → `confirmMouseSelection`, `msgs.MouseSelectCancelMsg` → `cancelMouseSelection`
- [X] T018 [US1] Em `components/terminal/terminal.go`: (1) no handler de `PaneFocusMsg{Focused: false}`, adicionar `m.mouseSelection = nil`; (2) no handler de `tea.WindowSizeMsg`, adicionar `m.mouseSelection = nil`; (3) em `View()`, adicionar ramo: se `m.mouseSelection != nil`, retornar `m.renderWithMouseSelection()`
- [X] T019 [US1] Atualizar `app/app.go` → `handleMouse()`: no bloco de passthrough PTY, adicionar condição: se `!m.layout.FocusedMouseEnabled()` (inner app sem mouse tracking) e o evento é sobre o painel terminal focado, calcular coordenadas pane-local (subtrair border + sidebarWidth), clampar, emitir `msgs.MouseSelectMsg{PaneID, Mouse: inner}`; NÃO emitir `PtyMouseMsg` neste caso
- [X] T020 [US1] Atualizar `app/app.go` → `handleKey()`: ANTES do bloco de forwarding PTY, verificar `m.layout.FocusedHasPendingSelection()`: se `true` e `msg.String() == "y"` → emitir `msgs.MouseSelectConfirmMsg{PaneID: int(m.layout.FocusedID())}` e retornar; se `"esc"` → emitir `msgs.MouseSelectCancelMsg` e retornar; demais teclas caem no forwarding normal
- [X] T021 [US1] Rodar `go test ./components/terminal/... ./tests/integration/...` — todos os testes T008-T014 devem passar

**Checkpoint**: MVP entregue. Seleção via mouse e cópia funcionais em modo normal (sem mouse tracking da aplicação interna). Validar manualmente quickstart §1 (ambos os modos de `mouse_auto_copy`).

---

## Phase 4: User Story 2 — Seleção com Mouse Quando a Aplicação Interna Usa Rastreamento de Mouse (Priority: P2)

**Goal**: Shift+clique+arrasto intercepta o evento para seleção Lumina mesmo quando a aplicação interna tem mouse tracking ativo, sem interferir com cliques normais (que vão ao PTY como antes).

**Independent Test**: Executar `./lumina`, abrir `vim` com `mouse=a` no painel terminal, verificar que clique simples chega ao vim normalmente; segurar Shift e arrastar → highlight de seleção Lumina aparece, vim não reage; soltar → texto copiado.

### Tests for User Story 2

- [X] T022 [P] [US2] Escrever integration test `TestMouseSelectMsg_ShiftBypassPTYTracking` em `tests/integration/mouse_select_test.go`: construir terminal model com `state.mouseAnyEvent = true` (simula inner app com tracking); enviar `MouseSelectMsg{Press}` com Shift → model atualiza `mouseSelection`; verificar que `PtyMouseMsg` NÃO foi emitido

### Implementation for User Story 2

- [X] T023 [US2] Atualizar `app/app.go` → `handleMouse()`: no bloco de passthrough PTY (onde `FocusedMouseEnabled() == true`), adicionar verificação de `msg.Shift`: se `msg.Shift == true`, INTERCEPTAR e emitir `msgs.MouseSelectMsg` (com pane-local coords) em vez de `msgs.PtyMouseMsg` — o Alt+wheel e outros atalhos existentes permanecem inalterados
- [X] T024 [US2] Rodar `go test ./tests/integration/...` — T022 deve passar; testes de PTY passthrough existentes devem continuar passando (zero regressão)

**Checkpoint**: US1 + US2 independentes e funcionais. Shift+drag funciona com aplicações mouse-aware.

---

## Phase 5: User Story 3 — Usuários Somente-Teclado Mantêm o Copy Mode (Priority: P3)

**Goal**: O copy mode existente permanece completamente funcional; a invariante de exclusão mútua entre `copy mode` e `mouse selection` é enforçada — entrar em um limpa o outro.

**Independent Test**: Acionar copy mode pelo atalho, navegar com h/j/k/l, selecionar com Shift+hjkl, copiar com `y` — comportamento idêntico ao anterior. Se `mouseSelection` estiver ativa antes de entrar no copy mode, ela deve desaparecer.

### Tests for User Story 3

- [X] T025 [P] [US3] Escrever `TestEnterCopyMode_ClearsMouseSelection` em `components/terminal/terminal_test.go`: terminal model com `mouseSelection` ativa recebe `msgs.EnterCopyModeMsg` → após o Update, `m.copy != nil` e `m.mouseSelection == nil`

### Implementation for User Story 3

- [X] T026 [US3] Em `components/terminal/copymode.go` → `enterCopyMode()`: adicionar `m.mouseSelection = nil` antes de inicializar `m.copy` (garantir exclusão mútua)
- [X] T027 [US3] Rodar `go test ./components/terminal/...` — T025 deve passar; todos os testes existentes de copy mode devem continuar passando

**Checkpoint**: Todas as três user stories funcionam e são testáveis independentemente.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Limpeza final, lint, validação end-to-end.

- [X] T028 [P] Rodar `gofmt -w .` e `golangci-lint run` — zero warnings novos
- [X] T029 [P] Rodar `go test ./...` — zero falhas
- [X] T030 Executar validação manual completa do `quickstart.md`: (a) seleção via drag com `mouse_auto_copy=true`; (b) seleção com confirmação `y` com `mouse_auto_copy=false`; (c) Shift+drag em terminal com vim `mouse=a`; (d) copy mode via teclado — sem regressão

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: sem dependências — pode começar imediatamente
- **Foundational (Phase 2)**: depende de Phase 1; BLOQUEIA todas as user stories (novos msgs, config field, layout methods)
- **US1 (Phase 3)**: depende de Phase 2 — implementa o core da feature (MVP)
- **US2 (Phase 4)**: depende de Phase 2; T023 modifica o mesmo `app.handleMouse` de T019 → recomendado após US1
- **US3 (Phase 5)**: depende de Phase 2; independente de US1/US2 (modifica `copymode.go`, arquivo distinto)
- **Polish (Phase 6)**: depende de todas as stories desejadas

### User Story Dependencies

- **US1 (P1)**: start após Phase 2 — core de todo o feature, sem dependências de US2/US3
- **US2 (P2)**: start após Phase 2 — T023 toca `handleMouse` (mesmo arquivo que T019 de US1); sequenciar após US1 para evitar conflito de merge
- **US3 (P3)**: start após Phase 2 — toca `copymode.go` (arquivo distinto de US1/US2), pode rodar em paralelo com US1 e US2

### Within Each User Story

- Tests DEVEM ser escritos e FALHAR antes da implementação correspondente
- `mouseselect.go` completo (T015+T016) antes de `terminal.go` (T017+T018)
- `terminal.go` completo antes de `app.go` (T019+T020)
- Task de smoke `go test` ao final de cada story

### Parallel Opportunities

- Phase 1: T002 e T003 em paralelo após T001
- Phase 2: T005 e T006 em paralelo após T004
- Phase 3 tests: T008–T014 podem ser escritos em paralelo (funções independentes)
- Phase 3 impl: T015+T016 em paralelo com T022 (test de US2 já pode ser escrito)
- Phase 5: US3 (T025–T027) pode rodar em paralelo com US1/US2 (arquivo distinto)
- Phase 6: T028 e T029 em paralelo

---

## Parallel Example: User Story 1

```bash
# Testes de US1 escritos em paralelo:
Task: "TestMouseSelection_StartAndUpdate em components/terminal/mouseselect_test.go"
Task: "TestMouseSelection_FinalizeAutoCopy em components/terminal/mouseselect_test.go"
Task: "TestMouseSelection_FinalizeManualConfirm em components/terminal/mouseselect_test.go"
Task: "TestMouseSelection_ConfirmAndCancel em components/terminal/mouseselect_test.go"
Task: "TestExtractMouseSelection_ReturnsCorrectText em components/terminal/mouseselect_test.go"
Task: "TestMouseSelectMsg_PressMotionRelease em tests/integration/mouse_select_test.go"
Task: "TestMouseSelectConfirmMsg / TestMouseSelectCancelMsg em tests/integration/mouse_select_test.go"
```

---

## Implementation Strategy

### MVP First (US1 apenas)

1. Phase 1 (Setup) — criar stubs
2. Phase 2 (Foundational) — config + msgs + layout methods
3. Phase 3 (US1) — seleção e cópia via mouse em modo normal
4. **STOP e VALIDAR** quickstart §1 manualmente
5. Merge como MVP se entrega incremental for desejada

### Incremental Delivery

1. Phase 1 + Phase 2 → fundação pronta
2. + US1 → mouse selection funcional (MVP!) — cobre 90% dos casos de uso
3. + US2 → Shift+drag para terminals com mouse tracking
4. + US3 → garantia de regressão do copy mode
5. + Polish → lint + validação fim-a-fim

### Parallel Team Strategy (se aplicável)

- Dev A: Phase 2 → US1 (core)
- Dev B: após Phase 2, US3 (`copymode.go` — arquivo distinto)
- Dev A (após US1): US2 (Shift+drag — mesmo `app.go`)

---

## Notes

- [P] = sem dependência bloqueante; pode rodar em qualquer ordem ou em paralelo
- `m.copy` e `m.mouseSelection` NUNCA ambos não-nil — invariante enforçada em T026
- `y` interceptado por `handleKey` APENAS quando `FocusedHasPendingSelection() == true` (T020); caso contrário, vai ao PTY normalmente
- `mouse_auto_copy` é lido de `config.Config` já carregado no startup — nenhum reload em runtime necessário
- Constitution II: 3 novos `tea.Msg` → integration tests obrigatórios (T013, T014, T022)
- Commit sugerido ao final de cada checkpoint (fim de phase)
- Rodar `go test ./...` zero falhas antes de qualquer merge (Constitution II)
