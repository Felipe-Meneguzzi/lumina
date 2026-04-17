# Feature Specification: UX Polish Pack

**Feature Branch**: `006-ux-polish-pack`
**Created**: 2026-04-17
**Status**: Draft
**Input**: User description: "TODO.txt — 8 itens de polimento de UX no Lumina: relógio na status bar, status bar sensível ao terminal focado (git), correção de render inicial quebrado em CLIs como Claude Code, substituir editor próprio por editor externo configurável (nano/vim/nvim), melhorar navegação da sidebar (enter entra em pasta, backspace volta, enter em arquivo abre no editor), cursor por terminal, criação de arquivos/pastas via sidebar (alt+d, alt+f), e estabilidade de render durante alta taxa de saída. Adicional via /speckit.clarify: mudar o foco para uma janela ao clicá-la com o mouse."

## Clarifications

### Session 2026-04-17

- Q: Semântica do clique do mouse em um painel não-focado? → A: Focus-and-pass-through — o clique muda o foco E é entregue ao componente clicado (terminal repassa ao PTY; sidebar seleciona o item sob o cursor).
- Q: Drag de seleção iniciado em painel não-focado? → A: O mousedown transfere foco imediatamente; o drag subsequente seleciona texto no painel recém-focado.
- Q: Indicador visual de painel focado? → A: Borda colorida (cor de destaque no focado, cinza nos demais) — padrão atual.
- Q: Indicador git dirty/clean na status bar? → A: Glifo compacto ao lado do nome da branch — `main ●` (dirty), `main ✓` (clean).
- Q: Backspace na raiz da sidebar? → A: Mensagem temporária na status bar: "Já na raiz".

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Render fiel e estável em terminais (Priority: P1)

Ao abrir um CLI como o Claude Code dentro de um painel do Lumina, o usuário vê a interface do CLI corretamente alinhada desde o primeiro frame. Durante execução de programas que produzem saída intensa (logs em tempo real, streams de build, acompanhamento de arquivos), o conteúdo permanece renderizado na ordem correta sem elementos sobrepostos, truncados ou fora do lugar — sem que o usuário precise redimensionar a janela para "consertar" a tela.

**Why this priority**: Sem rendering confiável o editor deixa de ser utilizável para a tarefa central (rodar ferramentas CLI). É a regressão mais visível e bloqueante da experiência atual.

**Independent Test**: Rodar no painel um CLI TUI conhecido por pintar um cabeçalho rich (ex.: Claude Code) imediatamente após iniciar, e verificar que o cabeçalho aparece alinhado sem precisar de resize. Em paralelo, rodar um loop que imprime muitas linhas por segundo (ex.: `yes | head -n 100000`) e confirmar que o conteúdo permanece íntegro.

**Acceptance Scenarios**:

1. **Given** o Lumina foi iniciado com um painel de terminal vazio, **When** o usuário executa um comando que renderiza um cabeçalho TUI rico (box/art ASCII) antes mesmo da primeira interação de teclado, **Then** o cabeçalho aparece corretamente alinhado no primeiro frame visível — sem linhas em branco extras no topo, sem caracteres sobrepostos e sem depender de redimensionamento.
2. **Given** um painel de terminal executando um comando com saída contínua em alta taxa, **When** a saída continua por pelo menos 30 segundos, **Then** o texto exibido permanece em ordem cronológica correta, sem linhas duplicadas, sobrepostas ou truncadas fora da borda do painel.
3. **Given** um painel com artefatos de render visíveis, **When** o usuário redimensiona a janela do Lumina, **Then** a tela se recompõe corretamente — mas esse redimensionamento não deve ser necessário para se obter o estado correto em operação normal.

---

### User Story 2 - Cursor, contexto e foco por clique (Priority: P1)

Cada painel de terminal mantém o próprio cursor visível e sua própria identidade de contexto (diretório, branch git, estado). Ao alternar o foco entre terminais (por atalho de teclado **ou por clique do mouse sobre o painel**), o cursor reaparece exatamente onde o usuário o deixou no terminal agora focado, e a status bar global reflete imediatamente o estado do terminal focado (branch git, diretório), não de outro painel.

**Why this priority**: Sem cursor visível o usuário fica desorientado ao digitar. E uma status bar que mostra git/contexto do painel errado leva a decisões erradas (checar branch antes de commit, por exemplo).

**Independent Test**: Abrir dois painéis em diretórios/branches git distintos. Alternar o foco e observar: (a) o cursor aparece no terminal focado na posição onde estava antes de perder o foco; (b) a status bar reflete branch e diretório do terminal focado em menos de um frame perceptível.

