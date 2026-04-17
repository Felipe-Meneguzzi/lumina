# Feature Specification: UX Fixes — Multi-Window Layout

**Feature Branch**: `003-ux-fixes-multiwindow`  
**Created**: 2026-04-16  
**Status**: Draft  
**Input**: User description: "Alguns detalhes, a sidebar precisa ser por janela e precisa ser redimensionada ou escondida, com um atalho por exemplo quero poder fechar e abrir no terminal que esta em foco, outra coisa, o terminal que abre deve ser uma sessão do lumina tambem, pois esta abrindo um powershell do windows apenas, o resource monitor preciso poder esconder tambem e deve ser unico para todos os terminais, quando abro uma janela nova, ele parece que se perde, pois os atalhos nao funcionam e se eu tento fechar a janela inicial ele diz q é a unica janela aberta"

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Sidebar por Janela com Toggle (Priority: P1)

O usuário trabalha com múltiplas janelas abertas no Lumina. Cada janela possui sua própria sidebar (explorador de arquivos). O usuário pode ocultar ou exibir a sidebar da janela atualmente em foco usando um atalho de teclado, sem afetar as demais janelas.

**Why this priority**: A sidebar compartilhada entre janelas cria confusão de contexto. Cada janela representa um espaço de trabalho independente e precisa de sua própria navegação de arquivos.

**Independent Test**: Abrir duas janelas, ocultar a sidebar da janela 1 com o atalho — a sidebar da janela 2 deve permanecer visível. A sidebar deve reaparecer ao pressionar o atalho novamente.

**Acceptance Scenarios**:

1. **Given** duas janelas abertas com sidebars visíveis, **When** o usuário pressiona o atalho de toggle com a janela 1 em foco, **Then** apenas a sidebar da janela 1 é ocultada; a janela 2 permanece inalterada.
2. **Given** a sidebar da janela em foco está oculta, **When** o usuário pressiona o atalho de toggle, **Then** a sidebar reaparece no mesmo estado (posição e tamanho) em que estava antes de ser ocultada.
3. **Given** uma janela em foco com sidebar visível, **When** o usuário pressiona o atalho de toggle, **Then** o painel do terminal (ou editor) expande para ocupar o espaço antes ocupado pela sidebar.

---

### User Story 2 — Terminal Abre com Shell do Lumina (Priority: P1)

Ao abrir um novo terminal dentro do Lumina, o shell iniciado é o shell padrão do ambiente do usuário (ex.: bash, zsh), e não um shell externo como PowerShell. O ambiente de terminal deve se comportar como uma sessão nativa do sistema onde o Lumina está rodando.

**Why this priority**: Abrir PowerShell em vez do shell esperado pelo usuário é um bug crítico que impede o uso básico do terminal no Linux/macOS.

**Independent Test**: Abrir um novo terminal no Lumina e verificar qual shell/processo está ativo. O prompt deve corresponder ao shell padrão configurado no sistema (ex.: `$SHELL`).

**Acceptance Scenarios**:

1. **Given** o Lumina rodando em Linux/macOS, **When** o usuário abre um novo painel de terminal, **Then** o processo iniciado é o shell padrão do sistema (`$SHELL` ou equivalente).
2. **Given** um terminal aberto no Lumina, **When** o usuário digita comandos, **Then** os comandos são executados no ambiente nativo do sistema operacional onde o Lumina está rodando.
3. **Given** o Lumina configurado com shell personalizado, **When** um novo terminal é aberto, **Then** o shell utilizado respeita a configuração do usuário (ex.: `config.toml`).

---

### User Story 3 — Resource Monitor Global e Ocultável (Priority: P2)

O monitor de recursos (CPU, memória, etc.) é exibido uma única vez para todos os terminais abertos, sem ser duplicado por janela. O usuário pode ocultar e exibir o monitor de recursos usando um atalho de teclado.

**Why this priority**: O monitor de recursos é informação global do sistema — duplicá-lo por janela desperdiça espaço e confunde o usuário. A capacidade de ocultá-lo é necessária para maximizar a área útil.

