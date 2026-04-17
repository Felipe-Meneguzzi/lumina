# Phase 0 Research — UX Polish Pack

**Feature**: 006-ux-polish-pack
**Date**: 2026-04-17

Esta fase resolve todas as áreas técnicas não-óbvias que a spec deixou em aberto, para que Phase 1 (design) possa ser executada sem `NEEDS CLARIFICATION`.

---

## R1 — Correção de render inicial quebrado (ex.: Claude Code no primeiro boot)

**Decision**: Introduzir um hook de "cold-start repaint" em `components/terminal/` que, ao receber o primeiro `TerminalResizeMsg` para um pane recém-criado, envia `SIGWINCH` ao processo filho via `pty.Setsize` **e** descarta qualquer conteúdo de framebuffer acumulado antes desse ponto, reemitindo a view após o primeiro read estável.

**Rationale**: A análise do fluxo atual revela a race condition: `exec.Start` dispara o shell com tamanho default do PTY (80×24); o shell imprime PS1 e qualquer CLI (ex.: Claude Code) pinta um cabeçalho TUI baseado nesse tamanho. Só depois chega `tea.WindowSizeMsg` e fazemos `pty.Setsize`. O conteúdo já pintado fica com coordenadas erradas — o VT engine (ultraviolet) tem o snapshot com as linhas no lugar errado até que o usuário force um repaint (resize da janela). A correção é inverter a ordem: calcular as dimensões-alvo no boot do programa (usando `tea.WindowSize()` ou lendo o TTY do host), aplicar `Setsize` antes do primeiro `io.Copy` e só então fazer read. Como Bubble Tea pode demorar para disparar o primeiro `WindowSizeMsg`, adicionamos um `tea.Cmd` que chama `tea.WindowSize()` dentro de `Init()` e usa o resultado para pré-configurar o PTY.

**Alternatives considered**:
- **Force-repaint genérico (enviar Ctrl+L a todo shell)**: rejeitado — polui o scrollback e não funciona em TUIs que não são shell (htop, vim).
- **Retardar o spawn do shell até o primeiro WindowSizeMsg**: rejeitado — introduz flash de pane vazio visível; e o usuário pode começar a digitar antes do shell estar pronto.
- **Ignorar bytes do PTY até `Setsize` ser aplicado**: rejeitado — se o shell emite o PS1 antes, ele fica invisível.

---

## R2 — Estabilidade de render sob alta taxa de saída

**Decision**: No `components/terminal/terminal.go`, coalescer múltiplas `PtyOutputMsg` consecutivas em um único `View()` por frame, usando o padrão drenagem + batch já idiomático em Bubble Tea: ao receber uma `PtyOutputMsg`, agendar um `tea.Cmd` que drene o canal do reader até vazio ou até um teto de tamanho (ex.: 64KB) antes de emitir a próxima mensagem de refresh.

**Rationale**: Bubble Tea processa `Update` serialmente; se cada `PtyOutputMsg` dispara uma renderização, sob 10.000 linhas/segundo o framework encara 10.000 ciclos `Update→View→render`, muito acima dos 16ms/frame. O resultado visual é o framebuffer sendo parcialmente reescrito enquanto o terminal host ainda está enviando o frame anterior — daí os artefatos que só somem com um resize (que força um repaint completo do cellbuf). Ao coalescer, entregamos ao VT engine um bloco consistente por frame.

**Alternatives considered**:
- **Throttle por timer fixo (ex.: emitir uma `PtyOutputMsg` a cada 16ms)**: rejeitado — introduz latência artificial em saídas esparsas (usuário digitando em shell) e degrada UX.
- **Alocar um framebuffer dedicado por pane e sincronizar via mutex**: rejeitado — viola o princípio de "no global mutable state" e confunde o modelo Elm.
- **Usar `charmbracelet/x/vt.TerminalBuffer.WriteChunked`** se existir: investigado — a API atual expõe `Write([]byte)` sem batching nativo; a drenagem fica na camada de leitura mesmo.

---

## R3 — Cursor por terminal visível apenas no painel focado

**Decision**: No `View()` do `components/terminal/`, renderizar o cursor (bloco/linha, conforme estilo VT) somente quando `m.focused == true`. Em painéis não-focados, substituir a célula do cursor pela mesma célula com atributo normal (sem bloco) — o VT engine da ultraviolet já mantém a posição do cursor internamente, o que muda é só a decoração do render.

