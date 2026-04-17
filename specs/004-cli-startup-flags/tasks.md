---

description: "Task list for feature 004-cli-startup-flags"
---

# Tasks: CLI Startup Flags

**Input**: Design documents from `/specs/004-cli-startup-flags/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli.md, quickstart.md

**Tests**: Incluídos. A Constitution de Lumina (Princípio II) exige unit tests para cada função exportada e para cada `tea.Model`. Os test tasks abaixo refletem essa exigência.

**Organization**: Agrupado por user story (US1 = `-mp`, US2 = `-sp`, US3 = `-sc`). MVP = US1.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Paralelizável (arquivos distintos, sem dependência de tarefas incompletas)
- **[Story]**: US1 / US2 / US3 — mapeia para a user story da spec
- Todos os paths são absolutos a partir da raiz do repo

## Path Conventions

Single project Go. Raiz do repo = `/home/menegas/fpm/lumina/`. Código em packages diretos (`cli/`, `components/layout/`, `components/terminal/`, `main.go`).

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Criar a estrutura mínima do novo pacote `cli/` antes de qualquer lógica.

- [X] T001 Criar diretório `cli/` com arquivo stub `cli/flags.go` contendo apenas `package cli` e comentário do pacote (`// Package cli parses Lumina startup flags.`)
- [X] T002 [P] Criar arquivo stub `cli/flags_test.go` com `package cli` + import de `testing` (nenhum teste ainda — evita breakage no `go test ./...`)
- [X] T003 [P] Criar arquivo stub `components/layout/bootstrap.go` com `package layout` + comentário do arquivo

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Estruturas compartilhadas que TODAS as user stories consomem. Sem elas, nenhuma story compila.

**⚠️ CRITICAL**: Concluir integralmente antes de começar qualquer user story.

- [X] T004 Definir tipo `Orient` (enum com `OrientNone`/`OrientHorizontal`/`OrientVertical`) e struct `StartupOverrides` (campos: `MaxPanes`, `StartPanes`, `StartOrient`, `StartCommand`, `FilePath`) em `cli/flags.go`
- [X] T005 Adicionar método `(o StartupOverrides) EffectiveMaxPanes() int` em `cli/flags.go` — retorna `o.MaxPanes` se >0, senão `max(4, o.StartPanes)` (default 4 com auto-raise)
- [X] T006 [P] Escrever unit tests em `cli/flags_test.go` para `EffectiveMaxPanes` cobrindo: `-mp` ausente + `StartPanes=0` → 4; `-mp=10` → 10; `-mp=0 + StartPanes=5` → 5; `-mp=3 + StartPanes=2` → 3
- [X] T007 Migrar `const maxPanes = 4` em `components/layout/layout.go` para campo `maxPanes int` de `Model`; inicializar para 4 no `New`; substituir todas as referências `maxPanes` por `m.maxPanes` (inclui a mensagem de warning em `handleSplit`, agora com `fmt.Sprintf`)
- [X] T008 Introduzir functional options em `components/layout/layout.go`: tipo `Option func(*Model)`, funções `WithMaxPanes(n int) Option`, `WithStartCommand(cmd string) Option`, `WithInitialLayout(orient cli.Orient, count int) Option`; atualizar assinatura de `New` para `New(cfg config.Config, opts ...Option) (Model, error)`
- [X] T009 [P] Ajustar `components/layout/layout_test.go`: garantir que os testes existentes continuam passando com a nova assinatura `New(cfg)` (sem opts); adicionar `TestLayoutNew_DefaultMaxPanesIs4` e `TestLayoutNew_WithMaxPanes_10`
- [X] T010 [P] Adicionar chamadas a `cli.ParseArgs(os.Args[1:]) (StartupOverrides, error)` em `main.go` (função ainda stub que retorna zero-value) para habilitar a fiação progressiva nas próximas stories; exibir erro em stderr + `os.Exit(2)` quando `err != nil`

**Checkpoint**: Código compila e todos os testes passam; comportamento default preservado (boot sem flags idêntico ao atual).

---

## Phase 3: User Story 1 — Override max panes at startup (Priority: P1) 🎯 MVP

**Goal**: `lumina -mp <N>` define o teto de painéis da sessão; splits respeitam o novo limite; default (4) permanece intacto quando a flag está ausente.

**Independent Test**: Rodar `./lumina -mp 10`, abrir painéis via atalho de split até o 10º; o 11º deve ser rejeitado com notificação. Rodar `./lumina` (sem flag) — 5º split segue rejeitado.

