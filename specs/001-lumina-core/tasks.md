---

description: "Task list for Lumina TUI Core"
---

# Tasks: Lumina TUI Core

**Input**: Design documents from `specs/001-lumina-core/`
**Prerequisites**: plan.md ✅, spec.md ✅, data-model.md ✅, research.md ✅, contracts/ ✅

**Tests**: Included — unit tests por componente `tea.Model`, integration tests para message flows.

**Organization**: Tasks agrupadas por user story para implementação e teste independentes.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Pode rodar em paralelo (arquivos diferentes, sem dependências)
- **[Story]**: US1=Terminal, US2=Sidebar, US3=Editor, US4=StatusBar

---

## Phase 1: Setup

**Purpose**: Inicialização do projeto Go e estrutura de diretórios

- [x] T001 Criar `go.mod` com `go 1.26` e adicionar dependências: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, `github.com/charmbracelet/bubbles`, `github.com/creack/pty`, `github.com/shirou/gopsutil/v3`, `github.com/BurntSushi/toml`
- [x] T002 Criar estrutura de diretórios: `app/`, `components/terminal/`, `components/sidebar/`, `components/editor/`, `components/statusbar/`, `msgs/`, `config/`, `tests/integration/`
- [x] T003 [P] Criar `.golangci.yml` na raiz com ruleset padrão e rodar `golangci-lint run` para validar configuração

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Infraestrutura compartilhada que TODOS os componentes dependem

**⚠️ CRÍTICO**: Nenhuma user story pode começar antes desta fase estar completa

- [x] T004 Criar `msgs/msgs.go` com todos os 8 tipos `tea.Msg` definidos em `specs/001-lumina-core/contracts/messages.md`: `FocusChangeMsg`, `PtyOutputMsg`, `PtyInputMsg`, `TerminalResizeMsg`, `SidebarResizeMsg`, `EditorResizeMsg`, `MetricsTickMsg`, `OpenFileMsg`, `ConfirmCloseMsg`, `CloseConfirmedMsg`, `CloseAbortedMsg`, `StatusBarNotifyMsg` e os tipos auxiliares (`FocusTarget`, `NotifyLevel`)
- [x] T005 [P] Criar `config/config.go` com struct `Config` (campos: `Shell string`, `MetricsInterval int`, `ShowHidden bool`, `SidebarWidth int`, `Theme string`) e função `LoadConfig() (Config, error)` que lê `~/.config/lumina/config.toml` com fallback para defaults embutidos via `BurntSushi/toml`
- [x] T006 [P] Criar `app/keymap.go` com struct `KeyMap` e bindings via `charmbracelet/bubbles/key`: `FocusTerminal` (Ctrl+1), `FocusSidebar` (Ctrl+2), `FocusEditor` (Ctrl+3), `Save` (Ctrl+S), `Quit` (Ctrl+C), `Help` (?)
- [x] T007 Criar `main.go` com entrypoint: parse `os.Args` (arquivo opcional), chamar `config.LoadConfig()`, instanciar `app.New(cfg)` e executar `tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()`

**Checkpoint**: Fundação pronta — implementação das user stories pode começar

---

## Phase 3: User Story 1 — Terminal Interativo (Priority: P1) 🎯 MVP

**Goal**: Painel de terminal funcional com PTY real, suporte a programas interativos e resize dinâmico

**Independent Test**: Abrir o Lumina → executar `echo hello` → saída aparece → executar `htop` → interativo → redimensionar janela → sem corrupção

### Tests for User Story 1 ⚠️

> **ESCREVER ANTES DA IMPLEMENTAÇÃO — devem falhar primeiro**

- [x] T008 [P] [US1] Escrever unit tests em `components/terminal/terminal_test.go`: testar `Update(PtyOutputMsg)` adiciona ao viewport, `Update(TerminalResizeMsg)` atualiza width/height, `View()` retorna string com dimensões corretas, shell auto-restart ao receber EOF

### Implementation for User Story 1

