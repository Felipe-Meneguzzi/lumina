# Research: Mouse Text Selection in Normal Mode

**Feature**: 005-mouse-select-copy
**Date**: 2026-04-17
**Status**: Complete — sem `NEEDS CLARIFICATION` pendentes

---

## R1 — Detecção de modificador Shift nos eventos de mouse do Bubble Tea

**Decision**: Usar `msg.Shift` diretamente em `app.handleMouse` (`tea.MouseMsg` já expõe
`Shift bool`).

**Rationale**:
- A struct `tea.MouseMsg` do Bubble Tea inclui os campos `Shift`, `Alt` e `Ctrl` para
  todos os eventos de mouse. Não é necessária nenhuma dependência adicional.
- O código existente em `mouse.go` já consulta `msg.Shift` ao converter para
  `teaModsToUV` — padrão estabelecido no projeto.
- A detecção em `app.handleMouse` (antes do passthrough PTY) é o único ponto central
  onde o Shift+drag pode ser interceptado sem modificar cada componente.

**Alternatives considered**:
- Interceptar no `terminal.Update()`: não funciona porque os eventos de mouse são
  atualmente roteados para o PTY antes de chegar ao `terminal.Model`.
- Adicionar um flag de estado em `app.Model` para rastrear "drag em andamento": complica
  o estado sem ganho — o `tea.MouseMsg.Action` já distingue Press/Motion/Release.

---

## R2 — Estrutura de estado da seleção de mouse

**Decision**: Novo struct `mouseSelection` em `components/terminal/mouseselect.go`,
análogo ao `copyState` existente em `copymode.go`.

```go
type mouseSelection struct {
    start   pos  // coordenadas pane-local (col, row) do início do drag
    end     pos  // atualizado no Motion, finalizado no Release
    pending bool // true quando mouse_auto_copy=false e aguarda confirmação 'y'
}
```

**Rationale**:
- `copyState` já usa `pos{x, y int}` para coordenadas viewport-local — reutilizar o
  tipo existente é a escolha mais coesa.
- Estado `pending bool` distingue "drag em andamento" de "aguardando confirmação",
  permitindo que `app.go` saiba quando interceptar `y` sem expor lógica interna.
- Arquivo separado `mouseselect.go` mantém a paridade com `copymode.go` e evita inflar
  `terminal.go` além do razoável.

**Alternatives considered**:
- Reutilizar `copyState` para a seleção de mouse: confunde dois modos distintos, aumenta
  complexidade ciclomática e quebraria a invariante de que `copy != nil` significa "copy
  mode ativo".
- Armazenar a seleção em `app.Model`: a seleção é específica de um painel terminal —
  state deve viver no modelo correspondente, conforme o padrão Bubble Tea de estado
  local.

---

## R3 — Novos tea.Msg necessários

**Decision**: Três novos tipos em `msgs/msgs.go`:

```go
// MouseSelectMsg roteia um evento de mouse para seleção Lumina no terminal (não PTY).
// Coordenadas X/Y são pane-local (0,0 = célula superior esquerda do conteúdo interno).
type MouseSelectMsg struct {
    PaneID int
    Mouse  tea.MouseMsg
}

// MouseSelectConfirmMsg confirma uma seleção pendente (copia para clipboard).
type MouseSelectConfirmMsg struct {
    PaneID int
}

// MouseSelectCancelMsg descarta uma seleção pendente sem alterar o clipboard.
type MouseSelectCancelMsg struct {
    PaneID int
}
```

**Rationale**:
- `MouseSelectMsg` segue o padrão de `PtyMouseMsg` existente — mesma estrutura
  (PaneID + Mouse), diferente semântica (seleção Lumina vs. forwarding PTY).
- `Confirm` e `Cancel` são necessários para o caminho `mouse_auto_copy=false`:
  app.go os emite ao interceptar `y` ou `esc`; o terminal os consome para copiar/descartar.
- Três tipos separados em vez de um tipo com campo `Action enum`: mais idiomático em
  Go (type switch) e garante que o compilador valide usos incorretos.
- Constitution II exige integration test para cada novo tipo em `msgs/msgs.go`.

**Alternatives considered**:
- Um único `MouseSelectMsg` com `Action enum (start/update/end/confirm/cancel)`:
  reduz tipos mas piora legibilidade do `Update()` switch.
- Tratar tudo em `app.go` sem novos msgs: não é possível — `app.go` não tem acesso
  direto ao estado interno do `terminal.Model`.

