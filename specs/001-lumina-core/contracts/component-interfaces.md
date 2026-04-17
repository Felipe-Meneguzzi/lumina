# Component Interface Contracts

**Project**: Lumina TUI Core
**Date**: 2026-04-16

Cada componente DEVE implementar `tea.Model` completo e as funções de construção abaixo.
Os contratos aqui são behavior contracts — não impõem implementação específica.

---

## tea.Model Contract (todos os componentes)

```go
// Todos os componentes DEVEM implementar:
Init() tea.Cmd           // retorna Cmds iniciais (ex: ticker, pty reader)
Update(tea.Msg) (tea.Model, tea.Cmd)  // puro: sem efeitos colaterais diretos
View() string            // renderização; DEVE respeitar m.width e m.height
```

**Invariantes**:
- `Update` DEVE retornar `(m, nil)` se a mensagem não for relevante para o componente
- `View` DEVE retornar string com exatamente `m.height` linhas e cada linha ≤ `m.width` colunas
- Nenhum componente DEVE chamar `os.Exit` — finalização via `tea.Quit`

---

## terminal.Model

```go
// Construção
func New(cfg config.Config) (terminal.Model, error)
// cfg.Shell → processo a iniciar; retorna erro se PTY falhar

// Comportamento esperado
// - Init(): retorna waitForPtyOutput(m.pty)
// - Update(PtyOutputMsg): adiciona ao viewport, retorna waitForPtyOutput
// - Update(PtyInputMsg): escreve m.pty, sem Cmd adicional
// - Update(TerminalResizeMsg): chama pty.Setsize, atualiza viewport dimensions
// - Update(tea.KeyMsg) quando focused: converte para PtyInputMsg
// - View(): retorna viewport.View() com borda de foco se focused
```

---

## sidebar.Model

```go
// Construção
func New(root string, cfg config.Config) sidebar.Model

// Comportamento esperado
// - Init(): carrega entries de root via os.ReadDir, popula list.Model; sem Cmds
// - Update(tea.KeyMsg) quando focused: delega para list.Model (up/down/enter)
//   - Enter em dir: toggle expanded, recarrega entries filhas
//   - Enter em arquivo: emite msgs.OpenFileMsg{Path: path}
// - Update(SidebarResizeMsg): atualiza m.width e m.height, propaga para list
// - View(): list.View() com borda de foco se focused
//   - Se m.width == 0: retorna "" (oculto)
```

---

## editor.Model

```go
// Construção
func New(cfg config.Config) editor.Model  // inicia Closed
func (m Model) Open(path string) (editor.Model, tea.Cmd)
// Lê arquivo via os.ReadFile, popula m.lines; retorna Cmd de notificação se erro

// Comportamento esperado
// - Update(tea.KeyMsg) quando focused:
//   - Caracteres imprimíveis: inserir em m.lines[cursor.Row] na posição cursor.Col
//   - Backspace: deletar caractere à esquerda do cursor
//   - Enter: inserir nova linha
//   - Setas: mover cursor (com boundary checks)
//   - Ctrl+S: salvar via os.WriteFile; emitir StatusBarNotifyMsg{Text:"Salvo"}
//   - Ctrl+W / Ctrl+Q: se dirty → emitir ConfirmCloseMsg; caso contrário fechar
// - Update(OpenFileMsg): chamar Open(msg.Path)
// - Update(EditorResizeMsg): atualizar viewport
// - View(): numeros de linha + conteúdo + cursor highlight + borda de foco
```

---

## statusbar.Model

```go
// Construção
func New(cfg config.Config) statusbar.Model

// Comportamento esperado
// - Init(): retorna tickMetrics(1 * time.Second)
// - Update(MetricsTickMsg): atualiza campos, retorna próximo tickMetrics Cmd
// - Update(StatusBarNotifyMsg): sobrescreve display por msg.Duration, depois restaura
// - Update(StatusBarResizeMsg): atualiza m.width
// - View(): string de 1 linha com:
//   "  CPU: 12.3%  MEM: 4.2/16GB  [branch]  ~/dir  "
//   Truncado para m.width colunas
```

---

## app.Model (root)

```go
// Construção
func New(cfg config.Config) (app.Model, error)
// Inicializa todos os componentes filhos; retorna erro se terminal PTY falhar

// Comportamento esperado
// - Init(): tea.Batch de todos os Init() filhos
// - Update(tea.WindowSizeMsg): computa dimensões, emite resize msgs para cada componente
// - Update(tea.KeyMsg):
//   - Atalhos globais (FocusSidebar, FocusTerminal, FocusEditor, Quit): processados aqui
//   - Se focus == FocusTerminal: converte para PtyInputMsg e delega para terminal
//   - Caso contrário: delega ao componente focado
// - Update(msgs.*): roteia para o componente correto via type-switch
// - View(): lip gloss.JoinHorizontal([sidebar, painel_ativo]) + statusbar
```

---

## Keymap Contract (app/keymap.go)

```go
type KeyMap struct {
    FocusTerminal key.Binding  // default: Ctrl+1
    FocusSidebar  key.Binding  // default: Ctrl+2
    FocusEditor   key.Binding  // default: Ctrl+3
    Save          key.Binding  // default: Ctrl+S (editor only)
    Quit          key.Binding  // default: Ctrl+C (fora do modo PTY raw)
    Help          key.Binding  // default: ?
}
```

**Invariantes**:
- Nenhum binding DEVE duplicar outro — validado em testes de configuração
- No modo PTY raw (terminal focado), TODOS os inputs vão direto ao PTY
  exceto o escape sequence de alternância de foco definido em `FocusTerminal`