**Rationale**: A ultraviolet (`charmbracelet/x/vt`) mantém o cursor como propriedade interna da `Terminal` struct; cada instância do terminal/pane já tem sua posição preservada naturalmente. O que estava faltando é o aspecto visual: hoje o render desenha o cursor em todos os panes, criando confusão. Ao gatear o desenho do cursor em `focused`, preservamos a posição (nada é perdido no blur) e mostramos apenas um de cada vez. Como bônus, isso também evita piscadas competindo entre terminais não-focados.

**Alternatives considered**:
- **Ocultar o cursor via ESC `\x1b[?25l` no PTY quando perde foco**: rejeitado — modifica o estado do PTY do processo filho (alguns apps, como vim, contam com `?25h`/`?25l` sendo deles); também reinicializa ao ganhar foco causando flicker.
- **Atenuar (dim) o cursor em vez de esconder**: considerado — rejeitado por adicionar complexidade sem benefício claro; com bordas coloridas já indicando foco (FR-004), o cursor apenas-no-focado é suficiente.

---

## R4 — Click-to-focus com pass-through

**Decision**: No `app/app.go`, interceptar `tea.MouseMsg` com `Action == tea.MouseActionPress` em `Update`. Resolver o `(X, Y)` do clique contra o `layout.Tree` atual usando uma função `HitTest(x, y int) (PaneID, target FocusTarget, rectLocalX, rectLocalY)`. Emitir em sequência: (a) `FocusChangeMsg{Target: target}` para mudar foco; (b) a mensagem de pass-through apropriada — `PtyMouseMsg` ou `MouseSelectMsg` para terminal, ou uma `SidebarMouseSelectMsg` (novo msg) para sidebar — com coordenadas locais ao pane. Ambos eventos são emitidos no mesmo `tea.Batch`, garantindo uma única re-render por clique.

**Rationale**: Bubble Tea entrega `MouseMsg` globalmente; sem intercepção, nenhum componente sabe se o clique pertence a ele. O `layout.Tree` já conhece os bounds de cada pane (`layout/bounds.go`). Um hit-test em O(log n) sobre a árvore resolve o destino; emitir foco primeiro e o evento depois garante que o pane já está `focused=true` ao processar o clique, o que casa com Q2 da spec (drag inicia no pane recém-focado). Usar `tea.Batch` em vez de enviar dois msgs separados evita dois ciclos `Update→View`, respeitando o budget de 16ms.

**Alternatives considered**:
- **Cada componente faz hit-test interno e decide se o clique lhe pertence**: rejeitado — viola o fluxo do layout; componentes não conhecem os bounds dos vizinhos.
- **Usar `MouseActionRelease` em vez de `Press`**: rejeitado — Q2 da spec define mousedown como o momento da transferência (para drag coerente).
- **Bubble Tea bounce (`tea.WithMouseCellMotion`) já tem wrapper "mouse.Bubble"**: investigado — não cobre a necessidade de hit-test contra layout tree custom.

---

## R5 — Editor externo em painel de terminal

**Decision**: Adicionar ao `config.Config` o campo `Editor string` (default `"nano"`, aceitos: `"nano"`, `"vim"`, `"nvim"` ou path absoluto). Quando o app recebe `OpenFileMsg{Path}`, criar um novo pane de terminal (via `PaneSplitMsg` se não há um painel livre, ou reaproveitar um painel temporário dedicado) cujo processo é `exec.Command(cfg.Editor, path)` em vez de `cfg.Shell`. Remover `components/editor/` e todos os msgs relacionados (`EditorResizeMsg`, `ConfirmCloseMsg`, `CloseConfirmedMsg`, `CloseAbortedMsg`) — estes são substituídos pelo ciclo de vida natural do PTY: quando o editor fecha, o pane morre como qualquer outro shell que fez `exit`.

**Rationale**: Delegar a um editor externo cumpre FR-017 (sem editor embutido) e aproveita a infraestrutura de PTY já madura. O usuário opera com sua ferramenta habitual, com seus próprios keymaps. Do lado do Lumina, o código removido reduz a superfície de bugs e testes. A pergunta "o que acontece quando o editor fecha" é resolvida pelo PTY EOF natural (pane close + foco passa para o próximo, conforme comportamento atual). Se o binário configurado não existe no PATH, `exec.LookPath` falha e emitimos `StatusBarNotifyMsg{Level: NotifyError, Text: "editor 'vim' não encontrado no PATH"}` sem criar o pane.