**Independent Test**: Abrir múltiplas janelas e verificar que o monitor de recursos aparece uma única vez na interface. Pressionar o atalho de toggle deve ocultar/exibir o monitor para toda a aplicação.

**Acceptance Scenarios**:

1. **Given** três janelas abertas no Lumina, **When** o usuário observa a interface, **Then** existe apenas um monitor de recursos visível, compartilhado entre todas as janelas.
2. **Given** o monitor de recursos visível, **When** o usuário pressiona o atalho de toggle, **Then** o monitor de recursos é ocultado e o espaço ocupado é redistribuído para os outros painéis.
3. **Given** o monitor de recursos oculto, **When** o usuário pressiona o atalho de toggle, **Then** o monitor de recursos volta a ser exibido com as métricas atualizadas.
4. **Given** uma nova janela aberta enquanto o monitor está oculto, **When** a janela é exibida, **Then** o monitor permanece oculto (o estado é global, não por janela).

---

### User Story 4 — Foco e Atalhos Funcionam em Novas Janelas (Priority: P1)

Ao abrir uma nova janela no Lumina, o foco é transferido corretamente para ela e todos os atalhos de teclado funcionam imediatamente, sem necessidade de interação manual para "ativar" a janela.

**Why this priority**: Atalhos que não funcionam em janelas novas tornam o multi-window inutilizável — o usuário precisa do teclado para navegar entre painéis, abrir arquivos e controlar o terminal.

**Independent Test**: Abrir uma nova janela via atalho e imediatamente tentar usar qualquer atalho do Lumina (navegar sidebar, alternar foco, fechar janela). Todos devem responder sem clique intermediário.

**Acceptance Scenarios**:

1. **Given** o Lumina com uma janela ativa, **When** o usuário abre uma nova janela pelo atalho, **Then** o foco é transferido para a nova janela e todos os atalhos respondem imediatamente.
2. **Given** múltiplas janelas abertas, **When** o usuário alterna o foco entre elas, **Then** a janela que recebe foco passa a responder a todos os atalhos e a anterior deixa de capturar atalhos exclusivos.
3. **Given** uma nova janela aberta, **When** o usuário pressiona o atalho de fechar janela, **Then** a janela em foco é fechada (não a janela original).

---

### User Story 5 — Fechar Janela Funciona Corretamente (Priority: P1)

O usuário pode fechar qualquer janela aberta no Lumina, incluindo a janela inicial, desde que não seja a última janela restante. O sistema não apresenta mensagem de erro incorreta ao tentar fechar a janela inicial enquanto outras janelas estão abertas.

**Why this priority**: Um bug que impede fechar a janela inicial enquanto outras estão abertas quebra o fluxo básico de gerenciamento de janelas.

**Independent Test**: Abrir duas janelas, colocar o foco na janela inicial e pressionar o atalho de fechar. A janela inicial deve ser fechada e a segunda janela deve permanecer ativa.

**Acceptance Scenarios**:

1. **Given** duas janelas abertas (incluindo a janela inicial), **When** o usuário fecha a janela inicial via atalho, **Then** a janela inicial é removida e o foco vai para a janela restante; nenhuma mensagem de erro é exibida.
2. **Given** apenas uma janela aberta, **When** o usuário tenta fechá-la, **Then** o sistema impede o fechamento e exibe uma mensagem informativa adequada.
3. **Given** três janelas abertas, **When** o usuário fecha qualquer uma delas (não necessariamente a inicial), **Then** as janelas restantes permanecem intactas e funcionais.

---

### Edge Cases

