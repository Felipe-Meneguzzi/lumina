# Data Model: CLI Startup Flags

**Feature**: 004-cli-startup-flags
**Date**: 2026-04-17

Este documento descreve as novas estruturas internas e mudanças em estruturas
existentes. Nada aqui é persistido — todos os dados abaixo vivem apenas em memória
durante uma sessão.

---

## Novo: `cli.StartupOverrides`

**Pacote**: `github.com/menegas/lumina/cli`

```go
type StartupOverrides struct {
    MaxPanes     int    // 0 = usar default (4). >0 = valor explícito do usuário.
    StartPanes   int    // 0 = sem pré-split (default 1 painel). >=1 = nº de panes iniciais.
    StartOrient  Orient // OrientNone | OrientHorizontal | OrientVertical
    StartCommand string // "" = usar shell default. Caso contrário, comando a rodar nos panes iniciais.
    FilePath     string // argumento posicional, se houver (compatibilidade com lumina <arquivo>)
}

type Orient int
const (
    OrientNone Orient = iota
    OrientHorizontal
    OrientVertical
)
```

**Fields & Validation**:

| Campo | Regra | Origem do erro |
|-------|-------|----------------|
| `MaxPanes` | se `-mp` informado: inteiro > 0 | parser (`flag.IntVar` com checagem pós-parse) |
| `StartPanes` + `StartOrient` | se `-sp` informado: casa com `^[hv][1-9][0-9]*$` | parser (regex + strconv) |
| `StartCommand` | se `-sc` informado: string não vazia; sem validação de executabilidade no parser (o terminal reporta falha em runtime) | parser |
| Conflito | se `MaxPanes>0` e `StartPanes>MaxPanes`: erro | validador em `StartupOverrides.Validate()` |
| Auto-raise | se `MaxPanes==0` e `StartPanes>4`: `Effective MaxPanes = StartPanes` | validador |

**State transitions**: N/A — struct imutável, construída uma vez em `main()` e consumida read-only pelo resto do programa.

---

## Modificado: `layout.Model`

**Pacote**: `github.com/menegas/lumina/components/layout`

Adicionar campos:

```go
type Model struct {
    // ... campos existentes ...
    maxPanes     int    // substitui a const de pacote; default 4 quando não especificado
    startCommand string // usado APENAS em buildInitialTree para override do shell dos panes iniciais
}
```

Adicionar options functional pattern:

```go
type Option func(*Model)

func WithMaxPanes(n int) Option      { return func(m *Model) { m.maxPanes = n } }
func WithStartCommand(cmd string) Option { return func(m *Model) { m.startCommand = cmd } }
func WithInitialLayout(orient Orient, count int) Option { ... }

func New(cfg config.Config, opts ...Option) (Model, error)
```

**Invariantes**:
- `m.maxPanes >= 1` sempre (zero é coagido para 4 no construtor).
- `m.startCommand` é lido apenas em `buildInitialTree` e `newTerminalLeafWithOverride`; `handleSplit` NUNCA lê esse campo — garante FR-010.
- `m.PaneCount()` ≤ `m.maxPanes` em qualquer momento pós-inicialização.

**Migração da const**:
- Remover `const maxPanes = 4` de `layout.go`.
- Substituir `if m.PaneCount() >= maxPanes` por `if m.PaneCount() >= m.maxPanes`.
- Atualizar a mensagem de warning em `handleSplit` para usar o valor dinâmico:
  `fmt.Sprintf("Máximo de %d painéis atingido", m.maxPanes)`.

---

## Modificado: `terminal.Model`

**Pacote**: `github.com/menegas/lumina/components/terminal`

Adicionar campo:

```go
type Model struct {
    // ... campos existentes ...
    shellOverride string // se não vazio, executa este comando no lugar de cfg.Shell
}
```

Adicionar construtor alternativo:

```go
func NewWithCommand(cfg config.Config, command string) (Model, error)
```

`NewWithCommand` é açúcar sintático sobre `New` que seta `shellOverride` antes de
chamar `startShell`. `startShell` passa a verificar `m.shellOverride`:

- se vazio → caminho atual (`buildShellCommand(m.shell, m.forceTheme, &env)`)
- se preenchido → `exec.Command("sh", "-c", m.shellOverride)` **OU**
  split com `shlex` seguido de `exec.Command(argv[0], argv[1:]...)`.

**Decision**: usar `sh -c <string>` — honra quoting/escape natural do shell do usuário
e atende FR-012 sem incluir dependência de shlex. O shell é o mesmo `sh` já exigido
por `/bin/sh` na plataforma (Constitution: Linux/macOS only).

---

## Modificado: `config.Config`

**Pacote**: `github.com/menegas/lumina/config`

Nenhuma mudança obrigatória. Campos novos de sessão (`MaxPanes`, `StartCommand`)
vivem em `cli.StartupOverrides` — não em `Config`. Motivo: `Config` é populado a
partir de `config.toml` (persistente) e as overrides são efêmeras. Manter a fronteira
evita que um `toml.DecodeFile` acidental sobrescreva decisões de CLI.

---

## Mensagens (`msgs.go`)

**Nenhuma nova `tea.Msg`** é adicionada. Toda a configuração é aplicada antes do
`tea.Program` iniciar; não há necessidade de comunicação cross-component em runtime
para esta feature. Isso também mantém a regra de testes de integração fora do escopo
(Constitution II: "Integration tests ... for every new `tea.Msg` type").

---

## Resumo das dependências

```text
main.go
   └── cli.ParseArgs(os.Args) → StartupOverrides
           └── StartupOverrides.Validate() → error ou ok
                   └── main aplica:
                          config.LoadConfig() (inalterado)
                          layout.New(cfg,
                              WithMaxPanes(overrides.EffectiveMaxPanes()),
                              WithStartCommand(overrides.StartCommand),
                              WithInitialLayout(overrides.StartOrient, overrides.StartPanes),
                          )
```