**Acceptance Scenarios**:

1. **Given** dois painéis A e B, cada um com uma linha de comando em edição parcial em posições diferentes, **When** o usuário alterna o foco de A para B e depois de volta para A, **Then** o cursor reaparece em A exatamente na posição em que estava, e desaparece (ou fica atenuado) em B enquanto este está desfocado.
2. **Given** o painel A está em `~/repo-x` na branch `feature/foo` e o painel B está em `~/repo-y` na branch `main`, **When** o usuário alterna o foco entre A e B, **Then** a status bar atualiza branch e caminho para refletir o painel focado em cada alternância.
3. **Given** um painel em um diretório que não é um repositório git, **When** esse painel recebe o foco, **Then** a status bar indica ausência de contexto git de forma clara (sem exibir o git do painel anterior).
4. **Given** o painel A está focado e o usuário clica com o mouse dentro da área do painel B, **When** o clique é liberado, **Then** o painel B recebe o foco, seu cursor torna-se visível e a status bar passa a refletir o contexto de B.
5. **Given** o usuário clica com o mouse sobre a área da sidebar enquanto um terminal estava focado, **When** o clique é liberado, **Then** o foco passa para a sidebar.

---

### User Story 3 - Sidebar como gerenciador de arquivos com editor externo (Priority: P2)

A sidebar funciona como um gerenciador de arquivos intuitivo: o usuário navega pastas com Enter (entra) e Backspace (volta um nível), cria pastas e arquivos por atalho (Alt+D / Alt+F), e ao abrir um arquivo ele é editado no editor externo preferido do usuário (nano, vim ou nvim) dentro de um painel de terminal — não em um editor reimplementado pelo Lumina. Ao criar uma pasta, o foco da sidebar passa para dentro da nova pasta; ao criar um arquivo, o editor externo abre imediatamente para edição.

**Why this priority**: Transforma a sidebar de um explorador passivo em uma ferramenta de trabalho real, delegando edição ao que o usuário já domina. Reduz a superfície de bugs do editor próprio.

**Independent Test**: Configurar o editor preferido (`nano`, `vim` ou `nvim`). Navegar uma árvore de diretórios pela sidebar usando apenas Enter/Backspace. Criar uma pasta nova com Alt+D, confirmar que o cursor da sidebar está dentro dela. Criar um arquivo com Alt+F e confirmar que o editor escolhido abre para editá-lo.

**Acceptance Scenarios**:

1. **Given** a sidebar está focada em uma pasta, **When** o usuário pressiona Enter, **Then** a sidebar entra na pasta e lista seu conteúdo; o diretório de trabalho dos terminais não muda.
2. **Given** a sidebar está em uma subpasta, **When** o usuário pressiona Backspace, **Then** a sidebar sobe um nível (não sai da raiz configurada como limite); o terminal não é afetado.
3. **Given** a sidebar está focada em um arquivo, **When** o usuário pressiona Enter, **Then** o editor externo configurado abre o arquivo em um painel de terminal.
4. **Given** a sidebar está em uma pasta qualquer, **When** o usuário pressiona Alt+D e digita um nome válido e confirma, **Then** a pasta é criada nesse local e a sidebar entra nela automaticamente.
5. **Given** a sidebar está em uma pasta qualquer, **When** o usuário pressiona Alt+F e digita um nome válido e confirma, **Then** o arquivo é criado e o editor externo configurado abre esse arquivo imediatamente.
6. **Given** o usuário tenta criar uma pasta/arquivo com nome já existente, **When** confirma, **Then** recebe uma mensagem de erro clara e a criação é abortada sem alterar o sistema de arquivos.
7. **Given** a configuração não define um editor externo, **When** o usuário tenta abrir um arquivo pela sidebar, **Then** o sistema usa `nano` como editor padrão.

---

### User Story 4 - Relógio visível na status bar (Priority: P3)

A status bar passa a exibir a hora atual, atualizada continuamente, para que o usuário tenha noção do tempo enquanto trabalha em foco profundo no terminal.

**Why this priority**: Melhoria de conveniência, não-bloqueante. Não altera fluxo de trabalho, mas é pedido direto do usuário.

**Independent Test**: Abrir o Lumina e observar a status bar por pelo menos 1 minuto; a hora exibida deve avançar em tempo real.

**Acceptance Scenarios**:

1. **Given** o Lumina está em execução, **When** o usuário olha para a status bar, **Then** vê a hora atual no formato HH:MM.
2. **Given** o Lumina está em execução há pelo menos um minuto, **When** o usuário observa a status bar, **Then** a hora exibida avança de acordo com o relógio do sistema sem necessidade de interação.

---

### Edge Cases

- Alternância rápida de foco entre muitos terminais não deve causar "vazamento" de cursor (cursor visível em mais de um painel ao mesmo tempo).
- Criar arquivo/pasta com nome contendo caracteres inválidos para o sistema de arquivos deve ser rejeitado com mensagem clara, sem travar o prompt da sidebar.
- Cancelar o prompt de criação (ESC) deve fechar o prompt sem criar nada e devolver o foco ao estado anterior.
- Abrir um arquivo binário ou muito grande pela sidebar deve delegar ao editor externo — o comportamento resultante é responsabilidade do editor, não do Lumina.
- Se o editor externo configurado não estiver instalado no PATH, o Lumina deve avisar claramente em vez de abrir um painel em branco.
- Um terminal sem contexto git (diretório sem `.git`) deve apresentar a status bar sem campo de branch, não um campo vazio ambíguo ou o último git conhecido.
- Saída que contém sequências de controle ANSI complexas (spinners, progress bars que reescrevem a mesma linha) deve continuar sendo renderizada corretamente sob alta taxa.
- Redimensionamento da janela durante saída intensa não pode corromper o buffer de scrollback.

## Requirements *(mandatory)*

### Functional Requirements

**Render e terminais**

- **FR-001**: O sistema DEVE garantir que o conteúdo renderizado por processos filhos em qualquer painel de terminal apareça corretamente desde o primeiro frame após a inicialização do processo, sem que o usuário precise redimensionar a janela para corrigir artefatos.
- **FR-002**: O sistema DEVE manter integridade visual (ordem, alinhamento e posicionamento correto dos elementos) em painéis recebendo saída contínua em alta taxa por pelo menos 30 segundos ininterruptos.
- **FR-003**: O sistema DEVE preservar e restaurar a posição do cursor de cada painel de terminal individualmente, mostrando o cursor apenas no painel atualmente focado.
- **FR-004**: O sistema DEVE indicar visualmente qual painel está focado pintando sua borda com uma cor de destaque distinta, enquanto os painéis não-focados mantêm borda cinza neutra. Este indicador é adicional à visibilidade do cursor.
- **FR-004a**: O sistema DEVE mover o foco para o painel (terminal ou sidebar) sob o cursor do mouse quando o usuário clicar sobre sua área.
- **FR-004b**: O clique que transfere foco DEVE também ser entregue ao painel clicado (pass-through) — terminais encaminham o evento ao PTY do processo filho; a sidebar seleciona o item sob o cursor. Não existe comportamento de "primeiro clique descartado".
- **FR-004c**: A transferência de foco DEVE ocorrer no evento de mousedown (não mouseup), de modo que um drag iniciado em painel não-focado selecione texto nesse painel já como focado.

**Status bar**

- **FR-005**: O sistema DEVE exibir a hora corrente (formato HH:MM) na status bar, atualizando-a continuamente sem ação do usuário.
- **FR-006**: A status bar DEVE refletir o estado git do terminal atualmente focado, exibindo o nome da branch seguido de um glifo de estado: `●` quando há alterações não commitadas (dirty) ou `✓` quando o working tree está limpo. A exibição deve atualizar-se em cada troca de foco.
- **FR-007**: A status bar DEVE indicar de forma clara a ausência de contexto git quando o terminal focado não está em um repositório git (em vez de exibir dados de outro painel).

**Sidebar — navegação**

- **FR-008**: Pressionar Enter com uma pasta selecionada na sidebar DEVE entrar na pasta (passar a listar seu conteúdo) sem alterar o diretório de trabalho de nenhum terminal.
- **FR-009**: Pressionar Backspace na sidebar DEVE subir um nível na árvore de diretórios, respeitando a raiz configurada como limite máximo. Quando a sidebar já está na raiz, o sistema DEVE exibir uma mensagem temporária na status bar ("Já na raiz") com duração de **2 segundos**, sem alterar o estado da sidebar.
- **FR-010**: Pressionar Enter com um arquivo selecionado na sidebar DEVE abrir esse arquivo no editor externo configurado, dentro de um painel de terminal.

**Sidebar — criação**

