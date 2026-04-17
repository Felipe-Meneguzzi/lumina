# Feature Specification: Mouse Text Selection in Normal Mode

**Feature Branch**: `005-mouse-select-copy`
**Created**: 2026-04-17
**Status**: Draft
**Input**: User description: "Quero poder usar o mouse pra selecionar e copiar o texto no modo normal também, deixa o modo copy para quando a pessoa não quer usar o mouse em nenhum momento"

## Clarifications

### Session 2026-04-17

- Q: Como acionar a cópia após seleção com mouse — automático ao soltar ou tecla explícita? → A: Configurável via flag `mouse_auto_copy` (padrão: ativado = cópia automática ao soltar o mouse); quando desativado, a seleção persiste e o usuário pressiona `y` para confirmar.
- Q: Quando `mouse_auto_copy` está desativado e há seleção pendente, `y` vai ao PTY, ao Lumina, ou a ambos? → A: `y` é consumido pelo Lumina (copia, não encaminhado ao PTY); sem seleção pendente, comportamento normal.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Select and Copy Text with Mouse in Normal Mode (Priority: P1)

O usuário está trabalhando no painel de terminal com uma aplicação que não usa rastreamento de mouse (ex.: shell simples, saída de comandos). Ele quer copiar um trecho de texto sem precisar ativar o copy mode.

**Why this priority**: É o core da feature. Elimina a necessidade de entrar no copy mode para quem usa mouse no fluxo normal de trabalho.

**Independent Test**: Pode ser testado executando o Lumina, focando o painel terminal, clicando e arrastando sobre o texto, acionando a cópia e verificando o conteúdo da área de transferência.

**Acceptance Scenarios**:

1. **Given** o painel terminal está focado em modo normal, **When** o usuário clica e arrasta sobre um trecho de texto, **Then** o texto selecionado é visivelmente destacado (realce visual).
2. **Given** há texto selecionado via arrasto do mouse e `mouse_auto_copy` está ATIVADO, **When** o usuário solta o botão do mouse, **Then** o texto selecionado é enviado à área de transferência e uma notificação de confirmação é exibida.
3. **Given** há texto selecionado via arrasto do mouse e `mouse_auto_copy` está DESATIVADO, **When** o usuário pressiona `y`, **Then** o texto selecionado é enviado à área de transferência e uma notificação de confirmação é exibida.
4. **Given** há texto selecionado com `mouse_auto_copy` desativado, **When** o usuário clica em outra posição ou pressiona `Esc`, **Then** a seleção é removida sem alterar a área de transferência.

---

### User Story 2 - Seleção com Mouse Quando a Aplicação Interna Usa Rastreamento de Mouse (Priority: P2)

O usuário está rodando vim, tmux ou outra aplicação que habilitou o rastreamento de mouse dentro do painel terminal. Ele quer selecionar conteúdo do terminal no Lumina (não interagir com o mouse da aplicação interna) sem entrar no copy mode.

**Why this priority**: Importante para quem usa TUIs com suporte a mouse e ainda quer poder selecionar texto do Lumina com o mouse sem mudar de modo.

**Independent Test**: Pode ser testado rodando vim (com `mouse=a`) no painel terminal, usando uma tecla modificadora + arrasto para selecionar texto e copiando.

**Acceptance Scenarios**:

1. **Given** a aplicação interna tem rastreamento de mouse ativo, **When** o usuário segura Shift e clica/arrasta, **Then** o Lumina intercepta o evento para seleção de texto em vez de encaminhá-lo à aplicação interna.
2. **Given** a aplicação interna tem rastreamento de mouse ativo e o usuário fez uma seleção com Shift+arrasto, **When** o usuário aciona a cópia, **Then** o texto selecionado é copiado para a área de transferência.
3. **Given** a aplicação interna tem rastreamento de mouse ativo, **When** o usuário faz um clique simples (sem modificador), **Then** o evento é encaminhado normalmente à aplicação interna.

---

### User Story 3 - Usuários Somente-Teclado Mantêm o Copy Mode (Priority: P3)

O usuário que não quer usar mouse em momento algum ainda pode invocar o copy mode existente para seleção e cópia de texto via teclado.

**Why this priority**: Preserva o fluxo atual para usuários que preferem teclado (estilo vi/tmux) e garante zero regressão no copy mode.

**Independent Test**: Pode ser testado invocando o copy mode pelo atalho de teclado, navegando e selecionando texto com teclado e confirmando a cópia.

**Acceptance Scenarios**:

1. **Given** o painel terminal está focado, **When** o usuário invoca o copy mode pelo atalho de teclado, **Then** o copy mode ativa normalmente com navegação e seleção via teclado.
2. **Given** o usuário está no copy mode, **When** ele copia o texto selecionado, **Then** o texto vai para a área de transferência como antes.

