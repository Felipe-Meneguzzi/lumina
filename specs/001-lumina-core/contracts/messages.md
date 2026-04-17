# Component Message Contracts: msgs/msgs.go

**Project**: Lumina TUI Core
**Date**: 2026-04-16
**Location**: `msgs/msgs.go`

Todos os `tea.Msg` customizados do Lumina são definidos neste arquivo.
Componentes NUNCA importam uns aos outros diretamente — comunicação é exclusivamente via mensagens.

---

## FocusChangeMsg

Emitida quando o usuário alterna o foco entre painéis.

```go
type FocusChangeMsg struct {
    Target FocusTarget // Terminal | Sidebar | Editor
}

type FocusTarget int

const (
    FocusTerminal FocusTarget = iota
    FocusSidebar
    FocusEditor
)
```

**Emitida por**: `app.Model.handleKey()` ao receber atalhos de alternância de foco
**Consumida por**: `app.Model.Update()` — atualiza `m.focus` e propaga `tea.FocusMsg`/`tea.BlurMsg`

---

## PtyOutputMsg

Carrega bytes lidos do PTY para o loop principal.

```go
type PtyOutputMsg struct {
    Data []byte
    Err  error
}
```

**Emitida por**: `terminal.waitForPtyOutput(pty *os.File) tea.Cmd` (goroutine de leitura)
**Consumida por**: `terminal.Model.Update()` — adiciona ao viewport e re-enfileira o Cmd de leitura

---

## PtyInputMsg

Envia input do usuário para o PTY quando o terminal está focado.

```go
type PtyInputMsg struct {
    Data []byte
}
```

**Emitida por**: `app.Model.Update()` ao receber `tea.KeyMsg` com `m.focus == FocusTerminal`
**Consumida por**: `terminal.Model.Update()` — escreve bytes no `pty.File`

---

## WindowResizeMsg (wrapper interno)

Propaga dimensões computadas para cada componente após o resize da janela.

```go
type TerminalResizeMsg struct {
    Width  int
    Height int
}

type SidebarResizeMsg struct {
    Width  int
    Height int
}

type EditorResizeMsg struct {
    Width  int
    Height int
}
```

**Emitida por**: `app.Model.Update()` ao receber `tea.WindowSizeMsg`
**Consumida por**: Cada componente respectivo — atualiza dimensões e propaga ao viewport/pty

---

## MetricsTickMsg

Transporta snapshot de métricas do sistema coletadas em background.

```go
type MetricsTickMsg struct {
    CPU      float64   // 0.0–100.0
    MemUsed  uint64    // bytes
    MemTotal uint64    // bytes
    CWD      string    // diretório atual
    GitBranch string   // branch git ou ""
    Tick     time.Time // timestamp da coleta
}
```

**Emitida por**: `statusbar.tickMetrics(interval) tea.Cmd` — `tea.Tick` goroutine
**Consumida por**: `statusbar.Model.Update()` — atualiza campos e re-enfileira o próximo tick

---

## OpenFileMsg

Solicita abertura de um arquivo no editor.

```go
type OpenFileMsg struct {
    Path string // caminho absoluto do arquivo
}
```

**Emitida por**: `sidebar.Model.Update()` ao selecionar um arquivo com Enter
**Consumida por**: `editor.Model.Update()` — lê o arquivo e popula o buffer

---

## ConfirmCloseMsg / CloseConfirmedMsg / CloseAbortedMsg

Fluxo de confirmação ao fechar o editor com alterações não salvas.

```go
type ConfirmCloseMsg struct{}   // editor solicita confirmação
type CloseConfirmedMsg struct{} // usuário confirmou descartar
type CloseAbortedMsg struct{}   // usuário cancelou
```

**Emitida por**: `editor.Model.Update()` quando `dirty == true` e close é solicitado
**Consumida por**: `app.Model.Update()` — exibe dialog de confirmação e emite a resposta

---

## StatusBarNotifyMsg

Exibe mensagem temporária na status bar (ex: "Arquivo salvo", "Erro ao ler métricas").

```go
type StatusBarNotifyMsg struct {
    Text     string
    Level    NotifyLevel // Info | Warning | Error
    Duration time.Duration
}

type NotifyLevel int

const (
    NotifyInfo NotifyLevel = iota
    NotifyWarning
    NotifyError
)
```

**Emitida por**: Qualquer componente
**Consumida por**: `statusbar.Model.Update()` — exibe por `Duration` e depois retorna ao modo normal
