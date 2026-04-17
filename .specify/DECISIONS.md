# Decisões Técnicas do Projeto

## Contexto do Projeto

Lumina é um editor de terminal estilo VSCode — uma TUI (Terminal User Interface) escrita em Go que combina painéis de terminal interativo, explorador de arquivos, status bar com métricas do sistema e (futuramente) edição de texto. Projetado para rodar inteiramente no terminal, com foco em produtividade e leveza.

---

## Decisões

### 1. Linguagem de Programação
**Decisão:** Go  
**Alternativas consideradas:** Rust, Python, C/C++  
**Justificativa:** Binário único sem runtime externo, excelente suporte a concorrência (goroutines para múltiplos painéis), ecossistema TUI maduro, compilação rápida, manutenibilidade alta para um projeto de longo prazo.  
**Trade-offs aceitos:** Verbosidade maior que Python; sem borrowing-checker como Rust (gerenciamento de memória mais manual).  
**Data:** 2026-04-16

---

### 2. Framework TUI
**Decisão:** Bubble Tea + Lip Gloss + Bubbles  
**Alternativas consideradas:** tview/tcell, termui, gocui, Charm Glow (standalone)  
**Justificativa:** Bubble Tea implementa arquitetura Elm (Model/Update/View) que escala bem para múltiplos componentes. Lip Gloss cuida de estilização declarativa. Bubbles fornece componentes prontos (viewport, textinput, spinner). Ecossistema coeso mantido pela Charm, comunidade ativa.  
**Trade-offs aceitos:** Curva de aprendizado inicial no modelo de mensagens assíncronas do Bubble Tea; menos controle de baixo nível que tcell puro.  
**Data:** 2026-04-16

---

### 3. Gerenciamento de PTY
**Decisão:** `creack/pty`  
**Alternativas consideradas:** `mvdan/sh`, syscall direto via `golang.org/x/sys`  
**Justificativa:** Os painéis de terminal do Lumina precisam de PTYs reais para suportar programas interativos (shell, htop, vim, etc.). `creack/pty` é a biblioteca canônica para esse padrão em Go — API simples (`pty.Start`, `pty.Setsize` para resize), bem mantida e amplamente adotada. `mvdan/sh` não executa programas que dependem de terminal real; syscall direto reintroduz toda a complexidade que `creack/pty` já resolve.  
**Trade-offs aceitos:** Sem suporte nativo a Windows (Linux/macOS only) — aceitável dado o perfil de usuário do projeto.  
**Data:** 2026-04-16

---

### 4. Coleta de Métricas do Sistema
**Decisão:** `github.com/shirou/gopsutil/v3`  
**Alternativas consideradas:** Leitura direta de `/proc`, bibliotecas de métricas específicas por OS  
**Justificativa:** Entrega CPU, memória, disco e rede com uma linha de código por métrica. Cross-platform por padrão. A latência mínima de amostragem de ~100ms é imperceptível em uma TUI com ciclo de atualização de 1–2 segundos. Leitura direta de `/proc` só valeria em ambientes embarcados ou sem dependências externas, o que não é o caso do Lumina.  
**Trade-offs aceitos:** Overhead ligeiramente maior que acesso direto ao kernel; dependência de biblioteca externa para funcionalidade que seria trivial implementar só para Linux.  
**Data:** 2026-04-16

---

### 5. Arquitetura de Componentes
**Decisão:** Componentes Compostos (Padrão B) — cada painel implementa `tea.Model` completo, composto pelo modelo raiz via delegação explícita.  
**Alternativas consideradas:** Modelo Único monolítico (Padrão A), Event Bus centralizado (Padrão C)  
**Justificativa:** Padrão idiomático do Bubble Tea, documentado nos exemplos oficiais. Cada componente (`terminal`, `sidebar`, `statusbar`) é testável isoladamente com estado encapsulado. O modelo raiz em `app/app.go` roteia mensagens via type-switch sem se tornar um God Object. Mensagens cross-componente são definidas em `msgs/msgs.go` com tipos explícitos, evitando acoplamento circular.  
**Trade-offs aceitos:** Comunicação entre componentes exige definição explícita de `tea.Msg` customizadas — mais verboso que um event bus, mas mantém o fluxo de dados unidirecional do Bubble Tea.  
**Data:** 2026-04-16

---

### 6. Breaking Change: PtyOutputMsg e PtyInputMsg ganham campo PaneID (feature 002-multiwindow)

**Decisão:** Adicionar `PaneID int` aos structs `msgs.PtyOutputMsg` e `msgs.PtyInputMsg`.  
**Justificativa:** Com múltiplos terminais PTY simultâneos (multiwindow feature), o layout manager precisa rotear o output de cada terminal para o `LeafNode` correto na árvore de painéis. O `PaneID` é capturado pelo closure de `waitForOutput` antes de `Init()` ser chamado, garantindo que cada goroutine de leitura etiquete seu output com o ID correto.  
**Impacto:** Todo código que constrói `PtyOutputMsg` ou `PtyInputMsg` com campos posicionais (sem nomear) precisaria ser atualizado; code que usa campos nomeados não é afetado (Go inicializa campos omitidos com zero value). A mudança mantém zero-value de `PaneID=0` funcional para o painel único do app.go legado durante a transição.  
**Alternativa considerada:** Manter structs sem PaneID e usar channels separados por terminal. Rejeitada: os channels não se integram ao modelo de mensagens do Bubble Tea (tea.Cmd deve retornar tea.Msg, não ler de channel diretamente sem wrapper).  
**Constituição:** Breaking change em `msgs/msgs.go` — bump MINOR da versão da Constituição conforme §Development Workflow.  
**Data:** 2026-04-16

---

### 7. Layout Manager: package components/layout com binary split tree (feature 002-multiwindow)

**Decisão:** Criar `components/layout` com arquitetura de árvore binária de splits (PaneNode interface → SplitNode | LeafNode), inspirada no Hyprland.  
**Justificativa:** A árvore binária permite qualquer combinação de 2, 3 ou 4 painéis sem layouts fixos. Cada split divide um nó em dois filhos com ratio configurável. O layout manager como package separado mantém `app/app.go` como roteador puro (responsabilidade única da Constituição §I). O `app.Model` delega toda a geometria de conteúdo ao `layout.Model`.  
**Alternativa considerada:** Grid fixo (2×1, 2×2). Rejeitada: inflexível para 3 painéis em L; não reproduz o UX do Hyprland.  
**UX Reference:** Hyprland DE — keybindings `Alt+H/J/K/L` para navegação, `Alt+|` e `Alt+_` para split, `Alt+Q` para fechar.  
**Data:** 2026-04-16

---

## Estrutura de Diretórios Definida

```
lumina/
├── main.go
├── app/
│   ├── app.go          # Model raiz — compõe e roteia entre componentes
│   └── keymap.go       # Keybindings globais
├── components/
│   ├── terminal/       # Painel de terminal (wraps creack/pty)
│   ├── sidebar/        # Explorador de arquivos
│   ├── statusbar/      # Métricas em tempo real (gopsutil)
│   └── editor/         # (futuro) buffer de texto
├── msgs/
│   └── msgs.go         # Tipos tea.Msg compartilhados entre componentes
└── config/
    └── config.go       # Configurações do usuário
```
