---
description: "Task list for feature implementation: UX Polish Pack"
---

# Tasks: UX Polish Pack

**Input**: Design documents from `/home/menegas/fpm/lumina/specs/006-ux-polish-pack/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/msgs.md, contracts/config.md, quickstart.md

**Tests**: Tests are MANDATORY for this feature per the Lumina constitution (§II) — every new `tea.Msg` gets an integration test, every altered `tea.Model` gets unit tests.

**Organization**: Tasks are grouped by user story (US1–US4 from spec.md). Each story is an independently testable increment.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- File paths are absolute from repo root

## Path Conventions

- Single-project Go TUI at `/home/menegas/fpm/lumina/` — all source under top-level directories (`app/`, `components/`, `msgs/`, `config/`, `tests/`)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish a clean baseline before any story work begins.

- [X] T001 Verify baseline: checkout branch `006-ux-polish-pack` and confirm `go build ./...` and `go test ./...` pass at current HEAD; capture baseline in shell for later regression comparison.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Cross-cutting `msgs/`, `config/`, and package-removal work that every user story depends on. No US work may begin until this phase is complete.

**⚠️ CRITICAL**: All tasks here touch shared files (`msgs/msgs.go`, `app/app.go`, `config/`) and must land in a single logical unit to avoid broken intermediate states.

- [X] T002 Add new message types to `/home/menegas/fpm/lumina/msgs/msgs.go` per `contracts/msgs.md` §1: `ClickFocusMsg`, `SidebarCreateMsg`, `SidebarCreatedMsg`, `OpenInExternalEditorMsg`, `ClockTickMsg`, `PaneCWDChangeMsg`, `PaneGitStateMsg`, `FocusedPaneContextMsg`.
- [X] T003 Remove deprecated message types from `/home/menegas/fpm/lumina/msgs/msgs.go` per `contracts/msgs.md` §3: `EditorResizeMsg`, `ConfirmCloseMsg`, `CloseConfirmedMsg`, `CloseAbortedMsg`.
- [X] T004 Remove `FocusEditor` from the `FocusTarget` iota in `/home/menegas/fpm/lumina/msgs/msgs.go`; update any reference that survived compilation.
- [X] T005 Trim `CWD` and `GitBranch` fields from `MetricsTickMsg` in `/home/menegas/fpm/lumina/msgs/msgs.go` per `contracts/msgs.md` §2; update the metrics ticker producer to stop emitting those fields.
- [X] T006 Remove package `/home/menegas/fpm/lumina/components/editor/` entirely (`editor.go`, `buffer.go`, and tests). Build must fail until T007 lands.
- [X] T007 Update `/home/menegas/fpm/lumina/app/app.go` to drop editor routing: remove the `editor.Model` field, its init, `Update` branches, and `View()` integration; remove any `switch` arms consuming `FocusEditor` or the deleted close-confirmation msgs. Build must succeed after this task.
- [X] T008 [P] Add `Editor string` field to `Config` struct in `/home/menegas/fpm/lumina/config/config.go` per `contracts/config.md`; extend `defaults()` to set `Editor = "nano"`; extend the `LoadConfig` path that treats empty values as absent.
- [X] T009 [P] Register new keybinding defaults in `/home/menegas/fpm/lumina/config/keybindings.go`: `sidebar.new_dir = alt+d`, `sidebar.new_file = alt+f`, `sidebar.parent = backspace`; ensure `sidebar.enter` remains bound to Enter (kept as-is, behaviour shifts in US3).
- [X] T010 [P] Register the same three bindings in `/home/menegas/fpm/lumina/app/keymap.go` so `Update` consumers can reference them via `m.cfg.Keys.*`.
- [X] T011 Unit test in `/home/menegas/fpm/lumina/config/config_test.go` asserting: (a) default `Editor == "nano"` quando a chave está ausente; (b) valor vazio é tratado como ausente (→ `"nano"`); (c) valor explícito faz round-trip por TOML; (d) valor não resolvível no PATH NÃO é reescrito por `LoadConfig` — a resolução fica para o momento do spawn (coberto por T040).

**Checkpoint**: `go build ./...` green, `go test ./...` still passes (or only fails in tests that assert removed behaviour — those tests should be deleted as part of T006/T007). User story phases may begin.

---

## Phase 3: User Story 1 — Render fiel e estável em terminais (Priority: P1) 🎯 MVP

**Goal**: Garantir que CLIs TUI pintem corretamente no primeiro frame e permaneçam íntegros sob alta taxa de saída — sem necessidade de resize.

**Independent Test**: Rodar `claude` em um pane recém-aberto e confirmar alinhamento do cabeçalho no primeiro frame; em paralelo, rodar `while true; do echo $RANDOM; done` por 30s e confirmar ausência de artefatos.

### Tests for User Story 1

- [X] T012 [P] [US1] Unit test for cold-start repaint hook in `/home/menegas/fpm/lumina/components/terminal/firstrender_test.go` — synthetic `tea.WindowSizeMsg` + simulated PTY byte stream asserts that `pty.Setsize` is invoked before the first `PtyOutputMsg` is consumed into the VT buffer.
- [X] T013 [P] [US1] Unit test for output coalescing in `/home/menegas/fpm/lumina/components/terminal/terminal_test.go` — alimenta 10.000 `PtyOutputMsg` em sucessão rápida; assert: (a) o terminal emite no máximo um render por janela de frame (~16 ms); (b) todos os bytes resultantes estão presentes no buffer VT final em ordem (ausência de perda sob batch). Este teste valida o *mecanismo* de coalescência; a durabilidade de SC-002 é validada no roteiro manual (T054 + quickstart).
- [X] T014 [US1] Integration test in `/home/menegas/fpm/lumina/tests/integration/first_render_test.go` — spin up an app Model with a terminal pane, inject a fake shell that emits a multi-line header on start, assert the resulting `View()` has header lines in expected cells without relying on a second `WindowSizeMsg`.
- [ ] T014b [US1] Integration test in `/home/menegas/fpm/lumina/tests/integration/high_output_test.go` — spawn real PTY rodando `yes | head -n 100000` num pane; drenar o canal de output até EOF e assert: (a) a última linha do buffer VT é `y`; (b) nenhuma linha do buffer contém caractere fora de `y\n`; (c) o teste termina em menos de 5 segundos. Marcado com `testing.Short()` skip.

### Implementation for User Story 1

- [X] T015 [P] [US1] Create `/home/menegas/fpm/lumina/components/terminal/firstrender.go` implementing the cold-start repaint flow from `research.md` §R1: on `TerminalResizeMsg` for a pane with `firstRenderDone == false`, call `pty.Setsize`, drain any buffered bytes, then unblock PTY reads and set the flag.
- [X] T016 [US1] Modify `/home/menegas/fpm/lumina/components/terminal/terminal.go` to defer the first read until `firstRenderDone == true`; ensure `Init()` returns a `tea.Cmd` that queries initial window size and prompts the first `TerminalResizeMsg`.
- [X] T017 [P] [US1] Implement PTY output batching/coalescing in `/home/menegas/fpm/lumina/components/terminal/terminal.go` per `research.md` §R2: drain the PTY reader channel up to 64KB or empty before emitting the next batched `PtyOutputMsg` to the VT engine.
- [X] T018 [US1] Add the `firstRenderDone` field to the terminal model struct in `/home/menegas/fpm/lumina/components/terminal/terminal.go` (see `data-model.md` §3).

**Checkpoint**: US1 é demonstrável: abrir `claude` renderiza correto; `yes | head -n 100000` não produz artefatos.

---

## Phase 4: User Story 2 — Cursor, contexto e foco por clique (Priority: P1)

**Goal**: Cursor aparece apenas no pane focado; clicar com o mouse em qualquer pane transfere foco e entrega o evento ao componente; a status bar reflete CWD e git do pane focado, com glifo `●`/`✓`.

**Independent Test**: Abrir 2 panes em repos/branches diferentes; clicar entre eles e observar cursor + status bar atualizarem imediatamente; iniciar drag em pane não-focado deve selecionar texto no pane que acabou de ser focado.

### Tests for User Story 2

- [X] T019 [P] [US2] Unit test `HitTest(x,y)` in `/home/menegas/fpm/lumina/components/layout/layout_test.go` — assert correct `PaneID`, `Target`, and local coordinates for clicks across multiple layouts (horizontal split, nested vertical split, sidebar region).
- [X] T020 [P] [US2] Unit test cursor gating in `/home/menegas/fpm/lumina/components/terminal/terminal_test.go` — two terminal models, one focused, assert `View()` of focused includes cursor cell and View() of unfocused does not.
- [X] T021 [P] [US2] Unit test OSC 7 parser in `/home/menegas/fpm/lumina/components/terminal/terminal_test.go` — feed `"\x1b]7;file://host/home/user/x\x07"` and assert `PaneCWDChangeMsg{CWD: "/home/user/x"}` is produced.
- [X] T022 [P] [US2] Unit test statusbar context consumption in `/home/menegas/fpm/lumina/components/statusbar/statusbar_test.go` — feed `FocusedPaneContextMsg{Branch: "main", GitDirty: true, CWD: "/x"}`, assert rendered output includes `main ●` and `/x`; then feed `FocusedPaneContextMsg{Branch: "", ...}` and assert git field disappears.
- [X] T023 [US2] Integration test click-to-focus in `/home/menegas/fpm/lumina/tests/integration/click_focus_test.go` — app Model with 2 terminal panes, inject `tea.MouseMsg{Action: Press, X: pane2.Bounds.X+2, Y: pane2.Bounds.Y+2}`, assert `FocusChangeMsg` emitted and pane 2's `m.focused == true`.
- [ ] T024 [US2] Integration test drag-from-unfocused in `/home/menegas/fpm/lumina/tests/integration/click_focus_test.go` (separate test func) — assert that a `MouseActionPress` immediately followed by `MouseActionMotion` transfers focus on Press and the motion becomes a selection in the newly-focused pane.
- [ ] T025 [US2] Integration test statusbar focus context flow in `/home/menegas/fpm/lumina/tests/integration/statusbar_focus_test.go` — focus pane A (with CWD in git repo dirty), then focus pane B (no git), assert status bar shows correct glyph then disappears.

### Implementation for User Story 2

- [ ] T026 [P] [US2] Gate cursor render on `m.focused` in `View()` of `/home/menegas/fpm/lumina/components/terminal/terminal.go` per `research.md` §R3.
- [X] T027 [P] [US2] Implement `HitTest(x, y int) (paneID int, target msgs.FocusTarget, localX, localY int, ok bool)` in `/home/menegas/fpm/lumina/components/layout/layout.go` (use `bounds.go` helpers).
- [X] T028 [US2] Wire mouse Press handling in `/home/menegas/fpm/lumina/app/app.go`: on `tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}`, call `HitTest`, emit `tea.Batch(ClickFocusMsg{...}, FocusChangeMsg{Target}, <pass-through>)` where pass-through is `PtyMouseMsg`, `MouseSelectMsg`, or a sidebar-targeted variant depending on `Target`.
- [X] T029 [P] [US2] Implement OSC 7 parser in `/home/menegas/fpm/lumina/components/terminal/terminal.go` — scan PTY bytes for `\x1b]7;file://` sequences, percent-decode path, emit `PaneCWDChangeMsg{PaneID, CWD}`.
- [X] T030 [US2] On `PaneCWDChangeMsg`, spawn `tea.Cmd` in `/home/menegas/fpm/lumina/components/terminal/terminal.go` that execs `git -C <cwd> symbolic-ref --short HEAD` and `git -C <cwd> status --porcelain` with `context.WithTimeout(ctx, 200*time.Millisecond)`; emit `PaneGitStateMsg{PaneID, Branch, Dirty}` with empty branch on non-repo / timeout.
- [X] T031 [US2] Consolidate focus context in `/home/menegas/fpm/lumina/components/layout/layout.go` (or `focus.go`): on any of `FocusChangeMsg`, `PaneCWDChangeMsg` (for the focused pane), or `PaneGitStateMsg` (for the focused pane), emit `FocusedPaneContextMsg{PaneID, CWD, GitBranch, GitDirty}`.
- [X] T032 [US2] Update `/home/menegas/fpm/lumina/components/statusbar/statusbar.go` to consume `FocusedPaneContextMsg` and render: `<cwd>  <branch> <●|✓>` when branch present; hide the git segment when branch is empty. Use Lip Gloss for glyph color (dirty ≠ clean).
- [X] T033 [US2] Stop the statusbar from reading the (now trimmed) `MetricsTickMsg.CWD`/`GitBranch`; verify compilation in `/home/menegas/fpm/lumina/components/statusbar/statusbar.go`.
- [ ] T034 [US2] Formalize focused-panel border styling in `/home/menegas/fpm/lumina/components/layout/render.go` per FR-004 clarification: focused pane border uses the accent color (consistent with Q3 answer), unfocused use neutral gray; ensure a single Lip Gloss style variable is the source of truth.

