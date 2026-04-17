# Phase 1 Data Model — UX Polish Pack

**Feature**: 006-ux-polish-pack
**Date**: 2026-04-17

Esta feature é puramente de UX — não introduz persistência nova além de um campo de configuração. "Data model" aqui significa as structs em memória que compõem o estado dos componentes e os msgs que trafegam entre eles.

## Entidades

### 1. Config (alterada)

`config.Config` em `config/config.go` — ganha um campo.

| Campo | Tipo | Default | Validação | Observações |
|---|---|---|---|---|
| `Editor` | `string` | `"nano"` | Valor aceito: `"nano"`, `"vim"`, `"nvim"` ou path absoluto resolvível por `exec.LookPath`. Se inválido, emite warning e usa `"nano"`. | Novo. Persistido em `config.toml` sob chave `editor`. |

Demais campos existentes permanecem inalterados.

### 2. SidebarCreatePrompt (nova, interna ao sidebar)

Submodel efêmero que existe enquanto o usuário digita o nome de um arquivo/pasta a criar. Vida curta: nasce ao pressionar alt+d/alt+f, morre ao confirmar (Enter) ou cancelar (ESC).

| Campo | Tipo | Propósito |
|---|---|---|
| `Kind` | `string` ("dir" \| "file") | Tipo do elemento a criar. |
| `Input` | `textinput.Model` | Captura de teclas com cursor, edição, etc. (bubbles). |
| `Err` | `string` | Mensagem de validação atual (ex.: "nome vazio", "já existe"). Exibida inline. |
| `ParentDir` | `string` | Diretório onde a criação ocorrerá (o diretório atual da sidebar no momento da ativação). |

Ciclo de vida:
1. `Active` — usuário digitando.
2. `Validating` (transiente, dentro de `Update` ao receber Enter) — checa nome.
3. `Creating` (síncrono, dentro de `Update`) — executa `os.Mkdir` ou `os.WriteFile`.
4. `Done` — submodel liberado; Enter confirma fluxo: se dir, sidebar navega para dentro; se file, app emite `OpenInExternalEditorMsg`.
5. `Cancelled` (ESC) — submodel liberado; sidebar volta ao estado anterior.

Invariantes:
- Enquanto `createPrompt != nil`, Backspace e setas atuam no `textinput`, **não** na navegação de diretórios.
- Apenas uma instância ativa por vez por sidebar.

### 3. TerminalPane (alterada — campos internos)

`terminal.Model` em `components/terminal/terminal.go` ganha/altera alguns campos (sem mudar a API externa).

| Campo | Tipo | Propósito |
|---|---|---|
| `cwd` | `string` | CWD reportado via OSC 7. Alimenta `PaneCWDChangeMsg`. Novo. |
| `gitBranch` | `string` | Última branch resolvida para este pane. Novo. |
| `gitDirty` | `bool` | Estado dirty/clean da última resolução. Novo. |
| `firstRenderDone` | `bool` | Flag usada para o fix de R1: falso até aplicarmos `pty.Setsize` e ler o primeiro chunk. Novo. |

### 4. Tipos tea.Msg (msgs/msgs.go — resumo)

Novos (detalhados em `contracts/msgs.md`):

- `ClickFocusMsg`
- `SidebarCreateMsg`
- `SidebarCreatedMsg`
- `OpenInExternalEditorMsg`
- `ClockTickMsg`
- `PaneCWDChangeMsg`
- `PaneGitStateMsg`
- `FocusedPaneContextMsg`

Removidos:

- `EditorResizeMsg`
- `ConfirmCloseMsg`
- `CloseConfirmedMsg`
- `CloseAbortedMsg`

Alterados:

- `FocusTarget`: remove o valor `FocusEditor`.

## Relacionamentos

```text
MouseMsg (tea)
    └── app.Update() — hit-test via layout.Tree
           └── ClickFocusMsg ──► layout.focus → PaneFocusMsg
                              └► (pass-through) PtyMouseMsg | MouseSelectMsg | SidebarMouseSelectMsg

Sidebar [alt+d|alt+f]
    └── createPrompt (submodel)
           ├── SidebarCreateMsg (on Enter) ──► sidebar cria o FS entry
           └── ESC ──► drop prompt

Sidebar [Enter on file]
    └── OpenInExternalEditorMsg ──► app spawna terminal pane com cfg.Editor

Terminal PTY output
    └── OSC 7 parser ──► PaneCWDChangeMsg ──► terminal roda `git` em tea.Cmd ──► PaneGitStateMsg

FocusChangeMsg | PaneGitStateMsg | PaneCWDChangeMsg
    └── layout consolida ──► FocusedPaneContextMsg ──► statusbar atualiza exibição

ClockTickMsg (tea.Tick 30s)
    └── statusbar atualiza m.now
```

## Estados observáveis (para testes)

- **Painel focado**: exatamente um, em todo momento (invariante testável via introspecção do `layout.Tree`).
- **Cursor visível**: exatamente no painel focado (terminal); sidebar e status bar não têm cursor de texto próprio.
- **Status bar context**: branch e CWD exibidos correspondem ao último `FocusedPaneContextMsg` recebido; se o último foi de um painel sem git, o glifo e a branch desaparecem.
- **CreatePrompt**: presença/ausência é booleana; quando presente, captura 100% das teclas exceto foco global.