**Alternatives considered**:
- **Spawn o editor em processo separado inline (sem pane)**: rejeitado — requer captura de TTY do host, pausando todo o Lumina.
- **Manter editor embutido como fallback quando externo falha**: rejeitado — contradiz FR-017 e adiciona ao debt que queremos eliminar.
- **Tornar `Editor` uma lista de preferência com fallback automático**: rejeitado — espec diz explicitamente "editor selecionado pelo usuário"; fallback automático mascara configuração errada.

---

## R6 — Navegação da sidebar + criação de arquivos/pastas

**Decision**: No `components/sidebar/sidebar.go`, binding `key.Enter` em diretório entra nele (já existente — verificar); `key.Enter` em arquivo emite `OpenFileMsg{Path}` ao invés do comportamento atual; novo binding `key.Backspace` sobe um nível (`filepath.Dir(currentDir)`) limitado à raiz configurada — quando já na raiz, emitir `StatusBarNotifyMsg{Level: NotifyInfo, Text: "Já na raiz", Duration: 2*time.Second}`. Criar `components/sidebar/create.go` contendo uma submodel `createPrompt` com campos `Kind (dir|file) string`, `Input textinput.Model`, `Err string`; acoplado ao `sidebar.Model` como `m.creating *createPrompt`. Atalhos `alt+d` e `alt+f` ativam `createPrompt`; ESC cancela; Enter confirma — validando (não-vazio, caracteres válidos, não-existente) e executando `os.Mkdir` ou `os.WriteFile` empty; em sucesso, recarrega a listagem; se pasta, navega para dentro dela; se arquivo, emite `OpenFileMsg` para abrir no editor externo.

**Rationale**: `textinput.Model` (bubbles) é o primitivo natural para prompt inline sem modal. Todas as operações de fs são rápidas o suficiente para serem feitas dentro de `Update` sem violar 16ms (tipicamente <1ms por arquivo). Erros ficam visíveis no campo `Err` do submodel sem tomar espaço de toda a UI. A raiz configurada já é o working dir do Lumina (comportamento atual mantido).

**Alternatives considered**:
- **Modal full-screen para criação**: rejeitado — viola UX de file manager (nnn/ranger fazem inline).
- **Executar criação via `tea.Cmd` goroutine**: rejeitado — overengineered para I/O síncrono sub-milissegundo.
- **Usar `os.MkdirAll` em vez de `os.Mkdir`**: rejeitado — criar diretórios intermediários ao digitar `a/b/c` é comportamento ambíguo e não pedido pela spec.

---

## R7 — Relógio na status bar

**Decision**: Adicionar à `statusbar.Model` um `time.Time` atualizado via `tea.Tick(30*time.Second, ...)`. No `View()`, formatar `HH:MM` no lado esquerdo ou próximo ao CWD (exato posicionamento a definir em plan visual; atualizar o style existente). Tick inicial em `Init()` retorna um `Cmd` que dispara o primeiro `ClockTickMsg` imediatamente; re-agenda após cada tick.

**Rationale**: 30s basta para cumprir SC-006 ("nunca diverge do relógio do sistema em mais de 60s"). Um ticker mais rápido (1s) seria render waste — o minuto só muda a cada 60s e o usuário não percebe a diferença dentro de um minuto. Reutilizar o tipo `tea.TickMsg` do framework em vez de criar `ClockTickMsg` custom também é uma opção, mas criar um `ClockTickMsg` explícito em `msgs/msgs.go` mantém o padrão "cross-component apenas via msgs tipados".

**Alternatives considered**:
- **Usar `MetricsTickMsg` (já existente) para pegar carona**: considerado — rejeitado porque `MetricsTickMsg` tem `Duration` de 1s, e usá-lo forçaria 60× mais renders do clock do que necessário. Melhor ter cadência própria.
- **Formato HH:MM:SS**: rejeitado — espec pede HH:MM; segundos aumentam render churn sem valor prático.

---

## R8 — Status bar sensível ao terminal focado (git + CWD)

**Decision**: Mover a responsabilidade de "qual branch git exibir" do ticker global de métricas para um refresh event-driven pelo foco. Novo msg `FocusedPaneContextMsg{PaneID, CWD, GitBranch, GitDirty}` é emitido pelo `layout` sempre que o foco muda de painel (ou quando o terminal focado publica um `PaneCWDChangeMsg`, novo, ao detectar OSC 7 / shell hook). A `statusbar.Model` passa a ler contexto desse msg em vez do `MetricsTickMsg`.