**Checkpoint**: Mouse clicks transferem foco, cursor visível só no focado, status bar sensível ao pane focado com glifos git.

---

## Phase 5: User Story 3 — Sidebar file manager + editor externo (Priority: P2)

**Goal**: Sidebar funciona como gerenciador: Enter/Backspace navegam, Alt+D/Alt+F criam, Enter em arquivo abre no editor externo configurado.

**Independent Test**: Pela sidebar: Alt+D → "foo" → Enter (entra em foo); Alt+F → "bar.txt" → Enter (abre no editor). Backspace até a raiz e +1 exibe "Já na raiz" na status bar.

### Tests for User Story 3

- [X] T035 [P] [US3] Unit test sidebar Backspace navigation (non-root and at-root) in `/home/menegas/fpm/lumina/components/sidebar/sidebar_test.go`.
- [X] T036 [P] [US3] Unit test sidebar Enter routing (directory → navigate; file → `OpenInExternalEditorMsg`) in `/home/menegas/fpm/lumina/components/sidebar/sidebar_test.go`.
- [X] T037 [P] [US3] Unit test `createPrompt` validation + creation in new file `/home/menegas/fpm/lumina/components/sidebar/create_test.go` — cases: empty name rejected, slash rejected, existing name rejected, valid dir created with `os.Mkdir`, valid file created with `os.WriteFile`, ESC cancels, Enter on valid name emits `SidebarCreatedMsg`.
- [X] T038 [US3] Integration test sidebar create-dir flow in `/home/menegas/fpm/lumina/tests/integration/sidebar_create_test.go` — Alt+D → "newdir" → Enter; assert filesystem has new dir and sidebar CWD is now inside it.
- [X] T039 [US3] Integration test sidebar create-file flow in `/home/menegas/fpm/lumina/tests/integration/sidebar_create_test.go` (separate test func) — Alt+F → "x.txt" → Enter; assert `OpenInExternalEditorMsg{Path: "<dir>/x.txt"}` is routed and a new terminal pane is spawned running the editor.
- [X] T040 [US3] Integration test external editor spawn + error notify in `/home/menegas/fpm/lumina/tests/integration/external_editor_test.go` — two cases: valid editor spawns pane; invalid editor emits `StatusBarNotifyMsg{Level: NotifyError}` without spawning.

