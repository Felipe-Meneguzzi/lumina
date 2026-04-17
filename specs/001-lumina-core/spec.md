# Feature Specification: Lumina — TUI Editor Core

**Feature Branch**: `001-lumina-core`
**Created**: 2026-04-16
**Status**: Draft
**Input**: User description: "Lumina é um editor de terminal estilo VSCode — uma TUI (Terminal User Interface)
que combina painéis de terminal interativo, explorador de arquivos, status bar com métricas do sistema e
editor de texto simples. Projetado para rodar inteiramente no terminal, com foco em produtividade e leveza."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Terminal Interativo (Priority: P1)

O usuário abre o Lumina e vê um painel de terminal funcional onde pode executar comandos de shell,
programas interativos (vim, htop, less) e ver a saída em tempo real. O painel responde corretamente
a resize da janela do terminal externo.

**Why this priority**: É a funcionalidade central do produto — sem um terminal funcional, as demais
funcionalidades (explorador, editor) perdem contexto de uso. Entrega valor imediato sozinha.

**Independent Test**: Abrir o Lumina, executar `ls -la` no painel de terminal e verificar a saída
formatada. Executar `htop` e verificar interatividade. Redimensionar a janela e confirmar que o
conteúdo se adapta sem corrupção.

**Acceptance Scenarios**:

1. **Given** o Lumina está aberto, **When** o usuário digita `echo hello` e pressiona Enter,
   **Then** o terminal exibe `hello` na linha seguinte
2. **Given** um programa interativo está rodando (ex: `top`), **When** o usuário pressiona `q`,
   **Then** o programa encerra e o prompt do shell é restaurado
3. **Given** o terminal tem o foco, **When** a janela externa é redimensionada,
   **Then** o conteúdo do PTY se adapta ao novo tamanho sem caracteres corrompidos

---

### User Story 2 — Explorador de Arquivos (Priority: P2)

O usuário navega pelo sistema de arquivos em um painel lateral (sidebar), expande/colapsa diretórios
e abre arquivos no painel de editor ou copia caminhos para o terminal.

**Why this priority**: Complementa o terminal fornecendo navegação visual — elimina a necessidade
de `ls`/`cd` constantes. Entrega valor como segundo MVP incremento.

**Independent Test**: Abrir o Lumina, alternar o foco para a sidebar, navegar até um diretório
aninhado usando as teclas de seta, e verificar que o caminho exibido é correto.

**Acceptance Scenarios**:

1. **Given** a sidebar está focada, **When** o usuário pressiona a seta para baixo,
   **Then** o item selecionado avança para o próximo arquivo/diretório na lista
2. **Given** um diretório está selecionado, **When** o usuário pressiona Enter ou a seta para direita,
   **Then** o diretório expande e exibe seus filhos
3. **Given** um arquivo está selecionado, **When** o usuário pressiona Enter,
   **Then** o arquivo é aberto no painel de editor (US3) ou, se o editor não estiver disponível,
   o caminho é copiado para o clipboard

---

### User Story 3 — Editor de Texto Simples (Priority: P3)

O usuário edita arquivos de texto diretamente no Lumina — abre um arquivo, modifica o conteúdo,
salva e fecha. Suporte a navegação básica (cursor, scroll), sem syntax highlighting em v1.

**Why this priority**: Completa o loop de edição sem precisar sair do Lumina. Depende da sidebar
(US2) para abertura de arquivos, mas é testável de forma independente via argumento de linha de
comando.

**Independent Test**: Abrir o Lumina com `lumina arquivo.txt`, editar uma linha, pressionar o
atalho de salvar, fechar e verificar que o arquivo no disco contém as alterações.

**Acceptance Scenarios**:

1. **Given** um arquivo está aberto no editor, **When** o usuário digita texto,
   **Then** o conteúdo é inserido na posição do cursor
2. **Given** o usuário fez alterações, **When** pressiona o atalho de salvar,
   **Then** o arquivo é gravado em disco e um indicador de "modificado" desaparece
3. **Given** o arquivo tem alterações não salvas, **When** o usuário tenta fechar o painel,
   **Then** o Lumina pergunta se deseja salvar antes de fechar

---

### User Story 4 — Status Bar com Métricas do Sistema (Priority: P2)

O status bar exibe continuamente métricas do sistema operacional (CPU, memória, disco) e informações
de contexto (diretório atual, branch git se aplicável), atualizando em tempo real sem impactar a
responsividade dos demais painéis.

**Why this priority**: Junto com o terminal (P1), a status bar diferencia o Lumina de um terminal
comum. Implementada em paralelo com US2.

**Independent Test**: Abrir o Lumina e observar a status bar por 5 segundos. Verificar que as
métricas de CPU e memória mudam ao rodar um processo pesado no terminal.

**Acceptance Scenarios**:

1. **Given** o Lumina está aberto, **When** nenhuma ação é realizada,
   **Then** a status bar atualiza as métricas de CPU e memória pelo menos 1 vez por segundo
