# Quickstart: Multiwindow Layout

**Branch**: `002-multiwindow` | **Date**: 2026-04-16

Guia para desenvolvedores que implementarão ou testarão a feature de multiwindow.

---

## Usando o Lumina com Multiwindow

### Abrir um segundo painel

| Ação | Tecla |
|---|---|
| Dividir painel atual (lado a lado) | `Alt+\` |
| Dividir painel atual (empilhado) | `Alt+-` |

O novo painel abre com um terminal vazio. O foco permanece no painel original.

### Navegar entre painéis

| Ação | Tecla |
|---|---|
| Foco → esquerda | `Alt+H` ou `Alt+←` |
| Foco → direita | `Alt+L` ou `Alt+→` |
| Foco → cima | `Alt+K` ou `Alt+↑` |
| Foco → baixo | `Alt+J` ou `Alt+↓` |

O painel ativo tem uma borda com cor de destaque. Os demais têm borda neutra.

### Abrir um arquivo em um painel específico

1. Navegue até o painel desejado com as teclas de foco.
2. Na sidebar (`Alt+B` para focar a sidebar), selecione o arquivo e pressione `Enter`.
3. O arquivo abre no painel ativo no momento da seleção.

### Redimensionar painéis

| Ação | Tecla |
|---|---|
| Expandir painel → direita | `Alt+Shift+L` |
| Recolher painel ← esquerda | `Alt+Shift+H` |
| Expandir painel ↓ baixo | `Alt+Shift+J` |
| Recolher painel ↑ cima | `Alt+Shift+K` |

Cada pressão ajusta o split em 5% do espaço disponível.

### Redimensionar a sidebar

| Ação | Tecla |
|---|---|
| Expandir sidebar | `Alt+Shift+]` |
| Recolher sidebar | `Alt+Shift+[` |

### Fechar painel

| Ação | Tecla |
|---|---|
| Fechar painel ativo | `Alt+Q` |

Se o painel tiver um arquivo com mudanças não salvas, aparece o diálogo de confirmação existente.
Se for o único painel aberto, o comando é ignorado.

---

## Para Desenvolvedores: Adicionando Suporte a um Novo Tipo de Conteúdo

### Passo 1 — Adicionar um novo `PaneKind`

Em `components/layout/layout.go`:

```go
const (
    KindTerminal PaneKind = iota
    KindEditor
    KindMinhaFeature  // NOVO
)
```

### Passo 2 — Inicializar o modelo no split

Em `components/layout/tree.go` (ou onde estiver a lógica de split):

```go
func newLeaf(kind PaneKind, cfg config.Config) (*LeafNode, error) {
    switch kind {
    case KindTerminal:
        m, err := terminal.New(cfg)
        // ...
    case KindEditor:
        m := editor.New(cfg)
        // ...
    case KindMinhaFeature:
        m := minhafeature.New(cfg)
        return &LeafNode{Kind: kind, model: m}, nil
    }
}
```

### Passo 3 — Rotear mensagens específicas

Em `components/layout/layout.go`, no `Update()`:

```go
case msgs.MinhaFeatureMsgTipo:
    // rotear para o leaf correto (ativo ou por PaneID)
```

### Passo 4 — Escrever o unit test

```go
// components/layout/layout_test.go
func TestSplitRoutesMsgToNewPane(t *testing.T) {
    m, _ := layout.New(cfg)
    m2, _ := m.Update(msgs.PaneSplitMsg{Direction: layout.SplitHorizontal})
    // verificar que m2 tem 2 painéis
}
```

---

## Para Desenvolvedores: Testando o Layout Manager em Isolamento

O `layout.Model` deve ser testado sem instanciar `app.Model`:

```go
cfg := config.Default()
m, err := layout.New(cfg)
require.NoError(t, err)

// Dividir
m2, _ := m.Update(msgs.PaneSplitMsg{Direction: layout.SplitHorizontal})
lm := m2.(layout.Model)
assert.Equal(t, 2, lm.PaneCount())

// Mover foco
m3, _ := lm.Update(msgs.PaneFocusMoveMsg{Direction: layout.FocusRight})
// ...

// Fechar
m4, _ := m3.(layout.Model).Update(msgs.PaneCloseMsg{})
assert.Equal(t, 1, m4.(layout.Model).PaneCount())
```

---

## Diagnóstico: Layout Não Redimensiona Corretamente

1. Verifique que `app.handleResize` envia `LayoutResizeMsg` com `Width = totalWidth - sidebarWidth` e `Height = totalHeight - statusBarHeight`.
2. Verifique que `layout.Update` propaga `TerminalResizeMsg` / `EditorResizeMsg` individualmente para cada `LeafNode`.
3. Para terminais: confirmar que `pty.Setsize` é chamado dentro do `tea.Cmd` retornado por `terminal.Update(TerminalResizeMsg{...})`.
