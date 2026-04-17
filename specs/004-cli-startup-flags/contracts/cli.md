# Contract: Lumina CLI Flags

**Feature**: 004-cli-startup-flags
**Surface**: Command-line interface exposta pelo binário `lumina`
**Audience**: Usuários finais e scripts que automatizam o boot do editor.

Este é o único contrato externo relevante para esta feature. A TUI não expõe
endpoints nem IPC; o "API" a ser respeitada é a linha de comando.

---

## Sinopse

```text
lumina [-mp <N>] [-sp <orient><N>] [-sc <command>] [<file>]
lumina --help | -h
lumina --version | -v | version
```

---

## Flags

### `-mp <N>` — Max Panes

- **Tipo**: inteiro ≥ 1
- **Default**: 4
- **Efeito**: define o número máximo de painéis permitidos durante toda a sessão.
  Tentativas de split que ultrapassem esse valor são rejeitadas com notificação na
  status bar (mensagem existente, atualizada para usar o valor dinâmico).
- **Erros**:
  - valor não numérico → `lumina: -mp inválido: esperado inteiro >= 1, recebi "<val>"` (exit 2)
  - valor `<= 0` → mesma mensagem acima (exit 2)

### `-sp <orient><N>` — Start Panes

- **Tipo**: string no formato `^[hv][1-9][0-9]*$`
  - `h` = horizontal (lado a lado)
  - `v` = vertical (empilhado)
  - `N` = inteiro ≥ 1
- **Default**: ausente → 1 painel (boot atual)
- **Efeito**: na inicialização, cria `N` painéis dispostos na orientação solicitada.
  Tamanhos proporcionais seguindo o mesmo algoritmo de `splitLeaf` usado em runtime.
- **Erros**:
  - formato inválido (`-sp 3`, `-sp h`, `-sp d2`, `-sp h0`) → `lumina: -sp inválido: esperado h<N> ou v<N> com N >= 1, recebi "<val>"` (exit 2)
  - `N > -mp` (quando `-mp` explícito) → `lumina: -sp h<N> excede -mp <M>: não é possível criar <N> painéis iniciais com teto <M>` (exit 2)

### `-sc <command>` — Start Command

- **Tipo**: string não vazia (pode conter espaços e argumentos, passada como uma unidade via `sh -c`)
- **Default**: ausente → shell default resolvido por `config.validateShell`
- **Efeito**: substitui o shell default APENAS nos painéis criados pela flag `-sp`
  (ou no único painel default, se `-sp` estiver ausente). Painéis criados depois via
  split manual abrem o shell default do `config.Config`.
- **Erros**:
  - string vazia (`-sc ""`) → `lumina: -sc inválido: comando não pode ser vazio` (exit 2)
  - comando inexistente → detectado em runtime; o painel mostra mensagem de erro via
    notify status e a sessão continua (FR-011). Não gera exit non-zero no boot.

### Argumento posicional `<file>`

- **Tipo**: path para arquivo a abrir no editor
- **Default**: ausente → boot com terminal no painel raiz (ou painéis iniciais de `-sp`)
- **Efeito**: inalterado em relação ao comportamento atual — emite `msgs.OpenFileMsg{Path: path}`.
- **Interação com `-sp`/`-sc`**: se ambos `<file>` e `-sp` estão presentes, o arquivo
  abre em UM dos painéis iniciais (o focado, conforme regra atual de `OpenFileMsg`).
  Se `-sc` está presente, o `OpenFileMsg` ainda é respeitado: o painel alvo troca
  para o editor e o start command é descartado nele (o comando só vale para painéis
  que permanecem como terminal).

### Flags existentes preservadas

- `--help`, `-h`: imprime o `flag.Usage` customizado com exemplos.
- `--version`, `-v`, `version`: inalterado.

---

## Ordem de processamento

1. Parse de flags com `flag.Parse()`.
2. `StartupOverrides.Validate()`:
   - valida ranges isolados (`MaxPanes`, `StartPanes`, `StartOrient`)
   - calcula `EffectiveMaxPanes()` (auto-raise quando `-mp` ausente e `-sp > 4`)
   - valida conflito explícito
3. Se qualquer erro → `fmt.Fprintln(os.Stderr, ...)` + `os.Exit(2)` antes de `config.LoadConfig`.
4. Se ok → `config.LoadConfig` → `app.New(cfg, ...)` → `tea.NewProgram(...)`.

---

## Exemplos canônicos

| Comando | Resultado |
|---------|-----------|
| `lumina` | 1 painel, shell default, teto 4. |
| `lumina -mp 10` | 1 painel, shell default, teto 10. |
| `lumina -sp h3` | 3 painéis horizontais, shell default em cada, teto `max(4, 3) = 4`. |
| `lumina -sp v2 -sc claude` | 2 painéis verticais, `claude` rodando em cada, teto 4. |
| `lumina -mp 10 -sp h3 -sc claude` | 3 painéis horizontais rodando `claude`, teto 10. |
| `lumina arquivo.txt -sp h2` | 2 painéis horizontais; `arquivo.txt` abre no foco inicial. |
| `lumina -mp 2 -sp h5` | erro: `-sp h5 excede -mp 2` (exit 2). |
| `lumina -sp d3` | erro: `-sp inválido: esperado h<N> ou v<N>` (exit 2). |

---

## Mensagem de `--help` esperada

```text
Lumina — TUI editor with splittable panes.

Usage:
  lumina [flags] [file]

Flags:
  -mp N             Max panes allowed in this session (default: 4)
  -sp <h|v>N        Pre-split layout on startup (e.g. h3 = 3 horizontal panes)
  -sc <command>     Run <command> in initial panes instead of the default shell
                    (applies only to panes created by -sp; later splits use the shell)
  --version, -v     Print version and exit
  --help, -h        Show this help

Examples:
  lumina
  lumina -mp 10 -sp h3 -sc claude
  lumina notes.md -sp v2
```