- **FR-011**: Pressionar Alt+D DEVE abrir um prompt inline na sidebar pedindo o nome da nova pasta; ao confirmar, o sistema DEVE criar a pasta no diretório atualmente navegado pela sidebar e em seguida entrar nela automaticamente.
- **FR-012**: Pressionar Alt+F DEVE abrir um prompt inline pedindo o nome do novo arquivo; ao confirmar, o sistema DEVE criar o arquivo vazio e abri-lo imediatamente no editor externo configurado.
- **FR-013**: O sistema DEVE rejeitar com mensagem clara qualquer tentativa de criação com nome vazio, contendo caracteres inválidos para o sistema de arquivos, ou já existente no diretório atual.
- **FR-014**: O sistema DEVE permitir cancelar o prompt de criação com ESC, fechando o prompt sem efeitos colaterais.

**Editor externo**

- **FR-015**: O sistema DEVE permitir que o usuário configure qual editor externo usar (nano, vim ou nvim) através do arquivo de configuração.
- **FR-016**: O sistema DEVE usar `nano` como editor padrão quando nenhuma configuração válida for encontrada.
- **FR-017**: O sistema NÃO DEVE manter um editor de texto próprio embarcado; toda edição de arquivos deve ser delegada ao editor externo configurado.
- **FR-018**: Quando o editor externo configurado não está disponível no PATH, o sistema DEVE (a) NÃO criar um painel de terminal, (b) exibir uma notificação de erro na status bar com o texto `editor '<nome>' não encontrado no PATH`, e (c) manter o valor configurado inalterado (sem fallback silencioso para `nano`).

### Key Entities

- **Painel de terminal focado**: O painel que recebe entrada do teclado e cuja identidade (diretório, branch git, cursor) alimenta a status bar.
- **Preferência de editor externo**: Configuração persistente que indica qual binário externo (`nano`, `vim`, `nvim`) deve ser usado para edição.
- **Diretório navegado pela sidebar**: Estado da sidebar independente do diretório de trabalho dos terminais — muda ao entrar/sair de pastas via Enter/Backspace.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Em 100% das inicializações de CLIs TUI conhecidas (Claude Code, htop, lazygit) em um painel novo, o primeiro frame renderizado está visualmente correto sem necessidade de redimensionamento.
- **SC-002**: Sob carga de saída de pelo menos 5.000 linhas por minuto por 5 minutos consecutivos, o usuário não observa nenhum artefato visual (texto sobreposto, linhas fora de ordem) no painel receptor.
- **SC-003**: Ao alternar o foco entre painéis, a status bar reflete o CWD do novo painel **no mesmo ciclo de `Update` em que o foco muda** (budget ≤16 ms, consistente com §IV do constitution). O campo git pode chegar em um segundo frame quando depende do `PaneGitStateMsg` assíncrono — nesse caso aparece em ≤250 ms (200 ms de timeout do `git` + um frame).
- **SC-004**: O cursor aparece visível em exatamente um painel a qualquer momento — nunca em zero (quando há terminal ativo) nem em dois ou mais simultaneamente.
- **SC-005**: Um usuário novo consegue criar uma pasta aninhada, um arquivo dentro dela, editá-lo e salvar usando apenas a sidebar e atalhos documentados, em menos de 60 segundos.
- **SC-006**: A hora exibida na status bar nunca diverge do relógio do sistema em mais de 60 segundos.
- **SC-007**: Nenhum caminho de criação de arquivo/pasta via sidebar causa corrupção de estado ou trava da UI quando entradas inválidas são fornecidas — todas falham com mensagem clara e a sidebar permanece utilizável.

## Assumptions

- O usuário está em Linux ou macOS; suporte Windows continua fora de escopo conforme o projeto atual.
- Pelo menos um dos editores suportados (`nano`, `vim`, `nvim`) está instalado no sistema do usuário; `nano` é assumido como padrão por estar presente na maioria das distribuições.
- A raiz de navegação da sidebar é o diretório de trabalho em que o Lumina foi iniciado, conforme comportamento atual do projeto.
- O estado git exibido na status bar limita-se à branch atual e a um indicador simples de sujo/limpo — histórico detalhado, ahead/behind e stash status ficam fora de escopo desta feature.
- "Alta taxa de saída" neste documento refere-se a saídas típicas de ferramentas de desenvolvimento (build logs, testes, streams de aplicação) — não a testes de stress sintéticos acima de ~10.000 linhas por segundo.
- Atalhos Alt+D e Alt+F são capturados pela sidebar apenas quando a sidebar está focada; não interferem em atalhos do terminal focado.
