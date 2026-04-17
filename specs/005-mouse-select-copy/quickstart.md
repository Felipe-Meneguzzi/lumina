# Quickstart: Mouse Text Selection in Normal Mode

**Feature**: 005-mouse-select-copy
**Date**: 2026-04-17

---

## Para o usuário

### Como selecionar e copiar texto com o mouse (modo normal)

**Sem precisar entrar no copy mode:**

1. Foque o painel terminal (clique nele ou use `Ctrl+hjkl` para navegar)
2. Clique e arraste sobre o texto que deseja copiar
3. O texto selecionado fica destacado em vídeo invertido
4. **Solte o botão do mouse** → texto copiado automaticamente para o clipboard ✓

**Se a aplicação interna usar mouse** (vim com `mouse=a`, htop, etc.):

1. Segure **Shift** antes de iniciar o arrasto
2. Shift+clique+arraste → seleção Lumina (ignora a aplicação interna)
3. Solte → copia automaticamente ✓

---

### Modo de confirmação manual (`mouse_auto_copy = false`)

Para usuários que não querem cópias automáticas ao soltar o mouse:

1. Edite `~/.config/lumina/config.toml`:
   ```toml
   mouse_auto_copy = false
   ```
2. Reinicie o Lumina
3. Agora, após o arrasto, a seleção fica visível aguardando confirmação:
   - Pressione **`y`** → copia e limpa a seleção
   - Pressione **`Esc`** ou clique fora → cancela sem copiar

---

### Copy mode (alternativa para quem não usa mouse)

O copy mode existente permanece inalterado para usuários somente-teclado:
- Acione pelo atalho padrão → navegue com `h/j/k/l` → selecione com `Shift+hjkl` → confirme com `y`
- Nenhuma alteração no comportamento do copy mode

---

## Para o desenvolvedor

### Resumo das mudanças

```
config/config.go
  + MouseAutoCopy bool `toml:"mouse_auto_copy"`  (default: true)

msgs/msgs.go
  + MouseSelectMsg{PaneID int, Mouse tea.MouseMsg}
  + MouseSelectConfirmMsg{PaneID int}
  + MouseSelectCancelMsg{PaneID int}

components/terminal/mouseselect.go  [NOVO]
  + type mouseSelection struct{start, end pos; pending bool}
  + startMouseSelection(x, y int)
  + updateMouseSelection(x, y int)
  + finalizeMouseSelection(x, y int, autoCopy bool) tea.Cmd
  + confirmMouseSelection() tea.Cmd
  + cancelMouseSelection()
  + extractMouseSelection() string
  + renderWithMouseSelection() string
  + HasMouseSelection() bool
  + HasPendingSelection() bool

components/terminal/terminal.go
  Model.mouseSelection *mouseSelection
  Update() → handle MouseSelectMsg, MouseSelectConfirmMsg, MouseSelectCancelMsg
  View()   → ramo adicional: renderWithMouseSelection() quando mouseSelection != nil

components/layout/layout.go
  + FocusedHasMouseSelection() bool
  + FocusedHasPendingSelection() bool

app/app.go
  handleMouse() → detecta drag Lumina (sem PTY tracking || Shift) → emite MouseSelectMsg
  handleKey()   → intercepta y/esc quando FocusedHasPendingSelection()

tests/integration/mouse_select_test.go  [NOVO]
  Integration tests para os 3 novos tea.Msg
```

### Fluxo de dados resumido

```
Usuário arrasta mouse
        │
app.handleMouse detects: sem PTY tracking (ou Shift)
        │
        ▼
MouseSelectMsg{PaneID, Mouse{Action:Press/Motion/Release, X_local, Y_local}}
        │
layout.Update → terminal.Update
        │
terminal.mouseSelection atualizada
        │
Release:
  auto_copy=true  → copyToClipboard → StatusBarNotifyMsg
  auto_copy=false → pending=true → highlight permanece
        │
app.handleKey detecta FocusedHasPendingSelection()
  y   → MouseSelectConfirmMsg → terminal: copy + clear
  esc → MouseSelectCancelMsg  → terminal: clear
```

### Executar e testar

```bash
# Build
go build -o lumina .

# Executar
./lumina

# Testes (zero falhas obrigatório antes de merge)
go test ./...

# Lint
golangci-lint run
```

### Configuração de desenvolvimento

Para testar os dois modos rapidamente sem editar `config.toml`:
```bash
# Simular mouse_auto_copy=false: edite config.toml antes de iniciar
echo 'mouse_auto_copy = false' >> ~/.config/lumina/config.toml
./lumina
```