- [x] T009 [P] [US1] Criar `components/terminal/terminal.go` com struct `Model` (campos: `pty *os.File`, `cmd *exec.Cmd`, `viewport viewport.Model`, `width int`, `height int`, `focused bool`) e construtor `New(cfg config.Config) (Model, error)` que inicia processo `$SHELL` via `creack/pty`
- [x] T010 [US1] Implementar `Init()` em `components/terminal/terminal.go` retornando `waitForPtyOutput(m.pty)` — função `tea.Cmd` que lê do PTY em goroutine e emite `msgs.PtyOutputMsg`
- [x] T011 [US1] Implementar `Update(msg tea.Msg)` em `components/terminal/terminal.go`: `PtyOutputMsg` → adicionar ao viewport e re-enfileirar `waitForPtyOutput`; `PtyInputMsg` → escrever bytes no `m.pty`; `TerminalResizeMsg` → chamar `pty.Setsize` e atualizar viewport
- [x] T012 [US1] Implementar shell auto-restart em `components/terminal/terminal.go`: quando `PtyOutputMsg.Err != nil` (EOF/processo morto), criar novo `exec.Cmd` com `$SHELL`, iniciar novo PTY e retornar `waitForPtyOutput` do novo PTY
- [x] T013 [US1] Implementar `View()` em `components/terminal/terminal.go`: retornar `m.viewport.View()` com borda Lip Gloss destacada se `m.focused == true`, borda simples caso contrário
- [x] T014 [US1] Criar `app/app.go` com struct `AppModel` (campos: `terminal terminal.Model`, `statusbar statusbar.Model`, `sidebar sidebar.Model`, `editor editor.Model`, `focus msgs.FocusTarget`, `width int`, `height int`), construtor `New(cfg config.Config) (AppModel, error)` e `Init()` retornando `tea.Batch` dos Init de todos os filhos
- [x] T015 [US1] Implementar `Update(tea.WindowSizeMsg)` em `app/app.go`: calcular dimensões de cada painel e emitir `msgs.TerminalResizeMsg`, `msgs.SidebarResizeMsg`, `msgs.EditorResizeMsg`, `msgs.StatusBarResizeMsg` (adicionar tipo a msgs.go)
- [x] T016 [US1] Implementar roteamento de `tea.KeyMsg` em `app/app.go`: atalhos globais do `KeyMap` processados aqui; quando `focus == FocusTerminal` converter para `msgs.PtyInputMsg` e delegar ao terminal
- [x] T017 [US1] Implementar `View()` em `app/app.go` (versão US1): `lipgloss.JoinVertical` de `terminal.View()` + `statusbar.View()` com layout de altura correta

**Checkpoint**: US1 completa — terminal funcional, testável de forma independente

---

## Phase 4: User Story 4 — Status Bar com Métricas (Priority: P2)

**Goal**: Barra inferior com CPU%, memória e branch git atualizando a cada 1s sem impactar o loop

**Independent Test**: Abrir Lumina → observar status bar 3s → métricas mudam → rodar `yes > /dev/null &` → CPU aumenta na próxima atualização

### Tests for User Story 4 ⚠️

- [x] T018 [P] [US4] Escrever unit tests em `components/statusbar/statusbar_test.go`: `Update(MetricsTickMsg)` atualiza campos, `Update(StatusBarNotifyMsg)` sobrescreve display temporariamente, `View()` trunca ao `m.width`

### Implementation for User Story 4

