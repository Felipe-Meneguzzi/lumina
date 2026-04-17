# Contract — `config.Config` additions

**Feature**: 006-ux-polish-pack

## New field: `Editor`

```go
type Config struct {
    // …existing fields unchanged…

    // Editor selects the external editor spawned by the sidebar when opening
    // or creating a file. Accepted values: "nano", "vim", "nvim", or an
    // absolute path to a binary. If the value cannot be resolved via
    // exec.LookPath at the moment a file is opened, Lumina does NOT spawn
    // a pane and surfaces an error via StatusBarNotifyMsg. The configured
    // value is preserved (no silent fallback).
    Editor string `toml:"editor"`
}
```

### Defaults

| Scenario | Value |
|---|---|
| `editor` absent from `config.toml` | `"nano"` |
| `editor` present but empty | `"nano"` (treated as absent) |
| `editor` present with value não-resolvível em `exec.LookPath` | Nenhum pane é criado; `StatusBarNotifyMsg{Level: NotifyError}` exibe `editor '<cfg.Editor>' não encontrado no PATH`. A configuração é preservada (sem fallback silencioso para `"nano"`). |

### `~/.config/lumina/config.toml` — example

```toml
# existing keys…
editor = "nvim"
```

### Migration

Existing `config.toml` files without the `editor` key continue to work unchanged. On first read, Lumina rewrites the file with the new default (same pattern used by every other added field in this project — see `writeDefaults` in `config/config.go`).

## Keybindings additions

New entries in `~/.config/lumina/keybindings.json` (defaults defined in `config/keybindings.go`):

| Action | Default binding | Context |
|---|---|---|
| `sidebar.new_dir` | `alt+d` | Sidebar focused |
| `sidebar.new_file` | `alt+f` | Sidebar focused |
| `sidebar.parent` | `backspace` | Sidebar focused |
| `sidebar.enter` | `enter` | Sidebar focused (existing — extend to trigger `OpenInExternalEditorMsg` for files) |

Mouse click focus is **not** exposed as a user-configurable keybinding — it is a fixed behaviour wired directly in `app/app.go`'s mouse handler, consistent with how other mouse-driven interactions (wheel scroll, drag-select) are hardwired today.

## Removed config surface

None. The deprecated editor component had no user-facing configuration.