### Tests for User Story 1

> **NOTE: Escrever testes primeiro e garantir que FALHAM antes da implementação.**

- [X] T011 [P] [US1] Escrever teste `TestParseArgs_MPFlag_Valid` em `cli/flags_test.go` cobrindo `-mp 10` → `overrides.MaxPanes == 10`
- [X] T012 [P] [US1] Escrever teste `TestParseArgs_MPFlag_Invalid` em `cli/flags_test.go` para casos `-mp 0`, `-mp -1`, `-mp abc` → retorna erro com mensagem incluindo o valor recebido
- [X] T013 [P] [US1] Escrever teste `TestPaneSplitMsg_AtCustomMax_IsNoop` em `components/layout/layout_test.go` — layout criado com `WithMaxPanes(2)` rejeita o 3º split

### Implementation for User Story 1

- [X] T014 [US1] Implementar parser real de `-mp` em `cli/flags.go` usando `flag.IntVar`: lê o valor, valida `>= 1`, preenche `StartupOverrides.MaxPanes`; mensagens de erro no formato `lumina: -mp inválido: esperado inteiro >= 1, recebi "<val>"`
- [X] T015 [US1] Fiação em `main.go`: aplicar `layout.WithMaxPanes(overrides.EffectiveMaxPanes())` na chamada a `app.New` (ou repasse via `config.Config` se necessário) — confirmar que `app.New` propaga para `layout.New`
- [X] T016 [US1] Atualizar a mensagem de warning em `handleSplit` (`components/layout/layout.go`) para usar `m.maxPanes` formatado: `fmt.Sprintf("Máximo de %d painéis atingido", m.maxPanes)`
- [X] T017 [US1] Rodar `go test ./cli/... ./components/layout/...` — todos os testes de T011-T013 devem agora passar

**Checkpoint**: MVP entregue. `-mp` funcional e default preservado. Validar manualmente quickstart §2.1 e §2.2.

---

## Phase 4: User Story 2 — Start with a pre-split layout (Priority: P2)

**Goal**: `lumina -sp h<N>` / `-sp v<N>` abre diretamente com N painéis dispostos horizontalmente ou verticalmente, já no primeiro frame renderizado.

**Independent Test**: Rodar `./lumina -sp h3` → primeiro frame mostra 3 painéis lado-a-lado. Rodar `./lumina -sp v2` → 2 painéis empilhados. Rodar `./lumina -sp h5` sem `-mp` → teto efetivo sobe para 5.

### Tests for User Story 2

- [X] T018 [P] [US2] Escrever teste `TestParseStartPanes_Valid` em `cli/flags_test.go` para `h1`, `h3`, `v2`, `v99` — retorna `(orient, count, nil)` corretos
- [X] T019 [P] [US2] Escrever teste `TestParseStartPanes_Invalid` em `cli/flags_test.go` para `3`, `h`, `d2`, `h0`, `h-1`, `""`, `hABC` — retorna erro
- [X] T020 [P] [US2] Escrever teste `TestValidate_Conflict_MPBelowSP` em `cli/flags_test.go` — `MaxPanes=2, StartPanes=5` → erro com mensagem contendo "excede"
- [X] T021 [P] [US2] Escrever teste `TestLayoutNew_WithInitialLayoutH3` em `components/layout/layout_test.go` — `layout.New(cfg, WithInitialLayout(OrientHorizontal, 3))` resulta em `PaneCount() == 3` e todos os panes em subtree horizontal
- [X] T022 [P] [US2] Escrever teste `TestLayoutNew_WithInitialLayoutV2` em `components/layout/layout_test.go` — análogo para vertical

### Implementation for User Story 2

