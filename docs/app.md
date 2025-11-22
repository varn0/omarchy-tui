# tui/app.go - Root TUI Application

## Purpose
Creates and manages the root `tview.Application` instance, composes the three-panel layout (categories, apps, bottom panel), and handles global application events like quit.

## Responsibilities
- Initialize the root `tview.Application` instance
- Create and compose the three main panels:
  - Left panel (categories) - via `categories_view.go`
  - Right panel (apps) - via `apps_view.go`
  - Bottom panel (info/config) - via `bottom_panel.go`
- Set up the layout using `tview.Flex` for panel arrangement
- Handle global keyboard events:
  - `q` key → quit application
  - `Esc` key → cancel current action (delegated to controller)
- Coordinate focus management between panels
- Start and run the application event loop
- Handle application-level errors

## Scope
- **In Scope:**
  - Root application initialization
  - Panel composition and layout
  - Global keyboard shortcuts
  - Application lifecycle management
  - Focus coordination

- **Out of Scope:**
  - Individual panel rendering logic (delegated to view modules)
  - Business logic (delegated to controller)
  - Configuration loading (handled by config module)
  - App execution (handled by exec module)

## Dependencies
- `github.com/rivo/tview` - TUI library
- `internal/tui/categories_view.go` - Categories panel
- `internal/tui/apps_view.go` - Apps panel
- `internal/tui/bottom_panel.go` - Bottom info panel
- `internal/tui/controller.go` - Business logic controller

## Interfaces
- **Input:**
  - `*config.OmarchyConfig` - Configuration data
- **Output:**
  - `error` - Error if application fails to start/run

## Key Functions
- `NewApp(config *config.OmarchyConfig) (*App, error)` - Create new application instance
- `(a *App) Run() error` - Start the application event loop
- `(a *App) setupLayout() *tview.Flex` - Create and arrange panels
- `(a *App) setupGlobalKeyHandlers()` - Register global keyboard shortcuts

## Layout Structure
```
+------------------+-----------------------------+
| Categories list  |  App list for selected cat  |
| (left panel)     |  (right panel)              |
+------------------------------------------------+
| Bottom Information / Configuration Panel       |
+------------------------------------------------+
```

## Error Handling
- Panel creation failures → return error during initialization
- Application run errors → return error from Run()
- Should handle panics gracefully if possible

## Notes
- Uses `tview.Flex` for flexible layout management
- Panel sizes should be configurable or use reasonable defaults
- Focus management is critical for keyboard navigation
- Future: may support layout resizing or theme configuration
- Should integrate with controller for state management