- [x] T019 [P] [US4] Criar `components/statusbar/statusbar.go` com struct `Model` (campos: `cpu float64`, `memUsed uint64`, `memTotal uint64`, `cwd string`, `gitBranch string`, `width int`, `notify *StatusNotify`) e construtor `New(cfg config.Config) Model`
- [x] T020 [US4] Implementar `Init()` em `components/statusbar/statusbar.go` retornando `tickMetrics(time.Duration(cfg.MetricsInterval) * time.Millisecond)` — `tea.Tick` que coleta `cpu.Percent`, `mem.VirtualMemory` via gopsutil e detecta branch git via `exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")`
- [x] T021 [US4] Implementar `Update(msg tea.Msg)` em `components/statusbar/statusbar.go`: `MetricsTickMsg` → atualizar campos e retornar próximo `tickMetrics`; `StatusBarNotifyMsg` → armazenar notificação com timer; resize → atualizar `m.width`
- [x] T022 [US4] Implementar `View()` em `components/statusbar/statusbar.go`: string de 1 linha `"  CPU: X.X%  MEM: X.X/XGB  [branch]  ~/dir  "` com Lip Gloss, truncada em `m.width` colunas
- [x] T023 [US4] Integrar statusbar em `app/app.go`: adicionar ao `AppModel`, incluir `statusbar.Init()` no `tea.Batch` de `App.Init()`, rotear `MetricsTickMsg` para statusbar em `Update()`

**Checkpoint**: US1 + US4 funcionam independentemente

---

## Phase 5: User Story 2 — Explorador de Arquivos (Priority: P2)

**Goal**: Sidebar de navegação hierárquica por arquivos via teclado, abrindo arquivos no editor

**Independent Test**: Ctrl+2 → foco na sidebar → navegar com setas → Enter em diretório expande → Enter em arquivo emite OpenFileMsg

### Tests for User Story 2 ⚠️

- [x] T024 [P] [US2] Escrever unit tests em `components/sidebar/sidebar_test.go`: `Update(tea.KeyMsg{down})` avança seleção, `Update(tea.KeyMsg{enter})` em diretório emite expand e recarrega, `Update(tea.KeyMsg{enter})` em arquivo emite `msgs.OpenFileMsg`, `View()` retorna `""` quando `width == 0`

### Implementation for User Story 2

- [x] T025 [P] [US2] Criar `components/sidebar/sidebar.go` com struct `Model` (campos: `list list.Model`, `root string`, `cwd string`, `expanded map[string]bool`, `focused bool`, `width int`, `height int`) e construtor `New(root string, cfg config.Config) Model`
- [x] T026 [US2] Implementar `Init()` em `components/sidebar/sidebar.go`: carregar entries do `root` via `os.ReadDir`, popular `list.Model` com itens formatados (prefixo `▸`/`▾` para dirs, espaço para arquivos)
- [x] T027 [US2] Implementar `Update(msg tea.Msg)` em `components/sidebar/sidebar.go`: delegar `tea.KeyMsg` ao `list.Model`; ao Enter detectar se item é dir (toggle `expanded`, recarregar filhos) ou arquivo (emitir `msgs.OpenFileMsg{Path: path}`); `SidebarResizeMsg` → atualizar width/height
- [x] T028 [US2] Implementar `View()` em `components/sidebar/sidebar.go`: retornar `list.View()` com borda Lip Gloss de foco; retornar `""` se `m.width == 0` (modo terminal estreito <80 cols)
- [x] T029 [US2] Integrar sidebar em `app/app.go`: adicionar ao `AppModel`, rotear `Ctrl+2` para `msgs.FocusChangeMsg{Target: msgs.FocusSidebar}`, rotear `msgs.OpenFileMsg` para `editor.Model.Update()`; atualizar `View()` com `lipgloss.JoinHorizontal(sidebar, painel_ativo)`

**Checkpoint**: US1 + US4 + US2 funcionam independentemente

---

## Phase 6: User Story 3 — Editor de Texto Simples (Priority: P3)

**Goal**: Edição de arquivos com inserção/deleção de texto, salvamento e confirmação de fechamento

**Independent Test**: `lumina arquivo.txt` → editor focado → digitar texto → Ctrl+S → verificar disco → editar sem salvar → Ctrl+W → dialog de confirmação aparece

### Tests for User Story 3 ⚠️

