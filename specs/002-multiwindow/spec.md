# Feature Specification: Multiwindow Layout

**Feature Branch**: `002-multiwindow`  
**Created**: 2026-04-16  
**Status**: Draft  
**Input**: User description: "Vamos adicionar uma função multiwindow, eu preciso poder abrir mais de um terminal ou arquivo na mesma sessão do lumina, em layout de 2,3 ou 4, além de poder redimencionar as janelas, incluindo a sidebar"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Dividir o espaço em múltiplos painéis (Priority: P1)

O usuário está editando um arquivo e quer abrir um segundo painel ao lado para comparar dois arquivos ou ter um terminal disponível enquanto edita. Ele aciona um atalho de teclado para dividir a tela e escolhe o layout desejado (2, 3 ou 4 painéis).

**Why this priority**: É o núcleo da feature. Sem a divisão de painéis, todo o restante não existe. Entrega valor imediato ao permitir edição ou monitoramento paralelo.

**Independent Test**: Pode ser testado abrindo o Lumina, acionando o comando de split e verificando que dois painéis independentes são exibidos e recebem foco alternadamente.

**Acceptance Scenarios**:

1. **Given** o Lumina está aberto com um arquivo ou terminal, **When** o usuário aciona o atalho de split horizontal ou vertical, **Then** a tela é dividida em dois painéis igualmente dimensionados, cada um exibindo seu próprio conteúdo.
2. **Given** dois painéis abertos, **When** o usuário aciona novamente o split no painel ativo, **Then** o painel ativo é subdividido, resultando em 3 painéis totais.
3. **Given** três painéis abertos, **When** o usuário aciona o split uma vez mais, **Then** o painel ativo é subdividido, resultando em 4 painéis totais (limite máximo).
4. **Given** 4 painéis já abertos, **When** o usuário tenta acionar o split novamente, **Then** o sistema exibe aviso de que o limite de 4 painéis foi atingido e não cria novo painel.

---

### User Story 2 — Navegar entre painéis (Priority: P2)

O usuário alterna o foco entre os painéis abertos para interagir com cada um independentemente, seja para editar um arquivo ou digitar comandos num terminal.

**Why this priority**: Sem navegação entre painéis, a divisão de tela seria inútil — o usuário ficaria preso em um único painel.

**Independent Test**: Pode ser testado abrindo dois painéis e verificando que o atalho de navegação transfere o foco visual (destaque de borda ativa) de um painel para o outro.

**Acceptance Scenarios**:

1. **Given** múltiplos painéis abertos, **When** o usuário usa o atalho de navegação direcional (esquerda/direita/cima/baixo), **Then** o foco se move para o painel vizinho na direção indicada.
2. **Given** o foco está no último painel da sequência, **When** o usuário navega além do limite, **Then** o foco retorna ao primeiro painel (navegação cíclica).
3. **Given** um painel com foco ativo, **When** o usuário interage (digita, rola), **Then** apenas o painel com foco responde — os demais permanecem inalterados.

---

### User Story 3 — Redimensionar painéis incluindo a sidebar (Priority: P3)

O usuário quer ajustar o tamanho de cada painel e da sidebar de arquivos para adequar o espaço ao seu fluxo de trabalho — por exemplo, dar mais espaço ao editor e menos ao terminal.

**Why this priority**: O redimensionamento melhora significativamente a ergonomia, mas os layouts com tamanhos iguais já entregam valor básico sem ele.

**Independent Test**: Pode ser testado abrindo dois painéis e verificando que o atalho de redimensionamento aumenta/diminui a largura/altura do painel ativo enquanto o vizinho ocupa o espaço restante.

**Acceptance Scenarios**:

1. **Given** dois painéis lado a lado, **When** o usuário aciona o atalho de expandir painel ativo, **Then** o painel ativo aumenta sua largura e o painel vizinho diminui proporcionalmente.
2. **Given** a sidebar visível, **When** o usuário aciona o atalho de expandir/recolher sidebar, **Then** a sidebar aumenta ou diminui sua largura, redistribuindo o espaço para os painéis de conteúdo.
3. **Given** um painel redimensionado ao mínimo configurado, **When** o usuário tenta recolhê-lo ainda mais, **Then** o sistema respeita o tamanho mínimo e não permite redução adicional.

---

### User Story 4 — Fechar um painel (Priority: P2)

O usuário não precisa mais de um painel e quer fechá-lo, voltando ao layout com menos painéis sem encerrar a sessão do Lumina.

**Why this priority**: Complemento natural ao split: o usuário deve poder desfazer a divisão assim como a criou.

**Independent Test**: Pode ser testado abrindo dois painéis, fechando um e verificando que o painel restante ocupa todo o espaço disponível anteriormente.

**Acceptance Scenarios**:

