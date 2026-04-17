# Research: Lumina TUI Core

**Feature**: 001-lumina-core
**Date**: 2026-04-16
**Source**: DECISIONS.md (projeto já tem decisões técnicas ratificadas)

---

## 1. Linguagem e Runtime

**Decision**: Go (latest stable minor)
**Rationale**: Binário único sem runtime externo; goroutines para concorrência de painéis;
ecossistema TUI maduro. Decisão ratificada em DECISIONS.md § 1.
**Alternatives considered**: Rust (sem borrowing sem GC overhead), Python (runtime pesado),
C++ (complexidade de build)

---

## 2. Framework TUI

**Decision**: Bubble Tea (Model/Update/View) + Lip Gloss (styles) + Bubbles (ready-made components)
**Rationale**: Arquitetura Elm garante fluxo unidirecional — cada painel é um `tea.Model`
independente, testável isoladamente. Lip Gloss abstrai ANSI sem escape codes manuais.
Bubbles fornece viewport (scroll) e textinput prontos. Decisão ratificada em DECISIONS.md § 2.
**Alternatives considered**: tview/tcell (mais controle mas sem arquitetura clara),
termui (sem manutenção ativa), gocui (não suporta composição de modelos)

**Pattern: Composite Components (Pattern B)**
Cada componente implementa `tea.Model` completo. `app.Model` (root) delega via type-switch:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKey(msg)
    case msgs.MetricsTickMsg:
        m.statusbar, cmd = m.statusbar.Update(msg)
    case msgs.FocusChangeMsg:
        m.setFocus(msg.Target)
    }
    // delegate to focused component
}
```

---

## 3. PTY Management

**Decision**: `creack/pty`
**Rationale**: API canônica para PTY em Go. `pty.Start(cmd)` cria o processo com PTY real;
`pty.Setsize(fd, rows, cols)` propaga resize. Suporta todos os programas interativos (vim, htop).
Decisão ratificada em DECISIONS.md § 3.
**Alternatives considered**: `mvdan/sh` (shell emulado, não PTY real), syscall direto (reimplementa
o que creack já resolve)

**PTY integration pattern com Bubble Tea**:
O output do PTY é lido em uma goroutine separada e enviado ao loop via `tea.Cmd`:

```go
func waitForPtyOutput(pty *os.File) tea.Cmd {
    return func() tea.Msg {
        buf := make([]byte, 4096)
        n, err := pty.Read(buf)
        return msgs.PtyOutputMsg{Data: buf[:n], Err: err}
    }
}
```

Resize: ao receber `tea.WindowSizeMsg`, propagar para PTY:

```go
case tea.WindowSizeMsg:
    pty.Setsize(m.pty, &pty.Winsize{
        Rows: uint16(msg.Height - statusBarHeight),
        Cols: uint16(msg.Width - sidebarWidth),
    })
```

---

## 4. Métricas do Sistema

**Decision**: `github.com/shirou/gopsutil/v3`
**Rationale**: CPU, memória, disco com uma linha por métrica. Cross-platform. Latência de
~100ms imperceptível em ciclos de 1s. Decisão ratificada em DECISIONS.md § 4.
**Alternatives considered**: `/proc` direto (Linux-only, verbose), biblioteca custom (sem benefício)

**Ticker pattern com Bubble Tea**:

```go
func tickMetrics(interval time.Duration) tea.Cmd {
    return tea.Tick(interval, func(t time.Time) tea.Msg {
        cpu, _ := cpu.Percent(0, false)
        mem, _ := mem.VirtualMemory()
        return msgs.MetricsTickMsg{
            CPU:    cpu[0],
            MemPct: mem.UsedPercent,
            Time:   t,
        }
    })
}
```

---

## 5. Editor de Texto (Buffer)

**Decision**: Implementação própria com `strings.Builder` + slice de linhas
**Rationale**: Editor simples (sem syntax highlighting em v1) não justifica dependência externa.
Representação como `[]string` (slice de linhas) é suficiente para arquivos até 10k linhas.
**Alternatives considered**: `go-text/rope` (overkill para v1), `charm/x/editor` (não estável)

**Buffer model**:
```
type Buffer struct {
    lines  []string
    cursor Cursor   // {Row, Col}
    dirty  bool
    path   string
}
```

Operações: inserir caractere na posição do cursor, deletar, mover cursor, scroll via
`bubbles/viewport`.

---

## 6. Explorador de Arquivos (Sidebar)

**Decision**: Implementação própria com `os.ReadDir` + `bubbles/list`
**Rationale**: `bubbles/list` fornece navegação por teclado e scrolling. `os.ReadDir` é
stdlib. Junção deles cobre 100% dos requisitos de US2.
**Alternatives considered**: `charmbracelet/filetree` (experimental, instável)

---

## 7. Configuração de Keybindings

**Decision**: `app/keymap.go` com `charmbracelet/bubbles/key` para binding definitions
**Rationale**: `key.Binding` do pacote bubbles é o padrão idiomático — integra com
`help.Model` para exibir atalhos automaticamente. Centralizado em um arquivo elimina
conflitos.

```go
type KeyMap struct {
    FocusTerminal key.Binding
    FocusSidebar  key.Binding
    FocusEditor   key.Binding
    Save          key.Binding
    Quit          key.Binding
}
```

---

## 8. Configuração do Usuário

**Decision**: TOML via `BurntSushi/toml` — arquivo em `~/.config/lumina/config.toml`
**Rationale**: TOML é legível, suportado por `BurntSushi/toml` (biblioteca canônica em Go),
e compatível com a estrutura de config simples do Lumina.
**Alternatives considered**: JSON (sem comentários), YAML (complexo para config simples)

---

## 9. Resolução de NEEDS CLARIFICATION (da spec)

Nenhum marcador `[NEEDS CLARIFICATION]` encontrado na spec. Todos os aspectos cobertos
pelas decisões em DECISIONS.md ou pelos defaults documentados em Assumptions.

**Edge cases mapeados**:
- Shell exit: `tea.Quit()` ou reiniciar shell — decidido: reiniciar shell (melhor UX)
- Terminal estreito (<80 cols): layout adaptativo, sidebar oculta abaixo de threshold
- Edição concorrente externa: mostrar aviso na status bar, recarregar com confirmação
- Conflito de atalhos com programas PTY: quando terminal tem foco, todos os inputs
  são passados direto ao PTY; atalhos globais só funcionam fora do modo PTY raw