### Implementation for User Story 3

- [X] T041 [US3] Implement `openInExternalEditor(path string, cfg config.Config)` helper and its handler in `/home/menegas/fpm/lumina/app/app.go`: `exec.LookPath(cfg.Editor)`; on success, spawn a new terminal pane via `PaneSplitMsg` + PTY running `cmd := exec.Command(cfg.Editor, path)`; on failure, emit `StatusBarNotifyMsg{Level: msgs.NotifyError, Text: "editor '<editor>' não encontrado no PATH"}`.
- [X] T042 [P] [US3] Implement `sidebar.parent` (Backspace) in `/home/menegas/fpm/lumina/components/sidebar/sidebar.go`: `filepath.Dir` unless at the configured root; on root, emit `StatusBarNotifyMsg{Level: NotifyInfo, Text: "Já na raiz", Duration: 2*time.Second}`.
- [X] T043 [P] [US3] Extend `sidebar.enter` behaviour in `/home/menegas/fpm/lumina/components/sidebar/sidebar.go`: on directory → navigate; on file → emit `OpenInExternalEditorMsg{Path}`.
- [X] T044 [US3] Create `/home/menegas/fpm/lumina/components/sidebar/create.go` with the `createPrompt` submodel (see `data-model.md` §2): `textinput.Model`, `Kind`, `Err`, `ParentDir`; constructor takes `Kind` and current sidebar dir.
- [X] T045 [US3] Wire Alt+D / Alt+F in `/home/menegas/fpm/lumina/components/sidebar/sidebar.go`: when pressed and no prompt active, instantiate `createPrompt` with corresponding `Kind`; route all subsequent keystrokes to the prompt while active.
- [X] T046 [US3] Implement confirm/cancel in `/home/menegas/fpm/lumina/components/sidebar/create.go`: ESC clears the prompt; Enter validates (non-empty, no path separators, no existing entry), on valid calls `os.Mkdir` or `os.WriteFile` and emits `SidebarCreatedMsg`; on invalid sets `m.Err` and keeps the prompt open.
- [X] T047 [US3] Handle `SidebarCreatedMsg` in `/home/menegas/fpm/lumina/app/app.go`: if `Kind == "dir"`, forward a navigate-into msg to sidebar; if `Kind == "file"`, emit `OpenInExternalEditorMsg{Path}`.
- [X] T048 [US3] Render the inline prompt in sidebar `View()` at `/home/menegas/fpm/lumina/components/sidebar/sidebar.go` — show the `textinput` plus `m.Err` below when non-empty; Lip Gloss styles only.