- [X] T023 [US2] Implementar função pura `ParseStartPanes(s string) (Orient, int, error)` em `cli/flags.go` — valida formato `^[hv][1-9][0-9]*$` (sem regex se preferir; ok usar `strings.HasPrefix` + `strconv.Atoi`); mensagens conforme contract
- [X] T024 [US2] Implementar `(o StartupOverrides) Validate() error` em `cli/flags.go` — detecta conflito explícito `MaxPanes > 0 && StartPanes > MaxPanes` com mensagem do contract; trata `StartPanes == 1` como no-op (equivalente ao default)
- [X] T025 [US2] Estender o parser de `main.go` para ler `-sp` via `flag.StringVar`, chamar `ParseStartPanes`, popular `StartPanes`/`StartOrient`; chamar `Validate()` após `flag.Parse()`
- [X] T026 [US2] Implementar `buildInitialTree(cfg config.Config, orient Orient, count int, startCommand string) (PaneNode, PaneID, error)` em `components/layout/bootstrap.go` — cria a folha raiz + aplica `count-1` `splitLeaf` na direção solicitada, reutilizando funções internas existentes
- [X] T027 [US2] Modificar `layout.New` para consumir `WithInitialLayout` e `WithStartCommand` — se `WithInitialLayout` estiver definido com `count > 1`, delega para `buildInitialTree`; caso contrário, mantém o caminho atual (folha única)
- [X] T028 [US2] Fiação em `main.go`: aplicar `layout.WithInitialLayout(overrides.StartOrient, overrides.StartPanes)` quando `StartPanes > 0`
- [X] T029 [US2] Rodar `go test ./cli/... ./components/layout/...` — todos os testes T018-T022 devem passar

**Checkpoint**: US1 + US2 independentes e funcionais. Validar manualmente quickstart §2.3 e §2.4 (sem `-sc` por enquanto).

---

## Phase 5: User Story 3 — Autorun a command in initial panes (Priority: P3)

**Goal**: `lumina -sc <cmd>` executa `<cmd>` nos painéis criados pela flag `-sp` (ou no único painel default, se `-sp` ausente). Splits manuais posteriores abrem o shell default, NÃO o comando.

**Independent Test**: Rodar `./lumina -sp h3 -sc claude` → cada um dos 3 painéis mostra `claude` rodando. Ao fazer split manual (atalho), o 4º painel mostra o shell default (prompt normal), sem `claude`.

### Tests for User Story 3

- [X] T030 [P] [US3] Escrever teste `TestParseArgs_SCFlag_Valid` em `cli/flags_test.go` — `-sc claude` e `-sc "claude --model opus"` populam `StartCommand` como string única
- [X] T031 [P] [US3] Escrever teste `TestParseArgs_SCFlag_Empty` em `cli/flags_test.go` — `-sc ""` retorna erro
- [X] T032 [P] [US3] Escrever teste `TestTerminalNewWithCommand_UsesOverride` em `components/terminal/terminal_test.go` — `NewWithCommand(cfg, "/bin/true")` inicia um processo com o override (verificar via `m.cmd.Path` ou equivalente sem quebrar o contrato de caixa-preta)
- [X] T033 [P] [US3] Escrever teste `TestLayoutHandleSplit_DoesNotInheritStartCommand` em `components/layout/layout_test.go` — criar layout com `WithStartCommand("echo test")` + `WithInitialLayout(OrientHorizontal, 2)`; simular `PaneSplitMsg`; verificar que o novo leaf foi criado por `newTerminalLeaf` (shell default), não pelo override

### Implementation for User Story 3

- [X] T034 [US3] Estender o parser em `main.go` para ler `-sc` via `flag.StringVar`; validar não-vazio quando a flag é informada
- [X] T035 [US3] Adicionar campo `shellOverride string` a `terminal.Model` em `components/terminal/terminal.go`; adicionar construtor `NewWithCommand(cfg config.Config, command string) (Model, error)` que seta `shellOverride` antes de `startShell`
- [X] T036 [US3] Modificar `startShell` em `components/terminal/terminal.go`: se `m.shellOverride != ""`, substituir `buildShellCommand(...)` por `exec.Command("sh", "-c", m.shellOverride)` com o mesmo env/pty setup
- [X] T037 [US3] Criar função auxiliar `newTerminalLeafWithCommand(id PaneID, cfg config.Config, command string) (*LeafNode, error)` em `components/layout/layout.go` (simétrica a `newTerminalLeaf`, usa `terminal.NewWithCommand`)
- [X] T038 [US3] Atualizar `buildInitialTree` em `components/layout/bootstrap.go` para usar `newTerminalLeafWithCommand` quando `startCommand != ""`; confirmar que `handleSplit` NÃO consulta `m.startCommand` (garante FR-010)
- [X] T039 [US3] Fiação em `main.go`: aplicar `layout.WithStartCommand(overrides.StartCommand)` quando `StartCommand != ""`
- [X] T040 [US3] Rodar `go test ./cli/... ./components/layout/... ./components/terminal/...` — todos os testes T030-T033 devem passar

**Checkpoint**: Todas as três user stories funcionam e são testáveis independentemente. Validar quickstart §2.5 e §2.6.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Ajustes finais — ajuda do CLI, README, validação fim-a-fim e lint.

