# Feature Specification: CLI Startup Flags

**Feature Branch**: `004-cli-startup-flags`
**Created**: 2026-04-17
**Status**: Draft
**Input**: User description: "Quero algumas features customizaveis, por exemplo alterar o maxPanes na inicialização, pra poder rodar lumina -mp 10, ai ele seta pra 10, mantem o default como 4 mas permite alterar, alem disso vai ter parametros de startPanes, que vai definir com quantos vai começar ja, quero passar por exemplo -sp h3 e vai ter 3 panes divididos horizontalmente, ou -sp v2 e ai vem 2 panes divididos verticalmente, outra flag q eu quero é a flag de -sc (StartCommand), pra poder passar -sc claude e ele abrir todos os panes com o claude inicialmente (Só nos iniciais, nao no split), ai ficaria lumina -mp 10 -sp h3 -sc claude (vai abrir o lumina, com maximo de 10 janelas, com 3 iniciais divididas horizontalmente e rodar o claude)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Override max panes at startup (Priority: P1)

Um usuário avançado precisa trabalhar com mais painéis simultâneos do que o limite padrão permite (ex.: acompanhar vários logs, serviços e terminais em paralelo). Ele inicia o Lumina passando uma flag que redefine o número máximo de painéis permitidos para aquela sessão, sem precisar editar arquivos de configuração.

**Why this priority**: É a capacidade mais básica das três e desbloqueia as outras duas. Sem ela, um usuário não consegue iniciar o editor já com mais painéis do que o default e fica preso ao teto atual.

**Independent Test**: Pode ser validada isoladamente executando `lumina -mp 10`, abrindo painéis via comandos de split até ultrapassar 4, e confirmando que o editor aceita criar painéis adicionais até o novo limite (10) e rejeita apenas o 11º.

**Acceptance Scenarios**:

1. **Given** o Lumina é iniciado sem nenhuma flag, **When** o usuário tenta criar o 5º painel, **Then** o editor bloqueia a criação (comportamento default preservado).
2. **Given** o Lumina é iniciado com `-mp 10`, **When** o usuário cria painéis sucessivos via split, **Then** o editor permite criar até 10 painéis e bloqueia somente o 11º.
3. **Given** o Lumina é iniciado com `-mp 1`, **When** o usuário tenta qualquer split, **Then** o editor rejeita o split e mantém apenas um painel.

---

### User Story 2 - Start with a pre-split layout (Priority: P2)

Um usuário que sempre começa sua rotina com um layout específico (ex.: 3 painéis horizontais para monitoramento) quer evitar repetir os atalhos de split toda vez. Ele passa uma flag que já cria o layout desejado no boot.

**Why this priority**: Depende do US1 quando o layout solicitado excede o teto atual, mas agrega valor direto ao fluxo recorrente do usuário e economiza passos manuais toda inicialização.

**Independent Test**: Executando `lumina -sp h3`, o editor abre diretamente com três painéis dispostos horizontalmente (lado a lado), sem que o usuário precise acionar nenhum atalho.

**Acceptance Scenarios**:

1. **Given** o usuário executa `lumina -sp h3`, **When** o editor termina de inicializar, **Then** três painéis de largura proporcional aparecem dispostos horizontalmente e o foco está em um deles.
2. **Given** o usuário executa `lumina -sp v2`, **When** o editor termina de inicializar, **Then** dois painéis aparecem empilhados verticalmente.
3. **Given** o usuário executa `lumina -sp h1`, **When** o editor termina de inicializar, **Then** apenas um painel aparece (equivale ao boot default).
4. **Given** o usuário executa `lumina -sp h5` sem `-mp`, **When** o editor inicia, **Then** o teto efetivo de painéis é elevado automaticamente para acomodar os 5 painéis iniciais.

---

### User Story 3 - Autorun a command in initial panes (Priority: P3)