**Checkpoint**: US3 é demonstrável: usuário cria pasta/arquivo pela sidebar e edita com o editor externo; editor inexistente é reportado na status bar.

---

## Phase 6: User Story 4 — Relógio na status bar (Priority: P3)

**Goal**: Exibir o relógio atualizado na status bar.

**Independent Test**: Abrir o Lumina, observar status bar: formato HH:MM presente e avançando.

### Tests for User Story 4

- [X] T049 [P] [US4] Unit test clock tick rendering in `/home/menegas/fpm/lumina/components/statusbar/statusbar_test.go` — fire a `ClockTickMsg{Now: fixed}` and assert `View()` contains formatted HH:MM; fire a second with advanced time and assert the displayed time updates.

### Implementation for User Story 4

- [X] T050 [P] [US4] Add `now time.Time` field to the model in `/home/menegas/fpm/lumina/components/statusbar/statusbar.go`; on `Init` return a `tea.Cmd` emitting `ClockTickMsg{Now: time.Now()}` immediately and re-arming via `tea.Tick(30*time.Second, ...)`.
- [X] T051 [P] [US4] Update `View()` in `/home/menegas/fpm/lumina/components/statusbar/statusbar.go` to render `m.now.Format("15:04")` in a designated segment (left side, before CWD).