---

## R4 — Configuração mouse_auto_copy

**Decision**: Novo campo `MouseAutoCopy bool `toml:"mouse_auto_copy"`` na struct `Config`
com default `true` em `defaults()`.

**Rationale**:
- Segue o padrão exato de todos os outros campos configuráveis (`ShowHidden`,
  `SidebarWidth`, `ForceShellTheme`): TOML tag, valor default em `defaults()`,
  escritos em `writeDefaults`.
- Default `true` implementa o comportamento "auto-copy" esperado pela maioria dos
  usuários (comportamento padrão de tmux, Alacritty, etc.).
- Usuários com configs existentes sem o campo recebem `true` automaticamente porque
  `cfg := defaults()` é chamado antes de `toml.DecodeFile`.

**Alternatives considered**:
- Flag de runtime (`-mac` / `--mouse-auto-copy`): a spec explicitamente pediu config
  persistente (FR-010), não flag efêmera.
- Variável de ambiente: inconsistente com o modelo de config atual do projeto.

---

## R5 — Renderização com seleção de mouse ativa

**Decision**: Nova função `renderWithMouseSelection()` em `mouseselect.go`, extraindo
a lógica compartilhada de `renderCopyMode()` para uma função helper `renderHighlighted`.

**Rationale**:
- `renderCopyMode()` e a nova função são estruturalmente idênticas — iteram células do
  viewport, aplicam `selectionStyle` à região selecionada. Extrair o loop para um helper
  elimina duplicação.
- Usar a `selectionStyle` existente (`lipgloss.NewStyle().Reverse(true)`) garante visual
  consistente entre copy mode e mouse selection (Constitution III).
- `View()` passa a ter três ramos:
  1. `m.copy != nil` → `renderCopyMode()`
  2. `m.mouseSelection != nil` → `renderWithMouseSelection()`
  3. default → `renderViewport()`

**Alternatives considered**:
- Reutilizar `renderCopyMode()` passando o estado como parâmetro: mudaria a assinatura
  de uma função testada sem necessidade — preferível criar novo entry point.

---

## R6 — Interceção de y/esc com seleção pendente

**Decision**: Em `app.handleKey`, adicionar verificação `m.layout.FocusedHasPendingSelection()`
antes do bloco de forwarding para PTY. Interceptar `y` → `MouseSelectConfirmMsg` e `esc` → `MouseSelectCancelMsg`.

**Rationale**:
- O forwarding para PTY acontece em `app.handleKey`, não em `terminal.Update()`. A
  interceção deve ser no mesmo nível, antes que o input chegue ao PTY.
- Não adiciona novo keybinding global (Constitution III) — é comportamento contextual
  condicional ao estado de seleção pendente.
- Mantém `y` comportando-se normalmente para o PTY quando não há seleção pendente
  (FR-009b).

**Alternatives considered**:
- Interceptar no `terminal.Update()`: não funciona para a tecla `y`, pois `app.go`
  já envia `PtyInputMsg` antes de o terminal processar qualquer key.
- Usar `key.Binding` para o `y` de confirmação: conflita com input normal do PTY para
  todos os outros contextos.

---

## R7 — Comportamento nas condições de borda

**Decision**:
- **Drag release fora dos limites do painel**: tratar como release na última posição
  válida (clamp nas coordenadas do painel).
- **Resize durante seleção**: descartar a seleção (limpar `m.mouseSelection = nil`
  no handler de `tea.WindowSizeMsg`).
- **Foco perdido durante seleção**: descartar ao receber `PaneFocusMsg{Focused: false}`.
- **Aplicação interna alterna mouse tracking**: a verificação de `FocusedMouseEnabled()`
  é feita a cada evento — se a aplicação desabilitar o tracking, o próximo click sem
  Shift já funciona como seleção Lumina normalmente.

**Rationale**:
- Descarte no resize é a abordagem mais simples e robusta: coordenadas viewport-local
  tornam-se inválidas após resize, e recalcular a seleção adicionaria complexidade sem
  benefício visível (o usuário precisaria refazer o drag de qualquer forma).
- Clamp de coordenadas evita panics em índices out-of-bounds sem precisar de tratamento
  de erro especial.

---

## Unknowns resolvidos

Nenhum `NEEDS CLARIFICATION` restante. Todas as decisões de design estão capturadas
acima e nos artefatos de data-model e contratos.
