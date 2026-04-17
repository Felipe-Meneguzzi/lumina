# Research: CLI Startup Flags

**Feature**: 004-cli-startup-flags
**Date**: 2026-04-17
**Status**: Complete — sem `NEEDS CLARIFICATION` pendentes

---

## R1 — Biblioteca de parsing de flags

**Decision**: Usar o pacote `flag` da stdlib do Go.

**Rationale**:
- Zero dependência nova (Constitution Technical Standards: "prefer standard library").
- Suporta naturalmente formas curtas (`-mp`, `-sp`, `-sc`) via `flag.IntVar` /
  `flag.StringVar`.
- `flag.Parse()` separa flags de argumentos posicionais — preserva o uso atual
  `lumina arquivo.txt`.
- Fornece `Usage()` customizável para atender FR-015 (mensagem de ajuda).

**Alternatives considered**:
- `spf13/pflag` / `urfave/cli` / `cobra`: mais expressivos (flags longas automáticas,
  subcomandos), mas injetam dependência significativa para um binário TUI que hoje não
  tem subcomandos nem precisa de completion shell.
- Parser manual em `main.go`: suficiente para o escopo mas perde `flag.Usage` grátis e
  dobra a superfície de bugs.

---

## R2 — Formato da flag `-sp`

**Decision**: Valor único string no padrão `^[hv][1-9][0-9]*$`, parseado por função
pura `ParseStartPanes(s string) (orient byte, count int, err error)`.

**Rationale**:
- É exatamente o que o usuário pediu (`-sp h3`, `-sp v2`) — reduz atrito de adoção.
- String única evita duas flags acopladas (`--orient` + `--count`), o que seria mais
  verboso sem benefício.
- Regex/parsing simples cabe em <15 linhas e gera mensagens de erro claras do tipo
  "formato inválido: esperado h<N> ou v<N>".

**Alternatives considered**:
- Duas flags separadas (`-so h -sn 3`): mais composável mas contraria a UX pedida.
- JSON/TOML inline: overkill.

---

## R3 — Onde armazenar o teto de painéis

**Decision**: Transformar `maxPanes` (hoje `const` no pacote `layout`) em campo
`maxPanes int` de `layout.Model`, inicializado via `layout.New(cfg, opts...)` com um
option functional pattern `WithMaxPanes(n int)`.

**Rationale**:
- Constitution I proíbe estado global mutável — `const` não é mutável mas também não
  permite variação por instância; transformar em campo alinha com o modelo Bubble Tea
  (estado local ao Model).
- Permite cobrir no teste `TestPaneSplitMsg_AtMaxPanes_IsNoop` variações de teto sem
  duplicar o arquivo.
- Mantém default 4 quando a opção não é passada — FR-002 atendido automaticamente.

**Alternatives considered**:
- Variável de pacote (`var maxPanes = 4`): mutação global, fere princípio I.
- Passar o teto dentro de `config.Config`: possível mas mistura config persistente
  (toml) com overrides efêmeros de CLI. Manter em `layout.Model` deixa a fronteira clara.

---

## R4 — Comando custom em painéis iniciais

**Decision**: Adicionar campo opcional `cfg.StartCommand string` que é populado só na
sessão (nunca lido de `config.toml`) e consumido por `terminal.New` quando a folha é
marcada como "inicial". Splits posteriores chamam `terminal.New` sem o override (o
campo efêmero vive em `layout.Model.startCommand` e é consumido apenas na construção
das folhas do `buildInitialTree`).

**Rationale**:
- FR-010 exige isolamento: só painéis iniciais rodam o comando. Encapsular no passo
  de bootstrap garante que `handleSplit` não herda o override por acidente.
- `terminal.Model.startShell` hoje usa `cfg.Shell`; adicionar um campo `shellOverride`
  (não `cfg.Shell`) e consumir via `buildShellCommand(override, ...)` custa ~10 linhas
  e mantém a assinatura pública compatível.
- Se `StartCommand` falhar ao iniciar, `startShell` já retorna erro — propagar esse
  erro para exibir "falha ao iniciar X" no painel (FR-011) via notificação.

**Alternatives considered**:
- Injetar o comando no shell (`sh -c "<cmd>; exec sh"`): funciona mas polui env,
  histórico de shell e complica detecção de erro de comando inexistente.
- Reutilizar `config.Config.Shell`: faria splits posteriores também rodarem o comando,
  violando FR-010.

---

## R5 — Pré-split antes do primeiro frame

**Decision**: Em `layout.New`, após criar a folha raiz, aplicar `count-1` splits
sequenciais na direção solicitada usando as funções internas já testadas
(`splitLeaf`), antes de devolver o `Model`. Isso garante que `View()` no primeiro
render já enxerga N folhas.

**Rationale**:
- Reutiliza o algoritmo existente — não duplica a lógica de tree construction.
- Satisfaz SC-002 (N painéis no primeiro frame, sem frame intermediário) porque o
  tree já está pronto quando o `tea.Program` inicia.
- Inicialização síncrona é aceitável: operações de split no tree são O(N) e N é
  pequeno (tipicamente ≤10).

**Alternatives considered**:
- Emitir N `PaneSplitMsg` como comandos iniciais (`Init()`): funcionaria mas o
  primeiro frame poderia mostrar um único painel antes dos splits chegarem, violando
  SC-002.
- Construir tree balanceado manualmente: mais eficiente mas duplica código testado
  em `splitLeaf`.

---

## R6 — Resolução de conflitos entre `-mp` e `-sp`

**Decision**: Duas regras explícitas:
1. Se `-mp` ausente e `-sp hN`/`-sp vN` com `N > 4`, elevar `maxPanes` automaticamente
   para `N` (FR-007).
2. Se `-mp M` presente e `N > M`, abortar o boot com erro claro antes da TUI abrir
   (FR-008).

**Rationale**:
- Alinhado ao que a spec já declara nos requisitos; escolha de "auto-raise" evita
  fricção no fluxo comum (usuário só lembrou do `-sp`).
- Escolha de "fail fast" quando o conflito é explícito respeita a intenção do
  usuário que digitou `-mp M` conscientemente.

**Alternatives considered**:
- Sempre abortar em conflito: mais previsível mas adiciona atrito no caso cotidiano.
- Sempre auto-raise: ignora silenciosamente a intenção explícita do usuário.

---

## R7 — Mensagem de `--help`

**Decision**: Sobrescrever `flag.Usage` para imprimir um bloco multi-linha com
exemplo real (`lumina -mp 10 -sp h3 -sc claude`), o significado de cada flag, defaults
e o uso posicional preservado (`lumina <arquivo>`).

**Rationale**: FR-015 exige ajuda; um exemplo concreto vale mais que descrição
abstrata — reforça SC-005 (usuário decorar o mnemônico).

---

## Unknowns resolvidos

Nenhum `NEEDS CLARIFICATION` restante no Technical Context. Todas as decisões de
design estão capturadas acima e nas assumptions do `spec.md`.