**Checkpoint**: US4 shippable.

---

## Phase 7: Polish & Cross-Cutting Concerns

- [X] T052 Run `go test ./...` from `/home/menegas/fpm/lumina/` — assert zero failures.
- [ ] T053 Run `golangci-lint run` from `/home/menegas/fpm/lumina/` — assert zero warnings.
- [ ] T054 Execute the manual roteiro in `/home/menegas/fpm/lumina/specs/006-ux-polish-pack/quickstart.md` end-to-end; tick each step in a working copy of the file or a scratch note.
- [X] T055 Append a DECISIONS.md entry in `/home/menegas/fpm/lumina/.specify/DECISIONS.md` (or project root `DECISIONS.md`, whichever is the canonical location in this repo) documenting: removal of `components/editor/`, trimming of `MetricsTickMsg`, introduction of `FocusedPaneContextMsg` pipeline, and the `Editor` config field.
- [X] T056 Verify `/home/menegas/fpm/lumina/CLAUDE.md` reflects the active feature and any new architectural notes (auto-updated by `update-agent-context.sh`; spot-check for consistency).

---

## Dependencies

**Phase order** (must complete before next):
1. Phase 1 (Setup) → Phase 2 (Foundational) → Phases 3–6 (user stories, parallelizable among themselves) → Phase 7 (Polish).

**Within Phase 2**: T002 and T003 must happen before T004 (enum change); T006 must precede T007 (removing editor package before updating app.go routing). T008–T011 are [P] among themselves and can run after T002.

**Cross-story dependencies** (after Phase 2 checkpoint):
- **US1 (Phase 3)**: fully independent once Phase 2 is done.
- **US2 (Phase 4)**: fully independent of US1; **however** T032 depends on `FocusedPaneContextMsg` being defined (Phase 2 T002).
- **US3 (Phase 5)**: depends on Phase 2 T008 (`Editor` config field) and Phase 2 T002 (`OpenInExternalEditorMsg`). Independent of US1/US2 otherwise.
- **US4 (Phase 6)**: depends on Phase 2 T002 (`ClockTickMsg`). Fully independent of US1/US2/US3 otherwise.

