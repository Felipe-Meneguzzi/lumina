# Tasks: UX Fixes — Multi-Window Layout

**Input**: Design documents from `specs/003-ux-fixes-multiwindow/`  
**Prerequisites**: plan.md ✓ spec.md ✓ research.md ✓ data-model.md ✓ contracts/ ✓

**Tests**: Incluídos para os bug fixes (TDD obrigatório pela Constituição, Princípio II).

**Organização**: Tarefas agrupadas por user story. US2 (shell) é executada antes de US5 (close bug) porque é a sua root cause.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Pode rodar em paralelo (arquivos diferentes, sem dependências incompletas)
- **[Story]**: User story correspondente

---

## Phase 1: Setup

**Purpose**: Verificar ambiente antes de alterar código

- [x] T001 Rodar `go build -o lumina .` e confirmar build passa sem erros
- [x] T002 Rodar `go test ./...` e confirmar zero falhas (baseline)

---

## Phase 2: Foundational — `validateShell()` (Bloqueia US2 e US5)

**Purpose**: Infraestrutura de validação de shell que resolve a root cause de US2 e US5 simultaneamente. Deve ser concluída antes de qualquer outra story.

**⚠️ CRÍTICO**: US2 (shell correto) e US5 (close bug causado por split falhando) dependem deste phase.

**TDD — escrever testes ANTES da implementação:**

- [x] T003 [P] Escrever `TestValidateShell` em `config/config_test.go`: testar shell válido retorna o mesmo, shell inválido (`"invalid-shell-xyz"`) retorna fallback, shell vazio retorna fallback, `$SHELL` env var é usado quando configured é inválido

**Implementação:**

- [x] T004 Adicionar import `"os/exec"` e função `validateShell(configured string) string` em `config/config.go`: iterar `[]string{configured, os.Getenv("SHELL"), "/bin/bash", "/bin/zsh", "/bin/sh"}`, usar `exec.LookPath()` para validar cada um, retornar o primeiro válido
- [x] T005 Em `config/config.go` `defaults()`: substituir atribuição direta por `cfg.Shell = validateShell(shell)`
- [x] T006 Em `config/config.go` `LoadConfig()`: após `toml.DecodeFile`, adicionar `cfg.Shell = validateShell(cfg.Shell)`
- [x] T007 Rodar `go test ./config/...` — todos os testes devem passar

**Checkpoint**: `validateShell` implementada e testada. Shell nunca será inválido.

---

## Phase 3: User Story 2 — Terminal com Shell Correto (P1)

**Goal**: Novos terminais sempre abrem com o shell padrão do sistema (não PowerShell ou shell inválido).

**Independent Test**: Iniciar Lumina; abrir novo painel (`Alt+|`); verificar que o processo iniciado no PTY corresponde ao shell do sistema (`$SHELL` ou `/bin/bash`).

- [x] T008 [US2] Em `app/app.go` `New()`: após criar o model, emitir `msgs.StatusBarNotifyMsg` informando qual shell está sendo usado (`cfg.Shell`) — notificação informativa de 3s ao startup
- [x] T009 [US2] Rodar `go build -o lumina .` e testar manualmente: iniciar Lumina, verificar shell correto no terminal

**Checkpoint**: US2 completa. Terminais abrem com shell correto. US5 (close bug) provavelmente já funciona — validar na Phase 6.

---

## Phase 4: User Story 4 — Foco Move para Novo Pane após Split (P1)

**Goal**: Após um split, o foco vai automaticamente para o novo pane criado (não permanece no original).

**Independent Test**: Abrir Lumina → `Alt+|` para split → imediatamente pressionar `Alt+Q` → o novo pane deve fechar (não o original).

**TDD — escrever testes ANTES da implementação:**

- [x] T010 Adicionar `TestSplitFocusMovesToNewPane` em `components/layout/layout_test.go`: criar layout com 1 pane, enviar `PaneSplitMsg{Direction: SplitHorizontal}`, verificar que `FocusedID()` retorna o novo pane ID (não o original)

**Implementação:**

