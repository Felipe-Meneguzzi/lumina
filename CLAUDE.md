# Lumina Development Guidelines

Auto-generated from feature plan 001-lumina-core. Last updated: 2026-04-17

## Project Overview

Lumina é um editor TUI em Go (estilo VSCode para terminal) com Bubble Tea.
Arquitetura: Componentes Compostos (Elm Model/Update/View) — cada painel é um `tea.Model`
independente, composto pelo `app.Model` raiz via delegação e mensagens tipadas.

## Active Technologies
- Go 1.26 + Bubble Tea, Lip Gloss, Bubbles (Charm), creack/pty, gopsutil/v3 (003-ux-fixes-multiwindow)
- N/A (sem persistência nova) (003-ux-fixes-multiwindow)
- Go 1.26 (já em uso no projeto — `go.mod`) + Bubble Tea, Lip Gloss, Bubbles, creack/pty, gopsutil/v3 (sem (004-cli-startup-flags)
- N/A — flags são efêmeras e não alteram `config.toml` (004-cli-startup-flags)
- Go 1.26 (já em uso — `go.mod`) + Bubble Tea, Lip Gloss, creack/pty, ultraviolet (charmbracelet/x/vt) — sem dependências novas (main)
- `~/.config/lumina/config.toml` — novo campo `mouse_auto_copy bool` na struct `Config` (main)
- Go 1.26.1 (conforme `go.mod`) + Bubble Tea v1.3.10, Lip Gloss v1.1.0, Bubbles v1.0.0, charmbracelet/ultraviolet + x/vt (emulação de terminal), creack/pty v1.1.24, gopsutil/v3 v3.24.5, BurntSushi/toml v1.6.0 (006-ux-polish-pack)
- `~/.config/lumina/config.toml` (novos campos `editor string`, opcionais); `~/.config/lumina/keybindings.json` (novas bindings para click-focus, alt+d, alt+f, backspace) (006-ux-polish-pack)

- **Language**: Go 1.26
- **TUI Framework**: Bubble Tea + Lip Gloss + Bubbles (Charm ecosystem)
- **PTY**: creack/pty (Linux/macOS only)
- **Metrics**: gopsutil/v3
- **Config**: BurntSushi/toml
- **Storage**: os.ReadFile / os.WriteFile (sistema de arquivos)

## Project Structure

```text
lumina/
├── main.go
├── app/
│   ├── app.go          # Model raiz — roteia mensagens entre componentes
│   └── keymap.go       # ÚNICO lugar para key.Binding — nunca hardcode em componentes
├── components/
│   ├── terminal/       # PTY wrapper (creack/pty)
│   ├── sidebar/        # File explorer (bubbles/list + os.ReadDir)
│   ├── editor/         # Text buffer ([]string + bubbles/viewport)
│   └── statusbar/      # Métricas em tempo real (gopsutil ticker)
├── msgs/
│   └── msgs.go         # TODOS os tea.Msg customizados — cross-component communication
├── config/
│   └── config.go       # Config struct + LoadConfig()
└── tests/
    └── integration/    # Message flow integration tests
```

## Commands

```bash
# Build
go build -o lumina .

# Run
./lumina
./lumina arquivo.txt

# Test
go test ./...

# Lint
golangci-lint run
```

## Code Style

- **Go**: `gofmt` obrigatório; `golangci-lint` com ruleset padrão no CI
- **Naming**: Exported PascalCase, unexported camelCase (convenção Go padrão)
- **Functions**: Máximo 40 linhas de código não-comentário; complexidade ciclomática ≤10
- **State**: Sem estado global mutável fora do pacote `config/`
- **Styles**: Usar Lip Gloss para tudo — ANSI escape codes diretos são PROIBIDOS
  (exceção: camada PTY raw em `components/terminal/`)
- **Messages**: Cross-component APENAS via tipos em `msgs/msgs.go` — sem imports circulares

## Architecture Rules

- Todo I/O assíncrono (PTY reads, ticker, file reads) DEVE ser `tea.Cmd` — nunca bloqueie `Update()`
- `Update()` DEVE retornar em ≤16ms (orçamento de frame para ≥30 FPS)
- `View()` DEVE retornar string com exatamente `m.height` linhas e cada linha ≤ `m.width` colunas
- Keybindings: APENAS em `app/keymap.go` via `key.Binding` do pacote bubbles
- PTY resize: propagar via `pty.Setsize` sempre que receber `tea.WindowSizeMsg`

## Testing Requirements

- Cada `tea.Model` exportado DEVE ter unit tests isolados (sem dependência de outros componentes)
- Feed de mensagens sintéticas (`tea.Msg`) para testar `Update()` — não mockar a TUI inteira
- Integration tests em `tests/integration/` para cada novo `tea.Msg` adicionado a `msgs/msgs.go`
- `go test ./...` DEVE passar com zero falhas antes de qualquer merge

## Spec Artifacts

### Feature 001: Lumina Core (baseline)
- **Spec**: `specs/001-lumina-core/spec.md`
- **Plan**: `specs/001-lumina-core/plan.md`

### Feature 002: Multiwindow Layout
- **Spec**: `specs/002-multiwindow/spec.md`
- **Plan**: `specs/002-multiwindow/plan.md`
- **Research**: `specs/002-multiwindow/research.md`
- **Data Model**: `specs/002-multiwindow/data-model.md`
- **Contracts**: `specs/002-multiwindow/contracts/`
- **Quickstart**: `specs/002-multiwindow/quickstart.md`
- **Tasks**: `specs/002-multiwindow/tasks.md`

### Feature 003: UX Fixes Multi-Window ← ACTIVE
- **Spec**: `specs/003-ux-fixes-multiwindow/spec.md`
- **Plan**: `specs/003-ux-fixes-multiwindow/plan.md`
- **Research**: `specs/003-ux-fixes-multiwindow/research.md`
- **Data Model**: `specs/003-ux-fixes-multiwindow/data-model.md`
- **Contracts**: `specs/003-ux-fixes-multiwindow/contracts/`
- **Quickstart**: `specs/003-ux-fixes-multiwindow/quickstart.md`

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->

## Recent Changes
- 006-ux-polish-pack: Added Go 1.26.1 (conforme `go.mod`) + Bubble Tea v1.3.10, Lip Gloss v1.1.0, Bubbles v1.0.0, charmbracelet/ultraviolet + x/vt (emulação de terminal), creack/pty v1.1.24, gopsutil/v3 v3.24.5, BurntSushi/toml v1.6.0
- main: Added Go 1.26 (já em uso — `go.mod`) + Bubble Tea, Lip Gloss, creack/pty, ultraviolet (charmbracelet/x/vt) — sem dependências novas
- 004-cli-startup-flags: Added Go 1.26 (já em uso no projeto — `go.mod`) + Bubble Tea, Lip Gloss, Bubbles, creack/pty, gopsutil/v3 (sem