- O que acontece se o usuário pressionar o atalho de toggle da sidebar muito rapidamente (debounce)?
- Como a sidebar se comporta ao redimensionar a janela do terminal (resize do terminal host)?
- O que acontece se o shell configurado não existir no sistema?
- Se o monitor de recursos for ocultado e uma nova sessão for iniciada, o estado persiste?
- O que acontece ao tentar fechar a última janela quando há processos em execução no terminal?

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Cada janela do Lumina DEVE ter sua própria instância de sidebar, independente das demais janelas.
- **FR-002**: O usuário DEVE poder ocultar e exibir a sidebar da janela em foco usando um atalho de teclado dedicado.
- **FR-003**: O espaço liberado pelo ocultamento da sidebar DEVE ser redistribuído para os demais painéis da mesma janela.
- **FR-004**: O estado de visibilidade da sidebar (oculta/visível) DEVE ser mantido por janela individualmente.
- **FR-004b**: O usuário DEVE poder redimensionar a largura da sidebar da janela em foco usando atalhos de teclado (aumentar/diminuir em incrementos fixos).
- **FR-004c**: O usuário DEVE poder redimensionar a sidebar arrastando sua borda com o mouse.
- **FR-005**: Todo novo painel de terminal aberto no Lumina DEVE iniciar o shell padrão do sistema operacional onde o Lumina está rodando.
- **FR-006**: O shell usado pelo terminal DEVE ser configurável pelo usuário nas configurações do Lumina.
- **FR-007**: O monitor de recursos DEVE existir como componente único e global na interface, não sendo duplicado por janela.
- **FR-008**: O usuário DEVE poder ocultar e exibir o monitor de recursos usando um atalho de teclado dedicado; o estado é global (afeta toda a aplicação).
- **FR-009**: Ao abrir uma nova janela, o foco da aplicação DEVE ser transferido imediatamente para ela, com todos os atalhos funcionando sem interação adicional.
- **FR-010**: O sistema de gerenciamento de janelas DEVE rastrear corretamente o número de janelas abertas, permitindo fechar qualquer janela que não seja a única remanescente.
- **FR-011**: A mensagem "única janela aberta" DEVE ser exibida APENAS quando de fato existe somente uma janela; nunca quando há mais de uma janela aberta.

### Key Entities

- **Window (Janela)**: Unidade de layout independente, contendo seus próprios painéis (terminal, editor, sidebar). Possui estado próprio de visibilidade de sidebar.
- **Sidebar**: Explorador de arquivos vinculado a uma Window específica. Estado de visibilidade isolado por Window.
- **Terminal Pane**: Painel que hospeda uma sessão PTY. Deve iniciar o shell padrão do sistema.
- **Resource Monitor**: Componente singleton que exibe métricas do sistema. Estado de visibilidade global, compartilhado entre todas as Windows.
- **Focus State**: Registro de qual Window/Pane está atualmente ativo e capturando eventos de teclado.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: O usuário consegue ocultar e exibir a sidebar de qualquer janela em até 1 ação de teclado, sem afetar outras janelas abertas.
- **SC-002**: 100% dos novos terminais abertos no Lumina em Linux/macOS iniciam com o shell nativo do sistema, nunca com um shell de outra plataforma.
- **SC-003**: O monitor de recursos aparece exatamente uma vez na interface, independente do número de janelas abertas (1 a N janelas).
- **SC-004**: Todos os atalhos de teclado do Lumina funcionam imediatamente após abrir uma nova janela, sem nenhuma interação manual de ativação.
- **SC-005**: O usuário consegue fechar qualquer janela (incluindo a inicial) quando existe mais de uma janela aberta, sem mensagem de erro incorreta.
- **SC-006**: Nenhuma regressão nos atalhos de janelas existentes após as correções de foco.

---

## Assumptions

- O ambiente primário de uso é Linux/macOS; comportamento em Windows é out of scope para este fix (o bug do PowerShell sugere que a implementação atual usa algum fallback Windows).
- O shell padrão é lido de variável de ambiente (`$SHELL`) ou de configuração do Lumina; não é hardcoded.
- A "janela" no contexto deste spec equivale ao conceito de Window no layout multi-window (feature 002), não ao terminal host (iTerm, Windows Terminal, etc.).
- O redimensionamento da sidebar é suportado via atalho de teclado (incrementos fixos) E via mouse (drag). A implementação prioritária é via keybind; mouse drag é desejável mas secundário.
- O estado de visibilidade do monitor de recursos não é persistido entre sessões do Lumina (volta ao padrão visível ao reiniciar).
- Esta feature pressupõe que a feature 002 (multi-window layout) está implementada como base.