1. **Given** múltiplos painéis abertos, **When** o usuário aciona o atalho de fechar painel ativo, **Then** o painel é removido e os demais redistribuem o espaço de forma proporcional.
2. **Given** apenas um painel restante, **When** o usuário tenta fechá-lo, **Then** o sistema impede o fechamento (pelo menos um painel deve estar sempre aberto) e exibe aviso.
3. **Given** um painel terminal fechado, **When** o painel é removido, **Then** o processo do terminal associado é encerrado corretamente.

---

### Edge Cases

- O que acontece quando o terminal é redimensionado pelo sistema operacional (mudança de tamanho da janela do emulador de terminal)? Os painéis devem ser redistribuídos automaticamente.
- O que acontece se o arquivo aberto num painel for deletado externamente enquanto o painel está visível?
- Como o sistema se comporta quando há 4 painéis em uma janela muito pequena (ex: 80×24 colunas)? Deve exibir aviso ou colapsar graciosamente.
- O que acontece com o histórico de scroll de cada painel ao redimensionar?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema DEVE permitir dividir qualquer painel ativo em dois subpainéis (split horizontal ou vertical) com um único atalho de teclado.
- **FR-002**: O sistema DEVE suportar layouts de 2, 3 e 4 painéis simultâneos, sendo 4 o limite máximo em uma sessão.
- **FR-003**: Cada painel DEVE poder exibir independentemente um editor de arquivo ou um terminal PTY.
- **FR-004**: O usuário DEVE poder navegar entre painéis usando atalhos direcionais de teclado.
- **FR-005**: O sistema DEVE indicar visualmente qual painel possui o foco ativo (ex: borda destacada).
- **FR-006**: O usuário DEVE poder redimensionar painéis incrementalmente usando atalhos de teclado, sem uso de mouse.
- **FR-007**: A sidebar de arquivos DEVE ser redimensionável pelo mesmo mecanismo de redimensionamento dos painéis.
- **FR-008**: O usuário DEVE poder fechar o painel ativo com um atalho de teclado, exceto quando for o único painel restante.
- **FR-009**: Ao fechar um painel, o espaço deve ser redistribuído automaticamente entre os painéis remanescentes.
- **FR-010**: Ao redimensionar a janela do emulador de terminal, todos os painéis DEVEM ser redistribuídos e o conteúdo (incluindo PTY) notificado do novo tamanho.
- **FR-011**: Cada painel DEVE manter seu próprio estado de scroll, cursor e conteúdo independentemente dos outros.
- **FR-012**: Todos os atalhos de teclado para operações de painel DEVEM ser configuráveis via arquivo de configuração.

### Key Entities

- **Painel (Pane)**: Unidade de visualização independente. Pode conter um editor de arquivo ou um terminal. Possui dimensões próprias, estado de foco e conteúdo.
- **Layout**: Arranjo dos painéis na área de trabalho. Define como os painéis são dispostos (horizontal, vertical, misto) e suas proporções relativas.
- **Sidebar**: Painel lateral de navegação de arquivos, redimensionável e colapsável, mas não sujeito ao limite de 4 painéis principais.
- **Foco Ativo**: Estado que indica qual painel recebe entrada do teclado. Apenas um painel possui foco por vez.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: O usuário consegue criar um layout de 2 painéis em no máximo 2 ações (atalhos de teclado) a partir de uma sessão com painel único.
- **SC-002**: A alternância de foco entre painéis ocorre visivelmente em menos de 1 ação (um único atalho de teclado).
- **SC-003**: O redimensionamento de painéis responde a cada ação do usuário sem atraso perceptível, mantendo renderização fluida.
- **SC-004**: Ao redimensionar a janela do emulador, todos os painéis se ajustam automaticamente sem intervenção do usuário.
- **SC-005**: 100% dos terminais abertos em painéis recebem a notificação de novo tamanho ao redimensionar, garantindo que o conteúdo PTY não apresente quebras de layout.
- **SC-006**: O fechamento de um painel não encerra a sessão do Lumina nem afeta o conteúdo dos demais painéis.
- **SC-007**: A feature funciona corretamente em janelas de terminal com largura mínima de 120 colunas e altura mínima de 30 linhas para layouts de 4 painéis.

## Assumptions

- Assume-se que mouse não é suportado — toda interação de split, navegação e redimensionamento ocorre via teclado.
- Assume-se que layouts mistos (ex: 2 painéis na coluna esquerda + 1 grande na direita) são escopo de versões futuras; v1 suporta splits binários recursivos (cada split divide um painel em dois).
- Assume-se que a sidebar permanece sempre no lado esquerdo e não pode ser movida para outros lados.
- Assume-se que o número máximo de 4 painéis é fixo e não configurável pelo usuário na v1.
- Assume-se que ao abrir o Lumina, a sessão sempre inicia com layout de painel único (o padrão atual não muda).
- Assume-se que cada painel pode conter apenas um arquivo ou terminal por vez (sem abas dentro do painel na v1).