**Rationale**: O `MetricsTickMsg` atual carrega `CWD` e `GitBranch` baseados no processo Lumina (não nos painéis filhos) — daí o bug "status bar mostra git errado". Para saber o CWD real de cada pane precisamos ou (a) escutar OSC 7 (`\x1b]7;file://host/path\x07`) no stream do PTY, que zsh/bash emitem com hooks conhecidos (`precmd`), ou (b) ler `/proc/<pid>/cwd` do processo filho (Linux-only, frágil sob subshells). Escolhemos (a): OSC 7 é o mecanismo padrão usado por iTerm2, wezterm, kitty — já foi resolvido pelo prompt padrão que o Lumina injeta com `force_shell_theme`. Para git, após receber `PaneCWDChangeMsg`, o terminal dispara um `tea.Cmd` que executa `git -C <cwd> symbolic-ref --short HEAD` e `git -C <cwd> status --porcelain` (com timeout de 200ms) e publica `PaneGitStateMsg{PaneID, Branch, Dirty}`. O `layout` consolida e reencaminha como `FocusedPaneContextMsg` para a statusbar sempre que o foco aponta para esse pane.

**Alternatives considered**:
- **Rodar `git` continuamente em ticker por painel**: rejeitado — custo de fork/exec em N painéis a cada N segundos; também introduz race com commits/checkouts do usuário.
- **Ignorar dirty status (só mostrar branch)**: rejeitado — spec pede glifo `●/✓` explicitamente.
- **Usar libgit2 ou go-git (parsing direto)**: rejeitado — introduz dependência pesada para um caso simples; `git` CLI está sempre instalado onde o usuário usa `git`.

---

## R9 — Remoção do componente editor embutido sem quebrar mensagens globais

**Decision**: Remoção em duas etapas lógicas (mas um único PR/task): (1) parar de rotear `OpenFileMsg` para `editor.Model` e passar a rotear para uma função `openInExternalEditor(path, cfg)` no `app`; (2) remover `components/editor/` e os msgs órfãos (`EditorResizeMsg`, `ConfirmCloseMsg`, `CloseConfirmedMsg`, `CloseAbortedMsg`). `FocusTarget.FocusEditor` no enum de `msgs` é removido — painéis que antes teriam foco "editor" agora são apenas `FocusTerminal` (pois o editor externo roda em um pane de terminal comum).

**Rationale**: Remover o enum value `FocusEditor` é uma mudança de API interna; como `msgs/msgs.go` não é consumido por terceiros (binário único), o único risco é esquecer algum `switch` exaustivo. A constituição (§V) exige MINOR bump em breaking changes de `msgs`, mas como o componente é interno e a feature 006 é a primeira a usar o campo novo, tratamos como refactor interno (sem bump) — documentamos no DECISIONS.md.

**Alternatives considered**:
- **Deixar `editor.Model` como stub não-usado**: rejeitado — debt morto.
- **Manter `FocusEditor` enum por compatibilidade**: rejeitado — nenhuma compatibilidade externa a preservar.

---

## Resumo dos itens que entram em msgs/msgs.go

Novos:
- `ClickFocusMsg{PaneID int, Target FocusTarget, LocalX, LocalY int}` — resultado do hit-test do layout para um mousedown.
- `SidebarCreateMsg{Kind string, Name string}` — confirmação do prompt de criação (Kind: "dir"|"file").
- `SidebarCreatedMsg{Kind, Path string}` — notificação pós-criação (para recarregar listing e, se file, abrir editor).
- `OpenInExternalEditorMsg{Path string}` — renomeia/substitui o antigo `OpenFileMsg` (ou podemos manter o nome).
- `ClockTickMsg{Now time.Time}` — ticker dedicado do relógio (cadência 30s).
- `PaneCWDChangeMsg{PaneID int, CWD string}` — emitido pelo terminal ao detectar OSC 7.
- `PaneGitStateMsg{PaneID int, Branch string, Dirty bool}` — resultado do git query em background.
- `FocusedPaneContextMsg{PaneID int, CWD, GitBranch string, GitDirty bool}` — reemitido pelo layout à statusbar após troca de foco.

Removidos:
- `EditorResizeMsg`, `ConfirmCloseMsg`, `CloseConfirmedMsg`, `CloseAbortedMsg`
- `FocusTarget.FocusEditor` (enum value removido)

Alterados:
- `OpenFileMsg` pode ser mantido como alias ou renomeado; decisão final em Phase 1 (contracts).
