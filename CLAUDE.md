# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Omarchy TUI is a keyboard-driven terminal user interface for browsing, configuring, and launching applications on Linux systems with Hyprland window manager integration. Written in Go using tview/tcell.

## Build & Run Commands

```bash
# Build
go build -o omarchy ./cmd/omarchy

# Run
./omarchy

# View logs (logs written to app.log in current directory)
tail -f app.log
```

No tests exist currently - this is a proof-of-concept.

## Architecture

```
cmd/omarchy/main.go          Entry point: init logger → load config → start TUI
    ↓
internal/tui/
├── app.go                   Root tview.Application, 2-panel layout, global event router
├── controller.go            State management, app launching, config operations
├── apps_view.go             Application list panel, action menu, keybinding dialog
└── bottom_panel.go          Info display panel for selected application

internal/config/
├── loader.go                YAML config loading, .desktop file auto-discovery
└── model.go                 Category, Application, OmarchyConfig structs

internal/exec/runner.go      Launch applications as detached background processes
internal/hypr/bindings.go    Parse/update Hyprland keybinding configuration
internal/logger/logger.go    Thread-safe file logging
```

## Key Design Patterns

**Centralized Event Router**: All keyboard events flow through `app.SetInputCapture()` in app.go. Global shortcuts (q, Esc) are consumed there; navigation events are forwarded to widgets. See `docs/NAVIGATION_SOLUTIONS.md` for details.

**MVC Pattern**: Config models in `internal/config/model.go`, views in `internal/tui/`, controller in `internal/tui/controller.go`. Controller notifies views via callback functions.

**Graceful Degradation**: Hyprland integration is optional - missing hypr config logs a warning but doesn't crash.

## Configuration Files

- **App config**: `~/.config/omarchy.conf.yaml` - categories and applications inventory
- **Hyprland bindings**: `~/.config/hypr/bindings.conf` - keybinding source (read/write)
- **Auto-discovery**: Scans `~/.local/share/applications/*.desktop` if config is empty

## Data Models

```go
type Application struct {
    Name, PackageName, Keybinding, Category, ConfigFile, Icon string
    CustomConfig map[string]string
}

type Category struct {
    ID, Name string
}

type OmarchyConfig struct {
    Categories    []Category
    AppsInventory []Application
}
```

Use `GetAppsByCategory(categoryID)` and `GetCategoryByID(id)` helper methods on OmarchyConfig.

## Keyboard Controls

- `q` - Quit
- `Esc` - Cancel/close dialogs
- `↑/↓` - Navigate application list
- `Enter` - Open action menu for selected app