Um usuário que sempre começa suas sessões rodando o mesmo programa (ex.: `claude`) em cada painel quer que o Lumina já execute esse comando em todos os painéis iniciais ao abrir, em vez de invocá-lo manualmente em cada um.

**Why this priority**: Complementar ao US2 — traz ganho de produtividade real mas só faz sentido quando a sessão inicia com múltiplos painéis pré-criados. Sem US2, o efeito cosmético se reduz a um único painel.

**Independent Test**: Executando `lumina -sp h3 -sc claude`, o editor abre com 3 painéis horizontais e o processo `claude` é executado em cada um deles; após o usuário criar um 4º painel via split, esse painel novo não roda `claude` — abre apenas o shell padrão.

**Acceptance Scenarios**:

1. **Given** o usuário executa `lumina -sp h3 -sc claude`, **When** o editor termina de inicializar, **Then** cada um dos 3 painéis exibe o programa `claude` em execução.
2. **Given** uma sessão foi iniciada com `-sc claude` em 3 painéis iniciais, **When** o usuário cria um novo painel via split manual, **Then** o novo painel abre o shell default do usuário, não o `claude`.
3. **Given** o usuário executa `lumina -sc claude` sem `-sp`, **When** o editor inicia, **Then** o único painel default executa `claude`.
4. **Given** o usuário passa `-sc` com um comando inexistente, **When** o editor inicia, **Then** cada painel exibe uma mensagem clara de falha ao iniciar o comando e a sessão permanece utilizável (fallback para shell ou prompt de erro visível).

---

### Edge Cases

- **Valores inválidos numéricos**: `-mp 0`, `-mp -5`, `-mp abc` devem ser rejeitados na inicialização com mensagem clara e encerrar antes de abrir o editor.
- **Formato inválido de `-sp`**: valores que não seguem `h<N>` ou `v<N>` (ex.: `-sp 3`, `-sp d2`, `-sp h`) devem ser rejeitados com mensagem explicando o formato esperado.
- **`-sp` acima de `-mp` explícito**: quando o usuário passa os dois e o número de painéis iniciais excede o máximo, a inicialização deve falhar com mensagem apontando o conflito.
- **`-sc` com aspas/argumentos**: comando composto (ex.: `-sc "claude --model opus"`) deve ser aceito como um único comando com argumentos.
- **Combinação com argumento posicional de arquivo**: hoje `lumina arquivo.txt` abre o arquivo; as novas flags devem coexistir com esse uso sem conflito de parsing.
- **`-sp h1` ou `-sp v1`**: equivale ao boot default (um painel). Não deve gerar erro.
- **`-mp` abaixo do número de painéis iniciais padrão**: `-mp 1` deve ser aceito e simplesmente impedir qualquer split posterior.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST aceitar a flag `-mp <N>` na inicialização e usar `N` como o número máximo de painéis permitidos naquela sessão, substituindo o default (4).
- **FR-002**: O sistema MUST preservar o comportamento default (máximo de 4 painéis) quando `-mp` não é fornecido.
- **FR-003**: O sistema MUST rejeitar `-mp` com valores não inteiros, zero ou negativos, exibindo mensagem de erro e encerrando antes de abrir a TUI.
- **FR-004**: O sistema MUST aceitar a flag `-sp <orient><N>` onde `orient` é `h` (horizontal/lado a lado) ou `v` (vertical/empilhado) e `N` é um inteiro ≥ 1.
- **FR-005**: O sistema MUST criar, na inicialização, exatamente `N` painéis dispostos na orientação indicada por `-sp`, com tamanhos proporcionais, antes do primeiro frame visível.
- **FR-006**: O sistema MUST rejeitar valores de `-sp` fora do formato `h<N>` / `v<N>` com mensagem de erro explicando o formato esperado.
- **FR-007**: O sistema MUST elevar o teto efetivo de painéis para acomodar `N` de `-sp` quando `-mp` não é informado e `N` excede o default.
- **FR-008**: O sistema MUST falhar a inicialização com mensagem clara quando `-mp` e `-sp` são ambos informados e `N` de `-sp` excede o valor de `-mp`.
- **FR-009**: O sistema MUST aceitar a flag `-sc <comando>` e executar esse comando em cada painel criado pela flag `-sp` (ou no painel default quando `-sp` está ausente).
- **FR-010**: O sistema MUST aplicar `-sc` APENAS aos painéis iniciais; painéis criados posteriormente via split abrem o shell default do usuário.
- **FR-011**: O sistema MUST, quando o comando de `-sc` falhar ao executar em um painel, exibir mensagem de erro nesse painel sem encerrar a sessão inteira.
- **FR-012**: O sistema MUST permitir comandos com argumentos em `-sc` (ex.: `-sc "claude --model opus"`) tratando a string inteira como um único comando composto.
- **FR-013**: O sistema MUST permitir combinar `-mp`, `-sp` e `-sc` em qualquer ordem na linha de comando.
- **FR-014**: O sistema MUST manter compatibilidade com o uso atual `lumina <arquivo>` (argumento posicional para abrir arquivo) quando combinado com as novas flags.
- **FR-015**: O sistema MUST expor as novas flags na mensagem de ajuda (`lumina --help` ou equivalente), descrevendo formato e valor default de cada uma.