- [x] T011 [US4] Em `components/layout/layout.go` `handleSplit()`: substituir `m.applyFocus(m.focused, true)` por `m.focused = m.nextID` + `m.applyFocus(m.nextID, true)`
- [x] T012 [US4] Em `components/layout/layout.go`: adicionar método `FocusedID() PaneID { return m.focused }` após `FocusedKind()`
- [x] T013 [US4] Rodar `go test ./components/layout/...` — todos os testes devem passar
- [x] T014 [US4] Rodar `go build -o lumina .` e testar manualmente: split → verificar foco no novo pane via borda highlighted

**Checkpoint**: US4 completa. Foco correto após split.

---

## Phase 5: User Story 5 — Fechar Janela Inicial Funciona (P1)

**Goal**: Fechar qualquer pane (incluindo o inicial) funciona quando há mais de um pane aberto. Nenhuma mensagem incorreta de "única janela".

**Independent Test**: Abrir Lumina → split (`Alt+|`) → `Alt+H` para voltar ao pane original → `Alt+Q` → pane original fecha, pane novo permanece.

**Validação** (não requer código novo — US2 + US4 devem ter resolvido o bug):

- [x] T015 [US5] Rodar `go build -o lumina .` e testar o cenário: split → foco no original → fechar original. Se funcionar: US5 está resolvida. Se ainda falhar: investigar `handleClose()` em `components/layout/layout.go` e `PaneCount()`.
- [x] T016 [US5] Se houver regressão: adicionar `TestCloseInitialPane` em `components/layout/layout_test.go` reproduzindo o bug, e corrigir a causa raiz no `handleClose()`.

**Checkpoint**: US5 completa. Fechar qualquer pane funciona corretamente.

---

## Phase 6: User Story 1 — Sidebar Toggle via Keybind (P1)

**Goal**: Usuário pode ocultar/exibir a sidebar do pane em foco com `Alt+B`. Estado de visibilidade é memorizado por pane.

**Independent Test**: Abrir dois panes; `Alt+B` no pane 1 → sidebar oculta apenas para pane 1; navegar para pane 2 → sidebar visível; voltar ao pane 1 → sidebar permanece oculta.

### Keybinding (sem dependências entre si — [P])

- [x] T017 [P] [US1] Em `config/keybindings.go`: adicionar `ToggleSidebar []string` e `ToggleStatusBar []string` à struct `Keybindings`; adicionar defaults `["alt+b"]` e `["alt+m"]` em `defaultKeybindings()`; adicionar em `Action()`, `GlobalKeys()` e `LoadKeybindings()`
- [x] T018 [P] [US1] Em `app/keymap.go`: adicionar `ToggleSidebar key.Binding` e `ToggleStatusBar key.Binding`; inicializar em `NewKeyMap()`; adicionar em `FullHelp()` e `ShortHelp()`

### State no Model

- [x] T019 [US1] Em `app/app.go` struct `Model`: adicionar campos `sidebarVisible bool`, `sidebarPrevWidth int`, `paneShowSidebar map[layout.PaneID]bool` (depende de T017, T018)
- [x] T020 [US1] Em `app/app.go` `New()`: inicializar `sidebarVisible = true`, `paneShowSidebar = make(map[layout.PaneID]bool)`

### Lógica de Toggle

- [x] T021 [US1] Em `app/app.go`: adicionar método `toggleSidebar() Model` que inverte estado para o pane focado em `paneShowSidebar[m.layout.FocusedID()]` e chama `applySidebarState(visible bool)`
- [x] T022 [US1] Em `app/app.go`: adicionar método `applySidebarState(visible bool) Model`: quando `visible=false` preservar `sidebarPrevWidth = sidebarWidth` e setar `sidebarWidth = 0`; quando `visible=true` restaurar `sidebarWidth = max(sidebarPrevWidth, 30)`
- [x] T023 [US1] Em `app/app.go` `handleKey()` switch: adicionar case `"toggle_sidebar"` → `return m.toggleSidebar(), nil`

### Propagação ao Mudar Foco entre Panes

- [x] T024 [US1] Em `app/app.go` `handleKey()`: nos cases `focus_pane_left/right/up/down`, após `updateLayout(PaneFocusMoveMsg{})`, chamar `applySidebarForFocusedPane()` no model retornado
- [x] T025 [US1] Em `app/app.go`: adicionar método `applySidebarForFocusedPane() Model` que lê `paneShowSidebar[m.layout.FocusedID()]` (default `true`) e chama `applySidebarState`

### Limpeza ao Fechar Pane