- [X] T041 [P] Sobrescrever `flag.Usage` em `main.go` para imprimir o bloco de ajuda do contract (`contracts/cli.md` seção "Mensagem de --help esperada") com exemplo `lumina -mp 10 -sp h3 -sc claude`
- [X] T042 [P] Atualizar `README.md`: adicionar seção "CLI flags" após a seção de uso atual, listando `-mp`, `-sp`, `-sc` com defaults e os exemplos canônicos da tabela do `contracts/cli.md`
- [X] T043 Rodar `gofmt -w .` e `golangci-lint run` — zero warnings novos
- [X] T044 Rodar `go test ./...` — zero falhas
- [X] T045 Executar validação manual completa do `quickstart.md` §§ 2.1–2.6, 3 e 4: confirmar todos os comportamentos listados, incluindo exit codes 2 nos casos de erro e continuidade da sessão no caso de comando inexistente

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: sem dependências, pode começar imediatamente
- **Foundational (Phase 2)**: depende de Phase 1; BLOQUEIA todas as user stories (maxPanes migration, StartupOverrides struct, parser skeleton)
- **US1 (Phase 3)**: depende de Phase 2
- **US2 (Phase 4)**: depende de Phase 2 (usa `EffectiveMaxPanes` + layout options)
- **US3 (Phase 5)**: depende de Phase 2 e reutiliza `buildInitialTree` de US2 (T038 referencia artefato introduzido em T026) — executar após US2 para minimizar conflitos de merge em `bootstrap.go`
- **Polish (Phase 6)**: depende de todas as stories desejadas estarem completas

### User Story Dependencies

- **US1 (P1)**: start após Phase 2; nenhum acoplamento com US2/US3
- **US2 (P2)**: start após Phase 2; independente de US1 no plano (usa mesma fundação `WithMaxPanes`, mas não depende de parser de `-mp` concluído)
- **US3 (P3)**: start após Phase 2; reutiliza `buildInitialTree` de US2 → recomendado sequenciar US2 → US3 em equipe pequena; em equipe maior, dá para paralelizar desde que T038 aguarde T026

### Parallel Opportunities

- Phase 1: T002 e T003 em paralelo após T001
- Phase 2: T009 e T010 em paralelo após T007+T008
- Phase 3 tests: T011, T012, T013 em paralelo
- Phase 4 tests: T018–T022 em paralelo
- Phase 5 tests: T030–T033 em paralelo
- Phase 6: T041 e T042 em paralelo

---

## Parallel Example: User Story 2

```bash
# Testes de US2 em paralelo (arquivos diferentes ou seções independentes):
Task: "TestParseStartPanes_Valid em cli/flags_test.go"
Task: "TestParseStartPanes_Invalid em cli/flags_test.go"
Task: "TestValidate_Conflict_MPBelowSP em cli/flags_test.go"
Task: "TestLayoutNew_WithInitialLayoutH3 em components/layout/layout_test.go"
Task: "TestLayoutNew_WithInitialLayoutV2 em components/layout/layout_test.go"
```

---

## Implementation Strategy

### MVP First (US1 apenas)

1. Phase 1 (Setup) — criar stubs
2. Phase 2 (Foundational) — migrar maxPanes para campo + StartupOverrides struct
3. Phase 3 (US1) — `-mp` funcional
4. **STOP e VALIDAR** quickstart §2.2
5. Merge como MVP se a entrega em etapas for desejada

### Incremental Delivery

1. Phase 1 + Phase 2 → fundação pronta
2. + US1 → `-mp` entregue (MVP)
3. + US2 → `-sp` entregue
4. + US3 → `-sc` entregue
5. + Polish → README + `--help` + lint + validação manual

### Parallel Team Strategy

- Dev A: Phase 2 → US1
- Dev B: após Phase 2, US2 (depende de `WithInitialLayout` + `buildInitialTree`)
- Dev C: após US2 T026, US3 (depende de `buildInitialTree` existir)
- Polish executada por quem terminar primeiro

---

## Notes

- [P] = arquivos distintos, sem dependências bloqueantes; pode rodar em paralelo
- Commits sugeridos a cada checkpoint (fim de fase)
- `handleSplit` NÃO deve ler `m.startCommand` — T038 verifica isso explicitamente e T033 cobre no teste
- Nenhum `tea.Msg` novo é adicionado → sem obrigação de integration test (Constitution II)
- Reforço: atualização do README é obrigatória (pedido explícito do usuário) — T042
