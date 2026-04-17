# Quickstart — UX Polish Pack manual validation

**Feature**: 006-ux-polish-pack
**Audience**: reviewer / developer validating the implementation end-to-end after tasks are completed.

This roteiro exercita cada user story da spec. Faça todos os passos em sequência, partindo de um checkout da branch `006-ux-polish-pack` com a implementação completa aplicada.

## 0. Build e setup

```bash
cd ~/fpm/lumina
go build -o lumina .
./lumina
```

Antes de iniciar, garanta `~/.config/lumina/config.toml` contendo (ou deixe o Lumina escrever no primeiro boot):

```toml
shell = "/bin/zsh"       # ou o shell preferido
editor = "nvim"          # ou "vim" / "nano"
metrics_interval = 1000
sidebar_width = 30
show_hidden = true
theme = "default"
mouse_auto_copy = true
selection_mode = "linear"
```

## 1. US1 — Render fiel e estável (P1)

### 1.1 First-frame correctness

1. Abra o Lumina. Em um painel vazio digite: `claude` (Claude Code CLI).
2. **Espere o prompt inicial aparecer (com o cabeçalho TUI).**
3. **Esperado**: o cabeçalho (logo + informações de versão + caixa de prompt) aparece alinhado no primeiro frame, sem linhas em branco extras no topo, sem caracteres sobrepostos.
4. **Falha**: cabeçalho quebrado que só se corrige ao redimensionar a janela.

### 1.2 High-rate output stability

1. Em qualquer painel: `yes | head -n 200000`
2. Observe o scroll por ~30 segundos (o comando termina em alguns segundos; se quiser sustentar, use `while true; do echo "foo $RANDOM"; done`).
3. **Esperado**: texto em ordem cronológica correta, sem sobreposições, sem linhas fora da borda.
4. **Falha**: linhas truncadas ou desalinhadas que só se corrigem com resize.

### 1.3 SC-002 — Sustentação sob alta taxa (5 min)

1. Em um painel, rode `while true; do echo "linha-$(date +%N)"; done`.
2. Deixe rodar por **5 minutos contínuos**.
3. Faça scroll para cima periodicamente e confirme: sem linhas sobrepostas, sem truncamento lateral, sem artefatos de frame parcial.
4. Interrompa com `Ctrl+C`.
5. **Esperado**: o shell responde no primeiro frame após o sinal; nenhum artefato acumulado.

## 2. US2 — Cursor, contexto e foco por clique (P1)

### 2.1 Split e foco por teclado

1. Pressione o atalho de split horizontal (padrão definido em keybindings). Agora há dois painéis lado-a-lado A e B.
2. Em A: digite `echo oi ` (sem Enter). Em B: digite `pwd`.
3. Alterne foco com o atalho; confirme que o cursor aparece em exatamente um pane e reaparece **no mesmo lugar** ao voltar.

### 2.2 Foco por clique

1. A está focado. Clique com o mouse dentro da área de B.
2. **Esperado**: a borda de B muda para a cor de destaque, a borda de A fica cinza. O cursor aparece em B.
3. Clique em qualquer ponto da sidebar.
4. **Esperado**: a sidebar ganha borda colorida; os painéis ficam cinza.

### 2.3 Drag a partir de pane não-focado

1. A está focado. Clique e arraste sobre texto visível em B.
2. **Esperado**: o foco transfere para B no mousedown, e o texto selecionado é o de B (a seleção inicia no pane recém-focado).

### 2.4 Status bar reflete o pane focado

1. Em A, `cd ~/fpm/lumina && git checkout -b teste-branch && touch x && git add x` — há alterações não-commitadas.
2. Em B, `cd ~` (ou outro diretório sem repo git).
3. Alterne foco entre A e B.
4. **Esperado**: ao focar A, status bar mostra `teste-branch ●`. Ao focar B, status bar **não** mostra nome de branch (campo git ausente).
5. Em A: `git commit -m "x"`.
6. Foque A.
7. **Esperado**: status bar mostra `teste-branch ✓`.

## 3. US3 — Sidebar como file manager com editor externo (P2)

### 3.1 Navegação

1. Foque a sidebar. Navegue até uma pasta com Enter.
2. Pressione Backspace: volta um nível.
3. **Esperado**: o diretório listado no sidebar muda; os terminais não mudam de CWD.

### 3.2 Backspace na raiz

1. Pressione Backspace até chegar à raiz (o working-dir em que você iniciou o Lumina).
2. Pressione Backspace mais uma vez.
3. **Esperado**: a status bar exibe "Já na raiz" por ~2 segundos; a sidebar permanece inalterada.

### 3.3 Criar pasta

1. Com a sidebar focada em um diretório qualquer, pressione Alt+D.
2. **Esperado**: um prompt inline aparece pedindo o nome.
3. Digite `nova-pasta` e pressione Enter.
4. **Esperado**: a pasta é criada e o cursor da sidebar já está dentro dela (a listagem mostra o conteúdo da nova pasta, que está vazio).

### 3.4 Criar arquivo

1. Dentro da pasta recém-criada, pressione Alt+F.
2. Digite `hello.txt` e pressione Enter.
3. **Esperado**: o arquivo é criado e o editor externo configurado (`nvim` no exemplo) abre em um painel de terminal com o arquivo `hello.txt` vazio.
4. Salve e saia do editor (`:wq`).
5. **Esperado**: o painel do editor fecha naturalmente (PTY EOF); o foco retorna ao layout.

### 3.5 Abrir arquivo existente

1. Na sidebar, navegue até um arquivo `.md` qualquer.
2. Pressione Enter.
3. **Esperado**: o editor externo abre o arquivo em um novo painel.

### 3.6 Validações de erro

1. Alt+D → digite `nova-pasta` (mesmo nome da que já existe) → Enter.
2. **Esperado**: prompt exibe erro inline ("já existe"); nenhum diretório novo é criado; o prompt continua aberto aguardando novo nome ou ESC.
3. Pressione ESC.
4. **Esperado**: prompt fecha sem mudanças.

### 3.7 Editor não encontrado

1. Edite `config.toml`: `editor = "editor-que-nao-existe"`. Reinicie Lumina.
2. Tente abrir um arquivo pela sidebar.
3. **Esperado**: status bar mostra notificação de erro: `editor 'editor-que-nao-existe' não encontrado no PATH`. Nenhum painel novo é criado.

## 4. US4 — Relógio na status bar (P3)

1. Abra o Lumina. Observe o lado esquerdo (ou conforme posicionamento definido) da status bar.
2. **Esperado**: um campo no formato `HH:MM` visível, correspondendo ao relógio do sistema.
3. Aguarde pelo menos 1 minuto sem tocar em nada.
4. **Esperado**: o valor exibido avança conforme o relógio (tolerância de 30s, pois o ticker é de 30s).

## 5. Testes automatizados

```bash
go test ./...
```

Deve passar sem falhas. Cobertura esperada:

- `components/terminal/firstrender_test.go` — regressão de first-frame.
- `components/sidebar/create_test.go` — criação dir/file, validações.
- `components/layout/layout_test.go` — hit-test de clique.
- `components/statusbar/statusbar_test.go` — clock tick, glifo git.
- `tests/integration/click_focus_test.go` — flow completo de `MouseMsg` → `ClickFocusMsg` → `FocusChangeMsg` + pass-through.
- `tests/integration/external_editor_test.go` — `OpenInExternalEditorMsg` → spawn de pane com o binário correto.

## 6. Saída

Ao final do roteiro, para sair: `Ctrl+C` ou o atalho de quit definido. Nenhum painel deve ter estado pendurado; o processo retorna 0.
