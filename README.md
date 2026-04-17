# Lumina

> **"We have Hyprland at home"** — the Hyprland at home.

## Instalação rápida

```bash
curl -fsSL https://raw.githubusercontent.com/menegas/lumina/main/install.sh | bash
```

Detecta OS/arquitetura (Linux/macOS, amd64/arm64), baixa o binário da release mais
recente e instala em `~/.local/bin`. Depois é só rodar `lumina` — no Windows use **WSL 2**.

### Atualizar

Se o Lumina já está instalado, o jeito mais simples é usar o próprio binário:

```bash
lumina --update
```

Consulta a última release do GitHub, compara com a versão instalada e, se houver
atualização, baixa o binário e substitui o atual in-place. Nada a fazer se já estiver
na versão mais recente.

Alternativamente, a mesma linha do instalador também atualiza:

```bash
curl -fsSL https://raw.githubusercontent.com/menegas/lumina/main/install.sh | bash
```

Para saber qual versão está rodando agora:

```bash
lumina --version
```

Opções avançadas, fixar versão, build a partir do fonte: veja [Instalação](#instalação) abaixo.

---

Lumina é um ambiente de trabalho TUI (Terminal User Interface) inspirado no [Hyprland](https://hyprland.org/),
escrito em Go com [Bubble Tea](https://github.com/charmbracelet/bubbletea).

**O alvo primário é o WSL**: usuários que vivem em `wsl` dentro do Windows Terminal e
querem a ergonomia de um window manager tiling (tile panes, chord shortcuts, foco por
teclado, redimensionamento fluido) sem depender de X11/Wayland, sem sair do terminal e
sem perder integração com o shell nativo da distro. Roda também em qualquer Linux/macOS
com um terminal moderno — mas as decisões de keybinding, detecção de shell e
comportamento default são tunadas para o caso WSL + Windows Terminal.

Dentro de uma única instância do Lumina você tem:

- múltiplos terminais reais (PTY), em tiles recursivos lado-a-lado / empilhados
- um editor de texto nativo embutido, com salvamento e proteção de alterações
- um file explorer por pane, com resize por teclado ou mouse
- um monitor de sistema (CPU, memória, branch git, CWD) no rodapé
- copy mode estilo tmux com cópia para o clipboard do host via OSC 52
- mouse passthrough para apps dentro do terminal (vim, htop, lazygit…)

> **Built on [Speckkit](https://github.com/github/spec-kit)** — Lumina usa Speckkit como
> base de desenvolvimento spec-driven. Features são projetadas via specs estruturadas
> (em `specs/`) que dirigem decisões de arquitetura, contratos e tasks de implementação
> antes de qualquer código ser escrito.

---

## Por que "Hyprland para WSL"?

O WSL entrega um Linux excelente em CLI, mas perde toda a camada de window manager
gráfica. Quem vive no terminal tipicamente cola tmux + vim + lazygit + htop num mosaico
de janelas do Windows Terminal, o que funciona mas tem atrito:

- cada terminal é uma sessão desacoplada — sem tiling nativo, sem copy-mode consistente
  entre elas, sem métricas unificadas;
- redimensionar pane requer mouse ou sequência de comandos do próprio Windows Terminal;
- atalhos do WM gráfico (Hyprland `SUPER+arrow`, `SUPER+v`, etc.) não existem.

Lumina mimetiza a experiência de um compositor tiling dentro de um único emulador de
terminal: `alt+b` / `alt+v` dividem, `alt+hjkl` movem foco, `alt+HJKL` redimensionam o
pane focado, `alt+shift+←→↑↓` movem a borda entre panes. A árvore binária de splits segue
o modelo mental de quem já usa Hyprland, i3 ou sway.

---

## Features

### Janelas e layout
- **Binary split tree** (inspirada no Hyprland): splits horizontais e verticais aninhados
  recursivamente, até 4 panes simultâneos
- **Foco espacial por teclado** (`alt+hjkl` ou `alt+arrows`) — o pane vizinho na direção
  da seta recebe foco, respeitando a geometria real
- **Redimensionamento relativo ao foco** (`alt+HJKL`) e **absoluto por borda** (`alt+shift+arrows`)
- **Sidebar por pane** — cada pane pode ter seu próprio file explorer visível/oculto e
  com largura independente

### Terminal
- **PTY real** usando o `$SHELL` do usuário (zsh / bash / fish) via `creack/pty`
- **Emulador VT** completo (`charmbracelet/x/vt`) com suporte a cores 24-bit, estilos e
  modos DEC
- **Scrollback** de 2000 linhas, navegável com `PgUp`/`PgDown` ou `Alt+Wheel` do mouse
- **Copy mode estilo tmux** (`alt+y`): cursor Vim-like (hjkl + `v` + `y`), seleção retangular
  com highlight visual, cópia para o clipboard do host via **OSC 52** — funciona inclusive
  através do SSH/WSL porque o sequência atravessa o terminal hospedeiro
- **Mouse passthrough**: quando o app dentro do terminal ativa mouse tracking (modos DEC
  1000/1002/1003, usados por vim, htop, tmux, lazygit), Lumina encaminha os eventos com
  coordenadas traduzidas para o interior do pane
- **OSC 7 / OSC 0/2**: captura CWD e título reportados pelo shell; reaproveitados pelo
  `Open terminal here` e pela status bar
- **Auto-restart** do shell ao sair (sem derrubar a sessão do Lumina)
- **Tema forçado opcional** (`force_shell_theme`): injeta um prompt oh-my-zsh-inspired
  para uniformizar shells que não têm configuração própria

### Editor
- Abre arquivos no **editor externo** configurado (`nano` por padrão, configurável via campo `editor` no `config.toml`)
- A sidebar inicia o editor externo ao abrir um arquivo; `ctrl+s` funciona dentro do editor externo normalmente

### Mouse
- **Click-to-focus** em qualquer pane (sidebar, editor, terminal)
- **Drag** na borda da sidebar para redimensionar
- **Seleção de texto por mouse** (drag) no terminal — arrastar o mouse seleciona texto com
  highlight visual; ao soltar o botão, o texto é copiado automaticamente para o clipboard
  do host via OSC 52 se `mouse_auto_copy = true` (default). Com `mouse_auto_copy = false`
  uma confirmação aparece na status bar.
- `selection_mode` controla o estilo de seleção: `"linear"` (padrão, estilo bloco de notas)
  ou `"block"` (retangular, estilo vim visual-block)
- **Alt+wheel** para scrollback do terminal (hotkey preservada mesmo quando o app pede
  mouse tracking, servindo como escape hatch)
- **Wheel sem Alt** passa direto para o app dentro do terminal quando ele está em modo
  de mouse

### Status bar
- CPU (%), memória usada/total, branch git, CWD do pane focado
- Título (OSC 0/2) reportado pelo app interno do terminal focado
- Notificações temporárias (save, copy, warnings)
- Ocultável com `alt+m`

---

## Instalação

**Requisitos**: Linux ou macOS. No Windows, use **WSL 2** com Ubuntu / Debian / Fedora —
PTY nativo do Windows não é suportado.

### Opção 1 — one-liner (recomendado)

Baixa o binário da release mais recente e instala em `~/.local/bin` (ou
`/usr/local/bin`, se disponível):

```bash
curl -fsSL https://raw.githubusercontent.com/menegas/lumina/main/install.sh | bash
```

Variáveis de ambiente opcionais:

| Variável | Default | Função |
|---|---|---|
| `LUMINA_VERSION` | `latest` | Tag a instalar (ex: `v0.3.1`) |
| `INSTALL_DIR`    | `~/.local/bin` | Diretório destino |
| `LUMINA_REPO`    | `menegas/lumina` | Override de fork |

Exemplo fixando versão e diretório:

```bash
LUMINA_VERSION=v0.3.1 INSTALL_DIR=/usr/local/bin \
  curl -fsSL https://raw.githubusercontent.com/menegas/lumina/main/install.sh | bash
```

O script detecta OS (`linux` / `darwin`) e arquitetura (`amd64` / `arm64`), valida o
checksum SHA256 (se a release publicar `checksums.txt`) e avisa se o diretório de
instalação não está no `PATH`.

> **Publicando releases**: o installer espera assets nomeados
> `lumina-<os>-<arch>` (ex: `lumina-linux-amd64`) anexados à release no GitHub.
> Opcionalmente um `checksums.txt` com linhas no formato `sha256  lumina-linux-amd64`.

### Opção 2 — build a partir do fonte

Requer Go 1.26+.

```bash
git clone https://github.com/menegas/lumina.git
cd lumina
go build -o lumina .
./lumina
```

Abrir um arquivo específico:

```bash
lumina path/to/file.txt
```

### CLI flags

| Flag | Descrição |
|------|-----------|
| `--update` | Verifica se há nova release no GitHub e instala se houver. |
| `--version`, `-v` | Imprime a versão instalada e sai. |
| `--help`, `-h` | Exibe a ajuda completa e sai. |

Flags de sessão (efêmeras — não alteram `config.toml`):

| Flag | Formato | Default | Descrição |
|------|---------|---------|-----------|
| `-mp` | `-mp N` | 4 | Número máximo de painéis permitidos na sessão. |
| `-sp` | `-sp h<N>` / `-sp v<N>` | 1 painel | Cria `N` painéis iniciais dispostos horizontalmente (`h`) ou verticalmente (`v`). |
| `-sc` | `-sc "<comando>"` | shell default | Executa `<comando>` nos painéis criados por `-sp` (apenas nos iniciais — splits manuais posteriores abrem o shell default). |

Exemplos:

```bash
lumina                                  # boot tradicional: 1 painel, teto 4
lumina -mp 10                           # teto 10, 1 painel inicial
lumina -sp h3                           # 3 painéis lado-a-lado
lumina -sp v2 -sc claude                # 2 painéis empilhados rodando claude
lumina -mp 10 -sp h3 -sc claude         # combinação completa
lumina notes.md -sp h2                  # arquivo + layout customizado
```

Regras de validação:

- `-mp < 1`, `-sp` fora do formato `h<N>`/`v<N>`, ou `-sc ""` abortam a inicialização
  com mensagem em stderr (exit code 2).
- Se `-mp` explícito for menor que `N` de `-sp`, a inicialização é abortada.
- Se `-mp` for omitido e `-sp hN` / `-sp vN` exceder o default (4), o teto efetivo
  sobe automaticamente para `N`.

Ver `lumina --help` para a mensagem de ajuda completa.

### Dica para WSL + Windows Terminal

Alguns atalhos padrão do Windows Terminal (ex: `alt+shift+arrow` para mover pane) são
capturados antes de chegar ao Lumina. Desbinde-os em *Settings → Actions* do Windows
Terminal para liberar o passthrough.

---

## Configuração

Na primeira execução, o Lumina cria dois arquivos em `~/.config/lumina/`:

| Arquivo | Função |
|---|---|
| `config.toml` | Configurações gerais (shell, tema, métricas, sidebar) |
| `keybindings.json` | Mapeamento de teclas de cada ação |

### config.toml

```toml
shell             = "/bin/zsh"   # Executável do shell para os PTYs. Default: $SHELL.
metrics_interval  = 1000         # Taxa de refresh da status bar em ms.
show_hidden       = true         # Mostrar dotfiles na sidebar.
sidebar_width     = 30           # Largura da sidebar em colunas.
theme             = "default"    # Tema da UI.
force_shell_theme = true         # Injeta o prompt customizado do Lumina no shell.
mouse_auto_copy   = true         # Copia automaticamente ao soltar o botão do mouse na seleção.
selection_mode    = "linear"     # Estilo de seleção: "linear" (padrão) ou "block" (retangular).
editor            = "nano"       # Editor externo usado pela sidebar ("nano"|"vim"|"nvim"|caminho absoluto).
```

No **WSL**, se `shell` apontar para um executável Windows (`.exe`), o Lumina rejeita
automaticamente e faz fallback para o primeiro POSIX shell disponível, avisando na
status bar.

### keybindings.json

Cada ação mapeia para uma lista de teclas — qualquer uma delas dispara a ação. A notação
segue a do Bubble Tea: `"ctrl+s"`, `"alt+h"`, `"f1"`, `"?"`.

```json
{
  "toggle_sidebar":   ["alt+e"],
  "split_horizontal": ["alt+b", "alt+|"],
  "enter_copy_mode":  ["alt+y"],
  "sidebar_new_dir":  ["alt+d"],
  "sidebar_new_file": ["alt+f"],
  "sidebar_parent":   ["backspace"]
}
```

Só inclua as ações que quiser sobrescrever; o restante herda o default.

---

## Keybindings

### Foco

| Ação | Tecla default | Descrição |
|---|---|---|
| Focar sidebar | `alt+1` / `f1` / `ctrl+1` | Move o foco para o file explorer |
| Focar terminal | `alt+2` / `f2` / `ctrl+2` | Move o foco para o terminal |
| Focar editor | `alt+3` / `f3` / `ctrl+3` | Move o foco para o editor |
| Abrir terminal aqui | `ctrl+t` | Novo terminal no CWD do pane ativo |

### Gerenciamento de panes

| Ação | Tecla default | Descrição |
|---|---|---|
| Split horizontal | `alt+b` | Divide o pane ativo lado-a-lado |
| Split vertical | `alt+v` | Divide o pane ativo empilhado |
| Fechar pane | `alt+q` | Fecha o pane; o irmão expande |

### Navegação entre panes

| Ação | Tecla default | Descrição |
|---|---|---|
| Foco para a esquerda | `alt+h` / `alt+←` | Move foco para o pane à esquerda |
| Foco para a direita | `alt+l` / `alt+→` | Move foco para o pane à direita |
| Foco para cima | `alt+k` / `alt+↑` | Move foco para o pane acima |
| Foco para baixo | `alt+j` / `alt+↓` | Move foco para o pane abaixo |

### Redimensionar — relativo ao pane focado

Mexe a borda adjacente ao pane ativo.

| Ação | Tecla default | Descrição |
|---|---|---|
| Crescer pane à direita | `alt+L` | Amplia o pane empurrando a borda direita |
| Diminuir pane à esquerda | `alt+H` | Estreita o pane puxando a borda direita |
| Crescer pane para baixo | `alt+J` | Amplia verticalmente empurrando a borda inferior |
| Diminuir pane para cima | `alt+K` | Encolhe verticalmente puxando a borda inferior |

### Redimensionar — borda absoluta

Setas movem a borda mais próxima na direção da tecla, independente de foco.

| Ação | Tecla default | Descrição |
|---|---|---|
| Borda → | `alt+shift+→` | Empurra a divisória vertical para a direita |
| Borda ← | `alt+shift+←` | Empurra a divisória vertical para a esquerda |
| Borda ↓ | `alt+shift+↓` | Empurra a divisória horizontal para baixo |
| Borda ↑ | `alt+shift+↑` | Empurra a divisória horizontal para cima |

> **WSL**: `alt+shift+arrow` pode estar capturado pelo Windows Terminal ("move pane").
> Desbinde nas configurações do Windows Terminal para o passthrough funcionar.

### Sidebar

| Ação | Tecla default | Descrição |
|---|---|---|
| Crescer sidebar | `alt+}` | +1 coluna |
| Diminuir sidebar | `alt+{` | −1 coluna |
| Toggle sidebar | `alt+e` | Mostra/oculta a sidebar do pane ativo |
| Subir para diretório pai | `backspace` | Navega para o diretório pai (apenas com sidebar em foco) |
| Novo diretório | `alt+d` | Cria novo diretório no local atual |
| Novo arquivo | `alt+f` | Cria novo arquivo no local atual |

### Copy mode (terminal)

Entra em um modo estilo tmux para selecionar texto com teclado e copiar para o clipboard
do host via OSC 52.

| Ação | Tecla default | Descrição |
|---|---|---|
| Entrar em copy mode | `alt+y` | Inicia seleção no canto inferior direito do pane |
| Mover cursor | `h` `j` `k` `l` ou setas | Movimento Vim-like |
| Estender seleção | `H` `J` `K` `L` / `shift+setas` | Âncora fixa, cursor se move |
| Alternar âncora | `v` | Redefine a âncora na posição do cursor |
| Ir para início/fim de linha | `0` / `$` | `home` / `end` |
| Ir para topo/base | `g` / `G` | — |
| Copiar e sair | `y` / `enter` | Envia OSC 52; mostra confirmação na status bar |
| Cancelar | `esc` / `q` / `ctrl+c` | Sai sem copiar |

Enquanto em copy mode a borda do pane vira **amarela** e todo teclado é consumido — o
shell não recebe nada até o modo terminar. O viewport congela: novo output do shell é
preservado em scrollback para você não perder o conteúdo selecionado.

### Scrollback do terminal

| Ação | Tecla default | Descrição |
|---|---|---|
| Subir no histórico | `PgUp` | 10 linhas |
| Descer no histórico | `PgDown` | 10 linhas |
| Subir 3 linhas | `alt+wheel up` | Rola para o histórico |
| Descer 3 linhas | `alt+wheel down` | Rola em direção ao live |
| Voltar ao live | Digitar qualquer tecla | Sai do modo scroll |

Em apps que usam alt-screen (vim, less, htop) o scrollback fica desativado — é o
comportamento correto, porque alt-screen nunca alimenta o histórico.

### Arquivo e aplicativo

| Ação | Tecla default | Descrição |
|---|---|---|
| Salvar arquivo | `ctrl+s` | Salva o arquivo aberto no editor ativo |
| Sair | `ctrl+c` | Encerra o Lumina (pede confirmação se houver não-salvos) |
| Ajuda | `?` | Abre a overlay de atalhos |
| Toggle status bar | `alt+m` | Mostra/oculta a barra de métricas |

---

## Arquitetura

Lumina segue a arquitetura Elm (Model / Update / View) via Bubble Tea. Cada painel é um
`tea.Model` independente, composto pelo `app.Model` raiz através de delegação e
mensagens tipadas — **sem imports circulares, sem estado global mutável**.

```
lumina/
├── main.go
├── app/
│   ├── app.go             # Model raiz — roteia mensagens entre componentes
│   └── keymap.go          # ÚNICO lugar onde bindings são declarados
├── components/
│   ├── layout/            # Binary split tree (Hyprland-inspired)
│   │   ├── layout.go      # Model, Update, View do gerenciador de panes
│   │   ├── tree.go        # Inserção, remoção e walk recursivos
│   │   ├── focus.go       # Busca espacial de vizinho por direção
│   │   ├── bounds.go      # Cálculo de retângulos de cada pane
│   │   └── render.go      # Composição final em string com bordas
│   ├── terminal/
│   │   ├── terminal.go    # Model principal + ciclo de vida do PTY
│   │   ├── scrollback.go  # Render composto (scrollback + live)
│   │   ├── copymode.go    # Estado + render do copy mode + OSC 52
│   │   ├── mouseselect.go # Seleção de texto por mouse com highlight + OSC 52
│   │   ├── mouse.go       # Callbacks do emulador (DEC modes, title, CWD, bell)
│   │   ├── keys.go        # Tradução de tea.KeyMsg → bytes do PTY
│   │   └── theme.go       # Injeção opcional do prompt customizado
│   ├── sidebar/           # File explorer (bubbles/list + os.ReadDir) + criação de arquivos/dirs
│   └── statusbar/         # Métricas (gopsutil ticker)
├── msgs/
│   └── msgs.go            # TODOS os tea.Msg customizados
├── config/
│   ├── config.go
│   └── keybindings.go
├── specs/                 # Spec-kit: specs e contratos de cada feature
└── tests/integration/     # Testes de fluxo cross-component
```

### Stack de bibliotecas

| Camada | Lib | Uso |
|---|---|---|
| TUI framework | [bubbletea](https://github.com/charmbracelet/bubbletea) | Runtime Model/Update/View |
| Estilo | [lipgloss](https://github.com/charmbracelet/lipgloss) | Bordas, cores, layout |
| Widgets | [bubbles](https://github.com/charmbracelet/bubbles) | viewport, list, help |
| Emulador VT | [charmbracelet/x/vt](https://github.com/charmbracelet/x) | Parser de escape sequences, scrollback, DEC modes |
| Render de células | [charmbracelet/ultraviolet](https://github.com/charmbracelet/ultraviolet) | Acesso a glyphs estilizados |
| PTY | [creack/pty](https://github.com/creack/pty) | fork+exec com pseudo-terminal |
| Métricas | [gopsutil/v3](https://github.com/shirou/gopsutil) | CPU, memória |
| Clipboard | [go-osc52](https://github.com/aymanbagabas/go-osc52) | Cópia via OSC 52 |
| Config | [BurntSushi/toml](https://github.com/BurntSushi/toml) | TOML parsing |

### Regras de arquitetura

- Todo I/O assíncrono (leituras de PTY, ticker, leituras de arquivo) **deve** ser
  `tea.Cmd` — nunca bloqueie `Update()`.
- `Update()` **deve** retornar em ≤16ms (orçamento de frame para ≥30 FPS).
- `View()` **deve** retornar string com exatamente `m.height` linhas e cada linha ≤
  `m.width` colunas.
- Keybindings **apenas** em `app/keymap.go` via `key.Binding`.
- Comunicação cross-component **apenas** via tipos em `msgs/msgs.go`.
- Estilos **apenas** via Lip Gloss — ANSI bruto é proibido fora de `components/terminal/`.
- PTY resize propaga via `pty.Setsize` sempre que chega `tea.WindowSizeMsg`.

---

## Desenvolvimento

```bash
# Rodar testes
go test ./...

# Lint
golangci-lint run

# Build
go build -o lumina .
```

Cada `tea.Model` exportado tem testes unitários isolados (sem dependência de outros
componentes), alimentados com `tea.Msg` sintéticos. Testes de integração em
`tests/integration/` cobrem novos `tea.Msg` adicionados a `msgs/msgs.go`.

---