- [x] T030 [P] [US3] Escrever unit tests em `components/editor/editor_test.go`: inserir caractere move cursor, Backspace remove caractere, Enter insere nova linha, `dirty` torna-se `true` após edição, `dirty` retorna `false` após save, `Update(ConfirmCloseMsg)` emite `CloseConfirmedMsg`
- [x] T031 [P] [US3] Escrever unit tests em `components/editor/buffer_test.go`: `InsertAt(row, col, ch)`, `DeleteAt(row, col)`, `SplitLine(row, col)` (Enter), cursor boundary checks (`Row` e `Col` nunca saem dos bounds)

### Implementation for User Story 3

- [x] T032 [P] [US3] Criar `components/editor/buffer.go` com struct `Buffer` (campos: `lines []string`, `cursor Cursor`) e métodos: `InsertAt(row, col int, ch rune)`, `DeleteAt(row, col int)`, `SplitLine(row, col int)`, `JoinLines(row int)`, `MoveCursor(dr, dc int)` com boundary checks
- [x] T033 [P] [US3] Criar `components/editor/editor.go` com struct `Model` (campos: `buf buffer.Buffer`, `path string`, `dirty bool`, `viewport viewport.Model`, `focused bool`, `width int`, `height int`) e construtores `New(cfg config.Config) Model` (estado fechado) e `Open(path string) (Model, error)` (lê `os.ReadFile`, popula `buf.lines`)
- [x] T034 [US3] Implementar `Update(msg tea.Msg)` em `components/editor/editor.go`: `tea.KeyMsg` → inserir/deletar/mover cursor e marcar `dirty = true`; `Ctrl+S` → `os.WriteFile`, `dirty = false`, emitir `msgs.StatusBarNotifyMsg{Text:"Salvo"}`; `Ctrl+W` → se `dirty` emitir `msgs.ConfirmCloseMsg`, senão fechar; `OpenFileMsg` → chamar `Open(path)`; `EditorResizeMsg` → atualizar viewport
- [x] T035 [US3] Implementar `View()` em `components/editor/editor.go`: números de linha + conteúdo com cursor highlight usando Lip Gloss; scrolling via `viewport.Model`; borda de foco; exibir `[*]` no título se `dirty`
- [x] T036 [US3] Implementar fluxo `ConfirmCloseMsg` em `app/app.go`: ao receber, renderizar overlay de confirmação "Descartar alterações? (s/n)"; `s` → emitir `msgs.CloseConfirmedMsg`; `n` → emitir `msgs.CloseAbortedMsg`; editor consome esses msgs para fechar ou cancelar
- [x] T037 [US3] Integrar editor em `app/app.go`: adicionar ao `AppModel`, rotear `Ctrl+3` para `FocusChangeMsg{FocusEditor}`, rotear `OpenFileMsg` do sidebar para editor, atualizar `View()` para mostrar editor quando ativo

**Checkpoint**: Todas as 4 user stories funcionam independentemente

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Melhorias transversais e validação final

