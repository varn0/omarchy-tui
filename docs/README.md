# Omarchy TUI - Module Specifications

Technical specifications for each module are located alongside their corresponding `.go` files in the project structure.

## Project Structure

```
omarchy/
├─ cmd/
│  └─ omarchy/
│     ├─ main.go
│     └─ main.md
├─ internal/
│  ├─ config/
│  │  ├─ loader.go
│  │  ├─ loader.md
│  │  ├─ model.go
│  │  └─ model.md
│  ├─ tui/
│  │  ├─ app.go
│  │  ├─ app.md
│  │  ├─ categories_view.go
│  │  ├─ categories_view.md
│  │  ├─ apps_view.go
│  │  ├─ apps_view.md
│  │  ├─ bottom_panel.go
│  │  ├─ bottom_panel.md
│  │  ├─ controller.go
│  │  └─ controller.md
│  └─ exec/
│     ├─ runner.go
│     └─ runner.md
└─ go.mod
```

## Module Specifications

### Entry Point
- **[cmd/omarchy/main.md](../cmd/omarchy/main.md)** - Application entry point (`cmd/omarchy/main.go`)
  - Initializes application, loads configuration, starts TUI

### Configuration Module
- **[internal/config/loader.md](../internal/config/loader.md)** - Configuration loader (`internal/config/loader.go`)
  - Loads and parses YAML configuration file
- **[internal/config/model.md](../internal/config/model.md)** - Data models (`internal/config/model.go`)
  - Defines configuration data structures

### TUI Module
- **[internal/tui/app.md](../internal/tui/app.md)** - Root TUI application (`internal/tui/app.go`)
  - Creates root tview.Application and composes panels
- **[internal/tui/categories_view.md](../internal/tui/categories_view.md)** - Categories view (`internal/tui/categories_view.go`)
  - Renders and handles category list selection
- **[internal/tui/apps_view.md](../internal/tui/apps_view.md)** - Applications view (`internal/tui/apps_view.go`)
  - Renders and handles application list for selected category
- **[internal/tui/bottom_panel.md](../internal/tui/bottom_panel.md)** - Bottom panel (`internal/tui/bottom_panel.go`)
  - Displays contextual information and configuration editor
- **[internal/tui/controller.md](../internal/tui/controller.md)** - TUI controller (`internal/tui/controller.go`)
  - Coordinates views, manages state, handles business logic

### Execution Module
- **[internal/exec/runner.md](../internal/exec/runner.md)** - Application runner (`internal/exec/runner.go`)
  - Launches external system programs

## Specification Format

Each specification file includes:
- **Purpose** - What the module does
- **Responsibilities** - Specific tasks handled
- **Scope** - What's in and out of scope
- **Dependencies** - Required modules and packages
- **Interfaces** - Input/output contracts
- **Key Functions** - Main functions and methods
- **Error Handling** - Error scenarios and handling
- **Notes** - Additional considerations and future enhancements

## Related Documentation

See the main [gpt-specs.md](../gpt-specs.md) for the complete project specification including:
- Technology stack
- YAML configuration format
- Domain model
- UI layout design
- Keyboard navigation
- Error handling strategy

## Implementation Notes

These specifications are based on:
- The `tview` library (github.com/rivo/tview) for TUI components
- Go standard library for file I/O and process execution
- YAML configuration format as specified in the main spec

Each module should be implemented following its specification, maintaining clear separation of concerns and proper error handling.