### Key Entities

- **Startup Configuration**: Representa as escolhas feitas pelo usuário na linha de comando para aquela execução. Atributos: teto de painéis, layout inicial (orientação + contagem), comando a executar nos painéis iniciais, arquivo a abrir. É efêmero — vive apenas para aquela sessão e não é persistido.
- **Initial Pane**: Painel criado pela flag `-sp` durante o boot. Se distingue de painéis criados depois porque pode rodar o `StartCommand` em vez do shell default.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Um usuário consegue iniciar o Lumina já com o layout desejado (painéis e comando) em um único comando de terminal, sem precisar de atalhos manuais após o boot.
- **SC-002**: A sessão inicializada com `-sp hN` exibe os `N` painéis no primeiro frame visível — não há frame intermediário mostrando um único painel seguido de splits animados.
- **SC-003**: Combinações inválidas de flags são rejeitadas antes da TUI abrir, em menos de 1 segundo, com mensagem de erro clara que indica a flag problemática e o formato correto.
- **SC-004**: O comportamento default (sem flags) permanece idêntico ao da versão anterior em 100% dos casos — nenhuma regressão perceptível para usuários que não adotam as novas flags.
- **SC-005**: Após uma semana de uso, um usuário que adotou o fluxo `lumina -sp hN -sc X` consegue descrever de memória o que cada flag faz — o mnemônico (maxPanes, startPanes, startCommand) é intuitivo o suficiente.

## Assumptions

- Os nomes curtos das flags (`-mp`, `-sp`, `-sc`) são a forma canônica; versões longas equivalentes (ex.: `--max-panes`, `--start-panes`, `--start-command`) são desejáveis mas não obrigatórias nesta iteração.
- O comando passado em `-sc` é executado no shell default do usuário (mesmo shell usado pelos painéis normais), não em um interpretador específico do Lumina.
- Erros de inicialização (flag inválida, conflito `-mp`/`-sp`) vão para stderr e o processo encerra com código de saída não-zero, sem abrir a TUI.
- A flag `-sp` com `N=1` é aceita e equivale a não passar a flag; não é tratada como erro.
- Não há persistência: as escolhas de `-mp`, `-sp`, `-sc` valem apenas para a sessão corrente e não alteram nenhum arquivo de configuração.
- A ordem relativa entre flags e o argumento posicional de arquivo é livre (parser padrão de flags trata flags antes do posicional; isso é suficiente para o escopo desta feature).
- Não há limite rígido superior imposto pelo produto em `-mp` — o limite prático vem da usabilidade do layout na tela (painéis muito pequenos deixam de ser úteis), e a decisão de quando parar fica com o usuário.
