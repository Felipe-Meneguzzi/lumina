# Implementation Plan: UX Polish Pack

**Branch**: `006-ux-polish-pack` | **Date**: 2026-04-17 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/home/menegas/fpm/lumina/specs/006-ux-polish-pack/spec.md`

## Summary

Pacote de polimento de UX no Lumina reunindo oito melhorias (relógio na status bar, status bar sensível ao terminal focado, correção de render inicial em CLIs TUI, substituição do editor embutido por editor externo configurável, navegação refinada da sidebar com criação de arquivos/pastas, cursor por terminal, foco por clique do mouse em qualquer painel, e estabilidade de render sob alta taxa de saída). Abordagem técnica: reaproveitar a arquitetura Bubble Tea já existente (componentes independentes comunicando-se via `msgs/msgs.go`); introduzir novos `tea.Msg` tipados para cada cross-component concern (clique → foco, troca de foco → contexto da status bar, criar arquivo/pasta); mover o refresh inicial do PTY para garantir que `pty.Setsize` preceda o primeiro read; delegar edição de arquivos a `nano`/`vim`/`nvim` spawned em um painel de terminal; e fazer a status bar derivar seu contexto git do terminal focado em vez de um ticker global.

## Technical Context

**Language/Version**: Go 1.26.1 (conforme `go.mod`)
**Primary Dependencies**: Bubble Tea v1.3.10, Lip Gloss v1.1.0, Bubbles v1.0.0, charmbracelet/ultraviolet + x/vt (emulação de terminal), creack/pty v1.1.24, gopsutil/v3 v3.24.5, BurntSushi/toml v1.6.0
**Storage**: `~/.config/lumina/config.toml` (novos campos `editor string`, opcionais); `~/.config/lumina/keybindings.json` (novas bindings para click-focus, alt+d, alt+f, backspace)
**Testing**: `go test ./...` — unit tests por `tea.Model`, integration tests em `tests/integration/` para cada novo `tea.Msg`
**Target Platform**: Linux e macOS (terminal com suporte a true-color; mouse tracking via SGR 1006)
**Project Type**: Desktop TUI (binário único)
**Performance Goals**: ≥30 FPS no render loop; `Update()` ≤16ms; `pty.Setsize` aplicado em ≤50ms de `tea.WindowSizeMsg`; saída sustentada de 5.000 linhas/min sem degradação visual (SC-002)
**Constraints**: Nenhuma nova dependência externa; todo I/O assíncrono via `tea.Cmd`; sem estado global mutável fora de `config/`; todos os atalhos via `app/keymap.go`; ANSI raw proibido fora de `components/terminal/` e da camada PTY
**Scale/Scope**: ≤32 painéis simultâneos (limite prático do layout tree atual); status bar com atualização contínua (1s para métricas, 30s para relógio, tick on-demand para git do painel focado)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Justification |
|---|---|---|
| I. Code Quality | ✅ PASS | Novas funções planejadas para ≤40 linhas; sem global mutable state fora de `config/`; cross-component apenas via novos `tea.Msg` em `msgs/msgs.go`. |
| II. Testing Standards | ✅ PASS | Cada novo `tea.Msg` recebe integration test em `tests/integration/`; componentes tocados (`statusbar`, `sidebar`, `terminal`, `app`, `layout`) ganham unit tests para novos fluxos; remoção do editor próprio reduz superfície de teste. |
| III. UX Consistency | ✅ PASS | Novas bindings (`alt+d`, `alt+f`, `backspace`, click) definidas em `app/keymap.go` + `config/keybindings.go`; estilos (indicador de foco, glifos git, notificação "Já na raiz") usam Lip Gloss; erros de criação surgem na status bar via `StatusBarNotifyMsg`. |
| IV. Performance Requirements | ✅ PASS | Render do clock usa ticker de 30s (não-bloqueante, `tea.Cmd`); git status do painel focado executado em `tea.Cmd` goroutine; click-to-focus é resolução O(log n) no layout tree, bem dentro dos 16ms de `Update()`; editor externo roda em PTY isolado, não afeta o render loop. |

Sem violações. Complexity Tracking vazio.

## Project Structure

### Documentation (this feature)

```text
specs/006-ux-polish-pack/
├── plan.md              # This file
├── research.md          # Phase 0 output — decisões técnicas e alternativas
├── data-model.md        # Phase 1 output — novas entidades e estados
├── quickstart.md        # Phase 1 output — roteiro de validação manual
├── contracts/
│   ├── msgs.md          # Phase 1 output — contratos de novos tea.Msg
│   └── config.md        # Phase 1 output — contrato de novos campos de config
└── tasks.md             # Phase 2 output (/speckit.tasks — NÃO criado aqui)
```

### Source Code (repository root)

```text
lumina/
├── main.go
├── cli/                           # (inalterado) — flags de startup (004)
├── app/
│   ├── app.go                     # roteamento de novos msgs (click, create, etc.)
│   ├── keymap.go                  # novas bindings: sidebar.NewDir/NewFile; layout.ClickFocus
│   └── app_test.go                # +testes de roteamento dos novos msgs
├── components/
│   ├── terminal/
│   │   ├── terminal.go            # correção de first-render; cursor só se focado
│   │   ├── firstrender.go         # [NOVO] hook de repaint pós-Setsize inicial
│   │   ├── mouse.go               # (inalterado) — mouse tracking já existe
│   │   └── terminal_test.go       # +testes first-render, cursor visibility
│   ├── sidebar/
│   │   ├── sidebar.go             # Backspace→up, Enter em arquivo→OpenFileMsg
│   │   ├── create.go              # [NOVO] prompt inline para NewDir/NewFile
│   │   └── sidebar_test.go        # +testes de navegação e criação
│   ├── statusbar/
│   │   ├── statusbar.go           # clock, glifo git, notificação temporária
│   │   └── statusbar_test.go      # +testes clock + git glifo por painel focado
│   ├── layout/
│   │   ├── layout.go              # resolver click em (x,y) → PaneID focável
│   │   ├── focus.go               # (inalterado, mas consumidor de ClickFocusMsg)
│   │   └── layout_test.go         # +teste hit-test de clique
│   └── editor/                    # [REMOVIDO] — substituído por editor externo
├── msgs/
│   └── msgs.go                    # +ClickFocusMsg, +SidebarCreateMsg, +OpenInExternalEditorMsg, +FocusedPaneContextMsg
├── config/
│   ├── config.go                  # +campo Editor (nano/vim/nvim); default "nano"
│   └── keybindings.go             # +entries alt+d, alt+f, backspace, click_focus
└── tests/
    └── integration/
        ├── click_focus_test.go    # [NOVO]
        ├── sidebar_create_test.go # [NOVO]
        ├── external_editor_test.go# [NOVO]
        ├── statusbar_focus_test.go# [NOVO]
        └── first_render_test.go   # [NOVO]
```

**Structure Decision**: Manter a estrutura monolítica existente (Option 1 — single project). O único componente removido é `components/editor/` (substituído por spawn de editor externo dentro de um terminal). Novos arquivos aderem ao padrão "um arquivo por concern lógico" já em uso.

## Complexity Tracking

> Nenhuma violação de constitution detectada; esta seção é intencionalmente vazia.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