- [x] T038 [P] Escrever integration tests em `tests/integration/messages_test.go`: fluxo `OpenFileMsg` sidebar→editor, fluxo `ConfirmCloseMsg`→`CloseConfirmedMsg` app→editor, fluxo `MetricsTickMsg` statusbar→statusbar (re-enfileiramento), `tea.WindowSizeMsg` propagado como resize msgs para todos os componentes
- [x] T039 [P] Implementar layout adaptativo em `app/app.go` para terminais estreitos: quando `width < 80`, `sidebar.width = 0` (View retorna `""`) e terminal/editor usa toda a largura
- [x] T040 [P] Implementar help overlay em `app/app.go` usando `bubbles/help` com o `KeyMap` definido em `app/keymap.go` — exibido ao pressionar `?`, fechado ao pressionar qualquer tecla
- [x] T041 [P] Detectar modificação externa de arquivo em `components/editor/editor.go`: ao receber foco, comparar `os.Stat(path).ModTime()` com o valor no momento do Open; se alterado, emitir `msgs.StatusBarNotifyMsg{Text:"Arquivo alterado externamente — recarregar? (r/i)"}` e aguardar input
- [x] T042 Executar validação completa do `specs/001-lumina-core/quickstart.md` contra as 4 user stories no ambiente de desenvolvimento
- [x] T043 [P] Rodar `golangci-lint run ./...` e corrigir todos os warnings; garantir `go test ./...` com zero falhas
- [x] T044 [P] Documentar atalhos de teclado e modo de uso em `README.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Sem dependências — pode começar imediatamente
- **Foundational (Phase 2)**: Depende de Setup — **BLOQUEIA todas as user stories**
- **US1 Terminal (Phase 3)**: Depende de Foundational — **BLOQUEIA integração das demais**
- **US4 Status Bar (Phase 4)**: Depende de Foundational — pode ser em paralelo com US1 se houver 2+ devs
- **US2 Sidebar (Phase 5)**: Depende de Foundational — pode ser em paralelo com US1/US4
- **US3 Editor (Phase 6)**: Depende de Foundational + US2 (para `OpenFileMsg`) — pode ser em paralelo com US1/US4
- **Polish (Phase 7)**: Depende de todas as user stories completas

### User Story Dependencies

- **US1 (P1)**: Autônoma após Foundational ✅
- **US4 (P2)**: Autônoma após Foundational ✅
- **US2 (P2)**: Autônoma após Foundational ✅
- **US3 (P3)**: Autônoma após Foundational; integra com US2 via `OpenFileMsg` (mas testável sem ela)

### Within Each User Story

- Tests DEVEM ser escritos e FALHAR antes da implementação
- Buffer antes de Model (US3)
- Model antes da integração em app.go
- Integração em app.go completa a story

### Parallel Opportunities

- Setup: T001, T002, T003 podem rodar em paralelo após T001
- Foundational: T004, T005, T006 podem rodar em paralelo (após T001 para go.mod)
- US4 + US2 + US3 podem começar em paralelo após Foundational (com equipe de 3+)
- Dentro de cada story: tests e model creation marcados [P] podem rodar em paralelo

---

## Parallel Example: User Story 1

```bash
# Iniciar em paralelo após T007 (main.go):
Task: "T008 - Escrever unit tests em terminal_test.go"
Task: "T009 - Criar terminal.go com struct Model e New()"

# Após T009 completar, iniciar em sequência:
Task: "T010 - Implementar Init() em terminal.go"
Task: "T011 - Implementar Update() em terminal.go"
Task: "T012 - Implementar shell auto-restart"
Task: "T013 - Implementar View() em terminal.go"
Task: "T014 - Criar app.go com AppModel"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Completar Phase 1 (Setup) + Phase 2 (Foundational)
2. Completar Phase 3 (US1 — Terminal)
3. **PARAR e VALIDAR**: executar cenários US1 do quickstart.md
4. Deploy/demo: Lumina com terminal funcional já entrega valor real

### Incremental Delivery

1. Setup + Foundational → base pronta
2. US1 → terminal MVP → validar → demo (MVP!)
3. US4 → status bar → validar → demo
4. US2 → sidebar → validar → demo
5. US3 → editor → validar → release v1.0

### Parallel Team Strategy (3 devs)

1. Todos completam Setup + Foundational juntos
2. Após Foundational:
   - Dev A: US1 (Terminal — P1, crítico)
   - Dev B: US4 (Status Bar — P2, simples, independente)
   - Dev C: US2 (Sidebar — P2) → depois US3 (Editor — P3)

---

## Notes

- [P] = arquivos diferentes, sem dependências entre si
- [USn] mapeia task à user story para rastreabilidade
- Cada user story deve ser completável e testável de forma independente
- Verificar que tests FALHAM antes de implementar
- Fazer commit após cada task ou grupo lógico
- Parar em cada checkpoint para validar a story independentemente
- Evitar: tasks vagas, conflitos de arquivo, dependências cross-story que quebram independência
