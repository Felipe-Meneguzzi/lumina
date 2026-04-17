# Quickstart: Lumina TUI Core

**Date**: 2026-04-16
**Purpose**: Guia de validação manual do MVP — verificar que todas as User Stories
funcionam corretamente após implementação.

---

## Pré-requisitos

- Go 1.22+ instalado
- Linux ou macOS (PTY não suportado no Windows)
- Terminal com suporte a cores (qualquer terminal moderno: kitty, alacritty, iTerm2, etc.)

---

## Build e Run

```bash
# No diretório raiz do projeto
go build -o lumina .
./lumina
```

Para abrir direto em um arquivo:

```bash
./lumina caminho/para/arquivo.txt
```

---

## Validação: US1 — Terminal Interativo

```bash
# 1. Abrir o Lumina
./lumina

# 2. O terminal deve estar focado por padrão (borda destacada)
#    Você deve ver o prompt do seu shell ($SHELL)

# 3. Executar um comando simples
echo "Lumina funciona!"
# Esperado: "Lumina funciona!" aparece no terminal

# 4. Executar programa interativo
htop
# Esperado: htop abre, responde a teclas, 'q' retorna ao shell

# 5. Testar resize
# Arraste a janela do terminal para redimensionar
# Esperado: conteúdo se adapta sem caracteres corrompidos ('#', '?', boxes estranhos)
```

**Critério de sucesso**: Todos os 3 testes passam sem corrupção visual ✅

---

## Validação: US2 — Explorador de Arquivos

```bash
# 1. Pressionar Ctrl+2 para focar a sidebar
# Esperado: borda da sidebar fica destacada, terminal perde destaque

# 2. Navegar com ↑↓
# Esperado: item selecionado muda visivelmente

# 3. Pressionar → ou Enter em um diretório
# Esperado: diretório expande e mostra filhos

# 4. Pressionar ← em um diretório expandido
# Esperado: diretório colapsa

# 5. Navegar até um arquivo de texto e pressionar Enter
# Esperado: arquivo abre no painel de editor (ou foco muda para editor)
```

**Critério de sucesso**: Navegação completa sem travamentos ✅

---

## Validação: US3 — Editor de Texto

```bash
# 1. Criar arquivo de teste
echo -e "linha 1\nlinha 2\nlinha 3" > /tmp/test-lumina.txt

# 2. Abrir no Lumina
./lumina /tmp/test-lumina.txt

# 3. Pressionar Ctrl+3 para focar o editor (ou já estar focado)
# Esperado: cursor visível na linha 1, coluna 0

# 4. Pressionar ↓ duas vezes para ir à linha 3
# Esperado: cursor move para "linha 3"

# 5. Pressionar End para ir ao fim da linha, depois digitar " EDITADA"
# Esperado: linha 3 agora mostra "linha 3 EDITADA"
#           um indicador de "modificado" (*) aparece no título ou status

# 6. Pressionar Ctrl+S para salvar
# Esperado: status bar mostra "Salvo" brevemente, indicador de modificado desaparece

# 7. Verificar no disco (em outro terminal)
cat /tmp/test-lumina.txt
# Esperado: "linha 3 EDITADA" está no arquivo

# 8. Fazer uma edição e tentar fechar com Ctrl+W sem salvar
# Esperado: dialog de confirmação aparece ("Descartar alterações?")
```

**Critério de sucesso**: Edição, salvamento e confirmação de fechamento funcionam ✅

---

## Validação: US4 — Status Bar com Métricas

```bash
# 1. Abrir o Lumina e observar a status bar (linha inferior)
./lumina
# Esperado: status bar exibe CPU%, memória e diretório atual

# 2. Aguardar 3 segundos
# Esperado: valores de CPU e memória atualizam pelo menos 2 vezes

# 3. No terminal integrado, rodar um processo pesado
yes > /dev/null &
# Esperado: % de CPU na status bar aumenta visivelmente na próxima atualização

# 4. Matar o processo
kill %1

# 5. Se estiver num repositório git, verificar que a branch aparece
# Esperado: "[main]" ou "[feature-branch]" exibido na status bar
```

**Critério de sucesso**: Métricas atualizam em tempo real sem travar a TUI ✅

---

## Validação: Performance (SC-001 a SC-006)

```bash
# SC-006: Tempo de inicialização
time ./lumina
# Esperado: TUI aparece em menos de 500ms (real < 0.5s)
# CTRL+C para sair

# SC-002: Resposta de foco
# Pressionar Ctrl+1, Ctrl+2, Ctrl+3 rapidamente
# Esperado: transição visual imediata (sem lag perceptível)

# SC-004: Editor com arquivo grande
python3 -c "print('\n'.join([f'linha {i}' for i in range(10000)]))" > /tmp/big.txt
./lumina /tmp/big.txt
# Navegar com PgDown várias vezes
# Esperado: scroll fluido sem travamento
```

---

## Atalhos de Teclado (referência rápida)

| Atalho | Ação |
|--------|------|
| `Ctrl+1` | Focar terminal |
| `Ctrl+2` | Focar sidebar |
| `Ctrl+3` | Focar editor |
| `Ctrl+S` | Salvar arquivo (no editor) |
| `Ctrl+W` | Fechar editor |
| `Ctrl+C` | Sair do Lumina (fora do modo PTY raw) |
| `?` | Exibir ajuda de atalhos |
