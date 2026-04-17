# Implementation Plan: Lumina TUI Core

**Branch**: `001-lumina-core` | **Date**: 2026-04-16 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/001-lumina-core/spec.md`

## Summary

Implementação do Lumina — editor TUI em Go com Bubble Tea que combina painel de terminal
interativo (PTY via creack/pty), explorador de arquivos (sidebar), editor de texto simples
e status bar com métricas em tempo real (gopsutil). Arquitetura de Componentes Compostos
(Pattern B): cada painel implementa `tea.Model` completo, composto pelo modelo raiz `app.Model`
via delegação explícita e mensagens tipadas em `msgs/msgs.go`.

## Technical Context

**Language/Version**: Go 1.26
**Primary Dependencies**: Bubble Tea, Lip Gloss, Bubbles (Charm), creack/pty, gopsutil/v3, BurntSushi/toml
**Storage**: Sistema de arquivos (`os.ReadFile` / `os.WriteFile`)
**Testing**: `go test ./...` — unitário por componente, integração por message flow
**Target Platform**: Linux / macOS (PTY — sem suporte a Windows em v1)
**Project Type**: TUI desktop application (binário único, sem runtime externo)
**Performance Goals**: ≥30 FPS render loop, startup <500ms, PTY resize ≤50ms, status bar tick 1s
**Constraints**: `Update()` ≤16ms, `View()` sem bloqueio I/O, writes PTY síncronos OK
**Scale/Scope**: Usuário único, janela única, até 10k linhas por arquivo no editor

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Gate | Status |
|-----------|------|--------|
| I. Code Quality | Pacotes com responsabilidade única | ✅ `terminal/`, `sidebar/`, `editor/`, `statusbar/`, `msgs/`, `config/` |
| I. Code Quality | Cross-component via `msgs/msgs.go` exclusivamente | ✅ Contratos em contracts/messages.md |
| II. Testing Standards | Unit tests isolados por `tea.Model` | ✅ Planejado por componente |
| II. Testing Standards | Integration tests por `tea.Msg` customizado | ✅ 8 tipos de Msg com contratos definidos |
| III. UX Consistency | Keybindings centralizados em `app/keymap.go` | ✅ KeyMap struct definida |
| III. UX Consistency | Estilos via Lip Gloss (sem ANSI direto) | ✅ Exceto camada PTY raw (permitido) |
| IV. Performance | `Update()` ≤16ms; I/O em goroutines via `tea.Cmd` | ✅ PTY reads e ticker são Cmds assíncronos |
| IV. Performance | Status bar ticker ≥1s em background | ✅ `tickMetrics(1*time.Second)` |

**Constitution Check pós-design**: Nenhuma violação identificada. Writes de input ao PTY
são síncronos mas rápidos (bytes diretos) — não violam o limite de 16ms em condições normais.

## Project Structure

### Documentation (this feature)

```text
specs/001-lumina-core/
├── plan.md                        # Este arquivo
├── research.md                    # Phase 0 — padrões e decisões técnicas
├── data-model.md                  # Phase 1 — entidades, campos, layout
├── quickstart.md                  # Phase 1 — guia de validação manual
├── contracts/
│   ├── messages.md                # Contratos de msgs/msgs.go (8 Msg types)
│   └── component-interfaces.md   # Contratos tea.Model por componente
└── tasks.md                       # Phase 2 output (/speckit.tasks — não criado aqui)
```

### Source Code (repository root)

```text
lumina/
├── main.go                  # Entrypoint: parse args, load config, run tea.Program
├── app/
│   ├── app.go               # Model raiz: compõe e roteia mensagens
│   └── keymap.go            # KeyMap struct — todos os key.Binding centralizados
├── components/
│   ├── terminal/
│   │   ├── terminal.go      # Model, Init, Update, View + PTY lifecycle
│   │   └── terminal_test.go # Unit tests isolados (mock PTY)
│   ├── sidebar/
│   │   ├── sidebar.go       # Model + bubbles/list wrapper
│   │   └── sidebar_test.go
│   ├── editor/
│   │   ├── editor.go        # Model, Init, Update, View
│   │   ├── buffer.go        # Operações no []string buffer (insert, delete, move)
│   │   └── editor_test.go
│   └── statusbar/
│       ├── statusbar.go     # Model + gopsutil ticker
│       └── statusbar_test.go
├── msgs/
│   └── msgs.go              # Todos os tea.Msg customizados (8 types)
├── config/
│   └── config.go            # Config struct + LoadConfig() com TOML + defaults
└── tests/
    └── integration/
        └── messages_test.go # Integration tests: message flows entre componentes
```

**Structure Decision**: Single project (Option 1). TUI pura sem frontend/backend.
`tests/integration/` separado de `components/*/` para clareza de escopo.

## Complexity Tracking

> Sem violações da constituição que requeiram justificativa.
