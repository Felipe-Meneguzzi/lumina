# Implementation Plan: Mouse Text Selection in Normal Mode

**Branch**: `005-mouse-select-copy` | **Date**: 2026-04-17 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/005-mouse-select-copy/spec.md`

## Summary

Adicionar seleção de texto com mouse ao painel terminal em modo normal, eliminando a
necessidade de entrar no copy mode para quem usa mouse. O comportamento de cópia é
configurável via `mouse_auto_copy` em `config.toml` (padrão: `true` = cópia automática
ao soltar o mouse; `false` = seleção persiste e usuário confirma com `y`).

Abordagem técnica: novo struct `mouseSelection` em `components/terminal/mouseselect.go`
(análogo ao `copyState` existente), três novos `tea.Msg` em `msgs/msgs.go`, detecção
de Shift+drag em `app.handleMouse` para bypass de PTY passthrough, e interceção de `y`
em `app.handleKey` quando há seleção pendente. O copy mode existente permanece
inalterado como caminho alternativo para usuários somente-teclado.

## Technical Context

**Language/Version**: Go 1.26 (já em uso — `go.mod`)
**Primary Dependencies**: Bubble Tea, Lip Gloss, creack/pty, ultraviolet (charmbracelet/x/vt) — sem dependências novas
**Storage**: `~/.config/lumina/config.toml` — novo campo `mouse_auto_copy bool` na struct `Config`
**Testing**: `go test ./...` — unit tests por pacote + integration tests obrigatórios para os 3 novos `tea.Msg` em `msgs/msgs.go`
**Target Platform**: Linux/macOS (mesma restrição atual do creack/pty)
**Project Type**: Desktop-app / CLI TUI (single binary Go)
**Performance Goals**: Eventos de drag processados em ≤16ms no `Update()`; `renderWithMouseSelection` reutiliza `strings.Builder` sem alocações adicionais no hot path
**Constraints**: Não interferir no passthrough de mouse das aplicações internas; Shift+drag ativo apenas no painel terminal focado; zero regressão no copy mode existente; `Update()` ≤16ms

## Constitution Check

Avaliação contra os quatro princípios de `/.specify/memory/constitution.md`:

| Princípio | Veredicto | Notas |
|-----------|-----------|-------|
| I. Code Quality | ✅ PASS | `mouseselect.go` tem responsabilidade única (seleção via mouse); funções ≤40 linhas; sem estado global; cross-component via `msgs.go` (3 novos tipos). |
| II. Testing Standards | ✅ PASS | Cada função pura ganha unit test isolado; os 3 novos `tea.Msg` exigem integration tests em `tests/integration/` (Constitution II). |
| III. UX Consistency | ✅ PASS | Nenhum novo keybinding global adicionado; o `y` de confirmação é interceptado no terminal, não em `keymap.go` (é comportamento contextual, não atalho global); highlight usa Lip Gloss (mesma `selectionStyle` de `copymode.go`); notificação via `StatusBarNotifyMsg` existente. |
| IV. Performance | ✅ PASS | Drag events são O(1) no `Update()` (só atualiza coordenadas); `renderWithMouseSelection` segue o mesmo padrão de `renderCopyMode` com `strings.Builder`; nenhum I/O bloqueante. |

**Gate**: PASS — prossegue para Phase 0 sem violações.

## Project Structure

### Documentation (this feature)

```text
specs/005-mouse-select-copy/
├── plan.md              # Este arquivo
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── mouse-selection.md  # Contratos dos novos tea.Msg e API pública do terminal
├── checklists/
│   └── requirements.md  # Criado por /speckit.specify
└── tasks.md             # Phase 2 output — criado por /speckit.tasks
```

### Source Code (repository root)

```text
lumina/
├── config/
│   └── config.go                     # [MODIFICADO] campo MouseAutoCopy bool; default true
├── msgs/
│   └── msgs.go                       # [MODIFICADO] +MouseSelectMsg, +MouseSelectConfirmMsg, +MouseSelectCancelMsg
├── components/
│   └── terminal/
│       ├── mouseselect.go            # [NOVO] mouseSelection struct + start/update/finalize/confirm/cancel/extract/render
│       ├── mouseselect_test.go       # [NOVO] unit tests das funções puras de seleção
│       ├── terminal.go               # [MODIFICADO] campo mouseSelection *mouseSelection; handle dos novos Msgs; View()
│       └── terminal_test.go          # [MODIFICADO] cobre MouseSelectMsg, confirm, cancel
├── components/
│   └── layout/
│       └── layout.go                 # [MODIFICADO] +FocusedHasMouseSelection(), +FocusedHasPendingSelection()
├── app/
│   └── app.go                        # [MODIFICADO] handleMouse (Shift+drag → MouseSelectMsg); handleKey (y/esc intercept)
└── tests/
    └── integration/
        └── mouse_select_test.go      # [NOVO] integration tests para os 3 novos tea.Msg
```

**Structure Decision**: Mantida a estrutura `components/*` atual. A lógica de seleção
via mouse ganha arquivo próprio `mouseselect.go` no pacote `terminal` (mesma separação
que `copymode.go`), evitando inflar `terminal.go`. Nenhum pacote novo é criado.

## Complexity Tracking

> Sem violações da Constitution — seção deliberadamente vazia.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| _(nenhuma)_ | — | — |