- [x] T026 [US1] Em `app/app.go` `handleKey()` case `close_pane`: capturar `closedID := m.layout.FocusedID()` antes do `updateLayout(PaneCloseMsg{})`; após retorno, `delete(m.paneShowSidebar, closedID)`

### Testes

- [x] T027 [US1] Adicionar `TestToggleSidebarPerPane` em `app/app_test.go`: model com 2 panes, toggle sidebar no pane 1, verificar `sidebarWidth == 0`; mudar foco para pane 2, verificar `sidebarWidth > 0`; voltar pane 1, verificar `sidebarWidth == 0`
- [x] T028 [US1] Rodar `go test ./...` — todos os testes devem passar
- [x] T029 [US1] Rodar `go build -o lumina .` e testar manualmente: `Alt+B` oculta/exibe sidebar; `Alt+Shift+]`/`[` ainda funciona para resize

**Checkpoint**: US1 (keybind) completa. Sidebar toggling funciona por pane.

---

## Phase 7: User Story 3 — Resource Monitor Toggle (P2)

**Goal**: `Alt+M` oculta/exibe o resource monitor globalmente. Estado é único para toda a aplicação.

**Independent Test**: Pressionar `Alt+M` → resource monitor desaparece e área de conteúdo expande; pressionar novamente → monitor reaparece.

### State e Keybinding

- [x] T030 [US3] Em `app/app.go` struct `Model`: adicionar campo `sbarVisible bool` (depende de T017-T018 já concluídos na Phase 6)
- [x] T031 [US3] Em `app/app.go` `New()`: inicializar `sbarVisible = true`

### Lógica de Toggle

- [x] T032 [US3] Em `app/app.go` `handleKey()`: adicionar case `"toggle_statusbar"` → `m.sbarVisible = !m.sbarVisible`; chamar `reapplyResize()` para recomputar altura do conteúdo
- [x] T033 [US3] Em `app/app.go`: adicionar método `reapplyResize() (tea.Model, tea.Cmd)` que chama `m.handleResize(tea.WindowSizeMsg{Width: m.width, Height: m.height})`
- [x] T034 [US3] Em `app/app.go` `handleResize()`: substituir `contentHeight := msg.Height - statusBarHeight` por lógica condicional: `effectiveStatusH := 0; if m.sbarVisible { effectiveStatusH = statusBarHeight }; contentHeight := msg.Height - effectiveStatusH`
- [x] T035 [US3] Em `app/app.go` `View()`: envolver `sbarView` em condicional: se `!m.sbarVisible`, não incluir no `JoinVertical`

### Testes

- [x] T036 [US3] Adicionar `TestToggleStatusBar` em `app/app_test.go`: verificar que após toggle `sbarVisible = false`, o `contentHeight` propagado ao layout é `m.height` (sem subtrair `statusBarHeight`); verificar que `View()` não inclui statusbar
- [x] T037 [US3] Rodar `go test ./...` — todos os testes devem passar
- [x] T038 [US3] Rodar `go build -o lumina .` e testar manualmente: `Alt+M` oculta/exibe monitor; área de conteúdo expande corretamente

**Checkpoint**: US3 completa. Resource monitor toggle funciona.

---

## Phase 8: User Story 1 — Mouse Drag Sidebar (P2 — complementa US1)

**Goal**: Usuário pode arrastar a borda entre sidebar e conteúdo para redimensionar.

**Independent Test**: Clicar e arrastar a borda da sidebar horizontalmente → largura da sidebar muda em tempo real.

- [x] T039 [US1] Em `main.go`: adicionar `tea.WithMouseAllMotion()` às opções do `tea.NewProgram(...)` (antes: `tea.WithAltScreen()`, depois: `tea.WithAltScreen(), tea.WithMouseAllMotion()`)
- [x] T040 [US1] Em `app/app.go` struct `Model`: adicionar `sidebarDragging bool` e `sidebarDragStartX int`
- [x] T041 [US1] Em `app/app.go`: extrair método `resizeSidebarTo(newW int) (tea.Model, tea.Cmd)` a partir da lógica existente em `resizeSidebar(delta int)` para aceitar largura absoluta (elimina duplicação)
- [x] T042 [US1] Em `app/app.go` `handleMouse()`: alterar retorno para `(Model, tea.Cmd)`; adicionar detecção de clique na borda `abs(msg.X - m.sidebarWidth) <= 1` para iniciar drag (`sidebarDragging = true`); adicionar case `MouseActionMotion` para atualizar `sidebarWidth` via `resizeSidebarTo(msg.X)`; adicionar case `MouseActionRelease` para encerrar drag
- [x] T043 [US1] Em `app/app.go` `Update()`: ajustar tipo de retorno do `handleMouse` de `Model` para `(tea.Model, tea.Cmd)` e propagar o cmd
- [x] T044 [US1] Rodar `go build -o lumina .` e testar manualmente: drag na borda da sidebar redimensiona em tempo real

