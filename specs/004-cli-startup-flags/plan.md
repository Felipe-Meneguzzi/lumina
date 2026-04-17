# Implementation Plan: CLI Startup Flags

**Branch**: `004-cli-startup-flags` | **Date**: 2026-04-17 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/004-cli-startup-flags/spec.md`

## Summary

Adicionar três flags de linha de comando (`-mp`, `-sp`, `-sc`) que parametrizam o boot
do Lumina:

- `-mp N` redefine o teto de painéis (`maxPanes`) para a sessão (default 4).
- `-sp <orient><N>` cria `N` painéis iniciais dispostos horizontalmente (`h`) ou
  verticalmente (`v`) antes do primeiro frame.
- `-sc <cmd>` substitui o shell default pelo comando informado APENAS nos painéis
  iniciais; splits posteriores continuam abrindo o shell default.

Abordagem técnica: parsear as flags em `main.go` (pacote `flag` do stdlib), validar e
converter em uma nova struct `StartupOverrides`, propagar para `config.Config` e
`layout.New`, transformar o `maxPanes` hardcoded em campo do `layout.Model`, e injetar
no `terminal.Model` uma opção `startCommandOverride` usada apenas na criação das folhas
iniciais (não no `handleSplit`). Nenhuma persistência em disco.

## Technical Context

**Language/Version**: Go 1.26 (já em uso no projeto — `go.mod`)
**Primary Dependencies**: Bubble Tea, Lip Gloss, Bubbles, creack/pty, gopsutil/v3 (sem
dependências novas)
**Storage**: N/A — flags são efêmeras e não alteram `config.toml`
**Testing**: `go test ./...` com unit tests por pacote + integration tests em
`tests/integration/` quando um novo `tea.Msg` for adicionado (não será o caso aqui)
**Target Platform**: Linux/macOS (creack/pty não suporta Windows — mesma restrição
atual do produto)
**Project Type**: Desktop-app / CLI TUI (single binary Go)
**Performance Goals**: Parsing + validação das flags em <1s; primeiro frame com N
painéis renderizado antes de qualquer entrada do usuário (FR/SC-002)
**Constraints**: Zero regressão no boot default; `Update()` mantém ≤16ms; `maxPanes`
agora é dado de instância, não mais `const` do pacote
**Scale/Scope**: ~150 linhas novas em `main.go` (flags + parser + mensagens de erro),
~20 linhas de ajuste em `layout.go` (campo + validação), ~15 linhas em `terminal.go`
(override do comando) e atualização de `README.md`

## Constitution Check

Avaliação contra os quatro princípios de `/.specify/memory/constitution.md`:

| Princípio | Veredicto | Notas |
|-----------|-----------|-------|
| I. Code Quality | ✅ PASS | Parser fica em um arquivo novo `cli/flags.go` (responsabilidade única); funções ≤40 linhas; sem estado global mutável — overrides fluem como parâmetro imutável. |
| II. Testing Standards | ✅ PASS | Cada função pura (parser `-sp`, validador `-mp`, resolver de conflitos) terá unit tests; `layout.Model` ganha teste para `maxPanes` configurável; `terminal.Model` ganha teste para start command override. Nenhum novo `tea.Msg` é adicionado → sem obrigação de integration test novo. |
| III. UX Consistency | ✅ PASS | Sem novos keybindings; sem ANSI raw; mensagens de erro de flag vão a stderr (antes da TUI), e a notificação de "teto atingido" em `handleSplit` já usa o sistema existente de `notifyStatus`. |
| IV. Performance | ✅ PASS | Parsing é síncrono em `main()` antes de iniciar o event loop — não impacta render. `handleSplit` continua delegando I/O ao `tea.Cmd` do `terminal.Init`. Criação paralela dos N painéis iniciais reutiliza o mesmo caminho de `newTerminalLeaf` já testado. |

**Gate**: PASS — prossegue para Phase 0 sem violações.

## Project Structure

### Documentation (this feature)

```text
specs/004-cli-startup-flags/
├── plan.md              # Este arquivo
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   └── cli.md           # Contrato das flags (CLI é a interface externa)
├── checklists/
│   └── requirements.md  # Criado por /speckit.specify
└── tasks.md             # Phase 2 output — criado por /speckit.tasks
```

### Source Code (repository root)

```text
lumina/
├── main.go                      # [MODIFICADO] parse flags; monta StartupOverrides; injeta em cfg e layout
├── cli/
│   └── flags.go                 # [NOVO] parser + validação das flags; struct StartupOverrides; func ParseArgs
│   └── flags_test.go            # [NOVO] unit tests de parsing/validação
├── config/
│   └── config.go                # [INALTERADO] — overrides NÃO persistem em config.toml
├── components/
│   ├── layout/
│   │   ├── layout.go            # [MODIFICADO] maxPanes vira campo de Model; New aceita opções de bootstrap
│   │   ├── layout_test.go       # [MODIFICADO] cobre maxPanes custom e boot multi-pane
│   │   └── bootstrap.go         # [NOVO] buildInitialTree(orient, count) retorna PaneNode inicial
│   └── terminal/
│       ├── terminal.go          # [MODIFICADO] New aceita opcional start command override
│       └── terminal_test.go     # [MODIFICADO] cobre start command override
├── msgs/
│   └── msgs.go                  # [INALTERADO] — nenhum tea.Msg novo
└── README.md                    # [MODIFICADO] seção "Uso" descreve as novas flags
```

**Structure Decision**: Mantida a estrutura `components/*` atual (Composite Components
pattern). O parser de CLI ganha um pacote próprio `cli/` — responsabilidade única,
desacoplada do `main` (facilita teste). `layout` recebe um arquivo `bootstrap.go`
dedicado para a construção do tree inicial multi-pane, evitando inflar `layout.go`
(que já está perto do limite razoável). README atualizado conforme pedido do usuário.

## Complexity Tracking

> Sem violações da Constitution — seção deliberadamente vazia.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| _(nenhuma)_ | — | — |
