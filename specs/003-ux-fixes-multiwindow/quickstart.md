# Quickstart: UX Fixes — Multi-Window Layout

**Feature**: 003-ux-fixes-multiwindow  
**Data**: 2026-04-16

---

## O que muda para o usuário

### Novos atalhos

| Atalho | Ação |
|--------|------|
| `Alt+B` | Toggle sidebar da janela em foco (oculta / exibe) |
| `Alt+M` | Toggle resource monitor (oculta / exibe globalmente) |

### Comportamento alterado

| Comportamento | Antes | Depois |
|--------------|-------|--------|
| Foco após split | Permanece no pane original | Move para o novo pane criado |
| Shell do terminal | Pode abrir shell incorreto se config.toml inválido | Valida e faz fallback automático |
| Sidebar: estado por pane | Global (uma visibilidade para todos) | Memoriza por pane focado |
| Sidebar: arrastar borda | Não funcionava | Arraste o divisor sidebar/conteúdo com o mouse |

---

## Guia rápido de uso

### Toggling sidebar

```
# Com qualquer pane em foco:
Alt+B     → oculta sidebar se visível, exibe se oculta

# Redimensionar via teclado (já existia):
Alt+Shift+]   → aumenta sidebar
Alt+Shift+[   → diminui sidebar

# Redimensionar via mouse (novo):
# Arraste a borda entre sidebar e conteúdo
```

### Toggling resource monitor

```
Alt+M     → oculta/exibe o monitor de CPU/RAM/disco
            (afeta toda a aplicação, não por pane)
```

### Fluxo de split com novo comportamento de foco

```
1. Abrir Lumina → pane 1 em foco (terminal)
2. Alt+| → split horizontal → pane 2 criado (direita)
           ← NOVO: foco vai automaticamente para pane 2
3. Alt+H → voltar foco para pane 1 (esquerda)
4. Alt+Q → fechar pane 1 (agora funciona corretamente)
```

---

## Configuração do shell

Se o terminal abrir com shell incorreto, criar `~/.config/lumina/config.toml`:

```toml
shell = "/bin/bash"   # ou /bin/zsh, /usr/bin/fish, etc.
```

O Lumina valida o shell na inicialização e faz fallback para `/bin/bash → /bin/zsh → /bin/sh` se o shell configurado não existir.

---

## Compatibilidade

- Sem breaking changes nos keybindings existentes
- `keybindings.json` existente continua funcionando — novos actions assumem defaults
- `config.toml` existente continua funcionando — validação não rejeita, apenas corrige