2. **Given** um processo pesado é executado no terminal, **When** o CPU ultrapassa 50%,
   **Then** o valor de CPU na status bar reflete a mudança na próxima atualização
3. **Given** o usuário está em um diretório git, **When** a status bar é renderizada,
   **Then** o nome da branch atual é exibido na status bar

---

### Edge Cases

- **Shell exit**: Quando o processo shell encerra, o Lumina reinicia automaticamente uma nova sessão PTY com o mesmo shell (`$SHELL`), sem intervenção do usuário
- Como o Lumina se comporta em terminais que não suportam cores ou tamanho reduzido (<80 colunas)?
- O que acontece quando o arquivo aberto no editor é modificado externamente enquanto está aberto?
- Como o foco de teclado é transferido entre painéis com atalhos conflitantes com programas no terminal?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema DEVE exibir um painel de terminal interativo com suporte a PTY real,
  capaz de executar qualquer programa interativo de terminal
- **FR-002**: O sistema DEVE propagar eventos de resize da janela ao PTY para que programas
  como `vim` e `htop` reajustem seu layout corretamente
- **FR-003**: O usuário DEVE poder navegar entre painéis (terminal, sidebar, editor) usando
  atalhos de teclado globais definidos centralmente
- **FR-004**: O sistema DEVE exibir uma sidebar de explorador de arquivos com navegação
  hierárquica por teclado (expandir/colapsar diretórios, selecionar arquivos)
- **FR-005**: O usuário DEVE poder abrir arquivos de texto para edição com suporte a
  inserção, deleção e navegação básica de cursor
- **FR-006**: O sistema DEVE persistir alterações no editor em disco ao acionar o comando de salvar
- **FR-007**: O sistema DEVE exibir métricas de CPU, memória e contexto (diretório, branch git)
  em uma status bar que atualiza com intervalo mínimo de 1 segundo
- **FR-008**: O sistema DEVE alertar o usuário antes de descartar alterações não salvas no editor
- **FR-009**: Todos os atalhos de teclado DEVEM ser definidos em um único arquivo de configuração
  central e não devem conflitar entre si
- **FR-010**: Quando o processo shell do painel de terminal encerrar, o sistema DEVE reiniciar
  automaticamente uma nova sessão PTY com o mesmo shell, sem intervenção do usuário

### Key Entities

- **Painel (Pane)**: Área visual delimitada que hospeda um componente (terminal, sidebar, editor).
  Possui estado de foco, dimensões e pode ser redimensionado
- **Sessão de Terminal**: Instância de PTY associada a um processo shell. Mantém histórico de
  output e estado de interatividade
- **Arquivo**: Representação de um arquivo do sistema de arquivos no editor — possui caminho,
  conteúdo em buffer, estado de modificação e cursor
- **Métricas**: Snapshot periódico de CPU %, memória usada/total, diretório de trabalho atual
  e branch git (quando disponível)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Um usuário experiente em terminal consegue abrir o Lumina, executar um comando no
  terminal integrado e visualizar a saída em menos de 5 segundos após a inicialização
- **SC-002**: A transição de foco entre painéis responde em menos de 100ms após o atalho de teclado
- **SC-003**: A status bar atualiza métricas a cada 1–2 segundos sem causar queda perceptível
  na responsividade do painel de terminal (sem frame drops visíveis)
- **SC-004**: O editor suporta arquivos de texto de até 10.000 linhas com scroll fluido sem
  degradação perceptível da interface
- **SC-005**: 100% dos atalhos de teclado são listados em uma única fonte de verdade, sem
  atalhos duplicados ou conflitantes entre componentes
- **SC-006**: O Lumina inicializa (exibe a TUI completa) em menos de 500ms em hardware moderno

## Assumptions

- O público-alvo são desenvolvedores e usuários avançados de terminal (Linux/macOS) — sem
  suporte a Windows em v1 (PTY não disponível nativamente)
- A implementação usa Go 1.26 (versão mais recente disponível)
- Sintaxe highlighting e suporte a múltiplas abas de editor são funcionalidades futuras (fora
  de escopo desta spec)
- O shell padrão do usuário (definido em `$SHELL`) é usado como processo inicial do painel de terminal
- A configuração de atalhos será lida de um arquivo em `~/.config/lumina/config.toml` se existir,
  com fallback para defaults embutidos
- Internacionalização e suporte a idiomas além do inglês/UTF-8 estão fora de escopo
- A aplicação não gerencia múltiplas janelas — uma instância = uma janela TUI com layout fixo
  (terminal + sidebar + editor + statusbar)

## Clarifications

### Session 2026-04-16

- Q: Qual versão do Go usar? → A: Go 1.26 (versão mais recente)
- Q: Comportamento quando o shell encerra? → A: Reiniciar automaticamente nova sessão PTY com o mesmo `$SHELL`