---

### Edge Cases

- O que acontece se o usuário iniciar um arrasto mas soltar o mouse fora dos limites do painel terminal?
- Como a seleção é limpa quando o usuário muda o foco para outro painel?
- O que acontece se o painel terminal for redimensionado enquanto há uma seleção ativa?
- Quando a aplicação interna alterna entre modos de rastreamento de mouse, a seleção é descartada?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Em modo normal, o painel terminal DEVE permitir clicar e arrastar para destacar visualmente uma região de texto.
- **FR-002**: O texto selecionado DEVE ser visivelmente distinguível do restante (ex.: highlight por inversão de vídeo), consistente com o visual atual do copy mode.
- **FR-003**: O comportamento de cópia ao finalizar uma seleção com mouse DEVE ser controlado pela opção de configuração `mouse_auto_copy` (padrão: ativado).
- **FR-003a**: Quando `mouse_auto_copy` estiver ATIVADO, o texto selecionado DEVE ser copiado automaticamente para a área de transferência ao soltar o botão do mouse.
- **FR-003b**: Quando `mouse_auto_copy` estiver DESATIVADO, a seleção DEVE permanecer visualmente destacada após o arrasto e o usuário DEVE pressionar `y` para confirmar a cópia; `Esc` ou clique fora DEVE cancelar a seleção sem alterar a área de transferência.
- **FR-004**: Um clique simples (sem arrasto) no terminal DEVE limpar qualquer seleção ativa.
- **FR-005**: Quando a aplicação interna tiver rastreamento de mouse ativo, segurar Shift durante o clique+arrasto DEVE ativar a seleção do Lumina em vez de encaminhar o evento à aplicação interna.
- **FR-006**: O copy mode (seleção via teclado) DEVE permanecer completamente funcional e sem alterações como alternativa para usuários somente-teclado.
- **FR-007**: Uma notificação de status DEVE confirmar a cópia bem-sucedida para a área de transferência, consistente com o comportamento atual do copy mode.
- **FR-008**: A seleção via mouse DEVE ser descartada quando o foco sair do painel terminal.
- **FR-009**: Enquanto houver uma seleção de mouse ativa com `mouse_auto_copy` ATIVADO (cópia já ocorreu ao soltar o mouse), teclas de entrada DEVEM ser encaminhadas normalmente à aplicação interna.
- **FR-009b**: Quando `mouse_auto_copy` está DESATIVADO e há uma seleção pendente aguardando `y`, a tecla `y` DEVE ser consumida pelo Lumina (aciona a cópia e não é encaminhada ao PTY). Todas as demais teclas DEVEM ser encaminhadas normalmente. Quando não há seleção pendente, `y` também é encaminhado normalmente.
- **FR-010**: A opção `mouse_auto_copy` DEVE ser declarada no arquivo de configuração do usuário e carregada na inicialização do editor.

### Key Entities *(include if feature involves data)*

- **Seleção de Mouse (MouseSelection)**: Região definida por ponto de início e ponto de fim no espaço do viewport do terminal; pode estar vazia (sem seleção ativa).
- **Conteúdo do Terminal**: Texto renderizado no viewport do painel terminal, incluindo possível histórico de scrollback visível.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Usuários conseguem selecionar e copiar texto inteiramente com o mouse em menos de 5 segundos, sem precisar entrar no copy mode.
- **SC-002**: A seleção visual é precisa: a região destacada corresponde exatamente ao texto sobre o qual o usuário arrastou o mouse.
- **SC-003**: O copy mode permanece completamente funcional sem regressões (todos os cenários de seleção via teclado continuam funcionando).
- **SC-004**: Quando a aplicação interna usa rastreamento de mouse, segurar Shift intercepta o evento 100% das vezes, sem escapes não intencionais para a aplicação interna.
- **SC-005**: O conteúdo da área de transferência só muda quando o usuário completa intencionalmente uma seleção (soltando o mouse com `mouse_auto_copy` ativado, ou pressionando `y` com `mouse_auto_copy` desativado) — nunca durante o arrasto nem em cliques simples.

## Assumptions

- O suporte a mouse já está habilitado no nível da aplicação Bubble Tea (evidenciado pelos eventos de mouse já tratados no `handleMouse`).
- A área de transferência do sistema é acessível via OSC 52, consistente com a implementação atual do copy mode.
- A seleção via mouse é restrita ao painel terminal; painéis de editor e sidebar estão fora do escopo desta feature.
- A tecla modificadora padrão para ignorar o passthrough de mouse da aplicação interna é Shift, seguindo a convenção de emuladores de terminal (tmux, Alacritty, iTerm2).
- Ambientes sem suporte a mouse (terminais remotos sem repasse de mouse, sessões SSH simples) estão fora do escopo.