**Within a story**: Tests [P] may be written in parallel with implementation [P] if following TDD; otherwise implement first, then test. Respect the constitution rule (§II) that bug-fix flows require a failing test first (applies to US1's render fixes).

---

## Parallel Execution Examples

**Phase 2 parallel batch (after T002–T007 finish serially)**:

```text
T008 (config Editor field)
T009 (keybindings defaults)
T010 (app keymap)
```

**US1 test+impl parallel batch**:

```text
T012 (firstrender_test.go)
T013 (terminal_test.go batching)
T015 (firstrender.go)
T017 (terminal.go batching)
```

T014 (integration) depends on T015+T017 landing.

**US2 test+impl parallel batch**:

```text
T019 (layout_test HitTest)
T020 (terminal_test cursor gating)
T021 (terminal_test OSC 7)
T022 (statusbar_test context)
T026 (terminal cursor gate)
T027 (layout HitTest)
T029 (terminal OSC 7)
```

T028 (mouse routing in app.go) sequences after T027; T030–T034 sequence after their upstream implementers.

**US3 parallel batch**:

```text
T035, T036 (sidebar_test)
T042, T043 (sidebar navigation implementation)
```

T037 (create_test) + T044 (create.go) sequenced together; T045–T048 depend on T044.

**US4 entirely parallel** (3 tasks, 3 files): T049, T050, T051.

---

## Implementation Strategy

**MVP**: Phases 1 + 2 + 3 (US1 — Render fiel e estável). This alone restores confidence in running CLIs like Claude Code inside Lumina and eliminates the most visible regression in the screenshot provided by the user.

**Incremental delivery order** (one PR per phase is reasonable, but Phase 2 must land atomically):

1. Phase 2 + Phase 3 (US1) — ship the render fixes as MVP.
2. Phase 4 (US2) — cursor, click-to-focus, status bar context; this is the second-highest user pain.
3. Phase 5 (US3) — sidebar + external editor overhaul; the deepest refactor, contains the editor package removal (already in Phase 2) but its consumer flow lands here.
4. Phase 6 (US4) — relógio; quick polish shipped last.
5. Phase 7 — gate before merge to `main`.

**Risk areas**:

- T030 (git query `tea.Cmd`): fork/exec inside an event loop is the riskiest piece; the 200ms timeout is the safety rail. Keep the command pipeline small and avoid shelling out to anything other than `git`.
- T017 (PTY output coalescing): a bad implementation can introduce latency on keystrokes; the batch should flush immediately (within a frame) if no further bytes arrive within a small window (~4ms).
- T028 (mouse routing): care needed to not break the existing drag-selection feature (`mouseselect.go`); T024 integration test exists specifically to catch that.

---

## Task Count Summary

| Phase | Tasks | Notes |
|---|---:|---|
| 1. Setup | 1 | Baseline check |
| 2. Foundational | 10 | Shared msgs/config/removal |
| 3. US1 — Render | 8 | 4 tests + 4 impl |
| 4. US2 — Cursor/focus/click | 16 | 7 tests + 9 impl |
| 5. US3 — Sidebar + editor | 14 | 6 tests + 8 impl |
| 6. US4 — Clock | 3 | 1 test + 2 impl |
| 7. Polish | 5 | Lint/test/docs |
| **Total** | **57** | |

## Independent Test Criteria per Story

- **US1**: `claude` starts correctly on first frame; `yes | head -n 200000` produces no artifacts. Covered by T014 integration test.
- **US2**: Clicking between 2 panes transfers focus and the status bar/git glyph update accordingly; drag from unfocused selects in the just-focused pane. Covered by T023–T025.
- **US3**: Create dir → navigate into it; create file → editor opens; editor-not-found → error notification. Covered by T038–T040.
- **US4**: HH:MM visible and advancing over ≥1 minute. Covered by T049.

## Suggested MVP Scope

**MVP = Phases 1 + 2 + 3** (12 tasks, ~US1 ready). Delivers the single most impactful fix (first-frame render correctness) — rest is polish that can ship over subsequent PRs.
