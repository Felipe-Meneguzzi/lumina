# Contract: Novos Keybindings

**Feature**: 003-ux-fixes-multiwindow  
**Arquivo**: `config/keybindings.go` + `~/.config/lumina/keybindings.json`

---

## Novos campos em `Keybindings`

### `toggle_sidebar`

| Campo | Valor |
|-------|-------|
| JSON key | `"toggle_sidebar"` |
| Tipo Go | `[]string` |
| Default | `["alt+b"]` |
| Ação | Oculta/exibe a sidebar da janela em foco |
| Escopo | Global (funciona mesmo com terminal em foco) |

### `toggle_statusbar`

| Campo | Valor |
|-------|-------|
| JSON key | `"toggle_statusbar"` |
| Tipo Go | `[]string` |
| Default | `["alt+m"]` |
| Ação | Oculta/exibe o resource monitor globalmente |
| Escopo | Global |

---

## Keybindings adicionados a `GlobalKeys()`

Ambos os novos keybindings DEVEM ser adicionados a `GlobalKeys()` para que não sejam encaminhados ao PTY quando um terminal está em foco.

```go
for _, k := range kb.ToggleSidebar   { reserved[k] = true }
for _, k := range kb.ToggleStatusBar { reserved[k] = true }
```

E a `Action()`:
```go
for _, k := range kb.ToggleSidebar   { if k == key { return "toggle_sidebar"   } }
for _, k := range kb.ToggleStatusBar { if k == key { return "toggle_statusbar" } }
```

E ao merge em `LoadKeybindings()`:
```go
if len(partial.ToggleSidebar) > 0   { kb.ToggleSidebar   = partial.ToggleSidebar   }
if len(partial.ToggleStatusBar) > 0 { kb.ToggleStatusBar = partial.ToggleStatusBar }
```

---

## Keybindings.json de usuário (backward compatible)

Usuários que já têm `~/.config/lumina/keybindings.json` não precisam alterar nada — os novos campos assumem os defaults automaticamente quando ausentes.
