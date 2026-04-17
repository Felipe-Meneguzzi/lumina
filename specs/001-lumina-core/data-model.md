# Data Model: Lumina TUI Core

**Feature**: 001-lumina-core
**Date**: 2026-04-16

---

## Entities

### AppModel (root)

O modelo raiz que compõe todos os componentes e roteia mensagens.

| Field | Type | Description |
|-------|------|-------------|
| `terminal` | `terminal.Model` | Painel de terminal ativo |
| `sidebar` | `sidebar.Model` | Explorador de arquivos |
| `editor` | `editor.Model` | Editor de texto |
| `statusbar` | `statusbar.Model` | Barra de métricas |
| `focus` | `FocusTarget` | Qual painel está ativo (enum) |
| `width` | `int` | Largura total da janela TUI |
| `height` | `int` | Altura total da janela TUI |

**State transitions**:
- `FocusTarget`: `FocusTerminal` → `FocusSidebar` → `FocusEditor` → `FocusTerminal`
- Resize: `tea.WindowSizeMsg` → propaga dimensões para todos os filhos

---

### terminal.Model

Encapsula um processo PTY e seu output buffer.

| Field | Type | Description |
|-------|------|-------------|
| `pty` | `*os.File` | File descriptor do PTY |
| `cmd` | `*exec.Cmd` | Processo shell em execução |
| `viewport` | `viewport.Model` | Scroll do output (bubbles) |
| `width` | `int` | Colunas disponíveis para o painel |
| `height` | `int` | Linhas disponíveis para o painel |
| `focused` | `bool` | Se está recebendo inputs do teclado |

**Validation rules**:
- `cmd` deve ser inicializado com `$SHELL` ou `/bin/sh` como fallback
- Se PTY encerrar (EOF), recriar automaticamente com novo processo shell
- `width` e `height` DEVEM ser propagados ao PTY via `pty.Setsize` em cada resize

---

### sidebar.Model

Explorador de arquivos hierárquico.

| Field | Type | Description |
|-------|------|-------------|
| `list` | `list.Model` | Componente de lista navegável (bubbles) |
| `root` | `string` | Diretório raiz exibido |
| `cwd` | `string` | Diretório de trabalho atual |
| `expanded` | `map[string]bool` | Estado expand/colapso por caminho |
| `focused` | `bool` | Se está recebendo inputs |
| `width` | `int` | Largura do painel sidebar |

**Validation rules**:
- `root` padrão: diretório onde o Lumina foi invocado (`os.Getwd()`)
- Entradas ocultas (prefixo `.`) são exibidas por padrão mas podem ser filtradas via config
- Ao selecionar um arquivo: emitir `msgs.OpenFileMsg{Path: path}`

---

### editor.Model

Buffer de texto para edição de arquivos.

| Field | Type | Description |
|-------|------|-------------|
| `lines` | `[]string` | Conteúdo do arquivo como slice de linhas |
| `cursor` | `Cursor` | Posição atual `{Row int, Col int}` |
| `path` | `string` | Caminho do arquivo no disco |
| `dirty` | `bool` | Se há alterações não salvas |
| `viewport` | `viewport.Model` | Scroll vertical (bubbles) |
| `focused` | `bool` | Se está recebendo inputs |
| `width` | `int` | Colunas disponíveis |
| `height` | `int` | Linhas disponíveis |

**State transitions**:
```
Closed → Open(path) → Editing → Saved
                     → CloseConfirm (if dirty) → Closed
                                               → Saved → Closed
```

**Validation rules**:
- `dirty` DEVE ser `true` em qualquer modificação ao buffer
- Salvar: `os.WriteFile(path, content, 0644)`
- Fechar com `dirty == true`: emitir `msgs.ConfirmCloseMsg` antes de descartar

---

### statusbar.Model

Exibição de métricas do sistema em tempo real.

| Field | Type | Description |
|-------|------|-------------|
| `cpu` | `float64` | CPU usage em % (0–100) |
| `memUsed` | `uint64` | Memória usada em bytes |
| `memTotal` | `uint64` | Memória total em bytes |
| `cwd` | `string` | Diretório de trabalho atual |
| `gitBranch` | `string` | Branch git atual (vazio se não for repo git) |
| `width` | `int` | Largura total para renderização |

**Validation rules**:
- Atualizar via `tea.Tick` com intervalo de 1 segundo (background goroutine)
- `gitBranch`: detectar via `git rev-parse --abbrev-ref HEAD` ou fallback para string vazia
- Renderização DEVE respeitar `width` — truncar campos se necessário

---

### Cursor

Posição do cursor no editor.

| Field | Type | Description |
|-------|------|-------------|
| `Row` | `int` | Linha (0-indexed) |
| `Col` | `int` | Coluna (0-indexed) |

**Validation rules**:
- `Row` DEVE estar no intervalo `[0, len(lines)-1]`
- `Col` DEVE estar no intervalo `[0, len(lines[Row])]` (após o último char = fim de linha)

---

### Config

Configuração do usuário carregada de `~/.config/lumina/config.toml`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `shell` | `string` | `$SHELL` | Shell a ser executado no terminal |
| `metrics_interval` | `int` | `1000` | Intervalo de atualização da status bar (ms) |
| `show_hidden` | `bool` | `true` | Exibir arquivos ocultos na sidebar |
| `sidebar_width` | `int` | `30` | Largura da sidebar em colunas |
| `theme` | `string` | `"default"` | Nome do tema de cores |

---

## Relationships

```
AppModel
  ├── terminal.Model   (1 instância, sempre ativa)
  ├── sidebar.Model    (1 instância, pode ser oculta em terminais estreitos)
  ├── editor.Model     (1 instância, pode estar Closed)
  └── statusbar.Model  (1 instância, sempre visível)

AppModel.focus → determina qual componente recebe tea.KeyMsg
msgs/msgs.go   → define todos os tea.Msg trocados entre componentes
config.Config  → lida uma vez na inicialização, compartilhada por referência
```

---

## Layout Dimensions

```
┌─────────────────────────────────────────────┐  ← height total
│ sidebar (width=30) │ editor ou terminal      │  ← height - 1
│                    │ (width = total - 30)    │
├────────────────────┴────────────────────────┤
│ statusbar (width = total, height = 1)        │
└─────────────────────────────────────────────┘
```

Quando `total_width < 80`:
- Sidebar fica oculta (`sidebar.visible = false`)
- Todo o espaço vai para o painel ativo