**Checkpoint**: US1 completa (keybind + mouse drag). Sidebar totalmente funcional.

---

## Phase Final: Polish & Cross-Cutting

- [x] T045 [P] Em `app/app.go` `New()`: adicionar `abs()` helper local se não existir (para detecção de borda no drag)
- [x] T046 [P] Em `app/keymap.go` `ShortHelp()`: adicionar `k.ToggleSidebar` e `k.ToggleStatusBar` na lista compacta
- [x] T047 Rodar `go test ./...` completo — zero falhas
- [x] T048 Rodar `go build -o lumina .` — zero warnings
- [x] T049 Verificar cenários do `quickstart.md` manualmente: toggle sidebar, toggle monitor, split→close, shell correto

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundational: validateShell) ← BLOQUEIA tudo
    ↓
Phase 3 (US2: Shell info) ──────────────────────┐
Phase 4 (US4: Focus após split) ─────────────────┤
    ↓                                            ↓
Phase 5 (US5: Validação close bug) ←────── depende de US2+US4
    ↓
Phase 6 (US1: Sidebar toggle keybind) ─────────┐
Phase 7 (US3: Monitor toggle) ─────────────────┤  (paralelas)
    ↓                                           ↓
Phase 8 (US1 mouse drag: depende de Phase 6)
    ↓
Phase Final (Polish)
```

### User Story Dependencies

- **US2 (P1)**: Depende do Foundational (Phase 2) — sem dependências entre stories
- **US4 (P1)**: Independente — sem dependências entre stories
- **US5 (P1)**: Validação — depende de US2 e US4 estarem concluídas
- **US1 (P1)**: Independente — sem dependências entre stories
- **US3 (P2)**: Compartilha keybindings com US1 (T017/T018) — iniciar após Phase 6

### Paralelas Dentro do Mesmo Phase

- **Phase 6**: T017 (keybindings.go) e T018 (keymap.go) são [P] — arquivos diferentes
- **Phase Final**: T045 e T046 são [P] — arquivos diferentes

---

## Parallel Example: User Story 1 (Phase 6)

```
# Podem rodar em paralelo (Phase 6):
Task T017: config/keybindings.go — adicionar ToggleSidebar/ToggleStatusBar
Task T018: app/keymap.go — adicionar bindings

# Sequencial após T017+T018:
Task T019: app/app.go — adicionar campos ao Model
Task T020: app/app.go — inicializar em New()
...
```

---

## Implementation Strategy

### MVP First (US2 + US4 + US5 — bugs críticos)

1. Completar **Phase 1** (Setup — 2 tasks)
2. Completar **Phase 2** (Foundational — 5 tasks, TDD)
3. Completar **Phase 3** (US2 — 2 tasks)
4. Completar **Phase 4** (US4 — 5 tasks, TDD)
5. **PARAR e VALIDAR Phase 5** (US5 — testar close bug resolvido)
6. Deploy/demo do fix crítico

### Incremental Delivery

1. Setup + Foundational → shell correto em todos os terminais
2. US2 + US4 + US5 → multi-window funcional sem bugs críticos
3. US1 (sidebar toggle keybind) → UX melhorada
4. US3 (monitor toggle) → UX melhorada
5. US1 (mouse drag) → UX polida

---

## Notes

- [P] = diferentes arquivos, sem dependências incompletas — podem ser paralelizados
- TDD obrigatório para bug fixes (Constituição, Princípio II): escrever teste → falhar → implementar → passar
- `go test ./...` deve passar com zero falhas antes de cada checkpoint
- Keybindings definidos SOMENTE em `app/keymap.go` + `config/keybindings.go` — não hardcodar em componentes
- Mouse drag em Phase 8 é independente das fases anteriores de US1 — pode ser adiado sem impactar o resto
