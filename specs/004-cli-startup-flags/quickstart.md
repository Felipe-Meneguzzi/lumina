# Quickstart: CLI Startup Flags

**Feature**: 004-cli-startup-flags
**Audience**: Desenvolvedor que vai implementar ou revisar a feature.

Este quickstart mostra o fluxo ponta-a-ponta em ~5 minutos: do build ao primeiro
uso, incluindo validação manual dos principais casos.

---

## 1. Build local

```bash
cd ~/fpm/lumina
go build -o lumina .
```

Nenhuma dependência nova precisa ser baixada — apenas `flag` da stdlib.

---

## 2. Casos de boot

### 2.1 Default (regressão zero)
```bash
./lumina
```
Esperado: 1 painel, shell default, teto 4 (comportamento idêntico ao da versão anterior).

### 2.2 Apenas teto custom
```bash
./lumina -mp 10
```
Esperado: 1 painel inicial. Splits via atalhos param no 10º painel.

### 2.3 Layout pré-splitado horizontal
```bash
./lumina -sp h3
```
Esperado: 3 painéis lado-a-lado no primeiro frame; foco em um deles; teto ajusta para `max(4, 3) = 4`.

### 2.4 Layout pré-splitado vertical com comando custom
```bash
./lumina -sp v2 -sc claude
```
Esperado: 2 painéis empilhados; em cada um, `claude` rodando. Ao fazer split manual
(atalho de keymap), o novo painel abre shell default (não `claude`).

### 2.5 Combinação completa (exemplo da spec)
```bash
./lumina -mp 10 -sp h3 -sc claude
```
Esperado: 3 painéis horizontais com `claude` rodando, teto 10.

### 2.6 Com arquivo posicional
```bash
./lumina README.md -sp h2
```
Esperado: 2 painéis horizontais; `README.md` abre no painel focado (converte o terminal focado em editor).

---

## 3. Casos de erro (devem falhar antes da TUI)

```bash
./lumina -mp 0
# stderr: lumina: -mp inválido: esperado inteiro >= 1, recebi "0"
# exit code: 2

./lumina -sp h
# stderr: lumina: -sp inválido: esperado h<N> ou v<N> com N >= 1, recebi "h"
# exit code: 2

./lumina -mp 2 -sp h5
# stderr: lumina: -sp h5 excede -mp 2: não é possível criar 5 painéis iniciais com teto 2
# exit code: 2

./lumina -sc ""
# stderr: lumina: -sc inválido: comando não pode ser vazio
# exit code: 2
```

Em todos os casos, **a TUI não abre** — erros vão direto para stderr.

---

## 4. Caso de comando inexistente (falha em runtime)

```bash
./lumina -sp h2 -sc nao-existe-esse-comando-aqui
```
Esperado: a TUI abre com 2 painéis; cada painel mostra mensagem do `sh` ("command not
found") e/ou o notify da status bar. A sessão permanece utilizável — o usuário pode
fechar painéis ou continuar trabalhando normalmente.

---

## 5. `--help`

```bash
./lumina --help
```
Esperado: bloco de ajuda listando as flags novas com exemplos (ver contract em
`contracts/cli.md` seção "Mensagem de --help esperada").

---

## 6. Validação automatizada

```bash
# unit tests de parsing
go test ./cli/...

# unit tests do layout com maxPanes custom
go test ./components/layout/...

# unit tests do terminal com shellOverride
go test ./components/terminal/...

# suite completa
go test ./...
```

Todos devem passar antes do merge (Constitution II).

---

## 7. Atualização do README

Após implementar, acrescentar uma seção "CLI flags" no `README.md` do projeto com os
mesmos exemplos canônicos da tabela do contract. O usuário pediu explicitamente:
**"Lembre-se de atualizar o README"**.

---

## Referências cruzadas

- Contrato completo da CLI: [contracts/cli.md](./contracts/cli.md)
- Decisões técnicas: [research.md](./research.md)
- Estruturas internas: [data-model.md](./data-model.md)
- Requisitos testáveis: [spec.md](./spec.md)
